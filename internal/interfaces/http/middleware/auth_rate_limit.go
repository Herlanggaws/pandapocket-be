package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"panda-pocket/internal/interfaces/http/handlers"

	"github.com/gin-gonic/gin"
)

const authLimitWindow = 15 * time.Minute

// AuthAttemptLimiter counts auth attempts in a fixed window.
type AuthAttemptLimiter struct {
	mu     sync.Mutex
	window time.Duration
	hits   map[string]authWindow
	now    func() time.Time
}

type authWindow struct {
	started time.Time
	count   int
}

func NewAuthAttemptLimiter(window time.Duration) *AuthAttemptLimiter {
	if window <= 0 {
		window = authLimitWindow
	}
	return &AuthAttemptLimiter{
		window: window,
		hits:   map[string]authWindow{},
		now:    time.Now,
	}
}

// Limit rejects the request when the IP, or the email when countEmail is set, exceeds max.
func (l *AuthAttemptLimiter) Limit(route string, max int, countEmail bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := []string{route + ":ip:" + c.ClientIP()}
		if countEmail {
			if email := emailFromBody(c); email != "" {
				keys = append(keys, route+":email:"+email)
			}
		}
		allowed, retryAfter := l.allow(keys, max)
		if allowed {
			c.Next()
			return
		}
		c.Header("Retry-After", strconv.Itoa(retryAfter))
		handlers.SendErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many attempts. Try again later.")
		c.Abort()
	}
}

func (l *AuthAttemptLimiter) allow(keys []string, max int) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	retryAfter := 1
	blocked := false
	for _, key := range keys {
		slot := l.hits[key]
		if slot.started.IsZero() || now.Sub(slot.started) >= l.window {
			slot = authWindow{started: now, count: 0}
		}
		slot.count++
		l.hits[key] = slot
		remaining := int(l.window.Seconds() - now.Sub(slot.started).Seconds())
		if remaining < 1 {
			remaining = 1
		}
		if slot.count > max {
			blocked = true
			if remaining > retryAfter {
				retryAfter = remaining
			}
		}
	}
	return !blocked, retryAfter
}

func emailFromBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(payload.Email))
}
