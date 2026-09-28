package paas

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamChatUsesVisibleContentAndFallsBackToReasoning(t *testing.T) {
	var gotThinking *bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		gotThinking = &req.EnableThinking
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
	if gotThinking == nil || *gotThinking {
		t.Fatal("enable_thinking should be false")
	}
}

func TestStreamChatFallsBackWhenContentEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"answer from reasoning\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test",
		baseURL:    server.URL,
		model:      "glm-5.3-flash",
		httpClient: server.Client(),
	}
	full, _, _, err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(string) error {
		return nil
	})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if full != "answer from reasoning" {
		t.Fatalf("got %q", full)
	}
}
