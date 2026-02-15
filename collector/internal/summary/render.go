package summary

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/summary.md.tmpl
var summaryTemplate string

// SummaryData holds all data for rendering the summary markdown.
type SummaryData struct {
	FXRate              *FXRateEntry
	GeneratedAt         string
	InsufficientSymbols []string
	KRRows              []SymbolRow
	USRows              []SymbolRow
}

// SymbolRow pairs a watchlist entry with its computed indicators.
type SymbolRow struct {
	Indicators SymbolIndicators
	Name       string
	Symbol     string
}

// FXRateEntry holds the latest exchange rate for display.
type FXRateEntry struct {
	Date string
	Pair string
	Rate float64
}

// RenderSummary executes the template with data and writes to outputPath atomically.
func RenderSummary(data SummaryData, outputPath string) error {
	tmpl, err := template.New("summary").Funcs(templateFuncMap()).Parse(summaryTemplate)
	if err != nil {
		return fmt.Errorf("parse summary template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("render summary template: %w", err)
	}

	return atomicWriteFile(outputPath, buf.Bytes())
}

// atomicWriteFile writes data to a temp file then renames to prevent partial writes.
func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory %s: %w", dir, err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("write temp file %s: %w", tmp, err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename %s to %s: %w", tmp, path, err)
	}

	return nil
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"fmtCross":    formatMACross,
		"fmtFXRate":   formatFXRate,
		"fmtOptPrice": formatOptionalPrice,
		"fmtPct":      formatOptionalPct,
		"fmtPos":      formatPosition,
		"fmtPrice":    formatPrice,
		"fmtRatio":    formatRatio,
	}
}

func formatOptionalPct(v *float64) string {
	if v == nil {
		return "-"
	}
	if *v >= 0 {
		return fmt.Sprintf("+%.2f%%", *v)
	}
	return fmt.Sprintf("%.2f%%", *v)
}

func formatPrice(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func formatOptionalPrice(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *v)
}

// formatPosition converts a 0.0-1.0 ratio to a 0-100% display string.
func formatPosition(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.0f%%", *v*100)
}

func formatMACross(v *string) string {
	if v == nil {
		return "-"
	}
	return *v
}

func formatRatio(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2fx", *v)
}

// formatFXRate formats a currency rate with comma thousand separators.
// Constraint: v must be non-negative. Negative values produce malformed output.
func formatFXRate(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	decPart := parts[1]

	n := len(intPart)
	if n <= 3 {
		return intPart + "." + decPart
	}

	var result strings.Builder
	for i, ch := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(ch)
	}
	return result.String() + "." + decPart
}
