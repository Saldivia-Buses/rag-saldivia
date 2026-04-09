package plc

import (
	"errors"
	"testing"
)

func ptr(f float64) *float64 { return &f }

func TestSafetyTierIsValid(t *testing.T) {
	tests := []struct {
		tier SafetyTier
		want bool
	}{
		{TierUnclassified, true},
		{TierSafe, true},
		{TierControlled, true},
		{TierCritical, true},
		{"invalid", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := tt.tier.IsValid(); got != tt.want {
			t.Errorf("%q.IsValid() = %v, want %v", tt.tier, got, tt.want)
		}
	}
}

func TestSafetyTierAllowsWrite(t *testing.T) {
	tests := []struct {
		tier SafetyTier
		want bool
	}{
		{TierUnclassified, false},
		{TierSafe, true},
		{TierControlled, true},
		{TierCritical, true},
	}
	for _, tt := range tests {
		if got := tt.tier.AllowsWrite(); got != tt.want {
			t.Errorf("%q.AllowsWrite() = %v, want %v", tt.tier, got, tt.want)
		}
	}
}

func TestSafetyTierRequiresTwoPersonApproval(t *testing.T) {
	tests := []struct {
		tier SafetyTier
		want bool
	}{
		{TierUnclassified, false},
		{TierSafe, false},
		{TierControlled, false},
		{TierCritical, true},
	}
	for _, tt := range tests {
		if got := tt.tier.RequiresTwoPersonApproval(); got != tt.want {
			t.Errorf("%q.RequiresTwoPersonApproval() = %v, want %v", tt.tier, got, tt.want)
		}
	}
}

func TestValidateWrite(t *testing.T) {
	tests := []struct {
		name    string
		tier    SafetyTier
		value   float64
		min     *float64
		max     *float64
		wantErr error
	}{
		{"unclassified rejected", TierUnclassified, 0, nil, nil, ErrUnclassified},
		{"safe allowed", TierSafe, 42, nil, nil, nil},
		{"safe ignores range", TierSafe, 999, ptr(0), ptr(100), nil},
		{"controlled in range", TierControlled, 50, ptr(0), ptr(100), nil},
		{"controlled at min", TierControlled, 0, ptr(0), ptr(100), nil},
		{"controlled at max", TierControlled, 100, ptr(0), ptr(100), nil},
		{"controlled below min", TierControlled, -1, ptr(0), ptr(100), ErrValueOutOfRange},
		{"controlled above max", TierControlled, 101, ptr(0), ptr(100), ErrValueOutOfRange},
		{"controlled nil min", TierControlled, -999, nil, ptr(100), nil},
		{"controlled nil max", TierControlled, 999, ptr(0), nil, nil},
		{"controlled nil both", TierControlled, 999, nil, nil, nil},
		{"critical in range", TierCritical, 50, ptr(0), ptr(100), nil},
		{"critical below min", TierCritical, -1, ptr(0), ptr(100), ErrValueOutOfRange},
		{"critical above max", TierCritical, 101, ptr(0), ptr(100), ErrValueOutOfRange},
		{"critical nil range", TierCritical, 999, nil, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWrite(tt.tier, tt.value, tt.min, tt.max)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateWriteInvalidTier(t *testing.T) {
	err := ValidateWrite("bogus", 0, nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid tier")
	}
}
