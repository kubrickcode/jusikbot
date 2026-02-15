package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/ratelimit"
)

func TestCollectFX(t *testing.T) {
	t.Run("full collection without gaps", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"amount": 1,
				"base": "USD",
				"start_date": "2024-01-01",
				"end_date": "2024-12-31",
				"rates": {
					"2024-06-03": {"KRW": 1380.50},
					"2024-06-04": {"KRW": 1382.00}
				}
			}`))
		}))
		defer srv.Close()

		fxClient := newTestClient(srv)
		collector := NewCollector(fxClient, ratelimit.RetryConfig{
			InitialBackoff: time.Millisecond,
			MaxAttempts:    1,
			MaxBackoff:     time.Millisecond,
		})
		collector.now = func() time.Time {
			return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		gaps := make(map[string]time.Time)
		rates, err := collector.CollectFX(context.Background(), "USD", "KRW", gaps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 2 {
			t.Fatalf("len(rates) = %d, want 2", len(rates))
		}
	})

	t.Run("incremental collection with gap", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the path uses gap+1 day as start
			if r.URL.Path == "/v1/2025-01-11..2025-01-15" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"amount": 1,
					"base": "USD",
					"start_date": "2025-01-11",
					"end_date": "2025-01-15",
					"rates": {
						"2025-01-13": {"KRW": 1475.00}
					}
				}`))
				return
			}
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer srv.Close()

		fxClient := newTestClient(srv)
		collector := NewCollector(fxClient, ratelimit.RetryConfig{
			InitialBackoff: time.Millisecond,
			MaxAttempts:    1,
			MaxBackoff:     time.Millisecond,
		})

		// Simulate gap: last data was 2025-01-10, today is 2025-01-15
		collector.now = func() time.Time {
			return time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		}

		gaps := map[string]time.Time{
			"USD/KRW": time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		}

		rates, err := collector.CollectFX(context.Background(), "USD", "KRW", gaps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("len(rates) = %d, want 1", len(rates))
		}
		if rates[0].Rate != 1475.00 {
			t.Errorf("rates[0].Rate = %v, want 1475.00", rates[0].Rate)
		}
	})

	t.Run("already up to date", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("should not call API when already up to date")
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		fxClient := newTestClient(srv)
		collector := NewCollector(fxClient, ratelimit.RetryConfig{
			InitialBackoff: time.Millisecond,
			MaxAttempts:    1,
			MaxBackoff:     time.Millisecond,
		})

		today := time.Now().Truncate(24 * time.Hour)
		collector.now = func() time.Time { return today }

		gaps := map[string]time.Time{
			"USD/KRW": today,
		}

		rates, err := collector.CollectFX(context.Background(), "USD", "KRW", gaps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 0 {
			t.Errorf("len(rates) = %d, want 0", len(rates))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"amount":1,"base":"USD","rates":{}}`))
		}))
		defer srv.Close()

		fxClient := newTestClient(srv)
		collector := NewCollector(fxClient, ratelimit.RetryConfig{
			InitialBackoff: time.Millisecond,
			MaxAttempts:    1,
			MaxBackoff:     time.Millisecond,
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectFX(ctx, "USD", "KRW", make(map[string]time.Time))
		if err == nil {
			t.Fatal("expected error for cancelled context")
		}
	})

	t.Run("returns domain FXRate type", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"amount": 1,
				"base": "USD",
				"start_date": "2025-01-01",
				"end_date": "2025-01-03",
				"rates": {
					"2025-01-02": {"KRW": 1466.73}
				}
			}`))
		}))
		defer srv.Close()

		fxClient := newTestClient(srv)
		collector := NewCollector(fxClient, ratelimit.RetryConfig{
			InitialBackoff: time.Millisecond,
			MaxAttempts:    1,
			MaxBackoff:     time.Millisecond,
		})

		rates, err := collector.CollectFX(context.Background(), "USD", "KRW", make(map[string]time.Time))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 1 {
			t.Fatalf("len(rates) = %d, want 1", len(rates))
		}

		got := rates[0]
		want := domain.FXRate{
			Date:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			Pair:   "USD/KRW",
			Rate:   1466.73,
			Source: "frankfurter",
		}
		if got.Date != want.Date {
			t.Errorf("Date = %v, want %v", got.Date, want.Date)
		}
		if got.Pair != want.Pair {
			t.Errorf("Pair = %q, want %q", got.Pair, want.Pair)
		}
		if got.Rate != want.Rate {
			t.Errorf("Rate = %v, want %v", got.Rate, want.Rate)
		}
		if got.Source != want.Source {
			t.Errorf("Source = %q, want %q", got.Source, want.Source)
		}
	})
}
