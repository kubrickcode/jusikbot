package kis

import (
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
)

func TestComputeStartDate(t *testing.T) {
	to := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("no gap uses lookback", func(t *testing.T) {
		gaps := map[string]time.Time{}

		got := computeStartDate(to, gaps, "005930")
		want := to.AddDate(0, 0, -defaultLookbackDays)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gap newer than lookback uses gap+1", func(t *testing.T) {
		lastDate := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		gaps := map[string]time.Time{"005930": lastDate}

		got := computeStartDate(to, gaps, "005930")
		want := lastDate.AddDate(0, 0, 1)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gap older than lookback uses lookback", func(t *testing.T) {
		lastDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		gaps := map[string]time.Time{"005930": lastDate}

		got := computeStartDate(to, gaps, "005930")
		want := to.AddDate(0, 0, -defaultLookbackDays)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gap for different symbol ignored", func(t *testing.T) {
		gaps := map[string]time.Time{"035720": time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC)}

		got := computeStartDate(to, gaps, "005930")
		want := to.AddDate(0, 0, -defaultLookbackDays)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestMarkAnomalies(t *testing.T) {
	entry := domain.WatchlistEntry{
		Market: domain.MarketKR,
		Name:   "삼성전자",
		Symbol: "005930",
		Type:   domain.SecurityTypeStock,
	}

	t.Run("normal prices no anomaly", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{AdjClose: 70000, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{AdjClose: 71000, Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
			{AdjClose: 70500, Date: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)},
		}

		result := markAnomalies(prices, entry)
		for i, p := range result {
			if p.IsAnomaly {
				t.Errorf("prices[%d].IsAnomaly = true, want false", i)
			}
		}
	})

	t.Run("large drop flagged as anomaly", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{AdjClose: 70000, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{AdjClose: 40000, Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
		}

		result := markAnomalies(prices, entry)
		if !result[0].IsAnomaly {
			// First row should never be anomaly
		}
		if !result[1].IsAnomaly {
			t.Error("prices[1].IsAnomaly = false, want true for 42% drop (> KR 30% threshold)")
		}
	})

	t.Run("first row never anomaly", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{AdjClose: 70000, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}

		result := markAnomalies(prices, entry)
		if result[0].IsAnomaly {
			t.Error("first row should never be anomaly")
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		result := markAnomalies(nil, entry)
		if len(result) != 0 {
			t.Errorf("len = %d, want 0", len(result))
		}
	})

	t.Run("at threshold boundary not flagged", func(t *testing.T) {
		// 30% drop exactly: 70000 * 0.7 = 49000
		prices := []domain.DailyPrice{
			{AdjClose: 70000, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{AdjClose: 49000, Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
		}

		result := markAnomalies(prices, entry)
		if result[1].IsAnomaly {
			t.Error("exactly 30% change should not be flagged (threshold is >30%)")
		}
	})
}
