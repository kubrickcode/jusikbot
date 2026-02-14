package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jusikbot/collector/internal/store"
)

func databaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	return url
}

func connectAndClean(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pool, err := store.ConnectDB(ctx, databaseURL(t))
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	for _, table := range []string{"price_history", "fx_rate", "schema_version"} {
		if _, err := pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			t.Fatalf("drop table %s: %v", table, err)
		}
	}

	return pool
}

func assertTableExists(t *testing.T, pool *pgxpool.Pool, tableName string) {
	t.Helper()
	var exists bool
	err := pool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`, tableName).Scan(&exists)
	if err != nil {
		t.Fatalf("check table %s: %v", tableName, err)
	}
	if !exists {
		t.Errorf("table %s does not exist", tableName)
	}
}

func TestRunMigrations(t *testing.T) {
	pool := connectAndClean(t)
	defer pool.Close()
	ctx := context.Background()

	t.Run("creates tables on first run", func(t *testing.T) {
		if err := store.RunMigrations(ctx, pool); err != nil {
			t.Fatalf("first migration run: %v", err)
		}

		assertTableExists(t, pool, "price_history")
		assertTableExists(t, pool, "fx_rate")
		assertTableExists(t, pool, "schema_version")

		var version int
		if err := pool.QueryRow(ctx, `SELECT MAX(version) FROM schema_version`).Scan(&version); err != nil {
			t.Fatalf("read schema version: %v", err)
		}
		if version != 1 {
			t.Errorf("schema version = %d, want 1", version)
		}
	})

	t.Run("idempotent on second run", func(t *testing.T) {
		if err := store.RunMigrations(ctx, pool); err != nil {
			t.Fatalf("second migration run: %v", err)
		}

		var count int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_version`).Scan(&count); err != nil {
			t.Fatalf("count schema_version rows: %v", err)
		}
		if count != 1 {
			t.Errorf("schema_version rows = %d, want 1 (duplicate detected)", count)
		}
	})
}

func TestRunMigrations_CheckConstraints(t *testing.T) {
	pool := connectAndClean(t)
	defer pool.Close()
	ctx := context.Background()

	if err := store.RunMigrations(ctx, pool); err != nil {
		t.Fatalf("migration: %v", err)
	}

	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "negative open rejects",
			sql:  `INSERT INTO price_history (symbol, date, open, high, low, close, adj_close, volume, source) VALUES ('T', '2024-01-01', -1, 100, 90, 95, 95, 1000, 'test')`,
		},
		{
			name: "high less than low rejects",
			sql:  `INSERT INTO price_history (symbol, date, open, high, low, close, adj_close, volume, source) VALUES ('T', '2024-01-01', 100, 80, 90, 95, 95, 1000, 'test')`,
		},
		{
			name: "negative volume rejects",
			sql:  `INSERT INTO price_history (symbol, date, open, high, low, close, adj_close, volume, source) VALUES ('T', '2024-01-01', 100, 110, 90, 95, 95, -1, 'test')`,
		},
		{
			name: "negative fx rate rejects",
			sql:  `INSERT INTO fx_rate (pair, date, rate, source) VALUES ('USD/KRW', '2024-01-01', -1, 'test')`,
		},
		{
			name: "zero adj_close rejects",
			sql:  `INSERT INTO price_history (symbol, date, open, high, low, close, adj_close, volume, source) VALUES ('T', '2024-01-01', 100, 110, 90, 95, 0, 1000, 'test')`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, tt.sql)
			if err == nil {
				t.Error("expected CHECK constraint violation, got nil")
			}
		})
	}
}

func TestRunMigrations_ValidInsert(t *testing.T) {
	pool := connectAndClean(t)
	defer pool.Close()
	ctx := context.Background()

	if err := store.RunMigrations(ctx, pool); err != nil {
		t.Fatalf("migration: %v", err)
	}

	t.Run("price_history accepts valid row", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
			INSERT INTO price_history (symbol, date, open, high, low, close, adj_close, volume, source)
			VALUES ('NVDA', '2024-01-02', 495.22, 498.50, 492.10, 496.30, 496.30, 40000000, 'tiingo')
		`)
		if err != nil {
			t.Fatalf("valid price insert: %v", err)
		}
	})

	t.Run("fx_rate accepts valid row", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
			INSERT INTO fx_rate (pair, date, rate)
			VALUES ('USD/KRW', '2024-01-02', 1305.50)
		`)
		if err != nil {
			t.Fatalf("valid fx_rate insert: %v", err)
		}
	})

	t.Run("price_history upsert on duplicate PK", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
			INSERT INTO price_history (symbol, date, open, high, low, close, adj_close, volume, source)
			VALUES ('NVDA', '2024-01-02', 500.00, 510.00, 495.00, 505.00, 505.00, 45000000, 'tiingo')
			ON CONFLICT (symbol, date) DO UPDATE SET
				open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low,
				close = EXCLUDED.close, adj_close = EXCLUDED.adj_close,
				volume = EXCLUDED.volume, fetched_at = NOW()
		`)
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}

		var adjClose float64
		if err := pool.QueryRow(ctx, `SELECT adj_close FROM price_history WHERE symbol = 'NVDA' AND date = '2024-01-02'`).Scan(&adjClose); err != nil {
			t.Fatalf("read back: %v", err)
		}
		if adjClose != 505.00 {
			t.Errorf("adj_close after upsert = %f, want 505.00", adjClose)
		}
	})
}

func TestConnectDB_InvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := store.ConnectDB(ctx, "postgres://invalid:5432/nodb?connect_timeout=1")
	if err == nil {
		t.Error("expected error for unreachable database, got nil")
	}
}
