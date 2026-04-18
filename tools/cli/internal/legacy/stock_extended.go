package legacy

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// ---------------------------------------------------------------------------
// BOM — Bill of Materials (STKPIEZA)
// ---------------------------------------------------------------------------

// ArticleCostHistoryReader — STK_COSTO_HIST (103,799 rows live, scrape
// 95,217). Monthly cost snapshots per article. Composite PK
// (articulo_id, anio_hist, mes_hist) — no AI PK. Already consumed as
// JSONB metadata on erp_articles by the metadata enricher's
// articleCostHistory spec; this reader materializes the native
// relational shape so structured history reports can read it without
// parsing JSON. Pareto #8 of the post-2.0.10 gap.
func ArticleCostHistoryReader(db *sql.DB) *CompositeKeyReader {
	return &CompositeKeyReader{
		DB:         db,
		Table:      "STK_COSTO_HIST",
		Target:     "erp_article_cost_history",
		DomainName: "stock",
		PKColumns:  []string{"articulo_id", "anio_hist", "mes_hist"},
		Columns:    "articulo_id, anio_hist, mes_hist, costo_hist, periodo_hist",
	}
}

// ArticleSupplierCostReader creates a reader for STKINSPR (189,863 rows
// live — Pareto #5 of the Phase 1 §Data migration gap post-2.0.10,
// ~17 % of uncovered row volume).
//
// Despite the STKINSPR name (probably "STock INSumo PRecios" historically),
// the table is a per-supplier cost ledger: one row per (artcod, ctacod)
// cost snapshot. XML-form scrape shows it consumed across stock/costos/,
// costos/, estadisticas/evolutivo_costo, evolucion_costos, evolucion_inc_costos,
// remitos/factura_stkinspr_ingmov (invoice-driven insert path). Maintained
// by invoice-import triggers plus periodic re-cost runs (the recalc flag).
//
// Single-subsystem in live data (siscod='01' in 100 % of rows), 7,431
// distinct artcod × 860 distinct ctacod (but only 190 K actual pairs, so
// sparsely populated). fecfac is ~empty (99.9 % zero) and preserved as
// NULLABLE DATE; fecult is the business "as-of" timestamp and is
// populated on 99.99 % of rows.
func ArticleSupplierCostReader(db *sql.DB) *GenericReader {
	return &GenericReader{
		DB:         db,
		Table:      "STKINSPR",
		Target:     "erp_article_costs",
		DomainName: "stock",
		PKColumn:   "idCosto",
		Columns: "idCosto, artcod, siscod, artcos, artpor__1, artpor__2, artpor__3, " +
			"artpro, ctacod, fecfac, fecult, movnro, movnpv, movfec, recalc",
	}
}

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
// JOINs STKPIEZA to bring in idPadre (parent article code) needed for article FK resolution.
type BOMHistoryReaderType struct {
	DB *sql.DB
}

func (r *BOMHistoryReaderType) LegacyTable() string { return "STK_BOM_HIST" }
func (r *BOMHistoryReaderType) SDATable() string     { return "erp_bom_history" }
func (r *BOMHistoryReaderType) Domain() string       { return "stock" }

func (r *BOMHistoryReaderType) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]LegacyRow, string, error) {
	lastID := int64(0)
	if resumeKey != "" {
		lastID, _ = strconv.ParseInt(resumeKey, 10, 64)
	}

	query := `SELECT h.id_stkbomhist, h.pieza_id, h.bom_variacion_id, h.fecha_costo,
		h.stkarticulohijo_id, h.stkarticulohijo_string, h.posicionfab_id,
		h.tipocalculo_costo, h.regcuenta_id, h.cuenta_id, h.cantidad,
		h.unidad_compra, h.unidad_uso, h.multiplo, h.unitario, h.recargo,
		h.porcentaje_region, h.sumaitems, h.sumaitems_clog,
		h.level_0, h.level_1, h.level_2, h.level_3, h.level_4, h.level_5, h.level_6, h.level_7, h.piezas,
		p.idPadre AS parent_article_code
		FROM STK_BOM_HIST h
		LEFT JOIN STKPIEZA p ON h.pieza_id = p.id_pieza
		WHERE h.id_stkbomhist > ?
		ORDER BY h.id_stkbomhist LIMIT ?`

	rows, err := r.DB.QueryContext(ctx, query, lastID, limit)
	if err != nil {
		return nil, "", fmt.Errorf("read STK_BOM_HIST: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, "", fmt.Errorf("columns STK_BOM_HIST: %w", err)
	}

	var result []LegacyRow
	var lastKey int64
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, "", fmt.Errorf("scan STK_BOM_HIST: %w", err)
		}
		row := make(LegacyRow, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
		lastKey = row.Int64("id_stkbomhist")
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate STK_BOM_HIST: %w", err)
	}

	return result, strconv.FormatInt(lastKey, 10), nil
}

// BOMHistoryReader creates a reader for STK_BOM_HIST with a STKPIEZA JOIN.
func BOMHistoryReader(db *sql.DB) *BOMHistoryReaderType {
	return &BOMHistoryReaderType{DB: db}
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
