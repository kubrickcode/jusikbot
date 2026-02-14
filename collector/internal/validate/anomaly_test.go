package validate

import (
	"math"
	"testing"

	"github.com/jusikbot/collector/internal/domain"
)

func TestIsPriceAnomaly(t *testing.T) {
	t.Run("KR market", func(t *testing.T) {
		tests := []struct {
			name        string
			current     float64
			previous    float64
			market      domain.Market
			secType     domain.SecurityType
			wantAnomaly bool
		}{
			{
				name:        "exactly at 30% threshold is not anomaly",
				current:     130,
				previous:    100,
				market:      domain.MarketKR,
				secType:     domain.SecurityTypeStock,
				wantAnomaly: false,
			},
			{
				name:        "just above 30% threshold is anomaly",
				current:     130.01,
				previous:    100,
				market:      domain.MarketKR,
				secType:     domain.SecurityTypeStock,
				wantAnomaly: true,
			},
			{
				name:        "negative change exactly at -30% threshold is not anomaly",
				current:     70,
				previous:    100,
				market:      domain.MarketKR,
				secType:     domain.SecurityTypeStock,
				wantAnomaly: false,
			},
			{
				name:        "negative change just beyond -30% threshold is anomaly",
				current:     69.99,
				previous:    100,
				market:      domain.MarketKR,
				secType:     domain.SecurityTypeStock,
				wantAnomaly: true,
			},
			{
				name:        "KR ETF uses same 30% threshold",
				current:     130.01,
				previous:    100,
				market:      domain.MarketKR,
				secType:     domain.SecurityTypeETF,
				wantAnomaly: true,
			},
			{
				name:        "normal daily move is not anomaly",
				current:     102,
				previous:    100,
				market:      domain.MarketKR,
				secType:     domain.SecurityTypeStock,
				wantAnomaly: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsPriceAnomaly(tt.current, tt.previous, tt.market, tt.secType)
				if got != tt.wantAnomaly {
					change := math.Abs(tt.current-tt.previous) / tt.previous * 100
					t.Errorf("IsPriceAnomaly(%.2f, %.2f, %s, %s) = %v, want %v (change=%.4f%%)",
						tt.current, tt.previous, tt.market, tt.secType, got, tt.wantAnomaly, change)
				}
			})
		}
	})

	t.Run("US ETF", func(t *testing.T) {
		tests := []struct {
			name        string
			current     float64
			previous    float64
			wantAnomaly bool
		}{
			{
				name:        "exactly at 15% threshold is not anomaly",
				current:     115,
				previous:    100,
				wantAnomaly: false,
			},
			{
				name:        "just above 15% threshold is anomaly",
				current:     115.01,
				previous:    100,
				wantAnomaly: true,
			},
			{
				name:        "negative change exactly at -15% is not anomaly",
				current:     85,
				previous:    100,
				wantAnomaly: false,
			},
			{
				name:        "negative change just beyond -15% is anomaly",
				current:     84.99,
				previous:    100,
				wantAnomaly: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsPriceAnomaly(tt.current, tt.previous, domain.MarketUS, domain.SecurityTypeETF)
				if got != tt.wantAnomaly {
					change := math.Abs(tt.current-tt.previous) / tt.previous * 100
					t.Errorf("IsPriceAnomaly(%.2f, %.2f, US, etf) = %v, want %v (change=%.4f%%)",
						tt.current, tt.previous, got, tt.wantAnomaly, change)
				}
			})
		}
	})

	t.Run("US individual stock", func(t *testing.T) {
		tests := []struct {
			name        string
			current     float64
			previous    float64
			wantAnomaly bool
		}{
			{
				name:        "exactly at 50% threshold is not anomaly",
				current:     150,
				previous:    100,
				wantAnomaly: false,
			},
			{
				name:        "just above 50% threshold is anomaly",
				current:     150.01,
				previous:    100,
				wantAnomaly: true,
			},
			{
				name:        "negative change exactly at -50% is not anomaly",
				current:     50,
				previous:    100,
				wantAnomaly: false,
			},
			{
				name:        "negative change just beyond -50% is anomaly",
				current:     49.99,
				previous:    100,
				wantAnomaly: true,
			},
			{
				name:        "35% change is within US stock threshold",
				current:     135,
				previous:    100,
				wantAnomaly: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := IsPriceAnomaly(tt.current, tt.previous, domain.MarketUS, domain.SecurityTypeStock)
				if got != tt.wantAnomaly {
					change := math.Abs(tt.current-tt.previous) / tt.previous * 100
					t.Errorf("IsPriceAnomaly(%.2f, %.2f, US, stock) = %v, want %v (change=%.4f%%)",
						tt.current, tt.previous, got, tt.wantAnomaly, change)
				}
			})
		}
	})

	t.Run("first data point skipped", func(t *testing.T) {
		// Why zero previous: first data point has no prior day to compare against.
		got := IsPriceAnomaly(100, 0, domain.MarketKR, domain.SecurityTypeStock)
		if got {
			t.Error("IsPriceAnomaly should return false when previous is zero (first data point)")
		}
	})
}

func TestCrossValidateAdjClose(t *testing.T) {
	t.Run("stock split day", func(t *testing.T) {
		tests := []struct {
			name        string
			splitFactor float64
			divCash     float64
			wantAnomaly bool
		}{
			{
				name:        "2-for-1 split is not anomaly",
				splitFactor: 2.0,
				divCash:     0,
				wantAnomaly: false,
			},
			{
				name:        "3-for-1 split is not anomaly",
				splitFactor: 3.0,
				divCash:     0,
				wantAnomaly: false,
			},
			{
				name:        "reverse split 0.5 is not anomaly",
				splitFactor: 0.5,
				divCash:     0,
				wantAnomaly: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := CrossValidateAdjClose(tt.splitFactor, tt.divCash)
				if got != tt.wantAnomaly {
					t.Errorf("CrossValidateAdjClose(splitFactor=%.2f, divCash=%.2f) = %v, want %v",
						tt.splitFactor, tt.divCash, got, tt.wantAnomaly)
				}
			})
		}
	})

	t.Run("dividend day", func(t *testing.T) {
		tests := []struct {
			name        string
			splitFactor float64
			divCash     float64
			wantAnomaly bool
		}{
			{
				name:        "cash dividend is not anomaly",
				splitFactor: 1.0,
				divCash:     1.50,
				wantAnomaly: false,
			},
			{
				name:        "large special dividend is not anomaly",
				splitFactor: 1.0,
				divCash:     10.00,
				wantAnomaly: false,
			},
			{
				name:        "split and dividend on same day is not anomaly",
				splitFactor: 2.0,
				divCash:     0.50,
				wantAnomaly: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := CrossValidateAdjClose(tt.splitFactor, tt.divCash)
				if got != tt.wantAnomaly {
					t.Errorf("CrossValidateAdjClose(splitFactor=%.2f, divCash=%.2f) = %v, want %v",
						tt.splitFactor, tt.divCash, got, tt.wantAnomaly)
				}
			})
		}
	})

	t.Run("normal day", func(t *testing.T) {
		tests := []struct {
			name        string
			splitFactor float64
			divCash     float64
			wantAnomaly bool
		}{
			{
				name:        "no corporate action is confirmed anomaly",
				splitFactor: 1.0,
				divCash:     0,
				wantAnomaly: true,
			},
			{
				name:        "zero split factor treated as no split",
				splitFactor: 0,
				divCash:     0,
				wantAnomaly: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := CrossValidateAdjClose(tt.splitFactor, tt.divCash)
				if got != tt.wantAnomaly {
					t.Errorf("CrossValidateAdjClose(splitFactor=%.2f, divCash=%.2f) = %v, want %v",
						tt.splitFactor, tt.divCash, got, tt.wantAnomaly)
				}
			})
		}
	})
}
