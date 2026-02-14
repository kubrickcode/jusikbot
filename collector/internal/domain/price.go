package domain

import "time"

// DailyPrice stores one trading day's OHLCV with adjusted close.
// Why DOUBLE PRECISION (float64): stock prices are not accounting figures;
// source APIs deliver floats, and indicator math gains 5-10x speed over NUMERIC.
type DailyPrice struct {
	AdjClose  float64
	Close     float64
	Date      time.Time
	FetchedAt time.Time
	High      float64
	IsAnomaly bool
	Low       float64
	Open      float64
	Source    string
	Symbol   string
	Volume   int64
}

// FXRate stores a single-day foreign exchange rate (e.g. USD/KRW).
type FXRate struct {
	Date      time.Time
	FetchedAt time.Time
	Pair      string
	Rate      float64
	Source    string
}
