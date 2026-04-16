package legacy

import (
	"testing"
)

func TestLegacyRow_Int64(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{"int64", int64(42), 42},
		{"int", int(10), 10},
		{"float64", float64(99.0), 99},
		{"string", "123", 123},
		{"bytes", []byte("456"), 456},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		row := LegacyRow{"col": tt.value}
		got := row.Int64("col")
		if got != tt.want {
			t.Errorf("Int64(%s) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestLegacyRow_String(t *testing.T) {
	tests := []struct {
		value any
		want  string
	}{
		{"hello", "hello"},
		{[]byte("bytes"), "bytes"},
		{42, "42"},
		{nil, ""},
	}
	for _, tt := range tests {
		row := LegacyRow{"col": tt.value}
		got := row.String("col")
		if got != tt.want {
			t.Errorf("String(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestLegacyRow_NullString(t *testing.T) {
	row := LegacyRow{"a": "hello", "b": nil, "c": ""}
	if got := row.NullString("a"); got == nil || *got != "hello" {
		t.Errorf("NullString(a) = %v, want 'hello'", got)
	}
	if got := row.NullString("b"); got != nil {
		t.Errorf("NullString(b) = %v, want nil", got)
	}
	if got := row.NullString("c"); got != nil {
		t.Errorf("NullString(c) = %v, want nil (empty string)", got)
	}
	if got := row.NullString("missing"); got != nil {
		t.Errorf("NullString(missing) = %v, want nil", got)
	}
}

func TestLegacyRow_Decimal(t *testing.T) {
	tests := []struct {
		value any
		want  string
	}{
		{nil, "0"},
		{[]byte("123.45"), "123.45"},
		{"99.99", "99.99"},
		{float64(42.5), "42.5"},
	}
	for _, tt := range tests {
		row := LegacyRow{"col": tt.value}
		got := row.Decimal("col")
		if got != tt.want {
			t.Errorf("Decimal(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestParseCompositeKey(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]string
	}{
		{"", map[string]string{}},
		{"A=1,B=2", map[string]string{"A": "1", "B": "2"}},
		{"ID=42", map[string]string{"ID": "42"}},
	}
	for _, tt := range tests {
		got := ParseCompositeKey(tt.input)
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("ParseCompositeKey(%q)[%s] = %q, want %q", tt.input, k, got[k], v)
			}
		}
	}
}

func TestFormatCompositeKey(t *testing.T) {
	row := LegacyRow{"A": "1", "B": "hello"}
	got := FormatCompositeKey([]string{"A", "B"}, row)
	if got != "A=1,B=hello" {
		t.Errorf("FormatCompositeKey = %q, want A=1,B=hello", got)
	}
}
