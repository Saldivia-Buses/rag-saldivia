// Package export provides CSV and Excel export for tabular data.
package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/xuri/excelize/v2"
)

// Column defines an exportable column.
type Column struct {
	Header string `json:"header"`
	Key    string `json:"key"`
	Format string `json:"format"` // "text", "number", "currency", "date", "percent"
}

// Row is a generic map of column key → value.
type Row map[string]any

// WriteCSV writes rows as CSV to the writer.
func WriteCSV(w io.Writer, columns []Column, rows []Row) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}
	if err := cw.Write(headers); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, row := range rows {
		record := make([]string, len(columns))
		for i, c := range columns {
			record[i] = formatValue(row[c.Key], c.Format)
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	return cw.Error()
}

// WriteExcel writes rows as an XLSX file to the writer.
func WriteExcel(w io.Writer, sheetName string, columns []Column, rows []Row) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("new sheet: %w", err)
	}
	f.SetActiveSheet(sheet)
	// Remove default "Sheet1" if different
	if sheetName != "Sheet1" {
		_ = f.DeleteSheet("Sheet1")
	}

	// Bold header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E2E8F0"}, Pattern: 1},
	})

	// Write headers
	for i, c := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, c.Header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Write data
	for rowIdx, row := range rows {
		for colIdx, c := range columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			val := row[c.Key]
			switch c.Format {
			case "currency", "number", "percent":
				if n, ok := toFloat64(val); ok {
					_ = f.SetCellFloat(sheetName, cell, n, 2, 64)
				} else {
					_ = f.SetCellValue(sheetName, cell, formatValue(val, c.Format))
				}
			default:
				_ = f.SetCellValue(sheetName, cell, formatValue(val, c.Format))
			}
		}
	}

	// Auto-fit column widths (approximate)
	for i, c := range columns {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		width := float64(len(c.Header)) * 1.3
		if width < 12 {
			width = 12
		}
		if width > 40 {
			width = 40
		}
		_ = f.SetColWidth(sheetName, colName, colName, width)
	}

	return f.Write(w)
}

func formatValue(v any, format string) string {
	if v == nil {
		return ""
	}
	switch format {
	case "currency":
		if n, ok := toFloat64(v); ok {
			return fmt.Sprintf("%.2f", n)
		}
	case "number":
		if n, ok := toFloat64(v); ok {
			return fmt.Sprintf("%.2f", n)
		}
	case "percent":
		if n, ok := toFloat64(v); ok {
			return fmt.Sprintf("%.1f%%", n*100)
		}
	case "date":
		switch d := v.(type) {
		case time.Time:
			return d.Format("2006-01-02")
		case string:
			return d
		}
	}
	return fmt.Sprintf("%v", v)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}
