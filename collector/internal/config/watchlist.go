package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jusikbot/collector/internal/domain"
)

type watchlistFileEntry struct {
	Market string `json:"market"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
}

var (
	validMarkets = map[string]domain.Market{
		"US": domain.MarketUS,
		"KR": domain.MarketKR,
	}
	validSecurityTypes = map[string]domain.SecurityType{
		"stock": domain.SecurityTypeStock,
		"etf":   domain.SecurityTypeETF,
	}
)

func LoadWatchlist(path string) ([]domain.WatchlistEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading watchlist: %w", err)
	}

	var raw []watchlistFileEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing watchlist JSON: %w", err)
	}

	if len(raw) == 0 {
		return nil, errors.New("watchlist is empty")
	}

	entries := make([]domain.WatchlistEntry, 0, len(raw))
	for i, r := range raw {
		entry, err := toWatchlistEntry(r)
		if err != nil {
			return nil, fmt.Errorf("watchlist entry [%d]: %w", i, err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func FilterByMarket(entries []domain.WatchlistEntry, market domain.Market) []domain.WatchlistEntry {
	var filtered []domain.WatchlistEntry
	for _, e := range entries {
		if e.Market == market {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func toWatchlistEntry(raw watchlistFileEntry) (domain.WatchlistEntry, error) {
	if raw.Symbol == "" {
		return domain.WatchlistEntry{}, errors.New("symbol is required")
	}

	market, ok := validMarkets[raw.Market]
	if !ok {
		return domain.WatchlistEntry{}, fmt.Errorf("invalid market %q (allowed: US, KR)", raw.Market)
	}

	secType, ok := validSecurityTypes[raw.Type]
	if !ok {
		return domain.WatchlistEntry{}, fmt.Errorf("invalid type %q (allowed: stock, etf)", raw.Type)
	}

	return domain.WatchlistEntry{
		Market: market,
		Name:   raw.Name,
		Symbol: raw.Symbol,
		Type:   secType,
	}, nil
}
