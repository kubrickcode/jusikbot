package summary

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jusikbot/collector/internal/domain"
)

// PriceReader abstracts DB read-back for summary generation.
// Satisfied by store.Repository.
type PriceReader interface {
	FetchPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]domain.DailyPrice, error)
	FetchFXRates(ctx context.Context, pair string, from, to time.Time) ([]domain.FXRate, error)
}

// Why 380: 252 trading days/year + ~128 calendar gap days ensures full 52-week coverage
// plus buffer for 200D MA calculation.
const priceHistoryLookbackDays = 380

// GenerateSummary reads prices from DB, computes indicators, and writes data/summary.md.
func GenerateSummary(ctx context.Context, reader PriceReader, watchlist []domain.WatchlistEntry, outputPath string) error {
	now := time.Now()
	from := now.AddDate(0, 0, -priceHistoryLookbackDays)
	to := now

	benchPrices, err := loadBenchmarkPrices(ctx, reader, from, to)
	if err != nil {
		return err
	}

	usRows, krRows, insufficientSymbols, err := computeAllIndicators(ctx, reader, watchlist, benchPrices, from, to)
	if err != nil {
		return err
	}

	fxEntry := loadLatestFXRate(ctx, reader, from, to)

	data := SummaryData{
		FXRate:              fxEntry,
		GeneratedAt:         now.UTC().Format("2006-01-02 15:04 UTC"),
		InsufficientSymbols: insufficientSymbols,
		KRRows:              krRows,
		USRows:              usRows,
	}

	return RenderSummary(data, outputPath)
}

func loadBenchmarkPrices(ctx context.Context, reader PriceReader, from, to time.Time) (map[domain.Market][]domain.DailyPrice, error) {
	benchPrices := make(map[domain.Market][]domain.DailyPrice, len(BenchmarkSymbols))
	for market, symbol := range BenchmarkSymbols {
		prices, err := reader.FetchPriceHistory(ctx, symbol, from, to)
		if err != nil {
			return nil, fmt.Errorf("fetch benchmark %s: %w", symbol, err)
		}
		benchPrices[market] = prices
	}
	return benchPrices, nil
}

func computeAllIndicators(
	ctx context.Context,
	reader PriceReader,
	watchlist []domain.WatchlistEntry,
	benchPrices map[domain.Market][]domain.DailyPrice,
	from, to time.Time,
) (usRows, krRows []SymbolRow, insufficientSymbols []string, err error) {
	for _, entry := range watchlist {
		prices, fetchErr := reader.FetchPriceHistory(ctx, entry.Symbol, from, to)
		if fetchErr != nil {
			return nil, nil, nil, fmt.Errorf("fetch prices for %s: %w", entry.Symbol, fetchErr)
		}

		if len(prices) == 0 {
			slog.Warn("no price data, skipping symbol", "symbol", entry.Symbol)
			continue
		}

		isBenchmark := BenchmarkSymbols[entry.Market] == entry.Symbol
		indicators := ComputeSymbolIndicators(prices, benchPrices[entry.Market], isBenchmark)

		row := SymbolRow{
			Indicators: indicators,
			Name:       entry.Name,
			Symbol:     entry.Symbol,
		}

		switch entry.Market {
		case domain.MarketUS:
			usRows = append(usRows, row)
		case domain.MarketKR:
			krRows = append(krRows, row)
		}

		if indicators.MADivergence200D == nil && len(prices) > 0 {
			insufficientSymbols = append(insufficientSymbols,
				fmt.Sprintf("%s (%s): 200D MA 데이터 부족 (< 200 거래일)", entry.Symbol, entry.Name))
		}
	}

	return usRows, krRows, insufficientSymbols, nil
}

func loadLatestFXRate(ctx context.Context, reader PriceReader, from, to time.Time) *FXRateEntry {
	fxRates, err := reader.FetchFXRates(ctx, "USD/KRW", from, to)
	if err != nil {
		slog.Warn("FX rate fetch skipped", "error", err)
		return nil
	}
	if len(fxRates) == 0 {
		return nil
	}

	latest := fxRates[len(fxRates)-1]
	return &FXRateEntry{
		Date: latest.Date.Format("2006-01-02"),
		Pair: latest.Pair,
		Rate: latest.Rate,
	}
}
