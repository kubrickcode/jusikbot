package tiingo

import (
	"context"
	"errors"
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

// CollectAll fetches daily prices for all entries sequentially with rate limiting.
// Invalid tickers are skipped with a warning. Returns partial results on error.
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
			func(ctx context.Context) ([]tiingoPrice, error) {
				return c.client.fetchPrices(ctx, entry.Symbol, from, to)
			},
		)
		if err != nil {
			if errors.Is(err, ErrTickerInvalid) {
				slog.Warn("skipping invalid ticker", "symbol", entry.Symbol)
				continue
			}
			return allPrices, fmt.Errorf("collect %s: %w", entry.Symbol, err)
		}

		prices, err := markAnomalies(raw, entry)
		if err != nil {
			return allPrices, fmt.Errorf("validate %s: %w", entry.Symbol, err)
		}

		allPrices = append(allPrices, prices...)
		slog.Info("collected", "rows", len(prices), "symbol", entry.Symbol)
	}

	return allPrices, nil
}

// markAnomalies converts raw Tiingo data to domain prices with anomaly detection.
// Uses adj_close for anomaly detection and splitFactor/divCash for cross-validation.
// Why cross-validation here: only Tiingo provides splitFactor/divCash; this context
// is lost after conversion to domain.DailyPrice.
func markAnomalies(raw []tiingoPrice, entry domain.WatchlistEntry) ([]domain.DailyPrice, error) {
	prices := make([]domain.DailyPrice, 0, len(raw))

	for i, r := range raw {
		p, err := toDailyPrice(r, entry.Symbol)
		if err != nil {
			return nil, fmt.Errorf("row %d date %q: %w", i, r.Date, err)
		}

		if i > 0 && isConfirmedAnomaly(r, raw[i-1].AdjClose, entry) {
			p.IsAnomaly = true
			slog.Warn("anomaly detected",
				"change_pct", fmt.Sprintf("%.1f%%", (r.AdjClose-raw[i-1].AdjClose)/raw[i-1].AdjClose*100),
				"date", p.Date.Format("2006-01-02"),
				"symbol", entry.Symbol,
			)
		}

		prices = append(prices, p)
	}

	return prices, nil
}

func isConfirmedAnomaly(r tiingoPrice, prevAdjClose float64, entry domain.WatchlistEntry) bool {
	return validate.IsPriceAnomaly(r.AdjClose, prevAdjClose, entry.Market, entry.Type) &&
		validate.CrossValidateAdjClose(r.SplitFactor, r.DivCash)
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
