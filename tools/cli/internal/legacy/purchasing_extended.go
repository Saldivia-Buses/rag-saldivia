package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Internal Requisitions (PEDIDOINT)
// ---------------------------------------------------------------------------

// InternalRequisitionReader creates a reader for PEDIDOINT (pedidos internos / internal requisitions).
// Has auto-increment idPed PK. 384K rows.
// artcod = article code, cant = requested quantity, movcps = qty ordered from supplier,
// movuso = qty used, pencps = pending purchase, penuso = pending usage.
// estado = workflow state, codprv = supplier code, nrofab = vehicle/unit number.
// idDeposito = warehouse, cpsproyecto_id = project, centrocostos_id = cost center.
func InternalRequisitionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PEDIDOINT",
		Target:     "erp_purchase_orders",
		DomainName: "purchasing",
		PKColumn:   "idPed",
		Columns: "idPed, artcod, artuni, cant, codprv, estado, fechaemi, fecreq, " +
			"movcps, movref, movuso, nrocom, nrofab, operador, pencps, penuso, " +
			"puesto, renglon, tipocomp, unicps, unimed, usuario, " +
			"idDeposito, ocpnro, confirmado, mensaje, archivo, directorio, " +
			"cpsproyecto_id, centrocostos_id",
	}
}

// ---------------------------------------------------------------------------
// Purchase Receipts (OCPRECIB)
// ---------------------------------------------------------------------------

// PurchaseReceiptReader creates a reader for OCPRECIB (recepciones de compra).
// Has auto-increment id_recepcion PK. 320K rows.
// Links to purchase order via (ocpfec, ocpnro) and internal requisition via idPed.
// reccps = qty received for purchase, recuso = qty received for usage.
// pedcps = qty ordered purchase, peduso = qty ordered usage.
// siscod = subsystem code, ctacod = supplier entity code.
func PurchaseReceiptReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "OCPRECIB",
		Target:     "erp_purchase_receipts",
		DomainName: "purchasing",
		PKColumn:   "id_recepcion",
		Columns: "id_recepcion, artcod, codest, ctacod, fecrec, fecreq, nrorec, " +
			"ocpfec, ocpnro, pedcps, peduso, reccps, recuso, siscod, " +
			"idPed, ocpren, observacion, vencimiento",
	}
}
