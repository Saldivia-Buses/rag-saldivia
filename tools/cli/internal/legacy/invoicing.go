package legacy

// Invoicing readers are intentionally deferred until Plan 19 (AFIP integration).
//
// Invoice migration requires:
// - CODCOMPROBANTE → invoice_type mapping (coupled to AFIP voucher types)
// - CAE/CAE_VTO fields only meaningful with AFIP config in place
// - Withholding readers (RET1598, RETGANAN, RETIVA) need AFIP tax regime mapping
// - Account movements (REG_MOVIMIENTOS) depend on invoices being migrated first
//
// These readers exist in the plan but are not wired until Plan 19 is implemented.
// See docs/plans/2.0.x-plan21-data-migration.md for the full specification.
