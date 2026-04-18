-- ─── Products domain (Phase 1 §Data migration — Pareto #6 + #18) ───────
-- PRODUCTOS / PRODUCTO_ATRIBUTOS / PRODUCTO_ATRIB_VALORES and related
-- lookup tables. The UI forms at producto/producto_atributos_valores.xml
-- etc. drive their views off these tables; each has a minimal sqlc read
-- query here to satisfy the Phase 0 "every migrated table has a sqlc
-- query" gate.

-- name: ListProductSections :many
SELECT id, tenant_id, legacy_id, name, sort_order, rubro_id, active, created_at
FROM erp_product_sections
WHERE tenant_id = $1
ORDER BY sort_order, name;

-- name: ListProducts :many
SELECT id, tenant_id, legacy_id, description, supplier_entity_id,
       supplier_code, created_at
FROM erp_products
WHERE tenant_id = $1
ORDER BY description
LIMIT $2 OFFSET $3;

-- name: ListProductAttributes :many
SELECT id, tenant_id, legacy_id, name, attribute_type, section_id,
       section_legacy_id, article_code, helper_xml, helper_dir, parameters,
       sort_order, active, print_label, print_value,
       active_in_quote, active_in_tech_sheet, quote_description,
       define_before_section_id, standard_additional,
       code, print_section_id, created_at
FROM erp_product_attributes
WHERE tenant_id = $1
  AND (sqlc.arg(active_only)::BOOLEAN = false OR active = true)
ORDER BY sort_order, name;

-- name: ListProductAttributeOptions :many
SELECT id, tenant_id, legacy_id, attribute_id, attribute_legacy_id,
       option_name, option_value, created_at
FROM erp_product_attribute_options
WHERE tenant_id = $1 AND attribute_id = $2
ORDER BY option_value;

-- name: ListProductAttributeValues :many
-- Values for a single product, ordered by attribute. The UI replaces
-- producto/producto_atributos_valores.xml with this.
SELECT id, tenant_id, legacy_id, product_id, product_legacy_id,
       attribute_id, attribute_legacy_id, value, quantity,
       quote_legacy_id, recorded_at, created_at
FROM erp_product_attribute_values
WHERE tenant_id = $1 AND product_id = $2
ORDER BY attribute_legacy_id;

-- name: ListProductAttributeHomologations :many
-- Cross-reference between attribute definitions and vehicle homologations
-- (erp_homologations). Backs the producto/producto_atributos_homologacion_qry
-- screen post-cutover.
SELECT id, tenant_id, legacy_id, attribute_id, attribute_legacy_id,
       homologation_id, homologation_legacy_id, created_at
FROM erp_product_attribute_homologations
WHERE tenant_id = $1 AND attribute_id = $2
ORDER BY homologation_legacy_id;
