package export

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

func TestWriteCSV(t *testing.T) {
	cols := []Column{
		{Header: "Name", Key: "name", Format: "text"},
		{Header: "Amount", Key: "amount", Format: "currency"},
		{Header: "Rate", Key: "rate", Format: "percent"},
	}
	rows := []Row{
		{"name": "Alice", "amount": 1234.56, "rate": 0.15},
		{"name": "Bob", "amount": 0.0, "rate": nil},
	}

	var buf bytes.Buffer
	err := WriteCSV(&buf, cols, rows)
	if err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	r := csv.NewReader(strings.NewReader(buf.String()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if len(records) != 3 { // header + 2 rows
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	// Header
	if records[0][0] != "Name" || records[0][1] != "Amount" || records[0][2] != "Rate" {
		t.Errorf("header mismatch: %v", records[0])
	}

	// Row 1
	if records[1][0] != "Alice" {
		t.Errorf("expected Alice, got %s", records[1][0])
	}
	if records[1][1] != "1234.56" {
		t.Errorf("expected 1234.56, got %s", records[1][1])
	}
	if records[1][2] != "15.0%" {
		t.Errorf("expected 15.0%%, got %s", records[1][2])
	}

	// Row 2 — nil rate
	if records[2][2] != "" {
		t.Errorf("expected empty for nil rate, got %s", records[2][2])
	}
}

func TestWriteExcel(t *testing.T) {
	cols := []Column{
		{Header: "Code", Key: "code", Format: "text"},
		{Header: "Total", Key: "total", Format: "currency"},
	}
	rows := []Row{
		{"code": "A001", "total": 5000.0},
		{"code": "B002", "total": 0.0},
	}

	var buf bytes.Buffer
	err := WriteExcel(&buf, "Test", cols, rows)
	if err != nil {
		t.Fatalf("WriteExcel: %v", err)
	}

	// Verify it produced valid XLSX (at least some bytes)
	if buf.Len() < 100 {
		t.Fatalf("XLSX too small: %d bytes", buf.Len())
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		value  any
		format string
		want   string
	}{
		{nil, "text", ""},
		{"hello", "text", "hello"},
		{42.5, "currency", "42.50"},
		{0.15, "percent", "15.0%"},
		{100, "number", "100.00"},
		{nil, "currency", ""},
	}
	for _, tt := range tests {
		got := formatValue(tt.value, tt.format)
		if got != tt.want {
			t.Errorf("formatValue(%v, %q) = %q, want %q", tt.value, tt.format, got, tt.want)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input any
		want  float64
		ok    bool
	}{
		{42.0, 42.0, true},
		{float32(3.14), 3.140000104904175, true}, // float32 precision
		{int(10), 10.0, true},
		{int64(100), 100.0, true},
		{"not a number", 0, false},
		{nil, 0, false},
	}
	for _, tt := range tests {
		got, ok := toFloat64(tt.input)
		if ok != tt.ok {
			t.Errorf("toFloat64(%v) ok=%v, want ok=%v", tt.input, ok, tt.ok)
		}
		if ok && got != tt.want {
			t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
