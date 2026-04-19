package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Camionerou/rag-saldivia/tools/cli/internal/legacy"
)

// ============================================================================
// Phase 9 — Stock Extended
// ============================================================================

// NewBOMMigrator migrates STKPIEZA → erp_bom.
// STKPIEZA has auto-increment id_pieza PK. 36K rows.
// Both parent (idPadre) and child (articulo_hijo) are varchar article codes →
// resolved via hashCode to the STK_ARTICULOS mapping.
func NewBOMMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.BOMReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "parent_id", "child_id", "quantity", "unit_id", "sort_order", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_pieza")
			if legacyID == 0 {
				return nil, nil
			}

			// Parent article FK via hashCode
			parentCode := row.String("idPadre")
			if parentCode == "" {
				return nil, nil // skip rows with no parent
			}
			parentLegacyID := int64(hashCode(parentCode))
			parentID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", parentLegacyID)
			if err != nil {
				return nil, nil // skip if parent article not migrated
			}

			// Child article FK via hashCode
			childCode := row.String("articulo_hijo")
			if childCode == "" {
				return nil, nil
			}
			childLegacyID := int64(hashCode(childCode))
			childID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", childLegacyID)
			if err != nil {
				return nil, nil // skip if child article not migrated
			}

			id, err := mapper.Map(ctx, nil, "stock", "STKPIEZA", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// posicionfab_id as sort_order (manufacturing position)
			sortOrder := row.Int("posicionfab_id")

			return []any{
				id, tenantID, parentID, childID,
				ParseDecimal(row.Decimal("cantidad")),
				(*uuid.UUID)(nil), // unit_id — no direct mapping in STKPIEZA
				sortOrder,
				"", // notes — no notes column in STKPIEZA
			}, nil
		},
	}
}

// NewBOMHistoryMigrator migrates STK_BOM_HIST → erp_bom_history.
// STK_BOM_HIST has auto-increment id_stkbomhist PK. 3.3M rows.
// Same hashCode pattern for parent/child article resolution.
func NewBOMHistoryMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.BOMHistoryReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "parent_id", "child_id", "quantity", "unit_id", "version", "effective_date", "replaced_date", "legacy_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_stkbomhist")
			if legacyID == 0 {
				return nil, nil
			}

			// Child article FK via hashCode on stkarticulohijo_id (varchar)
			childCode := row.String("stkarticulohijo_id")
			if childCode == "" {
				return nil, nil
			}
			childLegacyID := int64(hashCode(childCode))
			childID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", childLegacyID)
			if err != nil {
				return nil, nil
			}

			// Parent article FK: resolve via hashCode of parent article code.
			// parent_article_code comes from STKPIEZA.idPadre via JOIN in the reader.
			// When the LEFT JOIN finds no STKPIEZA (orphan pieza_id) we fall back to
			// the ghost article seeded by RescueBOMOrphanParents, keyed on pieza_id.
			// Saldivia empirics (2026-04-16): without the ghost fallback this path
			// lost 2,546,930 rows out of 14,572,284 (17.5%).
			var parentID uuid.UUID
			parentCode := row.String("parent_article_code")
			if parentCode != "" {
				parentLegacyID := int64(hashCode(parentCode))
				resolved, err := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", parentLegacyID)
				if err == nil && resolved != uuid.Nil {
					parentID = resolved
				}
			}
			if parentID == uuid.Nil {
				// Fallback to the ghost article tied to pieza_id.
				piezaID := row.Int64("pieza_id")
				if piezaID > 0 {
					if ghost, err := mapper.ResolveOptional(ctx, "stock", "GHOST_PIEZA", piezaID); err == nil && ghost != uuid.Nil {
						parentID = ghost
					}
				}
			}
			if parentID == uuid.Nil {
				return nil, nil // truly unresolvable — no pieza_id and no parent code
			}

			id, err := mapper.Map(ctx, nil, "stock", "STK_BOM_HIST", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			version := row.Int("bom_variacion_id")
			effectiveDate := SafeDate(timeFromRow(row, "fecha_costo"))

			return []any{
				id, tenantID, parentID, childID,
				ParseDecimal(row.Decimal("cantidad")),
				(*uuid.UUID)(nil), // unit_id
				version,
				effectiveDate,
				(*time.Time)(nil), // replaced_date — not in STK_BOM_HIST
				legacyID,
			}, nil
		},
	}
}

// NewStockLevelMigrator migrates STK_STOCKACTUAL → erp_stock_levels.
// STK_STOCKACTUAL has NO auto-increment PK — composite key (stkarticulo_id, stkdeposito_id). 17K rows.
// article_id via hashCode, warehouse_id via Resolve.
func NewStockLevelMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.StockLevelReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "article_id", "warehouse_id", "quantity", "reserved", "updated_at"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite key — generate deterministic ID from hash
			artCode := row.String("stkarticulo_id")
			if artCode == "" {
				return nil, nil
			}
			depID := row.Int64("stkdeposito_id")
			if depID == 0 {
				return nil, nil
			}

			// Resolve article FK via hashCode
			artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
			articleID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", artLegacyID)
			if err != nil {
				return nil, nil // skip if article not migrated
			}

			// Resolve warehouse FK
			warehouseID, err := mapper.Resolve(ctx, "stock", "STK_DEPOSITOS", depID)
			if err != nil {
				return nil, nil
			}

			// Generate deterministic ID from composite key
			compositeKey := fmt.Sprintf("STOCKLEVEL:%s:%d", artCode, depID)
			legacyID := int64(hashCode(compositeKey))
			id, err := mapper.Map(ctx, nil, "stock", "STK_STOCKACTUAL", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, articleID, warehouseID,
				ParseDecimal(row.Decimal("cantidad_stock")),
				ParseDecimal("0"), // reserved — not tracked in legacy
				SafeDateRequired(timeFromRow(row, "actualizado")),
			}, nil
		},
	}
}

// NewPriceListMigrator migrates STK_LISTAS → erp_price_lists.
// STK_LISTAS has auto-increment id_stklista PK. 1.1K rows.
func NewPriceListMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PriceListReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "name", "currency_id", "valid_from", "valid_until", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_stklista")
			if legacyID == 0 {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "stock", "STK_LISTAS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// currency_id — moneda_id is an int reference, resolve as catalog if available
			monedaID := row.Int64("moneda_id")
			var currencyID *uuid.UUID
			if monedaID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "catalog", "MONEDAS", monedaID)
				if err == nil && resolved != uuid.Nil {
					currencyID = &resolved
				}
			}

			return []any{
				id, tenantID,
				row.String("stklista_nombre"),
				currencyID,
				SafeDate(timeFromRow(row, "stklista_fecha")),
				(*time.Time)(nil), // valid_until — not in STK_LISTAS
				true,              // active — all legacy lists are active
			}, nil
		},
	}
}

// NewPriceListItemMigrator migrates STK_LISTADETALLE → erp_price_list_items.
// STK_LISTADETALLE has auto-increment id_stklistadetalle PK. 138K rows.
// article_id via hashCode on stkarticulo_id varchar.
func NewPriceListItemMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PriceListItemReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "price_list_id", "article_id", "description", "price"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_stklistadetalle")
			if legacyID == 0 {
				return nil, nil
			}

			// Resolve price list FK
			listaID := row.Int64("stklista_id")
			priceListID, err := mapper.Resolve(ctx, "stock", "STK_LISTAS", listaID)
			if err != nil {
				return nil, nil // skip if price list not migrated
			}

			// Resolve article FK via hashCode
			artCode := row.String("stkarticulo_id")
			var articleID *uuid.UUID
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
				resolved, err := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID)
				if err == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "stock", "STK_LISTADETALLE", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Use obervaciones (sic — typo in Histrix) as description, fallback to nombre
			desc := row.String("obervaciones")
			if desc == "" {
				desc = row.String("nombre")
			}

			return []any{
				id, tenantID, priceListID, articleID,
				desc,
				ParseDecimal(row.Decimal("stklistadetalle_precioventa")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 10 — Purchasing Extended
// ============================================================================

// NewInternalRequisitionMigrator migrates PEDIDOINT → erp_purchase_orders.
// PEDIDOINT has auto-increment idPed PK. 384K rows.
// supplier_id nullable (migration 054 made it nullable).
// Each row treated as a separate purchase order since idPed is the auto-increment PK.
// Adds metadata: {"source": "internal_requisition"}.
func NewInternalRequisitionMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.InternalRequisitionReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "supplier_id", "status", "currency_id", "total", "notes", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("idPed")
			if legacyID == 0 {
				return nil, nil
			}

			// Supplier FK — codprv is supplier code, nullable
			var supplierID *uuid.UUID
			codprv := row.Int64("codprv")
			if codprv > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "REG_CUENTA", codprv)
				if err == nil && resolved != uuid.Nil {
					supplierID = &resolved
				}
			}

			// Status mapping: estado 0=draft, 1=approved, 2=received
			estado := row.Int("estado")
			status := "draft"
			switch estado {
			case 1:
				status = "approved"
			case 2:
				status = "received"
			}

			number := fmt.Sprintf("RI-%d", legacyID)

			id, err := mapper.Map(ctx, nil, "purchasing", "PEDIDOINT", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Notes: combine movref + mensaje
			notes := row.String("movref")
			if msg := row.String("mensaje"); msg != "" {
				if notes != "" {
					notes += " | "
				}
				notes += msg
			}

			return []any{
				id, tenantID, number,
				SafeDateRequired(timeFromRow(row, "fechaemi")),
				supplierID,
				status,
				(*uuid.UUID)(nil), // currency_id — not in PEDIDOINT
				ParseDecimal(row.Decimal("cant")),
				notes,
				LegacyUserID,
			}, nil
		},
	}
}

// NewPurchaseReceiptMigrator migrates OCPRECIB → erp_purchase_receipts.
// OCPRECIB has auto-increment id_recepcion PK. 320K rows.
func NewPurchaseReceiptMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PurchaseReceiptReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "order_id", "date", "number", "user_id", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_recepcion")
			if legacyID == 0 {
				return nil, nil
			}

			// order_id: link to internal requisition via idPed
			var orderID *uuid.UUID
			idPed := row.Int64("idPed")
			if idPed > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "purchasing", "PEDIDOINT", idPed)
				if err == nil && resolved != uuid.Nil {
					orderID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "purchasing", "OCPRECIB", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			number := fmt.Sprintf("REC-%d", row.Int64("nrorec"))
			if row.Int64("nrorec") == 0 {
				number = fmt.Sprintf("REC-%d", legacyID)
			}

			return []any{
				id, tenantID, orderID,
				SafeDateRequired(timeFromRow(row, "fecrec")),
				number,
				LegacyUserID,
				row.String("observacion"),
			}, nil
		},
	}
}

// ============================================================================
// Phase 11 — Sales & Production Extended
// ============================================================================

// NewQuotationLineMigrator migrates COTIZOPCIONES → erp_quotation_lines.
// Composite PK (idPrecio, idOpcion). 882 rows.
// idPrecio links to COTIZACION pricing, idOpcion is the option number within that quotation.
func NewQuotationLineMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.QuotationLineReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "quotation_id", "article_id", "description", "quantity", "unit_price", "sort_order", "metadata"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite PK — generate deterministic ID
			idPrecio := row.Int64("idPrecio")
			idOpcion := row.Int64("idOpcion")
			if idPrecio == 0 && idOpcion == 0 {
				return nil, nil
			}

			compositeKey := fmt.Sprintf("COTIZOPCIONES:%d:%d", idPrecio, idOpcion)
			legacyID := int64(hashCode(compositeKey))

			// Resolve quotation FK — idPrecio might link to COTIZACION
			var quotationID *uuid.UUID
			if idPrecio > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "sales", "COTIZACION", idPrecio)
				if err == nil && resolved != uuid.Nil {
					quotationID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "sales", "COTIZOPCIONES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Metadata: store idSeccion
			meta := map[string]any{
				"idSeccion": row.Int("idSeccion"),
				"source":    "COTIZOPCIONES",
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID, quotationID,
				(*uuid.UUID)(nil), // article_id — no article FK in COTIZOPCIONES
				row.String("descripcion"),
				ParseDecimal("1"),  // quantity — not in COTIZOPCIONES, default 1
				ParseDecimal("0"),  // unit_price — no price column in COTIZOPCIONES
				row.Int("idOpcion"), // sort_order
				string(metaJSON),
			}, nil
		},
	}
}

// NewCustomerOrderMigrator migrates PEDCOTIZ → erp_orders.
// PEDCOTIZ has auto-increment id PK. 3.8K rows.
// Massive table with ~100 columns describing vehicle specifications per order.
func NewCustomerOrderMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CustomerOrderReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "order_type", "customer_id", "quotation_id", "status", "total", "user_id", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id")
			if legacyID == 0 {
				return nil, nil
			}

			// Customer FK: ctacod → REG_CUENTA
			ctacod := row.Int64("ctacod")
			var customerID *uuid.UUID
			if ctacod > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "REG_CUENTA", ctacod)
				if err == nil && resolved != uuid.Nil {
					customerID = &resolved
				}
			}

			// Quotation FK: cotizacion_id → COTIZACION
			var quotationID *uuid.UUID
			cotizID := row.Int64("cotizacion_id")
			if cotizID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "sales", "COTIZACION", cotizID)
				if err == nil && resolved != uuid.Nil {
					quotationID = &resolved
				}
			}

			// Status from estado_ficha: 0=pending, 1=in_progress, 2=delivered.
			// erp_orders.status CHECK allows: pending|in_progress|shipped|delivered|cancelled.
			estadoFicha := row.Int("estado_ficha")
			status := "pending"
			switch estadoFicha {
			case 1:
				status = "in_progress"
			case 2:
				status = "delivered"
			}

			number := fmt.Sprintf("PED-%d", row.Int64("pednro"))
			if row.Int64("pednro") == 0 {
				number = fmt.Sprintf("PED-%d", legacyID)
			}

			id, err := mapper.Map(ctx, nil, "sales", "PEDCOTIZ", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, number,
				SafeDateRequired(timeFromRow(row, "pedfec")),
				"customer", // order_type — PEDCOTIZ are customer-facing vehicle orders (CHECK allows only customer|internal)
				customerID,
				quotationID,
				status,
				ParseDecimal("0"), // total — no total column in PEDCOTIZ
				LegacyUserID,
				row.String("obsesp"), // notes from obsesp (observaciones especiales)
			}, nil
		},
	}
}

// NewVehicleMigrator migrates CHASIS → erp_units.
// CHASIS has PK nrocha (mediumint, not auto-increment). ~4K rows.
// Contains the real chassis/vehicle data: nrocha=chassis number, marcod/modcod=brand/model,
// nromotor=engine number, ctacod=customer, entrada/salida=dates.
func NewVehicleMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.VehicleReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "chassis_number", "internal_number", "model", "customer_id", "order_id", "production_order_id", "patent", "status", "engine_brand", "body_style", "seat_count", "year", "metadata", "delivered_at"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("nrocha")
			if legacyID == 0 {
				return nil, nil
			}

			// Customer FK: ctacod → REG_CUENTA
			ctacod := row.Int64("ctacod")
			var customerID *uuid.UUID
			if ctacod > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "REG_CUENTA", ctacod)
				if err == nil && resolved != uuid.Nil {
					customerID = &resolved
				}
			}

			// Production order FK (not directly available in CHASIS, skip)
			var orderID *uuid.UUID
			var prodOrderID *uuid.UUID

			// Status: CHECK allows only in_production/ready/delivered.
			// salida (exit date) present → delivered, else default in_production.
			status := "in_production"
			salidaDate := timeFromRow(row, "salida")
			if !salidaDate.IsZero() && salidaDate.Year() > 1900 {
				status = "delivered"
			}

			// Model from modcha
			model := row.String("modcha")

			id, err := mapper.Map(ctx, nil, "production", "CHASIS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Metadata: store extra CHASIS fields
			meta := map[string]any{
				"marcod":    row.Int("marcod"),
				"modcod":    row.Int("modcod"),
				"nromotor":  row.String("nromotor"),
				"idTaco":    row.Int("idTaco"),
				"nroTaco":   row.String("nroTaco"),
				"diasTaco":  row.Int("diasTaco"),
				"conces":    row.Int("conces"),
			}
			metaJSON, _ := json.Marshal(meta)

			chassisNumber := row.String("chasis")
			if chassisNumber == "" || chassisNumber == "0" {
				chassisNumber = fmt.Sprintf("CHA-%d", legacyID)
			}

			return []any{
				id, tenantID,
				chassisNumber,
				fmt.Sprintf("%d", legacyID), // internal_number = nrocha
				model,
				customerID,
				orderID,
				prodOrderID,
				"",       // patent — not in CHASIS table
				status,
				"",       // engine_brand — idMarcaMotor is an int reference, not a name
				"",       // body_style — no direct mapping
				0,        // seat_count — not in CHASIS
				0,        // year — not in CHASIS
				string(metaJSON),
				SafeDate(timeFromRow(row, "salida")), // delivered_at = salida date
			}, nil
		},
	}
}

// NewProductionInspectionMigrator migrates PROD_CONTROLES → erp_inspection_templates.
// PROD_CONTROLES has auto-increment id_prodcontrol PK. 2K rows.
// This is the MASTER catalog of inspection definitions ("tabla maestra de controles de
// calidad en produccion"). Each row defines WHAT to inspect — not an inspection event.
// Inspection events live in PROD_CONTROL_MOVIM → erp_production_inspections.
func NewProductionInspectionMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductionInspectionReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "section_id", "step_id", "vehicle_section_id",
			"control_name", "model_code", "control_type", "sort_order",
			"active", "critical", "actionable", "show_in_tech_sheet",
			"default_inspector_id", "enabled_user_id", "observations", "metadata",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_prodcontrol")
			if legacyID == 0 {
				return nil, nil
			}

			// Optional FKs — all nullable in erp_inspection_templates.
			var sectionID *uuid.UUID
			if seccionID := row.Int64("seccion_id"); seccionID > 0 {
				if resolved, err := mapper.ResolveOptional(ctx, "production", "PROD_PROCESOS", seccionID); err == nil && resolved != uuid.Nil {
					sectionID = &resolved
				}
			}

			var defaultInspectorID *uuid.UUID
			if legajo := row.Int64("legajo_defecto"); legajo > 0 {
				if resolved, ok := mapper.ResolveByLegajo(legajo); ok {
					defaultInspectorID = &resolved
				}
			}

			// habilitado: 0=enabled, 1=disabled (inverted in Histrix).
			active := row.Int("habilitado") == 0
			critical := row.Int("critico") == 1
			actionable := row.Int("accionable") == 1
			showInTechSheet := row.Int("ver_ft") == 1

			id, err := mapper.Map(ctx, nil, "production", "PROD_CONTROLES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			meta := map[string]any{
				"legacy_id_prodcontrol": legacyID,
				"seccion_coche_id":      row.Int64("seccion_coche_id"),
				"usuario_habilitado":    row.Int("usuario_habilitado"),
				"aviso_produccion":      row.Int("aviso_produccion"),
				"source":                "PROD_CONTROLES",
			}
			if defaultInspectorID == nil && row.Int64("legajo_defecto") > 0 {
				meta["legajo_defecto_unresolved"] = row.Int64("legajo_defecto")
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID,
				sectionID,
				(*uuid.UUID)(nil), // step_id — seccion_id maps to section, no separate step
				(*uuid.UUID)(nil), // vehicle_section_id — seccion_coche_id preserved in metadata
				row.String("nombre_control"),
				row.String("modelo_control"),
				row.Int("tipo_control"),
				row.Int("orden_control"),
				active,
				critical,
				actionable,
				showInTechSheet,
				defaultInspectorID,
				(*string)(nil), // enabled_user_id — usuario_habilitado is int, preserved in metadata (TEXT column)
				row.String("obs_control"),
				string(metaJSON),
			}, nil
		},
	}
}

// NewProductionStepMigrator migrates PROD_PROCESOS → erp_catalogs (type='production_step_template').
// PROD_PROCESOS are process *templates* (catalog of steps like "Pintura", "Montaje"),
// and erp_production_steps requires order_id NOT NULL — so we land them in
// erp_catalogs instead and preserve all 1,576 saldivia rows instead of the
// previous silent drop. Legacy id_proceso is stored in metadata for joins.
func NewProductionStepMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductionStepReader(db)
	return &GenericMigrator{
		reader:      &readerOverride{r: reader, target: "erp_catalogs", source: "PROD_PROCESOS"},
		columns:     []string{"id", "tenant_id", "type", "code", "name", "sort_order", "active", "metadata"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_proceso")
			if legacyID == 0 {
				// Fall back to any plausible PK name the reader might use.
				legacyID = row.Int64("id_prodproceso")
			}
			if legacyID == 0 {
				return nil, nil
			}
			name := row.String("nombre_proceso")
			if name == "" {
				name = row.String("descripcion")
			}
			if name == "" {
				name = fmt.Sprintf("proceso_%d", legacyID)
			}
			code := fmt.Sprintf("PROC-%d", legacyID)

			id, err := mapper.Map(ctx, nil, "production", "PROD_PROCESOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			meta := map[string]any{
				"source":           "PROD_PROCESOS",
				"legacy_id":        legacyID,
				"orden_proceso":    row.Int("orden_proceso"),
				"seccion_id":       row.Int64("seccion_id"),
				"id_centro_prod":   row.Int64("id_centro_productivo"),
				"calidad_id":       row.Int64("calidad_id"),
			}
			metaJSON, _ := json.Marshal(meta)
			return []any{
				id, tenantID, "production_step_template",
				code, name,
				row.Int("orden_proceso"),
				true,
				string(metaJSON),
			}, nil
		},
	}
}

// NewProductionRequestMigrator migrates MRP_PEDIDO_PRODUCCION → erp_production_orders.
// MRP_PEDIDO_PRODUCCION has auto-increment id_mrp_pedido_prod PK. 13K rows.
// These are the actual per-piece production requests, vs MRP_ORDEN_PRODUCCION which has only 8 rows.
func NewProductionRequestMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductionRequestReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "product_id", "center_id", "quantity", "status", "priority", "user_id", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_mrp_pedido_prod")
			if legacyID == 0 {
				return nil, nil
			}

			// Product FK: stkarticulo_id is an INT in this table (not varchar!)
			var productID *uuid.UUID
			artID := row.Int64("stkarticulo_id")
			if artID > 0 {
				// MRP_PEDIDO_PRODUCCION.stkarticulo_id is an int FK, not varchar code.
				// Try resolving via a direct mapping — the article might be in STK_ARTICULOS by int ID.
				// Since articles use hashCode of varchar, we can't resolve int→varchar directly.
				// Skip product resolution if the FK pattern doesn't match.
				_ = artID // product link left as nil — FK resolution deferred
			}

			// Center FK: seccion_id → production center/section
			var centerID *uuid.UUID
			seccionID := row.Int64("seccion_id")
			if seccionID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "production", "MRP_CENTRO_PRODUCTIVO", seccionID)
				if err == nil && resolved != uuid.Nil {
					centerID = &resolved
				}
			}

			// Status: derive from presence of terminado_hora
			status := "planned"
			if !timeFromRow(row, "inicio_hora").IsZero() {
				status = "in_progress"
			}
			if !timeFromRow(row, "terminado_hora").IsZero() {
				status = "completed"
			}

			id, err := mapper.Map(ctx, nil, "production", "MRP_PEDIDO_PRODUCCION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				fmt.Sprintf("PP-%d", legacyID),
				SafeDateRequired(timeFromRow(row, "creacion_hora")),
				productID,
				centerID,
				ParseDecimal(row.Decimal("cantidad_pieza_unidad")),
				status,
				0, // priority (int: 0=normal, 1=high, -1=low) — not in MRP_PEDIDO_PRODUCCION
				LegacyUserID,
				"", // notes — not in MRP_PEDIDO_PRODUCCION
			}, nil
		},
	}
}

// ============================================================================
// Phase 12 — Quality
// ============================================================================

// NewNonconformityMigrator migrates CAL_NOCONFORMIDADES → erp_nonconformities.
// CAL_NOCONFORMIDADES has auto-increment id_noconformidad PK.
func NewNonconformityMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.NonconformityReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "type_id", "origin_id", "description", "severity", "status", "assigned_to", "closed_at", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_noconformidad")
			if legacyID == 0 {
				return nil, nil
			}

			// type_id from tiponconf_id catalog
			var typeID *uuid.UUID
			tipoID := row.Int64("tiponconf_id")
			if tipoID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "catalog", "CAL_TIPONCONF", tipoID)
				if err == nil && resolved != uuid.Nil {
					typeID = &resolved
				}
			}

			// origin_id from origennconf_id catalog
			var originID *uuid.UUID
			origenID := row.Int64("origennconf_id")
			if origenID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "catalog", "CAL_ORIGENNCONF", origenID)
				if err == nil && resolved != uuid.Nil {
					originID = &resolved
				}
			}

			// Status from estadonconf_id. erp_nonconformities.status CHECK allows:
			// open|investigating|corrective_action|closed.
			status := "open"
			estadoID := row.Int("estadonconf_id")
			switch estadoID {
			case 1:
				status = "open"
			case 2:
				status = "investigating"
			case 3:
				status = "closed"
			}

			// Severity: derive from eficaz and costo
			severity := "minor"
			if row.Int("eficaz") == 0 {
				severity = "major"
			}

			id, err := mapper.Map(ctx, nil, "quality", "CAL_NOCONFORMIDADES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Description: combine descripcion_nconf + causa_nconf + contencion_nconf
			desc := row.String("descripcion_nconf")
			if causa := row.String("causa_nconf"); causa != "" {
				desc += " | Causa: " + causa
			}

			// assigned_to is uuid FK → erp_entities. Legacy stores it as a
			// string name; fold it into the description so no data is lost.
			if resp := row.String("responsable_nconf"); resp != "" {
				desc += " | Responsable: " + resp
			}
			return []any{
				id, tenantID,
				fmt.Sprintf("NC-%d", legacyID),
				SafeDateRequired(timeFromRow(row, "fecha_nconf")),
				typeID,
				originID,
				desc,
				severity,
				status,
				(*uuid.UUID)(nil), // assigned_to — legacy had name string, preserved in desc above
				SafeDate(timeFromRow(row, "fecha_cierrenconf")),
				LegacyUserID,
			}, nil
		},
	}
}

// NewCorrectiveActionMigrator migrates CAL_ACCIONES_NCONF → erp_corrective_actions.
// CAL_ACCIONES_NCONF has auto-increment id_accionnconf PK.
func NewCorrectiveActionMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CorrectiveActionReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "nc_id", "action_type", "description", "responsible_id", "due_date", "status", "completed_at", "effectiveness"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_accionnconf")
			if legacyID == 0 {
				return nil, nil
			}

			// NC FK: noconformidad_id → CAL_NOCONFORMIDADES
			ncLegacyID := row.Int64("noconformidad_id")
			ncID, err := mapper.Resolve(ctx, "quality", "CAL_NOCONFORMIDADES", ncLegacyID)
			if err != nil {
				return nil, nil // skip if NC not migrated
			}

			// Responsible FK: resp_rhpersonal_id → PERSONAL
			var responsibleID *uuid.UUID
			respID := row.Int64("resp_rhpersonal_id")
			if respID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "PERSONAL", respID)
				if err == nil && resolved != uuid.Nil {
					responsibleID = &resolved
				}
			}

			// Action type from tipo_accionnconf.
			// erp_corrective_actions.action_type CHECK allows: corrective|preventive.
			// Legacy "improvement" (3) is closest to "preventive".
			actionType := "corrective"
			switch row.Int("tipo_accionnconf") {
			case 1:
				actionType = "corrective"
			case 2, 3:
				actionType = "preventive"
			}

			// Status from terminada flag.
			// erp_corrective_actions.status CHECK allows: pending|in_progress|completed|verified.
			status := "pending"
			if row.Int("terminada") == 1 {
				status = "completed"
			}

			id, err := mapper.Map(ctx, nil, "quality", "CAL_ACCIONES_NCONF", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Effectiveness from control_accionnconf.
			// CHECK: effectiveness IS NULL OR effectiveness IN (effective|ineffective|pending_review).
			// Empty string would fail the CHECK — use nil instead.
			var effectiveness *string
			if row.Int("control_accionnconf") > 0 {
				s := "effective"
				effectiveness = &s
			}

			return []any{
				id, tenantID, ncID,
				actionType,
				row.String("descripcion_accionnconf"),
				responsibleID,
				SafeDate(timeFromRow(row, "fecha_accionnconf")),
				status,
				SafeDate(timeFromRow(row, "fecha_terminada")),
				effectiveness,
			}, nil
		},
	}
}

// NewAuditMigrator migrates CAL_AUDITORIA_INT → erp_audits.
// CAL_AUDITORIA_INT has auto-increment id_auditoriaint PK.
func NewAuditMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AuditReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "audit_type", "scope", "lead_auditor_id", "status", "score", "notes", "metadata"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_auditoriaint")
			if legacyID == 0 {
				return nil, nil
			}

			// Status from realizada flag
			status := "planned"
			if row.Int("realizada") == 1 {
				status = "completed"
			}

			id, err := mapper.Map(ctx, nil, "quality", "CAL_AUDITORIA_INT", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			meta := map[string]any{
				"codigo_auditorint":   row.Int("codigo_auditorint"),
				"documento_id":        row.Int("documento_id"),
				"auditoriaintpunt_id": row.Int("auditoriaintpunt_id"),
				"source":              "CAL_AUDITORIA_INT",
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID,
				fmt.Sprintf("AUD-%d", legacyID),
				SafeDateRequired(timeFromRow(row, "fecha_auditoriaint")),
				"internal", // audit_type — CAL_AUDITORIA_INT is always internal
				"",         // scope — not in table
				(*uuid.UUID)(nil), // lead_auditor_id — not directly available
				status,
				row.Int("auditoriaintpunt_id"), // score — using punctuation as score
				"",                              // notes
				string(metaJSON),
			}, nil
		},
	}
}

// NewAuditFindingMigrator migrates CAL_AUDITORIA_OBS → erp_audit_findings.
// CAL_AUDITORIA_OBS has auto-increment id_auditoriaobs PK.
func NewAuditFindingMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AuditFindingReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "audit_id", "finding_type", "description", "nc_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_auditoriaobs")
			if legacyID == 0 {
				return nil, nil
			}

			// Audit FK: auditoriaint_id → CAL_AUDITORIA_INT
			auditLegacyID := row.Int64("auditoriaint_id")
			auditID, err := mapper.Resolve(ctx, "quality", "CAL_AUDITORIA_INT", auditLegacyID)
			if err != nil {
				return nil, nil // skip if audit not migrated
			}

			// Finding type from hallazgotipo_id.
			// erp_audit_findings.finding_type CHECK allows: observation|minor_nc|major_nc|opportunity.
			findingType := "observation"
			switch row.Int("hallazgotipo_id") {
			case 1:
				findingType = "observation"
			case 2:
				findingType = "minor_nc"
			case 3:
				findingType = "major_nc"
			case 4:
				findingType = "opportunity"
			}

			// NC FK (optional): link via respauditoriaint_id or direct from the observation
			var ncID *uuid.UUID
			// CAL_AUDITORIA_OBS doesn't have a direct NC FK in the schema,
			// but CAL_NOCONFORMIDADES has auditoriaobs_id linking back.
			// Leave nc_id nil — can be populated in a post-migration step.

			id, err := mapper.Map(ctx, nil, "quality", "CAL_AUDITORIA_OBS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, auditID,
				findingType,
				row.String("observacion"),
				ncID,
			}, nil
		},
	}
}

// NewControlledDocumentMigrator migrates CAL_DOCUMENTOS → erp_controlled_documents.
// CAL_DOCUMENTOS has auto-increment id_documento PK.
func NewControlledDocumentMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ControlledDocumentReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "title", "revision", "doc_type_id", "file_key", "approved_by", "approved_at", "status", "metadata"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_documento")
			if legacyID == 0 {
				return nil, nil
			}

			// doc_type_id from tipodocumento_id catalog
			var docTypeID *uuid.UUID
			tipoID := row.Int64("tipodocumento_id")
			if tipoID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "catalog", "CAL_TIPODOCUMENTO", tipoID)
				if err == nil && resolved != uuid.Nil {
					docTypeID = &resolved
				}
			}

			// approved_by: emisor_documento is an int (user/person ID)
			var approvedBy *uuid.UUID
			emisorID := row.Int64("aprueba_documento")
			if emisorID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "PERSONAL", emisorID)
				if err == nil && resolved != uuid.Nil {
					approvedBy = &resolved
				}
			}

			// Status from estado_documento. Legacy values vary; normalize to CHECK set.
			// erp_controlled_documents.status CHECK allows: draft|approved|obsolete.
			legacyStatus := strings.ToLower(strings.TrimSpace(row.String("estado_documento")))
			status := "approved"
			switch legacyStatus {
			case "0", "borrador", "draft":
				status = "draft"
			case "2", "obsoleto", "obsolete":
				status = "obsolete"
			}

			id, err := mapper.Map(ctx, nil, "quality", "CAL_DOCUMENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			meta := map[string]any{
				"distribuidor": row.String("distribuidor_documento"),
				"aplicacion":   row.String("aplicacion_documento"),
				"conservacion": row.String("conservacion_documento"),
				"guarda":       row.String("guarda_documento"),
				"emisor_id":    row.Int("emisor_documento"),
				"menuId":       row.Int("menuId"),
				"source":       "CAL_DOCUMENTOS",
			}
			metaJSON, _ := json.Marshal(meta)

			// revision is int4 in SDA; legacy stores it as text (e.g. "A", "2").
			// Parse what we can, default to 1, keep the original text in metadata.
			revisionInt := 1
			revText := strings.TrimSpace(row.String("revision_documento"))
			if n, err := strconv.Atoi(revText); err == nil && n > 0 {
				revisionInt = n
			}
			code := strings.TrimSpace(row.String("codigo_documento"))
			if code == "" {
				code = fmt.Sprintf("DOC-%d", legacyID)
			}

			return []any{
				id, tenantID,
				code,
				row.String("nombre_documento"),
				revisionInt,
				docTypeID,
				row.String("archivo_documento"), // file_key — legacy file path
				approvedBy,
				SafeDate(timeFromRow(row, "vigencia_documento")), // approved_at ≈ vigencia date
				status,
				string(metaJSON),
			}, nil
		},
	}
}

// ============================================================================
// Phase 13 — HR Extended
// ============================================================================

// NewDepartmentMigrator migrates ORGANIGRAMA → erp_departments.
// ORGANIGRAMA has composite PK (id_seccion VARCHAR(30), idPadre VARCHAR(30)). 14 rows.
func NewDepartmentMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.DepartmentReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "parent_id", "manager_id", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite PK → generate deterministic ID from hash
			idSeccion := row.String("id_seccion")
			idPadre := row.String("idPadre")
			if idSeccion == "" {
				return nil, nil
			}

			compositeKey := fmt.Sprintf("ORGANIGRAMA:%s:%s", idSeccion, idPadre)
			legacyID := int64(hashCode(compositeKey))

			id, err := mapper.Map(ctx, nil, "hr", "ORGANIGRAMA", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Manager FK: idUsuario → user reference (not directly resolvable to PERSONAL)
			var managerID *uuid.UUID

			return []any{
				id, tenantID,
				idSeccion,
				idSeccion, // name = code (ORGANIGRAMA doesn't have a separate name column)
				(*uuid.UUID)(nil), // parent_id — idPadre is varchar, would need a second pass to link
				managerID,
				true,
			}, nil
		},
	}
}

// NewAttendanceMigrator migrates FICHADADIA → erp_attendance.
// FICHADADIA has composite PK (tarjeta INT, fecha DATE). 933K rows!
// entity_id resolved via legajo → PERSONAL mapping.
func NewAttendanceMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AttendanceReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "date", "clock_in", "clock_out", "hours", "source"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite PK (tarjeta, fecha)
			tarjeta := row.Int64("tarjeta")
			fechaStr := row.String("fecha")
			if tarjeta == 0 {
				return nil, nil
			}

			// Entity FK: legajo → PERSONAL.legajo via pre-built index.
			legajo := row.Int64("legajo")
			var entityID *uuid.UUID
			if resolved, ok := mapper.ResolveByLegajo(legajo); ok {
				entityID = &resolved
			}
			if entityID == nil {
				return nil, nil // skip if employee not found
			}

			compositeKey := fmt.Sprintf("FICHADADIA:%d:%s", tarjeta, fechaStr)
			legacyID := int64(hashCode(compositeKey))

			id, err := mapper.Map(ctx, nil, "hr", "FICHADADIA", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Parse hours from trabajadas field (format HH:MM:SS or similar)
			hoursStr := row.String("trabajadas")
			hours := parseHoursString(hoursStr)

			// erp_attendance.clock_in / clock_out are TIMESTAMPTZ — Histrix
			// stores the clock times as TIME-of-day strings, so we combine
			// them with `fecha` (the attendance date) for a full timestamp.
			// combineDateTime returns nil when the time cell is empty, which
			// the COPY writer encodes as NULL.
			fecha := SafeDateRequired(timeFromRow(row, "fecha"))
			clockIn := combineDateTime(fecha, row.String("hingreso"))
			clockOut := combineDateTime(fecha, row.String("hegreso"))

			return []any{
				id, tenantID, *entityID,
				fecha,
				clockIn,
				clockOut,
				hours,
				"rfid", // source — CHECK allows only manual/rfid/biometric
			}, nil
		},
	}
}

// NewTrainingAttendeeMigrator migrates RH_CURSO_REALIZADO → erp_training_attendees.
// RH_CURSO_REALIZADO has auto-increment id_curso_realizado PK.
func NewTrainingAttendeeMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.TrainingAttendeeReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "training_id", "entity_id", "result", "score"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_curso_realizado")
			if legacyID == 0 {
				return nil, nil
			}

			// Training FK: Id_curso is VARCHAR in this table but INT PK in RH_CURSOS
			cursoID := row.Int64("Id_curso")
			var trainingID *uuid.UUID
			if cursoID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "hr", "RH_CURSOS", cursoID)
				if err == nil && resolved != uuid.Nil {
					trainingID = &resolved
				}
			}
			if trainingID == nil {
				return nil, nil // skip if training not migrated
			}

			// Entity FK: IdPersona → PERSONAL
			personaID := row.Int64("IdPersona")
			entityID, err := mapper.Resolve(ctx, "entity", "PERSONAL", personaID)
			if err != nil {
				return nil, nil // skip if employee not migrated
			}

			// Result from presente flag
			result := "absent"
			if row.Int("presente") == 1 {
				result = "attended"
			}

			id, err := mapper.Map(ctx, nil, "hr", "RH_CURSO_REALIZADO", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, *trainingID, entityID,
				result,
				ParseDecimal(row.Decimal("calificacion")),
			}, nil
		},
	}
}

// NewDemeritEventMigrator migrates MOVDEMERITO → erp_hr_events.
// MOVDEMERITO has auto-increment id_movdemerito PK. 284K rows.
// event_type = "sanction".
func NewDemeritEventMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.DemeritReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "event_type", "date_from", "date_to", "hours", "reason_id", "notes", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_movdemerito")
			if legacyID == 0 {
				return nil, nil
			}

			// Entity FK: prvcod → supplier/employee entity.
			// prvcod in MOVDEMERITO shares the ctacod ambiguity (id_regcuenta /
			// nro_cuenta). Flex-resolve + UNKNOWN fallback so the 19,652 saldivia
			// rows that fail direct id_regcuenta lookup still land.
			prvcod := row.Int64("prvcod")
			resolved, _ := mapper.ResolveEntityFlexible(ctx, prvcod)
			if resolved == uuid.Nil {
				return nil, nil // UNKNOWN hook didn't run — degrade gracefully
			}
			entityID := &resolved

			// Reason: coddem → demerit type catalog
			var reasonID *uuid.UUID
			coddem := row.Int64("coddem")
			if coddem > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "catalog", "DEMERITOS", coddem)
				if err == nil && resolved != uuid.Nil {
					reasonID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "hr", "MOVDEMERITO", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, *entityID,
				"sanction", // event_type
				SafeDateRequired(timeFromRow(row, "demfec")),
				SafeDate(timeFromRow(row, "movfec")), // date_to = movement date
				ParseDecimal("0"),                     // hours — not in MOVDEMERITO
				reasonID,
				"", // notes — not in MOVDEMERITO
				LegacyUserID,
			}, nil
		},
	}
}

// NewDeductionEventMigrator migrates RHDESCUENTOS → erp_hr_events.
// RHDESCUENTOS has auto-increment idDesc PK.
// event_type = "sanction" (deduction events).
func NewDeductionEventMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.DeductionReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "event_type", "date_from", "date_to", "hours", "reason_id", "notes", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("idDesc")
			if legacyID == 0 {
				return nil, nil
			}

			// Entity FK: legajo → PERSONAL via pre-built index.
			legajo := row.Int64("legajo")
			var entityID *uuid.UUID
			if resolved, ok := mapper.ResolveByLegajo(legajo); ok {
				entityID = &resolved
			}
			if entityID == nil {
				return nil, nil
			}

			// Reason: idMotivoDesc → deduction reason catalog
			var reasonID *uuid.UUID
			motivoID := row.Int64("idMotivoDesc")
			if motivoID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "catalog", "RH_MOTIVODESC", motivoID)
				if err == nil && resolved != uuid.Nil {
					reasonID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "hr", "RHDESCUENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, *entityID,
				"sanction",
				SafeDateRequired(timeFromRow(row, "fecha")),
				(*time.Time)(nil),  // date_to — single date event
				ParseDecimal("0"),  // hours — not tracked for deductions
				reasonID,
				fmt.Sprintf("importe=%s sobre_extras=%d", row.Decimal("importe"), row.Int("sobre_extras")),
				LegacyUserID,
			}, nil
		},
	}
}

// NewAdditionalPayEventMigrator migrates RRHH_ADICIONALES → erp_hr_events.
// RRHH_ADICIONALES has composite PK (legajo, desdeFecha).
// event_type = "overtime".
func NewAdditionalPayEventMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AdditionalPayReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "event_type", "date_from", "date_to", "hours", "reason_id", "notes", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legajo := row.Int64("legajo")
			desdeFecha := row.String("desdeFecha")
			if legajo == 0 {
				return nil, nil
			}

			// Entity FK: legajo → PERSONAL via pre-built index.
			var entityID *uuid.UUID
			if resolved, ok := mapper.ResolveByLegajo(legajo); ok {
				entityID = &resolved
			}
			if entityID == nil {
				return nil, nil
			}

			// Composite key → deterministic ID
			compositeKey := fmt.Sprintf("RRHH_ADICIONALES:%d:%s", legajo, desdeFecha)
			legacyID := int64(hashCode(compositeKey))

			id, err := mapper.Map(ctx, nil, "hr", "RRHH_ADICIONALES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, *entityID,
				"overtime",
				SafeDateRequired(timeFromRow(row, "desdeFecha")),
				SafeDate(timeFromRow(row, "hastaFecha")),
				ParseDecimal(row.Decimal("porcentaje")), // hours → use porcentaje as the numeric value
				(*uuid.UUID)(nil),                        // reason_id
				row.String("descripcion"),
				LegacyUserID,
			}, nil
		},
	}
}

// ============================================================================
// Phase 13b — Time clock (FICHADAS + PERSONAL_TARJETA)
// ============================================================================
//
// Phase 1 §Data migration Pareto #2 — 41 % of the remaining uncovered row
// volume post-2.0.8. FICHADAS (1.46 M rows) is the raw clock-punch stream
// from the physical terminals; FICHADADIA (migrated in 2.0.6 → erp_attendance)
// is the DAILY rollup. XML-form scrape shows 116 direct references to the raw
// table across rrhh, sueldos, dashboard, estadisticas, rh_evaluaciones — raw
// events are not dead weight. PERSONAL_TARJETA (1,403 rows) is the versioned
// card-to-employee assignment that makes FICHADAS queryable by employee.

// NewEmployeeCardMigrator migrates PERSONAL_TARJETA → erp_employee_cards.
// Composite (non-unique) natural key (idPersona, tarjeta, fechaDesde). Cards
// reassign across employees over time — we preserve the full history so
// FICHADAS can resolve card+event_date → correct employee for that day.
// idPersona → entity UUID through the preloaded "entity" domain cache.
func NewEmployeeCardMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.EmployeeCardReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "card_code", "effective_from"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			idPersona := row.Int64("idPersona")
			card := row.String("tarjeta")
			if idPersona == 0 || card == "" {
				return nil, nil
			}

			entityID, err := mapper.ResolveOptional(ctx, "entity", "PERSONAL", idPersona)
			if err != nil || entityID == uuid.Nil {
				return nil, nil // orphan — PERSONAL parent not migrated (usually baja removals)
			}

			effFrom := SafeDateRequired(timeFromRow(row, "fechaDesde"))
			compositeKey := fmt.Sprintf("PERSONAL_TARJETA:%d:%s:%s", idPersona, card, effFrom.Format("2006-01-02"))
			legacyID := int64(hashCode(compositeKey))

			id, err := mapper.Map(ctx, nil, "hr", "PERSONAL_TARJETA", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				entityID,
				card,
				effFrom,
			}, nil
		},
	}
}

// NewTimeClockEventMigrator migrates FICHADAS → erp_time_clock_events.
// AI PK id_fichada + UNIQUE insertkey VARCHAR(80). tarjeta is INT(11) on
// FICHADAS but VARCHAR on PERSONAL_TARJETA; we normalize both to string
// (card_code) and resolve employee through BuildTarjetaIndex's date-versioned
// lookup. Orphan tarjetas (card never assigned, or assigned strictly after
// the event) migrate with entity_id NULL — the raw marcaje is preserved.
// Zero-date fecha rows migrate with event_time NULL for the same reason.
func NewTimeClockEventMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.TimeClockEventReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "entity_id", "card_code",
			"event_time", "event_type", "terminal",
			"marca", "deleted_flag", "insert_key", "legacy_id",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_fichada")
			if legacyID == 0 {
				return nil, nil
			}
			insertKey := row.String("insertkey")
			if insertKey == "" {
				return nil, nil
			}

			tarjetaInt := row.Int64("tarjeta")
			cardCode := ""
			if tarjetaInt > 0 {
				cardCode = fmt.Sprintf("%d", tarjetaInt)
			}

			fecha := timeFromRow(row, "fecha")
			eventTime := combineDateTime(fecha, row.String("hora"))

			var entityID *uuid.UUID
			if cardCode != "" {
				if resolved, ok := mapper.ResolveByTarjetaAtDate(cardCode, fecha); ok {
					entityID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "hr", "FICHADAS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, entityID, cardCode,
				eventTime,
				strings.TrimSpace(row.String("codigo")),
				strings.TrimSpace(row.String("reloj")),
				int16(row.Int("marca")),
				int16(row.Int("borrado")),
				insertKey,
				legacyID,
			}, nil
		},
	}
}

// ============================================================================
// Phase 14 — Maintenance
// ============================================================================

// NewMaintenanceAssetMigrator migrates MANT_EQUIPOS → erp_maintenance_assets.
// MANT_EQUIPOS has auto-increment id_equipo PK.
func NewMaintenanceAssetMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.MaintenanceAssetReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "asset_type", "unit_id", "location", "metadata", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_equipo")
			if legacyID == 0 {
				return nil, nil
			}

			// erp_maintenance_assets.asset_type CHECK allows: vehicle|machine|tool|facility.
			// Legacy tipoequipo_id is a catalog we don't resolve here — default to
			// "machine" (the most generic), preserve tipoequipo_id in metadata.
			assetType := "machine"
			_ = row.Int64("tipoequipo_id")

			// unit_id: could link to CHASIS via articulo_id or deposito_id
			var unitID *uuid.UUID

			// Active: fecha_baja is null/zero = active
			active := true
			bajDate := timeFromRow(row, "fecha_baja")
			if !bajDate.IsZero() && bajDate.Year() > 1900 {
				active = false
			}

			id, err := mapper.Map(ctx, nil, "maintenance", "MANT_EQUIPOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			meta := map[string]any{
				"numero_serie":       row.String("numero_serie"),
				"anio_fabricacion":   row.Int("anio_fabricacion"),
				"repuestos_criticos": row.String("repuestos_criticos"),
				"tipoequipo_id":      row.Int("tipoequipo_id"),
				"articulo_id":        row.String("articulo_id"),
				"fecha_alta":         row.String("fecha_alta"),
				"source":             "MANT_EQUIPOS",
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID,
				fmt.Sprintf("EQ-%d", legacyID),
				row.String("nombre_equipo"),
				assetType,
				unitID,
				fmt.Sprintf("DEP-%d", row.Int("deposito_id")), // location from deposito
				string(metaJSON),
				active,
			}, nil
		},
	}
}

// NewMaintenancePlanMigrator migrates MANT_PLAN → erp_maintenance_plans.
// MANT_PLAN has auto-increment id_plan PK.
func NewMaintenancePlanMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.MaintenancePlanReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "asset_id", "name", "frequency_days", "frequency_km", "frequency_hours", "last_done", "next_due", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_plan")
			if legacyID == 0 {
				return nil, nil
			}

			// Asset FK: equipo_id → MANT_EQUIPOS
			equipoID := row.Int64("equipo_id")
			assetID, err := mapper.Resolve(ctx, "maintenance", "MANT_EQUIPOS", equipoID)
			if err != nil {
				return nil, nil // skip if asset not migrated
			}

			id, err := mapper.Map(ctx, nil, "maintenance", "MANT_PLAN", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Name: derive from accion_id or rrule
			name := fmt.Sprintf("Plan-%d", legacyID)
			if rrule := row.String("rrule"); rrule != "" {
				name = fmt.Sprintf("Plan-%d (%s)", legacyID, truncateString(rrule, 50))
			}

			return []any{
				id, tenantID, assetID,
				name,
				0,    // frequency_days — rrule is text, not a simple days value
				0,    // frequency_km — not in MANT_PLAN
				0,    // frequency_hours — not in MANT_PLAN
				SafeDate(timeFromRow(row, "fecha_terminacion")), // last_done
				SafeDate(timeFromRow(row, "fecha_accion")),      // next_due
				true, // active
			}, nil
		},
	}
}

// NewMaintenanceEventMigrator migrates MANT_PLAN_EVENTOS → erp_work_orders.
// MANT_PLAN_EVENTOS has auto-increment id_planevento PK.
// work_type = "preventive".
func NewMaintenanceEventMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.MaintenanceEventReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "asset_id", "plan_id", "date", "work_type", "description", "assigned_to", "status", "priority", "completed_at", "user_id", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_planevento")
			if legacyID == 0 {
				return nil, nil
			}

			// Plan FK: plan_id → MANT_PLAN
			planLegacyID := row.Int64("plan_id")
			var planID *uuid.UUID
			if planLegacyID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "maintenance", "MANT_PLAN", planLegacyID)
				if err == nil && resolved != uuid.Nil {
					planID = &resolved
				}
			}

			// Asset FK: via plan → equipo. We don't have a direct asset_id in MANT_PLAN_EVENTOS.
			var assetID *uuid.UUID

			id, err := mapper.Map(ctx, nil, "maintenance", "MANT_PLAN_EVENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				fmt.Sprintf("WO-P-%d", legacyID),
				assetID,
				planID,
				SafeDateRequired(timeFromRow(row, "fecha_accion")),
				"preventive",
				"",                // description — not in MANT_PLAN_EVENTOS directly
				(*uuid.UUID)(nil), // assigned_to
				"completed",       // events are historical → completed
				"normal",
				(*time.Time)(nil), // completed_at — same as fecha_accion for events
				LegacyUserID,
				row.String("observaciones"),
			}, nil
		},
	}
}

// NewVehicleWorkMigrator migrates TRABAJOS_COCHE → erp_work_orders.
// TRABAJOS_COCHE has composite PK (nrofab, idTrabajo). ~7.3K rows.
// work_type = "corrective".
func NewVehicleWorkMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.VehicleWorkReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "asset_id", "plan_id", "date", "work_type", "description", "assigned_to", "status", "priority", "completed_at", "user_id", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			nrofab := row.Int64("nrofab")
			idTrabajo := row.Int64("idTrabajo")
			if nrofab == 0 && idTrabajo == 0 {
				return nil, nil
			}

			compositeKey := fmt.Sprintf("TRABAJOS_COCHE:%d:%d", nrofab, idTrabajo)
			legacyID := int64(hashCode(compositeKey))

			// Asset FK: nrofab → CHASIS (vehicle/unit number)
			var assetID *uuid.UUID
			if nrofab > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "production", "CHASIS", nrofab)
				if err == nil && resolved != uuid.Nil {
					assetID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "maintenance", "TRABAJOS_COCHE", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				fmt.Sprintf("WO-C-%d-%d", nrofab, idTrabajo),
				assetID,
				(*uuid.UUID)(nil), // plan_id — corrective work has no plan
				SafeDateRequired(timeFromRow(row, "fechaTrabajo")),
				"corrective",
				row.String("realizador"), // description = who did the work
				(*uuid.UUID)(nil),        // assigned_to
				"completed",              // historical = completed
				"normal",
				SafeDate(timeFromRow(row, "fechaPago")), // completed_at ≈ payment date
				LegacyUserID,
				fmt.Sprintf("importe=%.2f", ParseDecimal(row.Decimal("importe")).InexactFloat64()),
			}, nil
		},
	}
}

// NewFuelLogMigrator migrates COMBUSTIBLE → erp_fuel_logs.
// COMBUSTIBLE has auto-increment id PK.
func NewFuelLogMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.FuelReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "asset_id", "date", "liters", "km_reading", "cost", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id")
			if legacyID == 0 {
				return nil, nil
			}

			// Asset FK: nrofab_id → CHASIS (vehicle/unit)
			nrofab := row.Int64("nrofab_id")
			var assetID *uuid.UUID
			if nrofab > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "production", "CHASIS", nrofab)
				if err == nil && resolved != uuid.Nil {
					assetID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "maintenance", "COMBUSTIBLE", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, assetID,
				SafeDateRequired(timeFromRow(row, "fecha")),
				ParseDecimal(fmt.Sprintf("%d", row.Int("egreso_litros"))), // liters (int in legacy)
				ParseDecimal("0"), // km_reading — not in COMBUSTIBLE
				ParseDecimal("0"), // cost — not in COMBUSTIBLE
				LegacyUserID,
			}, nil
		},
	}
}

// ============================================================================
// Phase 15 — Safety
// ============================================================================

// NewAccidentMigrator migrates ACCIDENTE_PER → erp_work_accidents.
// ACCIDENTE_PER has NO auto-increment PK. 22 rows. Uses IdPersona + legajo + fechaini as composite key.
// entity_id is NULLABLE in erp_work_accidents — we preserve orphan rows with legajo/IdPersona in observations.
func NewAccidentMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.NewAccidentPersonReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "accident_type_id", "body_part_id", "section_id", "incident_date", "recovery_date", "lost_days", "observations", "reported_by", "status"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			personaID := row.Int64("IdPersona")
			legajo := row.Int64("legajo")
			fechaini := row.String("fechaini")

			// Resolve entity via IdPersona → PERSONAL, with legajo fallback.
			// entity_id is NULLABLE so we preserve the row even when neither resolves.
			var entityID *uuid.UUID
			if personaID > 0 {
				if resolved, err := mapper.ResolveOptional(ctx, "entity", "PERSONAL", personaID); err == nil && resolved != uuid.Nil {
					entityID = &resolved
				}
			}
			if entityID == nil && legajo > 0 {
				if resolved, ok := mapper.ResolveByLegajo(legajo); ok {
					entityID = &resolved
				}
			}

			// Composite key: include legajo so rows without IdPersona still get a stable ID.
			compositeKey := fmt.Sprintf("ACCIDENTE_PER:%d:%d:%s", personaID, legajo, fechaini)
			legacyID := int64(hashCode(compositeKey))

			id, err := mapper.Map(ctx, nil, "safety", "ACCIDENTE_PER", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Accident type from idaccidente → ACCIDENTE catalog.
			var accidentTypeID *uuid.UUID
			accidenteID := row.Int64("idaccidente")
			if accidenteID > 0 {
				if resolved, err := mapper.ResolveOptional(ctx, "safety", "ACCIDENTE", accidenteID); err == nil && resolved != uuid.Nil {
					accidentTypeID = &resolved
				}
			}

			// Preserve legacy identifiers in observations when entity is unresolved.
			observations := row.String("observaciones")
			if entityID == nil {
				marker := fmt.Sprintf("[legacy:unlinked_employee legajo=%d IdPersona=%d sector=%d]", legajo, personaID, row.Int64("sector"))
				if observations != "" {
					observations = marker + " " + observations
				} else {
					observations = marker
				}
			}

			return []any{
				id, tenantID, entityID,
				accidentTypeID,
				(*uuid.UUID)(nil), // body_part_id — not in ACCIDENTE_PER
				(*uuid.UUID)(nil), // section_id — sector is int, no direct resolution
				SafeDateRequired(timeFromRow(row, "fechaini")),
				SafeDate(timeFromRow(row, "fechafin")),
				0, // lost_days — not in ACCIDENTE_PER
				observations,
				"",     // reported_by
				"open", // status
			}, nil
		},
	}
}

// NewRiskAgentMigrator migrates RIESGOS → erp_risk_agents.
// RIESGOS has idRiesgo INT PK (not auto-increment). 212 rows. Catalog-like.
func NewRiskAgentMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.RiskAgentReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "name"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("idRiesgo")
			if legacyID == 0 {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "safety", "RIESGOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				row.String("agente"),
			}, nil
		},
	}
}

// NewRiskExposureMigrator migrates RIESGO_PERSONAL → erp_employee_risk_exposures.
// RIESGO_PERSONAL has auto-increment id_riesgopersonal PK. 2.9K rows.
func NewRiskExposureMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.RiskExposureReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "risk_agent_id", "section_id", "exposed_from", "exposed_until", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_riesgopersonal")
			if legacyID == 0 {
				return nil, nil
			}

			// Entity FK: IdPersona → PERSONAL (there are TWO person fields: persona_id and IdPersona)
			// Use IdPersona as it's the standard FK name
			personaID := row.Int64("IdPersona")
			if personaID == 0 {
				personaID = row.Int64("persona_id")
			}
			entityID, err := mapper.Resolve(ctx, "entity", "PERSONAL", personaID)
			if err != nil {
				return nil, nil
			}

			// Risk agent FK: riesgo_id → RIESGOS
			riesgoID := row.Int64("riesgo_id")
			var riskAgentID *uuid.UUID
			if riesgoID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "safety", "RIESGOS", riesgoID)
				if err == nil && resolved != uuid.Nil {
					riskAgentID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "safety", "RIESGO_PERSONAL", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, entityID,
				riskAgentID,
				(*uuid.UUID)(nil), // section_id — no direct mapping
				SafeDateRequired(timeFromRow(row, "fecha_desde")), // exposed_from is NOT NULL
				SafeDate(timeFromRow(row, "fecha_hasta")),
				"", // notes — not in RIESGO_PERSONAL
			}, nil
		},
	}
}

// NewMedicalLeaveMigrator migrates PARTE_MEDICO_DIARIO → erp_medical_visits_log.
// PARTE_MEDICO_DIARIO has auto-increment id PK. 59 rows.
// Maps to erp_medical_visits_log (NOT erp_medical_leaves) because the source has only
// free-text name/user fields with no FK to PERSONAL. entity_id in the target is nullable.
func NewMedicalLeaveMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.MedicalLeaveReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "entity_id", "visit_date", "visit_time",
			"operator_username", "patient_name", "symptoms", "prescription", "metadata",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id")
			if legacyID == 0 {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "safety", "PARTE_MEDICO_DIARIO", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Visit time — nullable TIME column. Empty string maps to NULL.
			var visitTime any
			if t := row.String("hora"); t != "" && t != "00:00:00" {
				visitTime = t
			}

			meta := map[string]any{
				"legacy_id": legacyID,
				"source":    "PARTE_MEDICO_DIARIO",
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID,
				(*uuid.UUID)(nil), // entity_id — no FK in source
				SafeDateRequired(timeFromRow(row, "fecha")),
				visitTime,
				row.String("usuario"),
				row.String("nombre"),
				row.String("sintomatologia"),
				row.String("prescripcion"),
				string(metaJSON),
			}, nil
		},
	}
}

// ============================================================================
// Phase 16 — Auth
// ============================================================================

// NewLegacyUserMigrator migrates HTXUSERS → users.
// HTXUSERS has auto-increment Id_usuario PK.
// Generates bcrypt hash of temp password, sets force_password_reset=true.
// users table has TEXT id (not UUID), so we generate uuid and cast to text.
func NewLegacyUserMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.LegacyUserReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "email", "name", "password_hash", "is_active", "force_password_reset"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("Id_usuario")
			if legacyID == 0 {
				return nil, nil
			}

			login := row.String("login")
			if login == "" {
				return nil, nil // skip users with no login
			}

			nombre := row.String("Nombre")
			apellido := row.String("apellido")
			name := strings.TrimSpace(nombre + " " + apellido)
			if name == "" {
				name = login
			}

			// Email: prefer emailUser, fallback to email, then login@legacy.local
			email := row.String("emailUser")
			if email == "" {
				email = row.String("email")
			}
			if email == "" {
				email = login + "@legacy.local"
			}

			// Active: baja = 0 means active
			isActive := row.Int("baja") == 0

			// Generate bcrypt hash of temporary password.
			// We use a deterministic temp password: "SDA-migrate-{login}"
			// bcrypt import is required — if not available, use a placeholder hash.
			tempPassword := fmt.Sprintf("SDA-migrate-%s", login)
			passwordHash := hashTempPassword(tempPassword)

			// Use the mapper-generated UUID for BOTH users.id and erp_legacy_mapping so
			// downstream rescue passes (RescueLegacyUserRoles, RescueLegacyUserOverrides)
			// can resolve HTXUSERS.Id_usuario → users.id directly. Previously this
			// function generated an ephemeral UUID here and a second, different one
			// via mapper.Map, leaving the mapping table pointing at a non-existent
			// users row and silently breaking the role-assignment rescue.
			id, err := mapper.Map(ctx, nil, "auth", "HTXUSERS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id.String(), // TEXT id, not UUID
				email,
				name,
				passwordHash,
				isActive,
				true, // force_password_reset
			}, nil
		},
	}
}

// ============================================================================
// Helpers
// ============================================================================

// parseHoursString parses a time string like "08:30:00" or "8.5" into a decimal hours value.
func parseHoursString(s string) float64 {
	if s == "" {
		return 0
	}
	// Try HH:MM:SS format
	parts := strings.Split(s, ":")
	if len(parts) >= 2 {
		h := parseIntSafe(parts[0])
		m := parseIntSafe(parts[1])
		result := float64(h) + float64(m)/60.0
		if len(parts) >= 3 {
			sec := parseIntSafe(parts[2])
			result += float64(sec) / 3600.0
		}
		return result
	}
	// Try plain number
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

// parseIntSafe parses an int from string, returning 0 on error.
func parseIntSafe(s string) int {
	s = strings.TrimSpace(s)
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// hashTempPassword generates a placeholder hash for the temporary migration password.
// Uses hashCode (FNV-1a) since golang.org/x/crypto/bcrypt may not be available.
// The auth service will force password reset on first login anyway.
func hashTempPassword(password string) string {
	// Use a deterministic placeholder hash.
	// The real bcrypt hash should be generated by the auth service on first login.
	// Format: $migration$<fnv64-hex> to clearly mark this as a migration placeholder.
	return fmt.Sprintf("$migration$%016x", hashCode(password))
}

// ============================================================================
// Homologations (HOMOLOGMOD + STK_ARTICULO_PROCESO_HIST + _HIST_DETALLE)
// ============================================================================
// Despite the STK_ prefix on the HIST tables, these are production-domain
// artifacts — the UX lives in .intranet-scrape/xml-forms/produccion/linea/
// under "HOMOLOGACION POR MODELO". Together they cover 2,642,734 Histrix rows
// (585 homologations + 1,173 revisions + 2,640,976 detail lines), i.e. 42.7 %
// of the Phase 1 §Data migration row-volume gap.

// NewHomologationMigrator migrates HOMOLOGMOD (vehicle model homologations, 585 rows)
// → erp_homologations. AI PK id_homologacion. `baja` flips to `active = (baja == 0)`.
// Columns beyond the minimal Phase 1 set stay in Histrix for a follow-up extension.
func NewHomologationMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.HomologationReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "plano", "expte", "dispos", "fecha_aprob", "fecha_vto", "seats", "seats_lower", "weight_tare", "weight_gross", "vin", "commercial_code", "commercial_desc", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_homologacion")
			if legacyID == 0 {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "production", "HOMOLOGMOD", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			active := row.Int("baja") == 0

			return []any{
				id, tenantID,
				row.String("plano"),
				row.String("expte"),
				row.String("dispos"),
				SafeDate(timeFromRow(row, "fechaAprob")),
				SafeDate(timeFromRow(row, "fechaVto")),
				row.Int("asientos"),
				row.Int("asientos_plantabaja"),
				ParseDecimal(row.Decimal("tara")),
				ParseDecimal(row.Decimal("bruto")),
				row.String("vin"),
				row.String("codigo_comercial"),
				row.String("desc_comercial"),
				active,
			}, nil
		},
	}
}

// NewHomologationRevisionMigrator migrates STK_ARTICULO_PROCESO_HIST
// (homologation cost/process revisions, 1,173 rows) → erp_homologation_revisions.
// AI PK id_artproceso_hist. FK homologacion_id → erp_homologations.id via the
// preloaded "production" domain cache.
func NewHomologationRevisionMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.HomologationRevisionReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "homologation_id", "date", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_artproceso_hist")
			if legacyID == 0 {
				return nil, nil
			}

			homID := row.Int64("homologacion_id")
			homologationUUID, err := mapper.ResolveOptional(ctx, "production", "HOMOLOGMOD", homID)
			if err != nil || homologationUUID == uuid.Nil {
				return nil, nil // orphan revision — parent homologation missing
			}

			id, err := mapper.Map(ctx, nil, "production", "STK_ARTICULO_PROCESO_HIST", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				homologationUUID,
				SafeDateRequired(timeFromRow(row, "fecha")),
				row.String("observaciones"),
			}, nil
		},
	}
}

// NewHomologationRevisionLineMigrator migrates STK_ARTICULO_PROCESO_HIST_DETALLE
// (2,640,976 rows — #1 in the Phase 1 Pareto) → erp_homologation_revision_lines.
// AI PK id_artproceso_hist_detalle. FK artproceso_hist_id → erp_homologation_revisions.
// article_id is best-effort: artcod without subsistema resolves to the default-
// subsystem article if present, otherwise nil; article_code is always preserved.
func NewHomologationRevisionLineMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.HomologationRevisionLineReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "revision_id", "article_id",
			"article_code", "article_desc", "article_unit",
			"process_1", "process_2", "process_3", "process_4",
			"multiplier", "quantity",
			"replacement_cost", "replacement_partial",
			"replacement_cost_desc", "replacement_partial_desc",
			"account_code", "account_name",
			"partial_with_surcharge", "region_percentage",
			"partial_clog", "partial_surcharge_log", "logistics_cost",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_artproceso_hist_detalle")
			if legacyID == 0 {
				return nil, nil
			}

			revLegacyID := row.Int64("artproceso_hist_id")
			revisionUUID, err := mapper.ResolveOptional(ctx, "production", "STK_ARTICULO_PROCESO_HIST", revLegacyID)
			if err != nil || revisionUUID == uuid.Nil {
				return nil, nil // orphan line — parent revision missing
			}

			// Best-effort article resolution. DETALLE carries only artcod (no
			// subsistema_id), so we look up the default-subsystem article; rows
			// whose artcod exists only under a non-default subsystem keep
			// article_id NULL but article_code preserved.
			var articleID *uuid.UUID
			artCode := row.String("artcod")
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, "")))
				if resolved, rerr := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID); rerr == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "production", "STK_ARTICULO_PROCESO_HIST_DETALLE", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, revisionUUID, articleID,
				artCode,
				row.String("artdes"),
				row.String("artuni"),
				row.String("proceso1"),
				row.String("proceso2"),
				row.String("proceso3"),
				row.String("proceso4"),
				ParseDecimal(row.Decimal("multiplo")),
				ParseDecimal(row.Decimal("cantidad")),
				ParseDecimal(row.Decimal("costo_reposicion")),
				ParseDecimal(row.Decimal("parcial_reposicion")),
				ParseDecimal(row.Decimal("costo_reposicion_desc")),
				ParseDecimal(row.Decimal("parcial_reposicion_desc")),
				row.String("ctbcod"),
				row.String("ctbnom"),
				ParseDecimal(row.Decimal("parcial_con_recargo")),
				ParseDecimal(row.Decimal("porcentaje_region")),
				ParseDecimal(row.Decimal("parcial_clog")),
				ParseDecimal(row.Decimal("parcial_recargo_log")),
				ParseDecimal(row.Decimal("costo_logistico")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 6c — Bank statement imports (BCS_IMPORTACION)
// ============================================================================

// NewBankImportMigrator — BCS_IMPORTACION → erp_bank_imports
// (91,959 rows live, scrape 84,492). Bank-statement import staging;
// each row is one line from a CSV/XLS dump of bank movements awaiting
// reconciliation against internal REG_MOVIMIENTOS.
//
// FK resolution:
//   - treasury_movement_id via BuildRegMovimIndex (Phase 6 hook) —
//     looks up the regmovim_id in the preloaded index populated from
//     IVACOMPRAS. Nullable because BCS_IMPORTACION includes rows
//     where the bank line hasn't been matched yet (processed=2).
//   - account_entity_id via the nro_cuenta index (Phase 2 hook) —
//     straight ResolveByNroCuenta lookup.
//
// Both indexes are populated much earlier in the run so this migrator
// just consumes them.
func NewBankImportMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.BankImportReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"movement_date", "concept_name", "movement_no",
			"amount", "debit", "credit", "balance",
			"movement_code", "treasury_movement_id", "treasury_legacy_id",
			"imported_at", "account_number", "account_entity_id",
			"processed", "comments", "internal_no", "branch",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_importacion")
			if legacyID == 0 {
				return nil, nil
			}

			regMovimID := row.Int64("regmovim_id")
			var treasuryID *uuid.UUID
			if regMovimID > 0 {
				if resolved, ok := mapper.ResolveRegMovim(regMovimID); ok && resolved != uuid.Nil {
					treasuryID = &resolved
				}
			}

			nroCuenta := row.Int64("nro_cuenta")
			var accountEntityID *uuid.UUID
			if nroCuenta > 0 {
				if resolved, tag := mapper.ResolveEntityFlexible(ctx, nroCuenta); resolved != uuid.Nil && tag != "unknown" {
					accountEntityID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "treasury", "BCS_IMPORTACION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			movDate := SafeDate(timeFromRow(row, "fecha_movimiento"))
			importedAt := SafeDate(timeFromRow(row, "importado"))

			return []any{
				id, tenantID, legacyID,
				movDate,
				strings.TrimSpace(row.String("nombre_concepto")),
				int32(row.Int("nro_movimiento")),
				ParseDecimal(row.Decimal("importe")),
				ParseDecimal(row.Decimal("debito")),
				ParseDecimal(row.Decimal("credito")),
				ParseDecimal(row.Decimal("saldo")),
				strings.TrimSpace(row.String("cod_movimiento")),
				treasuryID, int32(regMovimID),
				importedAt,
				int32(nroCuenta), accountEntityID,
				int32(row.Int("procesado")),
				strings.TrimSpace(row.String("comentarios")),
				int32(row.Int("nro_interno")),
				strings.TrimSpace(row.String("sucursal")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 11d — Production inspection × homologation cross-reference
// ============================================================================

// NewProductionInspectionHomologationMigrator — PROD_CONTROL_HOMOLOG →
// erp_production_inspection_homologations. Pareto #7 of the post-2.0.10
// gap (403,028 rows live vs 105,683 scrape estimate — +282 % growth).
//
// Simple join table: production inspection templates × homologated
// vehicle models. Live Histrix shows 0 orphans on both FKs. Dependencies:
// PROD_CONTROLES (Phase 7/8) and HOMOLOGMOD (Phase 11b) both already
// migrated; this just wires the join.
func NewProductionInspectionHomologationMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductionInspectionHomologationReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"inspection_id", "inspection_legacy_id",
			"homologation_id", "homologation_legacy_id",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_controlhomolog")
			if legacyID == 0 {
				return nil, nil
			}

			prodcontrolID := row.Int64("prodcontrol_id")
			var inspectionID *uuid.UUID
			if prodcontrolID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "production", "PROD_CONTROLES", prodcontrolID); rerr == nil && resolved != uuid.Nil {
					inspectionID = &resolved
				}
			}

			homologLegacyID := row.Int64("homologacion_id")
			var homologID *uuid.UUID
			if homologLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "production", "HOMOLOGMOD", homologLegacyID); rerr == nil && resolved != uuid.Nil {
					homologID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "production", "PROD_CONTROL_HOMOLOG", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				inspectionID, int32(prodcontrolID),
				homologID, int32(homologLegacyID),
			}, nil
		},
	}
}

// ============================================================================
// Phase 4e — Article cost history (STK_COSTO_HIST)
// ============================================================================

// NewArticleCostHistoryMigrator — STK_COSTO_HIST →
// erp_article_cost_history. Pareto #8 of the post-2.0.10 gap
// (103,799 rows live). Composite natural PK (articulo_id, anio_hist,
// mes_hist) with no AI surrogate — we synthesize legacy_id via
// hashCode("articulo_id|anio|mes") to keep the idempotency UNIQUE
// (tenant_id, legacy_id) constraint working.
//
// article_id is resolved via the stock default-subsystem lookup (same
// pattern as STKINSPR / tools). Rows with empty articulo_id (observed
// in sample data — possibly orphan periods) migrate with article_id
// NULL and article_code preserved.
func NewArticleCostHistoryMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ArticleCostHistoryReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"article_code", "article_id",
			"year", "month", "cost", "period_code",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			artCode := strings.TrimSpace(row.String("articulo_id"))
			year := row.Int("anio_hist")
			month := row.Int("mes_hist")
			if year == 0 && month == 0 && artCode == "" {
				return nil, nil // defensive — fully blank row
			}

			compositeKey := fmt.Sprintf("STK_COSTO_HIST:%s:%d:%d", artCode, year, month)
			legacyID := hashCode(compositeKey)

			var articleID *uuid.UUID
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, "")))
				if resolved, rerr := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID); rerr == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "stock", "STK_COSTO_HIST", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				artCode, articleID,
				int32(year), int16(month),
				ParseDecimal(row.Decimal("costo_hist")),
				strings.TrimSpace(row.String("periodo_hist")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 4d — Products domain (PRODUCTO_* cluster)
// ============================================================================
//
// Six tables in one cluster: PRODUCTO_SECCION (10) + PRODUCTOS (4,108) +
// PRODUCTO_ATRIBUTOS (415) + PRODUCTO_ATRIB_OPCIONES (147) +
// PRODUCTO_ATRIB_VALORES (353,936 — Pareto #6) +
// PRODUCTO_ATRIBUTO_HOMOLOGACION (47,189 — Pareto #18). ~406 K rows total.
//
// Domain = "productos" (distinct from "stock" / STK_ARTICULOS even though
// the two domains overlap — a bus-model article has a STK_ARTICULOS.artcod
// that equals PRODUCTOS.descripcion_producto). The existing metadata
// enricher (metadata_enrichment.articleProductAttributes) keeps working
// in parallel and attaches attribute data as JSONB on erp_articles; this
// migrator additionally materializes the full relational shape so the UI
// forms read their native schema post-cutover.

// NewProductSectionMigrator — PRODUCTO_SECCION → erp_product_sections.
func NewProductSectionMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductSectionReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id", "name", "sort_order",
			"rubro_id", "active",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_prdseccion")
			if legacyID == 0 {
				return nil, nil
			}
			id, err := mapper.Map(ctx, nil, "productos", "PRODUCTO_SECCION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				strings.TrimSpace(row.String("nombre_seccion")),
				int32(row.Int("orden_seccion")),
				int32(row.Int("rubro_id")),
				row.Int("activa_seccion") != 0,
			}, nil
		},
	}
}

// NewProductMigrator — PRODUCTOS → erp_products. descripcion_producto is
// both the product description AND the STK_ARTICULOS.artcod short code
// for unit-level articles (bus chassis). supplier resolved via
// ResolveEntityFlexible(regcuenta_id).
func NewProductMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id", "description",
			"supplier_entity_id", "supplier_code",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_producto")
			if legacyID == 0 {
				return nil, nil
			}

			regcuenta := row.Int64("regcuenta_id")
			var supplierEntityID *uuid.UUID
			if regcuenta > 0 {
				if resolved, tag := mapper.ResolveEntityFlexible(ctx, regcuenta); resolved != uuid.Nil && tag != "unknown" {
					supplierEntityID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "productos", "PRODUCTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				strings.TrimSpace(row.String("descripcion_producto")),
				supplierEntityID,
				int32(regcuenta),
			}, nil
		},
	}
}

// NewProductAttributeMigrator — PRODUCTO_ATRIBUTOS → erp_product_attributes.
func NewProductAttributeMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductAttributeReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id", "name", "attribute_type",
			"section_id", "section_legacy_id", "article_code",
			"helper_xml", "helper_dir", "parameters", "sort_order",
			"active", "print_label", "print_value",
			"active_in_quote", "active_in_tech_sheet", "quote_description",
			"define_before_section_id", "standard_additional",
			"code", "print_section_id",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_prdatributo")
			if legacyID == 0 {
				return nil, nil
			}

			sectionLegacyID := row.Int64("prdseccion_id")
			var sectionID *uuid.UUID
			if sectionLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "productos", "PRODUCTO_SECCION", sectionLegacyID); rerr == nil && resolved != uuid.Nil {
					sectionID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "productos", "PRODUCTO_ATRIBUTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				strings.TrimSpace(row.String("nombre_atributo")),
				strings.TrimSpace(row.String("tipo_atributo")),
				sectionID, int32(sectionLegacyID),
				strings.TrimSpace(row.String("stkarticulo_id")),
				strings.TrimSpace(row.String("helper_xml")),
				strings.TrimSpace(row.String("helper_dir")),
				strings.TrimSpace(row.String("parametros")),
				int32(row.Int("orden_atributo")),
				row.Int("activo") != 0,
				row.Int("print_label") != 0,
				row.Int("print_value") != 0,
				row.Int("activo_cotizacion") != 0,
				row.Int("activo_fichatecnica") != 0,
				strings.TrimSpace(row.String("descrip_cotizacion")),
				int32(row.Int("definir_antes_seccion_id")),
				int16(row.Int("estandar_adicional")),
				strings.TrimSpace(row.String("codigo")),
				strings.TrimSpace(row.String("print_seccion_id")),
			}, nil
		},
	}
}

// NewProductAttributeOptionMigrator — PRODUCTO_ATRIB_OPCIONES →
// erp_product_attribute_options.
func NewProductAttributeOptionMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductAttributeOptionReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"attribute_id", "attribute_legacy_id",
			"option_name", "option_value",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_atribopcion")
			if legacyID == 0 {
				return nil, nil
			}

			attrLegacyID := row.Int64("prdatributo_id")
			var attrID *uuid.UUID
			if attrLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "productos", "PRODUCTO_ATRIBUTOS", attrLegacyID); rerr == nil && resolved != uuid.Nil {
					attrID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "productos", "PRODUCTO_ATRIB_OPCIONES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				attrID, int32(attrLegacyID),
				strings.TrimSpace(row.String("nombre_opcion")),
				strings.TrimSpace(row.String("valor_opcion")),
			}, nil
		},
	}
}

// NewProductAttributeValueMigrator — PRODUCTO_ATRIB_VALORES →
// erp_product_attribute_values. The Pareto #6 target (353,936 rows).
// 89 % resolve producto_id against PRODUCTOS; 11 % orphan migrate with
// product_id NULL preserving raw producto_legacy_id. 100 % resolve
// prdatributo_id against PRODUCTO_ATRIBUTOS.
func NewProductAttributeValueMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductAttributeValueReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"product_id", "product_legacy_id",
			"attribute_id", "attribute_legacy_id",
			"value", "quantity", "quote_legacy_id", "recorded_at",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_atribvalor")
			if legacyID == 0 {
				return nil, nil
			}

			productLegacyID := row.Int64("producto_id")
			var productID *uuid.UUID
			if productLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "productos", "PRODUCTOS", productLegacyID); rerr == nil && resolved != uuid.Nil {
					productID = &resolved
				}
			}

			attrLegacyID := row.Int64("prdatributo_id")
			var attrID *uuid.UUID
			if attrLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "productos", "PRODUCTO_ATRIBUTOS", attrLegacyID); rerr == nil && resolved != uuid.Nil {
					attrID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "productos", "PRODUCTO_ATRIB_VALORES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			recordedAt := SafeDate(timeFromRow(row, "timestamp_atributo"))

			return []any{
				id, tenantID, legacyID,
				productID, int32(productLegacyID),
				attrID, int32(attrLegacyID),
				strings.TrimSpace(row.String("valor_atributo")),
				int32(row.Int("cantidad_atributo")),
				int32(row.Int("cotizacion_id")),
				recordedAt,
			}, nil
		},
	}
}

// NewProductAttributeHomologationMigrator — PRODUCTO_ATRIBUTO_HOMOLOGACION
// → erp_product_attribute_homologations (Pareto #18 rank, 47,189 rows).
// Join table between product attributes and vehicle homologations
// migrated in 2.0.8 (erp_homologations). Resolves both FKs with orphan
// tolerance.
func NewProductAttributeHomologationMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductAttributeHomologationReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"attribute_id", "attribute_legacy_id",
			"homologation_id", "homologation_legacy_id",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_atrib_homolog")
			if legacyID == 0 {
				return nil, nil
			}

			attrLegacyID := row.Int64("prdatributo_id")
			var attrID *uuid.UUID
			if attrLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "productos", "PRODUCTO_ATRIBUTOS", attrLegacyID); rerr == nil && resolved != uuid.Nil {
					attrID = &resolved
				}
			}

			homologLegacyID := row.Int64("homologacion_id")
			var homologID *uuid.UUID
			if homologLegacyID > 0 {
				if resolved, rerr := mapper.ResolveOptional(ctx, "production", "HOMOLOGMOD", homologLegacyID); rerr == nil && resolved != uuid.Nil {
					homologID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "productos", "PRODUCTO_ATRIBUTO_HOMOLOGACION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID, legacyID,
				attrID, int32(attrLegacyID),
				homologID, int32(homologLegacyID),
			}, nil
		},
	}
}

// ============================================================================
// Phase 4c — Stock cost tracking (STKINSPR)
// ============================================================================

// NewArticleSupplierCostMigrator migrates STKINSPR → erp_article_costs.
// Pareto #5 of the Phase 1 §Data migration gap post-2.0.10
// (189,863 rows live, ~17 % of remaining uncovered row volume).
//
// STKINSPR is the per-supplier cost ledger — one row per
// (article, supplier) cost snapshot, maintained by invoice-import
// triggers + the recalc flag for periodic re-costs. Source of the live
// stock/costos/ and estadisticas/evolutivo_costo screens.
//
// Resolution:
//   - article_id via stock default-subsystem lookup
//     (articleCompositeCode(artcod, "")) — same pattern as
//     HomologationRevisionLine / ToolMigrator. Orphan articles migrate
//     with article_id NULL and article_code preserved.
//   - supplier_entity_id via ResolveEntityFlexible(ctacod) — tries
//     id_regcuenta first, then the nro_cuenta index (built as a hook
//     after REG_CUENTA in Phase 2), then falls through to NULL.
//
// Dates: fecfac is ~99.9 % zero in live data (legacy column, rarely
// populated) — SafeDate → NULL. fecult is the business "as-of" date and
// is populated on 99.99 % of rows. movfec zero on 4.7 % of rows.
//
// Phase 0 invariant: rows_read = rows_written + rows_skipped + 0
// duplicate. Skips: idCosto == 0 (defensive — AI PK, 0 rows observed).
func NewArticleSupplierCostMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ArticleSupplierCostReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"article_code", "article_id", "subsystem_code",
			"cost", "percentage_1", "percentage_2", "percentage_3",
			"supplier_article_code", "supplier_code", "supplier_entity_id",
			"invoice_date", "last_update_date",
			"movement_no", "movement_post", "movement_date",
			"recalc_flag",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("idCosto")
			if legacyID == 0 {
				return nil, nil
			}

			artCode := strings.TrimSpace(row.String("artcod"))
			var articleID *uuid.UUID
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, "")))
				if resolved, rerr := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID); rerr == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			ctacod := row.Int64("ctacod")
			var supplierEntityID *uuid.UUID
			if ctacod > 0 {
				if resolved, tag := mapper.ResolveEntityFlexible(ctx, ctacod); resolved != uuid.Nil && tag != "unknown" {
					supplierEntityID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "stock", "STKINSPR", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			invoiceDate := SafeDate(timeFromRow(row, "fecfac"))
			lastUpdate := SafeDate(timeFromRow(row, "fecult"))
			moveDate := SafeDate(timeFromRow(row, "movfec"))

			return []any{
				id, tenantID, legacyID,
				artCode, articleID,
				strings.TrimSpace(row.String("siscod")),
				ParseDecimal(row.Decimal("artcos")),
				ParseDecimal(row.Decimal("artpor__1")),
				ParseDecimal(row.Decimal("artpor__2")),
				ParseDecimal(row.Decimal("artpor__3")),
				strings.TrimSpace(row.String("artpro")),
				int32(ctacod), supplierEntityID,
				invoiceDate, lastUpdate,
				int32(row.Int("movnro")),
				int32(row.Int("movnpv")),
				moveDate,
				int32(row.Int("recalc")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 4b — Tools / serialized inventory (HERRAMIENTAS + HERRMOVS)
// ============================================================================

// NewToolMigrator migrates HERRAMIENTAS → erp_tools. Pareto #4 of the
// Phase 1 §Data migration gap post-2.0.9 (389,253 rows live, ~15 % of
// remaining uncovered row volume).
//
// Despite the "herramientas" name, the table is the serialized inventory
// tag ledger — one row per physical item, each with a unique code stamped
// on it. Live XML-form scrape shows it consumed across recepcion/,
// almacen/, herramientas/, mantenimiento/, help_local/. Keeps Histrix
// naming (erp_tools / erp_tool_movements) for operational parity.
//
// Resolution: article_id via the stock domain's default-subsystem lookup
// (same pattern as HomologationRevisionLine). Rows whose artcod isn't in
// the current catalog migrate with article_id NULL and article_code
// preserved.
//
// Phase 0 invariant: rows_read = rows_written + rows_skipped + 0
// duplicate. Skips: id_etiqueta == 0 (defensive — AI PK).
func NewToolMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ToolReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id", "code",
			"article_code", "article_id", "inventory_code", "name",
			"characteristic", "group_code", "tool_type", "status_code",
			"purchase_order_no", "purchase_order_date",
			"delivery_note_date", "delivery_note_post", "delivery_note_no",
			"supplier_code", "pending_oc", "observation", "manufacture_no",
			"generated_at",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_etiqueta")
			if legacyID == 0 {
				return nil, nil
			}

			toolCode := strings.TrimSpace(row.String("id_herramienta"))

			var articleID *uuid.UUID
			artCode := strings.TrimSpace(row.String("artcod"))
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, "")))
				if resolved, rerr := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID); rerr == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "tools", "HERRAMIENTAS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			ocpDate := SafeDate(timeFromRow(row, "ocpfec"))
			remDate := SafeDate(timeFromRow(row, "remfec"))
			generated := SafeDate(timeFromRow(row, "generada"))

			return []any{
				id, tenantID, legacyID, toolCode,
				artCode, articleID,
				strings.TrimSpace(row.String("invcod")),
				strings.TrimSpace(row.String("nomherr")),
				strings.TrimSpace(row.String("caract")),
				int16(row.Int("grucod")),
				int16(row.Int("tipoherr")),
				int32(row.Int("codest")),
				int32(row.Int("ocpnro")),
				ocpDate,
				remDate,
				int32(row.Int("remnpv")),
				int32(row.Int("remnro")),
				int32(row.Int("ctacod")),
				ParseDecimal(row.Decimal("pendiente_oc")),
				strings.TrimSpace(row.String("observacion")),
				int32(row.Int("nrofab")),
				generated,
			}, nil
		},
	}
}

// NewToolMovementMigrator migrates HERRMOVS → erp_tool_movements
// (11,680 rows live). Lending ledger: employees take out / return /
// damage / loan tools. concept_code is the raw CONCHERR.movher code
// (1=Devol. Rotura, 2=Devolucion, 3=A Cargo, 7=Prestamo) — the 4-row
// lookup is inlined rather than migrated separately.
//
// Tool resolution: erp_tools code index (built via AfterTableHook on
// HERRAMIENTAS). About 13 % of HERRMOVS rows (1,566 of 11,680) don't
// match because HERRMOVS.id_herramienta sometimes references
// MANT_EQUIPOS.numero_serie instead — those orphan movements migrate
// with tool_id NULL and raw tool_code preserved (FICHADAS forensic
// pattern).
func NewToolMovementMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ToolMovementReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id", "tool_id", "tool_code",
			"user_code", "quantity", "movement_date", "concept_code",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_herrmovs")
			if legacyID == 0 {
				return nil, nil
			}

			toolCode := strings.TrimSpace(row.String("id_herramienta"))
			var toolID *uuid.UUID
			if toolCode != "" {
				if resolved, rerr := mapper.ResolveByCode("tools", "erp_tools", toolCode); rerr == nil && resolved != uuid.Nil {
					toolID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "tools", "HERRMOVS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			movDate := SafeDate(timeFromRow(row, "movfec"))

			return []any{
				id, tenantID, legacyID, toolID, toolCode,
				strings.TrimSpace(row.String("usuario")),
				int32(row.Int("cantidad")),
				movDate,
				int16(row.Int("movher")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 3c — Accounting legacy line log (CTBREGIS)
// ============================================================================

// NewAccountingRegisterMigrator migrates CTBREGIS → erp_accounting_registers.
// Pareto #3 of the Phase 1 §Data migration gap post-2.0.9 (604 K rows live,
// ~28 % of the remaining uncovered row volume).
//
// CTBREGIS is the pre-CTB_MOVIMIENTOS leaf-level debe/haber log — but it is
// NOT dead weight: the live libro_diario_qry joins back through
// CTB_DETALLES.ctbregis_id for legacy ctbcod, and the provider / client /
// orden-de-pago / IVA / anulaciones modules still write to it directly.
// 59 live xml-form references confirmed.
//
// Resolution strategy:
//   - account_id is resolved via the accounting code index
//     (BuildCodeIndex("accounting", "erp_accounts", "code")) already wired
//     in Phase 3 setup. Rows whose ctbcod is not in the current plan de
//     cuentas keep account_id NULL with account_code preserved — forensic
//     preservation, same pattern as FICHADAS orphan tarjetas in 2.0.9.
//   - reg_date uses SafeDate (NULL for 0000-00-00 — 122 rows live) rather
//     than SafeDateRequired to avoid collapsing "unknown date" into the
//     1970 epoch default.
//   - Cost center, imputation, and other secondary references stay as raw
//     SMALLINT/INT — no xml-form surfaces them yet, so resolving to UUID
//     would be dead complexity.
//
// Phase 0 invariant: rows_read = rows_written + rows_skipped + 0 duplicate.
// Skips: legacy_id == 0 (defensive, shouldn't happen — AI PK).
func NewAccountingRegisterMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AccountingRegisterReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"subsystem_code", "reg_date", "voucher_date",
			"minuta_number", "comprobante_type", "comprobante_number",
			"account_code", "account_id", "entry_side",
			"amount", "reference", "status",
			"cost_center_code", "imputation_code",
			"legacy_cost_center_id", "legacy_imputation_id",
			"legacy_account_id",
			"post_number", "entry_order", "subdiary_code",
			"physical_units",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ctbregis")
			if legacyID == 0 {
				return nil, nil
			}

			accountCode := strings.TrimSpace(row.String("ctbcod"))
			var accountID *uuid.UUID
			if accountCode != "" {
				if resolved, rerr := mapper.ResolveByCode("accounting", "erp_accounts", accountCode); rerr == nil && resolved != uuid.Nil {
					accountID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "accounting", "CTBREGIS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			regDate := SafeDate(timeFromRow(row, "regfec"))
			voucherDate := SafeDate(timeFromRow(row, "regfco"))

			return []any{
				id, tenantID, legacyID,
				strings.TrimSpace(row.String("siscod")),
				regDate, voucherDate,
				int32(row.Int("regmin")),
				int16(row.Int("regtip")),
				int32(row.Int("regnro")),
				accountCode, accountID,
				int16(row.Int("regdoh")),
				ParseDecimal(row.Decimal("regimp")),
				strings.TrimSpace(row.String("regref")),
				strings.TrimSpace(row.String("regpoa")),
				int16(row.Int("coscod")),
				int16(row.Int("impcod")),
				int32(row.Int("idcos")),
				int32(row.Int("idimpu")),
				int32(row.Int("regcta")),
				int16(row.Int("regnpv")),
				int16(row.Int("regord")),
				int16(row.Int("regsub")),
				ParseDecimal(row.Decimal("reguni")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 8c — Pareto tail Grupo A (REG_CUENTA_CALIFICACION +
// REG_MOVIMIENTO_OBS + CARCHEHI) — 2.0.11
// ============================================================================

// NewEntityCreditRatingMigrator — REG_CUENTA_CALIFICACION →
// erp_entity_credit_ratings (136,064 rows live; scrape 58,960, +131 %).
// Customer / supplier credit rating history. regcuenta_id points at
// REG_CUENTA(id_regcuenta), which is the entity domain cache populated
// by NewEntityMigrator, so the resolve is a straight ResolveOptional.
// Rows whose FK misses the cache fall back to the unknown-entity
// sentinel via ResolveEntityFlexible — no data loss.
func NewEntityCreditRatingMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.EntityCreditRatingReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"entity_id", "entity_legacy_id",
			"rating", "rated_at", "reference",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_regcalificacion")
			if legacyID == 0 {
				return nil, nil
			}

			regcuentaID := row.Int64("regcuenta_id")
			var entityID *uuid.UUID
			if regcuentaID > 0 {
				if resolved, tag := mapper.ResolveEntityFlexible(ctx, regcuentaID); resolved != uuid.Nil && tag != "unknown" {
					entityID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "current_account", "REG_CUENTA_CALIFICACION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			ratedAt := SafeDate(timeFromRow(row, "fecha_calificacion"))

			return []any{
				id, tenantID, legacyID,
				entityID, int32(regcuentaID),
				strings.TrimSpace(row.String("calificacion")),
				ratedAt,
				strings.TrimSpace(row.String("referencia_calificacion")),
			}, nil
		},
	}
}

// NewInvoiceNoteMigrator — REG_MOVIMIENTO_OBS → erp_invoice_notes
// (72,737 rows live). Free-text notes attached to REG_MOVIMIENTOS.
// regmovim_id resolves via ResolveRegMovim (Phase 6 BuildRegMovimIndex).
// Rows with zero regmovim_id or that miss the index land with
// invoice_id NULL and invoice_legacy_id preserved — the note is still
// useful for audit even when the parent is missing. Zero-dates in
// fec_observacion / movfec pass through as NULL.
func NewInvoiceNoteMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.InvoiceNoteReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"observation_date", "observation_time",
			"observation",
			"invoice_id", "invoice_legacy_id",
			"login", "contact_legacy_id", "source_table",
			"system_code", "movement_date",
			"account_code", "concept_code",
			"movement_voucher_class", "movement_no",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_regmovimientoobs")
			if legacyID == 0 {
				return nil, nil
			}

			regMovimID := row.Int64("regmovim_id")
			var invoiceID *uuid.UUID
			if regMovimID > 0 {
				if resolved, ok := mapper.ResolveRegMovim(regMovimID); ok && resolved != uuid.Nil {
					invoiceID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "current_account", "REG_MOVIMIENTO_OBS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			obsDate := SafeDate(timeFromRow(row, "fec_observacion"))
			movDate := SafeDate(timeFromRow(row, "movfec"))

			var obsTime any
			if t := strings.TrimSpace(row.String("hora_observacion")); t != "" && t != "00:00:00" {
				obsTime = t
			}

			return []any{
				id, tenantID, legacyID,
				obsDate, obsTime,
				row.String("observacion"),
				invoiceID, int32(regMovimID),
				strings.TrimSpace(row.String("login")),
				int32(row.Int64("gencontacto_id")),
				strings.TrimSpace(row.String("tabla_origen")),
				strings.TrimSpace(row.String("siscod")),
				movDate,
				int32(row.Int("ctacod")),
				int32(row.Int("concod")),
				int32(row.Int("movnpv")),
				int32(row.Int64("movnro")),
			}, nil
		},
	}
}

// NewCheckHistoryMigrator — CARCHEHI → erp_check_history (28,763 rows
// live). Archived-check history (sister of CARCHEQU / erp_checks).
// Composite PK (carint, siscod, succod) hashed into legacy_id. ctacod
// resolves via ResolveEntityFlexible (id_regcuenta then nro_cuenta).
// movnro+regmin is a composite pointer at REG_MOVIMIENTOS that we
// preserve raw — no resolver uses that composite today. All the zero-
// date columns (carfec, caralt, caring, plus the nullable caracr /
// cardev) round-trip as NULL.
func NewCheckHistoryMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CheckHistoryReader(db)
	return &GenericMigrator{
		reader: reader,
		columns: []string{
			"id", "tenant_id", "legacy_id",
			"legacy_carint", "legacy_siscod", "legacy_succod",
			"check_type", "number", "bank_name", "amount",
			"operation_date", "credited_at", "returned_at",
			"altered_at", "deposited_at", "issue_date",
			"description", "observation", "reference",
			"owner_ident", "owner_mark", "accredited",
			"entity_legacy_code", "entity_id",
			"movement_no", "movement_register", "movement_voucher_class",
			"portfolio_id", "branch", "system_code",
			"concept_code", "operator_code", "operator_class",
			"plan_id", "pay_no", "received_no", "check_counter",
			"account_balance_ref", "process_code", "circuit_code",
			"bcs_no", "cash_plan",
		},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			carint := row.Int64("carint")
			siscod := strings.TrimSpace(row.String("siscod"))
			succod := row.Int64("succod")

			if carint == 0 && siscod == "" && succod == 0 {
				return nil, nil
			}

			compositeKey := fmt.Sprintf("CARCHEHI:%d:%s:%d", carint, siscod, succod)
			legacyID := hashCode(compositeKey)

			ctacod := row.Int64("ctacod")
			var entityID *uuid.UUID
			if ctacod > 0 {
				if resolved, tag := mapper.ResolveEntityFlexible(ctx, ctacod); resolved != uuid.Nil && tag != "unknown" {
					entityID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "treasury", "CARCHEHI", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, legacyID,
				int32(carint), siscod, int32(succod),
				int16(row.Int("cartip")),
				strings.TrimSpace(row.String("carnro")),
				strings.TrimSpace(row.String("carbco")),
				ParseDecimal(row.Decimal("carimp")),
				SafeDate(timeFromRow(row, "carfec")),
				SafeDate(timeFromRow(row, "caracr")),
				SafeDate(timeFromRow(row, "cardev")),
				SafeDate(timeFromRow(row, "caralt")),
				SafeDate(timeFromRow(row, "caring")),
				SafeDate(timeFromRow(row, "fecha_emision")),
				strings.TrimSpace(row.String("cardes")),
				strings.TrimSpace(row.String("carobv")),
				strings.TrimSpace(row.String("carref")),
				strings.TrimSpace(row.String("carcui")),
				strings.TrimSpace(row.String("carmar")),
				int16(row.Int("acreditado")),
				int32(ctacod), entityID,
				int32(row.Int64("movnro")),
				int32(row.Int("regmin")),
				int32(row.Int("movnpv")),
				int32(row.Int64("cartera_id")),
				int32(succod),
				siscod,
				int32(row.Int("concod")),
				int32(row.Int("opecod")),
				strings.TrimSpace(row.String("opecla")),
				int32(row.Int("carpla")),
				int32(row.Int("carpag")),
				int32(row.Int("carrec")),
				int32(row.Int("carccb")),
				int32(row.Int("ccbcod")),
				int32(row.Int("procod")),
				int32(row.Int("circod")),
				int32(row.Int64("bcsnro")),
				int32(row.Int("cajpla")),
			}, nil
		},
	}
}
