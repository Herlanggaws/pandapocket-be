package ai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	domainAI "panda-pocket/internal/domain/ai"
)

func TestBuildReplyMetaKeepsServerNumbers(t *testing.T) {
	contextJSON := `{
		"primary_currency": {"symbol": "Rp"},
		"cashflow_monthly": {"total_income": 1550000, "total_spent": 5120000, "net_amount": -3570000},
		"cashflow_previous_month": {"net_amount": 3200000},
		"health": {"score": 0, "components": {"cashflow": 0}},
		"budgets": [
			{"category": "Utang", "period": "monthly", "percent_used": 93},
			{"category": "Nafkah", "period": "monthly", "percent_used": 94},
			{"category": "Pendidikan", "period": "monthly", "percent_used": 96},
			{"category": "Hiburan", "period": "monthly", "percent_used": 10}
		]
	}`
	raw := "```json\n" + `{
		"sources": ["health", "cashflow", "budgets", "nope", "health"],
		"follow_ups": ["  Berapa yang harus gua sisihin? ", "Pos mana yang paling jebol?", "tiga", "empat"],
		"actions": ["open_wallets", "open_health", "open_wallets", "delete_wallet"]
	}` + "\n```"
	encoded, ok := buildReplyMeta(raw, contextJSON, true)
	if !ok {
		t.Fatal("expected meta")
	}
	var meta storedReplyMeta
	if err := json.Unmarshal([]byte(encoded), &meta); err != nil {
		t.Fatal(err)
	}
	if !meta.CreditSpent {
		t.Fatal("credit_spent")
	}
	if strings.Join(meta.Sources, ",") != "health,cashflow,budgets" {
		t.Fatalf("sources=%v", meta.Sources)
	}
	if len(meta.FollowUps) != 3 || meta.FollowUps[0] != "Berapa yang harus gua sisihin?" {
		t.Fatalf("follow_ups=%v", meta.FollowUps)
	}
	if len(meta.Actions) != 2 || meta.Actions[0].Href != "/settings/wallets" || meta.Actions[1].Href != "/health" {
		t.Fatalf("actions=%v", meta.Actions)
	}
	if len(meta.Cards) != 3 {
		t.Fatalf("cards=%d", len(meta.Cards))
	}
	var budgets struct {
		Items []struct {
			Category    string  `json:"category"`
			PercentUsed float64 `json:"percent_used"`
		} `json:"items"`
	}
	if err := json.Unmarshal(meta.Cards[2], &budgets); err != nil {
		t.Fatal(err)
	}
	if len(budgets.Items) != 3 || budgets.Items[0].Category != "Pendidikan" || budgets.Items[0].PercentUsed != 96 {
		t.Fatalf("budgets=%+v", budgets.Items)
	}
	var health struct {
		Score       int     `json:"score"`
		Tone        string  `json:"tone"`
		PreviousNet float64 `json:"previous_net"`
	}
	if err := json.Unmarshal(meta.Cards[0], &health); err != nil {
		t.Fatal(err)
	}
	if health.Score != 0 || health.Tone != "deficit" || health.PreviousNet != 3200000 {
		t.Fatalf("health=%+v", health)
	}
}

func TestBuildReplyMetaDropsShortFollowUps(t *testing.T) {
	raw := `{"sources":["budgets"],"follow_ups":["Pos mana?","Berapa sisihin?"],"actions":[]}`
	encoded, ok := buildReplyMeta(raw, `{}`, true)
	if !ok {
		t.Fatal("expected meta")
	}
	var meta storedReplyMeta
	if err := json.Unmarshal([]byte(encoded), &meta); err != nil {
		t.Fatal(err)
	}
	if len(meta.FollowUps) != 0 {
		t.Fatalf("follow_ups=%v", meta.FollowUps)
	}
}

func TestBuildReplyMetaRejectsNonJSON(t *testing.T) {
	if _, ok := buildReplyMeta("IN_SCOPE", `{}`, true); ok {
		t.Fatal("expected reject")
	}
}

func TestAdvisorChatStoresReplyMeta(t *testing.T) {
	meta := `{"sources":["budgets"],"follow_ups":["Pos mana?"],"actions":["open_budgets"]}`
	streamer := &stubStreamer{
		full:      "Anggaran pendidikan sudah 96%.",
		metaReply: meta,
	}
	uc, _, threads, threadID := newTestChat(streamer)
	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "Kenapa budget saya jebol?")
	if err != nil {
		t.Fatal(err)
	}
	defer unsub()
	last := drain(ch)
	if last == nil || last.Kind != StreamDone {
		t.Fatalf("want done, got %#v", last)
	}
	msgs, err := threads.ListMessages(context.Background(), threadID)
	if err != nil {
		t.Fatal(err)
	}
	assistant := msgs[len(msgs)-1]
	if assistant.Role != domainAI.RoleAssistant || !strings.Contains(assistant.ReplyMetaJSON, `"credit_spent":true`) {
		t.Fatalf("meta=%q", assistant.ReplyMetaJSON)
	}
	if !strings.Contains(assistant.ReplyMetaJSON, `"open_budgets"`) {
		t.Fatalf("meta=%q", assistant.ReplyMetaJSON)
	}
}

func TestSetFeedbackStoresChoiceOnly(t *testing.T) {
	credits := NewCreditService(&memCredits{balance: domainAI.NewCreditBalance(1, false)}, emptySubs{})
	threads := newMemThreads()
	created, err := threads.Create(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := threads.AppendAssistant(context.Background(), created.ID, "Saran.", "", 0, 0); err != nil {
		t.Fatal(err)
	}
	msgs, err := threads.ListMessages(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	uc := NewThreadUseCases(credits, threads, alwaysPro{})
	if err := uc.SetFeedback(context.Background(), 1, created.ID, msgs[0].ID, "up"); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetFeedback(context.Background(), 1, created.ID, msgs[0].ID, "clear"); err != nil {
		t.Fatal(err)
	}
	stored, err := threads.FindMessage(context.Background(), created.ID, msgs[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored.ReplyMetaJSON, `"feedback"`) {
		t.Fatalf("feedback still set: %s", stored.ReplyMetaJSON)
	}
	if err := uc.SetFeedback(context.Background(), 1, created.ID, msgs[0].ID, "maybe"); err != domainAI.ErrInvalidFeedback {
		t.Fatalf("err=%v", err)
	}
}
