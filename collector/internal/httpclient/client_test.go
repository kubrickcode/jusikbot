package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token test-key" {
			t.Errorf("missing default header, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"close":150.0}]`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, map[string]string{"Authorization": "Token test-key"}, srv.Client(), 0)

	body, status, err := client.Get(context.Background(), "/prices")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
	if string(body) != `[{"close":150.0}]` {
		t.Errorf("body = %q, want JSON array", string(body))
	}
}

func TestGet_DefaultAndPerRequestHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token default" {
			t.Errorf("default header missing: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Custom") != "override" {
			t.Errorf("per-request header missing: %q", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, map[string]string{"Authorization": "Token default"}, srv.Client(), 0)

	_, _, err := client.Get(context.Background(), "/test", WithHeader("X-Custom", "override"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("startDate") != "2024-01-01" {
			t.Errorf("query param startDate = %q", r.URL.Query().Get("startDate"))
		}
		if r.URL.Query().Get("format") != "json" {
			t.Errorf("query param format = %q", r.URL.Query().Get("format"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, nil, srv.Client(), 0)
	_, _, err := client.Get(context.Background(), "/data",
		WithQueryParam("startDate", "2024-01-01"),
		WithQueryParam("format", "json"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_4xx_PermanentError(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"bad request", http.StatusBadRequest},
		{"unauthorized", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
		{"not found", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte("error detail"))
			}))
			defer srv.Close()

			client := NewClient(srv.URL, nil, srv.Client(), 0)
			_, status, err := client.Get(context.Background(), "/fail")

			if err == nil {
				t.Fatal("expected error for 4xx")
			}
			if status != tt.status {
				t.Errorf("status = %d, want %d", status, tt.status)
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatal("error should be *APIError")
			}
			if apiErr.IsRetryable {
				t.Error("4xx should not be retryable")
			}
		})
	}
}

func TestGet_429_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limit exceeded"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, nil, srv.Client(), 0)
	_, status, err := client.Get(context.Background(), "/limited")

	if status != 429 {
		t.Errorf("status = %d, want 429", status)
	}
	if !errors.Is(err, ErrRateLimited) {
		t.Error("429 should unwrap to ErrRateLimited")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("error should be *APIError")
	}
	if !apiErr.IsRetryable {
		t.Error("429 should be retryable")
	}
}

func TestGet_5xx_RetryableError(t *testing.T) {
	tests := []int{500, 502, 503, 504}
	for _, code := range tests {
		t.Run(http.StatusText(code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
				w.Write([]byte("server error"))
			}))
			defer srv.Close()

			client := NewClient(srv.URL, nil, srv.Client(), 0)
			_, status, err := client.Get(context.Background(), "/error")

			if status != code {
				t.Errorf("status = %d, want %d", status, code)
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatal("error should be *APIError")
			}
			if !apiErr.IsRetryable {
				t.Errorf("status %d should be retryable", code)
			}
		})
	}
}

func TestGet_ContextDeadlineExceeded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	client := NewClient(srv.URL, nil, srv.Client(), 0)
	_, _, err := client.Get(ctx, "/slow")

	if err == nil {
		t.Fatal("expected error on context timeout")
	}
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("error should wrap ErrTimeout, got: %v", err)
	}
}

func TestGet_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	client := NewClient(srv.URL, nil, srv.Client(), 0)
	_, _, err := client.Get(ctx, "/slow")

	if err == nil {
		t.Fatal("expected error on context cancellation")
	}
	if errors.Is(err, ErrTimeout) {
		t.Error("explicit cancel should NOT wrap ErrTimeout")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error should wrap context.Canceled, got: %v", err)
	}
}

func TestGet_BodySizeExceeded(t *testing.T) {
	largeBody := strings.Repeat("x", 1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, nil, srv.Client(), 512)
	_, _, err := client.Get(context.Background(), "/large")

	if err == nil {
		t.Fatal("expected error for oversized body")
	}
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Errorf("error should wrap ErrBodyTooLarge, got: %v", err)
	}
}

func TestGet_BodyWithinLimit(t *testing.T) {
	exactBody := strings.Repeat("x", 512)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(exactBody))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, nil, srv.Client(), 512)
	body, _, err := client.Get(context.Background(), "/exact")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) != 512 {
		t.Errorf("body length = %d, want 512", len(body))
	}
}
