package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jusikbot/collector/internal/config"
	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/fx"
	"github.com/jusikbot/collector/internal/httpclient"
	"github.com/jusikbot/collector/internal/kis"
	"github.com/jusikbot/collector/internal/ratelimit"
	"github.com/jusikbot/collector/internal/store"
	"github.com/jusikbot/collector/internal/tiingo"
	"golang.org/x/time/rate"
)

const (
	frankfurterBaseURL = "https://api.frankfurter.dev"
	kisBaseURL         = "https://openapi.koreainvestment.com:9443"
	tiingoBaseURL      = "https://api.tiingo.com"
)

// Why 5s: Tiingo allows 50 req/hr burst with backoff.
// 5s is conservative enough to avoid rate limiting under normal conditions.
var tiingoRetryCfg = ratelimit.RetryConfig{
	InitialBackoff: 5 * time.Second,
	MaxAttempts:    3,
	MaxBackoff:     60 * time.Second,
}

// Why 2s: Frankfurter has no rate limit, but retry with backoff for transient failures.
var fxRetryCfg = ratelimit.RetryConfig{
	InitialBackoff: 2 * time.Second,
	MaxAttempts:    3,
	MaxBackoff:     30 * time.Second,
}

// Why 1s: KIS personal accounts allow ~20 req/sec, but conservative to avoid throttling.
var kisRetryCfg = ratelimit.RetryConfig{
	InitialBackoff: 2 * time.Second,
	MaxAttempts:    3,
	MaxBackoff:     30 * time.Second,
}

type sourceCollector struct {
	env       config.Env
	repo      *store.Repository
	watchlist []domain.WatchlistEntry
}

func (c *sourceCollector) collectKIS(ctx context.Context) error {
	if c.env.KISAppKey == "" || c.env.KISAppSecret == "" {
		return fmt.Errorf("KIS_APP_KEY and KIS_APP_SECRET are required")
	}

	krEntries := config.FilterByMarket(c.watchlist, domain.MarketKR)
	if len(krEntries) == 0 {
		slog.Info("no KR symbols in watchlist, skipping kis")
		return nil
	}

	symbols := make([]string, len(krEntries))
	for i, e := range krEntries {
		symbols[i] = e.Symbol
	}

	gaps, err := c.repo.DetectGaps(ctx, symbols)
	if err != nil {
		return fmt.Errorf("detect gaps: %w", err)
	}

	// Why credentials appear in both places: KIS triple auth requires appkey/appsecret
	// in POST body for token issuance (TokenProvider) AND in GET headers for data APIs (httpclient).
	tokenProvider := kis.NewTokenProvider(kisBaseURL, c.env.KISAppKey, c.env.KISAppSecret, nil)
	httpClient := httpclient.NewClient(
		kisBaseURL,
		map[string]string{
			"appkey":    c.env.KISAppKey,
			"appsecret": c.env.KISAppSecret,
		},
		nil,
		0,
	)
	kisClient := kis.NewClient(httpClient, tokenProvider)

	// Why Every(56ms): ~18 req/sec matches KIS personal account limits (analysis.md).
	limiter := rate.NewLimiter(rate.Every(56*time.Millisecond), 1)
	collector := kis.NewCollector(kisClient, limiter, kisRetryCfg)

	prices, collectErr := collector.CollectAll(ctx, krEntries, gaps)
	return c.savePartialResults(ctx, prices, collectErr, "kis")
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
	return c.savePartialResults(ctx, prices, collectErr, "tiingo")
}

func (c *sourceCollector) collectFX(ctx context.Context) error {
	gaps, err := c.repo.DetectFXGaps(ctx, []string{"USD/KRW"})
	if err != nil {
		return fmt.Errorf("detect fx gaps: %w", err)
	}

	httpClient := httpclient.NewClient(frankfurterBaseURL, nil, nil, 0)
	fxClient := fx.NewClient(httpClient)
	collector := fx.NewCollector(fxClient, fxRetryCfg)

	rates, err := collector.CollectFX(ctx, "USD", "KRW", gaps)
	if err != nil {
		return fmt.Errorf("collect fx: %w", err)
	}

	if len(rates) == 0 {
		return nil
	}

	n, err := c.repo.UpsertFXRates(ctx, rates)
	if err != nil {
		return fmt.Errorf("upsert fx rates: %w", err)
	}
	slog.Info("fx rates saved", "rows", n)

	return nil
}

// savePartialResults persists collected prices and joins any collection/upsert errors.
// Why save before checking collectErr: CollectAll returns partial results on failure.
func (c *sourceCollector) savePartialResults(
	ctx context.Context,
	prices []domain.DailyPrice,
	collectErr error,
	source string,
) error {
	var upsertErr error
	if len(prices) > 0 {
		if collectErr != nil {
			slog.Warn("saving partial results before reporting error",
				"collected", len(prices), "error", collectErr, "source", source)
		}
		n, err := c.repo.UpsertPrices(ctx, prices)
		if err != nil {
			upsertErr = fmt.Errorf("upsert %s prices: %w", source, err)
		} else {
			slog.Info("prices saved", "rows", n, "source", source)
		}
	}
	return errors.Join(collectErr, upsertErr)
}
