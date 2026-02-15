package summary

import "github.com/jusikbot/collector/internal/domain"

// BenchmarkSymbols maps markets to their benchmark symbol.
// US → QQQ (Nasdaq 100), KR → 069500 (KODEX 200).
var BenchmarkSymbols = map[domain.Market]string{
	domain.MarketUS: "QQQ",
	domain.MarketKR: "069500",
}

// SymbolIndicators holds the indicator set for one symbol (14 columns in summary:
// AdjClose + 12 optional metrics + MACross categorical signal).
type SymbolIndicators struct {
	AdjClose         float64
	Change5D         *float64
	Change20D        *float64
	FiftyTwoWeekHigh *float64
	FiftyTwoWeekLow  *float64
	FiftyTwoWeekPos  *float64
	HV20D            *float64
	HV60D            *float64
	MACross          *string
	MADivergence50D  *float64
	MADivergence200D *float64
	RelativeBench20D *float64
	VolRatio         *float64
}

// ComputeSymbolIndicators computes all 14-column indicators for a single symbol.
// benchPrices provides the benchmark's price history for relative performance.
// isBenchmark skips RelativeBench20D (benchmark vs itself is meaningless).
func ComputeSymbolIndicators(
	prices []domain.DailyPrice,
	benchPrices []domain.DailyPrice,
	isBenchmark bool,
) SymbolIndicators {
	adjCloses := extractAdjCloses(prices)
	if len(adjCloses) == 0 {
		return SymbolIndicators{}
	}

	currentAdj := adjCloses[len(adjCloses)-1]

	high := FiftyTwoWeekHigh(prices)
	low := FiftyTwoWeekLow(prices)

	var pos *float64
	if high != nil && low != nil {
		pos = FiftyTwoWeekPosition(currentAdj, *high, *low)
	}

	ma50 := MovingAverage(prices, 50)
	ma200 := MovingAverage(prices, 200)

	change20D := PriceChange(prices, 20)

	var relBench *float64
	if !isBenchmark {
		benchChange := PriceChange(benchPrices, 20)
		relBench = RelativeBenchmark(change20D, benchChange)
	}

	return SymbolIndicators{
		AdjClose:         currentAdj,
		Change5D:         PriceChange(prices, 5),
		Change20D:        change20D,
		FiftyTwoWeekHigh: high,
		FiftyTwoWeekLow:  low,
		FiftyTwoWeekPos:  pos,
		HV20D:            HistoricalVolatility(prices, 20),
		HV60D:            HistoricalVolatility(prices, 60),
		MACross:          DetectMACross(prices, maCrossLookbackDays),
		MADivergence50D:  MADivergence(currentAdj, ma50),
		MADivergence200D: MADivergence(currentAdj, ma200),
		RelativeBench20D: relBench,
		VolRatio:         VolumeRatio(prices, 20),
	}
}
