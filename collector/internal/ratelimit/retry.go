package ratelimit

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"golang.org/x/time/rate"
)

// RetryConfig controls exponential backoff behavior.
type RetryConfig struct {
	InitialBackoff time.Duration
	MaxAttempts    int
	MaxBackoff     time.Duration
}

// WithRetry executes fn with exponential backoff + full jitter.
// isRetryable decides whether an error warrants retry; nil treats all errors as retryable.
// MaxAttempts is total invocations (1 = no retry).
func WithRetry[T any](
	ctx context.Context,
	cfg RetryConfig,
	isRetryable func(error) bool,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	var zero T

	if cfg.MaxAttempts < 1 {
		return zero, fmt.Errorf("RetryConfig.MaxAttempts must be >= 1, got %d", cfg.MaxAttempts)
	}

	var lastErr error

	for attempt := range cfg.MaxAttempts {
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Context cancelled → abort immediately, regardless of isRetryable.
		if ctx.Err() != nil {
			return zero, lastErr
		}

		if isRetryable != nil && !isRetryable(err) {
			break
		}

		backoff := backoffWithJitter(cfg, attempt)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}

	return zero, lastErr
}

// FetchWithRateLimit combines rate.Limiter.Wait + WithRetry.
// The limiter is waited on BEFORE each attempt, ensuring retries also respect rate limits.
func FetchWithRateLimit[T any](
	ctx context.Context,
	limiter *rate.Limiter,
	cfg RetryConfig,
	isRetryable func(error) bool,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	rateLimitedFn := func(ctx context.Context) (T, error) {
		var zero T
		if err := limiter.Wait(ctx); err != nil {
			return zero, err
		}
		return fn(ctx)
	}

	return WithRetry(ctx, cfg, isRetryable, rateLimitedFn)
}

// backoffWithJitter computes full jitter: rand * min(maxBackoff, initialBackoff * 2^attempt).
// Why cap at 62: float64 precision degrades beyond 2^63, causing overflow.
func backoffWithJitter(cfg RetryConfig, attempt int) time.Duration {
	ceiling := float64(cfg.InitialBackoff) * math.Pow(2, float64(min(attempt, 62)))
	if ceiling > float64(cfg.MaxBackoff) {
		ceiling = float64(cfg.MaxBackoff)
	}
	return time.Duration(rand.Float64() * ceiling)
}
