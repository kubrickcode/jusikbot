package tiingo

import (
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
)

func TestComputeStartDate(t *testing.T) {
	to := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("no gap uses lookback", func(t *testing.T) {
		gaps := map[string]time.Time{}

		got := computeStartDate(to, gaps, "AAPL")
		want := to.AddDate(0, 0, -defaultLookbackDays)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gap newer than lookback uses gap+1", func(t *testing.T) {
		lastDate := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		gaps := map[string]time.Time{"AAPL": lastDate}

		got := computeStartDate(to, gaps, "AAPL")
		want := lastDate.AddDate(0, 0, 1)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gap older than lookback uses lookback", func(t *testing.T) {
		lastDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		gaps := map[string]time.Time{"AAPL": lastDate}

		got := computeStartDate(to, gaps, "AAPL")
		want := to.AddDate(0, 0, -defaultLookbackDays)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gap for different symbol ignored", func(t *testing.T) {
		gaps := map[string]time.Time{"META": time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC)}

		got := computeStartDate(to, gaps, "AAPL")
		want := to.AddDate(0, 0, -defaultLookbackDays)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestMarkAnomalies(t *testing.T) {
	entry := domain.WatchlistEntry{
		Market: domain.MarketUS,
		Name:   "Test Stock",
		Symbol: "TEST",
		Type:   domain.SecurityTypeStock,
	}

	t.Run("normal prices have no anomaly", func(t *testing.T) {
		raw := []tiingoPrice{
			{AdjClose: 100.0, Close: 100.0, Date: "2024-01-15T00:00:00+00:00", High: 101.0, Low: 99.0, Open: 100.0, SplitFactor: 1.0, Volume: 1000},
			{AdjClose: 102.0, Close: 102.0, Date: "2024-01-16T00:00:00+00:00", High: 103.0, Low: 101.0, Open: 101.0, SplitFactor: 1.0, Volume: 1100},
		}

		prices, err := markAnomalies(raw, entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 2 {
			t.Fatalf("len(prices) = %d, want 2", len(prices))
		}

		for i, p := range prices {
			if p.IsAnomaly {
				t.Errorf("prices[%d].IsAnomaly = true, want false", i)
			}
		}
	})

	t.Run("large drop without corporate action flagged", func(t *testing.T) {
		raw := []tiingoPrice{
			{AdjClose: 100.0, Close: 100.0, Date: "2024-01-15T00:00:00+00:00", High: 101.0, Low: 99.0, Open: 100.0, SplitFactor: 1.0, Volume: 1000},
			{AdjClose: 30.0, Close: 30.0, Date: "2024-01-16T00:00:00+00:00", High: 31.0, Low: 29.0, Open: 30.0, SplitFactor: 1.0, DivCash: 0.0, Volume: 5000},
		}

		prices, err := markAnomalies(raw, entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !prices[1].IsAnomaly {
			t.Error("prices[1].IsAnomaly = false, want true for 70% drop")
		}
	})

	t.Run("large drop with stock split not flagged", func(t *testing.T) {
		raw := []tiingoPrice{
			{AdjClose: 100.0, Close: 100.0, Date: "2024-01-15T00:00:00+00:00", High: 101.0, Low: 99.0, Open: 100.0, SplitFactor: 1.0, Volume: 1000},
			{AdjClose: 30.0, Close: 30.0, Date: "2024-01-16T00:00:00+00:00", High: 31.0, Low: 29.0, Open: 30.0, SplitFactor: 3.0, DivCash: 0.0, Volume: 5000},
		}

		prices, err := markAnomalies(raw, entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prices[1].IsAnomaly {
			t.Error("prices[1].IsAnomaly = true, want false for stock split")
		}
	})

	t.Run("large drop with dividend not flagged", func(t *testing.T) {
		raw := []tiingoPrice{
			{AdjClose: 100.0, Close: 100.0, Date: "2024-01-15T00:00:00+00:00", High: 101.0, Low: 99.0, Open: 100.0, SplitFactor: 1.0, Volume: 1000},
			{AdjClose: 30.0, Close: 30.0, Date: "2024-01-16T00:00:00+00:00", High: 31.0, Low: 29.0, Open: 30.0, SplitFactor: 1.0, DivCash: 5.0, Volume: 5000},
		}

		prices, err := markAnomalies(raw, entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prices[1].IsAnomaly {
			t.Error("prices[1].IsAnomaly = true, want false for dividend")
		}
	})

	t.Run("first row never anomaly", func(t *testing.T) {
		raw := []tiingoPrice{
			{AdjClose: 100.0, Close: 100.0, Date: "2024-01-15T00:00:00+00:00", High: 101.0, Low: 99.0, Open: 100.0, SplitFactor: 1.0, Volume: 1000},
		}

		prices, err := markAnomalies(raw, entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prices[0].IsAnomaly {
			t.Error("first row should never be anomaly")
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		prices, err := markAnomalies([]tiingoPrice{}, entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("len(prices) = %d, want 0", len(prices))
		}
	})

	t.Run("invalid date returns error", func(t *testing.T) {
		raw := []tiingoPrice{
			{AdjClose: 100.0, Close: 100.0, Date: "bad-date", SplitFactor: 1.0},
		}

		_, err := markAnomalies(raw, entry)
		if err == nil {
			t.Fatal("expected error for invalid date")
		}
	})
}

func TestIsConfirmedAnomaly(t *testing.T) {
	entry := domain.WatchlistEntry{
		Market: domain.MarketUS,
		Symbol: "TEST",
		Type:   domain.SecurityTypeStock,
	}

	tests := []struct {
		name         string
		adjClose     float64
		prevAdjClose float64
		splitFactor  float64
		divCash      float64
		want         bool
	}{
		{"normal change", 105.0, 100.0, 1.0, 0.0, false},
		{"anomaly no corporate action", 30.0, 100.0, 1.0, 0.0, true},
		{"anomaly with split", 30.0, 100.0, 3.0, 0.0, false},
		{"anomaly with dividend", 30.0, 100.0, 1.0, 5.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tiingoPrice{AdjClose: tt.adjClose, SplitFactor: tt.splitFactor, DivCash: tt.divCash}
			got := isConfirmedAnomaly(r, tt.prevAdjClose, entry)
			if got != tt.want {
				t.Errorf("isConfirmedAnomaly = %v, want %v", got, tt.want)
			}
		})
	}
}
