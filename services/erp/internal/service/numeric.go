package service

import (
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

// Precision-safe numeric helpers for financial operations.
// Uses math/big internally — never converts through float64.

// zeroNumeric returns a pgtype.Numeric representing 0.
func zeroNumeric() pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}
}

// negateNumeric returns -n without precision loss.
func negateNumeric(n pgtype.Numeric) pgtype.Numeric {
	if !n.Valid || n.Int == nil {
		return zeroNumeric()
	}
	neg := new(big.Int).Neg(n.Int)
	return pgtype.Numeric{Int: neg, Exp: n.Exp, Valid: true}
}

// absNumeric returns |n| without precision loss.
func absNumeric(n pgtype.Numeric) pgtype.Numeric {
	if !n.Valid || n.Int == nil {
		return zeroNumeric()
	}
	abs := new(big.Int).Abs(n.Int)
	return pgtype.Numeric{Int: abs, Exp: n.Exp, Valid: true}
}

// isPositive returns true if n > 0.
func isPositive(n pgtype.Numeric) bool {
	if !n.Valid || n.Int == nil {
		return false
	}
	return n.Int.Sign() > 0
}

// isNegative returns true if n < 0.
func isNegative(n pgtype.Numeric) bool {
	if !n.Valid || n.Int == nil {
		return false
	}
	return n.Int.Sign() < 0
}

// addNumeric returns a + b without precision loss.
// Both values are normalized to the same exponent before adding.
func addNumeric(a, b pgtype.Numeric) pgtype.Numeric {
	if !a.Valid || a.Int == nil {
		return b
	}
	if !b.Valid || b.Int == nil {
		return a
	}

	// Normalize to same exponent (use the smaller exp = more decimal places)
	aInt, bInt, exp := normalizeExp(a, b)
	sum := new(big.Int).Add(aInt, bInt)
	return pgtype.Numeric{Int: sum, Exp: exp, Valid: true}
}

// normalizeExp brings two Numeric values to the same exponent.
// Returns the adjusted big.Int values and the common exponent.
func normalizeExp(a, b pgtype.Numeric) (*big.Int, *big.Int, int32) {
	if a.Exp == b.Exp {
		return new(big.Int).Set(a.Int), new(big.Int).Set(b.Int), a.Exp
	}

	aInt := new(big.Int).Set(a.Int)
	bInt := new(big.Int).Set(b.Int)

	if a.Exp > b.Exp {
		// a has fewer decimal places — scale it up
		diff := a.Exp - b.Exp
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil)
		aInt.Mul(aInt, scale)
		return aInt, bInt, b.Exp
	}
	// b has fewer decimal places — scale it up
	diff := b.Exp - a.Exp
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil)
	bInt.Mul(bInt, scale)
	return aInt, bInt, a.Exp
}

// pgText creates a valid pgtype.Text from a string.
func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// pgNumeric parses a numeric string into pgtype.Numeric (precision-safe).
func pgNumeric(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if s == "" {
		s = "0"
	}
	_ = n.Scan(s)
	return n
}
