package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

var errTransient = errors.New("transient failure")
var errPermanent = errors.New("permanent failure")

func alwaysRetryable(err error) bool {
	return !errors.Is(err, errPermanent)
}

func TestWithRetry_ImmediateSuccess(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond}

	var calls int
	result, err := WithRetry(context.Background(), cfg, alwaysRetryable, func(ctx context.Context) (string, error) {
		calls++
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %q, want %q", result, "ok")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 5, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	var calls int
	result, err := WithRetry(context.Background(), cfg, alwaysRetryable, func(ctx context.Context) (int, error) {
		calls++
		if calls < 3 {
			return 0, errTransient
		}
		return 42, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("result = %d, want 42", result)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestWithRetry_MaxAttemptsExhausted(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	var calls int
	_, err := WithRetry(context.Background(), cfg, alwaysRetryable, func(ctx context.Context) (string, error) {
		calls++
		return "", errTransient
	})

	if !errors.Is(err, errTransient) {
		t.Errorf("error = %v, want errTransient", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3 (MaxAttempts)", calls)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 5, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	var calls int
	_, err := WithRetry(context.Background(), cfg, alwaysRetryable, func(ctx context.Context) (string, error) {
		calls++
		return "", errPermanent
	})

	if !errors.Is(err, errPermanent) {
		t.Errorf("error = %v, want errPermanent", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (should not retry)", calls)
	}
}

func TestWithRetry_ZeroMaxAttempts(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 0, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	_, err := WithRetry(context.Background(), cfg, nil, func(ctx context.Context) (string, error) {
		t.Fatal("fn should not be called with MaxAttempts=0")
		return "", nil
	})

	if err == nil {
		t.Fatal("expected error for MaxAttempts=0")
	}
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 10, InitialBackoff: time.Second, MaxBackoff: 10 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())

	var calls int
	_, err := WithRetry(ctx, cfg, alwaysRetryable, func(ctx context.Context) (string, error) {
		calls++
		if calls == 2 {
			cancel()
		}
		return "", errTransient
	})

	if err == nil {
		t.Fatal("expected error on context cancellation")
	}
	if calls > 3 {
		t.Errorf("calls = %d, should stop quickly after cancel", calls)
	}
}

func TestWithRetry_NilIsRetryable_RetriesAll(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	var calls int
	_, err := WithRetry[string](context.Background(), cfg, nil, func(ctx context.Context) (string, error) {
		calls++
		return "", errPermanent
	})

	if calls != 3 {
		t.Errorf("calls = %d, want 3 (nil isRetryable should retry all)", calls)
	}
	if !errors.Is(err, errPermanent) {
		t.Errorf("error = %v, want errPermanent", err)
	}
}

func TestFetchWithRateLimit_RespectsLimiter(t *testing.T) {
	// 100 req/sec limiter — fast enough for test, but ensures Wait is called.
	limiter := rate.NewLimiter(rate.Every(10*time.Millisecond), 1)
	cfg := RetryConfig{MaxAttempts: 1, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	result, err := FetchWithRateLimit(context.Background(), limiter, cfg, nil, func(ctx context.Context) (string, error) {
		return "fetched", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "fetched" {
		t.Errorf("result = %q, want %q", result, "fetched")
	}
}

func TestFetchWithRateLimit_RetriesWithRateLimiter(t *testing.T) {
	limiter := rate.NewLimiter(rate.Inf, 1)
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	var fnCalls int
	result, err := FetchWithRateLimit(context.Background(), limiter, cfg, alwaysRetryable, func(ctx context.Context) (string, error) {
		fnCalls++
		if fnCalls < 3 {
			return "", errTransient
		}
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %q, want %q", result, "ok")
	}
	if fnCalls != 3 {
		t.Errorf("fnCalls = %d, want 3", fnCalls)
	}
}

func TestFetchWithRateLimit_ContextCancelledDuringWait(t *testing.T) {
	// Very slow limiter — context should cancel before token becomes available.
	limiter := rate.NewLimiter(rate.Every(10*time.Second), 1)

	// Consume the burst token.
	limiter.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cfg := RetryConfig{MaxAttempts: 1, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	_, err := FetchWithRateLimit(ctx, limiter, cfg, nil, func(ctx context.Context) (string, error) {
		t.Fatal("fn should not be called when limiter.Wait is cancelled")
		return "", nil
	})

	if err == nil {
		t.Fatal("expected error on context cancellation during rate limit wait")
	}
}

func TestBackoffWithJitter_BoundsCheck(t *testing.T) {
	cfg := RetryConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxAttempts:    5,
		MaxBackoff:     2 * time.Second,
	}

	for attempt := range 10 {
		backoff := backoffWithJitter(cfg, attempt)
		if backoff < 0 {
			t.Errorf("attempt %d: negative backoff %v", attempt, backoff)
		}

		ceiling := float64(cfg.InitialBackoff) * float64(int64(1)<<min(attempt, 30))
		if ceiling > float64(cfg.MaxBackoff) {
			ceiling = float64(cfg.MaxBackoff)
		}
		if float64(backoff) > ceiling {
			t.Errorf("attempt %d: backoff %v exceeds ceiling %v", attempt, backoff, time.Duration(ceiling))
		}
	}
}
