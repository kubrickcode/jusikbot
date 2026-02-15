package fx

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/ratelimit"
)

const defaultLookbackDays = 365

// Collector orchestrates FX rate collection with retry support.
type Collector struct {
	client   *Client
	retryCfg ratelimit.RetryConfig

	// Why injectable: enables deterministic testing without time-dependent flakiness.
	now func() time.Time
}

func NewCollector(client *Client, retryCfg ratelimit.RetryConfig) *Collector {
	return &Collector{
		client:   client,
		now:      time.Now,
		retryCfg: retryCfg,
	}
}

// CollectFX returns nil when data is already up to date.
func (c *Collector) CollectFX(
	ctx context.Context,
	base, target string,
	gaps map[string]time.Time,
) ([]domain.FXRate, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	pair := base + "/" + target
	to := c.now().Truncate(24 * time.Hour)
	from := computeStartDate(to, gaps, pair)

	if !from.Before(to) {
		slog.Info("fx already up to date", "pair", pair)
		return nil, nil
	}

	rates, err := ratelimit.WithRetry(ctx, c.retryCfg, IsRetryable,
		func(ctx context.Context) ([]domain.FXRate, error) {
			return c.client.FetchRates(ctx, base, target, from, to)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("collect fx %s: %w", pair, err)
	}

	slog.Info("fx collected", "pair", pair, "rows", len(rates))
	return rates, nil
}

func computeStartDate(to time.Time, gaps map[string]time.Time, pair string) time.Time {
	from := to.AddDate(0, 0, -defaultLookbackDays)
	if lastDate, ok := gaps[pair]; ok {
		// Why +1 day: last recorded date is already in DB, start from next day.
		candidate := lastDate.AddDate(0, 0, 1)
		if candidate.After(from) {
			from = candidate
		}
	}
	return from
}
