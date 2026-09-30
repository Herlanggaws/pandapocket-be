package ai

import (
	"context"
	"errors"
	"testing"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
)

type stubVision struct {
	reply string
	err   error
	calls int
	mime  string
}

func (s *stubVision) CompleteVision(_ context.Context, _, _ string, mime string, _ []byte, _ int) (string, error) {
	s.calls++
	s.mime = mime
	if s.err != nil {
		return "", s.err
	}
	return s.reply, nil
}

func jpegBytes() []byte {
	return []byte{0xff, 0xd8, 0xff, 0x00, 0x11}
}

func receiptCategories() func(context.Context, int) ([]ReceiptCategory, error) {
	return func(context.Context, int) ([]ReceiptCategory, error) {
		return []ReceiptCategory{{ID: 7, Name: "Makanan"}}, nil
	}
}

func receiptUseCase(bal *domainAI.CreditBalance, vision *stubVision, checker entitlement.Checker) (*ReceiptScanUseCase, *memCredits) {
	repo := &memCredits{balance: bal}
	credits := NewCreditService(repo, emptySubs{})
	return NewReceiptScanUseCase(credits, checker, receiptCategories(), vision), repo
}

func TestReceiptScanDebitsAfterUsableAmount(t *testing.T) {
	vision := &stubVision{reply: "```json\n{\"merchant\":\"Indomaret\",\"amount\":18500,\"date\":\"2026-09-30\",\"category_name\":\" makanan \",\"currency_code\":\"idr\"}\n```"}
	uc, repo := receiptUseCase(domainAI.NewCreditBalance(1, false), vision, alwaysPro{})

	resp, err := uc.Execute(context.Background(), 1, jpegBytes())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Merchant != "Indomaret" || resp.Amount != 18500 || resp.Date != "2026-09-30" || resp.CurrencyCode != "IDR" {
		t.Fatalf("resp=%+v", resp)
	}
	if resp.CategoryID == nil || *resp.CategoryID != 7 {
		t.Fatalf("category=%v", resp.CategoryID)
	}
	if vision.calls != 1 || vision.mime != "image/jpeg" {
		t.Fatalf("calls=%d mime=%s", vision.calls, vision.mime)
	}
	if repo.balance.IncludedUsed != 1 {
		t.Fatalf("used=%d", repo.balance.IncludedUsed)
	}
}

func TestReceiptScanUnreadableDoesNotDebit(t *testing.T) {
	vision := &stubVision{reply: `{"merchant":"","amount":0}`}
	uc, repo := receiptUseCase(domainAI.NewCreditBalance(1, false), vision, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, jpegBytes())
	if !errors.Is(err, ErrReceiptUnreadable) {
		t.Fatalf("err=%v", err)
	}
	if vision.calls != 1 || repo.balance.IncludedUsed != 0 {
		t.Fatalf("calls=%d used=%d", vision.calls, repo.balance.IncludedUsed)
	}
}

func TestReceiptScanZeroCreditsSkipsModel(t *testing.T) {
	bal := domainAI.NewCreditBalance(1, false)
	bal.IncludedUsed = domainAI.IncludedGrant()
	vision := &stubVision{reply: `{"amount":10}`}
	uc, repo := receiptUseCase(bal, vision, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, jpegBytes())
	if !errors.Is(err, domainAI.ErrCreditsRequired) {
		t.Fatalf("err=%v", err)
	}
	if vision.calls != 0 || repo.balance.IncludedUsed != domainAI.IncludedGrant() {
		t.Fatalf("calls=%d used=%d", vision.calls, repo.balance.IncludedUsed)
	}
}

func TestReceiptScanFreeDoesNotCallModel(t *testing.T) {
	vision := &stubVision{reply: `{"amount":10}`}
	uc, _ := receiptUseCase(domainAI.NewCreditBalance(1, false), vision, neverPro{})

	_, err := uc.Execute(context.Background(), 1, jpegBytes())
	if !errors.Is(err, entitlement.ErrPremiumRequired) {
		t.Fatalf("err=%v", err)
	}
	if vision.calls != 0 {
		t.Fatalf("calls=%d", vision.calls)
	}
}

func TestReceiptScanUpstreamDoesNotDebit(t *testing.T) {
	vision := &stubVision{err: errors.New("upstream down")}
	uc, repo := receiptUseCase(domainAI.NewCreditBalance(1, false), vision, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, jpegBytes())
	if !errors.Is(err, domainAI.ErrUpstream) {
		t.Fatalf("err=%v", err)
	}
	if repo.balance.IncludedUsed != 0 {
		t.Fatalf("used=%d", repo.balance.IncludedUsed)
	}
}
