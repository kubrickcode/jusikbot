package tiingo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/httpclient"
)

const testAPIKey = "test-api-key"

func newTestClient(srv *httptest.Server) *Client {
	hc := httpclient.NewClient(
		srv.URL,
		map[string]string{"Authorization": "Token " + testAPIKey},
		srv.Client(),
		0,
	)
	return NewClient(hc)
}

func TestFetchPrices(t *testing.T) {
	t.Run("normal JSON response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/tiingo/daily/AAPL/prices" {
				t.Errorf("path = %q, want /tiingo/daily/AAPL/prices", r.URL.Path)
			}
			if got := r.URL.Query().Get("startDate"); got != "2024-01-01" {
				t.Errorf("startDate = %q, want 2024-01-01", got)
			}
			if got := r.URL.Query().Get("endDate"); got != "2024-01-31" {
				t.Errorf("endDate = %q, want 2024-01-31", got)
			}
			if got := r.Header.Get("Authorization"); got != "Token test-api-key" {
				t.Errorf("Authorization = %q, want Token test-api-key", got)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{
					"adjClose": 150.25, "adjHigh": 151.0, "adjLow": 149.5,
					"adjOpen": 150.0, "adjVolume": 55000000,
					"close": 150.25, "date": "2024-01-15T00:00:00+00:00",
					"divCash": 0.0, "high": 151.0, "low": 149.5,
					"open": 150.0, "splitFactor": 1.0, "volume": 55000000
				},
				{
					"adjClose": 155.50, "adjHigh": 156.0, "adjLow": 154.0,
					"adjOpen": 155.0, "adjVolume": 48000000,
					"close": 155.50, "date": "2024-01-16T00:00:00+00:00",
					"divCash": 0.0, "high": 156.0, "low": 154.0,
					"open": 155.0, "splitFactor": 1.0, "volume": 48000000
				}
			]`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		prices, err := client.fetchPrices(context.Background(), "AAPL", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 2 {
			t.Fatalf("len(prices) = %d, want 2", len(prices))
		}

		if prices[0].AdjClose != 150.25 {
			t.Errorf("prices[0].AdjClose = %v, want 150.25", prices[0].AdjClose)
		}
		if prices[0].Volume != 55000000 {
			t.Errorf("prices[0].Volume = %v, want 55000000", prices[0].Volume)
		}
		if prices[0].SplitFactor != 1.0 {
			t.Errorf("prices[0].SplitFactor = %v, want 1.0", prices[0].SplitFactor)
		}
		if prices[0].DivCash != 0.0 {
			t.Errorf("prices[0].DivCash = %v, want 0.0", prices[0].DivCash)
		}
		if prices[1].Close != 155.50 {
			t.Errorf("prices[1].Close = %v, want 155.50", prices[1].Close)
		}
	})

	t.Run("rate limit non-JSON body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Rate limit exceeded. Please try again later."))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		_, err := client.fetchPrices(context.Background(), "AAPL", from, to)
		if err == nil {
			t.Fatal("expected error for rate limit response")
		}
		if !errors.Is(err, httpclient.ErrRateLimited) {
			t.Errorf("error should wrap ErrRateLimited, got: %v", err)
		}
	})

	t.Run("rate limit empty body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		_, err := client.fetchPrices(context.Background(), "AAPL", from, to)
		if err == nil {
			t.Fatal("expected error for empty body")
		}
		if !errors.Is(err, httpclient.ErrRateLimited) {
			t.Errorf("error should wrap ErrRateLimited, got: %v", err)
		}
	})

	t.Run("404 invalid ticker", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("<html>Not Found</html>"))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		_, err := client.fetchPrices(context.Background(), "INVALID", from, to)
		if err == nil {
			t.Fatal("expected error for 404 response")
		}
		if !errors.Is(err, ErrTickerInvalid) {
			t.Errorf("error should wrap ErrTickerInvalid, got: %v", err)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		prices, err := client.fetchPrices(context.Background(), "AAPL", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("len(prices) = %d, want 0", len(prices))
		}
	})
}

func TestToDailyPrice(t *testing.T) {
	t.Run("converts to domain type", func(t *testing.T) {
		raw := tiingoPrice{
			AdjClose:    150.25,
			Close:       150.25,
			Date:        "2024-01-15T00:00:00+00:00",
			DivCash:     0.0,
			High:        151.0,
			Low:         149.5,
			Open:        150.0,
			SplitFactor: 1.0,
			Volume:      55000000,
		}

		p, err := toDailyPrice(raw, "AAPL")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.Symbol != "AAPL" {
			t.Errorf("Symbol = %q, want AAPL", p.Symbol)
		}
		if p.Source != "tiingo" {
			t.Errorf("Source = %q, want tiingo", p.Source)
		}
		if p.AdjClose != 150.25 {
			t.Errorf("AdjClose = %v, want 150.25", p.AdjClose)
		}
		if p.Volume != 55000000 {
			t.Errorf("Volume = %d, want 55000000", p.Volume)
		}

		expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		if !p.Date.Equal(expectedDate) {
			t.Errorf("Date = %v, want %v", p.Date, expectedDate)
		}
	})

	t.Run("parses fractional second date format", func(t *testing.T) {
		raw := tiingoPrice{
			AdjClose:    100.0,
			Close:       100.0,
			Date:        "2024-06-01T00:00:00.000Z",
			DivCash:     0.0,
			High:        101.0,
			Low:         99.0,
			Open:        100.0,
			SplitFactor: 1.0,
			Volume:      1000000,
		}

		p, err := toDailyPrice(raw, "TEST")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		if !p.Date.Equal(expectedDate) {
			t.Errorf("Date = %v, want %v", p.Date, expectedDate)
		}
	})

	t.Run("rejects invalid date format", func(t *testing.T) {
		raw := tiingoPrice{Date: "not-a-date"}

		_, err := toDailyPrice(raw, "AAPL")
		if err == nil {
			t.Fatal("expected error for invalid date")
		}
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			"rate limited sentinel",
			fmt.Errorf("wrapped: %w", httpclient.ErrRateLimited),
			true,
		},
		{
			"retryable API error 500",
			&httpclient.APIError{IsRetryable: true, StatusCode: 500, URL: "/test"},
			true,
		},
		{
			"retryable API error 429",
			&httpclient.APIError{IsRetryable: true, StatusCode: 429, URL: "/test"},
			true,
		},
		{
			"non-retryable API error 404",
			&httpclient.APIError{IsRetryable: false, StatusCode: 404, URL: "/test"},
			false,
		},
		{
			"ticker invalid",
			ErrTickerInvalid,
			false,
		},
		{
			"generic error",
			errors.New("something broke"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.want {
				t.Errorf("IsRetryable = %v, want %v", got, tt.want)
			}
		})
	}
}
