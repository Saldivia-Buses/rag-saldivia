package legacy

import "database/sql"

// QuotationReader creates a reader for COTIZACION.
func QuotationReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "COTIZACION",
		Target:     "erp_quotations",
		DomainName: "sales",
		PKColumn:   "idCotiz",
		Columns:    "idCotiz, fecha, ctacod, ctanom, total, moneda_id, formaPago",
	}
}

// OrderReader creates a reader for PEDIDOINT (internal purchase requisitions / orders).
func OrderReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PEDIDOINT",
		Target:     "erp_orders",
		DomainName: "sales",
		PKColumn:   "idPed",
		Columns:    "idPed, artcod, fechaemi, codprv, estado, usuario, nrocom, tipocomp",
	}
}
