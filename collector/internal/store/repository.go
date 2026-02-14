package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jusikbot/collector/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertPrices bulk-inserts or updates price_history via temp table + CopyFrom + INSERT ON CONFLICT.
// Why temp table: pgx CopyFrom does not support ON CONFLICT directly.
func (r *Repository) UpsertPrices(ctx context.Context, prices []domain.DailyPrice) (int64, error) {
	if len(prices) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin upsert prices: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		CREATE TEMP TABLE tmp_prices (
			adj_close  DOUBLE PRECISION NOT NULL,
			close      DOUBLE PRECISION NOT NULL,
			date       DATE             NOT NULL,
			high       DOUBLE PRECISION NOT NULL,
			is_anomaly BOOLEAN          NOT NULL DEFAULT FALSE,
			low        DOUBLE PRECISION NOT NULL,
			open       DOUBLE PRECISION NOT NULL,
			source     TEXT             NOT NULL,
			symbol     TEXT             NOT NULL,
			volume     BIGINT           NOT NULL
		) ON COMMIT DROP
	`); err != nil {
		return 0, fmt.Errorf("create temp prices table: %w", err)
	}

	// Why fetched_at excluded: server-side NOW() used for both insert (DEFAULT) and update (SET).
	// DailyPrice.FetchedAt is read-only, populated by FetchPriceHistory.
	columns := []string{"adj_close", "close", "date", "high", "is_anomaly", "low", "open", "source", "symbol", "volume"}
	if _, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"tmp_prices"},
		columns,
		pgx.CopyFromSlice(len(prices), func(i int) ([]any, error) {
			p := prices[i]
			return []any{p.AdjClose, p.Close, p.Date, p.High, p.IsAnomaly, p.Low, p.Open, p.Source, p.Symbol, p.Volume}, nil
		}),
	); err != nil {
		return 0, fmt.Errorf("copy prices to temp table: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		INSERT INTO price_history (adj_close, close, date, high, is_anomaly, low, open, source, symbol, volume)
		SELECT adj_close, close, date, high, is_anomaly, low, open, source, symbol, volume
		FROM tmp_prices
		ON CONFLICT (symbol, date) DO UPDATE SET
			adj_close  = EXCLUDED.adj_close,
			close      = EXCLUDED.close,
			high       = EXCLUDED.high,
			is_anomaly = EXCLUDED.is_anomaly,
			low        = EXCLUDED.low,
			open       = EXCLUDED.open,
			source     = EXCLUDED.source,
			volume     = EXCLUDED.volume,
			fetched_at = NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("upsert prices from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit upsert prices: %w", err)
	}

	return tag.RowsAffected(), nil
}

// UpsertFXRates bulk-inserts or updates fx_rate via the same temp table pattern.
func (r *Repository) UpsertFXRates(ctx context.Context, rates []domain.FXRate) (int64, error) {
	if len(rates) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin upsert fx rates: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		CREATE TEMP TABLE tmp_fx_rates (
			date   DATE             NOT NULL,
			pair   TEXT             NOT NULL,
			rate   DOUBLE PRECISION NOT NULL,
			source TEXT             NOT NULL
		) ON COMMIT DROP
	`); err != nil {
		return 0, fmt.Errorf("create temp fx_rates table: %w", err)
	}

	// Why fetched_at excluded: same rationale as UpsertPrices.
	columns := []string{"date", "pair", "rate", "source"}
	if _, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"tmp_fx_rates"},
		columns,
		pgx.CopyFromSlice(len(rates), func(i int) ([]any, error) {
			rate := rates[i]
			return []any{rate.Date, rate.Pair, rate.Rate, rate.Source}, nil
		}),
	); err != nil {
		return 0, fmt.Errorf("copy fx rates to temp table: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		INSERT INTO fx_rate (date, pair, rate, source)
		SELECT date, pair, rate, source
		FROM tmp_fx_rates
		ON CONFLICT (pair, date) DO UPDATE SET
			rate       = EXCLUDED.rate,
			source     = EXCLUDED.source,
			fetched_at = NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("upsert fx rates from temp table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit upsert fx rates: %w", err)
	}

	return tag.RowsAffected(), nil
}

// DetectGaps returns the last recorded date per symbol for incremental collection.
// Symbols with no data are absent from the returned map.
func (r *Repository) DetectGaps(ctx context.Context, symbols []string) (map[string]time.Time, error) {
	if len(symbols) == 0 {
		return make(map[string]time.Time), nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT symbol, MAX(date)
		FROM price_history
		WHERE symbol = ANY($1)
		GROUP BY symbol
	`, symbols)
	if err != nil {
		return nil, fmt.Errorf("detect price gaps: %w", err)
	}
	defer rows.Close()

	gaps := make(map[string]time.Time, len(symbols))
	for rows.Next() {
		var symbol string
		var lastDate time.Time
		if err := rows.Scan(&symbol, &lastDate); err != nil {
			return nil, fmt.Errorf("scan price gap row: %w", err)
		}
		gaps[symbol] = lastDate
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price gap rows: %w", err)
	}

	return gaps, nil
}

// DetectFXGaps returns the last recorded date per currency pair for incremental FX collection.
func (r *Repository) DetectFXGaps(ctx context.Context, pairs []string) (map[string]time.Time, error) {
	if len(pairs) == 0 {
		return make(map[string]time.Time), nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT pair, MAX(date)
		FROM fx_rate
		WHERE pair = ANY($1)
		GROUP BY pair
	`, pairs)
	if err != nil {
		return nil, fmt.Errorf("detect fx gaps: %w", err)
	}
	defer rows.Close()

	gaps := make(map[string]time.Time, len(pairs))
	for rows.Next() {
		var pair string
		var lastDate time.Time
		if err := rows.Scan(&pair, &lastDate); err != nil {
			return nil, fmt.Errorf("scan fx gap row: %w", err)
		}
		gaps[pair] = lastDate
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fx gap rows: %w", err)
	}

	return gaps, nil
}

// FetchPriceHistory retrieves price data for a symbol within a date range, sorted ascending.
func (r *Repository) FetchPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]domain.DailyPrice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT adj_close, close, date, fetched_at, high, is_anomaly, low, open, source, symbol, volume
		FROM price_history
		WHERE symbol = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`, symbol, from, to)
	if err != nil {
		return nil, fmt.Errorf("fetch price history for %s: %w", symbol, err)
	}
	defer rows.Close()

	prices := make([]domain.DailyPrice, 0)
	for rows.Next() {
		var p domain.DailyPrice
		if err := rows.Scan(&p.AdjClose, &p.Close, &p.Date, &p.FetchedAt, &p.High, &p.IsAnomaly, &p.Low, &p.Open, &p.Source, &p.Symbol, &p.Volume); err != nil {
			return nil, fmt.Errorf("scan price row: %w", err)
		}
		prices = append(prices, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price rows: %w", err)
	}

	return prices, nil
}

// FetchFXRates retrieves FX rate data for a currency pair within a date range, sorted ascending.
func (r *Repository) FetchFXRates(ctx context.Context, pair string, from, to time.Time) ([]domain.FXRate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT date, fetched_at, pair, rate, source
		FROM fx_rate
		WHERE pair = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`, pair, from, to)
	if err != nil {
		return nil, fmt.Errorf("fetch fx rates for %s: %w", pair, err)
	}
	defer rows.Close()

	rates := make([]domain.FXRate, 0)
	for rows.Next() {
		var fr domain.FXRate
		if err := rows.Scan(&fr.Date, &fr.FetchedAt, &fr.Pair, &fr.Rate, &fr.Source); err != nil {
			return nil, fmt.Errorf("scan fx rate row: %w", err)
		}
		rates = append(rates, fr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fx rate rows: %w", err)
	}

	return rates, nil
}
