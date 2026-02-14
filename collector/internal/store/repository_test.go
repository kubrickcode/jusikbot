package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/store"
)

func setupRepository(t *testing.T) *store.Repository {
	t.Helper()
	pool := connectAndClean(t)
	t.Cleanup(pool.Close)
	ctx := context.Background()

	if err := store.RunMigrations(ctx, pool); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return store.NewRepository(pool)
}

func TestUpsertPrices(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	t.Run("inserts new rows", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{
				AdjClose: 496.30, Close: 496.30,
				Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				High: 498.50, Low: 492.10, Open: 495.22,
				Source: "tiingo", Symbol: "NVDA", Volume: 40000000,
			},
			{
				AdjClose: 497.00, Close: 497.00,
				Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				High: 500.00, Low: 494.00, Open: 496.50,
				Source: "tiingo", Symbol: "NVDA", Volume: 35000000,
			},
		}
		affected, err := repo.UpsertPrices(ctx, prices)
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}
		if affected != 2 {
			t.Errorf("rows affected = %d, want 2", affected)
		}
	})

	t.Run("updates on duplicate key", func(t *testing.T) {
		seed := []domain.DailyPrice{
			{
				AdjClose: 496.30, Close: 496.30,
				Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				High: 498.50, Low: 492.10, Open: 495.22,
				Source: "tiingo", Symbol: "NVDA", Volume: 40000000,
			},
		}
		if _, err := repo.UpsertPrices(ctx, seed); err != nil {
			t.Fatalf("seed: %v", err)
		}

		updated := []domain.DailyPrice{
			{
				AdjClose: 505.00, Close: 505.00,
				Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				High: 510.00, Low: 495.00, Open: 500.00,
				Source: "tiingo", Symbol: "NVDA", Volume: 45000000,
			},
		}
		affected, err := repo.UpsertPrices(ctx, updated)
		if err != nil {
			t.Fatalf("upsert update: %v", err)
		}
		if affected != 1 {
			t.Errorf("rows affected = %d, want 1", affected)
		}

		history, err := repo.FetchPriceHistory(ctx, "NVDA",
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(history) != 1 {
			t.Fatalf("history len = %d, want 1", len(history))
		}
		if history[0].AdjClose != 505.00 {
			t.Errorf("adj_close = %f, want 505.00", history[0].AdjClose)
		}
	})

	t.Run("empty slice returns zero", func(t *testing.T) {
		affected, err := repo.UpsertPrices(ctx, nil)
		if err != nil {
			t.Fatalf("upsert empty: %v", err)
		}
		if affected != 0 {
			t.Errorf("rows affected = %d, want 0", affected)
		}
	})
}

func TestUpsertPrices_CheckViolation(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	tests := []struct {
		name   string
		prices []domain.DailyPrice
	}{
		{
			name: "high less than low rejects",
			prices: []domain.DailyPrice{
				{
					AdjClose: 95.00, Close: 95.00,
					Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					High: 80.00, Low: 90.00, Open: 100.00,
					Source: "test", Symbol: "BAD", Volume: 1000,
				},
			},
		},
		{
			name: "negative volume rejects",
			prices: []domain.DailyPrice{
				{
					AdjClose: 95.00, Close: 95.00,
					Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
					High: 110.00, Low: 90.00, Open: 100.00,
					Source: "test", Symbol: "BAD", Volume: -1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.UpsertPrices(ctx, tt.prices)
			if err == nil {
				t.Error("expected CHECK constraint violation, got nil")
			}
		})
	}
}

func TestUpsertFXRates(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	t.Run("inserts new rows", func(t *testing.T) {
		rates := []domain.FXRate{
			{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1305.50, Source: "frankfurter"},
			{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1310.25, Source: "frankfurter"},
		}
		affected, err := repo.UpsertFXRates(ctx, rates)
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}
		if affected != 2 {
			t.Errorf("rows affected = %d, want 2", affected)
		}
	})

	t.Run("updates on duplicate key", func(t *testing.T) {
		seed := []domain.FXRate{
			{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1305.50, Source: "frankfurter"},
		}
		if _, err := repo.UpsertFXRates(ctx, seed); err != nil {
			t.Fatalf("seed: %v", err)
		}

		updated := []domain.FXRate{
			{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1300.00, Source: "frankfurter"},
		}
		affected, err := repo.UpsertFXRates(ctx, updated)
		if err != nil {
			t.Fatalf("upsert update: %v", err)
		}
		if affected != 1 {
			t.Errorf("rows affected = %d, want 1", affected)
		}

		fetched, err := repo.FetchFXRates(ctx, "USD/KRW",
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(fetched) != 1 {
			t.Fatalf("rates len = %d, want 1", len(fetched))
		}
		if fetched[0].Rate != 1300.00 {
			t.Errorf("rate = %f, want 1300.00", fetched[0].Rate)
		}
	})

	t.Run("empty slice returns zero", func(t *testing.T) {
		affected, err := repo.UpsertFXRates(ctx, nil)
		if err != nil {
			t.Fatalf("upsert empty: %v", err)
		}
		if affected != 0 {
			t.Errorf("rows affected = %d, want 0", affected)
		}
	})

	t.Run("negative rate rejects", func(t *testing.T) {
		bad := []domain.FXRate{
			{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: -1.0, Source: "test"},
		}
		_, err := repo.UpsertFXRates(ctx, bad)
		if err == nil {
			t.Error("expected CHECK constraint violation, got nil")
		}
	})
}

func TestDetectGaps(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	prices := []domain.DailyPrice{
		{AdjClose: 100, Close: 100, Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), High: 110, Low: 90, Open: 100, Source: "tiingo", Symbol: "NVDA", Volume: 1000},
		{AdjClose: 105, Close: 105, Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), High: 110, Low: 100, Open: 103, Source: "tiingo", Symbol: "NVDA", Volume: 2000},
		{AdjClose: 200, Close: 200, Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), High: 210, Low: 190, Open: 200, Source: "tiingo", Symbol: "META", Volume: 3000},
	}
	if _, err := repo.UpsertPrices(ctx, prices); err != nil {
		t.Fatalf("setup prices: %v", err)
	}

	t.Run("returns last date per symbol", func(t *testing.T) {
		gaps, err := repo.DetectGaps(ctx, []string{"NVDA", "META", "ASML"})
		if err != nil {
			t.Fatalf("detect gaps: %v", err)
		}

		nvdaDate, ok := gaps["NVDA"]
		if !ok {
			t.Fatal("NVDA not found in gaps")
		}
		wantNVDA := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
		if !nvdaDate.Equal(wantNVDA) {
			t.Errorf("NVDA last date = %v, want %v", nvdaDate, wantNVDA)
		}

		metaDate, ok := gaps["META"]
		if !ok {
			t.Fatal("META not found in gaps")
		}
		wantMETA := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		if !metaDate.Equal(wantMETA) {
			t.Errorf("META last date = %v, want %v", metaDate, wantMETA)
		}

		if _, ok := gaps["ASML"]; ok {
			t.Error("ASML should not be in gaps (no data)")
		}
	})

	t.Run("empty symbols returns empty map", func(t *testing.T) {
		gaps, err := repo.DetectGaps(ctx, nil)
		if err != nil {
			t.Fatalf("detect gaps empty: %v", err)
		}
		if len(gaps) != 0 {
			t.Errorf("gaps len = %d, want 0", len(gaps))
		}
	})
}

func TestDetectFXGaps(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	rates := []domain.FXRate{
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1305.50, Source: "frankfurter"},
		{Date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1310.25, Source: "frankfurter"},
	}
	if _, err := repo.UpsertFXRates(ctx, rates); err != nil {
		t.Fatalf("setup fx rates: %v", err)
	}

	gaps, err := repo.DetectFXGaps(ctx, []string{"USD/KRW", "EUR/KRW"})
	if err != nil {
		t.Fatalf("detect fx gaps: %v", err)
	}

	usdDate, ok := gaps["USD/KRW"]
	if !ok {
		t.Fatal("USD/KRW not in gaps")
	}
	if !usdDate.Equal(time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("USD/KRW last date = %v, want 2024-01-05", usdDate)
	}

	if _, ok := gaps["EUR/KRW"]; ok {
		t.Error("EUR/KRW should not be in gaps (no data)")
	}
}

func TestFetchPriceHistory(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	prices := []domain.DailyPrice{
		{AdjClose: 100, Close: 100, Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), High: 110, Low: 90, Open: 100, Source: "tiingo", Symbol: "NVDA", Volume: 1000},
		{AdjClose: 105, Close: 105, Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), High: 110, Low: 100, Open: 103, Source: "tiingo", Symbol: "NVDA", Volume: 2000},
		{AdjClose: 110, Close: 110, Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), High: 115, Low: 105, Open: 107, Source: "tiingo", Symbol: "NVDA", Volume: 3000},
		{AdjClose: 200, Close: 200, Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), High: 210, Low: 190, Open: 200, Source: "tiingo", Symbol: "META", Volume: 5000},
	}
	if _, err := repo.UpsertPrices(ctx, prices); err != nil {
		t.Fatalf("setup prices: %v", err)
	}

	t.Run("filters by symbol and date range", func(t *testing.T) {
		history, err := repo.FetchPriceHistory(ctx, "NVDA",
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(history) != 2 {
			t.Fatalf("history len = %d, want 2", len(history))
		}
		if history[0].Date.After(history[1].Date) {
			t.Error("expected ascending date order")
		}
	})

	t.Run("excludes other symbols", func(t *testing.T) {
		history, err := repo.FetchPriceHistory(ctx, "NVDA",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		for _, p := range history {
			if p.Symbol != "NVDA" {
				t.Errorf("unexpected symbol %s in NVDA results", p.Symbol)
			}
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		history, err := repo.FetchPriceHistory(ctx, "ASML",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(history) != 0 {
			t.Errorf("history len = %d, want 0", len(history))
		}
	})

	t.Run("populates all fields including fetched_at", func(t *testing.T) {
		history, err := repo.FetchPriceHistory(ctx, "NVDA",
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(history) != 1 {
			t.Fatalf("history len = %d, want 1", len(history))
		}
		p := history[0]
		if p.FetchedAt.IsZero() {
			t.Error("fetched_at should not be zero")
		}
		if p.Source != "tiingo" {
			t.Errorf("source = %s, want tiingo", p.Source)
		}
	})
}

func TestFetchFXRates(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	rates := []domain.FXRate{
		{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1305.50, Source: "frankfurter"},
		{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1310.25, Source: "frankfurter"},
		{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1315.00, Source: "frankfurter"},
	}
	if _, err := repo.UpsertFXRates(ctx, rates); err != nil {
		t.Fatalf("setup fx rates: %v", err)
	}

	t.Run("filters by pair and date range", func(t *testing.T) {
		fetched, err := repo.FetchFXRates(ctx, "USD/KRW",
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(fetched) != 2 {
			t.Fatalf("rates len = %d, want 2", len(fetched))
		}
		if fetched[0].Date.After(fetched[1].Date) {
			t.Error("expected ascending date order")
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		fetched, err := repo.FetchFXRates(ctx, "EUR/KRW",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(fetched) != 0 {
			t.Errorf("rates len = %d, want 0", len(fetched))
		}
	})

	t.Run("populates all fields including fetched_at", func(t *testing.T) {
		fetched, err := repo.FetchFXRates(ctx, "USD/KRW",
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if len(fetched) != 1 {
			t.Fatalf("rates len = %d, want 1", len(fetched))
		}
		r := fetched[0]
		if r.FetchedAt.IsZero() {
			t.Error("fetched_at should not be zero")
		}
		if r.Source != "frankfurter" {
			t.Errorf("source = %s, want frankfurter", r.Source)
		}
	})
}
