package kis

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"valid date", "20240115", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), false},
		{"year boundary", "20251231", time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), false},
		{"empty string", "", time.Time{}, true},
		{"short format", "240115", time.Time{}, true},
		{"hyphenated", "2024-01-15", time.Time{}, true},
		{"invalid month", "20241315", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"integer price", "53900", 53900.0, false},
		{"decimal price", "150.25", 150.25, false},
		{"zero", "0", 0, false},
		{"empty returns zero", "", 0, false},
		{"whitespace returns zero", "  ", 0, false},
		{"padded value", " 53900 ", 53900.0, false},
		{"invalid", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFloat64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFloat64(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseFloat64(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"normal volume", "12345678", 12345678, false},
		{"zero", "0", 0, false},
		{"empty returns zero", "", 0, false},
		{"whitespace returns zero", "   ", 0, false},
		{"padded value", " 12345 ", 12345, false},
		{"decimal fails", "123.45", 0, true},
		{"invalid", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseInt64(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToDailyPrice(t *testing.T) {
	t.Run("converts normal row", func(t *testing.T) {
		row := kisOutputRow{
			AcmlVol:      "12345678",
			StckBsopDate: "20240115",
			StckClpr:     "53900",
			StckHgpr:     "54500",
			StckLwpr:     "53700",
			StckOprc:     "54300",
		}

		p, ok, err := toDailyPrice(row, "005930")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected ok=true for valid row")
		}

		if p.Symbol != "005930" {
			t.Errorf("Symbol = %q, want 005930", p.Symbol)
		}
		if p.Source != "kis" {
			t.Errorf("Source = %q, want kis", p.Source)
		}
		if p.Open != 54300 {
			t.Errorf("Open = %v, want 54300", p.Open)
		}
		if p.High != 54500 {
			t.Errorf("High = %v, want 54500", p.High)
		}
		if p.Low != 53700 {
			t.Errorf("Low = %v, want 53700", p.Low)
		}
		if p.Close != 53900 {
			t.Errorf("Close = %v, want 53900", p.Close)
		}
		if p.AdjClose != 53900 {
			t.Errorf("AdjClose = %v, want 53900 (same as close for KIS)", p.AdjClose)
		}
		if p.Volume != 12345678 {
			t.Errorf("Volume = %d, want 12345678", p.Volume)
		}

		wantDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		if !p.Date.Equal(wantDate) {
			t.Errorf("Date = %v, want %v", p.Date, wantDate)
		}
	})

	t.Run("skips empty date row", func(t *testing.T) {
		row := kisOutputRow{StckBsopDate: ""}

		_, ok, err := toDailyPrice(row, "005930")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected ok=false for empty date")
		}
	})

	t.Run("skips zero date row", func(t *testing.T) {
		row := kisOutputRow{StckBsopDate: "0"}

		_, ok, err := toDailyPrice(row, "005930")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Error("expected ok=false for zero date")
		}
	})

	t.Run("rejects invalid date", func(t *testing.T) {
		row := kisOutputRow{StckBsopDate: "not-a-date", StckClpr: "100"}

		_, _, err := toDailyPrice(row, "005930")
		if err == nil {
			t.Fatal("expected error for invalid date")
		}
	})

	t.Run("rejects invalid price", func(t *testing.T) {
		row := kisOutputRow{
			StckBsopDate: "20240115",
			StckClpr:     "abc",
			StckHgpr:     "100",
			StckLwpr:     "100",
			StckOprc:     "100",
			AcmlVol:      "100",
		}

		_, _, err := toDailyPrice(row, "005930")
		if err == nil {
			t.Fatal("expected error for invalid price")
		}
	})
}
