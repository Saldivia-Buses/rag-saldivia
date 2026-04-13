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

			cerrado := row.Int("cerrado")
			status := "open"
			if cerrado == 1 {
				status = "closed"
			}

			// Parse year from nombre_ejercicio or comienza date
			year := 0
			if nombre := row.String("nombre_ejercicio"); nombre != "" {
				fmt.Sscanf(nombre, "%d", &year)
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

			number := fmt.Sprintf("AS-%d", row.Int64("nro_minuta"))
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

			// doh: 0=debe, 1=haber
			doh := row.Int("doh")
			amount := ParseDecimal(row.Decimal("importe"))

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
			code := row.String("id_stkarticulo")
			if code == "" {
				return nil, nil
			}
			// Use a hash of the code as pseudo legacy ID for mapping
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
			artLegacyID := int64(hashCode(artCode))
			articleID, err := mapper.Resolve(ctx, "stock", "STK_ARTICULOS", artLegacyID)
			if err != nil {
				return nil, nil // skip if article not found
			}

			warehouseLegacyID := row.Int64("stkdeposito_id")
			warehouseID, err := mapper.Resolve(ctx, "stock", "STK_DEPOSITOS", warehouseLegacyID)
			if err != nil {
				return nil, nil // skip if warehouse not found
			}

			// Movement type — positive quantity = in, negative = out
			quantity := ParseDecimal(row.Decimal("cantidad"))
			movType := "in"
			if quantity.IsNegative() {
				movType = "out"
				quantity = quantity.Abs()
			}

			if quantity.IsZero() {
				return nil, nil // skip zero-quantity movements
			}

			id, err := mapper.Map(ctx, nil, "stock", "STK_MOVIMIENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, articleID, warehouseID, movType, quantity,
				ParseDecimal(row.Decimal("precio_costo")),
				LegacyUserID,
				row.String("referencia"),
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
			supplierID, err := mapper.Resolve(ctx, "entity", "REG_CUENTA", supplierLegacyID)
			if err != nil {
				return nil, nil // skip if supplier not found
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
			artLegacyID := int64(hashCode(artCode))
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
			customerID, err := mapper.Resolve(ctx, "entity", "REG_CUENTA", customerLegacyID)
			if err != nil {
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
			artLegacyID := int64(hashCode(artCode))
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

			id := uuid.New()

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

// hashCode generates a stable int64 hash from a string code using FNV-1a 64-bit.
// Used for STK_ARTICULOS where PK is varchar, not int.
// FNV-1a has much better collision resistance than DJB hash.
func hashCode(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	v := h.Sum64()
	if v == 0 {
		v = 1
	}
	return v
}
