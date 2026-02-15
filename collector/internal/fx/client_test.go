package fx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/httpclient"
)

func newTestClient(srv *httptest.Server) *Client {
	hc := httpclient.NewClient(srv.URL, nil, srv.Client(), 0)
	return NewClient(hc)
}

func TestFetchRates(t *testing.T) {
	t.Run("normal JSON response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/2025-01-01..2025-01-10" {
				t.Errorf("path = %q, want /v1/2025-01-01..2025-01-10", r.URL.Path)
			}
			if got := r.URL.Query().Get("from"); got != "USD" {
				t.Errorf("from = %q, want USD", got)
			}
			if got := r.URL.Query().Get("to"); got != "KRW" {
				t.Errorf("to = %q, want KRW", got)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"amount": 1,
				"base": "USD",
				"start_date": "2025-01-01",
				"end_date": "2025-01-10",
				"rates": {
					"2025-01-02": {"KRW": 1466.73},
					"2025-01-03": {"KRW": 1470.50}
				}
			}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

		rates, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 2 {
			t.Fatalf("len(rates) = %d, want 2", len(rates))
		}

		if rates[0].Rate != 1466.73 {
			t.Errorf("rates[0].Rate = %v, want 1466.73", rates[0].Rate)
		}
		if rates[0].Pair != "USD/KRW" {
			t.Errorf("rates[0].Pair = %q, want USD/KRW", rates[0].Pair)
		}
		if rates[0].Source != "frankfurter" {
			t.Errorf("rates[0].Source = %q, want frankfurter", rates[0].Source)
		}

		expectedDate := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
		if !rates[0].Date.Equal(expectedDate) {
			t.Errorf("rates[0].Date = %v, want %v", rates[0].Date, expectedDate)
		}

		if rates[1].Rate != 1470.50 {
			t.Errorf("rates[1].Rate = %v, want 1470.50", rates[1].Rate)
		}
	})

	t.Run("empty rates object", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"amount": 1,
				"base": "USD",
				"start_date": "2025-01-01",
				"end_date": "2025-01-01",
				"rates": {}
			}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		rates, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 0 {
			t.Errorf("len(rates) = %d, want 0", len(rates))
		}
	})

	t.Run("server error 500", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

		_, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err == nil {
			t.Fatal("expected error for 500 response")
		}
		var apiErr *httpclient.APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected APIError, got: %T", err)
		}
		if apiErr.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
		}
	})

	t.Run("bad request 400", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"invalid date range"}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		_, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err == nil {
			t.Fatal("expected error for 400 response")
		}
	})

	t.Run("missing target currency in rates", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"amount": 1,
				"base": "USD",
				"start_date": "2025-01-01",
				"end_date": "2025-01-02",
				"rates": {
					"2025-01-02": {"EUR": 0.92}
				}
			}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

		_, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err == nil {
			t.Fatal("expected error when target currency missing from rates")
		}
	})

	t.Run("malformed JSON response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid json`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

		_, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err == nil {
			t.Fatal("expected error for malformed JSON")
		}
	})

	t.Run("rates sorted by date ascending", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"amount": 1,
				"base": "USD",
				"start_date": "2025-01-01",
				"end_date": "2025-01-05",
				"rates": {
					"2025-01-03": {"KRW": 1470.00},
					"2025-01-01": {"KRW": 1460.00},
					"2025-01-02": {"KRW": 1465.00}
				}
			}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)

		rates, err := client.FetchRates(context.Background(), "USD", "KRW", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 3 {
			t.Fatalf("len(rates) = %d, want 3", len(rates))
		}

		for i := 1; i < len(rates); i++ {
			if !rates[i].Date.After(rates[i-1].Date) {
				t.Errorf("rates not sorted: rates[%d].Date=%v <= rates[%d].Date=%v",
					i, rates[i].Date, i-1, rates[i-1].Date)
			}
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
			"non-retryable API error 400",
			&httpclient.APIError{IsRetryable: false, StatusCode: 400, URL: "/test"},
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
