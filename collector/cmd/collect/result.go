package main

import (
	"fmt"
	"log/slog"
	"time"
)

type SourceResult struct {
	Elapsed time.Duration
	Error   error
	Source  string
}

func (r SourceResult) isOK() bool {
	return r.Error == nil
}

func reportResults(results []SourceResult, totalElapsed time.Duration) {
	successCount := 0
	for _, r := range results {
		if r.isOK() {
			successCount++
		}
	}

	for _, r := range results {
		if r.isOK() {
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

func formatSourceSummary(source string, isOK bool, elapsed time.Duration) string {
	status := "OK"
	if !isOK {
		status = "FAIL"
	}
	return fmt.Sprintf("%s: %s | %s", source, status, elapsed.Round(time.Millisecond))
}
