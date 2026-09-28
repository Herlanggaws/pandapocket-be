package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

const defaultBaseURL = "https://api.frankfurter.app"

// Client fetches the ECB daily rate book from Frankfurter.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return NewClientWithBaseURL(os.Getenv("FX_RATES_BASE_URL"))
}

func NewClientWithBaseURL(baseURL string) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type latestResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// LatestEUR returns units of each currency per 1 EUR, including EUR itself.
func (c *Client) LatestEUR(ctx context.Context) ([]domainFinance.FxQuote, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/latest?from=EUR", nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frankfurter status %d", response.StatusCode)
	}

	var payload latestResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !strings.EqualFold(payload.Base, "EUR") || len(payload.Rates) == 0 {
		return nil, fmt.Errorf("frankfurter response missing EUR rates")
	}
	asOf, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		return nil, fmt.Errorf("frankfurter date: %w", err)
	}
	amount := payload.Amount
	if amount == 0 {
		amount = 1
	}

	quotes := make([]domainFinance.FxQuote, 0, len(payload.Rates)+1)
	quotes = append(quotes, domainFinance.FxQuote{Code: "EUR", UnitsPerEUR: 1, AsOf: asOf})
	for code, rate := range payload.Rates {
		if rate <= 0 {
			continue
		}
		quotes = append(quotes, domainFinance.FxQuote{
			Code:        strings.ToUpper(code),
			UnitsPerEUR: rate / amount,
			AsOf:        asOf,
		})
	}
	return quotes, nil
}
