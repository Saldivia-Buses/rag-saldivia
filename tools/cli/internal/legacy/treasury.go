package legacy

import "database/sql"

// BankImportReader — BCS_IMPORTACION (91,959 rows live, scrape 84,492).
// Bank-statement import staging: rows arrive from supplier CSV/XLS
// dumps of bank movements and sit here until concil screens match them
// against internal REG_MOVIMIENTOS. Live UI: bancos_local/
// bcs_importacion_qry.xml + surrounding bcsmovim_* forms. Pareto rank 1
// of the remaining long tail post-Pareto #8.
func BankImportReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "BCS_IMPORTACION",
		Target:     "erp_bank_imports",
		DomainName: "treasury",
		PKColumn:   "id_importacion",
		Columns: "id_importacion, fecha_movimiento, nombre_concepto, " +
			"nro_movimiento, importe, debito, credito, saldo, " +
			"cod_movimiento, regmovim_id, importado, nro_cuenta, " +
			"procesado, comentarios, nro_interno, sucursal",
	}
}

// BankAccountReader creates a reader for CAR_BANCOS.
func BankAccountReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAR_BANCOS",
		Target:     "erp_bank_accounts",
		DomainName: "treasury",
		PKColumn:   "id_carbanco",
		Columns:    "id_carbanco, nombre_banco",
	}
}

// CashRegisterReader creates a reader for CAJ_PUESTOS.
func CashRegisterReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAJ_PUESTOS",
		Target:     "erp_cash_registers",
		DomainName: "treasury",
		PKColumn:   "id_cajpuesto",
		Columns:    "id_cajpuesto, descripcion_puesto",
	}
}

// CheckReader creates a reader for CARCHEQU (cheques).
// PK is composite (carint, siscod, succod) but we use carint for ordering.
func CheckReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CARCHEQU",
		Target:     "erp_checks",
		DomainName: "treasury",
		PKColumn:   "carint",
		Columns:    "carint, cartip, carnro, carbco, carimp, carfec, caracr, ctacod, cardes, carobv, fecha_emision, siscod",
	}
}

// CashMovementReader creates a reader for CAJMOVIM (cash register movements).
// ~116K rows. PK is composite (cajcod, cajcta, cajfec, cajnpv, cajnro, concod, regmin, siscod, succod).
// Maps to erp_treasury_movements with movement_type resolution from cajcod/concod.
func CashMovementReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "CAJMOVIM",
		Target:     "erp_treasury_movements",
		DomainName: "treasury",
		PKColumns:  []string{"cajcod", "cajcta", "cajfec", "cajnpv", "cajnro", "concod", "regmin", "siscod", "succod"},
		Columns: "cajcod, cajcta, cajest, cajfec, cajhor, cajimp, cajnom, cajnpv, cajnro, " +
			"cajpla, cajref__1, cajref__2, cajref__3, codent, concod, coscod, impcod, " +
			"opecla, opecod, regmin, siscod, succod",
	}
}

// CashVoucherReader creates a reader for CAJCOMPR (cash voucher references).
// ~108K rows. PK is composite (nrocom, movord).
// Links voucher numbers to cash movements in erp_treasury_movements.
func CashVoucherReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "CAJCOMPR",
		Target:     "erp_treasury_movements",
		DomainName: "treasury",
		PKColumns:  []string{"nrocom", "movord"},
		Columns: "concod, condic, ctaant, ctacod, ctanom, entreg, movco1, movco2, movcta, " +
			"movfec, movicc, movict, movim1, movim2, movimp, movinc, movipv, movlet, " +
			"movnet, movnpv, movnro, movord, movref__1, movref__2, movref__3, movsa1, " +
			"movsa2, movsal, movuni, movvto, nrocom, opecla, prineg, regmin, siscod, " +
			"succod, vencod, zoncod",
	}
}

// BankMovementReader creates a reader for CAR_MOVIMIENTOS (bank movements).
// ~162 rows. PK is id_carmovimiento (auto_increment).
// Maps to erp_treasury_movements linking bank account, cash, and check references.
func BankMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAR_MOVIMIENTOS",
		Target:     "erp_treasury_movements",
		DomainName: "treasury",
		PKColumn:   "id_carmovimiento",
		Columns:    "id_carmovimiento, tipo_movimiento, referencia_movimiento, fecha_movimiento, cajmovimiento_id, carvalor_id, regcuenta_id",
	}
}

// CashCountReader creates a reader for CAJ_PUESTO_ARQUEOS (cash register counts/reconciliation).
// ~14 rows. PK is id_cajpuestoarqueo (auto_increment).
// Maps to erp_cash_counts for cash register balancing records.
func CashCountReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAJ_PUESTO_ARQUEOS",
		Target:     "erp_cash_counts",
		DomainName: "treasury",
		PKColumn:   "id_cajpuestoarqueo",
		Columns:    "id_cajpuestoarqueo, cajpuesto_id, cajformapago_id, orden_arqueo",
	}
}
