package legacy

import "database/sql"

// ProductionCenterReader creates a reader for MRP_CENTRO_PRODUCTIVO.
func ProductionCenterReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MRP_CENTRO_PRODUCTIVO",
		Target:     "erp_production_centers",
		DomainName: "production",
		PKColumn:   "id_centro_productivo",
		Columns:    "id_centro_productivo, nombre_centroproductivo",
	}
}

// ProductionOrderReader creates a reader for MRP_ORDEN_PRODUCCION.
func ProductionOrderReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MRP_ORDEN_PRODUCCION",
		Target:     "erp_production_orders",
		DomainName: "production",
		PKColumn:   "id_mrporden",
		Columns:    "id_mrporden, fecha_orden, fecha_cierre, mrpestado_id, centro_productivo_id, orden_comentarios, login_id",
	}
}
