package migration

import (
	"testing"
	"time"
)

func TestMapInvoiceStatus(t *testing.T) {
	tests := []struct {
		input string
		want  string
		err   bool
	}{
		{"anulada", "cancelled", false},
		{"pagada", "paid", false},
		{"pendiente", "posted", false},
		{"borrador", "draft", false},
		{"confirmada", "posted", false},
		{"unknown", "", true},
	}
	for _, tt := range tests {
		got, err := MapInvoiceStatus(tt.input)
		if tt.err && err == nil {
			t.Errorf("MapInvoiceStatus(%q) expected error", tt.input)
		}
		if !tt.err && err != nil {
			t.Errorf("MapInvoiceStatus(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("MapInvoiceStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapEntryType(t *testing.T) {
	tests := []struct {
		input string
		want  string
		err   bool
	}{
		{"MA", "manual", false},
		{"AU", "auto", false},
		{"AJ", "adjustment", false},
		{"RV", "reversal", false},
		{"XX", "", true},
	}
	for _, tt := range tests {
		got, err := MapEntryType(tt.input)
		if tt.err && err == nil {
			t.Errorf("MapEntryType(%q) expected error", tt.input)
		}
		if !tt.err && got != tt.want {
			t.Errorf("MapEntryType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSafeDate(t *testing.T) {
	zero := time.Time{}
	old := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	valid := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)

	if SafeDate(zero) != nil {
		t.Error("SafeDate(zero) should be nil")
	}
	if SafeDate(old) != nil {
		t.Error("SafeDate(year<1900) should be nil")
	}
	result := SafeDate(valid)
	if result == nil || !result.Equal(valid) {
		t.Errorf("SafeDate(valid) = %v, want %v", result, valid)
	}
}

func TestSafeDateRequired(t *testing.T) {
	zero := time.Time{}
	result := SafeDateRequired(zero)
	if result.Year() != 1970 {
		t.Errorf("SafeDateRequired(zero) year = %d, want 1970", result.Year())
	}

	valid := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !SafeDateRequired(valid).Equal(valid) {
		t.Error("SafeDateRequired(valid) should return same date")
	}
}

func TestParseDecimal(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "0"},
		{"", "0"},
		{"123.45", "123.45"},
		{"-50.00", "-50"},
		{"not_a_number", "0"},
	}
	for _, tt := range tests {
		got := ParseDecimal(tt.input)
		if got.String() != tt.want {
			t.Errorf("ParseDecimal(%q) = %s, want %s", tt.input, got.String(), tt.want)
		}
	}
}

func TestMapCodComprobante(t *testing.T) {
	tests := []struct {
		input int
		want  string
		err   bool
	}{
		{1, "invoice_a", false},
		{6, "invoice_b", false},
		{3, "credit_note_a", false},
		{99, "", true},
	}
	for _, tt := range tests {
		got, err := MapCodComprobante(tt.input)
		if tt.err && err == nil {
			t.Errorf("MapCodComprobante(%d) expected error", tt.input)
		}
		if !tt.err && got != tt.want {
			t.Errorf("MapCodComprobante(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
