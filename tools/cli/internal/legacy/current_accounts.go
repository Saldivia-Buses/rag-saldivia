package legacy

import "database/sql"

// AccountMovementReader creates a reader for REG_MOVIMIENTOS (current account movements).
// ~291K rows. PK is id_regmovim (bigint auto_increment).
// This is the most complex legacy table — 70+ columns covering invoices, credit notes,
// debit notes, receipts, and all financial document types across subsystems.
// Maps to erp_account_movements.
func AccountMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REG_MOVIMIENTOS",
		Target:     "erp_account_movements",
		DomainName: "current_account",
		PKColumn:   "id_regmovim",
		Columns: "id_regmovim, subsistema_id, regcuenta_id, referencia_movimiento, " +
			"concepto_id, puesto_movimiento, cuenta_movimiento, letra_movimiento, " +
			"vencimiento_movimiento, reg_periodo, reg_fecha, neto_movimiento, " +
			"importe_iva, impmi, importe_periodo, alicuota_iva, alirni, " +
			"saldo_movimiento, importe_movimiento, pag_movimiento, mar_movimiento, " +
			"rec_movimiento, pun_movimiento, novedades_movimiento, " +
			"bonificaciones_movimiento, fpv_movimiento, fecha_movimiento, " +
			"cpsmovimiento_id, nro_movimiento, cuota_movimiento, importe_exento, " +
			"importe_nogravado, importe_percepcion, fecha_cae, numero_cae, " +
			"regconcepto_afip, referencia_afip, tiposervicio_afip, " +
			"cajmovimiento_id, conciliado, cajvalor_id, regcuentavendedor_id, " +
			"cae_i, regcuentatransporte_id, regforma_id, domicilio_movimiento, " +
			"cantidad_bultos, valor_declarado, nropag, nro_cuenta, concod, " +
			"presentado, login, regmin, bienes_capital, codigo_numero, nrofab, " +
			"descuento_1, descuento_2, nombre_movimiento, cuit_movimiento, " +
			"dni_movimiento, movicc, movinc, movipv, movref__3, remito, succod, " +
			"zoncod, iva_id, catganancias_id, importe_ibr, importe_bon, moneda_id, " +
			"logregion_id, regmovimrel_id, genprovincia_id",
	}
}

// AccountMovementDetailReader creates a reader for REG_DETALLES (movement line items).
// ~234K rows. PK is id_regdetalle (int auto_increment).
// Sub-lines per account movement: articles, services, discounts, tax breakdowns.
// Maps as metadata on erp_account_movements.
func AccountMovementDetailReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REG_DETALLES",
		Target:     "erp_account_movement_details",
		DomainName: "current_account",
		PKColumn:   "id_regdetalle",
		Columns: "id_regdetalle, regmovim_id, sercod, descripcion_detalle, " +
			"codigo_detalle, iva_detalle, rni_detalle, cantidad_detalle, " +
			"total_detalle, iva_servicio, reg_servicio, alicuota_iva, " +
			"total_percepcion, ctbdeb, ctbhab, orden, neto_detalle, " +
			"precio_venta, exento_detalle, per_detalle, facservicio_id, " +
			"cuenta_debe_id, cuenta_haber_id, tabla_detalle, porcentaje_desc0, " +
			"porcentaje_desc1, porcentaje_desc2, bonificacion_1, bonificacion_2, " +
			"bonificacion_3, cantidad_pendiente, cantidad_entregada, " +
			"logincidencia_id, stkrubro_id, codigo_serie, nro_partida, color_id, " +
			"referencia_detalle, uid",
	}
}

// PaymentAllocationHeaderReader creates a reader for CCTENCAB (payment allocation headers).
// ~9K rows. PK is composite (concod, ctacod, movfec, movlet, movnpv, movnro, siscod, succod).
// Represents the header of a payment allocation linking receipts to invoices.
// Maps to erp_payment_allocations.
func PaymentAllocationHeaderReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "CCTENCAB",
		Target:     "erp_payment_allocations",
		DomainName: "current_account",
		PKColumns:  []string{"concod", "ctacod", "movfec", "movlet", "movnpv", "movnro", "siscod", "succod"},
		Columns: "camdes, cancuo, carobs__1, carobs__2, carobs__3, carobs__4, carobs__5, " +
			"concod, ctacod, ctajud, ctanom, encges, encsel, feccar, fecexc, moncod, " +
			"movcam, movent, movfec, movfin, movimp, movlet, movnpv, movnro, movseg, " +
			"movsen, ncarga, opecla, seccod, siscod, submai, succod, vencod, vtacod, " +
			"vtamai, vtasub, nrofab",
	}
}

// PaymentAllocationLineReader creates a reader for CCTIMPUT (payment allocation lines).
// ~132K rows. PK is id_cctimput (int auto_increment).
// Each line represents the application of a payment to a specific invoice/document.
// Links to erp_payment_allocations and erp_account_movements via regmovim_id.
func PaymentAllocationLineReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "CCTIMPUT",
		Target:     "erp_payment_allocations",
		DomainName: "current_account",
		PKColumn:   "id_cctimput",
		Columns: "id_cctimput, concod, ctacod, movcta, movfec, movicc, movife, movimp, " +
			"movinc, movint, movipv, movlet, movnpv, movnro, movsal, siscod, succod, " +
			"regmin, ctacoddes, fecha, nroimp, saldo_original, regmovim_id, regmovim0_id",
	}
}

// EntityCreditRatingReader — REG_CUENTA_CALIFICACION (136,064 rows live,
// scrape 58,960, +131 %). Customer / supplier credit rating history, one
// row per rating event. FK regcuenta_id → REG_CUENTA(id_regcuenta) — a
// straight entity-domain resolve via ResolveOptional. Pareto tail Grupo A
// rank 1 (post-2.0.10). Targets erp_entity_credit_ratings.
func EntityCreditRatingReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REG_CUENTA_CALIFICACION",
		Target:     "erp_entity_credit_ratings",
		DomainName: "current_account",
		PKColumn:   "id_regcalificacion",
		Columns: "id_regcalificacion, regcuenta_id, calificacion, " +
			"fecha_calificacion, referencia_calificacion",
	}
}

// InvoiceNoteReader — REG_MOVIMIENTO_OBS (72,737 rows live, scrape 72,737).
// Per-invoice free-text notes attached to REG_MOVIMIENTOS. FK regmovim_id
// resolves via BuildRegMovimIndex (Phase 6). Targets erp_invoice_notes.
func InvoiceNoteReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "REG_MOVIMIENTO_OBS",
		Target:     "erp_invoice_notes",
		DomainName: "current_account",
		PKColumn:   "id_regmovimientoobs",
		Columns: "id_regmovimientoobs, fec_observacion, hora_observacion, " +
			"observacion, regmovim_id, login, gencontacto_id, tabla_origen, " +
			"siscod, movfec, ctacod, concod, movnpv, movnro",
	}
}
