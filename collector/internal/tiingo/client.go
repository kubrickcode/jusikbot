package tiingo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/httpclient"
)

const sourceName = "tiingo"

// ErrTickerInvalid signals that the requested symbol does not exist on Tiingo.
var ErrTickerInvalid = errors.New("ticker not found on tiingo")

// tiingoPrice represents a single row from the Tiingo daily prices API.
// Why float64 for Volume: Tiingo may serialize integer volumes as floats in JSON.
type tiingoPrice struct {
	AdjClose    float64 `json:"adjClose"`
	Close       float64 `json:"close"`
	Date        string  `json:"date"`
	DivCash     float64 `json:"divCash"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Open        float64 `json:"open"`
	SplitFactor float64 `json:"splitFactor"`
	Volume      float64 `json:"volume"`
}

// Client wraps an httpclient.Client configured for the Tiingo API.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a Tiingo API client.
// The httpClient must be pre-configured with base URL and Authorization header.
func NewClient(httpClient *httpclient.Client) *Client {
	return &Client{http: httpClient}
}

// fetchPrices calls the Tiingo daily prices API and returns the raw parsed response.
// Rate limit detection: Tiingo returns HTTP 200 with non-JSON body when rate limited.
// Valid JSON arrays start with '[' after trimming whitespace.
func (c *Client) fetchPrices(ctx context.Context, symbol string, from, to time.Time) ([]tiingoPrice, error) {
	path := fmt.Sprintf("/tiingo/daily/%s/prices", symbol)

	body, _, err := c.http.Get(ctx, path,
		httpclient.WithQueryParam("startDate", from.Format("2006-01-02")),
		httpclient.WithQueryParam("endDate", to.Format("2006-01-02")),
	)
	if err != nil {
		var apiErr *httpclient.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return nil, fmt.Errorf("symbol %s: %w", symbol, ErrTickerInvalid)
		}
		return nil, fmt.Errorf("fetch prices for %s: %w", symbol, err)
	}

	trimmed := bytes.TrimLeftFunc(body, unicode.IsSpace)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, fmt.Errorf("symbol %s: unexpected response body: %w", symbol, httpclient.ErrRateLimited)
	}

	var prices []tiingoPrice
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, fmt.Errorf("parse tiingo response for %s: %w", symbol, err)
	}

	return prices, nil
}

// IsRetryable determines whether an error from the Tiingo client warrants retry.
// Retryable: rate limiting (HTTP 429 or body-level), server errors (5xx).
// Non-retryable: invalid ticker (404), parse errors.
func IsRetryable(err error) bool {
	if errors.Is(err, httpclient.ErrRateLimited) {
		return true
	}
	var apiErr *httpclient.APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable
	}
	return false
}

func toDailyPrice(r tiingoPrice, symbol string) (domain.DailyPrice, error) {
	date, err := parseTiingoDate(r.Date)
	if err != nil {
		return domain.DailyPrice{}, fmt.Errorf("parse date %q: %w", r.Date, err)
	}

	return domain.DailyPrice{
		AdjClose: r.AdjClose,
		Close:    r.Close,
		Date:     date,
		High:     r.High,
		Low:      r.Low,
		Open:     r.Open,
		Source:   sourceName,
		Symbol:   symbol,
		Volume:   int64(r.Volume),
	}, nil
}

// parseTiingoDate handles both RFC3339 ("...+00:00") and fractional seconds ("...000Z") formats.
func parseTiingoDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}
