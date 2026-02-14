package domain

import (
	"context"
	"time"
)

// StockDataFetcher is the common contract for stock data API clients.
// Why interface: allows swapping Tiingo/KIS without touching orchestration code (ADR-0001).
type StockDataFetcher interface {
	FetchDailyPrices(ctx context.Context, symbol string, from time.Time, to time.Time) ([]DailyPrice, error)
}
