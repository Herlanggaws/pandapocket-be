package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"panda-pocket/internal/domain/entitlement"
)

type memRepo struct {
	byUser map[int]StoredToken
}

func (r *memRepo) FindByUserID(_ context.Context, userID int) (*StoredToken, error) {
	token, ok := r.byUser[userID]
	if !ok {
		return nil, ErrNotFound
	}
	copy := token
	return &copy, nil
}

func (r *memRepo) FindActiveByHash(_ context.Context, hash string) (*StoredToken, error) {
	for _, token := range r.byUser {
		if token.TokenHash == hash && token.RevokedAt == nil {
			copy := token
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (r *memRepo) Save(_ context.Context, token StoredToken) error {
	if r.byUser == nil {
		r.byUser = map[int]StoredToken{}
	}
	token.RevokedAt = nil
	r.byUser[token.UserID] = token
	return nil
}

func (r *memRepo) Revoke(_ context.Context, userID int, at time.Time) error {
	token, ok := r.byUser[userID]
	if !ok || token.RevokedAt != nil {
		return ErrNotFound
	}
	token.RevokedAt = &at
	r.byUser[userID] = token
	return nil
}

func TestIssueStoresHashAndRotateInvalidates(t *testing.T) {
	repo := &memRepo{}
	svc := NewTokenService(repo, entitlement.StaticChecker{Pro: true})
	first, err := svc.Issue(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(first.Token, "bb_mcp_") || len(first.Prefix) != len("bb_mcp_")+8 {
		t.Fatalf("token=%s prefix=%s", first.Token, first.Prefix)
	}
	if repo.byUser[7].TokenHash == first.Token {
		t.Fatal("plaintext must not be stored")
	}
	userID, err := svc.Authenticate(context.Background(), first.Token)
	if err != nil || userID != 7 {
		t.Fatalf("user=%d err=%v", userID, err)
	}

	second, err := svc.Issue(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(context.Background(), first.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("old token err=%v", err)
	}
	if userID, err = svc.Authenticate(context.Background(), second.Token); err != nil || userID != 7 {
		t.Fatalf("new user=%d err=%v", userID, err)
	}
}

func TestRevokeAndFreeAreRejected(t *testing.T) {
	repo := &memRepo{}
	svc := NewTokenService(repo, entitlement.StaticChecker{Pro: true})
	issued, err := svc.Issue(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Revoke(context.Background(), 3); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(context.Background(), issued.Token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked err=%v", err)
	}
	status, err := svc.Status(context.Background(), 3)
	if err != nil || status.Active {
		t.Fatalf("status=%+v err=%v", status, err)
	}

	free := NewTokenService(repo, entitlement.StaticChecker{Pro: false})
	if _, err := free.Issue(context.Background(), 3); !errors.Is(err, entitlement.ErrPremiumRequired) {
		t.Fatalf("free issue err=%v", err)
	}
}

type stubReader struct {
	limit int
	start string
	end   string
}

func (s *stubReader) ListWallets(context.Context, int) (any, error) {
	return map[string]any{"wallets": []any{}}, nil
}
func (s *stubReader) ListTransactions(_ context.Context, _ int, startDate, endDate string, limit int) (any, error) {
	s.start, s.end, s.limit = startDate, endDate, limit
	return map[string]any{"transactions": []any{}}, nil
}
func (s *stubReader) ListBudgets(context.Context, int) (any, error) { return map[string]any{}, nil }
func (s *stubReader) ListGoals(context.Context, int) (any, error)   { return map[string]any{}, nil }
func (s *stubReader) NetWorth(context.Context, int) (any, error)    { return map[string]any{}, nil }

func TestToolsCallRequiresProAndCapsTransactions(t *testing.T) {
	now := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	reader := &stubReader{}
	server := NewServer(entitlement.StaticChecker{Pro: true}, reader)
	server.now = func() time.Time { return now }

	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_transactions","arguments":{"limit":500}}}`)
	raw, isNote, err := server.Dispatch(context.Background(), 1, body)
	if err != nil || isNote {
		t.Fatalf("note=%v err=%v", isNote, err)
	}
	var resp struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Result.IsError {
		t.Fatalf("tool error %s", resp.Result.Content[0].Text)
	}
	if reader.limit != 50 || reader.end != "2026-10-01" || reader.start != "2026-09-01" {
		t.Fatalf("window start=%s end=%s limit=%d", reader.start, reader.end, reader.limit)
	}

	free := NewServer(entitlement.StaticChecker{Pro: false}, reader)
	raw, _, err = free.Dispatch(context.Background(), 1, []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_wallets"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Result.IsError || !strings.Contains(resp.Result.Content[0].Text, "PREMIUM_REQUIRED") {
		t.Fatalf("free result=%s", raw)
	}
}

func TestInitializeAndNotification(t *testing.T) {
	server := NewServer(entitlement.StaticChecker{Pro: true}, &stubReader{})
	raw, isNote, err := server.Dispatch(context.Background(), 1, []byte(`{"jsonrpc":"2.0","id":"a","method":"initialize"}`))
	if err != nil || isNote || !strings.Contains(string(raw), protocolVersion) {
		t.Fatalf("note=%v err=%v body=%s", isNote, err, raw)
	}
	_, isNote, err = server.Dispatch(context.Background(), 1, []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	if err != nil || !isNote {
		t.Fatalf("note=%v err=%v", isNote, err)
	}
}
