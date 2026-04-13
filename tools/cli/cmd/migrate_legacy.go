package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/Camionerou/rag-saldivia/tools/cli/internal/legacy"
	"github.com/Camionerou/rag-saldivia/tools/cli/internal/migration"
)

var migrateLegacyCmd = &cobra.Command{
	Use:   "migrate-legacy",
	Short: "Migrar datos de Histrix (MySQL) a SDA (PostgreSQL)",
	Long: `Migra datos historicos de Histrix (MySQL) a SDA (PostgreSQL).

MySQL es read-only. IDs se mapean de INT a UUID. La migracion es idempotente
(INSERT ON CONFLICT DO NOTHING + tabla de mapping).

Ejemplos:
  sda migrate-legacy --dry-run --tenant=saldivia
  sda migrate-legacy --tenant=saldivia --skip-dry-run-for=saldivia
  sda migrate-legacy --tenant=saldivia --domains=catalog,entity
  sda migrate-legacy --resume --resume-run-id=<uuid> --tenant=saldivia`,
	RunE: runMigrateLegacy,
}

func init() {
	migrateLegacyCmd.Flags().String("mysql-dsn", "", "MySQL DSN (e.g., user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=true)")
	migrateLegacyCmd.Flags().String("pg-dsn", "", "PostgreSQL DSN (e.g., postgres://user:pass@host:5432/db?sslmode=disable)")
	migrateLegacyCmd.Flags().String("tenant", "", "Tenant slug to migrate")
	migrateLegacyCmd.Flags().StringSlice("domains", nil, "Domains to migrate (empty=all)")
	migrateLegacyCmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no INSERTs, validates transformations)")
	migrateLegacyCmd.Flags().Bool("resume", false, "Resume a failed migration run")
	migrateLegacyCmd.Flags().String("resume-run-id", "", "UUID of the run to resume")
	migrateLegacyCmd.Flags().StringSlice("skip-dry-run-for", nil, "Tenants allowed to skip dry-run requirement")
	migrateLegacyCmd.Flags().Bool("validate-only", false, "Run post-migration validation without migrating")

	migrateLegacyCmd.MarkFlagRequired("tenant")
}

func runMigrateLegacy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	tenantSlug, _ := cmd.Flags().GetString("tenant")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	resume, _ := cmd.Flags().GetBool("resume")
	resumeRunID, _ := cmd.Flags().GetString("resume-run-id")
	domains, _ := cmd.Flags().GetStringSlice("domains")
	validateOnly, _ := cmd.Flags().GetBool("validate-only")

	// --- Connect to databases ---
	mysqlDSN, _ := cmd.Flags().GetString("mysql-dsn")
	if mysqlDSN == "" {
		mysqlDSN = os.Getenv("MYSQL_LEGACY_DSN")
	}
	if mysqlDSN == "" {
		return fmt.Errorf("--mysql-dsn or MYSQL_LEGACY_DSN required")
	}

	pgDSN, _ := cmd.Flags().GetString("pg-dsn")
	if pgDSN == "" {
		pgDSN = os.Getenv("POSTGRES_TENANT_URL")
	}
	if pgDSN == "" {
		return fmt.Errorf("--pg-dsn or POSTGRES_TENANT_URL required")
	}

	slog.Info("connecting to MySQL", "dsn", maskDSN(mysqlDSN))
	mysqlDB, err := legacy.Connect(mysqlDSN)
	if err != nil {
		return fmt.Errorf("mysql connect: %w", err)
	}
	defer mysqlDB.Close()

	slog.Info("connecting to PostgreSQL", "dsn", maskDSN(pgDSN))
	pgPool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		return fmt.Errorf("pg connect: %w", err)
	}
	defer pgPool.Close()

	// --- Validate-only mode ---
	if validateOnly {
		return migration.PostMigrationValidation(ctx, mysqlDB, pgPool, tenantSlug)
	}

	// --- Dry-run enforcement ---
	if !dryRun && !resume {
		allowList, _ := cmd.Flags().GetStringSlice("skip-dry-run-for")
		if !containsStr(allowList, tenantSlug) {
			lastStatus, err := migration.FindLastDryRun(ctx, pgPool, tenantSlug)
			if err != nil || lastStatus != "dry_run_ok" {
				return fmt.Errorf("no dry_run_ok found for tenant %q — run --dry-run first, or add to --skip-dry-run-for", tenantSlug)
			}
		}
	}

	// --- Build orchestrator ---
	orch := migration.NewOrchestrator(mysqlDB, pgPool, tenantSlug)

	// Register all migrators in FK dependency order
	registerMigrators(orch, mysqlDB, tenantSlug)

	// --- Pre-validation (dry-run only) ---
	var dryRunID uuid.UUID
	if dryRun {
		dryRunID = uuid.New()
		pgPool.Exec(ctx,
			`INSERT INTO erp_migration_runs (id, tenant_id, mode) VALUES ($1, $2, 'dry_run')`,
			dryRunID, tenantSlug,
		)

		report, err := migration.RunPreValidation(ctx, mysqlDB, pgPool, tenantSlug, dryRunID)
		if err != nil {
			return fmt.Errorf("pre-validation: %w", err)
		}
		migration.PrintPreValidationReport(report)

		if report.FixManual > 0 {
			pgPool.Exec(ctx,
				`UPDATE erp_migration_runs SET status = 'dry_run_failed', completed_at = now() WHERE id = $1`,
				dryRunID,
			)
			return fmt.Errorf("%d blocking issues found — fix them and re-run", report.FixManual)
		}

		fmt.Println("\nPre-validation passed. Running dry-run migration...")
	}

	// --- Resume ---
	if resume {
		if resumeRunID == "" {
			return fmt.Errorf("--resume-run-id required with --resume")
		}
		return orch.Resume(ctx, resumeRunID)
	}

	// --- Run migration (pass dryRunID so orchestrator reuses it instead of creating duplicate) ---
	if err := orch.RunWithID(ctx, dryRunID, domains, dryRun); err != nil {
		return err
	}

	// --- Post-migration validation (prod mode) ---
	if !dryRun {
		fmt.Println("\nRunning post-migration validation...")
		migration.PostMigrationValidation(ctx, mysqlDB, pgPool, tenantSlug)
	}

	return nil
}

// registerMigrators adds all table migrators in FK dependency order.
func registerMigrators(orch *migration.Orchestrator, mysqlDB *sql.DB, tenantID string) {
	// Phase 2: Catalogs + Entities (no FK dependencies)
	for _, cm := range legacy.DefaultCatalogMappings() {
		orch.RegisterMigrators(migration.NewCatalogMigrator(mysqlDB, cm, tenantID))
	}
	orch.RegisterMigrators(
		migration.NewEntityMigrator(mysqlDB, tenantID),
		migration.NewEmployeeMigrator(mysqlDB, tenantID),
	)

	// Phase 3: Financial (depends on entities + catalogs)
	orch.RegisterMigrators(
		migration.NewCostCenterMigrator(mysqlDB, tenantID),
		migration.NewAccountMigrator(mysqlDB, tenantID),
		migration.NewFiscalYearMigrator(mysqlDB, tenantID),
		migration.NewJournalEntryMigrator(mysqlDB, tenantID),
	)

	// Setup hook: build code index for accounts (CTB_CUENTAS FK is by varchar code)
	// and entry date cache (for journal line date resolution).
	// These run after accounts + journal entries are migrated but before journal lines.
	orch.AddSetupHook(func(ctx context.Context, mapper *migration.Mapper) error {
		if err := mapper.BuildCodeIndex(ctx, "accounting", "erp_accounts", "code"); err != nil {
			return fmt.Errorf("build account code index: %w", err)
		}
		if err := mapper.BuildEntryDateCache(ctx); err != nil {
			return fmt.Errorf("build entry date cache: %w", err)
		}
		return nil
	})

	orch.RegisterMigrators(
		migration.NewJournalLineMigrator(mysqlDB, tenantID),
		migration.NewBankAccountMigrator(mysqlDB, tenantID),
		migration.NewCashRegisterMigrator(mysqlDB, tenantID),
		migration.NewCheckMigrator(mysqlDB, tenantID),
	)

	// Phase 4: Operational (depends on entities + stock)
	orch.RegisterMigrators(
		migration.NewWarehouseMigrator(mysqlDB, tenantID),
		migration.NewArticleMigrator(mysqlDB, tenantID),
		migration.NewStockMovementMigrator(mysqlDB, tenantID),
		migration.NewPurchaseOrderMigrator(mysqlDB, tenantID),
		migration.NewPurchaseOrderLineMigrator(mysqlDB, tenantID),
		migration.NewQuotationMigrator(mysqlDB, tenantID),
		migration.NewProductionCenterMigrator(mysqlDB, tenantID),
		migration.NewProductionOrderMigrator(mysqlDB, tenantID),
	)

	// Phase 5: HR (depends on entities)
	orch.RegisterMigrators(
		migration.NewEmployeeDetailMigrator(mysqlDB, tenantID),
		migration.NewAbsenceMigrator(mysqlDB, tenantID),
		migration.NewTrainingMigrator(mysqlDB, tenantID),
	)

	// Phase 6: Invoicing — headers (depends on entities)
	// IVAVENTAS/IVACOMPRAS are the invoice masters, NOT FACREMIT.
	orch.RegisterMigrators(
		migration.NewSalesInvoiceMigrator(mysqlDB, tenantID),
		migration.NewPurchaseInvoiceMigrator(mysqlDB, tenantID),
	)

	// Setup hook: build regmovim_id → invoice UUID index (FACDETAL needs it).
	// Must run AFTER IVAVENTAS + IVACOMPRAS migration, BEFORE FACDETAL.
	orch.AddSetupHook(func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildRegMovimIndex(ctx, mysqlDB)
	})

	// Phase 6b: Invoice lines + delivery notes (depends on invoice headers)
	orch.RegisterMigrators(
		migration.NewInvoiceLineMigrator(mysqlDB, tenantID),
		migration.NewDeliveryNoteMigrator(mysqlDB, tenantID),
		migration.NewDeliveryNoteAltMigrator(mysqlDB, tenantID),
		migration.NewDeliveryNoteLineMigrator(mysqlDB, tenantID),
	)

	// Phase 7: Tax entries from IVAIMPORTES (depends on invoice headers)
	orch.RegisterMigrators(
		migration.NewTaxEntrySalesMigrator(mysqlDB, tenantID),
		migration.NewTaxEntryPurchasesMigrator(mysqlDB, tenantID),
	)

	// Phase 8: Withholdings (depends on entities)
	orch.RegisterMigrators(
		migration.NewWithholdingGainsMigrator(mysqlDB, tenantID),
		migration.NewWithholdingIVAMigrator(mysqlDB, tenantID),
		migration.NewWithholding1598Migrator(mysqlDB, tenantID),
	)

	// Phase 9: Stock extended — BOM, stock levels, price lists
	orch.RegisterMigrators(
		migration.NewBOMMigrator(mysqlDB, tenantID),
		migration.NewBOMHistoryMigrator(mysqlDB, tenantID),
		migration.NewStockLevelMigrator(mysqlDB, tenantID),
		migration.NewPriceListMigrator(mysqlDB, tenantID),
		migration.NewPriceListItemMigrator(mysqlDB, tenantID),
	)

	// Phase 10: Purchasing extended — internal requisitions, receipts
	orch.RegisterMigrators(
		migration.NewInternalRequisitionMigrator(mysqlDB, tenantID),
		migration.NewPurchaseReceiptMigrator(mysqlDB, tenantID),
	)

	// Phase 11: Sales & Production extended
	orch.RegisterMigrators(
		migration.NewQuotationLineMigrator(mysqlDB, tenantID),
		migration.NewCustomerOrderMigrator(mysqlDB, tenantID),
		migration.NewVehicleMigrator(mysqlDB, tenantID),
		migration.NewProductionRequestMigrator(mysqlDB, tenantID),
		migration.NewProductionStepMigrator(mysqlDB, tenantID),
		migration.NewProductionInspectionMigrator(mysqlDB, tenantID),
	)

	// Phase 12: Quality (ISO 9001)
	orch.RegisterMigrators(
		migration.NewNonconformityMigrator(mysqlDB, tenantID),
		migration.NewCorrectiveActionMigrator(mysqlDB, tenantID),
		migration.NewAuditMigrator(mysqlDB, tenantID),
		migration.NewAuditFindingMigrator(mysqlDB, tenantID),
		migration.NewControlledDocumentMigrator(mysqlDB, tenantID),
	)

	// Phase 13: HR extended — departments, attendance, events
	orch.RegisterMigrators(
		migration.NewDepartmentMigrator(mysqlDB, tenantID),
		migration.NewAttendanceMigrator(mysqlDB, tenantID),
		migration.NewTrainingAttendeeMigrator(mysqlDB, tenantID),
		migration.NewDemeritEventMigrator(mysqlDB, tenantID),
		migration.NewDeductionEventMigrator(mysqlDB, tenantID),
		migration.NewAdditionalPayEventMigrator(mysqlDB, tenantID),
	)

	// Phase 14: Maintenance & Fleet
	orch.RegisterMigrators(
		migration.NewMaintenanceAssetMigrator(mysqlDB, tenantID),
		migration.NewMaintenancePlanMigrator(mysqlDB, tenantID),
		migration.NewMaintenanceEventMigrator(mysqlDB, tenantID),
		migration.NewVehicleWorkMigrator(mysqlDB, tenantID),
		migration.NewFuelLogMigrator(mysqlDB, tenantID),
	)

	// Phase 15: Safety — risk agents first (catalog), then exposures
	orch.RegisterMigrators(
		migration.NewRiskAgentMigrator(mysqlDB, tenantID),
		migration.NewAccidentMigrator(mysqlDB, tenantID),
		migration.NewRiskExposureMigrator(mysqlDB, tenantID),
		migration.NewMedicalLeaveMigrator(mysqlDB, tenantID),
	)

	// Phase 16: Auth — legacy users
	orch.RegisterMigrators(
		migration.NewLegacyUserMigrator(mysqlDB, tenantID),
	)
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func maskDSN(dsn string) string {
	// Hide password in DSN for logging.
	// Handles both URI (postgres://user:pass@host) and MySQL (user:pass@tcp(host)) formats.
	re := regexp.MustCompile(`(://[^:]+:|^[^:]+:)([^@]+)(@)`)
	masked := re.ReplaceAllString(dsn, "${1}***${3}")
	if masked != dsn {
		return masked
	}
	return dsn
}
