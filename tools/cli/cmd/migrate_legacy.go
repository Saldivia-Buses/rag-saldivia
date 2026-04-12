package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

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
	migrateLegacyCmd.Flags().Int("mysql-rps", 0, "Rate limit MySQL reads (rows per second, 0=unlimited)")

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
		pgDSN = env("POSTGRES_PLATFORM_URL", "postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable")
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
			return fmt.Errorf("tenant %q requires --dry-run first, or add to --skip-dry-run-for", tenantSlug)
		}
		lastStatus, err := migration.FindLastDryRun(ctx, pgPool, tenantSlug)
		if err != nil || lastStatus != "dry_run_ok" {
			return fmt.Errorf("no dry_run_ok found for tenant %q — run --dry-run first", tenantSlug)
		}
	}

	// --- Build orchestrator ---
	orch := migration.NewOrchestrator(mysqlDB, pgPool, tenantSlug)

	// Register all migrators in FK dependency order
	registerMigrators(orch, mysqlDB, tenantSlug)

	// --- Pre-validation (dry-run only) ---
	if dryRun {
		runID := uuid.New()
		pgPool.Exec(ctx,
			`INSERT INTO erp_migration_runs (id, tenant_id, mode) VALUES ($1, $2, 'dry_run')`,
			runID, tenantSlug,
		)

		report, err := migration.RunPreValidation(ctx, mysqlDB, pgPool, tenantSlug, runID)
		if err != nil {
			return fmt.Errorf("pre-validation: %w", err)
		}
		migration.PrintPreValidationReport(report)

		if report.FixManual > 0 {
			pgPool.Exec(ctx,
				`UPDATE erp_migration_runs SET status = 'dry_run_failed', completed_at = now() WHERE id = $1`,
				runID,
			)
			return fmt.Errorf("%d blocking issues found — fix them and re-run", report.FixManual)
		}

		// Continue with dry-run migration (test transformations)
		fmt.Println("\nPre-validation passed. Running dry-run migration...")
	}

	// --- Resume ---
	if resume {
		if resumeRunID == "" {
			return fmt.Errorf("--resume-run-id required with --resume")
		}
		return orch.Resume(ctx, resumeRunID)
	}

	// --- Run migration ---
	if err := orch.Run(ctx, domains, dryRun); err != nil {
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
	// Hide password in DSN for logging
	if idx := strings.Index(dsn, ":"); idx > 0 {
		if end := strings.Index(dsn[idx:], "@"); end > 0 {
			return dsn[:idx+1] + "***" + dsn[idx+end:]
		}
	}
	return dsn
}
