package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthAttemptLimiterBlocksAfterMax(t *testing.T) {
	limiter := NewAuthAttemptLimiter(15 * time.Minute)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", limiter.Limit("login", 2, true), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for i := 0; i < 2; i++ {
		status := postLogin(router, `{"email":"ada@example.com"}`)
		if status != http.StatusNoContent {
			t.Fatalf("attempt %d status=%d", i+1, status)
		}
	}
	recorder := postLoginRecorder(router, `{"email":"Ada@Example.com"}`)
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status=%d", recorder.Code)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("missing Retry-After")
	}
}

func TestAuthAttemptLimiterCountsEmailSeparatelyFromIP(t *testing.T) {
	limiter := NewAuthAttemptLimiter(15 * time.Minute)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", limiter.Limit("login", 1, true), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	if status := postLogin(router, `{"email":"one@example.com"}`); status != http.StatusNoContent {
		t.Fatalf("first status=%d", status)
	}
	if status := postLogin(router, `{"email":"two@example.com"}`); status != http.StatusTooManyRequests {
		t.Fatalf("same IP status=%d", status)
	}
}

func postLogin(router *gin.Engine, body string) int {
	return postLoginRecorder(router, body).Code
}

func postLoginRecorder(router *gin.Engine, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
