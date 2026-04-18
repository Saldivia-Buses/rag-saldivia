// Package plc provides PLC communication clients and safety types for
// industrial control systems. Supports Modbus TCP and OPC-UA protocols.
//
// WARNING: This package provides raw protocol clients. For production use
// with physical PLCs, always go through a service that enforces safety tiers,
// rate limits, audit logging, and credential management (e.g., BigBrother).
// Importing this package directly to talk to production PLCs bypasses all
// safety controls and is strongly discouraged.
package plc

import (
	"errors"
	"fmt"
)

// SafetyTier classifies PLC registers by the risk level of writing to them.
type SafetyTier string

const (
	// TierUnclassified: auto-discovered registers. Read-only until a human
	// classifies them. Writing is forbidden.
	TierUnclassified SafetyTier = "unclassified"

	// TierSafe: informational registers (display text, counters).
	// Writing requires bigbrother.plc.write permission.
	TierSafe SafetyTier = "safe"

	// TierControlled: operational registers (temperature setpoint, speed).
	// Writing requires bigbrother.plc.write + value within [min, max] range
	// + rate limit (max_writes_per_min).
	TierControlled SafetyTier = "controlled"

	// TierCritical: safety-critical registers (emergency stop, pressure relief).
	// Writing requires bigbrother.admin + two-person approval.
	TierCritical SafetyTier = "critical"
)

var (
	ErrWriteNotAllowed = errors.New("write not allowed for this safety tier")
	ErrValueOutOfRange = errors.New("value out of allowed range")
	ErrUnclassified    = errors.New("register is unclassified — classify before writing")
)

// ValidTiers contains all valid safety tier values.
var ValidTiers = []SafetyTier{TierUnclassified, TierSafe, TierControlled, TierCritical}

// IsValid returns true if the tier is a recognized value.
func (t SafetyTier) IsValid() bool {
	for _, v := range ValidTiers {
		if t == v {
			return true
		}
	}
	return false
}

// AllowsWrite returns true if the tier allows writes (with appropriate checks).
// Unclassified registers never allow writes.
func (t SafetyTier) AllowsWrite() bool {
	return t == TierSafe || t == TierControlled || t == TierCritical
}

// RequiresTwoPersonApproval returns true if writes to this tier need a second approver.
func (t SafetyTier) RequiresTwoPersonApproval() bool {
	return t == TierCritical
}

// ValidateWrite checks whether a write operation is allowed based on the safety
// tier and value range. For controlled tier, validates that value is within
// [min, max] range. For critical tier, only checks range — the two-person
// approval must be enforced by the caller.
//
// Returns nil if the write is allowed, or an error explaining why not.
func ValidateWrite(tier SafetyTier, value float64, min, max *float64) error {
	if !tier.IsValid() {
		return fmt.Errorf("invalid safety tier: %q", tier)
	}

	if tier == TierUnclassified {
		return ErrUnclassified
	}

	if tier == TierControlled || tier == TierCritical {
		if min != nil && value < *min {
			return fmt.Errorf("%w: %v < min %v", ErrValueOutOfRange, value, *min)
		}
		if max != nil && value > *max {
			return fmt.Errorf("%w: %v > max %v", ErrValueOutOfRange, value, *max)
		}
	}

	return nil
}
