package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Invoices (FACREMIT, FACREMIMP)
// ---------------------------------------------------------------------------

// InvoiceReader creates a reader for FACREMIT (facturas + remitos de venta/compra).
// FACREMIT has no auto-increment PK — it uses composite (ctacod, remfec, remnpv, remnro).
// movfec__N / movnpv__N / movnro__N are denormalized payment schedule columns (up to 10 slots).
// The transform function flattens these into rows or ignores them as needed.
func InvoiceReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "FACREMIT",
		Target:     "erp_invoices",
		DomainName: "invoicing",
		PKColumns:  []string{"ctacod", "remfec", "remnpv", "remnro"},
		Columns: "ctacod, remfec, remnpv, remnro, remefe, remval, remrea, " +
			"remmod, remlio, rembul, tracod, cierre, ocpnro, ocpnpv, ocpfec, " +
			"movfec__1, movfec__2, movfec__3, movfec__4, movfec__5, " +
			"movfec__6, movfec__7, movfec__8, movfec__9, movfec__10, " +
			"movnpv__1, movnpv__2, movnpv__3, movnpv__4, movnpv__5, " +
			"movnpv__6, movnpv__7, movnpv__8, movnpv__9, movnpv__10, " +
			"movnro__1, movnro__2, movnro__3, movnro__4, movnro__5, " +
			"movnro__6, movnro__7, movnro__8, movnro__9, movnro__10",
	}
}

// InvoiceImportReader creates a reader for FACREMIMP (facturas de importación).
// Has auto-increment id_facremimp PK. Links to a REG_MOVIMIENTOS entry and a FACREMIT row
// via (ctacod, movfec, movnpv, movnro) and (remfec, remnpv, remnro).
func InvoiceImportReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "FACREMIMP",
		Target:     "erp_invoices",
		DomainName: "invoicing",
		PKColumn:   "id_facremimp",
		Columns:    "id_facremimp, ctacod, movfec, movnpv, movnro, remfec, remnpv, remnro, regmovim_id",
	}
}

// ---------------------------------------------------------------------------
// Invoice lines (FACDETAL)
// ---------------------------------------------------------------------------

// InvoiceLineReader creates a reader for FACDETAL (detalle de facturas).
// Has auto-increment id_detal PK. Linked to FACREMIT via (ctacod, regfec, remnro).
// artnet is unit price, artali is IVA rate, artiva is IVA amount, arttot is line total.
func InvoiceLineReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "FACDETAL",
		Target:     "erp_invoice_lines",
		DomainName: "invoicing",
		PKColumn:   "id_detal",
		Columns: "id_detal, ctacod, regfec, remnro, nropue, concod, " +
			"artcod, artcan, artnet, artcom, artsal, artniv, " +
			"artali, artiva, arttot, detalle, " +
			"nrofac, nroimp, pueimp, siscod, movfec, movlet, regmin, regmovim_id",
	}
}

// ---------------------------------------------------------------------------
// Delivery notes (REMITO, REMDETAL)
// ---------------------------------------------------------------------------

// DeliveryNoteReader creates a reader for REMITO (remitos de compra).
// Composite PK (numero, puesto). Links to entity via ctacod and vehicle via nrofab.
func DeliveryNoteReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "REMITO",
		Target:     "erp_invoices",
		DomainName: "invoicing",
		PKColumns:  []string{"numero", "puesto"},
		Columns:    "numero, puesto, fecha, nrofab, ctacod",
	}
}

// DeliveryNoteLineReader creates a reader for REMDETAL (detalle de remitos de compra).
// Has auto-increment idRemdet PK. Linked to REMITO via idRemito.
func DeliveryNoteLineReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REMDETAL",
		Target:     "erp_invoice_lines",
		DomainName: "invoicing",
		PKColumn:   "idRemdet",
		Columns: "idRemdet, idRemito, artcod, unidadCompra, cantCompra, " +
			"fecreq, movref, cantUso, nrofab, " +
			"pendienteCompra, PendienteUso, unidadUso, idPed",
	}
}

// ---------------------------------------------------------------------------
// VAT ledger — Sales (IVAVENTAS, IVAVENTAS2)
// ---------------------------------------------------------------------------

// SalesVATReader creates a reader for IVAVENTAS (IVA ventas — new format with AI PK).
// Has auto-increment id_ivaventa PK. 9K rows.
// codcom = voucher type, codlet = letter (A/B/C), nronpv = point of sale,
// nrocom = voucher number, execom = exempt, nogcom = non-taxable,
// rnicom = not registered, totcom = total.
func SalesVATReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "IVAVENTAS",
		Target:     "erp_tax_entries",
		DomainName: "invoicing",
		PKColumn:   "id_ivaventa",
		Columns: "id_ivaventa, codcom, codlet, ctacod, ctanom, " +
			"execom, feciva, fecreg, ivacod, nogcom, " +
			"nrocom, nrodoc, nronpv, rnicom, tipdoc, totcom, regmovim_id",
	}
}

// SalesVATLegacyReader creates a reader for IVAVENTAS2 (IVA ventas — legacy format, no AI PK).
// Composite PK (feciva, codcom, codlet, nronpv, nrocom, ctacod). 3K rows.
// Same column structure as IVAVENTAS but without id_ivaventa and regmovim_id.
func SalesVATLegacyReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "IVAVENTAS2",
		Target:     "erp_tax_entries",
		DomainName: "invoicing",
		PKColumns:  []string{"feciva", "codcom", "codlet", "nronpv", "nrocom", "ctacod"},
		Columns: "codcom, codlet, ctacod, ctanom, " +
			"execom, feciva, fecreg, ivacod, nogcom, " +
			"nrocom, nrodoc, nronpv, rnicom, tipdoc, totcom",
	}
}

// ---------------------------------------------------------------------------
// VAT ledger — Purchases (IVACOMPRAS, IVACOMPS)
// ---------------------------------------------------------------------------

// PurchaseVATReader creates a reader for IVACOMPRAS (IVA compras — new format with AI PK).
// Has auto-increment id_ivacompra PK. 125K rows.
// Same column semantics as IVAVENTAS but for purchases. Includes refresh flag.
func PurchaseVATReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "IVACOMPRAS",
		Target:     "erp_tax_entries",
		DomainName: "invoicing",
		PKColumn:   "id_ivacompra",
		Columns: "id_ivacompra, codcom, codlet, ctacod, ctanom, " +
			"execom, feciva, fecreg, ivacod, nogcom, " +
			"nrocom, nrodoc, nronpv, rnicom, tipdoc, totcom, regmovim_id, refresh",
	}
}

// PurchaseVATLegacyReader creates a reader for IVACOMPS (IVA compras — legacy denormalized format).
// Composite PK (siscod, succod, ivcfec, ivclet, ivcnpv, ivcnro, ctacod). 12K rows.
// Heavily denormalized: up to 10 cost-center slots (coscod__N), 3 IVA rate tiers
// (ivcali__N, ivcgra__N, ivciva__N), 10 tax-code amounts (ivciim__N, ivciix__N),
// and 4 withholding buckets (ivcre1__N, ivcre2, ivcre3, ivcre4__N).
func PurchaseVATLegacyReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "IVACOMPS",
		Target:     "erp_tax_entries",
		DomainName: "invoicing",
		PKColumns:  []string{"siscod", "succod", "ivcfec", "ivclet", "ivcnpv", "ivcnro", "ctacod"},
		Columns: "siscod, succod, ivcfec, ivclet, ivcnpv, ivcnro, ctacod, " +
			"ctanom, ctanro, doccod, ivacod, comcod, concod, prvcod, " +
			"ivcexe, ivcnog, ivctot, ivcint, ivc063, ivccam, ivcuni, " +
			"ivclis, ivcreg, ivcman, regmin, " +
			"ivcali__1, ivcali__2, ivcali__3, " +
			"ivcgra__1, ivcgra__2, ivcgra__3, " +
			"ivciva__1, ivciva__2, ivciva__3, " +
			"ivcrni__1, ivcrni__2, ivcrni__3, " +
			"ivciim__1, ivciim__2, ivciim__3, ivciim__4, ivciim__5, " +
			"ivciim__6, ivciim__7, ivciim__8, ivciim__9, ivciim__10, " +
			"ivciix__1, ivciix__2, ivciix__3, ivciix__4, ivciix__5, " +
			"ivciix__6, ivciix__7, ivciix__8, ivciix__9, ivciix__10, " +
			"ivccim__1, ivccim__2, ivccim__3, ivccim__4, ivccim__5, " +
			"ivccim__6, ivccim__7, ivccim__8, ivccim__9, ivccim__10, " +
			"coscod__1, coscod__2, coscod__3, coscod__4, coscod__5, " +
			"coscod__6, coscod__7, coscod__8, coscod__9, coscod__10, " +
			"ivcre1__1, ivcre1__2, ivcre1__3, " +
			"ivcre2, ivcre3, " +
			"ivcre4__1, ivcre4__2, ivcre4__3",
	}
}

// ---------------------------------------------------------------------------
// VAT amounts (IVAIMPORTES)
// ---------------------------------------------------------------------------

// VATAmountReader creates a reader for IVAIMPORTES (importes de IVA por alícuota).
// Has auto-increment id_ivaimporte PK. 242K rows — the biggest tax table.
// Each row is one tax-rate line for one voucher. Links to IVAVENTAS/IVACOMPRAS
// via ivacv_id. tipoiva distinguishes sales (1) vs purchases (2).
// impali = tax rate %, impgra = taxable base, impimp = tax amount.
func VATAmountReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "IVAIMPORTES",
		Target:     "erp_tax_entries",
		DomainName: "invoicing",
		PKColumn:   "id_ivaimporte",
		Columns: "id_ivaimporte, codcom, codgra, codlet, ctacod, feciva, " +
			"impali, impcod, impgra, impimp, nrocom, nronpv, tipoiva, ivacv_id",
	}
}

// ---------------------------------------------------------------------------
// Withholdings — IIBB (RETACUMU)
// ---------------------------------------------------------------------------

// WithholdingIIBBReader creates a reader for RETACUMU (retenciones IIBB acumuladas).
// Composite PK (siscod, ctacod, acufec, year, mes, nropag). 39K rows.
// acuret = withholding amount, acupag = payment amount.
func WithholdingIIBBReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "RETACUMU",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumns:  []string{"siscod", "ctacod", "acufec", "year", "mes", "nropag"},
		Columns:    "siscod, ctacod, acufec, year, mes, nropag, acuret, acupag",
	}
}

// ---------------------------------------------------------------------------
// Withholdings — Gains (RETGANAN)
// ---------------------------------------------------------------------------

// WithholdingGainsReader creates a reader for RETGANAN (retenciones ganancias).
// Has auto-increment id_retganan (UNIQUE). 9.8K rows.
// ganfec = date, ganbru = gross, gannet = net, ganpor = rate %, gantot = total.
// catganancias_id = gains category.
func WithholdingGainsReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RETGANAN",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   "id_retganan",
		Columns: "id_retganan, concod, ctacod, ganacu, ganbru, ganfec, gannet, " +
			"gannro, ganpor, gantot, movnro, regmin, siscod, catganancias_id, catgan",
	}
}

// ---------------------------------------------------------------------------
// Withholdings — IVA (RETIVA)
// ---------------------------------------------------------------------------

// WithholdingIVAReader creates a reader for RETIVA (retenciones/percepciones IVA).
// Has auto-increment id_retiva (UNIQUE). 33 rows.
// ivafec = date, ivabru = gross, ivanet = net, ivapor = rate %, ivatot = total.
func WithholdingIVAReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RETIVA",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   "id_retiva",
		Columns: "id_retiva, concod, ctacod, ivaacu, ivabru, ivafec, ivanet, " +
			"ivanro, ivapor, ivatot, movnro, regmin, siscod, cativa_id, cativa",
	}
}

// ---------------------------------------------------------------------------
// Withholdings — IIBB Reg 1598 (RET1598)
// ---------------------------------------------------------------------------

// Withholding1598Reader creates a reader for RET1598 (retenciones IIBB régimen 1598).
// Has auto-increment id_ret1598 (UNIQUE). 8.8K rows.
// Includes entity denormalized fields (ctanom, ctanro, ctadir, ctaloc, ctanib)
// because Histrix stored them inline for IIBB reporting.
func Withholding1598Reader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "RET1598",
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   "id_ret1598",
		Columns: "id_ret1598, alicuota, codcomp, codnpv, compfec, compnro, " +
			"ctacod, ctadir, ctaloc, ctanib, ctanom, ctanro, " +
			"fecret, nroret, totcomp, totimpon, totret",
	}
}

// ---------------------------------------------------------------------------
// Backward-compatible generic withholding reader (used by stub migrators)
// ---------------------------------------------------------------------------

// WithholdingReader creates a generic reader for any withholding table by PK column.
// Deprecated: use the specific WithholdingGainsReader, WithholdingIVAReader,
// Withholding1598Reader, or WithholdingIIBBReader instead. This exists only to
// keep the stub migrators compiling until they are rewritten with correct transforms.
func WithholdingReader(db *sql.DB, table, pkColumn string) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      table,
		Target:     "erp_withholdings",
		DomainName: "invoicing",
		PKColumn:   pkColumn,
		Columns:    pkColumn + ", ctacod",
	}
}
