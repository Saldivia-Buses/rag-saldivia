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

// CashMovementReader and BankMovementReader are deferred — treasury movement migration
// requires mapping CAJ_MOVIMIENTOS + CAR_MOVIMIENTOS into a unified erp_treasury_movements
// table with complex movement_type resolution. Will be implemented alongside invoicing.

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
