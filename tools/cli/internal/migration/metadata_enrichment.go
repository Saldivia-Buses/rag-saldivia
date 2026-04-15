package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MetadataEnricher reads secondary legacy tables and UPDATEs existing SDA records
// to add JSONB metadata arrays. Phase 17 of the Histrix→SDA migration.
type MetadataEnricher struct {
	mysql    *sql.DB
	pg       *pgxpool.Pool
	tenantID string
	mapper   *Mapper
}

// NewMetadataEnricher creates a new enricher.
func NewMetadataEnricher(mysql *sql.DB, pg *pgxpool.Pool, tenantID string, mapper *Mapper) *MetadataEnricher {
	return &MetadataEnricher{
		mysql:    mysql,
		pg:       pg,
		tenantID: tenantID,
		mapper:   mapper,
	}
}

// enrichSpec defines how to read a legacy table and merge its rows as a JSONB array
// into a parent SDA record's metadata column.
type enrichSpec struct {
	name        string // human-readable name for logging
	query       string // MySQL query (must return all needed columns)
	sdaTable    string // target PostgreSQL table with metadata JSONB column
	domain      string // mapper domain for FK resolution
	legacyTable string // mapper legacy_table for FK resolution
	fkColumn    string // MySQL column containing the parent FK value
	metadataKey string // JSON key to merge (e.g., "product_attributes")
	fkIsCode    bool   // true if FK is a varchar article code needing hashCode resolution
	// transform converts a MySQL row to a compact JSON-friendly map.
	// Only include essential columns to keep metadata small.
	transform func(row map[string]any) map[string]any
}

// RunMetadataEnrichment executes all Phase 17 enrichments in sequence.
func RunMetadataEnrichment(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	e := NewMetadataEnricher(mysqlDB, pgPool, tenantID, mapper)

	specs := e.allSpecs()
	total := len(specs)
	for i, spec := range specs {
		slog.Info("metadata enrichment", "step", fmt.Sprintf("%d/%d", i+1, total), "name", spec.name, "sda_table", spec.sdaTable)
		start := time.Now()
		updated, skipped, err := e.enrich(ctx, spec)
		if err != nil {
			return fmt.Errorf("enrich %s: %w", spec.name, err)
		}
		slog.Info("enrichment done", "name", spec.name, "updated", updated, "skipped", skipped, "duration", time.Since(start).Round(time.Millisecond))
	}
	return nil
}

// enrich reads all rows from the MySQL query, groups by parent FK, resolves UUIDs,
// and batch-UPDATEs the parent records' metadata JSONB.
func (e *MetadataEnricher) enrich(ctx context.Context, spec enrichSpec) (updated, skipped int, err error) {
	rows, err := e.mysql.QueryContext(ctx, spec.query)
	if err != nil {
		slog.Warn("enrichment query failed (table may not exist)", "name", spec.name, "err", err)
		return 0, 0, nil // non-fatal: table might not exist in this legacy instance
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return 0, 0, fmt.Errorf("get columns: %w", err)
	}

	// Read all rows and group by parent FK
	groups := make(map[string]*enrichGroup) // fkValue → group

	for rows.Next() {
		// Scan into generic map
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return 0, 0, fmt.Errorf("scan row: %w", err)
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalizeValue(values[i])
		}

		// Extract FK value
		fkRaw := row[spec.fkColumn]
		fkStr := fmt.Sprintf("%v", fkRaw)
		if fkStr == "" || fkStr == "0" || fkStr == "<nil>" {
			skipped++
			continue
		}

		// Transform row to compact form
		item := spec.transform(row)
		if item == nil {
			skipped++
			continue
		}

		if g, ok := groups[fkStr]; ok {
			g.items = append(g.items, item)
		} else {
			groups[fkStr] = &enrichGroup{items: []map[string]any{item}}
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("iterate rows: %w", err)
	}

	// Resolve parent UUIDs
	for fkStr, g := range groups {
		var parentID uuid.UUID
		var resolveErr error

		if spec.fkIsCode {
			// Article FK via hashCode
			legacyID := int64(hashCode(fkStr))
			parentID, resolveErr = e.mapper.Resolve(ctx, spec.domain, spec.legacyTable, legacyID)
		} else {
			legacyID := parseInt64(fkStr)
			if legacyID == 0 {
				skipped += len(g.items)
				delete(groups, fkStr)
				continue
			}
			parentID, resolveErr = e.mapper.ResolveOptional(ctx, spec.domain, spec.legacyTable, legacyID)
		}

		if resolveErr != nil || parentID == uuid.Nil {
			skipped += len(g.items)
			delete(groups, fkStr)
			continue
		}
		g.parentUUID = parentID
	}

	// Batch UPDATE in chunks of 100
	const batchSize = 100
	batch := make([]*enrichGroup, 0, batchSize)
	for _, g := range groups {
		batch = append(batch, g)
		if len(batch) >= batchSize {
			n, err := e.flushBatch(ctx, spec.sdaTable, spec.metadataKey, batch)
			if err != nil {
				return updated, skipped, err
			}
			updated += n
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		n, err := e.flushBatch(ctx, spec.sdaTable, spec.metadataKey, batch)
		if err != nil {
			return updated, skipped, err
		}
		updated += n
	}

	return updated, skipped, nil
}

// enrichGroup holds a resolved parent UUID and its associated metadata items.
type enrichGroup struct {
	parentUUID uuid.UUID
	items      []map[string]any
}

// flushBatch executes batched UPDATEs for metadata enrichment.
// Each UPDATE merges a JSON key into the metadata column using || operator.
func (e *MetadataEnricher) flushBatch(ctx context.Context, sdaTable, metadataKey string, groups []*enrichGroup) (int, error) {
	if len(groups) == 0 {
		return 0, nil
	}

	tx, err := e.pg.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Disable triggers for bulk update
	if _, err := tx.Exec(ctx, "SET LOCAL session_replication_role = 'replica'"); err != nil {
		slog.Warn("could not disable triggers for enrichment", "err", err)
	}

	updated := 0
	for _, g := range groups {
		jsonData, err := json.Marshal(map[string]any{metadataKey: g.items})
		if err != nil {
			return 0, fmt.Errorf("marshal metadata: %w", err)
		}

		tag, err := tx.Exec(ctx,
			fmt.Sprintf(`UPDATE %s SET metadata = COALESCE(metadata, '{}'::jsonb) || $1::jsonb WHERE id = $2 AND tenant_id = $3`, sdaTable),
			string(jsonData), g.parentUUID, e.tenantID,
		)
		if err != nil {
			return 0, fmt.Errorf("update %s metadata: %w", sdaTable, err)
		}
		if tag.RowsAffected() > 0 {
			updated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit enrichment batch: %w", err)
	}
	return updated, nil
}

// allSpecs returns all enrichment specifications for Phase 17.
func (e *MetadataEnricher) allSpecs() []enrichSpec {
	return []enrichSpec{
		// === erp_articles metadata ===
		e.articleProductAttributes(),
		e.articleLocations(),
		e.articleCostHistory(),
		e.articleReplacementCosts(),
		e.articleSuppliers(),
		e.articleCostInspections(),

		// === erp_entities metadata ===
		e.entityQualifications(),

		// === erp_units metadata ===
		e.unitAccessories(),
		e.unitVehicleHistory(),
		e.unitLCMCertificates(),
		e.unitManufacturingCerts(),
		e.unitDeliveries(),

		// === erp_employee_details metadata ===
		e.employeeIssuedTools(),
		e.employeeCategoryHistory(),
		e.employeeMultiskillMatrix(),
	}
}

// --- erp_articles enrichments ---

func (e *MetadataEnricher) articleProductAttributes() enrichSpec {
	// PRODUCTO_ATRIB_VALORES → PRODUCTOS (via producto_id) → STK_ARTICULOS (via descripcion_producto = artcod).
	// We join through PRODUCTOS to resolve the article code, then use hashCode for FK resolution.
	return enrichSpec{
		name:        "article_product_attributes",
		query:       "SELECT pav.id_atribvalor, p.descripcion_producto AS artcod, pav.prdatributo_id, pav.valor_atributo, pav.cantidad_atributo FROM PRODUCTO_ATRIB_VALORES pav JOIN PRODUCTOS p ON p.id_producto = pav.producto_id WHERE pav.producto_id IS NOT NULL ORDER BY p.descripcion_producto",
		sdaTable:    "erp_articles",
		domain:      "stock",
		legacyTable: "STK_ARTICULOS",
		fkColumn:    "artcod",
		metadataKey: "product_attributes",
		fkIsCode:    true,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"attribute_id": row["prdatributo_id"],
				"value":        truncStr(row["valor_atributo"], 100),
				"quantity":     row["cantidad_atributo"],
			}
		},
	}
}

func (e *MetadataEnricher) articleLocations() enrichSpec {
	return enrichSpec{
		name:        "article_locations",
		query:       "SELECT id_articulo_locacion, stkarticulo_id, stklocacion_id, stock_minimo, stock_maximo, stock_actual FROM STK_ARTICULO_LOCACION ORDER BY stkarticulo_id",
		sdaTable:    "erp_articles",
		domain:      "stock",
		legacyTable: "STK_ARTICULOS",
		fkColumn:    "stkarticulo_id",
		metadataKey: "locations",
		fkIsCode:    true,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"location_id": row["stklocacion_id"],
				"stock_min":   row["stock_minimo"],
				"stock_max":   row["stock_maximo"],
				"stock":       row["stock_actual"],
			}
		},
	}
}

func (e *MetadataEnricher) articleCostHistory() enrichSpec {
	// Combines STK_COSTOS (cost movements, 15K rows, FK: stkarticulo_id varchar) +
	// STK_COSTO_HIST (historical costs, 103K rows, FK: articulo_id varchar) via UNION ALL.
	return enrichSpec{
		name: "article_cost_history",
		query: `SELECT stkarticulo_id AS art_id, precio_costo AS cost, fecha_movimiento AS date, 'current' AS source FROM STK_COSTOS
		         UNION ALL
		         SELECT articulo_id AS art_id, costo_hist AS cost, CONCAT(anio_hist, '-', LPAD(mes_hist, 2, '0'), '-01') AS date, 'history' AS source FROM STK_COSTO_HIST
		         ORDER BY art_id`,
		sdaTable:    "erp_articles",
		domain:      "stock",
		legacyTable: "STK_ARTICULOS",
		fkColumn:    "art_id",
		metadataKey: "cost_history",
		fkIsCode:    true,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"cost":   row["cost"],
				"date":   formatDateVal(row["date"]),
				"source": row["source"],
			}
		},
	}
}

func (e *MetadataEnricher) articleReplacementCosts() enrichSpec {
	// STK_COSTO_REPOSICION (current) + STK_COSTO_REPOSICION_HIST (historical).
	// Both link to article via stkarticulo_id (varchar).
	return enrichSpec{
		name: "article_replacement_costs",
		query: `SELECT stkarticulo_id AS art_id, costo_proveedor, costo_final, moneda_id, origen, ultimo_cambio AS date, 'current' AS source FROM STK_COSTO_REPOSICION
		         UNION ALL
		         SELECT r.stkarticulo_id AS art_id, h.costo_proveedor, 0 AS costo_final, h.moneda_id, h.origen, h.modificado AS date, 'history' AS source
		         FROM STK_COSTO_REPOSICION_HIST h JOIN STK_COSTO_REPOSICION r ON r.id_costoreposicion = h.costoreposicion_id
		         ORDER BY art_id`,
		sdaTable:    "erp_articles",
		domain:      "stock",
		legacyTable: "STK_ARTICULOS",
		fkColumn:    "art_id",
		metadataKey: "replacement_costs",
		fkIsCode:    true,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"supplier_cost": row["costo_proveedor"],
				"final_cost":    row["costo_final"],
				"currency_id":   row["moneda_id"],
				"origin":        row["origen"],
				"date":          formatDateVal(row["date"]),
				"source":        row["source"],
			}
		},
	}
}

func (e *MetadataEnricher) articleSuppliers() enrichSpec {
	return enrichSpec{
		name:        "article_suppliers",
		query:       "SELECT id_stkarticuloproveedor, stkarticulo_id, regcuenta_id, prv_articulo FROM STK_ARTICULOS_PROVEEDOR ORDER BY stkarticulo_id",
		sdaTable:    "erp_articles",
		domain:      "stock",
		legacyTable: "STK_ARTICULOS",
		fkColumn:    "stkarticulo_id",
		metadataKey: "supplier_articles",
		fkIsCode:    true,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"supplier_id":      row["regcuenta_id"],
				"supplier_article": truncStr(row["prv_articulo"], 60),
			}
		},
	}
}

func (e *MetadataEnricher) articleCostInspections() enrichSpec {
	// STKINSPR — article cost tracking entries, FK via artcod (varchar).
	return enrichSpec{
		name:        "article_cost_inspections",
		query:       "SELECT idCosto, artcod, artcos, ctacod, fecfac, fecult, siscod, movnro FROM STKINSPR ORDER BY artcod",
		sdaTable:    "erp_articles",
		domain:      "stock",
		legacyTable: "STK_ARTICULOS",
		fkColumn:    "artcod",
		metadataKey: "cost_inspections",
		fkIsCode:    true,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"cost":       row["artcos"],
				"entity_id":  row["ctacod"],
				"invoice_dt": formatDateVal(row["fecfac"]),
				"last_dt":    formatDateVal(row["fecult"]),
				"subsystem":  row["siscod"],
			}
		},
	}
}

// --- erp_entities enrichments ---

func (e *MetadataEnricher) entityQualifications() enrichSpec {
	return enrichSpec{
		name:        "entity_qualifications",
		query:       "SELECT id_regcalificacion, regcuenta_id, calificacion, fecha_calificacion, referencia_calificacion FROM REG_CUENTA_CALIFICACION ORDER BY regcuenta_id",
		sdaTable:    "erp_entities",
		domain:      "entity",
		legacyTable: "REG_CUENTA",
		fkColumn:    "regcuenta_id",
		metadataKey: "qualifications",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"rating":    truncStr(row["calificacion"], 40),
				"date":      formatDateVal(row["fecha_calificacion"]),
				"reference": truncStr(row["referencia_calificacion"], 100),
			}
		},
	}
}

// --- erp_units enrichments ---

func (e *MetadataEnricher) unitAccessories() enrichSpec {
	// ACCESORIOS_COCHE → erp_units via nrofab (int FK to CHASIS.nrocha).
	return enrichSpec{
		name:        "unit_accessories",
		query:       "SELECT id_accesorio, nrofab, artcod, artdes, fecha, estado, cantidad, precio_unitario, precio_adicional FROM ACCESORIOS_COCHE ORDER BY nrofab",
		sdaTable:    "erp_units",
		domain:      "production",
		legacyTable: "CHASIS",
		fkColumn:    "nrofab",
		metadataKey: "accessories",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"code":        truncStr(row["artcod"], 10),
				"description": truncStr(row["artdes"], 60),
				"date":        formatDateVal(row["fecha"]),
				"status":      row["estado"],
				"quantity":    row["cantidad"],
				"unit_price":  row["precio_unitario"],
			}
		},
	}
}

func (e *MetadataEnricher) unitVehicleHistory() enrichSpec {
	// CARCHEHI — linked to vehicles via carint = CHASIS.nrocha.
	// Only rows where carint has a matching CHASIS record will be enriched.
	return enrichSpec{
		name:        "unit_vehicle_history",
		query:       "SELECT carint, siscod, succod, cardes, carimp, carfec, carnro, carbco, cartip FROM CARCHEHI ORDER BY carint",
		sdaTable:    "erp_units",
		domain:      "production",
		legacyTable: "CHASIS",
		fkColumn:    "carint",
		metadataKey: "vehicle_history",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"subsystem":   row["siscod"],
				"branch":      row["succod"],
				"description": truncStr(row["cardes"], 35),
				"amount":      row["carimp"],
				"date":        formatDateVal(row["carfec"]),
				"number":      row["carnro"],
				"type":        row["cartip"],
			}
		},
	}
}

func (e *MetadataEnricher) unitLCMCertificates() enrichSpec {
	// CERTIFICADOLCM has composite PK (nrofab, fecha). FK via nrofab → CHASIS.nrocha.
	return enrichSpec{
		name:        "unit_lcm_certificates",
		query:       "SELECT nrofab, fecha, nrocert, observ FROM CERTIFICADOLCM ORDER BY nrofab",
		sdaTable:    "erp_units",
		domain:      "production",
		legacyTable: "CHASIS",
		fkColumn:    "nrofab",
		metadataKey: "lcm_certificates",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"date":         formatDateVal(row["fecha"]),
				"cert_number":  row["nrocert"],
				"observations": truncStr(row["observ"], 100),
			}
		},
	}
}

func (e *MetadataEnricher) unitManufacturingCerts() enrichSpec {
	// CERTFABRICACION has PK nrofab. FK via nrofab → CHASIS.nrocha.
	return enrichSpec{
		name:        "unit_manufacturing_certs",
		query:       "SELECT nrofab, fechaCert, fechaFac, nrovin, lcm, expte, plano, disposicion, tara, bruto, certestado_id FROM CERTFABRICACION ORDER BY nrofab",
		sdaTable:    "erp_units",
		domain:      "production",
		legacyTable: "CHASIS",
		fkColumn:    "nrofab",
		metadataKey: "manufacturing_certs",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"cert_date":   formatDateVal(row["fechaCert"]),
				"invoice_dt":  formatDateVal(row["fechaFac"]),
				"vin":         truncStr(row["nrovin"], 50),
				"lcm":         truncStr(row["lcm"], 30),
				"file_number": truncStr(row["expte"], 20),
				"tare":        row["tara"],
				"gross":       row["bruto"],
				"status":      row["certestado_id"],
			}
		},
	}
}

func (e *MetadataEnricher) unitDeliveries() enrichSpec {
	// ENTREGASCOCHES has PK id_entregacoche. FK via nrofab → CHASIS.nrocha.
	return enrichSpec{
		name:        "unit_deliveries",
		query:       "SELECT id_entregacoche, nrofab, entrega, nombre, domicilio, dni, recibepor, hora, entrega_estimada FROM ENTREGASCOCHES ORDER BY nrofab",
		sdaTable:    "erp_units",
		domain:      "production",
		legacyTable: "CHASIS",
		fkColumn:    "nrofab",
		metadataKey: "deliveries",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"date":           formatDateVal(row["entrega"]),
				"recipient":      truncStr(row["nombre"], 100),
				"address":        truncStr(row["domicilio"], 100),
				"dni":            truncStr(row["dni"], 10),
				"received_by":    truncStr(row["recibepor"], 100),
				"estimated_date": formatDateVal(row["entrega_estimada"]),
			}
		},
	}
}

// --- erp_employee_details enrichments ---

func (e *MetadataEnricher) employeeIssuedTools() enrichSpec {
	// PERSONAL_ARTICULOS → erp_employee_details via IdPersona → PERSONAL.
	return enrichSpec{
		name:        "employee_issued_tools",
		query:       "SELECT id_personalart, IdPersona, idPuesto, artcod, dias_recambio, observacion FROM PERSONAL_ARTICULOS ORDER BY IdPersona",
		sdaTable:    "erp_employee_details",
		domain:      "entity",
		legacyTable: "PERSONAL",
		fkColumn:    "IdPersona",
		metadataKey: "issued_tools",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"article_code":  truncStr(row["artcod"], 20),
				"position_id":   row["idPuesto"],
				"replacement_d": row["dias_recambio"],
				"notes":         truncStr(row["observacion"], 100),
			}
		},
	}
}

func (e *MetadataEnricher) employeeCategoryHistory() enrichSpec {
	// CATEGHISTORIA → erp_employee_details via IdPersona → PERSONAL.
	return enrichSpec{
		name:        "employee_category_history",
		query:       "SELECT idHist, IdPersona, idSindicato, IdCateg, fechaDesde FROM CATEGHISTORIA ORDER BY IdPersona",
		sdaTable:    "erp_employee_details",
		domain:      "entity",
		legacyTable: "PERSONAL",
		fkColumn:    "IdPersona",
		metadataKey: "category_history",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"union_id":    row["idSindicato"],
				"category_id": row["IdCateg"],
				"from_date":   formatDateVal(row["fechaDesde"]),
			}
		},
	}
}

func (e *MetadataEnricher) employeeMultiskillMatrix() enrichSpec {
	// RH_POLIVALENCIAS → erp_employee_details via IdPersona → PERSONAL.
	return enrichSpec{
		name:        "employee_multiskill_matrix",
		query:       "SELECT id_rhpolivalencia, IdPersona, seccion_id, polivalencianivel_id FROM RH_POLIVALENCIAS ORDER BY IdPersona",
		sdaTable:    "erp_employee_details",
		domain:      "entity",
		legacyTable: "PERSONAL",
		fkColumn:    "IdPersona",
		metadataKey: "multiskill_matrix",
		fkIsCode:    false,
		transform: func(row map[string]any) map[string]any {
			return map[string]any{
				"section_id": row["seccion_id"],
				"level_id":   row["polivalencianivel_id"],
			}
		},
	}
}

// --- Helper functions ---

// normalizeValue converts database driver types to JSON-friendly Go types.
func normalizeValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		if val.IsZero() || val.Year() < 1900 {
			return nil
		}
		return val.Format("2006-01-02")
	default:
		return val
	}
}

// formatDateVal formats a date value (string or time.Time) as YYYY-MM-DD.
func formatDateVal(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		if val.IsZero() || val.Year() < 1900 {
			return nil
		}
		return val.Format("2006-01-02")
	case string:
		if val == "" || val == "0000-00-00" || strings.HasPrefix(val, "0000") {
			return nil
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// truncStr truncates a string value to maxLen characters.
func truncStr(v any, maxLen int) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("%v", v)
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// parseInt64 parses a string as int64, returning 0 on failure.
func parseInt64(s string) int64 {
	var n int64
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}
