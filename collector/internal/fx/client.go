package fx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/httpclient"
)

const sourceName = "frankfurter"

type frankfurterResponse struct {
	Amount    float64                          `json:"amount"`
	Base      string                           `json:"base"`
	EndDate   string                           `json:"end_date"`
	Rates     map[string]map[string]float64    `json:"rates"`
	StartDate string                           `json:"start_date"`
}

// Client wraps an httpclient.Client configured for the Frankfurter API.
type Client struct {
	http *httpclient.Client
}

// NewClient creates a Frankfurter API client.
// Why no auth: Frankfurter API is free and requires no authentication.
func NewClient(httpClient *httpclient.Client) *Client {
	return &Client{http: httpClient}
}

// FetchRates returns rates sorted by date ascending.
func (c *Client) FetchRates(ctx context.Context, base, target string, from, to time.Time) ([]domain.FXRate, error) {
	path := fmt.Sprintf("/v1/%s..%s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	body, _, err := c.http.Get(ctx, path,
		httpclient.WithQueryParam("from", base),
		httpclient.WithQueryParam("to", target),
	)
	if err != nil {
		return nil, fmt.Errorf("fetch fx rates %s/%s: %w", base, target, err)
	}

	var resp frankfurterResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse frankfurter response: %w", err)
	}

	pair := base + "/" + target
	rates := make([]domain.FXRate, 0, len(resp.Rates))

	for dateStr, currencies := range resp.Rates {
		rate, ok := currencies[target]
		if !ok {
			return nil, fmt.Errorf("target currency %s missing in rates for %s", target, dateStr)
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date %q: %w", dateStr, err)
		}

		rates = append(rates, domain.FXRate{
			Date:   date,
			Pair:   pair,
			Rate:   rate,
			Source: sourceName,
		})
	}

	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Date.Before(rates[j].Date)
	})

	return rates, nil
}

// IsRetryable classifies errors for retry decisions.
// Why not retry 4xx: client errors indicate permanent failures (bad request, not found).
func IsRetryable(err error) bool {
	var apiErr *httpclient.APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable
	}
	return false
}
