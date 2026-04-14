package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// ValidationRule defines a pre-validation check against legacy data.
type ValidationRule struct {
	Name        string
	Domain      string
	LegacyTable string
	Constraint  string // SDA constraint/trigger name that would be violated
	Query       string // MySQL query returning problematic rows
	Transform   string // default resolution: "skip", "fix_manual", "auto_fix"
}

// DefaultValidationRules returns pre-validation rules for Plan 17+18 constraints.
func DefaultValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Name:        "invalid_dates_zero",
			Domain:      "global",
			LegacyTable: "CTB_MOVIMIENTOS",
			Constraint:  "date_not_zero",
			Query:       "SELECT id_movimiento as legacy_id, CAST(fecha_movimiento AS CHAR) as detail FROM CTB_MOVIMIENTOS WHERE CAST(fecha_movimiento AS CHAR) = '0000-00-00' LIMIT 1000",
			Transform:   "auto_fix",
		},
		{
			Name:        "journal_unbalanced",
			Domain:      "accounting",
			LegacyTable: "CTB_MOVIMIENTOS",
			Constraint:  "erp_journal_balance",
			Query:       "SELECT id_movimiento as legacy_id, CAST(ABS(COALESCE(debe,0) - COALESCE(haber,0)) AS CHAR) as detail FROM CTB_MOVIMIENTOS WHERE ABS(COALESCE(debe,0) - COALESCE(haber,0)) > 0.01 LIMIT 1000",
			Transform:   "fix_manual",
		},
		{
			Name:        "checks_zero_amount",
			Domain:      "treasury",
			LegacyTable: "CARCHEQU",
			Constraint:  "erp_checks_amount_check",
			Query:       "SELECT carint as legacy_id, CAST(carimp AS CHAR) as detail FROM CARCHEQU WHERE carimp <= 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "invoice_dates_zero",
			Domain:      "invoicing",
			LegacyTable: "IVAVENTAS",
			Constraint:  "date_not_zero",
			Query:       "SELECT id_ivaventa as legacy_id, CAST(feciva AS CHAR) as detail FROM IVAVENTAS WHERE CAST(feciva AS CHAR) = '0000-00-00' LIMIT 1000",
			Transform:   "auto_fix",
		},
		{
			Name:        "stock_movements_zero_qty",
			Domain:      "stock",
			LegacyTable: "STK_MOVIMIENTOS",
			Constraint:  "erp_stock_movements_quantity_check",
			Query:       "SELECT id_stkmovimiento as legacy_id, CAST(cantidad AS CHAR) as detail FROM STK_MOVIMIENTOS WHERE cantidad = 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "orphan_journal_lines",
			Domain:      "accounting",
			LegacyTable: "CTB_DETALLES",
			Constraint:  "erp_journal_lines_entry_id_fkey",
			Query:       "SELECT d.id_detalle as legacy_id, CAST(d.movimiento_id AS CHAR) as detail FROM CTB_DETALLES d LEFT JOIN CTB_MOVIMIENTOS m ON d.movimiento_id = m.id_movimiento WHERE m.id_movimiento IS NULL LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "importe_not_numeric",
			Domain:      "accounting",
			LegacyTable: "CTB_DETALLES",
			Constraint:  "numeric_cast",
			Query:       "SELECT id_detalle as legacy_id, importe as detail FROM CTB_DETALLES WHERE importe IS NOT NULL AND importe != '' AND importe NOT REGEXP '^-?[0-9]+(\\\\.[0-9]+)?$' LIMIT 1000",
			Transform:   "skip",
		},
		// --- Phase 18 pre-validation rules (H7 fix) ---
		{
			Name:        "invoice_lines_qty_zero",
			Domain:      "invoicing",
			LegacyTable: "FACDETAL",
			Constraint:  "erp_invoice_lines_quantity_check",
			Query:       "SELECT CAST(remnro AS SIGNED) as legacy_id, CAST(artcan AS CHAR) as detail FROM FACDETAL WHERE artcan <= 0 LIMIT 1000",
			Transform:   "auto_fix",
		},
		{
			Name:        "invoice_total_zero",
			Domain:      "invoicing",
			LegacyTable: "IVAVENTAS",
			Constraint:  "erp_invoices_total_check",
			Query:       "SELECT id_ivaventa as legacy_id, CAST(totcom AS CHAR) as detail FROM IVAVENTAS WHERE totcom = 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "withholding_gains_zero",
			Domain:      "withholding",
			LegacyTable: "RETGANAN",
			Constraint:  "erp_withholdings_amount_check",
			Query:       "SELECT id_retganan as legacy_id, CAST(gantot AS CHAR) as detail FROM RETGANAN WHERE gantot = 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "withholding_iva_zero",
			Domain:      "withholding",
			LegacyTable: "RETIVA",
			Constraint:  "erp_withholdings_amount_check",
			Query:       "SELECT id_retiva as legacy_id, CAST(ivatot AS CHAR) as detail FROM RETIVA WHERE ivatot = 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "treasury_amount_zero",
			Domain:      "treasury",
			LegacyTable: "CAJMOVIM",
			Constraint:  "erp_treasury_movements_amount_check",
			Query:       "SELECT CAST(cajnro AS SIGNED) as legacy_id, CAST(cajimp AS CHAR) as detail FROM CAJMOVIM WHERE cajimp = 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "account_movement_amount_zero",
			Domain:      "accounting",
			LegacyTable: "REG_MOVIMIENTOS",
			Constraint:  "erp_account_movements_amount_check",
			Query:       "SELECT id_regmovim as legacy_id, CAST(importe_movimiento AS CHAR) as detail FROM REG_MOVIMIENTOS WHERE importe_movimiento = 0 LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "attendance_orphan_entity",
			Domain:      "hr",
			LegacyTable: "FICHADADIA",
			Constraint:  "erp_attendance_entity_fkey",
			Query:       "SELECT f.tarjeta as legacy_id, CAST(f.legajo AS CHAR) as detail FROM FICHADADIA f LEFT JOIN PERSONAL p ON f.legajo = p.legajo WHERE p.legajo IS NULL LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "production_inspection_result_invalid",
			Domain:      "production",
			LegacyTable: "PROD_CONTROL_MOVIM",
			Constraint:  "erp_production_inspections_result_check",
			Query:       "SELECT id_controlmovim as legacy_id, CONCAT('conforme=', conforme, ' no_conforme=', no_conforme) as detail FROM PROD_CONTROL_MOVIM WHERE conforme = 0 AND no_conforme = 0 AND no_aplica = 0 LIMIT 1000",
			Transform:   "auto_fix",
		},
		{
			Name:        "treasury_movement_type_unknown",
			Domain:      "treasury",
			LegacyTable: "CAJMOVIM",
			Constraint:  "erp_treasury_movements_type_check",
			Query:       "SELECT CAST(cajnro AS SIGNED) as legacy_id, CAST(concod AS CHAR) as detail FROM CAJMOVIM WHERE concod NOT IN (0,1,2,3,4,5,6,7,8,9) LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "requisition_number_empty",
			Domain:      "purchasing",
			LegacyTable: "PEDIDOINT",
			Constraint:  "erp_purchase_orders_number_check",
			Query:       "SELECT idPed as legacy_id, CAST(nrocom AS CHAR) as detail FROM PEDIDOINT WHERE nrocom = 0 OR nrocom IS NULL LIMIT 1000",
			Transform:   "skip",
		},
		{
			Name:        "article_hashcode_collision",
			Domain:      "stock",
			LegacyTable: "STK_ARTICULOS",
			Constraint:  "hashcode_uniqueness",
			Query:       "SELECT COUNT(*) as legacy_id, CAST(COUNT(DISTINCT id_stkarticulo) AS CHAR) as detail FROM STK_ARTICULOS HAVING COUNT(*) != COUNT(DISTINCT id_stkarticulo)",
			Transform:   "fix_manual",
		},
	}
}

// ValidationReport summarizes pre-validation results.
type ValidationReport struct {
	RunID      uuid.UUID
	TotalCount int
	AutoFix    int
	Skip       int
	FixManual  int
	Issues     []ValidationIssue
}

// ValidationIssue represents a single problematic row.
type ValidationIssue struct {
	Domain      string
	LegacyTable string
	LegacyID    int64
	Constraint  string
	Detail      string
	Resolution  string
}

// RunPreValidation executes all validation rules against the legacy MySQL database.
func RunPreValidation(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, runID uuid.UUID) (*ValidationReport, error) {
	rules := DefaultValidationRules()
	report := &ValidationReport{RunID: runID}

	for _, rule := range rules {
		slog.Info("pre-validation", "rule", rule.Name, "table", rule.LegacyTable)

		rows, err := mysqlDB.QueryContext(ctx, rule.Query)
		if err != nil {
			slog.Warn("pre-validation query failed (table may not exist)", "rule", rule.Name, "err", err)
			continue
		}

		var issues []ValidationIssue
		for rows.Next() {
			var legacyID int64
			var detail string
			if err := rows.Scan(&legacyID, &detail); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan %s: %w", rule.Name, err)
			}
			issues = append(issues, ValidationIssue{
				Domain:      rule.Domain,
				LegacyTable: rule.LegacyTable,
				LegacyID:    legacyID,
				Constraint:  rule.Constraint,
				Detail:      detail,
				Resolution:  rule.Transform,
			})
		}
		rows.Close()

		// Record issues in PostgreSQL
		for _, issue := range issues {
			detailJSON, _ := json.Marshal(map[string]string{"value": issue.Detail})
			pgPool.Exec(ctx,
				`INSERT INTO erp_migration_validation_issues
				 (tenant_id, run_id, domain, legacy_table, legacy_id, constraint_name, details, resolution)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				tenantID, runID, issue.Domain, issue.LegacyTable, issue.LegacyID,
				issue.Constraint, detailJSON, issue.Resolution,
			)
		}

		report.Issues = append(report.Issues, issues...)
		for _, i := range issues {
			switch i.Resolution {
			case "auto_fix":
				report.AutoFix++
			case "skip":
				report.Skip++
			case "fix_manual":
				report.FixManual++
			}
		}
		report.TotalCount += len(issues)

		slog.Info("pre-validation done", "rule", rule.Name, "issues", len(issues))
	}

	return report, nil
}

// PostMigrationValidation runs count and checksum validations after migration.
func PostMigrationValidation(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string) error {
	type countCheck struct {
		Name    string
		MySQL   string
		PG      string
	}

	checks := []countCheck{
		{"entities_customer", "SELECT COUNT(*) FROM REG_CUENTA WHERE subsistema_id = '02'", "SELECT COUNT(*) FROM erp_entities WHERE tenant_id=$1 AND type='customer'"},
		{"entities_supplier", "SELECT COUNT(*) FROM REG_CUENTA WHERE subsistema_id = '01'", "SELECT COUNT(*) FROM erp_entities WHERE tenant_id=$1 AND type='supplier'"},
		{"entities_employee", "SELECT COUNT(*) FROM PERSONAL", "SELECT COUNT(*) FROM erp_entities WHERE tenant_id=$1 AND type='employee'"},
		{"accounts", "SELECT COUNT(DISTINCT id_ctbcuenta) FROM CTB_CUENTAS", "SELECT COUNT(*) FROM erp_accounts WHERE tenant_id=$1"},
		{"journal_entries", "SELECT COUNT(*) FROM CTB_MOVIMIENTOS", "SELECT COUNT(*) FROM erp_journal_entries WHERE tenant_id=$1"},
		{"warehouses", "SELECT COUNT(*) FROM STK_DEPOSITOS", "SELECT COUNT(*) FROM erp_warehouses WHERE tenant_id=$1"},
		// Phase 18 — extended post-validation checks
		{"invoices", "SELECT (SELECT COUNT(*) FROM IVAVENTAS) + (SELECT COUNT(*) FROM IVACOMPRAS)", "SELECT COUNT(*) FROM erp_invoices WHERE tenant_id=$1 AND invoice_type != 'delivery_note'"},
		{"bom_current", "SELECT COUNT(*) FROM STKPIEZA", "SELECT COUNT(*) FROM erp_bom WHERE tenant_id=$1"},
		{"bom_history", "SELECT COUNT(*) FROM STK_BOM_HIST", "SELECT COUNT(*) FROM erp_bom_history WHERE tenant_id=$1"},
		{"attendance", "SELECT COUNT(*) FROM FICHADADIA", "SELECT COUNT(*) FROM erp_attendance WHERE tenant_id=$1"},
	}

	fmt.Println("\n=== POST-MIGRATION VALIDATION ===")
	allOK := true
	for _, c := range checks {
		var mysqlCount, pgCount int

		row := mysqlDB.QueryRowContext(ctx, c.MySQL)
		row.Scan(&mysqlCount)

		pgRow := pgPool.QueryRow(ctx, c.PG, tenantID)
		pgRow.Scan(&pgCount)

		status := "OK"
		if mysqlCount != pgCount {
			status = "MISMATCH"
			allOK = false
		}
		fmt.Printf("  %-25s mysql=%-6d pg=%-6d %s\n", c.Name, mysqlCount, pgCount, status)
	}

	// Financial checksum: journal balance
	var mysqlBalance, pgBalance decimal.Decimal
	mysqlDB.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(CAST(importe AS DECIMAL(16,2))),0) FROM CTB_DETALLES WHERE doh=0").Scan(&mysqlBalance)
	pgPool.QueryRow(ctx,
		"SELECT COALESCE(SUM(debit),0) FROM erp_journal_lines WHERE tenant_id=$1", tenantID).Scan(&pgBalance)

	diff := mysqlBalance.Sub(pgBalance).Abs()
	balanceStatus := "OK"
	if diff.GreaterThan(decimal.NewFromFloat(0.01)) {
		balanceStatus = "MISMATCH"
		allOK = false
	}
	fmt.Printf("  %-25s diff=%s %s\n", "journal_debit_balance", diff.StringFixed(2), balanceStatus)

	// Phase 18 — Tax entry balance (IVAIMPORTES.impimp vs erp_tax_entries.tax_amount)
	var mysqlTaxSum, pgTaxSum decimal.Decimal
	mysqlDB.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(ABS(impimp)),0) FROM IVAIMPORTES").Scan(&mysqlTaxSum)
	pgPool.QueryRow(ctx,
		"SELECT COALESCE(SUM(ABS(tax_amount)),0) FROM erp_tax_entries WHERE tenant_id=$1", tenantID).Scan(&pgTaxSum)

	taxDiff := mysqlTaxSum.Sub(pgTaxSum).Abs()
	taxStatus := "OK"
	if taxDiff.GreaterThan(decimal.NewFromFloat(0.01)) {
		taxStatus = "MISMATCH"
		allOK = false
	}
	fmt.Printf("  %-25s diff=%s %s\n", "tax_entry_balance", taxDiff.StringFixed(2), taxStatus)

	// Phase 18 — Treasury balance (CAJMOVIM absolute amounts vs erp_treasury_movements.amount)
	var mysqlTreasurySum, pgTreasurySum decimal.Decimal
	mysqlDB.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(ABS(cajimp)),0) FROM CAJMOVIM").Scan(&mysqlTreasurySum)
	pgPool.QueryRow(ctx,
		"SELECT COALESCE(SUM(ABS(amount)),0) FROM erp_treasury_movements WHERE tenant_id=$1", tenantID).Scan(&pgTreasurySum)

	treasuryDiff := mysqlTreasurySum.Sub(pgTreasurySum).Abs()
	treasuryStatus := "OK"
	if treasuryDiff.GreaterThan(decimal.NewFromFloat(0.01)) {
		treasuryStatus = "MISMATCH"
		allOK = false
	}
	fmt.Printf("  %-25s diff=%s %s\n", "treasury_balance", treasuryDiff.StringFixed(2), treasuryStatus)

	if allOK {
		fmt.Println("\n  RESULT: ALL CHECKS PASSED")
	} else {
		fmt.Println("\n  RESULT: SOME CHECKS FAILED — review mismatches")
	}

	return nil
}

// PrintPreValidationReport prints the pre-validation report to stdout.
func PrintPreValidationReport(report *ValidationReport) {
	fmt.Printf("\n=== PRE-VALIDATION REPORT ===\n")
	fmt.Printf("Run ID: %s\n", report.RunID)
	fmt.Printf("Total issues: %d\n", report.TotalCount)
	fmt.Printf("  auto_fix:   %d (mapped automatically in transformer)\n", report.AutoFix)
	fmt.Printf("  skip:       %d (rows skipped in writer)\n", report.Skip)
	fmt.Printf("  fix_manual: %d (block prod run)\n", report.FixManual)

	if report.FixManual > 0 {
		fmt.Println("\nBlocking issues (fix_manual, need operator action):")
		for _, i := range report.Issues {
			if i.Resolution == "fix_manual" {
				fmt.Printf("  - %s#%d: %s (%s)\n", i.LegacyTable, i.LegacyID, i.Constraint, i.Detail)
			}
		}
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Fix blocking issues in MySQL legacy")
		fmt.Println("  2. Re-run: sda migrate-legacy --dry-run --tenant=<slug>")
		fmt.Println("  3. When report says 'all clear', run prod")
	} else {
		fmt.Println("\n  ALL CLEAR — no blocking issues. Ready for prod run.")
	}
}
