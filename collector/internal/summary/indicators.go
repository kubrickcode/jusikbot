package summary

import (
	"math"
	"slices"

	"github.com/jusikbot/collector/internal/domain"
)

const (
	maCrossLookbackDays = 5
	maCrossShortDays    = 50
	maCrossLongDays     = 200
	tradingDaysPerYear  = 252
)

// FiftyTwoWeekHigh returns the highest adj_close among non-anomaly entries.
func FiftyTwoWeekHigh(prices []domain.DailyPrice) *float64 {
	adjCloses := extractAdjCloses(prices)
	if len(adjCloses) == 0 {
		return nil
	}
	high := slices.Max(adjCloses)
	return &high
}

// FiftyTwoWeekLow returns the lowest adj_close among non-anomaly entries.
func FiftyTwoWeekLow(prices []domain.DailyPrice) *float64 {
	adjCloses := extractAdjCloses(prices)
	if len(adjCloses) == 0 {
		return nil
	}
	low := slices.Min(adjCloses)
	return &low
}

// FiftyTwoWeekPosition returns (current - low) / (high - low).
// Returns nil when high equals low (division by zero).
func FiftyTwoWeekPosition(current, high, low float64) *float64 {
	if high == low {
		return nil
	}
	pos := (current - low) / (high - low)
	return &pos
}

// MovingAverage returns the simple moving average of adj_close for the last N
// non-anomaly entries. Returns nil if fewer than N entries exist.
func MovingAverage(prices []domain.DailyPrice, days int) *float64 {
	adjCloses := extractAdjCloses(prices)
	return movingAverageAt(adjCloses, len(adjCloses)-1, days)
}

// MADivergence returns (current - ma) / ma * 100 as a percentage.
// Returns nil when ma is nil or zero.
func MADivergence(current float64, ma *float64) *float64 {
	if ma == nil || *ma == 0 {
		return nil
	}
	div := (current - *ma) / *ma * 100
	return &div
}

// DetectMACross checks if 50D and 200D MAs crossed within the last lookbackDays
// trading days. Returns "GC" for golden cross, "DC" for dead cross, nil otherwise.
func DetectMACross(prices []domain.DailyPrice, lookbackDays int) *string {
	adjCloses := extractAdjCloses(prices)

	end := len(adjCloses) - 1
	if end < maCrossLongDays-1 {
		return nil
	}

	start := max(end-lookbackDays, maCrossLongDays-1)

	type maGap struct {
		diff  float64
		valid bool
	}

	gaps := make([]maGap, 0, lookbackDays+1)
	for i := start; i <= end; i++ {
		ma50 := movingAverageAt(adjCloses, i, maCrossShortDays)
		ma200 := movingAverageAt(adjCloses, i, maCrossLongDays)
		if ma50 == nil || ma200 == nil {
			gaps = append(gaps, maGap{valid: false})
			continue
		}
		gaps = append(gaps, maGap{diff: *ma50 - *ma200, valid: true})
	}

	var result *string
	for i := 1; i < len(gaps); i++ {
		if !gaps[i-1].valid || !gaps[i].valid {
			continue
		}
		if gaps[i-1].diff <= 0 && gaps[i].diff > 0 {
			gc := "GC"
			result = &gc
		} else if gaps[i-1].diff >= 0 && gaps[i].diff < 0 {
			dc := "DC"
			result = &dc
		}
	}

	return result
}

// PriceChange returns the percentage change of adj_close over the last N trading
// days (non-anomaly). Returns nil if fewer than N+1 non-anomaly entries exist.
func PriceChange(prices []domain.DailyPrice, days int) *float64 {
	adjCloses := extractAdjCloses(prices)
	if len(adjCloses) <= days {
		return nil
	}

	current := adjCloses[len(adjCloses)-1]
	past := adjCloses[len(adjCloses)-1-days]
	if past == 0 {
		return nil
	}

	change := (current - past) / past * 100
	return &change
}

// RelativeBenchmark returns stockChange - benchChange (arithmetic difference).
// Returns nil if either input is nil.
func RelativeBenchmark(stockChange, benchChange *float64) *float64 {
	if stockChange == nil || benchChange == nil {
		return nil
	}
	rel := *stockChange - *benchChange
	return &rel
}

// VolumeRatio returns the most recent volume divided by the trailing N-day average
// volume (non-anomaly, excluding the current day).
// Returns nil if fewer than N+1 entries or trailing average is zero.
func VolumeRatio(prices []domain.DailyPrice, days int) *float64 {
	volumes := extractVolumes(prices)
	if len(volumes) < days+1 {
		return nil
	}

	trailing := volumes[len(volumes)-1-days : len(volumes)-1]
	var sum int64
	for _, v := range trailing {
		sum += v
	}
	avg := float64(sum) / float64(days)
	if avg == 0 {
		return nil
	}

	current := float64(volumes[len(volumes)-1])
	ratio := current / avg
	return &ratio
}

// HistoricalVolatility returns annualized log-return volatility as a percentage.
// Uses sample standard deviation (N-1 denominator), subtracts mean, excludes anomalies.
// Returns nil if fewer than days+1 non-anomaly entries exist or days < 2.
func HistoricalVolatility(prices []domain.DailyPrice, days int) *float64 {
	adjCloses := extractAdjCloses(prices)
	if len(adjCloses) < days+1 || days < 2 {
		return nil
	}

	start := len(adjCloses) - days - 1
	returns := make([]float64, days)
	for i := range days {
		if adjCloses[start+i] == 0 {
			return nil
		}
		returns[i] = math.Log(adjCloses[start+i+1] / adjCloses[start+i])
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(days)

	var sumSqDev float64
	for _, r := range returns {
		dev := r - mean
		sumSqDev += dev * dev
	}
	variance := sumSqDev / float64(days-1)

	hv := math.Sqrt(variance) * math.Sqrt(tradingDaysPerYear) * 100
	return &hv
}

func extractAdjCloses(prices []domain.DailyPrice) []float64 {
	result := make([]float64, 0, len(prices))
	for _, p := range prices {
		if !p.IsAnomaly {
			result = append(result, p.AdjClose)
		}
	}
	return result
}

func extractVolumes(prices []domain.DailyPrice) []int64 {
	result := make([]int64, 0, len(prices))
	for _, p := range prices {
		if !p.IsAnomaly {
			result = append(result, p.Volume)
		}
	}
	return result
}

func movingAverageAt(adjCloses []float64, endIdx int, days int) *float64 {
	if endIdx < days-1 || endIdx < 0 || days <= 0 {
		return nil
	}

	start := endIdx - days + 1
	var sum float64
	for i := start; i <= endIdx; i++ {
		sum += adjCloses[i]
	}
	avg := sum / float64(days)
	return &avg
}
