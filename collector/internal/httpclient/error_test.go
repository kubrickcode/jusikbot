package httpclient

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		wantSub  string
	}{
		{
			name:    "with body",
			err:     &APIError{StatusCode: 500, URL: "https://api.example.com/data", Body: "internal error"},
			wantSub: "HTTP 500 GET https://api.example.com/data: internal error",
		},
		{
			name:    "without body",
			err:     &APIError{StatusCode: 404, URL: "https://api.example.com/data"},
			wantSub: "HTTP 404 GET https://api.example.com/data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantSub {
				t.Errorf("Error() = %q, want %q", got, tt.wantSub)
			}
		})
	}
}

func TestAPIError_Unwrap_RateLimited(t *testing.T) {
	err := &APIError{StatusCode: 429, URL: "https://api.example.com", IsRetryable: true}
	if !errors.Is(err, ErrRateLimited) {
		t.Error("429 APIError should unwrap to ErrRateLimited")
	}
}

func TestAPIError_Unwrap_NonRateLimited(t *testing.T) {
	tests := []int{400, 401, 403, 404, 500, 503}
	for _, code := range tests {
		err := &APIError{StatusCode: code, URL: "https://example.com"}
		if errors.Is(err, ErrRateLimited) {
			t.Errorf("status %d should not unwrap to ErrRateLimited", code)
		}
	}
}

func TestAPIError_ErrorsAs(t *testing.T) {
	original := &APIError{StatusCode: 503, URL: "https://api.example.com", IsRetryable: true, Body: "unavailable"}
	wrapped := fmt.Errorf("fetch prices: %w", original)

	var target *APIError
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should extract APIError from wrapped chain")
	}
	if target.StatusCode != 503 {
		t.Errorf("StatusCode = %d, want 503", target.StatusCode)
	}
	if !target.IsRetryable {
		t.Error("IsRetryable should be true for 503")
	}
}

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantLen int
		hasDots bool
	}{
		{
			name:    "empty",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "under limit",
			input:   []byte("short body"),
			wantLen: 10,
		},
		{
			name:    "exact limit",
			input:   []byte(strings.Repeat("a", maxErrorBodyLength)),
			wantLen: maxErrorBodyLength,
		},
		{
			name:    "over limit",
			input:   []byte(strings.Repeat("a", maxErrorBodyLength+100)),
			wantLen: maxErrorBodyLength + 3, // "..."
			hasDots: true,
		},
		{
			name:    "multibyte utf8 boundary",
			input:   append([]byte(strings.Repeat("a", maxErrorBodyLength-1)), []byte("日本語テスト")...),
			hasDots: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateBody(tt.input)
			if tt.wantLen > 0 && len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.hasDots && !strings.HasSuffix(got, "...") {
				t.Error("truncated body should end with ...")
			}
		})
	}
}
