package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Quotation Lines (COTIZOPCIONES)
// ---------------------------------------------------------------------------

// QuotationLineReader creates a reader for COTIZOPCIONES (quotation option lines).
// Composite PK (idPrecio, idOpcion). 882 rows.
// Each row is a priced option within a quotation. idSeccion = section grouping.
func QuotationLineReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "COTIZOPCIONES",
		Target:     "erp_quotation_lines",
		DomainName: "sales",
		PKColumns:  []string{"idPrecio", "idOpcion"},
		Columns:    "idPrecio, idOpcion, descripcion, idSeccion",
	}
}

// ---------------------------------------------------------------------------
// Quotation Other Items (COTIZOTROS)
// ---------------------------------------------------------------------------

// QuotationOtherReader creates a reader for COTIZOTROS (quotation miscellaneous items).
// Has auto-increment idMovim PK. 97 rows.
// idCotiz = parent quotation, idSeccion = section, descripcion = free-text description.
func QuotationOtherReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "COTIZOTROS",
		Target:     "erp_quotation_lines",
		DomainName: "sales",
		PKColumn:   "idMovim",
		Columns:    "idMovim, idCotiz, idSeccion, descripcion",
	}
}

// ---------------------------------------------------------------------------
// Customer Orders (PEDCOTIZ)
// ---------------------------------------------------------------------------

// CustomerOrderReader creates a reader for PEDCOTIZ (pedidos de cotizacion / customer orders).
// Has auto-increment id PK. 3.8K rows.
// Massive table with ~100 columns describing vehicle specifications per order:
// seat types, air conditioning, audio, windows, painting, lighting, etc.
// ctacod = customer entity code, nrocha = vehicle/chassis number,
// pedfec = order date, terminacion = scheduled completion date.
// cotizacion_id = linked quotation, homologacion_id = vehicle homologation.
func CustomerOrderReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PEDCOTIZ",
		Target:     "erp_orders",
		DomainName: "sales",
		PKColumn:   "id",
		Columns: "id, arruga, bagno, barcom, cabeza, cabasinc, calcan, caltip, " +
			"calefn, calobs, camare, cantasi, canttv, cartel, cerrad, colair, " +
			"colaud, colcor, colint, colvid, conaac, contacto, cortina, ctacod, " +
			"ctepin, tipcartel, email, equipa, habili, marair, marasi, maresp, " +
			"microf, miniba, modair, nroasi, nrocar, nrocha, obsesp, paqluz, " +
			"pedfec, pednro, piston, porpaq, posair, preair, preaud, previd, " +
			"prvcod, ptabga, reloj, revfec, revnro, siscod, telnro, tipasi, " +
			"tipcal, tipcor, tipesp, tiphab, tippin, tippta, tipvtn, trabas, " +
			"vtnmov, tipserv, localidades, pana, colorllantas, colorespej, " +
			"portavasos, portalatas, mesas, condensador, gabinetes, tipotv, " +
			"prvtv, jacks, prvaudio, vtnpuerta, ninterno, idconta, colorPana, " +
			"luneta, cortePint, fotoPana, cucheta, terminacion, fecha_herreria, " +
			"fecha_chaperia, terminacionFinal, ingresoLinea, ingresoHora, " +
			"tanque_id, homologacion_id, login, estado_ficha, txtcabasinc, " +
			"cabasrem, txtcabasrem, apbracen, cotizacion_id, ultima_modificacion, " +
			"marcado_pendientes, entrega_probable, empresa_id",
	}
}
