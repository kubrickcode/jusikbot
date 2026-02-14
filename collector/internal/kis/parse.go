package kis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jusikbot/collector/internal/domain"
)

const sourceName = "kis"

// kisOutputRow represents a single row from KIS daily chart price API output2.
// Why all fields are strings: KIS API returns all numeric values as strings.
type kisOutputRow struct {
	AcmlVol      string `json:"acml_vol"`
	StckBsopDate string `json:"stck_bsop_date"`
	StckClpr     string `json:"stck_clpr"`
	StckHgpr     string `json:"stck_hgpr"`
	StckLwpr     string `json:"stck_lwpr"`
	StckOprc     string `json:"stck_oprc"`
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("20060102", s)
}

// parseFloat64 converts a KIS numeric string to float64.
// Returns 0 for empty/whitespace strings (blank rows from API).
func parseFloat64(s string) (float64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, nil
	}
	return strconv.ParseFloat(trimmed, 64)
}

// parseInt64 converts a KIS volume string to int64.
// Returns 0 for empty/whitespace strings.
func parseInt64(s string) (int64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, nil
	}
	return strconv.ParseInt(trimmed, 10, 64)
}

// toDailyPrice converts a KIS output row to a domain.DailyPrice.
// Returns ok=false for empty rows (blank or zero date).
// Why adjClose == close: KIS FID_ORG_ADJ_PRC=0 returns split-adjusted prices only;
// dividends are NOT reflected (see analysis.md).
func toDailyPrice(row kisOutputRow, symbol string) (domain.DailyPrice, bool, error) {
	if row.StckBsopDate == "" || row.StckBsopDate == "0" {
		return domain.DailyPrice{}, false, nil
	}

	date, err := parseDate(row.StckBsopDate)
	if err != nil {
		return domain.DailyPrice{}, false, fmt.Errorf("parse date %q: %w", row.StckBsopDate, err)
	}

	close_, err := parseFloat64(row.StckClpr)
	if err != nil {
		return domain.DailyPrice{}, false, fmt.Errorf("parse close %q: %w", row.StckClpr, err)
	}

	high, err := parseFloat64(row.StckHgpr)
	if err != nil {
		return domain.DailyPrice{}, false, fmt.Errorf("parse high %q: %w", row.StckHgpr, err)
	}

	low, err := parseFloat64(row.StckLwpr)
	if err != nil {
		return domain.DailyPrice{}, false, fmt.Errorf("parse low %q: %w", row.StckLwpr, err)
	}

	open, err := parseFloat64(row.StckOprc)
	if err != nil {
		return domain.DailyPrice{}, false, fmt.Errorf("parse open %q: %w", row.StckOprc, err)
	}

	volume, err := parseInt64(row.AcmlVol)
	if err != nil {
		return domain.DailyPrice{}, false, fmt.Errorf("parse volume %q: %w", row.AcmlVol, err)
	}

	return domain.DailyPrice{
		AdjClose: close_,
		Close:    close_,
		Date:     date,
		High:     high,
		Low:      low,
		Open:     open,
		Source:   sourceName,
		Symbol:   symbol,
		Volume:   volume,
	}, true, nil
}
