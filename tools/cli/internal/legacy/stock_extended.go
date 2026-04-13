package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// BOM — Bill of Materials (STKPIEZA)
// ---------------------------------------------------------------------------

// BOMReader creates a reader for STKPIEZA (bill of materials / pieces).
// Has auto-increment id_pieza PK. 36K rows.
// articulo_hijo = child article code, idPadre = parent article code.
// cantidad = quantity per unit, cant_uso = usage quantity.
// bom_variacion_id = BOM variant, posicionfab_id = manufacturing position.
func BOMReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STKPIEZA",
		Target:     "erp_bom",
		DomainName: "stock",
		PKColumn:   "id_pieza",
		Columns: "id_pieza, articulo_hijo, idPadre, cantidad, costo_automatico, " +
			"recalc, recargo, cant_uso, bom_variacion_id, posicionfab_id, " +
			"desde_fecha, hasta_fecha, creacion_fecha, modificacion_fecha",
	}
}

// ---------------------------------------------------------------------------
// BOM History (STK_BOM_HIST)
// ---------------------------------------------------------------------------

// BOMHistoryReader creates a reader for STK_BOM_HIST (BOM cost history snapshots).
// Has auto-increment id_stkbomhist PK. 3.3M rows.
// Each row captures a point-in-time cost calculation for a BOM piece.
// level_0..level_7 = cost breakdown by BOM depth, piezas = total pieces count.
// tipocalculo_costo = cost calculation method, regcuenta_id = supplier entity.
func BOMHistoryReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STK_BOM_HIST",
		Target:     "erp_bom_history",
		DomainName: "stock",
		PKColumn:   "id_stkbomhist",
		Columns: "id_stkbomhist, pieza_id, bom_variacion_id, fecha_costo, " +
			"stkarticulohijo_id, stkarticulohijo_string, posicionfab_id, " +
			"tipocalculo_costo, regcuenta_id, cuenta_id, cantidad, " +
			"unidad_compra, unidad_uso, multiplo, unitario, recargo, " +
			"porcentaje_region, sumaitems, sumaitems_clog, " +
			"level_0, level_1, level_2, level_3, level_4, level_5, level_6, level_7, piezas",
	}
}

// ---------------------------------------------------------------------------
// Stock Levels (STK_STOCKACTUAL)
// ---------------------------------------------------------------------------

// StockLevelReader creates a reader for STK_STOCKACTUAL (current stock levels per article+warehouse).
// No auto-increment PK — composite key (stkarticulo_id, stkdeposito_id). 17K rows.
// cantidad_stock = current quantity, tipo_ajuste = last adjustment type.
func StockLevelReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "STK_STOCKACTUAL",
		Target:     "erp_stock_levels",
		DomainName: "stock",
		PKColumns:  []string{"stkdeposito_id", "stkarticulo_id"},
		Columns:    "stkarticulo_id, stkdeposito_id, cantidad_stock, actualizado, tipo_ajuste",
	}
}

// ---------------------------------------------------------------------------
// Price Lists (STK_LISTAS)
// ---------------------------------------------------------------------------

// PriceListReader creates a reader for STK_LISTAS (price list headers).
// Has auto-increment id_stklista PK. 1.1K rows.
// moneda_id = currency for cost, moneda_idvta = currency for sale price.
// cotizacion_genmonedahis = exchange rate snapshot used to generate the list.
// porcentaje_desc/porcentaje_desc1 = discount percentages.
func PriceListReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STK_LISTAS",
		Target:     "erp_price_lists",
		DomainName: "stock",
		PKColumn:   "id_stklista",
		Columns: "id_stklista, stklista_nombre, moneda_id, cotizacion_genmonedahis, " +
			"porcentaje_desc, porcentaje_desc1, moneda_idvta, cotizacion_monedavta, " +
			"stklista_descripcion, stklista_pie, stklista_fecha",
	}
}

// ---------------------------------------------------------------------------
// Price List Items (STK_LISTADETALLE)
// ---------------------------------------------------------------------------

// PriceListItemReader creates a reader for STK_LISTADETALLE (price list line items).
// Has auto-increment id_stklistadetalle PK. 138K rows.
// stklistadetalle_rentabilidad = margin %, stklistadetalle_precioventa = sale price,
// stklistadetalle_precioventa_2 = alternate sale price, stklistadetalle_costo = cost.
// dto_* columns = discount terms (percentage, date range, quantity threshold).
func PriceListItemReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STK_LISTADETALLE",
		Target:     "erp_price_list_items",
		DomainName: "stock",
		PKColumn:   "id_stklistadetalle",
		Columns: "id_stklistadetalle, stkarticulo_id, stklista_id, " +
			"stklistadetalle_rentabilidad, stklistadetalle_precioventa, " +
			"stklistadetalle_precioventa_2, stklistadetalle_costo, vigencia, " +
			"obervaciones, nombre, moneda_id, cotizacion, " +
			"dto_porcentaje, dto_fechadesde, dto_fechahasta, dto_cantidad, modificado",
	}
}
