package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"panda-pocket/internal/domain/entitlement"
)

const (
	protocolVersion        = "2025-03-26"
	maxTransactionRows     = 50
	defaultTransactionDays = 30
	premiumToolMessage     = "PREMIUM_REQUIRED: MCP is available on Pro."
)

type Reader interface {
	ListWallets(ctx context.Context, userID int) (any, error)
	ListTransactions(ctx context.Context, userID int, startDate, endDate string, limit int) (any, error)
	ListBudgets(ctx context.Context, userID int) (any, error)
	ListGoals(ctx context.Context, userID int) (any, error)
	NetWorth(ctx context.Context, userID int) (any, error)
}

type Server struct {
	entitlements entitlement.Checker
	reader       Reader
	now          func() time.Time
}

func NewServer(entitlements entitlement.Checker, reader Reader) *Server {
	return &Server{
		entitlements: entitlements,
		reader:       reader,
		now:          time.Now,
	}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type toolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type transactionArgs struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Limit     int    `json:"limit"`
}

// Dispatch runs one JSON-RPC message. isNotification is true when the client sent no id.
func (s *Server) Dispatch(ctx context.Context, userID int, body []byte) (response []byte, isNotification bool, err error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(body, &probe); err != nil {
		return s.mustEncode(rpcResponse{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: -32700, Message: "parse error"},
		}), false, nil
	}
	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Method == "" {
		return s.mustEncode(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32600, Message: "invalid request"},
		}), false, nil
	}
	if !hasRPCID(probe["id"]) {
		return nil, true, nil
	}
	result, rpcErr := s.dispatchMethod(ctx, userID, req)
	return s.mustEncode(rpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
		Error:   rpcErr,
	}), false, nil
}

func (s *Server) dispatchMethod(ctx context.Context, userID int, req rpcRequest) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return initializeResult(), nil
	case "ping", "notifications/initialized":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": toolDefinitions()}, nil
	case "tools/call":
		return s.callTool(ctx, userID, req.Params)
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func (s *Server) callTool(ctx context.Context, userID int, params json.RawMessage) (any, *rpcError) {
	var call toolCall
	if len(params) > 0 && string(params) != "null" {
		if err := json.Unmarshal(params, &call); err != nil {
			return nil, &rpcError{Code: -32602, Message: "invalid params"}
		}
	}
	isPro, err := s.entitlements.IsPro(ctx, userID)
	if err != nil {
		return toolError("Could not check Pro access."), nil
	}
	if !isPro {
		return toolError(premiumToolMessage), nil
	}
	payload, err := s.readTool(ctx, userID, call)
	if err != nil {
		return toolError(err.Error()), nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return toolError("Could not encode the tool result."), nil
	}
	return toolOK(string(raw)), nil
}

func (s *Server) readTool(ctx context.Context, userID int, call toolCall) (any, error) {
	switch call.Name {
	case "list_wallets":
		return s.reader.ListWallets(ctx, userID)
	case "list_budgets":
		return s.reader.ListBudgets(ctx, userID)
	case "list_goals":
		return s.reader.ListGoals(ctx, userID)
	case "net_worth":
		return s.reader.NetWorth(ctx, userID)
	case "list_transactions":
		start, end, limit, err := resolveTransactionWindow(call.Arguments, s.now())
		if err != nil {
			return nil, err
		}
		return s.reader.ListTransactions(ctx, userID, start, end, limit)
	default:
		return nil, errors.New("unknown tool")
	}
}

func resolveTransactionWindow(raw json.RawMessage, now time.Time) (string, string, int, error) {
	args := transactionArgs{}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", "", 0, errors.New("invalid list_transactions arguments")
		}
	}
	limit := args.Limit
	if limit <= 0 || limit > maxTransactionRows {
		limit = maxTransactionRows
	}
	end := now
	start := now.AddDate(0, 0, -defaultTransactionDays)
	if args.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", args.EndDate)
		if err != nil {
			return "", "", 0, errors.New("end_date must be YYYY-MM-DD")
		}
		end = parsed
		start = parsed.AddDate(0, 0, -defaultTransactionDays)
	}
	if args.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", args.StartDate)
		if err != nil {
			return "", "", 0, errors.New("start_date must be YYYY-MM-DD")
		}
		start = parsed
	}
	return start.Format("2006-01-02"), end.Format("2006-01-02"), limit, nil
}

func hasRPCID(id json.RawMessage) bool {
	if len(id) == 0 || string(id) == "null" {
		return false
	}
	return true
}

func initializeResult() map[string]any {
	return map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
		"serverInfo":      map[string]any{"name": "berbudget", "version": "1.0.0"},
		"instructions":    "Read-only access to this user's Berbudget wallets, transactions, budgets, goals, and net worth. Never claim a transaction was saved.",
	}
}

func toolDefinitions() []map[string]any {
	empty := map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
	return []map[string]any{
		{"name": "list_wallets", "description": "List the user's non-archived wallets and balances.", "inputSchema": empty},
		{"name": "list_transactions", "description": "List income and expense rows. Defaults to the last 30 days, at most 50 rows.", "inputSchema": transactionSchema()},
		{"name": "list_budgets", "description": "List the user's budgets.", "inputSchema": empty},
		{"name": "list_goals", "description": "List the user's savings goals.", "inputSchema": empty},
		{"name": "net_worth", "description": "Net worth summary in the user's primary currency.", "inputSchema": empty},
	}
}

func transactionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"start_date": map[string]any{"type": "string", "description": "YYYY-MM-DD. Defaults to 30 days before end_date."},
			"end_date":   map[string]any{"type": "string", "description": "YYYY-MM-DD. Defaults to today."},
			"limit":      map[string]any{"type": "integer", "description": "Max rows. Capped at 50."},
		},
		"additionalProperties": false,
	}
}

func toolOK(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": false,
	}
}

func toolError(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": true,
	}
}

func (s *Server) mustEncode(resp rpcResponse) []byte {
	raw, err := json.Marshal(resp)
	if err != nil {
		return []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"internal error"}}`)
	}
	return raw
}
