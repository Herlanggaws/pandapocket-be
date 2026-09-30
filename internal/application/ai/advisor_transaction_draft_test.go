package ai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	domainAI "panda-pocket/internal/domain/ai"
)

func testDraftCatalog() TransactionDraftCatalog {
	return TransactionDraftCatalog{
		Wallets: []DraftWalletOption{
			{ID: 1, Name: "Tunai", IsDefault: true},
			{ID: 2, Name: "BCA"},
		},
		Categories: []DraftCategoryOption{
			{ID: 4, Name: "Makanan", Kind: "expense"},
			{ID: 9, Name: "Gaji", Kind: "income"},
		},
	}
}

func TestQuickRecordIntent(t *testing.T) {
	record, confident := quickRecordIntent("catat pengeluaran kopi 25000")
	if !record || !confident {
		t.Fatalf("record=%v confident=%v", record, confident)
	}
	record, confident = quickRecordIntent("berapa pengeluaran kopi")
	if record || !confident {
		t.Fatalf("question record=%v confident=%v", record, confident)
	}
	record, confident = quickRecordIntent("Record coffee 25000 today")
	if !record || !confident {
		t.Fatalf("english record=%v confident=%v", record, confident)
	}
	record, confident = quickRecordIntent("transfer 100rb dari Tunai ke BCA")
	if !record || !confident {
		t.Fatalf("transfer record=%v confident=%v", record, confident)
	}
	_, confident = quickRecordIntent("25000")
	if confident {
		t.Fatal("bare amount should be classified by the model")
	}
}

func TestResolveTransactionDraftExpense(t *testing.T) {
	raw := `{"kind":"expense","amount":25000,"date":"","note":"kopi","wallet_name":"","category_name":"Makanan"}`
	draft, miss := resolveTransactionDraft(raw, testDraftCatalog(), "2026-09-30")
	if miss != "" || draft == nil {
		t.Fatalf("miss=%s draft=%v", miss, draft)
	}
	if draft.Kind != "expense" || draft.Amount != 25000 || draft.WalletID != 1 || draft.Date != "2026-09-30" {
		t.Fatalf("draft=%+v", draft)
	}
	if draft.CategoryID == nil || *draft.CategoryID != 4 || draft.Status != DraftStatusPending {
		t.Fatalf("category=%v status=%s", draft.CategoryID, draft.Status)
	}
}

func TestResolveTransactionDraftMissingAmount(t *testing.T) {
	raw := `{"kind":"expense","amount":0,"note":"kopi","wallet_name":"Tunai","category_name":"Makanan"}`
	draft, miss := resolveTransactionDraft(raw, testDraftCatalog(), "2026-09-30")
	if draft != nil || miss != "amount" {
		t.Fatalf("draft=%v miss=%s", draft, miss)
	}
}

func TestResolveTransactionDraftIgnoresUnknownCategory(t *testing.T) {
	raw := `{"kind":"expense","amount":10000,"wallet_name":"Tunai","category_name":"Gaji"}`
	draft, miss := resolveTransactionDraft(raw, testDraftCatalog(), "2026-09-30")
	if miss != "" || draft == nil || draft.CategoryID != nil {
		t.Fatalf("miss=%s draft=%+v", miss, draft)
	}
}

func TestResolveTransactionDraftTransfer(t *testing.T) {
	raw := `{"kind":"transfer","amount":100000,"wallet_name":"Tunai","to_wallet_name":"BCA","note":"geser"}`
	draft, miss := resolveTransactionDraft(raw, testDraftCatalog(), "2026-09-30")
	if miss != "" || draft == nil || draft.ToWalletID == nil || *draft.ToWalletID != 2 || draft.WalletID != 1 {
		t.Fatalf("miss=%s draft=%+v", miss, draft)
	}

	same := `{"kind":"transfer","amount":100000,"wallet_name":"Tunai","to_wallet_name":"Tunai"}`
	draft, miss = resolveTransactionDraft(same, testDraftCatalog(), "2026-09-30")
	if draft != nil || miss != "transfer" {
		t.Fatalf("same wallet draft=%v miss=%s", draft, miss)
	}
}

func TestAdvisorChatRecordDraftSpendsOnce(t *testing.T) {
	streamer := &stubStreamer{classifyReply: `{"kind":"expense","amount":25000,"note":"kopi","wallet_name":"Tunai","category_name":"Makanan"}`}
	uc, repo, threads, threadID := newTestChat(streamer)
	uc.EnableTransactionDrafts(func(context.Context, int) (TransactionDraftCatalog, error) {
		return testDraftCatalog(), nil
	})

	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "catat pengeluaran kopi 25000")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer unsub()
	last := drain(ch)
	if last == nil || last.Kind != StreamDone {
		t.Fatalf("want done, got %#v", last)
	}
	if streamer.streamCalls != 0 || streamer.completeCalls != 1 {
		t.Fatalf("stream=%d complete=%d", streamer.streamCalls, streamer.completeCalls)
	}
	if repo.balance.Available() != domainAI.IncludedGrant()-1 {
		t.Fatalf("available=%d", repo.balance.Available())
	}
	msgs, err := threads.ListMessages(context.Background(), threadID)
	if err != nil {
		t.Fatal(err)
	}
	var draft TransactionDraft
	if err := json.Unmarshal([]byte(msgs[len(msgs)-1].DraftJSON), &draft); err != nil {
		t.Fatal(err)
	}
	if draft.Amount != 25000 || draft.Kind != "expense" {
		t.Fatalf("draft=%+v", draft)
	}
}

func TestAdvisorChatRecordMissingAmountDoesNotSpend(t *testing.T) {
	streamer := &stubStreamer{classifyReply: `{"kind":"expense","amount":0,"note":"kopi","wallet_name":"Tunai"}`}
	uc, repo, threads, threadID := newTestChat(streamer)
	uc.EnableTransactionDrafts(func(context.Context, int) (TransactionDraftCatalog, error) {
		return testDraftCatalog(), nil
	})
	ch, unsub, err := uc.Start(context.Background(), 1, threadID, "catat pengeluaran kopi")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer unsub()
	last := drain(ch)
	if last == nil || last.Kind != StreamDone {
		t.Fatalf("want done, got %#v", last)
	}
	if last.Credits == nil || last.Credits.Available != domainAI.IncludedGrant() {
		t.Fatalf("credits=%v", last.Credits)
	}
	if repo.balance.Available() != domainAI.IncludedGrant() {
		t.Fatalf("persisted=%d", repo.balance.Available())
	}
	msgs, err := threads.ListMessages(context.Background(), threadID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msgs[len(msgs)-1].Content, "Jumlahnya") {
		t.Fatalf("content=%q", msgs[len(msgs)-1].Content)
	}
}

func TestMarkDraftSavedIsIdempotent(t *testing.T) {
	threads := newMemThreads()
	created, err := threads.Create(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	raw := `{"kind":"expense","amount":10,"date":"2026-09-30","note":"kopi","wallet_id":1,"wallet_name":"Tunai","status":"pending"}`
	if err := threads.AppendAssistant(context.Background(), created.ID, "Draf siap.", raw, 0, 0); err != nil {
		t.Fatal(err)
	}
	msgs, err := threads.ListMessages(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	uc := NewThreadUseCases(NewCreditService(&memCredits{balance: domainAI.NewCreditBalance(1, false)}, emptySubs{}), threads, alwaysPro{})
	if err := uc.MarkDraftSaved(context.Background(), 1, created.ID, msgs[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := uc.MarkDraftSaved(context.Background(), 1, created.ID, msgs[0].ID); err != nil {
		t.Fatal(err)
	}
	saved, err := threads.FindMessage(context.Background(), created.ID, msgs[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(saved.DraftJSON, `"status":"saved"`) {
		t.Fatalf("draft=%s", saved.DraftJSON)
	}
}
