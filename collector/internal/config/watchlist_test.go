package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jusikbot/collector/internal/domain"
)

func writeTestWatchlist(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "watchlist.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test watchlist: %v", err)
	}
	return path
}

func TestLoadWatchlist_Valid(t *testing.T) {
	path := writeTestWatchlist(t, `[
		{"symbol": "NVDA", "name": "NVIDIA", "market": "US", "type": "stock"},
		{"symbol": "069500", "name": "KODEX 200", "market": "KR", "type": "etf"}
	]`)

	entries, err := LoadWatchlist(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	got := entries[0]
	if got.Symbol != "NVDA" {
		t.Errorf("Symbol = %q, want %q", got.Symbol, "NVDA")
	}
	if got.Name != "NVIDIA" {
		t.Errorf("Name = %q, want %q", got.Name, "NVIDIA")
	}
	if got.Market != domain.MarketUS {
		t.Errorf("Market = %q, want %q", got.Market, domain.MarketUS)
	}
	if got.Type != domain.SecurityTypeStock {
		t.Errorf("Type = %q, want %q", got.Type, domain.SecurityTypeStock)
	}
}

func TestLoadWatchlist_InvalidJSON(t *testing.T) {
	path := writeTestWatchlist(t, `not json`)

	_, err := LoadWatchlist(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadWatchlist_EmptyArray(t *testing.T) {
	path := writeTestWatchlist(t, `[]`)

	_, err := LoadWatchlist(path)
	if err == nil {
		t.Fatal("expected error for empty watchlist, got nil")
	}
}

func TestLoadWatchlist_FileNotFound(t *testing.T) {
	_, err := LoadWatchlist("/nonexistent/watchlist.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadWatchlist_MissingSymbol(t *testing.T) {
	path := writeTestWatchlist(t, `[
		{"name": "NVIDIA", "market": "US", "type": "stock"}
	]`)

	_, err := LoadWatchlist(path)
	if err == nil {
		t.Fatal("expected error for missing symbol, got nil")
	}
}

func TestLoadWatchlist_InvalidMarket(t *testing.T) {
	path := writeTestWatchlist(t, `[
		{"symbol": "TSLA", "name": "Tesla", "market": "JP", "type": "stock"}
	]`)

	_, err := LoadWatchlist(path)
	if err == nil {
		t.Fatal("expected error for invalid market, got nil")
	}
}

func TestLoadWatchlist_InvalidSecurityType(t *testing.T) {
	path := writeTestWatchlist(t, `[
		{"symbol": "TSLA", "name": "Tesla", "market": "US", "type": "bond"}
	]`)

	_, err := LoadWatchlist(path)
	if err == nil {
		t.Fatal("expected error for invalid security type, got nil")
	}
}

func TestFilterByMarket_US(t *testing.T) {
	entries := []domain.WatchlistEntry{
		{Symbol: "NVDA", Market: domain.MarketUS},
		{Symbol: "069500", Market: domain.MarketKR},
		{Symbol: "QQQ", Market: domain.MarketUS},
	}

	us := FilterByMarket(entries, domain.MarketUS)
	if len(us) != 2 {
		t.Fatalf("got %d US entries, want 2", len(us))
	}
	if us[0].Symbol != "NVDA" {
		t.Errorf("first US symbol = %q, want %q", us[0].Symbol, "NVDA")
	}
	if us[1].Symbol != "QQQ" {
		t.Errorf("second US symbol = %q, want %q", us[1].Symbol, "QQQ")
	}
}

func TestFilterByMarket_KR(t *testing.T) {
	entries := []domain.WatchlistEntry{
		{Symbol: "NVDA", Market: domain.MarketUS},
		{Symbol: "069500", Market: domain.MarketKR},
	}

	kr := FilterByMarket(entries, domain.MarketKR)
	if len(kr) != 1 {
		t.Fatalf("got %d KR entries, want 1", len(kr))
	}
	if kr[0].Symbol != "069500" {
		t.Errorf("KR symbol = %q, want %q", kr[0].Symbol, "069500")
	}
}

func TestFilterByMarket_NoMatch(t *testing.T) {
	entries := []domain.WatchlistEntry{
		{Symbol: "NVDA", Market: domain.MarketUS},
	}

	kr := FilterByMarket(entries, domain.MarketKR)
	if len(kr) != 0 {
		t.Fatalf("got %d KR entries, want 0", len(kr))
	}
}
