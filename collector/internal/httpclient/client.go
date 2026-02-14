package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultMaxBodySize = 10 * 1024 * 1024
	defaultTimeout     = 30 * time.Second
)

// ErrBodyTooLarge signals a response body exceeding the configured limit.
var ErrBodyTooLarge = errors.New("response body too large")

// Client wraps net/http with base URL, default headers, and safety limits.
type Client struct {
	baseURL     string
	headers     map[string]string
	httpClient  *http.Client
	maxBodySize int64
}

// NewClient creates a Client. Pass nil httpClient for defaults.
// Why explicit *http.Client: enables test doubles via httptest.
func NewClient(baseURL string, headers map[string]string, httpClient *http.Client, maxBodySize int64) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if maxBodySize <= 0 {
		maxBodySize = defaultMaxBodySize
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	return &Client{
		baseURL:     baseURL,
		headers:     headers,
		httpClient:  httpClient,
		maxBodySize: maxBodySize,
	}
}

// RequestOption customizes a single request.
type RequestOption func(*requestConfig)

type requestConfig struct {
	headers     map[string]string
	queryParams map[string]string
}

func WithHeader(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.headers[key] = value
	}
}

func WithQueryParam(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.queryParams[key] = value
	}
}

// Get sends an HTTP GET and returns body bytes, status code, and error.
// 2xx → (body, status, nil), 4xx → permanent error, 5xx → retryable error.
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) ([]byte, int, error) {
	cfg := &requestConfig{
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	reqURL, err := c.buildURL(path, cfg.queryParams)
	if err != nil {
		return nil, 0, fmt.Errorf("build request URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, 0, fmt.Errorf("request to %s: %w", reqURL, ErrTimeout)
		}
		if ctx.Err() != nil {
			return nil, 0, fmt.Errorf("request to %s: %w", reqURL, ctx.Err())
		}
		return nil, 0, fmt.Errorf("request to %s: %w", reqURL, err)
	}
	defer resp.Body.Close()

	// Why +1: detect overflow without separate HEAD request.
	limited := io.LimitReader(resp.Body, c.maxBodySize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body from %s: %w", reqURL, err)
	}

	if int64(len(body)) > c.maxBodySize {
		return nil, resp.StatusCode, fmt.Errorf("response from %s (%d bytes): %w", reqURL, len(body), ErrBodyTooLarge)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, resp.StatusCode, nil
	}

	apiErr := &APIError{
		Body:        truncateBody(body),
		IsRetryable: resp.StatusCode == 429 || resp.StatusCode >= 500,
		StatusCode:  resp.StatusCode,
		URL:         reqURL,
	}
	return nil, resp.StatusCode, apiErr
}

func (c *Client) buildURL(path string, queryParams map[string]string) (string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL %q: %w", c.baseURL, err)
	}

	ref, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("parse path %q: %w", path, err)
	}

	resolved := base.ResolveReference(ref)

	if len(queryParams) > 0 {
		q := resolved.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		resolved.RawQuery = q.Encode()
	}

	return resolved.String(), nil
}
