package legacy

import "database/sql"

// AccountReader creates a reader for CTB_CUENTAS (plan de cuentas).
// PK is composite (id_cuenta varchar, ejercicio_id int) but id_ctbcuenta is AI unique.
func AccountReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CTB_CUENTAS",
		Target:     "erp_accounts",
		DomainName: "accounting",
		PKColumn:   "id_ctbcuenta",
		Columns:    "id_ctbcuenta, id_cuenta, nombre_cuenta, cuenta_superior_id, tipo_id, habilitada, ctbcentro_id, ejercicio_id",
	}
}

// CostCenterReader creates a reader for CTB_CENTROS.
func CostCenterReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CTB_CENTROS",
		Target:     "erp_cost_centers",
		DomainName: "accounting",
		PKColumn:   "id_ctbcentro",
		Columns:    "id_ctbcentro, nombre_centro, referencia_centro",
	}
}

// FiscalYearReader creates a reader for CTB_EJERCICIOS.
func FiscalYearReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CTB_EJERCICIOS",
		Target:     "erp_fiscal_years",
		DomainName: "accounting",
		PKColumn:   "id_ejercicio",
		Columns:    "id_ejercicio, nombre_ejercicio, comienza, finaliza, cerrado, activo, cuenta_resultado_id",
	}
}

// JournalEntryReader creates a reader for CTB_MOVIMIENTOS (asientos cabecera).
func JournalEntryReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CTB_MOVIMIENTOS",
		Target:     "erp_journal_entries",
		DomainName: "accounting",
		PKColumn:   "id_movimiento",
		Columns:    "id_movimiento, nro_minuta, fecha_movimiento, ejercicio_id, referencia, concepto_id, usuario_modificacion, debe, haber",
	}
}

// JournalLineReader creates a reader for CTB_DETALLES (asientos detalle).
// Note: CTB_DETALLES.importe is varchar(45) — needs CAST.
// The entry date is resolved from the already-migrated journal entry in PostgreSQL.
func JournalLineReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CTB_DETALLES",
		Target:     "erp_journal_lines",
		DomainName: "accounting",
		PKColumn:   "id_detalle",
		Columns:    "id_detalle, movimiento_id, cuenta_id, ctbcentro_id, doh, CAST(importe AS DECIMAL(16,2)) as importe, referencia, orden",
	}
}

// AccountingRegisterReader creates a reader for CTBREGIS (legacy accounting
// line log, 604,579 rows as of 2026-04-18 — Pareto #3 of the Phase 1 §Data
// migration gap post-2.0.9, ~28 % of the remaining uncovered row volume).
//
// CTBREGIS is the leaf-level debe/haber line under the old (pre-CTB_MOVIMIENTOS)
// accounting schema. It is still read and written by the live Histrix UI:
// contabilidad/qry/libro_diario_qry.xml joins through CTB_DETALLES.ctbregis_id
// to pull the legacy ctbcod; proveedores_loc + clientes_local + ordenpago +
// iva + estadisticas + anulaciones all reference CTBREGIS directly. Not a
// waiver candidate.
//
// Shape: id_ctbregis INT AI PRI, siscod, regfec (zero-date for 122 rows),
// regmin (minuta), regtip, regnro, ctbcod VARCHAR(12), regdoh TINYINT
// (1=debe/2=haber), regimp DECIMAL(13,2), regref, regfco, regpoa VARCHAR(1),
// plus cost-center / imputation / order metadata.
func AccountingRegisterReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CTBREGIS",
		Target:     "erp_accounting_registers",
		DomainName: "accounting",
		PKColumn:   "id_ctbregis",
		Columns: "id_ctbregis, siscod, regfec, regfco, regmin, regtip, regnro, " +
			"ctbcod, regdoh, regimp, regref, regpoa, coscod, impcod, " +
			"idcos, idimpu, regcta, regnpv, regord, regsub, reguni",
	}
}
