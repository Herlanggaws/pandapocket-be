package paas

import (
	"bufio"
	"bytes"
	"context"
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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey:  os.Getenv("PAAS_AI_API_KEY"),
		baseURL: strings.TrimRight(getEnv("PAAS_AI_BASE_URL", defaultBaseURL), "/"),
		model:   getEnv("PAAS_AI_MODEL", defaultModel),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) Configured() bool { return c.apiKey != "" }
func (c *Client) Model() string    { return c.model }

// StreamChat writes content deltas to onDelta and returns full text + usage (usage may be 0 if provider omits it).
func (c *Client) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) (string, int, int, error) {
	if !c.Configured() {
		return "", 0, 0, fmt.Errorf("PAAS_AI_API_KEY is not configured")
	}

	body, err := json.Marshal(ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return "", 0, 0, err
	}

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
