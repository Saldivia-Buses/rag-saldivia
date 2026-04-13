package legacy

import "database/sql"

// WarehouseReader creates a reader for STK_DEPOSITOS (DEPOSITO is a view).
func WarehouseReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STK_DEPOSITOS",
		Target:     "erp_warehouses",
		DomainName: "stock",
		PKColumn:   "id_stkdeposito",
		Columns:    "id_stkdeposito, nombre_deposito, direccion_deposito, inactivo",
	}
}

// ArticleReader creates a reader for STK_ARTICULOS.
// PK is composite (id_stkarticulo varchar, subsistema_id varchar) but we need
// a synthetic AI column for ordering. We'll use a subquery with ROW_NUMBER.
// For simplicity, we read using the UNIQUE combination.
func ArticleReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STK_ARTICULOS",
		Target:     "erp_articles",
		DomainName: "stock",
		PKColumn:   "id_stkarticulo",
		Columns:    "id_stkarticulo, nombre_articulo, stkfamilia_id, stkrubro_id, unidad_id, tipo_articulo, stock_minimo, stock_maximo, stock_reposicion, precio_costo, precio_promedio, baja_articulo, subsistema_id",
	}
}

// StockMovementReader creates a reader for STK_MOVIMIENTOS.
func StockMovementReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STK_MOVIMIENTOS",
		Target:     "erp_stock_movements",
		DomainName: "stock",
		PKColumn:   "id_stkmovimiento",
		Columns:    "id_stkmovimiento, stkarticulo_id, stkdeposito_id, stkconcepto_id, fecha_movimiento, cantidad, precio_costo, referencia, subsistema_id",
	}
}

// BOMReader, BOMHistoryReader, StockLevelReader, PriceListReader,
// PriceListItemReader are in stock_extended.go.
