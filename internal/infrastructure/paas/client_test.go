package paas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamChatUsesVisibleContentAndFallsBackToReasoning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertThinkingOmitted(t, r)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"think\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test",
		baseURL:    server.URL,
		model:      "glm-5.3-flash",
		httpClient: server.Client(),
	}
	var streamed strings.Builder
	full, _, _, err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(delta string) error {
		streamed.WriteString(delta)
		return nil
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if full != "hello" || streamed.String() != "hello" {
		t.Fatalf("got full %q streamed %q", full, streamed.String())
	}
}

func TestCompleteChatStreamsAndRaisesSmallTokenCap(t *testing.T) {
	var gotMaxTokens int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if !req.Stream {
			t.Errorf("complete chat should stream, got stream=%v", req.Stream)
		}
		if req.MaxTokens != nil {
			gotMaxTokens = *req.MaxTokens
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"IN_SCOPE\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test",
		baseURL:    server.URL,
		model:      "glm-5.3-flash",
		httpClient: server.Client(),
	}
	text, err := client.CompleteChat(context.Background(), []Message{{Role: "user", Content: "gaji"}}, 16)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if text != "IN_SCOPE" {
		t.Fatalf("got %q", text)
	}
	if gotMaxTokens < minVisibleMaxTokens {
		t.Fatalf("max_tokens %d below floor %d", gotMaxTokens, minVisibleMaxTokens)
	}
}

func assertThinkingOmitted(t *testing.T, r *http.Request) {
	t.Helper()
	body, _ := io.ReadAll(r.Body)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Errorf("decode request: %v", err)
		return
	}
	if _, ok := raw["enable_thinking"]; ok {
		t.Fatal("enable_thinking must be omitted; glm-5.3-flash rejects false")
	}
}

func TestStreamChatIgnoresReasoningWhenContentEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"The user asks in Indonesian: Saran alokasi gaji\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test",
		baseURL:    server.URL,
		model:      "glm-5.3-flash",
		httpClient: server.Client(),
	}
	var streamed strings.Builder
	full, _, _, err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(delta string) error {
		streamed.WriteString(delta)
		return nil
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if full != "" || streamed.Len() != 0 {
		t.Fatalf("got full %q streamed %q", full, streamed.String())
	}
}

func TestCompleteVisionUsesImagePartsAndVisionModel(t *testing.T) {
	image := []byte{0xff, 0xd8, 0xff, 0x00}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Errorf("decode: %v", err)
		}
		if raw["model"] != "vision-test" {
			t.Errorf("model=%v", raw["model"])
		}
		if _, ok := raw["enable_thinking"]; ok {
			t.Error("enable_thinking must be omitted")
		}
		encoded, _ := json.Marshal(raw["messages"])
		if !strings.Contains(string(encoded), base64.StdEncoding.EncodeToString(image)) {
			t.Errorf("image missing from request")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"{\\\"amount\\\":1}\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := &Client{
		apiKey:      "test",
		baseURL:     server.URL,
		model:       "glm-5.3-flash",
		visionModel: "vision-test",
		httpClient:  server.Client(),
	}
	text, err := client.CompleteVision(context.Background(), "system", "user", "image/jpeg", image, 128)
	if err != nil {
		t.Fatal(err)
	}
	if text != `{"amount":1}` {
		t.Fatalf("got %q", text)
	}
}

func TestCompleteVisionRequiresVisionModel(t *testing.T) {
	client := &Client{apiKey: "test", model: "glm-5.3-flash"}
	_, err := client.CompleteVision(context.Background(), "s", "u", "image/jpeg", []byte{1}, 256)
	if err == nil || !strings.Contains(err.Error(), "PAAS_AI_VISION_MODEL") {
		t.Fatalf("err=%v", err)
	}
}
