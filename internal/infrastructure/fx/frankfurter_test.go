package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientLatestEUR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/latest" || r.URL.Query().Get("from") != "EUR" {
			t.Fatalf("unexpected request %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"amount":1,"base":"EUR","date":"2026-09-26","rates":{"USD":1.17,"IDR":17800}}`))
	}))
	defer server.Close()

	quotes, err := NewClientWithBaseURL(server.URL).LatestEUR(context.Background())
	if err != nil {
		t.Fatalf("LatestEUR: %v", err)
	}
	if len(quotes) != 3 {
		t.Fatalf("quotes = %d, want 3", len(quotes))
	}
	byCode := map[string]float64{}
	for _, quote := range quotes {
		byCode[quote.Code] = quote.UnitsPerEUR
		if quote.AsOf.Format("2006-01-02") != "2026-09-26" {
			t.Fatalf("as_of = %s", quote.AsOf.Format("2006-01-02"))
		}
	}
	if byCode["EUR"] != 1 || byCode["USD"] != 1.17 || byCode["IDR"] != 17800 {
		t.Fatalf("rates = %#v", byCode)
	}
}

func TestClientLatestEUR_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := NewClientWithBaseURL(server.URL).LatestEUR(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
