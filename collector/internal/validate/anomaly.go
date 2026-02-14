package validate

import (
	"math"

	"github.com/jusikbot/collector/internal/domain"
)

// Anomaly detection thresholds per market and security type.
// Why these values: KR 30% matches KRX daily price limit band,
// US ETF 15% reflects lower expected volatility of diversified funds,
// US stock 50% accommodates high-volatility individual equities.
const (
	ThresholdKR      = 0.30
	ThresholdUSETF   = 0.15
	ThresholdUSStock = 0.50
)

// IsPriceAnomaly returns true when the adj_close percentage change between
// consecutive trading days exceeds the market+type-specific threshold.
// Returns false for the first data point (previous == 0).
func IsPriceAnomaly(current, previous float64, market domain.Market, secType domain.SecurityType) bool {
	if previous == 0 {
		return false
	}

	changeRatio := math.Abs(current-previous) / previous
	threshold := resolveThreshold(market, secType)

	return changeRatio > threshold
}

// CrossValidateAdjClose returns true (confirmed anomaly) when no corporate
// action explains the adj_close deviation. A splitFactor != 1 or divCash > 0
// indicates a legitimate corporate action.
func CrossValidateAdjClose(splitFactor, divCash float64) bool {
	isSplit := splitFactor != 1.0 && splitFactor != 0
	isDividend := divCash > 0

	return !isSplit && !isDividend
}

func resolveThreshold(market domain.Market, secType domain.SecurityType) float64 {
	if market == domain.MarketKR {
		return ThresholdKR
	}

	if secType == domain.SecurityTypeETF {
		return ThresholdUSETF
	}

	return ThresholdUSStock
}
