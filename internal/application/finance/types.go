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

// WalletResponse represents a wallet in the response
type WalletResponse struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
