package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jusikbot/collector/internal/collector"
	"github.com/jusikbot/collector/internal/config"
	"github.com/jusikbot/collector/internal/store"
	"github.com/jusikbot/collector/internal/summary"
)

const (
	summaryOutputPath = "../data/summary.md"
	watchlistPath     = "config/watchlist.json"
)

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

	sc := &sourceCollector{
		env:       env,
		repo:      store.NewRepository(pool),
		watchlist: watchlist,
	}

	sources := sc.buildSources(target)
	results := collector.CollectAll(ctx, sources)
	collector.ReportResults(results, time.Since(started))

	// Intent: summary는 부가 출력이므로 실패해도 수집 exit code에 반영하지 않음.
	absOutputPath, _ := filepath.Abs(summaryOutputPath)
	if err := summary.GenerateSummary(ctx, sc.repo, watchlist, summaryOutputPath); err != nil {
		slog.Error("summary generation failed", "error", err, "path", absOutputPath)
	} else {
		slog.Info("summary generated", "path", absOutputPath)
	}

	return collector.AggregateErrors(results)
}

func resolveTargets(target string) []string {
	if target == "all" {
		return []string{"tiingo", "kis", "fx"}
	}
	return []string{target}
}

func (c *sourceCollector) buildSources(target string) []collector.Source {
	targets := resolveTargets(target)
	sources := make([]collector.Source, 0, len(targets))

	for _, t := range targets {
		switch t {
		case "tiingo":
			sources = append(sources, collector.Source{Name: "tiingo", Collect: c.collectTiingo})
		case "kis":
			sources = append(sources, collector.Source{Name: "kis", Collect: c.collectKIS})
		case "fx":
			sources = append(sources, collector.Source{Name: "fx", Collect: c.collectFX})
		default:
			slog.Warn("unknown collection target, skipping", "target", t)
		}
	}

	return sources
}
