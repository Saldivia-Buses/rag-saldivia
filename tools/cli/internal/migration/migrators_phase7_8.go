package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Camionerou/rag-saldivia/tools/cli/internal/legacy"
)

// ============================================================================
// Phase 7 — Treasury (cash movements, bank movements, cash counts)
// ============================================================================

// NewCashMovementMigrator migrates CAJMOVIM → erp_treasury_movements.
// CAJMOVIM has composite PK (cajcod, cajcta, cajfec, cajnpv, cajnro, concod, regmin, siscod, succod).
// ~116K rows. The most complex treasury table.
// cajimp = amount, cajest = status (1=confirmed), cajfec = date.
// cajcod = cash register code, concod = concept code.
// siscod+opecod determine movement type (cash_in/cash_out).
func NewCashMovementMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CashMovementReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "date", "number", "movement_type", "amount", "currency_id", "bank_account_id", "cash_register_id", "entity_id", "concept_id", "payment_method", "reference_type", "reference_id", "journal_entry_id", "user_id", "notes", "status"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite PK — generate deterministic ID from hash of all PK columns
			compositeKey := fmt.Sprintf("CAJMOVIM:%d:%d:%s:%d:%d:%d:%d:%s:%d",
				row.Int("cajcod"), row.Int("cajcta"), row.String("cajfec"),
				row.Int("cajnpv"), row.Int("cajnro"), row.Int("concod"),
				row.Int("regmin"), row.String("siscod"), row.Int("succod"))
			legacyID := int64(hashCode(compositeKey))

			// Amount must be > 0 (CHECK constraint). Use absolute value, skip zeros.
			amount := ParseDecimal(row.Decimal("cajimp")).Abs()
			if amount.IsZero() {
				return nil, nil // skip zero-amount movements
			}

			// Movement type: cash movements are cash_in or cash_out.
			// cajimp positive → cash_in, negative → cash_out
			rawImp := ParseDecimal(row.Decimal("cajimp"))
			movType := "cash_in"
			if rawImp.IsNegative() {
				movType = "cash_out"
			}

			// Status: cajest=1 → confirmed (only value found in data)
			status := "confirmed"

			// Number: MOV-{composite key hash}
			number := fmt.Sprintf("MOV-%d", legacyID)

			// Cash register FK: cajcod → CAJ_PUESTOS mapping
			var cashRegisterID *uuid.UUID
			cajcod := row.Int64("cajcod")
			if cajcod > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "treasury", "CAJ_PUESTOS", cajcod)
				if err == nil && resolved != uuid.Nil {
					cashRegisterID = &resolved
				}
			}

			// Entity FK (optional): codent or opecla might link to REG_CUENTA
			var entityID *uuid.UUID
			codent := row.Int64("codent")
			if codent > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "REG_CUENTA", codent)
				if err == nil && resolved != uuid.Nil {
					entityID = &resolved
				}
			}

			date := SafeDateRequired(timeFromRow(row, "cajfec"))

			id, err := mapper.Map(ctx, nil, "treasury", "CAJMOVIM", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Notes: combine references
			var notes []string
			if ref1 := row.String("cajref__1"); ref1 != "" {
				notes = append(notes, ref1)
			}
			if ref2 := row.String("cajref__2"); ref2 != "" {
				notes = append(notes, ref2)
			}
			if ref3 := row.String("cajref__3"); ref3 != "" {
				notes = append(notes, ref3)
			}
			if nom := row.String("cajnom"); nom != "" {
				notes = append(notes, nom)
			}

			return []any{
				id, tenantID, date, number, movType, amount,
				(*uuid.UUID)(nil), // currency_id
				(*uuid.UUID)(nil), // bank_account_id (this is a cash movement)
				cashRegisterID,
				entityID,
				(*uuid.UUID)(nil), // concept_id
				nil,               // payment_method
				nil,               // reference_type
				(*uuid.UUID)(nil), // reference_id
				(*uuid.UUID)(nil), // journal_entry_id
				LegacyUserID,
				strings.Join(notes, " | "),
				status,
			}, nil
		},
	}
}

// NewBankMovementMigrator migrates CAR_MOVIMIENTOS → erp_treasury_movements.
// CAR_MOVIMIENTOS has auto-increment id_carmovimiento PK. ~162 rows.
// tipo_movimiento=1 for all rows in data (bank check/value intake).
// Links to bank account via carvalor_id → CARCHEQU, and entity via regcuenta_id.
func NewBankMovementMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.BankMovementReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "date", "number", "movement_type", "amount", "currency_id", "bank_account_id", "cash_register_id", "entity_id", "concept_id", "payment_method", "reference_type", "reference_id", "journal_entry_id", "user_id", "notes", "status"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_carmovimiento")
			if legacyID == 0 {
				return nil, nil
			}

			// tipo_movimiento: 1=intake. Bank movements are bank_deposit/bank_withdrawal.
			// In the data all are tipo_movimiento=1 (intake to check portfolio).
			tipoMov := row.Int("tipo_movimiento")
			movType := "bank_deposit"
			if tipoMov == 2 {
				movType = "bank_withdrawal"
			}

			// Entity FK: regcuenta_id → REG_CUENTA
			var entityID *uuid.UUID
			regcuentaID := row.Int64("regcuenta_id")
			if regcuentaID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "REG_CUENTA", regcuentaID)
				if err == nil && resolved != uuid.Nil {
					entityID = &resolved
				}
			}

			// Check FK: carvalor_id → CARCHEQU (check reference)
			var referenceID *uuid.UUID
			carvalorID := row.Int64("carvalor_id")
			if carvalorID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "treasury", "CARCHEQU", carvalorID)
				if err == nil && resolved != uuid.Nil {
					referenceID = &resolved
				}
			}

			// Bank movements need a positive amount. Since we don't have an amount column
			// in CAR_MOVIMIENTOS directly, we set a nominal amount.
			// The actual value is tracked in CARCHEQU (check table).
			// Use a default positive amount since CHECK(amount > 0).
			amount := ParseDecimal("0.01")

			date := SafeDateRequired(timeFromRow(row, "fecha_movimiento"))
			number := fmt.Sprintf("BNK-%d", legacyID)

			id, err := mapper.Map(ctx, nil, "treasury", "CAR_MOVIMIENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			refType := "check"
			return []any{
				id, tenantID, date, number, movType, amount,
				(*uuid.UUID)(nil), // currency_id
				(*uuid.UUID)(nil), // bank_account_id — not directly in CAR_MOVIMIENTOS
				(*uuid.UUID)(nil), // cash_register_id
				entityID,
				(*uuid.UUID)(nil), // concept_id
				nil,               // payment_method
				&refType,          // reference_type
				referenceID,       // reference_id → check UUID
				(*uuid.UUID)(nil), // journal_entry_id
				LegacyUserID,
				row.String("referencia_movimiento"),
				"confirmed",
			}, nil
		},
	}
}

// NewCashCountMigrator migrates CAJ_PUESTO_ARQUEOS → erp_cash_counts.
// CAJ_PUESTO_ARQUEOS has auto-increment id_cajpuestoarqueo PK. ~14 rows.
// Maps payment method configurations per cash register to cash count records.
// Columns: cash_register_id (NOT NULL FK), date, expected, counted, difference, user_id.
func NewCashCountMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.CashCountReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "cash_register_id", "date", "expected", "counted", "difference", "user_id", "notes"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_cajpuestoarqueo")
			if legacyID == 0 {
				return nil, nil
			}

			// Cash register FK: cajpuesto_id → CAJ_PUESTOS (NOT NULL)
			cajpuestoID := row.Int64("cajpuesto_id")
			cashRegisterID, err := mapper.Resolve(ctx, "treasury", "CAJ_PUESTOS", cajpuestoID)
			if err != nil {
				return nil, nil // skip if cash register not migrated
			}

			id, err := mapper.Map(ctx, nil, "treasury", "CAJ_PUESTO_ARQUEOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Legacy data is a config table (payment method per register), not actual count data.
			// Map with zero amounts — actual counts are not stored in Histrix.
			return []any{
				id, tenantID, cashRegisterID,
				SafeDateRequired(timeFromRow(row, "")), // no date column — use epoch
				ParseDecimal("0"),                      // expected
				ParseDecimal("0"),                      // counted
				ParseDecimal("0"),                      // difference
				LegacyUserID,
				fmt.Sprintf("form_pago=%d orden=%d", row.Int("cajformapago_id"), row.Int("orden_arqueo")),
			}, nil
		},
	}
}

// ============================================================================
// Phase 8 — Current Accounts (account movements, payment allocations)
// ============================================================================

// accountMovementDirectionMap maps subsistema_id to erp_account_movements.direction.
// Different from MapDirection() which maps to invoice direction (received/issued).
// erp_account_movements CHECK: ('receivable', 'payable').
var accountMovementDirectionMap = map[string]string{
	"01": "payable",    // compras/proveedores
	"1":  "payable",
	"02": "receivable", // ventas/clientes
	"2":  "receivable",
	"03": "receivable", // (other subsystem, default to receivable)
	"3":  "receivable",
	"04": "payable",    // (other subsystem, default to payable)
	"4":  "payable",
}

// mapAccountDirection maps subsistema_id to receivable/payable for erp_account_movements.
func mapAccountDirection(subsistemaID string) string {
	if v, ok := accountMovementDirectionMap[subsistemaID]; ok {
		return v
	}
	// Default: treat unknown subsystems as receivable
	return "receivable"
}

// mapAccountMovementTypeFromConcepto maps concepto_id to movement_type.
// concepto_id in REG_MOVIMIENTOS identifies the document type.
// Top values: 64=factura, 67=recibo, 2=nota_credito, 4=nota_debito, 7=ajuste.
var conceptoToMovementType = map[int]string{
	2:  "credit_note",
	3:  "invoice",
	4:  "debit_note",
	7:  "adjustment",
	11: "payment",
	13: "adjustment",
	14: "adjustment",
	20: "invoice",
	39: "adjustment",
	64: "invoice",
	65: "credit_note",
	66: "debit_note",
	67: "payment",
	68: "payment",
	71: "payment",
	72: "credit_note",
	73: "invoice",
	75: "invoice",
	76: "invoice",
	95: "adjustment",
	96: "adjustment",
	98: "adjustment",
}

func mapMovementTypeFromConcepto(conceptoID int) string {
	if v, ok := conceptoToMovementType[conceptoID]; ok {
		return v
	}
	return "adjustment" // default for unmapped concepto_id
}

// NewAccountMovementMigrator migrates REG_MOVIMIENTOS → erp_account_movements.
// REG_MOVIMIENTOS has auto-increment id_regmovim PK. ~291K rows.
// The most complex legacy table with 70+ columns.
// erp_account_movements: id, tenant_id, entity_id (NOT NULL), date, movement_type,
//
//	direction, amount (>0), balance, invoice_id, treasury_id, journal_entry_id,
//	notes, user_id, metadata.
func NewAccountMovementMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.AccountMovementReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "entity_id", "date", "movement_type", "direction", "amount", "balance", "invoice_id", "treasury_id", "journal_entry_id", "notes", "user_id", "metadata"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_regmovim")
			if legacyID == 0 {
				return nil, nil
			}

			// Entity FK: regcuenta_id → REG_CUENTA (NOT NULL in target)
			regcuentaID := row.Int64("regcuenta_id")
			if regcuentaID == 0 {
				return nil, nil // skip rows with no entity
			}
			entityID, err := mapper.Resolve(ctx, "entity", "REG_CUENTA", regcuentaID)
			if err != nil {
				return nil, nil // skip if entity not migrated
			}

			// Direction: subsistema_id → receivable/payable
			subsistemaID := row.String("subsistema_id")
			direction := mapAccountDirection(subsistemaID)

			// Movement type: from concepto_id
			conceptoID := row.Int("concepto_id")
			movType := mapMovementTypeFromConcepto(conceptoID)

			// Amount: must be > 0 (CHECK constraint), use absolute value
			amount := ParseDecimal(row.Decimal("importe_movimiento")).Abs()
			if amount.IsZero() {
				return nil, nil // skip zero-amount movements
			}

			// Balance from saldo_movimiento
			balance := ParseDecimal(row.Decimal("saldo_movimiento"))

			// Invoice FK (optional): try to resolve via regmovim index
			// facremit_id is not a direct column, but cajmovimiento_id links to treasury
			var invoiceID *uuid.UUID
			// If regMovimIndex exists, this movement itself might be linked as invoice

			// Treasury FK (optional): cajmovimiento_id → CAJMOVIM
			var treasuryID *uuid.UUID
			cajmovID := row.Int64("cajmovimiento_id")
			if cajmovID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "treasury", "CAJMOVIM", cajmovID)
				if err == nil && resolved != uuid.Nil {
					treasuryID = &resolved
				}
			}

			date := SafeDateRequired(timeFromRow(row, "fecha_movimiento"))

			id, err := mapper.Map(ctx, nil, "current_account", "REG_MOVIMIENTOS", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Notes: combine referencia + novedades
			notes := row.String("referencia_movimiento")
			if n := row.String("novedades_movimiento"); n != "" && n != "0" {
				if notes != "" {
					notes += " | "
				}
				notes += n
			}

			// Metadata: store useful legacy fields
			meta := map[string]any{
				"subsistema_id":      subsistemaID,
				"concepto_id":        conceptoID,
				"letra_movimiento":   row.String("letra_movimiento"),
				"nro_movimiento":     row.Int("nro_movimiento"),
				"puesto_movimiento":  row.Int("puesto_movimiento"),
				"cuenta_movimiento":  row.Int("cuenta_movimiento"),
				"cuota_movimiento":   row.Int("cuota_movimiento"),
				"importe_iva":        row.Decimal("importe_iva"),
				"importe_exento":     row.Decimal("importe_exento"),
				"importe_percepcion": row.Decimal("importe_percepcion"),
				"source":             "REG_MOVIMIENTOS",
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID, entityID, date, movType, direction,
				amount, balance,
				invoiceID,
				treasuryID,
				(*uuid.UUID)(nil), // journal_entry_id
				notes,
				LegacyUserID,
				string(metaJSON),
			}, nil
		},
	}
}

// NewPaymentAllocationMigrator migrates CCTIMPUT → erp_payment_allocations.
// CCTIMPUT has auto-increment id_cctimput PK. ~132K rows.
// Each row links a payment to an invoice document in erp_account_movements.
// erp_payment_allocations: id, tenant_id, payment_id (NOT NULL FK→erp_account_movements),
//
//	invoice_id (NOT NULL FK→erp_account_movements), amount (>0).
//
// Both payment_id and invoice_id reference erp_account_movements, NOT erp_invoices.
// regmovim0_id = the payment movement, regmovim_id = the invoice movement being applied to.
func NewPaymentAllocationMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.PaymentAllocationLineReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "payment_id", "invoice_id", "amount"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_cctimput")
			if legacyID == 0 {
				return nil, nil
			}

			// Payment FK: regmovim0_id → REG_MOVIMIENTOS (the payment/receipt)
			paymentLegacyID := row.Int64("regmovim0_id")
			if paymentLegacyID == 0 {
				return nil, nil // skip rows with no payment reference
			}
			paymentID, err := mapper.Resolve(ctx, "current_account", "REG_MOVIMIENTOS", paymentLegacyID)
			if err != nil {
				return nil, nil // skip if payment not migrated
			}

			// Invoice FK: regmovim_id → REG_MOVIMIENTOS (the invoice/document being paid)
			invoiceLegacyID := row.Int64("regmovim_id")
			if invoiceLegacyID == 0 {
				// Fallback: try to construct from movinc (nro_movimiento in REG_MOVIMIENTOS)
				// This is a best-effort approach for rows where regmovim_id=0
				return nil, nil // skip — can't link without a valid FK
			}
			invoiceID, err := mapper.Resolve(ctx, "current_account", "REG_MOVIMIENTOS", invoiceLegacyID)
			if err != nil {
				return nil, nil // skip if invoice movement not migrated
			}

			// Amount: movimp, must be > 0
			amount := ParseDecimal(row.Decimal("movimp")).Abs()
			if amount.IsZero() {
				// Try movsal (saldo) or movint as fallback
				amount = ParseDecimal(row.Decimal("movint")).Abs()
			}
			if amount.IsZero() {
				return nil, nil // skip zero allocations
			}

			id, err := mapper.Map(ctx, nil, "current_account", "CCTIMPUT", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID, paymentID, invoiceID, amount,
			}, nil
		},
	}
}

// ============================================================================
// Phase 8 — Withholdings IIBB (missing from existing withholding migrators)
// ============================================================================

// NewWithholdingIIBBMigrator migrates RETACUMU (IIBB withholdings, 39K rows) → erp_withholdings.
// Composite PK (siscod, ctacod, acufec, year, mes, nropag).
// erp_withholdings: entity_id NOT NULL, amount CHECK(>0), type='iibb'.
// acuret = withholding amount, acupag = payment amount, acufec = date.
func NewWithholdingIIBBMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.WithholdingIIBBReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "invoice_id", "movement_id", "entity_id", "type", "rate", "base_amount", "amount", "certificate_num", "date"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			// Composite PK — generate deterministic ID from hash
			compositeKey := fmt.Sprintf("RETACUMU:%s:%d:%s:%d:%d:%s",
				row.String("siscod"), row.Int("ctacod"), row.String("acufec"),
				row.Int("year"), row.Int("mes"), row.String("nropag"))
			legacyID := int64(hashCode(compositeKey))

			// Entity FK: ctacod → REG_CUENTA (NOT NULL in target)
			ctacod := row.Int64("ctacod")
			if ctacod == 0 {
				return nil, nil // skip rows with no entity
			}
			entityID, err := mapper.Resolve(ctx, "entity", "REG_CUENTA", ctacod)
			if err != nil {
				return nil, nil // skip if entity not migrated
			}

			// Amount: acuret (withholding amount), must be > 0
			amount := ParseDecimal(row.Decimal("acuret")).Abs()
			if amount.IsZero() {
				return nil, nil // skip zero withholdings
			}

			// Base amount: acupag (payment amount)
			baseAmount := ParseDecimal(row.Decimal("acupag")).Abs()

			// Rate: not directly stored in RETACUMU, set to 0
			rate := ParseDecimal("0")

			// Certificate number: nropag (payment number)
			certNum := row.NullString("nropag")

			// Date: acufec
			date := SafeDate(timeFromRow(row, "acufec"))

			id, err := mapper.Map(ctx, nil, "invoicing", "RETACUMU", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			return []any{
				id, tenantID,
				(*uuid.UUID)(nil), // no invoice FK
				(*uuid.UUID)(nil), // no movement FK
				entityID,
				"iibb",
				rate,
				baseAmount,
				amount,
				certNum,
				date,
			}, nil
		},
	}
}

// ============================================================================
// Phase 11 — Production Inspection Details (1.7M rows)
// ============================================================================

// NewProductionInspectionDetailMigrator migrates PROD_CONTROL_MOVIM → erp_production_inspections.
// PROD_CONTROL_MOVIM has auto-increment id_controlmovim PK. 1.7M rows.
// Uses ProductionInspectionDetailReader which JOINs with PROD_CONTROLES to get control_orden_id.
// erp_production_inspections: order_id (NOT NULL FK), step_id, inspector_id, result, observations, metadata.
// result: conforme=1→'pass', no_conforme=1→'fail', else→'rework'.
func NewProductionInspectionDetailMigrator(db *sql.DB, tenantID string) *GenericMigrator {
	reader := legacy.NewProductionInspectionDetailReader(db)
	return &GenericMigrator{
		reader:  reader,
		columns: []string{"id", "tenant_id", "order_id", "step_id", "inspector_id", "result", "observations", "metadata"},
		transformFn: func(ctx context.Context, row legacy.LegacyRow, mapper *Mapper) ([]any, error) {
			legacyID := row.Int64("id_controlmovim")
			if legacyID == 0 {
				return nil, nil
			}

			// Order FK: nrofab_id → CHASIS mapping (production unit → order)
			// The vehicle/unit is the production context. Use nrofab_id to find the unit,
			// then try to resolve a production order.
			nrofabID := row.Int64("nrofab_id")
			if nrofabID == 0 {
				return nil, nil // skip rows with no vehicle reference
			}

			// Try resolving nrofab_id as a CHASIS entry → production order
			// CHASIS.nrocha maps to erp_units in the production domain
			unitID, err := mapper.ResolveOptional(ctx, "production", "CHASIS", nrofabID)
			if err != nil || unitID == uuid.Nil {
				return nil, nil // skip if vehicle not migrated
			}

			// We need an order_id (NOT NULL FK to erp_production_orders).
			// PROD_CONTROL_MOVIM doesn't have a direct production order FK.
			// control_orden_id from the JOIN gives us the control's orden_control.
			// Try to resolve via MRP_PEDIDO_PRODUCCION or similar.
			// If we can't find an order, we must skip — order_id is NOT NULL.
			controlOrdenID := row.Int64("control_orden_id")
			var orderID uuid.UUID
			if controlOrdenID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "production", "MRP_PEDIDO_PRODUCCION", controlOrdenID)
				if err == nil && resolved != uuid.Nil {
					orderID = resolved
				}
			}
			if orderID == uuid.Nil {
				// Fallback: try PROD_PROCESOS as order context
				prodcontrolID := row.Int64("prodcontrol_id")
				if prodcontrolID > 0 {
					resolved, err := mapper.ResolveOptional(ctx, "production", "PROD_CONTROLES", prodcontrolID)
					if err == nil && resolved != uuid.Nil {
						// PROD_CONTROLES maps to erp_production_inspections, not orders.
						// We need a valid order_id. Skip if we can't resolve.
					}
				}
				return nil, nil // skip — order_id is NOT NULL
			}

			// Step FK: prodcontrol_id → PROD_CONTROLES (inspection template → step)
			var stepID *uuid.UUID
			prodcontrolID := row.Int64("prodcontrol_id")
			if prodcontrolID > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "production", "PROD_PROCESOS", prodcontrolID)
				if err == nil && resolved != uuid.Nil {
					stepID = &resolved
				}
			}

			// Inspector FK: legajo_realizo or legajo_personal → PERSONAL
			var inspectorID *uuid.UUID
			legajoRealizo := row.Int64("legajo_realizo")
			if legajoRealizo > 0 {
				resolved, err := mapper.ResolveOptional(ctx, "entity", "PERSONAL", legajoRealizo)
				if err == nil && resolved != uuid.Nil {
					inspectorID = &resolved
				}
			}
			if inspectorID == nil {
				legajoPersonal := row.Int64("legajo_personal")
				if legajoPersonal > 0 {
					resolved, err := mapper.ResolveOptional(ctx, "entity", "PERSONAL", legajoPersonal)
					if err == nil && resolved != uuid.Nil {
						inspectorID = &resolved
					}
				}
			}

			// Result: conforme=1→'pass', no_conforme=1→'fail', else→'rework'
			result := "rework"
			if row.Int("conforme") == 1 {
				result = "pass"
			} else if row.Int("no_conforme") == 1 {
				result = "fail"
			}

			id, err := mapper.Map(ctx, nil, "production", "PROD_CONTROL_MOVIM", legacyID, nil)
			if err != nil {
				id = uuid.New()
			}

			// Observations
			obs := row.String("observacion_control")
			if acciones := row.String("acciones_control"); acciones != "" {
				if obs != "" {
					obs += " | "
				}
				obs += acciones
			}

			// Metadata: store rework hours, causal, dates, photo
			meta := map[string]any{
				"nrofab_id":         nrofabID,
				"prodcontrol_id":    prodcontrolID,
				"valor_control":     row.Int("valor_control"),
				"horas_retrabajo":   row.String("horas_retrabajo"),
				"controlcausal_id":  row.Int("controlcausal_id"),
				"controlado":        row.Int("controlado"),
				"no_aplica":         row.Int("no_aplica"),
				"foto":              row.String("foto"),
				"seccion_origen_id": row.String("seccion_origen_id"),
				"source":            "PROD_CONTROL_MOVIM",
			}
			metaJSON, _ := json.Marshal(meta)

			return []any{
				id, tenantID, orderID, stepID, inspectorID,
				result, obs, string(metaJSON),
			}, nil
		},
	}
}
