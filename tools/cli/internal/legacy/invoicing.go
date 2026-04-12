package legacy

import "database/sql"

// InvoiceReader creates a reader for IVAVENTAS (IVA sales book — our best source for issued invoices).
// FACREMIT has no PK and is a remitos table, not the actual invoice table.
// IVAVENTAS has all invoice data with AI PK.
func InvoiceReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "IVAVENTAS",
		Target:     "erp_invoices",
		DomainName: "invoicing",
		PKColumn:   "id_ivaventa",
		Columns:    "id_ivaventa, codcom, codlet, nronpv, nrocom, feciva, ctacod, ctanom, totcom, execom, nogcom, regmovim_id",
	}
}

// PurchaseInvoiceReader creates a reader for IVACOMPRAS (facturas recibidas).
func PurchaseInvoiceReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "IVACOMPRAS",
		Target:     "erp_invoices",
		DomainName: "invoicing",
		PKColumn:   "id_ivacompra",
		Columns:    "id_ivacompra, codcom, codlet, nronpv, nrocom, feciva, ctacod, ctanom, totcom, execom, nogcom, regmovim_id",
	}
}

// InvoiceLineReader creates a reader for FACDETAL (detalle factura).
func InvoiceLineReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "FACDETAL",
		Target:     "erp_invoice_lines",
		DomainName: "invoicing",
		PKColumn:   "id_detal",
		Columns:    "id_detal, ctacod, nrofac, nropue, regfec, artcod, detalle, artcan, artnet, artali, artiva, arttot, concod",
	}
}

// WithholdingIIBBReader creates a reader for RET1598.
func WithholdingIIBBReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RET1598",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   "id_ret1598",
		Columns:    "id_ret1598, ctacod, fecret, nroret, alicuota, totimpon, totret, ctanom",
	}
}

// WithholdingGananciasReader creates a reader for RETGANAN.
func WithholdingGananciasReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RETGANAN",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   "id_retganan",
		Columns:    "id_retganan, ctacod, ganfec, gannro, ganpor, ganbru, gantot",
	}
}

// WithholdingIVAReader creates a reader for RETIVA.
func WithholdingIVAReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RETIVA",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   "id_retiva",
		Columns:    "id_retiva, ctacod, ivafec, ivanro, ivapor, ivabru, ivatot",
	}
}

// AccountMovementReader creates a reader for REG_MOVIMIENTOS.
// This table tracks current account (cuenta corriente) movements.
func AccountMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REG_MOVIMIENTOS",
		Target:     "erp_account_movements",
		DomainName: "invoicing",
		PKColumn:   "id_regmovimiento",
		Columns:    "id_regmovimiento, regcuenta_id, fecha_movimiento, tipo_movimiento, importe_movimiento, saldo_movimiento, referencia_movimiento",
	}
}
