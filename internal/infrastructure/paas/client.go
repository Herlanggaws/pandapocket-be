package paas

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://ai.paas.id"
const defaultModel = "glm-5.3-flash"

// 1024 is spent entirely on reasoning for a finance snapshot, so the visible
// answer never starts. 4096 leaves room for that thinking plus a short reply.
const streamMaxTokens = 4096

// glm-5.3-flash always thinks. A non-stream completion with a small cap spends
// every token on reasoning_content and returns empty visible content. Streaming
// still yields content. Requests below this floor are raised so the label or
// answer can appear after that reasoning.
const minVisibleMaxTokens = 256

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest intentionally omits enable_thinking. glm-5.3-flash rejects false
// ("always engages in thinking") and rejects non-boolean values. Omitting the
// field still streams a visible answer.
type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
	MaxTokens *int      `json:"max_tokens,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
	} `json:"choices"`
}

type visionContentPart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *visionImageURL `json:"image_url,omitempty"`
}

type visionImageURL struct {
	URL string `json:"url"`
}

type visionMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type visionChatRequest struct {
	Model     string          `json:"model"`
	Messages  []visionMessage `json:"messages"`
	Stream    bool            `json:"stream"`
	MaxTokens *int            `json:"max_tokens,omitempty"`
}

type Client struct {
	apiKey      string
	baseURL     string
	model       string
	visionModel string
	httpClient  *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey:      os.Getenv("PAAS_AI_API_KEY"),
		baseURL:     strings.TrimRight(getEnv("PAAS_AI_BASE_URL", defaultBaseURL), "/"),
		model:       getEnv("PAAS_AI_MODEL", defaultModel),
		visionModel: os.Getenv("PAAS_AI_VISION_MODEL"),
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *Client) Configured() bool       { return c.apiKey != "" }
func (c *Client) Model() string          { return c.model }
func (c *Client) VisionConfigured() bool { return c.apiKey != "" && c.visionModel != "" }

// CompleteChat returns visible assistant text. This model is streamed even for
// callers that only need the final string: a non-stream body spends the token
// cap on reasoning and comes back with empty content.
func (c *Client) CompleteChat(ctx context.Context, messages []Message, maxTokens int) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("PAAS_AI_API_KEY is not configured")
	}
	limit := maxTokens
	if limit < minVisibleMaxTokens {
		limit = minVisibleMaxTokens
	}
	text, _, _, err := c.streamChat(ctx, messages, limit, nil)
	return strings.TrimSpace(text), err
}

// StreamChat writes content deltas to onDelta and returns full text + usage (usage may be 0 if provider omits it).
func (c *Client) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) (string, int, int, error) {
	if !c.Configured() {
		return "", 0, 0, fmt.Errorf("PAAS_AI_API_KEY is not configured")
	}
	return c.streamChat(ctx, messages, streamMaxTokens, onDelta)
}

// CompleteVision sends one image as an OpenAI-style content part. The chat
// model stays on PAAS_AI_MODEL; receipt scans never fall back to it.
func (c *Client) CompleteVision(ctx context.Context, systemPrompt, userText, mime string, image []byte, maxTokens int) (string, error) {
	if !c.Configured() {
		return "", fmt.Errorf("PAAS_AI_API_KEY is not configured")
	}
	if c.visionModel == "" {
		return "", fmt.Errorf("PAAS_AI_VISION_MODEL is not configured")
	}
	limit := maxTokens
	if limit < minVisibleMaxTokens {
		limit = minVisibleMaxTokens
	}
	dataURL := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(image)
	body, err := json.Marshal(visionChatRequest{
		Model: c.visionModel,
		Messages: []visionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: []visionContentPart{
				{Type: "text", Text: userText},
				{Type: "image_url", ImageURL: &visionImageURL{URL: dataURL}},
			}},
		},
		Stream:    true,
		MaxTokens: &limit,
	})
	if err != nil {
		return "", err
	}
	text, _, _, err := c.postStream(ctx, body, nil)
	return strings.TrimSpace(text), err
}

func (c *Client) streamChat(ctx context.Context, messages []Message, maxTokens int, onDelta func(string) error) (string, int, int, error) {
	body, err := json.Marshal(ChatRequest{
		Model:     c.model,
		Messages:  messages,
		Stream:    true,
		MaxTokens: &maxTokens,
	})
	if err != nil {
		return "", 0, 0, err
	}
	return c.postStream(ctx, body, onDelta)
}

func (c *Client) postStream(ctx context.Context, body []byte, onDelta func(string) error) (string, int, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", 0, 0, fmt.Errorf("paas chat failed: status %d: %s", resp.StatusCode, string(raw))
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		// reasoning_content is scratchpad (often English, and it restates the
		// question). Only visible content is the reply.
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		full.WriteString(delta)
		if onDelta != nil {
			if err := onDelta(delta); err != nil {
				return full.String(), 0, 0, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), 0, 0, err
	}
	return full.String(), 0, 0, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
