package summary

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
)

type stubPriceReader struct {
	prices   map[string][]domain.DailyPrice
	fxRates  map[string][]domain.FXRate
	priceErr map[string]error
	fxErr    map[string]error
}

func (s *stubPriceReader) FetchPriceHistory(_ context.Context, symbol string, _, _ time.Time) ([]domain.DailyPrice, error) {
	if s.priceErr != nil {
		if err, ok := s.priceErr[symbol]; ok {
			return nil, err
		}
	}
	return s.prices[symbol], nil
}

func (s *stubPriceReader) FetchFXRates(_ context.Context, pair string, _, _ time.Time) ([]domain.FXRate, error) {
	if s.fxErr != nil {
		if err, ok := s.fxErr[pair]; ok {
			return nil, err
		}
	}
	return s.fxRates[pair], nil
}

func TestGenerateSummary(t *testing.T) {
	t.Run("full pipeline renders all sections", func(t *testing.T) {
		reader := &stubPriceReader{
			prices: map[string][]domain.DailyPrice{
				"NVDA":   makePriceSeries(baseDate, repeatFloat(100, 30), 1000),
				"QQQ":    makePriceSeries(baseDate, repeatFloat(100, 30), 1000),
				"069500": makePriceSeries(baseDate, repeatFloat(35000, 30), 500),
			},
			fxRates: map[string][]domain.FXRate{
				"USD/KRW": {
					{Date: time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC), Pair: "USD/KRW", Rate: 1345.50, Source: "frankfurter"},
				},
			},
		}

		watchlist := []domain.WatchlistEntry{
			{Symbol: "NVDA", Name: "NVIDIA", Market: domain.MarketUS, Type: domain.SecurityTypeStock},
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Market: domain.MarketUS, Type: domain.SecurityTypeETF},
			{Symbol: "069500", Name: "KODEX 200", Market: domain.MarketKR, Type: domain.SecurityTypeETF},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := GenerateSummary(context.Background(), reader, watchlist, outputPath); err != nil {
			t.Fatalf("GenerateSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		content := string(got)
		if !strings.Contains(content, "## US Stocks") {
			t.Error("missing US Stocks section")
		}
		if !strings.Contains(content, "## KR Stocks") {
			t.Error("missing KR Stocks section")
		}
		if !strings.Contains(content, "## Exchange Rate") {
			t.Error("missing Exchange Rate section")
		}
		if !strings.Contains(content, "| NVDA |") {
			t.Error("missing NVDA row")
		}
		if !strings.Contains(content, "| QQQ |") {
			t.Error("missing QQQ row")
		}
		if !strings.Contains(content, "| 069500 |") {
			t.Error("missing 069500 row")
		}
		if !strings.Contains(content, "1,345.50") {
			t.Error("missing FX rate")
		}
	})

	t.Run("skips symbols with no price data", func(t *testing.T) {
		reader := &stubPriceReader{
			prices: map[string][]domain.DailyPrice{
				"NVDA": {},
				"QQQ":  makePriceSeries(baseDate, repeatFloat(100, 30), 1000),
			},
			fxRates: map[string][]domain.FXRate{},
		}

		watchlist := []domain.WatchlistEntry{
			{Symbol: "NVDA", Name: "NVIDIA", Market: domain.MarketUS, Type: domain.SecurityTypeStock},
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Market: domain.MarketUS, Type: domain.SecurityTypeETF},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := GenerateSummary(context.Background(), reader, watchlist, outputPath); err != nil {
			t.Fatalf("GenerateSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		content := string(got)
		if strings.Contains(content, "| NVDA |") {
			t.Error("NVDA should be skipped (no data)")
		}
		if !strings.Contains(content, "| QQQ |") {
			t.Error("QQQ should be present")
		}
	})

	t.Run("insufficient data adds notes", func(t *testing.T) {
		// 30 entries: enough for 20D but not 200D
		reader := &stubPriceReader{
			prices: map[string][]domain.DailyPrice{
				"NEW": makePriceSeries(baseDate, repeatFloat(50, 30), 1000),
				"QQQ": makePriceSeries(baseDate, repeatFloat(100, 30), 1000),
			},
			fxRates: map[string][]domain.FXRate{},
		}

		watchlist := []domain.WatchlistEntry{
			{Symbol: "NEW", Name: "New Stock", Market: domain.MarketUS, Type: domain.SecurityTypeStock},
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Market: domain.MarketUS, Type: domain.SecurityTypeETF},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := GenerateSummary(context.Background(), reader, watchlist, outputPath); err != nil {
			t.Fatalf("GenerateSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		content := string(got)
		if !strings.Contains(content, "*Notes:*") {
			t.Error("missing Notes section for insufficient data")
		}
		if !strings.Contains(content, "NEW (New Stock)") {
			t.Error("missing NEW in insufficient data notes")
		}
	})

	t.Run("propagates benchmark fetch error", func(t *testing.T) {
		benchErr := errors.New("db connection lost")
		reader := &stubPriceReader{
			prices:   map[string][]domain.DailyPrice{},
			fxRates:  map[string][]domain.FXRate{},
			priceErr: map[string]error{"QQQ": benchErr},
		}

		watchlist := []domain.WatchlistEntry{
			{Symbol: "NVDA", Name: "NVIDIA", Market: domain.MarketUS, Type: domain.SecurityTypeStock},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		err := GenerateSummary(context.Background(), reader, watchlist, outputPath)
		if err == nil {
			t.Fatal("expected error from benchmark fetch, got nil")
		}
		if !errors.Is(err, benchErr) {
			t.Errorf("expected wrapped benchErr, got: %v", err)
		}
	})

	t.Run("propagates symbol fetch error", func(t *testing.T) {
		symbolErr := errors.New("timeout reading NVDA")
		reader := &stubPriceReader{
			prices: map[string][]domain.DailyPrice{
				"QQQ":    makePriceSeries(baseDate, repeatFloat(100, 30), 1000),
				"069500": makePriceSeries(baseDate, repeatFloat(35000, 30), 500),
			},
			fxRates:  map[string][]domain.FXRate{},
			priceErr: map[string]error{"NVDA": symbolErr},
		}

		watchlist := []domain.WatchlistEntry{
			{Symbol: "NVDA", Name: "NVIDIA", Market: domain.MarketUS, Type: domain.SecurityTypeStock},
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Market: domain.MarketUS, Type: domain.SecurityTypeETF},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		err := GenerateSummary(context.Background(), reader, watchlist, outputPath)
		if err == nil {
			t.Fatal("expected error from symbol fetch, got nil")
		}
		if !errors.Is(err, symbolErr) {
			t.Errorf("expected wrapped symbolErr, got: %v", err)
		}
	})

	t.Run("no FX data omits exchange rate section", func(t *testing.T) {
		reader := &stubPriceReader{
			prices: map[string][]domain.DailyPrice{
				"QQQ": makePriceSeries(baseDate, repeatFloat(100, 30), 1000),
			},
			fxRates: map[string][]domain.FXRate{},
		}

		watchlist := []domain.WatchlistEntry{
			{Symbol: "QQQ", Name: "Invesco QQQ Trust", Market: domain.MarketUS, Type: domain.SecurityTypeETF},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := GenerateSummary(context.Background(), reader, watchlist, outputPath); err != nil {
			t.Fatalf("GenerateSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		if strings.Contains(string(got), "## Exchange Rate") {
			t.Error("Exchange Rate section should be absent when no FX data")
		}
	})
}
