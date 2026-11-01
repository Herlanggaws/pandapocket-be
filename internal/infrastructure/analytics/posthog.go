package analytics

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/posthog/posthog-go"
)

var (
	mu          sync.Mutex
	client      posthog.Client
	testCapture func(userID int, event string, properties map[string]interface{})
)

// SetTestCapture installs a hook used by unit tests. Pass nil to clear.
func SetTestCapture(fn func(userID int, event string, properties map[string]interface{})) {
	mu.Lock()
	defer mu.Unlock()
	testCapture = fn
}

// InitFromEnv configures PostHog when POSTHOG_API_KEY is set. Safe to call with empty key.
func InitFromEnv() {
	apiKey := os.Getenv("POSTHOG_API_KEY")
	if apiKey == "" {
		return
	}
	endpoint := os.Getenv("POSTHOG_HOST")
	if endpoint == "" {
		endpoint = "https://us.i.posthog.com"
	}
	c, err := posthog.NewWithConfig(apiKey, posthog.Config{Endpoint: endpoint})
	if err != nil {
		log.Printf("posthog init skipped: %v", err)
		return
	}
	mu.Lock()
	client = c
	mu.Unlock()
}

// Close flushes queued events. No-op when not initialized.
func Close() {
	mu.Lock()
	c := client
	client = nil
	mu.Unlock()
	if c != nil {
		_ = c.Close()
	}
}

// Capture sends a product event for the user. Never include amounts, notes, or email.
func Capture(userID int, event string, properties map[string]interface{}) {
	mu.Lock()
	c := client
	hook := testCapture
	mu.Unlock()
	if hook != nil {
		hook(userID, event, properties)
	}
	if c == nil || userID <= 0 || event == "" {
		return
	}
	props := posthog.NewProperties()
	for key, value := range properties {
		props.Set(key, value)
	}
	if err := c.Enqueue(posthog.Capture{
		DistinctId: strconv.Itoa(userID),
		Event:      event,
		Properties: props,
	}); err != nil {
		log.Printf("posthog capture %s: %v", event, err)
	}
}
