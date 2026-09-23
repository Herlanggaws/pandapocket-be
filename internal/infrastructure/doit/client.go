package doit

import (
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

const defaultBaseURL = "https://pay.doit.id"

// CreatePaymentRequest is the body for POST /v1/payments.
type CreatePaymentRequest struct {
	Amount    int                    `json:"amount"`
	Rail      string                 `json:"rail,omitempty"`
	Reference string                 `json:"reference"`
	ExpiresIn int                    `json:"expires_in,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ReturnURL string                 `json:"return_url,omitempty"`
}

// CreatePaymentResponse is the relevant subset of Doit payment create response.
type CreatePaymentResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Amount    int    `json:"amount"`
	Reference string `json:"reference"`
	HostedURL string `json:"hosted_url"`
}

// Client talks to pay.doit.id.
type Client struct {
	apiKey     string
	baseURL    string
	returnURL  string
	httpClient *http.Client
}

// NewClient constructs a Client from env (DOIT_API_KEY, DOIT_RETURN_URL, optional DOIT_BASE_URL).
func NewClient() *Client {
	return &Client{
		apiKey:    os.Getenv("DOIT_API_KEY"),
		baseURL:   getEnv("DOIT_BASE_URL", defaultBaseURL),
		returnURL: os.Getenv("DOIT_RETURN_URL"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ReturnURL exposes the configured return URL for checkout.
func (c *Client) ReturnURL() string { return c.returnURL }

// Configured reports whether an API key is present.
func (c *Client) Configured() bool { return c.apiKey != "" }

// CreatePayment creates a one-shot payment. idempotencyKey is required by Doit.
func (c *Client) CreatePayment(ctx context.Context, idempotencyKey string, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("DOIT_API_KEY is not configured")
	}
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key is required")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/payments", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("doit create payment failed: status %d: %s", resp.StatusCode, truncateDoitBody(respBody))
	}

	var out CreatePaymentResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	if out.HostedURL == "" || out.ID == "" {
		return nil, fmt.Errorf("doit create payment returned incomplete response")
	}
	return &out, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncateDoitBody(body []byte) string {
	const max = 300
	s := strings.TrimSpace(string(body))
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
