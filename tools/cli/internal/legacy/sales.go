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

// QuotationOptionReader — COTIZOPMOVIM (28,573 rows live, scrape
// 28,626). "OPCIONES POR COTIZACION" — free-text option lines per
// quotation section. idMovim PK + idCotiz FK + idSeccion + descripcion.
// Pareto tail Grupo B rank 3 (post-2.0.10).
func QuotationOptionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "COTIZOPMOVIM",
		Target:     "erp_quotation_section_items",
		DomainName: "sales",
		PKColumn:   "idMovim",
		Columns:    "idMovim, idCotiz, idSeccion, descripcion",
	}
}
