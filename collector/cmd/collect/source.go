package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jusikbot/collector/internal/config"
	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/httpclient"
	"github.com/jusikbot/collector/internal/ratelimit"
	"github.com/jusikbot/collector/internal/store"
	"github.com/jusikbot/collector/internal/tiingo"
	"golang.org/x/time/rate"
)

const tiingoBaseURL = "https://api.tiingo.com"

// Why 5s: Tiingo allows 50 req/hr burst with backoff.
// 5s is conservative enough to avoid rate limiting under normal conditions.
var tiingoRetryCfg = ratelimit.RetryConfig{
	InitialBackoff: 5 * time.Second,
	MaxAttempts:    3,
	MaxBackoff:     60 * time.Second,
}

type sourceCollector struct {
	env       config.Env
	repo      *store.Repository
	watchlist []domain.WatchlistEntry
}

func (c *sourceCollector) collect(ctx context.Context, source string) SourceResult {
	started := time.Now()
	var err error

	switch source {
	case "tiingo":
		err = c.collectTiingo(ctx)
	case "kis":
		slog.Warn("not implemented yet", "source", "kis")
	case "fx":
		slog.Warn("not implemented yet", "source", "fx")
	default:
		err = fmt.Errorf("unknown source %q", source)
	}

	return SourceResult{
		Elapsed: time.Since(started),
		Error:   err,
		Source:  source,
	}
}

func (c *sourceCollector) collectTiingo(ctx context.Context) error {
	if c.env.TiingoAPIKey == "" {
		return fmt.Errorf("TIINGO_API_KEY is not set")
	}

	usEntries := config.FilterByMarket(c.watchlist, domain.MarketUS)
	if len(usEntries) == 0 {
		slog.Info("no US symbols in watchlist, skipping tiingo")
		return nil
	}

	symbols := make([]string, len(usEntries))
	for i, e := range usEntries {
		symbols[i] = e.Symbol
	}

	gaps, err := c.repo.DetectGaps(ctx, symbols)
	if err != nil {
		return fmt.Errorf("detect gaps: %w", err)
	}

	httpClient := httpclient.NewClient(
		tiingoBaseURL,
		map[string]string{"Authorization": "Token " + c.env.TiingoAPIKey},
		nil,
		0,
	)
	tiingoClient := tiingo.NewClient(httpClient)
	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)
	collector := tiingo.NewCollector(tiingoClient, limiter, tiingoRetryCfg)

	prices, collectErr := collector.CollectAll(ctx, usEntries, gaps)

	// Why save before checking error: CollectAll returns partial results on failure.
	var upsertErr error
	if len(prices) > 0 {
		if collectErr != nil {
			slog.Warn("saving partial results before reporting error",
				"collected", len(prices), "error", collectErr)
		}
		n, err := c.repo.UpsertPrices(ctx, prices)
		if err != nil {
			upsertErr = fmt.Errorf("upsert tiingo prices: %w", err)
		} else {
			slog.Info("tiingo prices saved", "rows", n)
		}
	}

	return errors.Join(collectErr, upsertErr)
}
