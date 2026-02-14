package kis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jusikbot/collector/internal/domain"
	"github.com/jusikbot/collector/internal/httpclient"
)

func newStubTokenProvider(t *testing.T, token string) *TokenProvider {
	t.Helper()
	srv := httptest.NewServer(validTokenHandler(token, 86400))
	t.Cleanup(srv.Close)
	return NewTokenProvider(srv.URL, "test-key", "test-secret", srv.Client())
}

func newTestKISClient(t *testing.T, srv *httptest.Server, tokenProvider *TokenProvider) *Client {
	t.Helper()
	hc := httpclient.NewClient(
		srv.URL,
		map[string]string{
			"appkey":    "test-key",
			"appsecret": "test-secret",
		},
		srv.Client(),
		0,
	)
	return NewClient(hc, tokenProvider)
}

func kisSuccessResponse(rows []kisOutputRow) kisResponse {
	return kisResponse{
		MsgCode: "MCA00000",
		Msg:     "정상처리 되었습니다.",
		Output2: rows,
		RtCode:  "0",
	}
}

func TestFetchDailyPrices(t *testing.T) {
	t.Run("single page response", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "test-bearer-token")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != dailyPricePath {
				t.Errorf("path = %q, want %s", r.URL.Path, dailyPricePath)
			}
			if got := r.Header.Get("authorization"); got != "Bearer test-bearer-token" {
				t.Errorf("authorization = %q, want Bearer test-bearer-token", got)
			}
			if got := r.Header.Get("tr_id"); got != trIDDailyChart {
				t.Errorf("tr_id = %q, want %s", got, trIDDailyChart)
			}
			if got := r.Header.Get("appkey"); got != "test-key" {
				t.Errorf("appkey = %q, want test-key", got)
			}
			if got := r.URL.Query().Get("FID_INPUT_ISCD"); got != "005930" {
				t.Errorf("FID_INPUT_ISCD = %q, want 005930", got)
			}
			if got := r.URL.Query().Get("FID_ORG_ADJ_PRC"); got != "0" {
				t.Errorf("FID_ORG_ADJ_PRC = %q, want 0", got)
			}

			resp := kisSuccessResponse([]kisOutputRow{
				{StckBsopDate: "20240116", StckOprc: "72000", StckHgpr: "72500", StckLwpr: "71500", StckClpr: "72200", AcmlVol: "15000000"},
				{StckBsopDate: "20240115", StckOprc: "71000", StckHgpr: "72000", StckLwpr: "70500", StckClpr: "71800", AcmlVol: "12000000"},
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

		prices, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 2 {
			t.Fatalf("len(prices) = %d, want 2", len(prices))
		}

		// Verify ascending sort order
		if !prices[0].Date.Before(prices[1].Date) {
			t.Errorf("prices not sorted ascending: %v >= %v", prices[0].Date, prices[1].Date)
		}

		if prices[0].Close != 71800 {
			t.Errorf("prices[0].Close = %v, want 71800", prices[0].Close)
		}
		if prices[1].Close != 72200 {
			t.Errorf("prices[1].Close = %v, want 72200", prices[1].Close)
		}
		if prices[0].Symbol != "005930" {
			t.Errorf("Symbol = %q, want 005930", prices[0].Symbol)
		}
		if prices[0].Source != "kis" {
			t.Errorf("Source = %q, want kis", prices[0].Source)
		}
	})

	t.Run("multi-page pagination", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "token")

		var pageCount int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pageCount++
			var resp kisResponse

			endDate := r.URL.Query().Get("FID_INPUT_DATE_2")

			switch {
			case endDate == "20240331":
				// Page 1: dates 20240301 - 20240331 (oldest > from)
				resp = kisSuccessResponse([]kisOutputRow{
					{StckBsopDate: "20240331", StckOprc: "100", StckHgpr: "100", StckLwpr: "100", StckClpr: "100", AcmlVol: "1000"},
					{StckBsopDate: "20240215", StckOprc: "98", StckHgpr: "99", StckLwpr: "97", StckClpr: "98", AcmlVol: "900"},
				})
			case endDate == "20240214":
				// Page 2: dates from start range (oldest <= from)
				resp = kisSuccessResponse([]kisOutputRow{
					{StckBsopDate: "20240201", StckOprc: "95", StckHgpr: "96", StckLwpr: "94", StckClpr: "95", AcmlVol: "800"},
					{StckBsopDate: "20240115", StckOprc: "90", StckHgpr: "91", StckLwpr: "89", StckClpr: "90", AcmlVol: "700"},
				})
			default:
				resp = kisSuccessResponse(nil)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

		prices, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 4 {
			t.Fatalf("len(prices) = %d, want 4", len(prices))
		}
		if pageCount != 2 {
			t.Errorf("pages fetched = %d, want 2", pageCount)
		}

		// Verify ascending order after merge
		for i := 1; i < len(prices); i++ {
			if !prices[i-1].Date.Before(prices[i].Date) {
				t.Errorf("prices[%d].Date (%v) >= prices[%d].Date (%v)",
					i-1, prices[i-1].Date, i, prices[i].Date)
			}
		}
	})

	t.Run("empty response", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "token")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := kisSuccessResponse(nil)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		prices, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 0 {
			t.Errorf("len(prices) = %d, want 0", len(prices))
		}
	})

	t.Run("KIS API error", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "token")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := kisResponse{
				MsgCode: "EGW00123",
				Msg:     "유효하지 않은 토큰입니다.",
				RtCode:  "1",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

		_, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err == nil {
			t.Fatal("expected error for KIS API error response")
		}
	})

	t.Run("empty rows filtered out", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "token")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := kisSuccessResponse([]kisOutputRow{
				{StckBsopDate: "20240115", StckOprc: "100", StckHgpr: "100", StckLwpr: "100", StckClpr: "100", AcmlVol: "1000"},
				{StckBsopDate: "", StckOprc: "", StckHgpr: "", StckLwpr: "", StckClpr: "", AcmlVol: ""},
				{StckBsopDate: "0", StckOprc: "0", StckHgpr: "0", StckLwpr: "0", StckClpr: "0", AcmlVol: "0"},
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		prices, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prices) != 1 {
			t.Errorf("len(prices) = %d, want 1 (empty rows filtered)", len(prices))
		}
	})

	t.Run("max pages safety", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "token")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return the endDate as row date so cursor advances (no stale) but never reaches from
			endDate := r.URL.Query().Get("FID_INPUT_DATE_2")
			resp := kisSuccessResponse([]kisOutputRow{
				{StckBsopDate: endDate, StckOprc: "100", StckHgpr: "100", StckLwpr: "100", StckClpr: "100", AcmlVol: "1000"},
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

		_, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err == nil {
			t.Fatal("expected error for max pages exceeded")
		}
		if !errors.Is(err, ErrMaxPagesReached) {
			t.Errorf("error should wrap ErrMaxPagesReached, got: %v", err)
		}
	})

	t.Run("stale cursor breaks pagination early", func(t *testing.T) {
		stubTP := newStubTokenProvider(t, "token")

		var pageCount int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pageCount++
			// Always return a fixed date regardless of endDate param.
			// After page 1, cursor becomes 20240531 and stays there → stale.
			resp := kisSuccessResponse([]kisOutputRow{
				{StckBsopDate: "20240601", StckOprc: "100", StckHgpr: "100", StckLwpr: "100", StckClpr: "100", AcmlVol: "1000"},
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		client := newTestKISClient(t, srv, stubTP)
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

		prices, err := client.FetchDailyPrices(context.Background(), "005930", from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pageCount != 2 {
			t.Errorf("pageCount = %d, want 2 (stale cursor should break after page 2)", pageCount)
		}
		if len(prices) != 2 {
			t.Errorf("len(prices) = %d, want 2 (one from each page before break)", len(prices))
		}
	})
}

func TestBuildNextCursor(t *testing.T) {
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("empty prices stops pagination", func(t *testing.T) {
		cursor := buildNextCursor(nil, from)
		if cursor.hasMore {
			t.Error("hasMore = true, want false for empty prices")
		}
	})

	t.Run("oldest at from boundary stops", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
			{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}
		cursor := buildNextCursor(prices, from)
		if cursor.hasMore {
			t.Error("hasMore = true, want false when oldest == from")
		}
	})

	t.Run("oldest before from stops", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
			{Date: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
		}
		cursor := buildNextCursor(prices, from)
		if cursor.hasMore {
			t.Error("hasMore = true, want false when oldest < from")
		}
	})

	t.Run("oldest after from continues with adjusted date", func(t *testing.T) {
		prices := []domain.DailyPrice{
			{Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
			{Date: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)},
		}
		cursor := buildNextCursor(prices, from)
		if !cursor.hasMore {
			t.Fatal("hasMore = false, want true when oldest > from")
		}

		wantEnd := time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC)
		if !cursor.endDate.Equal(wantEnd) {
			t.Errorf("endDate = %v, want %v", cursor.endDate, wantEnd)
		}
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"rate limited", fmt.Errorf("wrapped: %w", httpclient.ErrRateLimited), true},
		{"retryable 500", &httpclient.APIError{IsRetryable: true, StatusCode: 500, URL: "/test"}, true},
		{"retryable 429", &httpclient.APIError{IsRetryable: true, StatusCode: 429, URL: "/test"}, true},
		{"non-retryable 400", &httpclient.APIError{IsRetryable: false, StatusCode: 400, URL: "/test"}, false},
		{"generic error", errors.New("something broke"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.want {
				t.Errorf("IsRetryable = %v, want %v", got, tt.want)
			}
		})
	}
}
