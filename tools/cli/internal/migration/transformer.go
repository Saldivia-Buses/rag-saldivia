package migration

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

// Status mapping tables (legacy → SDA). Unknown values cause fatal errors.

var invoiceStatusMap = map[string]string{
	"anulada":    "cancelled",
	"pagada":     "paid",
	"pendiente":  "posted",
	"borrador":   "draft",
	"confirmada": "posted",
}

var entryTypeMap = map[string]string{
	"MA": "manual",
	"AU": "auto",
	"AJ": "adjustment",
	"RV": "reversal",
}

var treasuryStatusMap = map[string]string{
	"confirmado": "confirmed",
	"anulado":    "cancelled",
	"pendiente":  "pending",
}

var fiscalStatusMap = map[string]string{
	"abierto": "open",
	"cerrado": "closed",
}

var codComprobanteMap = map[int]string{
	1:  "invoice_a",
	2:  "debit_note_a",
	3:  "credit_note_a",
	6:  "invoice_b",
	7:  "debit_note_b",
	8:  "credit_note_b",
	11: "invoice_c",
	12: "debit_note_c",
	13: "credit_note_c",
}

// MapInvoiceStatus maps a legacy invoice status to SDA. Fatal on unknown value.
func MapInvoiceStatus(legacy string) (string, error) {
	if v, ok := invoiceStatusMap[legacy]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown invoice status %q — add mapping to transformer.go", legacy)
}

// MapEntryType maps a legacy journal entry type to SDA. Fatal on unknown value.
func MapEntryType(legacy string) (string, error) {
	if v, ok := entryTypeMap[legacy]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown entry type %q — add mapping to transformer.go", legacy)
}

// MapTreasuryStatus maps a legacy treasury movement status. Fatal on unknown value.
func MapTreasuryStatus(legacy string) (string, error) {
	if v, ok := treasuryStatusMap[legacy]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown treasury status %q — add mapping to transformer.go", legacy)
}

// MapFiscalStatus maps a legacy fiscal year status. Fatal on unknown value.
func MapFiscalStatus(legacy string) (string, error) {
	if v, ok := fiscalStatusMap[legacy]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown fiscal status %q — add mapping to transformer.go", legacy)
}

// MapCodComprobante maps Histrix CODCOMPROBANTE to SDA invoice type.
func MapCodComprobante(code int) (string, error) {
	if v, ok := codComprobanteMap[code]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown CODCOMPROBANTE %d — add mapping to transformer.go", code)
}

// SafeDate converts a time.Time, returning nil for zero dates (MySQL 0000-00-00).
func SafeDate(t time.Time) *time.Time {
	if t.IsZero() || t.Year() < 1900 {
		return nil
	}
	return &t
}

// SafeDateRequired converts a time.Time, returning a default date for zero dates.
// Use this for NOT NULL date columns.
func SafeDateRequired(t time.Time) time.Time {
	if t.IsZero() || t.Year() < 1900 {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return t
}

// ParseDecimal parses a string value into decimal.Decimal.
// Logs a warning on invalid input instead of silently returning zero.
func ParseDecimal(s string) decimal.Decimal {
	if s == "" || s == "0" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		slog.Warn("invalid decimal value, defaulting to zero", "value", s, "err", err)
		return decimal.Zero
	}
	return d
}

// directionMap maps Histrix subsistema_id to SDA direction.
var directionMap = map[string]string{
	"01": "received",
	"1":  "received",
	"02": "issued",
	"2":  "issued",
}

// accountMovementTypeMap maps Histrix comprobante prefix to SDA movement type.
var accountMovementTypeMap = map[string]string{
	"FC": "invoice",
	"RC": "payment",
	"NC": "credit_note",
	"ND": "debit_note",
}

// movementTypeMap maps Histrix treasury movement types to SDA.
var movementTypeMap = map[string]string{
	"ING":  "income",
	"EGR":  "expense",
	"TRF":  "transfer",
	"AJU":  "adjustment",
	"DEP":  "deposit",
	"RET":  "withdrawal",
	"CHQE": "check_issued",
	"CHQR": "check_received",
}

// MapDirection maps Histrix subsistema_id to SDA direction ("received" or "issued").
func MapDirection(subsistemaID string) (string, error) {
	if v, ok := directionMap[subsistemaID]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown subsistema_id %q — add mapping to transformer.go", subsistemaID)
}

// MapAccountMovementType maps Histrix comprobante prefix (FC, RC, NC, ND) to SDA account movement type.
func MapAccountMovementType(prefix string) (string, error) {
	if v, ok := accountMovementTypeMap[prefix]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown account movement type %q — add mapping to transformer.go", prefix)
}

// MapMovementType maps Histrix treasury movement type to SDA.
func MapMovementType(legacy string) (string, error) {
	if v, ok := movementTypeMap[legacy]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown movement type %q — add mapping to transformer.go", legacy)
}

// MapInvoiceType maps Histrix (codcom, codlet) to SDA invoice_type string.
// codcom: 1=Factura, 2=Nota de Débito, 3=Nota de Crédito.
// codlet: A, B, C, E (fiscal letter).
func MapInvoiceType(codcom int, codlet string) string {
	key := fmt.Sprintf("%d:%s", codcom, codlet)
	result, ok := invoiceTypeMap[key]
	if !ok {
		slog.Warn("unknown codcom+codlet combo, defaulting to invoice_a",
			"codcom", codcom, "codlet", codlet)
		return "invoice_a"
	}
	return result
}

var invoiceTypeMap = map[string]string{
	"1:A": "invoice_a",
	"1:B": "invoice_b",
	"1:C": "invoice_c",
	"1:E": "invoice_e",
	"2:A": "debit_note_a",
	"2:B": "debit_note_b",
	"3:A": "credit_note_a",
	"3:B": "credit_note_b",
	"3:C": "credit_note_c",
}

// LegacyUserID is the constant user_id for migrated records.
const LegacyUserID = "legacy-import"
