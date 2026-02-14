package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jusikbot/collector/internal/config"
	"github.com/jusikbot/collector/internal/store"
)

const watchlistPath = "config/watchlist.json"

func main() {
	target, dryRun := parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, target, dryRun); err != nil {
		slog.Error("collector failed", "error", err)
		os.Exit(1)
	}
}

func parseFlags() (string, bool) {
	target := flag.String("target", "all", "collection target: tiingo, kis, fx, all")
	dryRun := flag.Bool("dry-run", false, "validate configuration without collecting data")
	flag.Parse()

	validTargets := map[string]bool{
		"all":    true,
		"fx":     true,
		"kis":    true,
		"tiingo": true,
	}

	if !validTargets[*target] {
		fmt.Fprintf(os.Stderr, "invalid target %q (allowed: tiingo, kis, fx, all)\n", *target)
		os.Exit(1)
	}

	return *target, *dryRun
}

func run(ctx context.Context, target string, dryRun bool) error {
	started := time.Now()
	slog.Info("collector starting", "dry_run", dryRun, "target", target)

	env, err := config.LoadEnv()
	if err != nil {
		return fmt.Errorf("load environment config: %w", err)
	}

	watchlist, err := config.LoadWatchlist(watchlistPath)
	if err != nil {
		return fmt.Errorf("load watchlist: %w", err)
	}
	slog.Info("watchlist loaded", "entries", len(watchlist))

	if dryRun {
		slog.Info("dry-run mode: skipping DB connection and collection")
		return nil
	}

	pool, err := store.ConnectDB(ctx, env.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	if err := store.RunMigrations(ctx, pool); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	targets := resolveTargets(target)
	results := make([]SourceResult, 0, len(targets))

	for _, t := range targets {
		if ctx.Err() != nil {
			slog.Warn("cancelled before collecting", "source", t)
			results = append(results, SourceResult{
				Elapsed: 0,
				Error:   ctx.Err(),
				Source:  t,
			})
			continue
		}

		result := collectSource(ctx, t)
		results = append(results, result)
	}

	reportResults(results, time.Since(started))
	return nil
}

func resolveTargets(target string) []string {
	if target == "all" {
		return []string{"tiingo", "kis", "fx"}
	}
	return []string{target}
}

func collectSource(ctx context.Context, source string) SourceResult {
	started := time.Now()

	switch source {
	case "tiingo":
		slog.Warn("not implemented yet", "source", "tiingo")
	case "kis":
		slog.Warn("not implemented yet", "source", "kis")
	case "fx":
		slog.Warn("not implemented yet", "source", "fx")
	}

	return SourceResult{
		Elapsed: time.Since(started),
		Error:   nil,
		Source:  source,
	}
}
