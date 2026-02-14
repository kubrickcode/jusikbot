package config

import (
	"testing"
)

func TestLoadEnv_AllRequired(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("TIINGO_API_KEY", "test-tiingo-key")
	t.Setenv("KIS_APP_KEY", "test-kis-key")
	t.Setenv("KIS_APP_SECRET", "test-kis-secret")

	cfg, err := LoadEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://localhost:5432/test" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://localhost:5432/test")
	}
	if cfg.TiingoAPIKey != "test-tiingo-key" {
		t.Errorf("TiingoAPIKey = %q, want %q", cfg.TiingoAPIKey, "test-tiingo-key")
	}
	if cfg.KISAppKey != "test-kis-key" {
		t.Errorf("KISAppKey = %q, want %q", cfg.KISAppKey, "test-kis-key")
	}
	if cfg.KISAppSecret != "test-kis-secret" {
		t.Errorf("KISAppSecret = %q, want %q", cfg.KISAppSecret, "test-kis-secret")
	}
}

func TestLoadEnv_MissingDatabaseURL(t *testing.T) {
	t.Setenv("TIINGO_API_KEY", "key")
	t.Setenv("KIS_APP_KEY", "key")
	t.Setenv("KIS_APP_SECRET", "secret")

	_, err := LoadEnv()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
}

func TestLoadEnv_MissingTiingoAPIKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("KIS_APP_KEY", "key")
	t.Setenv("KIS_APP_SECRET", "secret")

	_, err := LoadEnv()
	if err == nil {
		t.Fatal("expected error for missing TIINGO_API_KEY, got nil")
	}
}

func TestLoadEnv_MissingKISAppKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("TIINGO_API_KEY", "key")
	t.Setenv("KIS_APP_SECRET", "secret")

	_, err := LoadEnv()
	if err == nil {
		t.Fatal("expected error for missing KIS_APP_KEY, got nil")
	}
}

func TestLoadEnv_MissingKISAppSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("TIINGO_API_KEY", "key")
	t.Setenv("KIS_APP_KEY", "key")

	_, err := LoadEnv()
	if err == nil {
		t.Fatal("expected error for missing KIS_APP_SECRET, got nil")
	}
}
