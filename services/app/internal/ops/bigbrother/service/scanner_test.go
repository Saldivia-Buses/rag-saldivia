package service

import (
	"testing"
)

func TestMaxNewDevicesPerScan_Is50(t *testing.T) {
	t.Parallel()
	// maxNewDevicesPerScan is a security constant that prevents MAC spoofing
	// from bloating the device table. It must remain at 50.
	if maxNewDevicesPerScan != 50 {
		t.Fatalf("expected maxNewDevicesPerScan to be 50, got %d", maxNewDevicesPerScan)
	}
}

func TestNilIfEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		isNil bool
	}{
		{name: "empty string", input: "", isNil: true},
		{name: "non-empty string", input: "hello", isNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := nilIfEmpty(tt.input)
			if tt.isNil && got != nil {
				t.Fatalf("expected nil for empty string, got %q", *got)
			}
			if !tt.isNil {
				if got == nil {
					t.Fatal("expected non-nil for non-empty string")
				}
				if *got != tt.input {
					t.Fatalf("expected %q, got %q", tt.input, *got)
				}
			}
		})
	}
}
