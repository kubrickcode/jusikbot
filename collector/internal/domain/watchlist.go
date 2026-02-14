package domain

// Market identifies the stock exchange region.
type Market string

const (
	MarketUS Market = "US"
	MarketKR Market = "KR"
)

// SecurityType distinguishes individual stocks from ETFs.
// Why this matters: anomaly thresholds differ by type (US ETF 15%, US stock 50%).
type SecurityType string

const (
	SecurityTypeStock SecurityType = "stock"
	SecurityTypeETF   SecurityType = "etf"
)

// WatchlistEntry represents a single tracked symbol loaded from watchlist.json.
type WatchlistEntry struct {
	Market Market
	Name   string
	Symbol string
	Type   SecurityType
}
