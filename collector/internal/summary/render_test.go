package summary

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "overwrite golden files with actual output")

func ptrString(s string) *string { return &s }

func TestFormatOptionalPct(t *testing.T) {
	tests := []struct {
		name  string
		input *float64
		want  string
	}{
		{"nil returns dash", nil, "-"},
		{"positive with plus sign", ptrFloat(12.345), "+12.35%"},
		{"negative without plus", ptrFloat(-5.678), "-5.68%"},
		{"zero shows plus sign", ptrFloat(0), "+0.00%"},
		{"small positive", ptrFloat(0.001), "+0.00%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOptionalPct(tt.input)
			if got != tt.want {
				t.Errorf("formatOptionalPct = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"US stock price", 875.284, "875.28"},
		{"KR stock price", 35210.0, "35210.00"},
		{"zero", 0, "0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPrice(tt.input)
			if got != tt.want {
				t.Errorf("formatPrice = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatOptionalPrice(t *testing.T) {
	tests := []struct {
		name  string
		input *float64
		want  string
	}{
		{"nil returns dash", nil, "-"},
		{"valid price", ptrFloat(120.50), "120.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOptionalPrice(tt.input)
			if got != tt.want {
				t.Errorf("formatOptionalPrice = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPosition(t *testing.T) {
	tests := []struct {
		name  string
		input *float64
		want  string
	}{
		{"nil returns dash", nil, "-"},
		{"midpoint shows 50%", ptrFloat(0.5), "50%"},
		{"at high shows 100%", ptrFloat(1.0), "100%"},
		{"at low shows 0%", ptrFloat(0.0), "0%"},
		{"75th percentile", ptrFloat(0.75), "75%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPosition(tt.input)
			if got != tt.want {
				t.Errorf("formatPosition = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatMACross(t *testing.T) {
	tests := []struct {
		name  string
		input *string
		want  string
	}{
		{"nil returns dash", nil, "-"},
		{"golden cross", ptrString("GC"), "GC"},
		{"dead cross", ptrString("DC"), "DC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMACross(tt.input)
			if got != tt.want {
				t.Errorf("formatMACross = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatRatio(t *testing.T) {
	tests := []struct {
		name  string
		input *float64
		want  string
	}{
		{"nil returns dash", nil, "-"},
		{"normal ratio", ptrFloat(1.234), "1.23x"},
		{"high ratio", ptrFloat(3.0), "3.00x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRatio(tt.input)
			if got != tt.want {
				t.Errorf("formatRatio = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatFXRate(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"typical KRW rate", 1345.50, "1,345.50"},
		{"small rate", 1.23, "1.23"},
		{"large rate", 12345.67, "12,345.67"},
		{"round number", 1000.00, "1,000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFXRate(tt.input)
			if got != tt.want {
				t.Errorf("formatFXRate = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderSummary(t *testing.T) {
	t.Run("golden file comparison", func(t *testing.T) {
		gc := "GC"
		data := SummaryData{
			GeneratedAt: "2025-02-15 14:30 UTC",
			USRows: []SymbolRow{
				{
					Symbol: "NVDA",
					Name:   "NVIDIA",
					Indicators: SymbolIndicators{
						AdjClose:         875.28,
						Change5D:         ptrFloat(2.34),
						Change20D:        ptrFloat(15.67),
						FiftyTwoWeekHigh: ptrFloat(974.00),
						FiftyTwoWeekLow:  ptrFloat(450.00),
						FiftyTwoWeekPos:  ptrFloat(0.81),
						HV20D:            ptrFloat(45.67),
						HV60D:            ptrFloat(38.90),
						MACross:          nil,
						MADivergence50D:  ptrFloat(5.23),
						MADivergence200D: ptrFloat(12.34),
						RelativeBench20D: ptrFloat(8.45),
						VolRatio:         ptrFloat(1.23),
					},
				},
				{
					Symbol: "QQQ",
					Name:   "Invesco QQQ Trust",
					Indicators: SymbolIndicators{
						AdjClose:         485.12,
						Change5D:         ptrFloat(-1.05),
						Change20D:        ptrFloat(7.22),
						FiftyTwoWeekHigh: ptrFloat(510.00),
						FiftyTwoWeekLow:  ptrFloat(380.00),
						FiftyTwoWeekPos:  ptrFloat(0.808),
						HV20D:            ptrFloat(18.50),
						HV60D:            ptrFloat(16.80),
						MACross:          &gc,
						MADivergence50D:  ptrFloat(2.10),
						MADivergence200D: ptrFloat(8.55),
						RelativeBench20D: nil,
						VolRatio:         ptrFloat(0.95),
					},
				},
			},
			KRRows: []SymbolRow{
				{
					Symbol: "069500",
					Name:   "KODEX 200",
					Indicators: SymbolIndicators{
						AdjClose:         35210.00,
						Change5D:         ptrFloat(-0.42),
						Change20D:        ptrFloat(3.15),
						FiftyTwoWeekHigh: ptrFloat(37500.00),
						FiftyTwoWeekLow:  ptrFloat(30100.00),
						FiftyTwoWeekPos:  ptrFloat(0.69),
						HV20D:            ptrFloat(22.30),
						HV60D:            ptrFloat(20.10),
						MACross:          nil,
						MADivergence50D:  ptrFloat(1.50),
						MADivergence200D: ptrFloat(4.20),
						RelativeBench20D: nil,
						VolRatio:         ptrFloat(1.10),
					},
				},
			},
			FXRate: &FXRateEntry{
				Pair: "USD/KRW",
				Rate: 1345.50,
				Date: "2025-02-14",
			},
			InsufficientSymbols: nil,
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := RenderSummary(data, outputPath); err != nil {
			t.Fatalf("RenderSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		goldenPath := "testdata/golden_us_kr_fx.md"
		if *updateGolden {
			if err := os.MkdirAll("testdata", 0755); err != nil {
				t.Fatalf("create testdata dir: %v", err)
			}
			if err := os.WriteFile(goldenPath, got, 0644); err != nil {
				t.Fatalf("write golden file: %v", err)
			}
			t.Log("golden file updated")
			return
		}

		golden, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("read golden file (run with -update-golden to create): %v", err)
		}

		if string(got) != string(golden) {
			t.Errorf("output does not match golden file.\n--- GOT ---\n%s\n--- WANT ---\n%s", got, golden)
		}
	})

	t.Run("US only without FX or notes", func(t *testing.T) {
		data := SummaryData{
			GeneratedAt: "2025-01-01 00:00 UTC",
			USRows: []SymbolRow{
				{
					Symbol: "TEST",
					Name:   "Test Stock",
					Indicators: SymbolIndicators{
						AdjClose: 100.00,
					},
				},
			},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := RenderSummary(data, outputPath); err != nil {
			t.Fatalf("RenderSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		content := string(got)
		if !strings.Contains(content, "## US Stocks") {
			t.Error("missing US Stocks section")
		}
		if strings.Contains(content, "## KR Stocks") {
			t.Error("KR Stocks section should be absent")
		}
		if strings.Contains(content, "## Exchange Rate") {
			t.Error("Exchange Rate section should be absent")
		}
		if strings.Contains(content, "*Notes:*") {
			t.Error("Notes section should be absent")
		}
	})

	t.Run("insufficient data notes rendered", func(t *testing.T) {
		data := SummaryData{
			GeneratedAt: "2025-01-01 00:00 UTC",
			USRows: []SymbolRow{
				{
					Symbol:     "NEW",
					Name:       "New Stock",
					Indicators: SymbolIndicators{AdjClose: 50.00},
				},
			},
			InsufficientSymbols: []string{
				"NEW (New Stock): 200D MA 데이터 부족 (< 200 거래일)",
			},
		}

		outputDir := t.TempDir()
		outputPath := filepath.Join(outputDir, "summary.md")

		if err := RenderSummary(data, outputPath); err != nil {
			t.Fatalf("RenderSummary failed: %v", err)
		}

		got, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		content := string(got)
		if !strings.Contains(content, "*Notes:*") {
			t.Error("missing Notes section")
		}
		if !strings.Contains(content, "NEW (New Stock): 200D MA") {
			t.Error("missing insufficient data note for NEW")
		}
	})
}

func TestAtomicWriteFile(t *testing.T) {
	t.Run("creates parent directories", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sub", "dir", "file.md")

		if err := atomicWriteFile(path, []byte("test content")); err != nil {
			t.Fatalf("atomicWriteFile failed: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read output: %v", err)
		}
		if string(got) != "test content" {
			t.Errorf("content = %q, want %q", got, "test content")
		}
	})

	t.Run("no leftover tmp file on success", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file.md")

		if err := atomicWriteFile(path, []byte("data")); err != nil {
			t.Fatalf("atomicWriteFile failed: %v", err)
		}

		if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
			t.Error("tmp file should not exist after successful write")
		}
	})
}
