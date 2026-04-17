package migration

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/google/uuid"

	"github.com/Camionerou/rag-saldivia/tools/cli/internal/legacy"
)

// GenericMigrator implements TableMigrator using a legacy.Reader and a transform function.
type GenericMigrator struct {
	reader         legacy.Reader
	columns        []string
	conflictCol    string
	transformFn    func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error)
}

func (m *GenericMigrator) LegacyTable() string    { return m.reader.LegacyTable() }
func (m *GenericMigrator) SDATable() string        { return m.reader.SDATable() }
func (m *GenericMigrator) Domain() string          { return m.reader.Domain() }
func (m *GenericMigrator) Columns() []string       { return m.columns }
func (m *GenericMigrator) ConflictColumn() string  { return m.conflictCol }
func (m *GenericMigrator) Reader() interface {
	ReadBatch(ctx context.Context, resumeKey string, limit int) ([]map[string]any, string, error)
} {
	return &readerAdapter{r: m.reader}
}

func (m *GenericMigrator) Transform(ctx context.Context, row map[string]any, mapper *Mapper) ([]any, error) {
	return m.transformFn(ctx, legacy.LegacyRow(row), mapper)
}

// readerAdapter adapts legacy.Reader to the interface expected by orchestrator.
type readerAdapter struct {
	r legacy.Reader
}

func (a *readerAdapter) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]map[string]any, string, error) {
	rows, key, err := a.r.ReadBatch(ctx, resumeKey, limit)
	if err != nil {
		return nil, "", err
	}
	result := make([]map[string]any, len(rows))
	for i, r := range rows {
		result[i] = map[string]any(r)
	}
	return result, key, nil
}

// --- Catalog Migrators ---

func NewCatalogMigrator(db *sql.DB, cm legacy.CatalogMapping, tenantID string) *GenericMigrator {
	reader := legacy.CatalogReader(db, cm)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "type", "code", "name", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64(cm.PKColumn)
			if legacyID == 0 {
				return nil, nil // skip
			}
			id, err := mapper.Map(ctx, nil, "catalog", cm.LegacyTable, legacyID, nil)
			if err != nil {
				// In dry-run mapper.Map with nil tx won't work; generate UUID
				id = uuid.New()
			}
			code := row.String(cm.CodeColumn)
			name := row.String(cm.NameColumn)
			if name == "" {
				name = code
			}
			return []any{id, tenantID, cm.CatalogType, fmt.Sprintf("%v", code), name, true}, nil
		},
	}
}

// --- Entity Migrators ---

func NewEntityMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.EntityReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "type", "code", "name", "email", "phone", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_regcuenta")
			if legacyID == 0 {
				return nil, nil
			}
			// subsistema_id: 01=compras(supplier), 02=ventas(customer)
			sub := row.String("subsistema_id")
			entityType := "customer"
			if sub == "01" || sub == "1" {
				entityType = "supplier"
			}

			name := row.String("nombre_cuenta")
			if razon := row.String("razon_social"); razon != "" {
				name = razon
			}

			baja := row.Int("baja")
			active := baja == 0

			code := fmt.Sprintf("%d", row.Int64("nro_cuenta"))
			if code == "0" {
				code = fmt.Sprintf("REG-%d", legacyID)
			}

			creator := row.NullString("creator_user_id")
			id, err := mapper.Map(ctx, nil, "entity", "REG_CUENTA", legacyID, creator)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, entityType, code, name,
				row.NullString("email_cuenta"),
				row.NullString("tel_cuenta"),
				active,
			}, nil
		},
	}
}

func NewEmployeeMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.EmployeeReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "type", "code", "name", "email", "phone", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("IdPersona")
			if legacyID == 0 {
				return nil, nil
			}

			nombre := row.String("Nombre")
			apellido := row.String("apellido")
			name := apellido + ", " + nombre
			if apellido == "" {
				name = nombre
			}

			legajo := row.Int("legajo")
			code := fmt.Sprintf("EMP-%d", legajo)
			if legajo == 0 {
				code = fmt.Sprintf("EMP-%d", legacyID)
			}

			egreso := row.String("fechaEgreso")
			active := egreso == "" || egreso == "0000-00-00"

			id, err := mapper.Map(ctx, nil, "entity", "PERSONAL", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, "employee", code, name,
				row.NullString("email"),
				row.NullString("telefono"),
				active,
			}, nil
		},
	}
}

// --- Accounting Migrators ---

func NewCostCenterMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CostCenterReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ctbcentro")
			if legacyID == 0 {
				return nil, nil
			}
			id, err := mapper.Map(ctx, nil, "accounting", "CTB_CENTROS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID,
				row.String("referencia_centro"),
				row.String("nombre_centro"),
				true,
			}, nil
		},
	}
}

func NewAccountMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AccountReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "account_type", "is_detail", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ctbcuenta")
			if legacyID == 0 {
				return nil, nil
			}
			// tipo_id: 1=activo, 2=pasivo, 3=patrimonio, 4=ingreso, 5=egreso
			tipoID := row.Int("tipo_id")
			accountType := "asset"
			switch tipoID {
			case 2:
				accountType = "liability"
			case 3:
				accountType = "equity"
			case 4:
				accountType = "income"
			case 5:
				accountType = "expense"
			}

			habilitada := row.Int("habilitada")

			id, err := mapper.Map(ctx, nil, "accounting", "CTB_CUENTAS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				row.String("id_cuenta"),
				row.String("nombre_cuenta"),
				accountType,
				true, // is_detail (will be updated post-migration based on children)
				habilitada != 0,
			}, nil
		},
	}
}

func NewFiscalYearMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.FiscalYearReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "year", "start_date", "end_date", "status"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ejercicio")
			if legacyID == 0 {
				return nil, nil
			}

			// Histrix `cerrado=1` does not carry the result_account_id the new
			// fiscal-close workflow requires. Always import as 'open' — the
			// 10 legacy-closed ejercicios can be formally closed via the UI
			// after assigning their result account.
			_ = row.Int("cerrado")
			status := "open"

			// Parse year from nombre_ejercicio or comienza date
			year := 0
			if nombre := row.String("nombre_ejercicio"); nombre != "" {
				_, _ = fmt.Sscanf(nombre, "%d", &year)
			}
			if year == 0 {
				// Try to extract from start date
				if t, ok := row["comienza"].(time.Time); ok && !t.IsZero() {
					year = t.Year()
				}
			}

			id, err := mapper.Map(ctx, nil, "accounting", "CTB_EJERCICIOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			comienza := SafeDateRequired(timeFromRow(row, "comienza"))
			finaliza := SafeDateRequired(timeFromRow(row, "finaliza"))

			return []any{id, tenantID, year, comienza, finaliza, status}, nil
		},
	}
}

func NewJournalEntryMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.JournalEntryReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "fiscal_year_id", "concept", "entry_type", "user_id", "status"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_movimiento")
			if legacyID == 0 {
				return nil, nil
			}

			// Map fiscal year FK
			fyID := row.Int64("ejercicio_id")
			var fiscalYearID *uuid.UUID
			if fyID > 0 {
				resolved, err := mapper.Resolve(ctx, "accounting", "CTB_EJERCICIOS", fyID)
				if err == nil {
					fiscalYearID = &resolved
				}
			}

			// Use id_movimiento (the PK) to guarantee uniqueness. nro_minuta
			// repeats across Histrix fiscal years and would drop ~200K rows
			// to the tenant_id+number UNIQUE constraint.
			number := fmt.Sprintf("AS-%d-%d", row.Int64("nro_minuta"), legacyID)
			date := SafeDateRequired(timeFromRow(row, "fecha_movimiento"))

			usuario := row.NullString("usuario_modificacion")
			id, err := mapper.Map(ctx, nil, "accounting", "CTB_MOVIMIENTOS", legacyID, usuario)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, number, date, fiscalYearID,
				row.String("referencia"),
				"manual", // default — Histrix doesn't distinguish entry types in this table
				LegacyUserID,
				"posted", // all legacy entries are posted
			}, nil
		},
	}
}

func NewJournalLineMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.JournalLineReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entry_id", "account_id", "entry_date", "debit", "credit", "description", "sort_order"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_detalle")
			if legacyID == 0 {
				return nil, nil
			}

			// Resolve entry FK
			entryLegacyID := row.Int64("movimiento_id")
			entryID, err := mapper.Resolve(ctx, "accounting", "CTB_MOVIMIENTOS", entryLegacyID)
			if err != nil {
				return nil, nil // skip orphan lines (parent entry not migrated)
			}

			// Resolve account FK via code index (built after account migration)
			accountCode := row.String("cuenta_id")
			accountID, err := mapper.ResolveByCode("accounting", "erp_accounts", accountCode)
			if err != nil {
				return nil, nil // skip lines with unknown account codes
			}

			// doh: 0=debe, 1=haber. Legacy can have negative amounts.
			doh := row.Int("doh")
			amount := ParseDecimal(row.Decimal("importe")).Abs()

			debit := amount
			credit := ParseDecimal("0")
			if doh == 1 {
				debit = ParseDecimal("0")
				credit = amount
			}

			id, err := mapper.Map(ctx, nil, "accounting", "CTB_DETALLES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Resolve entry date from the already-migrated journal entry
			entryDate := mapper.GetEntryDate(entryID)
			if entryDate.IsZero() {
				entryDate = SafeDateRequired(time.Time{})
			}

			return []any{
				id, tenantID, entryID, accountID,
				entryDate,
				debit, credit,
				row.String("referencia"),
				row.Int("orden"),
			}, nil
		},
	}
}

// --- Treasury Migrators ---

func NewBankAccountMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.BankAccountReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "bank_name", "account_number", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_carbanco")
			if legacyID == 0 {
				return nil, nil
			}
			id, err := mapper.Map(ctx, nil, "treasury", "CAR_BANCOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID,
				row.String("nombre_banco"),
				fmt.Sprintf("BANK-%d", legacyID), // no account_number in legacy
				true,
			}, nil
		},
	}
}

func NewCashRegisterMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CashRegisterReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "name", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_cajpuesto")
			if legacyID == 0 {
				return nil, nil
			}
			id, err := mapper.Map(ctx, nil, "treasury", "CAJ_PUESTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID,
				row.String("descripcion_puesto"),
				true,
			}, nil
		},
	}
}

func NewCheckMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CheckReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "direction", "number", "bank_name", "amount", "issue_date", "due_date", "status", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("carint")
			if legacyID == 0 {
				return nil, nil
			}

			// cartip: 0=received, 1=issued (conventional)
			direction := "received"
			if row.Int("cartip") == 1 {
				direction = "issued"
			}

			amount := ParseDecimal(row.Decimal("carimp"))
			if amount.IsZero() {
				return nil, nil // skip zero-amount checks
			}

			id, err := mapper.Map(ctx, nil, "treasury", "CARCHEQU", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, direction,
				row.String("carnro"),
				row.String("carbco"),
				amount,
				SafeDateRequired(timeFromRow(row, "fecha_emision")),
				SafeDateRequired(timeFromRow(row, "caracr")),
				"in_portfolio",
				row.String("carobv"),
			}, nil
		},
	}
}

// --- Stock Migrators ---

func NewWarehouseMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.WarehouseReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "location", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_stkdeposito")
			if legacyID == 0 {
				return nil, nil
			}
			id, err := mapper.Map(ctx, nil, "stock", "STK_DEPOSITOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			inactive := row.Int("inactivo")
			return []any{
				id, tenantID,
				fmt.Sprintf("DEP-%d", legacyID),
				row.String("nombre_deposito"),
				row.String("direccion_deposito"),
				inactive == 0,
			}, nil
		},
	}
}

func NewArticleMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ArticleReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "article_type", "min_stock", "max_stock", "reorder_point", "last_cost", "avg_cost", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			rawCode := row.String("id_stkarticulo")
			if rawCode == "" {
				return nil, nil
			}
			// STK_ARTICULOS PK is composite (id_stkarticulo, subsistema_id).
			// Same id may appear under multiple subsystems — preserve them as
			// distinct articles by namespacing the code and hash for any subsystem
			// other than "01" (the dominant one).
			subsistema := row.String("subsistema_id")
			code := articleCompositeCode(rawCode, subsistema)
			legacyID := int64(hashCode(code))

			baja := row.Int("baja_articulo")
			active := baja == 0

			// tipo_articulo mapping
			artType := "material"
			switch row.Int("tipo_articulo") {
			case 1:
				artType = "product"
			case 2:
				artType = "material"
			case 3:
				artType = "consumable"
			}

			id, err := mapper.Map(ctx, nil, "stock", "STK_ARTICULOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, code,
				row.String("nombre_articulo"),
				artType,
				ParseDecimal(row.Decimal("stock_minimo")),
				ParseDecimal(row.Decimal("stock_maximo")),
				ParseDecimal(row.Decimal("stock_reposicion")),
				ParseDecimal(row.Decimal("precio_costo")),
				ParseDecimal(row.Decimal("precio_promedio")),
				active,
			}, nil
		},
	}
}

func NewStockMovementMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.StockMovementReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "article_id", "warehouse_id", "movement_type", "quantity", "unit_cost", "user_id", "notes", "created_at"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_stkmovimiento")
			if legacyID == 0 {
				return nil, nil
			}

			// Resolve article FK by code hash
			artCode := row.String("stkarticulo_id")
			artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
			articleID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", artLegacyID)
			if err != nil {
				return nil, nil // skip if article not found
			}

			// Warehouse FK — fall back to the UNASSIGNED warehouse when the legacy
			// row has stkdeposito_id = 0 (allowed in Histrix) or points to a depot
			// that no longer exists. Prefix notes with [legacy:unassigned_warehouse]
			// so the rows can be filtered and reassigned from an admin UI later.
			warehouseLegacyID := row.Int64("stkdeposito_id")
			unassigned := false
			var warehouseID uuid.UUID
			if warehouseLegacyID > 0 {
				resolved, err := mapper.Resolve(ctx, "stock", "STK_DEPOSITOS", warehouseLegacyID)
				if err != nil {
					fallbackID := mapper.UnassignedWarehouseID()
					if fallbackID == uuid.Nil {
						return nil, nil // fallback not initialised — skip row
					}
					warehouseID = fallbackID
					unassigned = true
				} else {
					warehouseID = resolved
				}
			} else {
				fallbackID := mapper.UnassignedWarehouseID()
				if fallbackID == uuid.Nil {
					return nil, nil
				}
				warehouseID = fallbackID
				unassigned = true
			}

			// Movement type — positive quantity = in, negative = out.
			// Zero-quantity legacy rows (173K on saldivia) represent adjustment
			// events without a numeric delta: depot reclassification, corrections,
			// stock counts that confirmed 0 movement. Post-migration 061 relaxed
			// the CHECK from (> 0) to (>= 0) so we can preserve them — flagged in
			// notes for operator review.
			quantity := ParseDecimal(row.Decimal("cantidad"))
			movType := "in"
			if quantity.IsNegative() {
				movType = "out"
				quantity = quantity.Abs()
			}
			zeroQty := quantity.IsZero()

			id, err := mapper.Map(ctx, nil, "stock", "STK_MOVIMIENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			notes := row.String("referencia")
			if unassigned {
				if notes != "" {
					notes = "[legacy:unassigned_warehouse] " + notes
				} else {
					notes = "[legacy:unassigned_warehouse]"
				}
			}
			if zeroQty {
				if notes != "" {
					notes = "[legacy:zero_qty] " + notes
				} else {
					notes = "[legacy:zero_qty]"
				}
			}

			return []any{
				id, tenantID, articleID, warehouseID, movType, quantity,
				ParseDecimal(row.Decimal("precio_costo")),
				LegacyUserID,
				notes,
				SafeDateRequired(timeFromRow(row, "fecha_movimiento")),
			}, nil
		},
	}
}

// --- Purchasing Migrators ---

func NewPurchaseOrderMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PurchaseOrderReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "supplier_id", "status", "total", "notes", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_cpsmovimiento")
			if legacyID == 0 {
				return nil, nil
			}

			supplierLegacyID := row.Int64("regcuenta_id")
			supplierID, _ := mapper.ResolveEntityFlexible(ctx, supplierLegacyID)
			if supplierID == uuid.Nil {
				return nil, nil // UNKNOWN hook missing — archive-skips will keep it
			}

			// estado: ''=draft, 1=approved, 2=received
			estado := row.String("cpsestado_id")
			status := "draft"
			switch estado {
			case "1":
				status = "approved"
			case "2":
				status = "received"
			}

			number := fmt.Sprintf("OC-%d", row.Int64("cps_numero"))
			if row.Int64("cps_numero") == 0 {
				number = fmt.Sprintf("OC-%d", legacyID)
			}

			id, err := mapper.Map(ctx, nil, "purchasing", "CPS_MOVIMIENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, number,
				SafeDateRequired(timeFromRow(row, "cpsfecha")),
				supplierID, status,
				ParseDecimal(row.Decimal("cpsimporte")),
				row.String("cpsobservacion"),
				LegacyUserID,
			}, nil
		},
	}
}

func NewPurchaseOrderLineMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PurchaseOrderLineReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "order_id", "article_id", "quantity", "unit_price", "received_qty"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_cpsdetalle")
			if legacyID == 0 {
				return nil, nil
			}

			orderLegacyID := row.Int64("cpsmovimiento_id")
			orderID, err := mapper.Resolve(ctx, "purchasing", "CPS_MOVIMIENTOS", orderLegacyID)
			if err != nil {
				return nil, nil
			}

			artCode := row.String("stkarticulo_id")
			artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
			articleID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", artLegacyID)
			if err != nil {
				return nil, nil
			}

			quantity := ParseDecimal(row.Decimal("cantidad_compra"))
			if quantity.IsZero() || quantity.IsNegative() {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "purchasing", "CPS_DETALLE", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, orderID, articleID, quantity,
				ParseDecimal(row.Decimal("costo_unitario")),
				ParseDecimal(row.Decimal("recibido_compra")),
			}, nil
		},
	}
}

// --- Sales Migrators ---

func NewQuotationMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.QuotationReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "customer_id", "status", "total", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("idCotiz")
			if legacyID == 0 {
				return nil, nil
			}

			customerLegacyID := row.Int64("ctacod")
			customerID, _ := mapper.ResolveEntityFlexible(ctx, customerLegacyID)
			if customerID == uuid.Nil {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "sales", "COTIZACION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				fmt.Sprintf("COT-%d", legacyID),
				SafeDateRequired(timeFromRow(row, "fecha")),
				customerID,
				"approved", // legacy quotations are finalized
				ParseDecimal(row.Decimal("total")),
				LegacyUserID,
			}, nil
		},
	}
}

// --- Production Migrators ---

func NewProductionCenterMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductionCenterReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "code", "name", "active"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_centro_productivo")
			if legacyID == 0 {
				return nil, nil
			}
			id, err := mapper.Map(ctx, nil, "production", "MRP_CENTRO_PRODUCTIVO", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}
			return []any{
				id, tenantID,
				fmt.Sprintf("CP-%d", legacyID),
				row.String("nombre_centroproductivo"),
				true,
			}, nil
		},
	}
}

func NewProductionOrderMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.ProductionOrderReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "product_id", "status", "quantity", "user_id", "notes"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_mrporden")
			if legacyID == 0 {
				return nil, nil
			}

			// mrpestado_id mapping
			status := "planned"
			switch row.Int("mrpestado_id") {
			case 1:
				status = "in_progress"
			case 2:
				status = "completed"
			case 3:
				status = "cancelled"
			}

			// Resolve product_id from first article in MRP_ORDEN_DETALLE (via subquery in reader)
			artCode := row.String("first_article_code")
			if artCode == "" {
				return nil, nil // skip production orders with no articles
			}
			artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
			productID, err := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID)
			if err != nil || productID == uuid.Nil {
				return nil, nil // skip if article not migrated
			}

			quantity := ParseDecimal(row.Decimal("first_quantity"))
			if quantity.IsZero() {
				quantity = ParseDecimal("1")
			}

			id, err := mapper.Map(ctx, nil, "production", "MRP_ORDEN_PRODUCCION", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				fmt.Sprintf("OP-%d", legacyID),
				SafeDateRequired(timeFromRow(row, "fecha_orden")),
				productID,
				status,
				quantity,
				LegacyUserID,
				row.String("orden_comentarios"),
			}, nil
		},
	}
}

// --- HR Migrators ---

func NewEmployeeDetailMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.EmployeeDetailReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "position", "hire_date", "termination_date", "schedule_type"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("IdPersona")
			if legacyID == 0 {
				return nil, nil
			}

			entityID, err := mapper.Resolve(ctx, "entity", "PERSONAL", legacyID)
			if err != nil {
				return nil, nil // skip if entity not migrated
			}

			// Map PERSONAL→EmployeeDetail idempotently (separate domain so it does
			// not collide with the entity UUID for the same IdPersona).
			id, err := mapper.Map(ctx, nil, "hr", "PERSONAL_DETAIL", legacyID, nil)
			if err != nil {
				return nil, fmt.Errorf("map PERSONAL_DETAIL %d: %w", legacyID, err)
			}

			// Schedule type
			scheduleType := "full_time"
			horario := row.Int("horario")
			if horario == 2 {
				scheduleType = "part_time"
			} else if horario == 3 {
				scheduleType = "shifts"
			}

			return []any{
				id, tenantID, entityID,
				"", // position — no direct mapping from PERSONAL
				SafeDate(timeFromRow(row, "fecing")),
				SafeDate(timeFromRow(row, "fechaEgreso")),
				scheduleType,
			}, nil
		},
	}
}

func NewAbsenceMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AbsenceReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "entity_id", "event_type", "date_from", "date_to", "notes", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id")
			if legacyID == 0 {
				return nil, nil
			}

			personaID := row.Int64("IdPersona")
			entityID, err := mapper.Resolve(ctx, "entity", "PERSONAL", personaID)
			if err != nil {
				return nil, nil
			}

			id, err := mapper.Map(ctx, nil, "hr", "FRANCOS_PER", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, entityID,
				"absence", // default event type
				SafeDateRequired(timeFromRow(row, "fechainicio")),
				SafeDate(timeFromRow(row, "fechafin")),
				row.String("observaciones"),
				LegacyUserID,
			}, nil
		},
	}
}

func NewTrainingMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.TrainingReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "name", "description", "status"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("Id_curso")
			if legacyID == 0 {
				return nil, nil
			}

			status := "completed" // legacy courses are historical
			if row.Int("situacion") == 0 {
				status = "planned"
			}

			id, err := mapper.Map(ctx, nil, "hr", "RH_CURSOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				row.String("descripcion"),
				row.String("contenido"),
				status,
			}, nil
		},
	}
}

// --- Invoicing Migrators (Phase 6) ---

// NewSalesInvoiceMigrator migrates IVAVENTAS (sales invoice headers, 9K rows) → erp_invoices.
// IVAVENTAS is the real invoice master table, NOT FACREMIT (which is delivery notes).
// Column mapping: id_ivaventa=PK, codcom+codlet→invoice_type, ctacod→entity,
// feciva→date, totcom→total, nronpv+nrocom→number.
func NewSalesInvoiceMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.SalesVATReader(db)
	// Override the target: SalesVATReader targets erp_tax_entries, but for invoice headers
	// we wrap it to target erp_invoices.
	return &GenericMigrator{
		reader:      &readerOverride{r: reader, target: "erp_invoices"},
		columns:     []string{"id", "tenant_id", "number", "date", "due_date", "invoice_type", "direction", "entity_id", "currency_id", "subtotal", "tax_amount", "total", "afip_cae", "afip_cae_due", "status", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ivaventa")
			if legacyID == 0 {
				return nil, nil
			}

			codcom := row.Int("codcom")
			codlet := row.String("codlet")
			invoiceType := MapInvoiceType(codcom, codlet)

			// Entity FK: ctacod → REG_CUENTA. ctacod is ambiguous — use the
			// flex resolver with UNKNOWN fallback instead of dropping rows.
			entityLegacyID := row.Int64("ctacod")
			resolvedEnt, _ := mapper.ResolveEntityFlexible(ctx, entityLegacyID)
			var entityID *uuid.UUID
			if resolvedEnt != uuid.Nil {
				entityID = &resolvedEnt
			}
			if entityID == nil {
				return nil, nil // fallback missing — let it drop cleanly
			}

			// Number: LETTER-PV-NUMERO (e.g., A-0001-00012345)
			number := fmt.Sprintf("%s-%04d-%08d", codlet, row.Int64("nronpv"), row.Int64("nrocom"))

			id, err := mapper.Map(ctx, nil, "invoicing", "IVAVENTAS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				number,
				SafeDateRequired(timeFromRow(row, "feciva")),
				(*time.Time)(nil),     // no due date in IVAVENTAS
				invoiceType,
				"issued",              // fixed: IVAVENTAS = sales = issued
				entityID,
				(*uuid.UUID)(nil),     // no currency FK in IVAVENTAS
				ParseDecimal(row.Decimal("totcom")), // subtotal ≈ total (tax detail in IVAIMPORTES)
				ParseDecimal("0"),     // real tax comes from IVAIMPORTES
				ParseDecimal(row.Decimal("totcom")),
				(*string)(nil),        // no CAE in IVAVENTAS
				(*time.Time)(nil),     // no CAE due
				"posted",              // all IVA ledger entries are posted
				LegacyUserID,
			}, nil
		},
	}
}

// NewPurchaseInvoiceMigrator migrates IVACOMPRAS (purchase invoice headers, 124K rows) → erp_invoices.
// Same structure as IVAVENTAS but direction = "received".
func NewPurchaseInvoiceMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PurchaseVATReader(db)
	return &GenericMigrator{
		reader:      &readerOverride{r: reader, target: "erp_invoices"},
		columns:     []string{"id", "tenant_id", "number", "date", "due_date", "invoice_type", "direction", "entity_id", "currency_id", "subtotal", "tax_amount", "total", "afip_cae", "afip_cae_due", "status", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ivacompra")
			if legacyID == 0 {
				return nil, nil
			}

			codcom := row.Int("codcom")
			codlet := row.String("codlet")
			invoiceType := MapInvoiceType(codcom, codlet)

			// Entity FK: ctacod → REG_CUENTA (flex + UNKNOWN fallback, same as sales).
			entityLegacyID := row.Int64("ctacod")
			resolvedEnt, _ := mapper.ResolveEntityFlexible(ctx, entityLegacyID)
			var entityID *uuid.UUID
			if resolvedEnt != uuid.Nil {
				entityID = &resolvedEnt
			}
			if entityID == nil {
				return nil, nil
			}

			number := fmt.Sprintf("%s-%04d-%08d", codlet, row.Int64("nronpv"), row.Int64("nrocom"))

			id, err := mapper.Map(ctx, nil, "invoicing", "IVACOMPRAS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				number,
				SafeDateRequired(timeFromRow(row, "feciva")),
				(*time.Time)(nil),
				invoiceType,
				"received",            // fixed: IVACOMPRAS = purchases = received
				entityID,
				(*uuid.UUID)(nil),
				ParseDecimal(row.Decimal("totcom")),
				ParseDecimal("0"),
				ParseDecimal(row.Decimal("totcom")),
				(*string)(nil),
				(*time.Time)(nil),
				"posted",
				LegacyUserID,
			}, nil
		},
	}
}

// NewInvoiceLineMigrator migrates FACDETAL (invoice lines, 194K rows) → erp_invoice_lines.
// FK resolution: FACDETAL.regmovim_id links to IVAVENTAS/IVACOMPRAS.regmovim_id.
// The regmovim→invoice index must be built before this migrator runs (setup hook).
func NewInvoiceLineMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.InvoiceLineReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "invoice_id", "article_id", "description", "quantity", "unit_price", "tax_rate", "tax_amount", "line_total", "sort_order"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_detal")
			if legacyID == 0 {
				return nil, nil
			}

			// Resolve invoice FK via regmovim_id index
			regMovimID := row.Int64("regmovim_id")
			var invoiceID *uuid.UUID
			if regMovimID > 0 {
				resolved, ok := mapper.ResolveRegMovim(regMovimID)
				if ok {
					invoiceID = &resolved
				}
			}
			if invoiceID == nil {
				return nil, nil // skip orphan lines with no linked invoice
			}

			// Resolve article FK (optional — via code hash, same as stock)
			var articleID *uuid.UUID
			artCode := row.String("artcod")
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
				resolved, err := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID)
				if err == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			// Unit price: artnet is the net amount. If artcan > 1, divide to get per-unit.
			quantity := row.Int64("artcan")
			if quantity <= 0 {
				quantity = 1
			}
			artNet := ParseDecimal(row.Decimal("artnet"))
			unitPrice := artNet
			if quantity > 1 {
				unitPrice = artNet.Div(ParseDecimal(fmt.Sprintf("%d", quantity)))
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "FACDETAL", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, *invoiceID, articleID,
				row.String("detalle"),
				ParseDecimal(fmt.Sprintf("%d", quantity)),
				unitPrice,
				ParseDecimal(row.Decimal("artali")),
				ParseDecimal(row.Decimal("artiva")),
				ParseDecimal(row.Decimal("arttot")),
				0, // no ordering column in FACDETAL
			}, nil
		},
	}
}

// NewDeliveryNoteMigrator migrates FACREMIT (delivery notes, 4K rows) → erp_invoices.
// FACREMIT is NOT an invoice table — it's delivery notes (remitos).
// Composite PK: (ctacod, remfec, remnpv, remnro). We hash the composite key for legacy_id.
func NewDeliveryNoteMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.InvoiceReader(db) // InvoiceReader reads FACREMIT
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "due_date", "invoice_type", "direction", "entity_id", "currency_id", "subtotal", "tax_amount", "total", "afip_cae", "afip_cae_due", "status", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite PK → hash for legacy_id
			ctacod := row.Int64("ctacod")
			remnpv := row.Int64("remnpv")
			remnro := row.Int64("remnro")
			compositeKey := fmt.Sprintf("%d:%s:%d:%d", ctacod, row.String("remfec"), remnpv, remnro)
			legacyID := int64(hashCode(compositeKey))
			if legacyID == 0 {
				legacyID = 1
			}

			// Entity FK with flex + UNKNOWN fallback so NOT NULL never fails.
			resolvedEnt, _ := mapper.ResolveEntityFlexible(ctx, ctacod)
			entityID := &resolvedEnt
			if resolvedEnt == uuid.Nil {
				entityID = nil
			}
			if entityID == nil {
				return nil, nil // UNKNOWN hook missing — archive-skips will catch it
			}

			number := fmt.Sprintf("REM-%d-%d", remnpv, remnro)

			id, err := mapper.Map(ctx, nil, "invoicing", "FACREMIT", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				number,
				SafeDateRequired(timeFromRow(row, "remfec")),
				(*time.Time)(nil),     // no due date
				"delivery_note",       // fixed invoice_type
				"issued",              // default direction
				entityID,
				(*uuid.UUID)(nil),     // no currency
				ParseDecimal("0"),     // no subtotal
				ParseDecimal("0"),     // no tax
				ParseDecimal(row.Decimal("remval")),
				(*string)(nil),        // no CAE
				(*time.Time)(nil),     // no CAE due
				"posted",
				LegacyUserID,
			}, nil
		},
	}
}

// NewDeliveryNoteAltMigrator migrates REMITO (alt delivery notes, 3.8K rows) → erp_invoices.
// Composite PK: (numero, puesto). Links to entity via ctacod.
func NewDeliveryNoteAltMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.DeliveryNoteReader(db) // reads REMITO
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "number", "date", "due_date", "invoice_type", "direction", "entity_id", "currency_id", "subtotal", "tax_amount", "total", "afip_cae", "afip_cae_due", "status", "user_id"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			numero := row.Int64("numero")
			puesto := row.Int64("puesto")
			compositeKey := fmt.Sprintf("REMITO:%d:%d", numero, puesto)
			legacyID := int64(hashCode(compositeKey))
			if legacyID == 0 {
				legacyID = 1
			}

			// Entity FK with flex + UNKNOWN fallback.
			ctacod := row.Int64("ctacod")
			resolvedEnt, _ := mapper.ResolveEntityFlexible(ctx, ctacod)
			entityID := &resolvedEnt
			if resolvedEnt == uuid.Nil {
				entityID = nil
			}
			if entityID == nil {
				return nil, nil
			}

			number := fmt.Sprintf("REMITO-%d-%d", puesto, numero)

			id, err := mapper.Map(ctx, nil, "invoicing", "REMITO", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				number,
				SafeDateRequired(timeFromRow(row, "fecha")),
				(*time.Time)(nil),
				"delivery_note",
				"issued",
				entityID,
				(*uuid.UUID)(nil),
				ParseDecimal("0"),
				ParseDecimal("0"),
				ParseDecimal("0"), // REMITO has no total column
				(*string)(nil),
				(*time.Time)(nil),
				"posted",
				LegacyUserID,
			}, nil
		},
	}
}

// NewDeliveryNoteLineMigrator migrates REMDETAL (delivery note lines, 5K rows) → erp_invoice_lines.
// PK: idRemdet. Links to REMITO via idRemito.
func NewDeliveryNoteLineMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.DeliveryNoteLineReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "invoice_id", "article_id", "description", "quantity", "unit_price", "tax_rate", "tax_amount", "line_total", "sort_order"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("idRemdet")
			if legacyID == 0 {
				return nil, nil
			}

			// idRemito is the physical auto-inc PK in REMITO; REMITO's logical PK is
			// composite (numero, puesto) and is hashed as legacy_id by the REMITO
			// migrator. BuildRemitoIndex resolves idRemito → UUID via that hash.
			remitoID := row.Int64("idRemito")
			var invoiceID *uuid.UUID
			if resolved, ok := mapper.ResolveByRemitoID(remitoID); ok {
				invoiceID = &resolved
			}
			if invoiceID == nil {
				return nil, nil // skip orphan lines
			}

			// Article FK (optional)
			var articleID *uuid.UUID
			artCode := row.String("artcod")
			if artCode != "" {
				artLegacyID := int64(hashCode(articleCompositeCode(artCode, row.String("subsistema_id"))))
				resolved, err := mapper.ResolveOptional(ctx, "stock", "STK_ARTICULOS", artLegacyID)
				if err == nil && resolved != uuid.Nil {
					articleID = &resolved
				}
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "REMDETAL", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			qty := ParseDecimal(row.Decimal("cantCompra"))
			if qty.IsZero() {
				qty = ParseDecimal("1")
			}

			return []any{
				id, tenantID, *invoiceID, articleID,
				row.String("artcod"), // no description column — use article code
				qty,
				ParseDecimal("0"), // no unit price in REMDETAL
				ParseDecimal("0"), // no tax rate
				ParseDecimal("0"), // no tax amount
				ParseDecimal("0"), // no line total
				0,                 // no ordering
			}, nil
		},
	}
}

// --- Tax Entry Migrators (Phase 7: IVAIMPORTES) ---

// NewTaxEntrySalesMigrator migrates IVAIMPORTES rows for sales (tipoiva=1) → erp_tax_entries.
// IVAIMPORTES has one row per tax-rate per invoice. ivacv_id links to IVAVENTAS.id_ivaventa.
func NewTaxEntrySalesMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.VATAmountReader(db)
	return &GenericMigrator{
		reader:      &readerOverride{r: reader, target: "erp_tax_entries", source: "IVAIMPORTES_SALES"},
		columns:     []string{"id", "tenant_id", "invoice_id", "period", "direction", "net_amount", "tax_rate", "tax_amount"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ivaimporte")
			if legacyID == 0 {
				return nil, nil
			}

			// Filter: only sales (tipoiva=1)
			tipoiva := row.Int("tipoiva")
			if tipoiva != 1 {
				return nil, nil // skip purchases — handled by NewTaxEntryPurchasesMigrator
			}

			// Resolve invoice FK: ivacv_id → IVAVENTAS
			ivacvID := row.Int64("ivacv_id")
			var invoiceID *uuid.UUID
			if ivacvID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "invoicing", "IVAVENTAS", ivacvID)
				if err == nil && resolved != uuid.Nil {
					invoiceID = &resolved
				}
			}

			// Period from feciva (YYYYMM)
			feciva := timeFromRow(row, "feciva")
			period := ""
			if !feciva.IsZero() {
				period = fmt.Sprintf("%d%02d", feciva.Year(), feciva.Month())
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "IVAIMPORTES_SALES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				invoiceID,
				period,
				"sales",
				ParseDecimal(row.Decimal("impgra")),
				ParseDecimal(row.Decimal("impali")),
				ParseDecimal(row.Decimal("impimp")),
			}, nil
		},
	}
}

// NewTaxEntryPurchasesMigrator migrates IVAIMPORTES rows for purchases (tipoiva=2) → erp_tax_entries.
// ivacv_id links to IVACOMPRAS.id_ivacompra.
func NewTaxEntryPurchasesMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.VATAmountReader(db)
	return &GenericMigrator{
		reader:      &readerOverride{r: reader, target: "erp_tax_entries", source: "IVAIMPORTES_PURCHASES"},
		columns:     []string{"id", "tenant_id", "invoice_id", "period", "direction", "net_amount", "tax_rate", "tax_amount"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ivaimporte")
			if legacyID == 0 {
				return nil, nil
			}

			// Filter: only purchases (tipoiva=2)
			tipoiva := row.Int("tipoiva")
			if tipoiva != 2 {
				return nil, nil
			}

			// Resolve invoice FK: ivacv_id → IVACOMPRAS
			ivacvID := row.Int64("ivacv_id")
			var invoiceID *uuid.UUID
			if ivacvID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "invoicing", "IVACOMPRAS", ivacvID)
				if err == nil && resolved != uuid.Nil {
					invoiceID = &resolved
				}
			}

			feciva := timeFromRow(row, "feciva")
			period := ""
			if !feciva.IsZero() {
				period = fmt.Sprintf("%d%02d", feciva.Year(), feciva.Month())
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "IVAIMPORTES_PURCHASES", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				invoiceID,
				period,
				"purchases",
				ParseDecimal(row.Decimal("impgra")),
				ParseDecimal(row.Decimal("impali")),
				ParseDecimal(row.Decimal("impimp")),
			}, nil
		},
	}
}

// --- Withholding Migrators (Phase 8) ---

// NewWithholdingGainsMigrator migrates RETGANAN (gains withholding, 9.8K rows) → erp_withholdings.
// Columns: id_retganan=PK, ctacod=entity, ganfec=date, ganpor=rate%, ganbru=base, gantot=total.
// ctacod is ambiguous in Histrix (sometimes id_regcuenta, sometimes nro_cuenta)
// — ResolveEntityFlexible tries both and falls back to the "UNKNOWN" entity
// seeded by EnsureUnknownEntity rather than dropping the row.
func NewWithholdingGainsMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.WithholdingGainsReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "invoice_id", "movement_id", "entity_id", "type", "rate", "base_amount", "amount", "certificate_num", "date"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_retganan")
			if legacyID == 0 {
				return nil, nil
			}

			entityLegacyID := row.Int64("ctacod")
			resolved, _ := mapper.ResolveEntityFlexible(ctx, entityLegacyID)
			var entityID *uuid.UUID
			if resolved != uuid.Nil {
				entityID = &resolved
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "RETGANAN", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				(*uuid.UUID)(nil), // no invoice FK in RETGANAN
				(*uuid.UUID)(nil), // no movement FK
				entityID,
				"gains",
				ParseDecimal(row.Decimal("ganpor")),
				ParseDecimal(row.Decimal("ganbru")),
				ParseDecimal(row.Decimal("gantot")),
				row.NullString("gannro"),
				SafeDateRequired(timeFromRow(row, "ganfec")),
			}, nil
		},
	}
}

// NewWithholdingIVAMigrator migrates RETIVA (IVA withholding, 33 rows) → erp_withholdings.
// Columns: id_retiva=PK, ctacod=entity, ivafec=date, ivapor=rate%, ivabru=base, ivatot=total.
func NewWithholdingIVAMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.WithholdingIVAReader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "invoice_id", "movement_id", "entity_id", "type", "rate", "base_amount", "amount", "certificate_num", "date"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_retiva")
			if legacyID == 0 {
				return nil, nil
			}

			entityLegacyID := row.Int64("ctacod")
			resolved, _ := mapper.ResolveEntityFlexible(ctx, entityLegacyID)
			var entityID *uuid.UUID
			if resolved != uuid.Nil {
				entityID = &resolved
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "RETIVA", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				(*uuid.UUID)(nil),
				(*uuid.UUID)(nil),
				entityID,
				"iva",
				ParseDecimal(row.Decimal("ivapor")),
				ParseDecimal(row.Decimal("ivabru")),
				ParseDecimal(row.Decimal("ivatot")),
				row.NullString("ivanro"),
				SafeDateRequired(timeFromRow(row, "ivafec")),
			}, nil
		},
	}
}

// NewWithholding1598Migrator migrates RET1598 (IIBB reg 1598, 8.7K rows) → erp_withholdings.
// Columns: id_ret1598=PK, ctacod=entity, fecret=date, alicuota=rate%, totimpon=base, totret=total.
func NewWithholding1598Migrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.Withholding1598Reader(db)
	return &GenericMigrator{
		reader:      reader,
		columns:     []string{"id", "tenant_id", "invoice_id", "movement_id", "entity_id", "type", "rate", "base_amount", "amount", "certificate_num", "date"},
		conflictCol: "",
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_ret1598")
			if legacyID == 0 {
				return nil, nil
			}

			entityLegacyID := row.Int64("ctacod")
			resolved, _ := mapper.ResolveEntityFlexible(ctx, entityLegacyID)
			var entityID *uuid.UUID
			if resolved != uuid.Nil {
				entityID = &resolved
			}

			id, err := mapper.Map(ctx, nil, "invoicing", "RET1598", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				(*uuid.UUID)(nil),
				(*uuid.UUID)(nil),
				entityID,
				"iibb",
				ParseDecimal(row.Decimal("alicuota")),
				ParseDecimal(row.Decimal("totimpon")),
				ParseDecimal(row.Decimal("totret")),
				row.NullString("nroret"),
				SafeDateRequired(timeFromRow(row, "fecret")),
			}, nil
		},
	}
}

// --- Reader adapters ---

// readerOverride wraps a legacy.Reader to override the target and/or source table names.
// Used when a single MySQL table feeds multiple SDA tables (e.g., IVAIMPORTES → sales + purchases).
type readerOverride struct {
	r      legacy.Reader
	target string // override SDA table name
	source string // override legacy table name (for progress tracking uniqueness)
}

func (o *readerOverride) LegacyTable() string {
	if o.source != "" {
		return o.source
	}
	return o.r.LegacyTable()
}
func (o *readerOverride) SDATable() string { return o.target }
func (o *readerOverride) Domain() string   { return o.r.Domain() }
func (o *readerOverride) ReadBatch(ctx context.Context, resumeKey string, limit int) ([]legacy.LegacyRow, string, error) {
	return o.r.ReadBatch(ctx, resumeKey, limit)
}

// --- Helpers ---

func timeFromRow(row legacy.LegacyRow, col string) time.Time {
	v, ok := row[col]
	if !ok || v == nil {
		return time.Time{}
	}
	switch t := v.(type) {
	case time.Time:
		return t
	case []byte:
		parsed, _ := time.Parse("2006-01-02", string(t))
		return parsed
	case string:
		parsed, _ := time.Parse("2006-01-02", t)
		return parsed
	default:
		return time.Time{}
	}
}

// hashCode generates a stable non-negative int64 hash from a string code using
// FNV-1a 64-bit, clearing the sign bit. Used for tables where the PK is varchar
// (e.g. STK_ARTICULOS) or composite, so they can be stored as int64 legacy_id in
// erp_legacy_mapping without bit-63 overflow collisions.
func hashCode(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	v := h.Sum64() & 0x7FFFFFFFFFFFFFFF // clear sign bit — guarantees positive int64
	if v == 0 {
		v = 1
	}
	return int64(v)
}

// articleCompositeCode namespaces an article code by its subsystem so rows
// that share id_stkarticulo across different subsistema_id values in
// STK_ARTICULOS (the legacy PK is composite) remain distinct after migration.
// Subsystem "01" is the dominant one and is treated as the default (no prefix)
// so existing FKs referencing id_stkarticulo keep resolving via hashCode(code).
func articleCompositeCode(rawCode, subsistema string) string {
	if subsistema == "" || subsistema == "01" {
		return rawCode
	}
	return "SS" + subsistema + "|" + rawCode
}
