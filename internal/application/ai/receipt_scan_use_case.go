package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
)

const (
	receiptScanMaxTokens = 4096
	ReceiptMaxImageBytes = 4 << 20
	receiptMerchantMax   = 120
)

var (
	ErrReceiptInvalid    = errors.New("receipt image invalid")
	ErrReceiptUnreadable = errors.New("receipt unreadable")
)

type ReceiptCategory struct {
	ID   int
	Name string
}

type ReceiptScanResponse struct {
	Merchant     string       `json:"merchant"`
	Amount       float64      `json:"amount"`
	Date         string       `json:"date,omitempty"`
	CategoryID   *int         `json:"category_id,omitempty"`
	CurrencyCode string       `json:"currency_code,omitempty"`
	Credits      *CreditsView `json:"credits"`
}

type VisionCompleter interface {
	CompleteVision(ctx context.Context, systemPrompt, userText, mime string, image []byte, maxTokens int) (string, error)
}

type ReceiptScanUseCase struct {
	credits      *CreditService
	entitlements entitlement.Checker
	categories   func(ctx context.Context, userID int) ([]ReceiptCategory, error)
	vision       VisionCompleter
}

func NewReceiptScanUseCase(
	credits *CreditService,
	entitlements entitlement.Checker,
	categories func(ctx context.Context, userID int) ([]ReceiptCategory, error),
	vision VisionCompleter,
) *ReceiptScanUseCase {
	return &ReceiptScanUseCase{
		credits:      credits,
		entitlements: entitlements,
		categories:   categories,
		vision:       vision,
	}
}

func (uc *ReceiptScanUseCase) Execute(ctx context.Context, userID int, image []byte) (*ReceiptScanResponse, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	mime, err := receiptImageMIME(image)
	if err != nil {
		return nil, err
	}
	creditsBefore, err := uc.credits.View(ctx, userID)
	if err != nil {
		return nil, err
	}
	if creditsBefore.Available < 1 {
		return nil, domainAI.ErrCreditsRequired
	}

	categories := []ReceiptCategory{}
	if uc.categories != nil {
		categories, err = uc.categories(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	reply, err := uc.vision.CompleteVision(ctx, receiptScanPrompt(), receiptUserText(categories), mime, image, receiptScanMaxTokens)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			return nil, domainAI.ErrNotConfigured
		}
		return nil, domainAI.ErrUpstream
	}

	draft, err := parseReceiptDraft(reply, categories)
	if err != nil {
		return nil, err
	}
	creditsAfter, err := uc.credits.SpendOne(ctx, userID)
	if err != nil {
		return nil, err
	}
	draft.Credits = creditsAfter
	return draft, nil
}

func receiptImageMIME(image []byte) (string, error) {
	if len(image) == 0 || len(image) > ReceiptMaxImageBytes {
		return "", ErrReceiptInvalid
	}
	if len(image) >= 12 && string(image[0:4]) == "RIFF" && string(image[8:12]) == "WEBP" {
		return "image/webp", nil
	}
	switch kind := http.DetectContentType(image); kind {
	case "image/jpeg", "image/png", "image/webp":
		return kind, nil
	default:
		return "", ErrReceiptInvalid
	}
}

func receiptScanPrompt() string {
	return `You extract one expense from a receipt photo for Berbudget.
Return JSON only, no markdown fences:
{"merchant":"","amount":0,"date":"YYYY-MM-DD","category_name":"","currency_code":""}
amount is the final amount paid, not a line subtotal. Use 0 when the total is unreadable.
Do not invent an amount.
date is the receipt date when printed, otherwise "".
currency_code is the printed ISO 4217 code, otherwise "".
category_name must be copied exactly from the allowed list in the user message, or "" when none fits.
Do not convert currency.`
}

func receiptUserText(categories []ReceiptCategory) string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		if name := strings.TrimSpace(category.Name); name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "Allowed categories: none."
	}
	return "Allowed categories: " + strings.Join(names, ", ")
}

func parseReceiptDraft(raw string, categories []ReceiptCategory) (*ReceiptScanResponse, error) {
	text := stripJSONFence(raw)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return nil, ErrReceiptUnreadable
	}

	var parsed struct {
		Merchant     string  `json:"merchant"`
		Amount       float64 `json:"amount"`
		Date         string  `json:"date"`
		CategoryName string  `json:"category_name"`
		CurrencyCode string  `json:"currency_code"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &parsed); err != nil || parsed.Amount <= 0 {
		return nil, ErrReceiptUnreadable
	}

	draft := &ReceiptScanResponse{
		Merchant:     truncateMerchant(parsed.Merchant),
		Amount:       parsed.Amount,
		Date:         validReceiptDate(parsed.Date),
		CategoryID:   matchReceiptCategory(parsed.CategoryName, categories),
		CurrencyCode: validCurrencyCode(parsed.CurrencyCode),
	}
	return draft, nil
}

func stripJSONFence(raw string) string {
	text := strings.TrimSpace(raw)
	if !strings.HasPrefix(text, "```") {
		return text
	}
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	if idx := strings.LastIndex(text, "```"); idx >= 0 {
		text = text[:idx]
	}
	return strings.TrimSpace(text)
}

func truncateMerchant(value string) string {
	merchant := strings.TrimSpace(value)
	if len(merchant) <= receiptMerchantMax {
		return merchant
	}
	return merchant[:receiptMerchantMax]
}

func validReceiptDate(value string) string {
	date := strings.TrimSpace(value)
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ""
	}
	return date
}

func validCurrencyCode(value string) string {
	code := strings.TrimSpace(value)
	if len(code) != 3 {
		return ""
	}
	for _, char := range code {
		if !unicode.IsLetter(char) {
			return ""
		}
	}
	return strings.ToUpper(code)
}

func matchReceiptCategory(name string, categories []ReceiptCategory) *int {
	key := normalizeLabel(name)
	if key == "" {
		return nil
	}
	for _, category := range categories {
		if normalizeLabel(category.Name) == key {
			id := category.ID
			return &id
		}
	}
	return nil
}

func normalizeLabel(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}
