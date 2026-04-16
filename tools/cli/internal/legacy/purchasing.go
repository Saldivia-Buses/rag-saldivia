package legacy

import "database/sql"

// PurchaseOrderReader creates a reader for CPS_MOVIMIENTOS (replaced CPSENCAB).
func PurchaseOrderReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CPS_MOVIMIENTOS",
		Target:     "erp_purchase_orders",
		DomainName: "purchasing",
		PKColumn:   "id_cpsmovimiento",
		Columns:    "id_cpsmovimiento, cps_numero, cpsfecha, regcuenta_id, cpsestado_id, moneda_id, cpsimporte, cpsobservacion, usuario",
	}
}

// PurchaseOrderLineReader creates a reader for CPS_DETALLE.
func PurchaseOrderLineReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CPS_DETALLE",
		Target:     "erp_purchase_order_lines",
		DomainName: "purchasing",
		PKColumn:   "id_cpsdetalle",
		Columns:    "id_cpsdetalle, cpsmovimiento_id, stkarticulo_id, cantidad_compra, costo_unitario, recibido_compra",
	}
}
