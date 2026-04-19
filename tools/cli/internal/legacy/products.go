package legacy

import "database/sql"

// ---------------------------------------------------------------------------
// Products domain (PRODUCTO_*)
// ---------------------------------------------------------------------------
// PRODUCTOS is the master catalog of "productos terminados" (bus/unit
// models). It's distinct from STK_ARTICULOS (stock articles): the same
// short code (e.g. "434") appears in both — a bus model IS an article in
// stock AND a product in this catalog. The existing metadata_enrichment
// pipeline joins through PRODUCTOS.descripcion_producto → STK_ARTICULOS
// to attach attributes as JSONB on erp_articles; this migration now ALSO
// materializes the full relational shape so the UI forms
// (producto/producto_atributos_valores.xml etc.) can read their native
// schema post-cutover.
//
// Pareto #6 (PRODUCTO_ATRIB_VALORES, 353,936 rows) + Pareto #18
// (PRODUCTO_ATRIBUTO_HOMOLOGACION, 47,189 rows) both ride this migrator
// cluster along with the 4 parent tables needed for FK resolution.

// ProductSectionReader — PRODUCTO_SECCION (10 rows).
func ProductSectionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PRODUCTO_SECCION",
		Target:     "erp_product_sections",
		DomainName: "productos",
		PKColumn:   "id_prdseccion",
		Columns:    "id_prdseccion, nombre_seccion, orden_seccion, rubro_id, activa_seccion",
	}
}

// ProductReader — PRODUCTOS (4,108 rows). descripcion_producto doubles as
// the short article code used in STK_ARTICULOS for units that are also
// stock items (bus chassis / completed models).
func ProductReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PRODUCTOS",
		Target:     "erp_products",
		DomainName: "productos",
		PKColumn:   "id_producto",
		Columns:    "id_producto, descripcion_producto, regcuenta_id",
	}
}

// ProductAttributeReader — PRODUCTO_ATRIBUTOS (415 rows). Attribute
// definitions per section (check/varchar/file/combo/etc).
func ProductAttributeReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PRODUCTO_ATRIBUTOS",
		Target:     "erp_product_attributes",
		DomainName: "productos",
		PKColumn:   "id_prdatributo",
		Columns: "id_prdatributo, nombre_atributo, tipo_atributo, prdseccion_id, " +
			"stkarticulo_id, helper_xml, helper_dir, parametros, orden_atributo, " +
			"activo, print_label, print_value, activo_cotizacion, activo_fichatecnica, " +
			"descrip_cotizacion, definir_antes_seccion_id, estandar_adicional, " +
			"codigo, print_seccion_id",
	}
}

// ProductAttributeOptionReader — PRODUCTO_ATRIB_OPCIONES (147 rows).
// Enumerated option lists for enum-typed attributes.
func ProductAttributeOptionReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PRODUCTO_ATRIB_OPCIONES",
		Target:     "erp_product_attribute_options",
		DomainName: "productos",
		PKColumn:   "id_atribopcion",
		Columns:    "id_atribopcion, prdatributo_id, nombre_opcion, valor_opcion",
	}
}

// ProductAttributeValueReader — PRODUCTO_ATRIB_VALORES (353,936 rows live,
// scrape estimate was 153,835 — +130 % growth). The Pareto #6 target: one
// row per (product, attribute) snapshot. 89 % (315,762) resolve producto_id
// against PRODUCTOS; 11 % (38,174) migrate with product_id NULL preserving
// the raw producto_legacy_id. 100 % resolve prdatributo_id against
// PRODUCTO_ATRIBUTOS (zero orphan).
func ProductAttributeValueReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PRODUCTO_ATRIB_VALORES",
		Target:     "erp_product_attribute_values",
		DomainName: "productos",
		PKColumn:   "id_atribvalor",
		Columns: "id_atribvalor, producto_id, prdatributo_id, valor_atributo, " +
			"cantidad_atributo, timestamp_atributo, cotizacion_id",
	}
}

// ProductAttributeHomologationReader — PRODUCTO_ATRIBUTO_HOMOLOGACION
// (47,189 rows — rank 18 in the original Pareto). Join table between
// product attributes and vehicle homologations (erp_homologations from
// 2.0.8 / HOMOLOGMOD).
func ProductAttributeHomologationReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "PRODUCTO_ATRIBUTO_HOMOLOGACION",
		Target:     "erp_product_attribute_homologations",
		DomainName: "productos",
		PKColumn:   "id_atrib_homolog",
		Columns:    "id_atrib_homolog, prdatributo_id, homologacion_id",
	}
}
