package finance

// TransactionResponse represents a transaction in the response
type TransactionResponse struct {
	ID          int              `json:"id"`
	UserID      int              `json:"user_id"`
	Category    CategoryResponse `json:"category"`
	CurrencyID  int              `json:"currency_id"`
	Amount      float64          `json:"amount"`
	Description string           `json:"description"`
	Date        string           `json:"date"`
	Type        string           `json:"type"`
	CreatedAt   string           `json:"created_at"`
}
