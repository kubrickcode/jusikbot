package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

// SourceFunc collects data from a single source. Returns nil on success.
type SourceFunc func(ctx context.Context) error

// Source pairs a name with its collection function.
type Source struct {
	Collect SourceFunc
	Name    string
}

// SourceResult captures the outcome of a single source collection.
type SourceResult struct {
	Elapsed time.Duration
	Error   error
	Source  string
}

func (r SourceResult) IsOK() bool {
	return r.Error == nil
}

// CollectAll runs source functions in parallel using plain errgroup.
// Why plain errgroup (not WithContext): WithContext cancels all goroutines on first
// failure, preventing partial success. Plain errgroup lets all sources complete
// independently (plan.md design correction from data-collection-specialist).
func CollectAll(ctx context.Context, sources []Source) []SourceResult {
	results := make([]SourceResult, len(sources))
	var g errgroup.Group

	for i, src := range sources {
		results[i].Source = src.Name
		g.Go(func() error {
			started := time.Now()
			err := src.Collect(ctx)
			results[i].Elapsed = time.Since(started)
			results[i].Error = err
			return nil // Always nil: errors captured in results, not errgroup
		})
	}

	_ = g.Wait() // Always nil: errors captured in results, not errgroup
	return results
}

// ReportResults logs per-source outcomes and overall summary.
func ReportResults(results []SourceResult, totalElapsed time.Duration) {
	successCount := 0
	for _, r := range results {
		if r.IsOK() {
			successCount++
		}
	}

	for _, r := range results {
		if r.IsOK() {
			slog.Info(formatSourceSummary(r.Source, true, r.Elapsed))
		} else {
			slog.Error(formatSourceSummary(r.Source, false, r.Elapsed), "error", r.Error)
		}
	}

	slog.Info("collection complete",
		"elapsed", totalElapsed.Round(time.Millisecond).String(),
		"result", fmt.Sprintf("%d/%d OK", successCount, len(results)),
	)
}

// AggregateErrors joins all source errors into a single error.
// Returns nil when all sources succeeded.
func AggregateErrors(results []SourceResult) error {
	var errs []error
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.Source, r.Error))
		}
	}
	return errors.Join(errs...)
}

func formatSourceSummary(source string, isOK bool, elapsed time.Duration) string {
	status := "OK"
	if !isOK {
		status = "FAIL"
	}
	return fmt.Sprintf("%s: %s | %s", source, status, elapsed.Round(time.Millisecond))
}
