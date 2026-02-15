package summary

import (
	"testing"

	"github.com/jusikbot/collector/internal/domain"
)

func TestComputeSymbolIndicators(t *testing.T) {
	t.Run("sufficient data returns all indicators", func(t *testing.T) {
		// 210 entries: enough for 200D MA
		adjCloses := make([]float64, 210)
		for i := range adjCloses {
			adjCloses[i] = 100 + float64(i)*0.1
		}
		prices := makePriceSeries(baseDate, adjCloses, 1000)
		benchPrices := makePriceSeries(baseDate, adjCloses, 1000)

		result := ComputeSymbolIndicators(prices, benchPrices, false)

		if result.AdjClose == 0 {
			t.Error("AdjClose should not be zero")
		}
		if result.FiftyTwoWeekHigh == nil {
			t.Error("FiftyTwoWeekHigh should not be nil")
		}
		if result.FiftyTwoWeekLow == nil {
			t.Error("FiftyTwoWeekLow should not be nil")
		}
		if result.FiftyTwoWeekPos == nil {
			t.Error("FiftyTwoWeekPos should not be nil")
		}
		if result.MADivergence50D == nil {
			t.Error("MADivergence50D should not be nil")
		}
		if result.MADivergence200D == nil {
			t.Error("MADivergence200D should not be nil")
		}
		if result.Change5D == nil {
			t.Error("Change5D should not be nil")
		}
		if result.Change20D == nil {
			t.Error("Change20D should not be nil")
		}
		if result.VolRatio == nil {
			t.Error("VolRatio should not be nil")
		}
		if result.HV20D == nil {
			t.Error("HV20D should not be nil")
		}
		if result.HV60D == nil {
			t.Error("HV60D should not be nil")
		}
	})

	t.Run("benchmark symbol has nil RelativeBench", func(t *testing.T) {
		prices := makePriceSeries(baseDate, repeatFloat(100, 30), 1000)
		benchPrices := makePriceSeries(baseDate, repeatFloat(100, 30), 1000)

		result := ComputeSymbolIndicators(prices, benchPrices, true)

		if result.RelativeBench20D != nil {
			t.Errorf("RelativeBench20D = %v, want nil for benchmark", *result.RelativeBench20D)
		}
	})

	t.Run("non-benchmark computes RelativeBench", func(t *testing.T) {
		// Stock: 100 → 110 (10% change over 20 days)
		stockCloses := make([]float64, 25)
		for i := range stockCloses {
			stockCloses[i] = 100
		}
		stockCloses[24] = 110
		stockPrices := makePriceSeries(baseDate, stockCloses, 1000)

		// Bench: constant 100 (0% change)
		benchPrices := makePriceSeries(baseDate, repeatFloat(100, 25), 1000)

		result := ComputeSymbolIndicators(stockPrices, benchPrices, false)

		if result.RelativeBench20D == nil {
			t.Fatal("RelativeBench20D should not be nil")
		}
		if *result.RelativeBench20D <= 0 {
			t.Errorf("RelativeBench20D = %v, want positive (stock outperformed)", *result.RelativeBench20D)
		}
	})

	t.Run("empty prices returns zero AdjClose", func(t *testing.T) {
		result := ComputeSymbolIndicators(nil, nil, false)

		if result.AdjClose != 0 {
			t.Errorf("AdjClose = %v, want 0", result.AdjClose)
		}
		if result.FiftyTwoWeekHigh != nil {
			t.Error("FiftyTwoWeekHigh should be nil for empty data")
		}
	})
}

func TestBenchmarkSymbols(t *testing.T) {
	if BenchmarkSymbols[domain.MarketUS] != "QQQ" {
		t.Errorf("US benchmark = %q, want %q", BenchmarkSymbols[domain.MarketUS], "QQQ")
	}
	if BenchmarkSymbols[domain.MarketKR] != "069500" {
		t.Errorf("KR benchmark = %q, want %q", BenchmarkSymbols[domain.MarketKR], "069500")
	}
}
