package summary

import (
	"math"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
)

var baseDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func makePrice(date time.Time, adjClose float64, volume int64) domain.DailyPrice {
	return domain.DailyPrice{
		AdjClose: adjClose,
		Close:    adjClose,
		Date:     date,
		High:     adjClose * 1.02,
		Low:      adjClose * 0.98,
		Open:     adjClose,
		Source:   "test",
		Symbol:   "TEST",
		Volume:   volume,
	}
}

func makePriceAnomaly(date time.Time, adjClose float64, volume int64) domain.DailyPrice {
	p := makePrice(date, adjClose, volume)
	p.IsAnomaly = true
	return p
}

func makePriceSeries(startDate time.Time, adjCloses []float64, volume int64) []domain.DailyPrice {
	prices := make([]domain.DailyPrice, len(adjCloses))
	for i, ac := range adjCloses {
		prices[i] = makePrice(startDate.AddDate(0, 0, i), ac, volume)
	}
	return prices
}

func repeatFloat(val float64, n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = val
	}
	return s
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func assertNil(t *testing.T, got *float64, label string) {
	t.Helper()
	if got != nil {
		t.Errorf("%s = %v, want nil", label, *got)
	}
}

func assertAlmostEqual(t *testing.T, got *float64, want float64, tolerance float64, label string) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s = nil, want %v", label, want)
	}
	if !almostEqual(*got, want, tolerance) {
		t.Errorf("%s = %v, want %v (±%v)", label, *got, want, tolerance)
	}
}

func ptrFloat(v float64) *float64 { return &v }

func TestFiftyTwoWeekHigh(t *testing.T) {
	t.Run("returns max adj_close", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 120, 110, 115, 105}, 1000)

		got := FiftyTwoWeekHigh(prices)
		assertAlmostEqual(t, got, 120, 0.001, "FiftyTwoWeekHigh")
	})

	t.Run("excludes anomaly entries", func(t *testing.T) {
		prices := []domain.DailyPrice{
			makePrice(baseDate, 100, 1000),
			makePriceAnomaly(baseDate.AddDate(0, 0, 1), 999, 1000),
			makePrice(baseDate.AddDate(0, 0, 2), 110, 1000),
		}

		got := FiftyTwoWeekHigh(prices)
		assertAlmostEqual(t, got, 110, 0.001, "FiftyTwoWeekHigh")
	})

	t.Run("empty prices returns nil", func(t *testing.T) {
		assertNil(t, FiftyTwoWeekHigh(nil), "FiftyTwoWeekHigh")
	})

	t.Run("all anomalies returns nil", func(t *testing.T) {
		prices := []domain.DailyPrice{
			makePriceAnomaly(baseDate, 100, 1000),
		}
		assertNil(t, FiftyTwoWeekHigh(prices), "FiftyTwoWeekHigh")
	})
}

func TestFiftyTwoWeekLow(t *testing.T) {
	t.Run("returns min adj_close", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 120, 80, 115, 105}, 1000)

		got := FiftyTwoWeekLow(prices)
		assertAlmostEqual(t, got, 80, 0.001, "FiftyTwoWeekLow")
	})

	t.Run("excludes anomaly entries", func(t *testing.T) {
		prices := []domain.DailyPrice{
			makePrice(baseDate, 100, 1000),
			makePriceAnomaly(baseDate.AddDate(0, 0, 1), 1, 1000),
			makePrice(baseDate.AddDate(0, 0, 2), 90, 1000),
		}

		got := FiftyTwoWeekLow(prices)
		assertAlmostEqual(t, got, 90, 0.001, "FiftyTwoWeekLow")
	})

	t.Run("empty prices returns nil", func(t *testing.T) {
		assertNil(t, FiftyTwoWeekLow(nil), "FiftyTwoWeekLow")
	})
}

func TestFiftyTwoWeekPosition(t *testing.T) {
	tests := []struct {
		name    string
		current float64
		high    float64
		low     float64
		want    *float64
	}{
		{
			name:    "midpoint returns 0.5",
			current: 150,
			high:    200,
			low:     100,
			want:    ptrFloat(0.5),
		},
		{
			name:    "at high returns 1.0",
			current: 200,
			high:    200,
			low:     100,
			want:    ptrFloat(1.0),
		},
		{
			name:    "at low returns 0.0",
			current: 100,
			high:    200,
			low:     100,
			want:    ptrFloat(0.0),
		},
		{
			name:    "high equals low returns nil",
			current: 100,
			high:    100,
			low:     100,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FiftyTwoWeekPosition(tt.current, tt.high, tt.low)
			if tt.want == nil {
				assertNil(t, got, "FiftyTwoWeekPosition")
			} else {
				assertAlmostEqual(t, got, *tt.want, 0.001, "FiftyTwoWeekPosition")
			}
		})
	}
}

func TestMovingAverage(t *testing.T) {
	t.Run("exact N entries", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 110, 120}, 1000)

		got := MovingAverage(prices, 3)
		assertAlmostEqual(t, got, 110, 0.001, "MovingAverage(3)")
	})

	t.Run("more than N uses last N", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{50, 100, 110, 120}, 1000)

		got := MovingAverage(prices, 3)
		assertAlmostEqual(t, got, 110, 0.001, "MovingAverage(3)")
	})

	t.Run("insufficient data returns nil", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 110}, 1000)
		assertNil(t, MovingAverage(prices, 3), "MovingAverage(3)")
	})

	t.Run("excludes anomaly entries", func(t *testing.T) {
		prices := []domain.DailyPrice{
			makePrice(baseDate, 100, 1000),
			makePriceAnomaly(baseDate.AddDate(0, 0, 1), 999, 1000),
			makePrice(baseDate.AddDate(0, 0, 2), 110, 1000),
			makePrice(baseDate.AddDate(0, 0, 3), 120, 1000),
		}

		got := MovingAverage(prices, 3)
		assertAlmostEqual(t, got, 110, 0.001, "MovingAverage(3)")
	})
}

func TestMADivergence(t *testing.T) {
	t.Run("positive divergence above MA", func(t *testing.T) {
		ma := 100.0
		got := MADivergence(110, &ma)
		assertAlmostEqual(t, got, 10.0, 0.001, "MADivergence")
	})

	t.Run("negative divergence below MA", func(t *testing.T) {
		ma := 100.0
		got := MADivergence(90, &ma)
		assertAlmostEqual(t, got, -10.0, 0.001, "MADivergence")
	})

	t.Run("nil MA returns nil", func(t *testing.T) {
		assertNil(t, MADivergence(100, nil), "MADivergence")
	})

	t.Run("zero MA returns nil", func(t *testing.T) {
		zero := 0.0
		assertNil(t, MADivergence(100, &zero), "MADivergence")
	})
}

func TestDetectMACross(t *testing.T) {
	t.Run("golden cross within lookback", func(t *testing.T) {
		// 200 entries at 95, then 5 entries at 200 → 50D crosses above 200D
		adjCloses := append(repeatFloat(95, 200), repeatFloat(200, 5)...)
		prices := makePriceSeries(baseDate, adjCloses, 1000)

		got := DetectMACross(prices, 5)
		if got == nil {
			t.Fatal("DetectMACross = nil, want GC")
		}
		if *got != "GC" {
			t.Errorf("DetectMACross = %q, want %q", *got, "GC")
		}
	})

	t.Run("dead cross within lookback", func(t *testing.T) {
		// 200 entries at 105, then 5 entries at 10 → 50D crosses below 200D
		adjCloses := append(repeatFloat(105, 200), repeatFloat(10, 5)...)
		prices := makePriceSeries(baseDate, adjCloses, 1000)

		got := DetectMACross(prices, 5)
		if got == nil {
			t.Fatal("DetectMACross = nil, want DC")
		}
		if *got != "DC" {
			t.Errorf("DetectMACross = %q, want %q", *got, "DC")
		}
	})

	t.Run("no cross returns nil", func(t *testing.T) {
		adjCloses := repeatFloat(100, 210)
		prices := makePriceSeries(baseDate, adjCloses, 1000)

		got := DetectMACross(prices, 5)
		if got != nil {
			t.Errorf("DetectMACross = %q, want nil", *got)
		}
	})

	t.Run("insufficient data returns nil", func(t *testing.T) {
		prices := makePriceSeries(baseDate, repeatFloat(100, 50), 1000)

		got := DetectMACross(prices, 5)
		if got != nil {
			t.Errorf("DetectMACross = %q, want nil", *got)
		}
	})
}

func TestPriceChange(t *testing.T) {
	t.Run("positive 5D change", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 101, 102, 103, 104, 110}, 1000)

		got := PriceChange(prices, 5)
		assertAlmostEqual(t, got, 10.0, 0.001, "PriceChange(5)")
	})

	t.Run("negative change", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 99, 98, 97, 96, 90}, 1000)

		got := PriceChange(prices, 5)
		assertAlmostEqual(t, got, -10.0, 0.001, "PriceChange(5)")
	})

	t.Run("insufficient data returns nil", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 110}, 1000)
		assertNil(t, PriceChange(prices, 5), "PriceChange(5)")
	})

	t.Run("excludes anomaly entries", func(t *testing.T) {
		prices := []domain.DailyPrice{
			makePrice(baseDate, 100, 1000),
			makePriceAnomaly(baseDate.AddDate(0, 0, 1), 999, 1000),
			makePrice(baseDate.AddDate(0, 0, 2), 105, 1000),
		}

		got := PriceChange(prices, 1)
		assertAlmostEqual(t, got, 5.0, 0.001, "PriceChange(1)")
	})
}

func TestRelativeBenchmark(t *testing.T) {
	t.Run("normal computation", func(t *testing.T) {
		stock := 10.0
		bench := 3.0
		got := RelativeBenchmark(&stock, &bench)
		assertAlmostEqual(t, got, 7.0, 0.001, "RelativeBenchmark")
	})

	t.Run("nil stock returns nil", func(t *testing.T) {
		bench := 3.0
		assertNil(t, RelativeBenchmark(nil, &bench), "RelativeBenchmark")
	})

	t.Run("nil bench returns nil", func(t *testing.T) {
		stock := 10.0
		assertNil(t, RelativeBenchmark(&stock, nil), "RelativeBenchmark")
	})
}

func TestVolumeRatio(t *testing.T) {
	t.Run("current vs trailing average", func(t *testing.T) {
		// 20 days at volume 100, last day (21st) at volume 200
		adjCloses := repeatFloat(100, 21)
		prices := makePriceSeries(baseDate, adjCloses, 100)
		prices[20].Volume = 200

		got := VolumeRatio(prices, 20)
		// trailing avg = 20 * 100 / 20 = 100
		// ratio = 200 / 100 = 2.0
		assertAlmostEqual(t, got, 2.0, 0.001, "VolumeRatio")
	})

	t.Run("insufficient data returns nil", func(t *testing.T) {
		prices := makePriceSeries(baseDate, repeatFloat(100, 5), 1000)
		assertNil(t, VolumeRatio(prices, 20), "VolumeRatio")
	})

	t.Run("zero trailing average returns nil", func(t *testing.T) {
		// 21 entries: trailing 20 have volume 0, current has volume 100
		prices := makePriceSeries(baseDate, repeatFloat(100, 21), 0)
		prices[20].Volume = 100
		assertNil(t, VolumeRatio(prices, 20), "VolumeRatio")
	})

	t.Run("excludes anomaly entries", func(t *testing.T) {
		// 5 non-anomaly + 1 anomaly → 5 clean entries, need 5+1 for VolumeRatio(4)
		prices := makePriceSeries(baseDate, repeatFloat(100, 5), 100)
		prices = append(prices, makePriceAnomaly(baseDate.AddDate(0, 0, 5), 100, 9999))

		got := VolumeRatio(prices, 4)
		// trailing 4 = [100,100,100,100], avg=100, current=100, ratio=1.0
		assertAlmostEqual(t, got, 1.0, 0.001, "VolumeRatio")
	})
}

func TestHistoricalVolatility(t *testing.T) {
	t.Run("known computation", func(t *testing.T) {
		// 5 prices → 4 log returns
		prices := makePriceSeries(baseDate, []float64{100, 105, 103, 108, 110}, 1000)

		got := HistoricalVolatility(prices, 4)

		// Manually computed:
		// returns: ln(1.05), ln(103/105), ln(108/103), ln(110/108)
		// = 0.04879, -0.01923, 0.04738, 0.01835
		// mean = 0.02382
		// sample variance (N-1=3) = 0.001021
		// annualized = sqrt(0.001021) * sqrt(252) * 100 ≈ 50.73
		assertAlmostEqual(t, got, 50.73, 0.5, "HistoricalVolatility(4)")
	})

	t.Run("constant prices returns zero", func(t *testing.T) {
		prices := makePriceSeries(baseDate, repeatFloat(100, 25), 1000)

		got := HistoricalVolatility(prices, 20)
		assertAlmostEqual(t, got, 0.0, 0.001, "HistoricalVolatility(20)")
	})

	t.Run("insufficient data returns nil", func(t *testing.T) {
		prices := makePriceSeries(baseDate, []float64{100, 105}, 1000)
		assertNil(t, HistoricalVolatility(prices, 20), "HistoricalVolatility(20)")
	})

	t.Run("excludes anomaly entries", func(t *testing.T) {
		prices := []domain.DailyPrice{
			makePrice(baseDate, 100, 1000),
			makePriceAnomaly(baseDate.AddDate(0, 0, 1), 999, 1000),
			makePrice(baseDate.AddDate(0, 0, 2), 100, 1000),
			makePrice(baseDate.AddDate(0, 0, 3), 100, 1000),
		}

		// After filtering: [100, 100, 100] → 2 returns, both 0
		got := HistoricalVolatility(prices, 2)
		assertAlmostEqual(t, got, 0.0, 0.001, "HistoricalVolatility(2)")
	})
}
