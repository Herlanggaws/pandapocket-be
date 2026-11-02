package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appAI "panda-pocket/internal/application/ai"
	domainAI "panda-pocket/internal/domain/ai"

	"github.com/gin-gonic/gin"
)

func TestMapReceiptErrTimeoutIs504(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	h := &AIAdvisorHandlers{}

	h.mapReceiptErr(c, appAI.ErrReceiptTimeout)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("status=%d", recorder.Code)
	}
	var body APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error == nil || body.Error.ErrorCode != "RECEIPT_SCAN_TIMEOUT" {
		t.Fatalf("body=%+v", body.Error)
	}
}

func TestMapReceiptErrCanceledIs499(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	h := &AIAdvisorHandlers{}

	h.mapReceiptErr(c, context.Canceled)

	if recorder.Code != 499 {
		t.Fatalf("status=%d", recorder.Code)
	}
}

func TestMapReceiptErrUpstreamIs502(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	h := &AIAdvisorHandlers{}

	h.mapReceiptErr(c, domainAI.ErrUpstream)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status=%d", recorder.Code)
	}
	var body APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error == nil || body.Error.ErrorCode != "RECEIPT_SCAN_ERROR" {
		t.Fatalf("body=%+v", body.Error)
	}
}
