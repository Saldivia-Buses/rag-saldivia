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
// Includes a subquery to get the first article from MRP_ORDEN_DETALLE.
func ProductionOrderReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "MRP_ORDEN_PRODUCCION",
		Target:     "erp_production_orders",
		DomainName: "production",
		PKColumn:   "id_mrporden",
		Columns:    "id_mrporden, fecha_orden, fecha_cierre, mrpestado_id, centro_productivo_id, orden_comentarios, login_id, (SELECT stkarticulo_id FROM MRP_ORDEN_DETALLE WHERE mrporden_id = MRP_ORDEN_PRODUCCION.id_mrporden LIMIT 1) as first_article_code, (SELECT objetivo_ordendetalle FROM MRP_ORDEN_DETALLE WHERE mrporden_id = MRP_ORDEN_PRODUCCION.id_mrporden LIMIT 1) as first_quantity",
	}
}
