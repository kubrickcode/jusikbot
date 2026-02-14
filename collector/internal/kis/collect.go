package kis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/ratelimit"
	"github.com/jusikbot/collector/internal/validate"
	"golang.org/x/time/rate"
)

const defaultLookbackDays = 365

// Collector orchestrates sequential symbol collection with rate limiting and anomaly detection.
type Collector struct {
	client   *Client
	limiter  *rate.Limiter
	retryCfg ratelimit.RetryConfig
}

func NewCollector(client *Client, limiter *rate.Limiter, retryCfg ratelimit.RetryConfig) *Collector {
	return &Collector{
		client:   client,
		limiter:  limiter,
		retryCfg: retryCfg,
	}
}

// CollectAll fetches daily prices for all KR entries sequentially with rate limiting.
// Returns partial results on error.
func (c *Collector) CollectAll(
	ctx context.Context,
	entries []domain.WatchlistEntry,
	gaps map[string]time.Time,
) ([]domain.DailyPrice, error) {
	var allPrices []domain.DailyPrice
	to := time.Now().Truncate(24 * time.Hour)

	for _, entry := range entries {
		if ctx.Err() != nil {
			return allPrices, ctx.Err()
		}

		from := computeStartDate(to, gaps, entry.Symbol)
		if !from.Before(to) {
			slog.Info("already up to date", "symbol", entry.Symbol)
			continue
		}

		raw, err := ratelimit.FetchWithRateLimit(ctx, c.limiter, c.retryCfg, IsRetryable,
			func(ctx context.Context) ([]domain.DailyPrice, error) {
				return c.client.FetchDailyPrices(ctx, entry.Symbol, from, to)
			},
		)
		if err != nil {
			return allPrices, fmt.Errorf("collect %s: %w", entry.Symbol, err)
		}

		prices := markAnomalies(raw, entry)
		allPrices = append(allPrices, prices...)
		slog.Info("collected", "rows", len(prices), "symbol", entry.Symbol)
	}

	return allPrices, nil
}

// markAnomalies flags price anomalies using only IsPriceAnomaly (no cross-validation).
// Why no CrossValidateAdjClose: KIS does not provide splitFactor/divCash fields.
// For KR market, 30% threshold matches KRX price limits, making IsPriceAnomaly sufficient.
func markAnomalies(prices []domain.DailyPrice, entry domain.WatchlistEntry) []domain.DailyPrice {
	for i := 1; i < len(prices); i++ {
		if validate.IsPriceAnomaly(prices[i].AdjClose, prices[i-1].AdjClose, entry.Market, entry.Type) {
			prices[i].IsAnomaly = true
			slog.Warn("anomaly detected",
				"change_pct", fmt.Sprintf("%.1f%%", (prices[i].AdjClose-prices[i-1].AdjClose)/prices[i-1].AdjClose*100),
				"date", prices[i].Date.Format("2006-01-02"),
				"symbol", entry.Symbol,
			)
		}
	}
	return prices
}

func computeStartDate(to time.Time, gaps map[string]time.Time, symbol string) time.Time {
	from := to.AddDate(0, 0, -defaultLookbackDays)
	if lastDate, ok := gaps[symbol]; ok {
		// Why +1 day: last recorded date is already in DB, start from next day.
		candidate := lastDate.AddDate(0, 0, 1)
		if candidate.After(from) {
			from = candidate
		}
	}
	return from
}
