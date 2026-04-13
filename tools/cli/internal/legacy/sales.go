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

// QuotationLineReader, QuotationOtherReader, CustomerOrderReader
// are in sales_extended.go.
