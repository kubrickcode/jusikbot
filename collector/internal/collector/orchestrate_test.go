package collector_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/collector"
)

func TestCollectAll_AllSuccess(t *testing.T) {
	sources := []collector.Source{
		{Name: "source-a", Collect: func(_ context.Context) error { return nil }},
		{Name: "source-b", Collect: func(_ context.Context) error { return nil }},
		{Name: "source-c", Collect: func(_ context.Context) error { return nil }},
	}

	results := collector.CollectAll(context.Background(), sources)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.IsOK() {
			t.Errorf("source %s should succeed, got: %v", r.Source, r.Error)
		}
		if r.Elapsed <= 0 {
			t.Errorf("source %s should have positive elapsed time", r.Source)
		}
	}
}

func TestCollectAll_PartialFailure(t *testing.T) {
	tokenErr := errors.New("token expired")
	sources := []collector.Source{
		{Name: "tiingo", Collect: func(_ context.Context) error { return nil }},
		{Name: "kis", Collect: func(_ context.Context) error { return tokenErr }},
		{Name: "fx", Collect: func(_ context.Context) error { return nil }},
	}

	results := collector.CollectAll(context.Background(), sources)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if !results[0].IsOK() {
		t.Errorf("tiingo should succeed, got: %v", results[0].Error)
	}
	if results[1].IsOK() {
		t.Error("kis should fail")
	}
	if !errors.Is(results[1].Error, tokenErr) {
		t.Errorf("kis error = %v, want %v", results[1].Error, tokenErr)
	}
	if !results[2].IsOK() {
		t.Errorf("fx should succeed, got: %v", results[2].Error)
	}
}

func TestCollectAll_ParallelExecution(t *testing.T) {
	delay := 100 * time.Millisecond
	sources := []collector.Source{
		{Name: "a", Collect: func(_ context.Context) error { time.Sleep(delay); return nil }},
		{Name: "b", Collect: func(_ context.Context) error { time.Sleep(delay); return nil }},
		{Name: "c", Collect: func(_ context.Context) error { time.Sleep(delay); return nil }},
	}

	started := time.Now()
	results := collector.CollectAll(context.Background(), sources)
	elapsed := time.Since(started)

	// Why 2x: if sequential, ~300ms. If parallel, ~100ms. 2x threshold is generous.
	if elapsed > 2*delay {
		t.Errorf("expected parallel execution (~%s), took %s", delay, elapsed)
	}

	for _, r := range results {
		if !r.IsOK() {
			t.Errorf("source %s failed: %v", r.Source, r.Error)
		}
	}
}

func TestCollectAll_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sources := []collector.Source{
		{Name: "a", Collect: func(ctx context.Context) error { return ctx.Err() }},
		{Name: "b", Collect: func(ctx context.Context) error { return ctx.Err() }},
	}

	results := collector.CollectAll(ctx, sources)

	for _, r := range results {
		if r.IsOK() {
			t.Errorf("source %s should fail with cancelled context", r.Source)
		}
	}
}

func TestCollectAll_Empty(t *testing.T) {
	results := collector.CollectAll(context.Background(), nil)

	if len(results) != 0 {
		t.Errorf("expected 0 results for nil sources, got %d", len(results))
	}
}

func TestCollectAll_PreservesSourceOrder(t *testing.T) {
	names := []string{"tiingo", "kis", "fx"}
	sources := make([]collector.Source, len(names))
	for i, n := range names {
		sources[i] = collector.Source{
			Name:    n,
			Collect: func(_ context.Context) error { return nil },
		}
	}

	results := collector.CollectAll(context.Background(), sources)

	for i, r := range results {
		if r.Source != names[i] {
			t.Errorf("results[%d].Source = %q, want %q", i, r.Source, names[i])
		}
	}
}

func TestAggregateErrors_NoErrors(t *testing.T) {
	results := []collector.SourceResult{
		{Source: "a"},
		{Source: "b"},
	}

	if err := collector.AggregateErrors(results); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAggregateErrors_JoinsMultipleErrors(t *testing.T) {
	results := []collector.SourceResult{
		{Source: "tiingo", Error: errors.New("timeout")},
		{Source: "kis"},
		{Source: "fx", Error: errors.New("connection refused")},
	}

	err := collector.AggregateErrors(results)
	if err == nil {
		t.Fatal("expected error")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "tiingo: timeout") {
		t.Errorf("error should contain 'tiingo: timeout', got: %s", errStr)
	}
	if !strings.Contains(errStr, "fx: connection refused") {
		t.Errorf("error should contain 'fx: connection refused', got: %s", errStr)
	}
}

func TestAggregateErrors_Empty(t *testing.T) {
	if err := collector.AggregateErrors(nil); err != nil {
		t.Errorf("expected nil for nil input, got %v", err)
	}
}
