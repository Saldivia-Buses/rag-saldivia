package legacy

import "database/sql"

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

// CashMovementReader creates a reader for CAJ_MOVIMIENTOS.
func CashMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAJ_MOVIMIENTOS",
		Target:     "erp_treasury_movements",
		DomainName: "treasury",
		PKColumn:   "id_cajmovimiento",
		Columns:    "id_cajmovimiento, cajconcepto_id, cajpuesto_id, importe_movimiento, referencia_movimiento, fecha_movimiento, cajregistro_id, usuario_id",
	}
}

// BankMovementReader creates a reader for CAR_MOVIMIENTOS.
func BankMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CAR_MOVIMIENTOS",
		Target:     "erp_treasury_movements",
		DomainName: "treasury",
		PKColumn:   "id_carmovimiento",
		Columns:    "id_carmovimiento, tipo_movimiento, referencia_movimiento, fecha_movimiento, cajmovimiento_id, regcuenta_id",
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
