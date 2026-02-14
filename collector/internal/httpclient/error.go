package httpclient

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const maxErrorBodyLength = 512

var (
	ErrRateLimited = errors.New("rate limited")
	ErrTimeout     = errors.New("request timed out")
)

// APIError represents an HTTP response indicating failure.
// Why Body is string: truncated for safe log inclusion, not raw bytes.
type APIError struct {
	Body        string
	IsRetryable bool
	StatusCode  int
	URL         string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("HTTP %d GET %s: %s", e.StatusCode, e.URL, e.Body)
	}
	return fmt.Sprintf("HTTP %d GET %s", e.StatusCode, e.URL)
}

// Unwrap returns sentinel errors for errors.Is matching.
// 429 → ErrRateLimited. Timeout errors are wrapped via fmt.Errorf in Client.Get, not via APIError.
func (e *APIError) Unwrap() error {
	if e.StatusCode == 429 {
		return ErrRateLimited
	}
	return nil
}

func truncateBody(body []byte) string {
	if len(body) <= maxErrorBodyLength {
		return string(body)
	}
	truncated := body[:maxErrorBodyLength]
	for !utf8.Valid(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}
	return string(truncated) + "..."
}
