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

// TestCombineDateTime guards the FICHADADIA clock_in / clock_out encoding fix.
// Before this helper, NewAttendanceMigrator passed *string directly into a
// TIMESTAMPTZ column and every batch failed with "cannot find encode plan ...
// *string into binary format for timestamptz (OID 1184)" once the pipeline
// finally had a non-empty legajo index and tried to write real rows.
func TestCombineDateTime(t *testing.T) {
	day := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		date    time.Time
		timeStr string
		wantNil bool
		wantH   int
		wantM   int
		wantS   int
	}{
		{"hh:mm:ss full", day, "08:30:45", false, 8, 30, 45},
		{"hh:mm short", day, "14:15", false, 14, 15, 0},
		{"iso timestamp", day, "2099-12-31 23:59:59", false, 23, 59, 59},
		{"empty time → nil", day, "", true, 0, 0, 0},
		{"zero date → nil", time.Time{}, "08:30:45", true, 0, 0, 0},
		{"garbage → nil", day, "not-a-time", true, 0, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := combineDateTime(tc.date, tc.timeStr)
			if tc.wantNil {
				if got != nil {
					t.Errorf("combineDateTime(%v, %q) = %v, want nil", tc.date, tc.timeStr, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("combineDateTime(%v, %q) = nil, want non-nil", tc.date, tc.timeStr)
			}
			if got.Year() != tc.date.Year() || got.Month() != tc.date.Month() || got.Day() != tc.date.Day() {
				t.Errorf("date part = %v, want date %v", got, tc.date)
			}
			if got.Hour() != tc.wantH || got.Minute() != tc.wantM || got.Second() != tc.wantS {
				t.Errorf("time part = %02d:%02d:%02d, want %02d:%02d:%02d",
					got.Hour(), got.Minute(), got.Second(), tc.wantH, tc.wantM, tc.wantS)
			}
		})
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
