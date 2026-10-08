package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appMCP "panda-pocket/internal/application/mcp"
	"panda-pocket/internal/domain/entitlement"

	"github.com/gin-gonic/gin"
)

type mcpTokenRepo struct {
	stored *appMCP.StoredToken
}

func (r *mcpTokenRepo) FindByUserID(context.Context, int) (*appMCP.StoredToken, error) {
	if r.stored == nil {
		return nil, appMCP.ErrNotFound
	}
	copy := *r.stored
	return &copy, nil
}

func (r *mcpTokenRepo) FindActiveByHash(_ context.Context, hash string) (*appMCP.StoredToken, error) {
	if r.stored == nil || r.stored.TokenHash != hash || r.stored.RevokedAt != nil {
		return nil, appMCP.ErrNotFound
	}
	copy := *r.stored
	return &copy, nil
}

func (r *mcpTokenRepo) Save(_ context.Context, token appMCP.StoredToken) error {
	r.stored = &token
	return nil
}

func (r *mcpTokenRepo) Revoke(context.Context, int, time.Time) error {
	return nil
}

func TestMCPServeAcceptsSmallBody(t *testing.T) {
	handler, token := newMCPHandler(t)
	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)

	status, payload := serveMCP(handler, token, body)
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, payload)
	}
	var rpc struct {
		Result map[string]any `json:"result"`
		Error  any            `json:"error"`
	}
	if err := json.Unmarshal(payload, &rpc); err != nil {
		t.Fatal(err)
	}
	if rpc.Error != nil || rpc.Result == nil {
		t.Fatalf("payload=%s", payload)
	}
}

func TestMCPServeRejectsBodyOverLimit(t *testing.T) {
	handler, token := newMCPHandler(t)
	body := bytes.Repeat([]byte("x"), mcpMaxBodyBytes+1)

	status, payload := serveMCP(handler, token, body)
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d body=%s", status, payload)
	}
	var resp APIResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil || resp.Error.ErrorCode != "MCP_BODY_TOO_LARGE" {
		t.Fatalf("payload=%s", payload)
	}
}

func newMCPHandler(t *testing.T) (*MCPHandlers, string) {
	t.Helper()
	repo := &mcpTokenRepo{}
	tokens := appMCP.NewTokenService(repo, entitlement.StaticChecker{Pro: true})
	issued, err := tokens.Issue(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	server := appMCP.NewServer(entitlement.StaticChecker{Pro: true}, nil)
	return NewMCPHandlers(tokens, server), issued.Token
}

func serveMCP(handler *MCPHandlers, token string, body []byte) (int, []byte) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	ctx.Request = request
	handler.Serve(ctx)
	return recorder.Code, recorder.Body.Bytes()
}
