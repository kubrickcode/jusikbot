package kis

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

const (
	dailyPricePath = "/uapi/domestic-stock/v1/quotations/inquire-daily-itemchartprice"
	maxPages       = 10
	trIDDailyChart = "FHKST03010100"
)

var ErrMaxPagesReached = errors.New("max pagination pages reached")

// kisResponse represents the KIS daily chart price API response.
type kisResponse struct {
	MsgCode string         `json:"msg_cd"`
	Msg     string         `json:"msg1"`
	Output2 []kisOutputRow `json:"output2"`
	RtCode  string         `json:"rt_cd"`
}

// kisCursor tracks date-based pagination state.
// Why date-based instead of tr_cont headers: httpclient.Client does not expose response
// headers. Date-range adjustment achieves the same result for daily chart data.
type kisCursor struct {
	endDate time.Time
	hasMore bool
}

// Client wraps an httpclient.Client configured for the KIS API.
type Client struct {
	http  *httpclient.Client
	token *TokenProvider
}

func NewClient(httpClient *httpclient.Client, tokenProvider *TokenProvider) *Client {
	return &Client{
		http:  httpClient,
		token: tokenProvider,
	}
}

// FetchDailyPrices fetches all pages of daily prices for a symbol within the date range.
// Returns data sorted ascending by date. Implements domain.StockDataFetcher.
func (c *Client) FetchDailyPrices(ctx context.Context, symbol string, from, to time.Time) ([]domain.DailyPrice, error) {
	var allPrices []domain.DailyPrice
	cursor := kisCursor{endDate: to, hasMore: true}

	for page := range maxPages {
		if !cursor.hasMore {
			break
		}

		prevEndDate := cursor.endDate
		prices, nextCursor, err := c.fetchDailyPricesPage(ctx, symbol, from, cursor.endDate)
		if err != nil {
			return allPrices, fmt.Errorf("page %d for %s: %w", page, symbol, err)
		}

		allPrices = append(allPrices, prices...)
		cursor = nextCursor

		// Why break on stale cursor: prevents redundant API calls when a symbol
		// has no data before the current page boundary.
		if cursor.hasMore && cursor.endDate.Equal(prevEndDate) {
			cursor.hasMore = false
			break
		}
	}

	if cursor.hasMore {
		return allPrices, fmt.Errorf("symbol %s: %w", symbol, ErrMaxPagesReached)
	}

	sort.Slice(allPrices, func(i, j int) bool {
		return allPrices[i].Date.Before(allPrices[j].Date)
	})

	return allPrices, nil
}

func (c *Client) fetchDailyPricesPage(
	ctx context.Context,
	symbol string,
	from, to time.Time,
) ([]domain.DailyPrice, kisCursor, error) {
	accessToken, err := c.token.Token(ctx)
	if err != nil {
		return nil, kisCursor{}, fmt.Errorf("obtain token: %w", err)
	}

	body, _, err := c.http.Get(ctx, dailyPricePath,
		httpclient.WithHeader("authorization", "Bearer "+accessToken),
		httpclient.WithHeader("tr_id", trIDDailyChart),
		httpclient.WithQueryParam("FID_COND_MRKT_DIV_CODE", "J"),
		httpclient.WithQueryParam("FID_INPUT_DATE_1", from.Format("20060102")),
		httpclient.WithQueryParam("FID_INPUT_DATE_2", to.Format("20060102")),
		httpclient.WithQueryParam("FID_INPUT_ISCD", symbol),
		httpclient.WithQueryParam("FID_ORG_ADJ_PRC", "0"),
		httpclient.WithQueryParam("FID_PERIOD_DIV_CODE", "D"),
	)
	if err != nil {
		return nil, kisCursor{}, err
	}

	var resp kisResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, kisCursor{}, fmt.Errorf("parse KIS response for %s: %w", symbol, err)
	}

	if resp.RtCode != "0" {
		return nil, kisCursor{}, fmt.Errorf("KIS API error for %s (code=%s): %s", symbol, resp.MsgCode, resp.Msg)
	}

	prices, err := parseOutputRows(resp.Output2, symbol)
	if err != nil {
		return nil, kisCursor{}, err
	}

	nextCursor := buildNextCursor(prices, from)
	return prices, nextCursor, nil
}

// buildNextCursor determines if more pages are needed.
// KIS returns data newest-first. If the oldest row's date is still after `from`,
// there may be more data. Set endDate to one day before the oldest date.
func buildNextCursor(prices []domain.DailyPrice, from time.Time) kisCursor {
	if len(prices) == 0 {
		return kisCursor{hasMore: false}
	}

	oldestDate := prices[0].Date
	for _, p := range prices[1:] {
		if p.Date.Before(oldestDate) {
			oldestDate = p.Date
		}
	}

	if oldestDate.After(from) {
		return kisCursor{
			endDate: oldestDate.AddDate(0, 0, -1),
			hasMore: true,
		}
	}

	return kisCursor{hasMore: false}
}

func parseOutputRows(rows []kisOutputRow, symbol string) ([]domain.DailyPrice, error) {
	prices := make([]domain.DailyPrice, 0, len(rows))
	for i, row := range rows {
		price, ok, err := toDailyPrice(row, symbol)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}
		if !ok {
			continue
		}
		prices = append(prices, price)
	}
	return prices, nil
}

// IsRetryable determines whether an error from the KIS client warrants retry.
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
