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
	migrateLegacyCmd.Flags().Int("batch-size", migration.DefaultBatchSize,
		"Rows per batch. Auto-clamped per table to stay under PG's 65535 bind-param limit.")

	// Fast-path flags. Default ("insert") preserves existing behaviour: the
	// legacy multi-value INSERT writer and the sequential table runner.
	//
	//   --writer=copy              → COPY + staging + read/write pipeline
	//   --writer=parallel-copy     → parallel COPY workers + pipeline
	//   --copy-workers=N           → worker count for parallel-copy (default 4)
	//
	// Empirical speedups on a local pg16 + 100K-row bom_history batch:
	//   INSERT baseline            1.00x (36K rows/sec)
	//   COPY + pipeline            1.80x (77K rows/sec)
	//   ParallelCopy×8 + pipeline  1.98x (85K rows/sec)
	//
	// Writer-only (no pipeline, no read latency), parallel COPY hits 2.90x.
	migrateLegacyCmd.Flags().String("writer", "insert", "Write strategy: insert | copy | parallel-copy")
	migrateLegacyCmd.Flags().Int("copy-workers", 4, "Parallel COPY workers when --writer=parallel-copy")

	// --archive: after the normal migration, scan every legacy MySQL table
	// not already owned by a first-class migrator (and not HTX/sabredav/asterisk
	// system noise) and park its rows in erp_legacy_archive as JSONB. Gives
	// the "no silent data loss" guarantee.
	migrateLegacyCmd.Flags().Bool("archive", false, "Archive every non-migrated, non-HTX legacy table into erp_legacy_archive as JSONB")
	migrateLegacyCmd.Flags().Bool("archive-only", false, "Skip the structured migration; only run the universal JSONB archive over non-migrated tables")

	// --archive-skips: whenever a migrator transform rejects a row (returns
	// nil because of a zero PK, missing FK that no ghost covered, etc.) the
	// pipeline writes that raw legacy row to erp_legacy_archive in the same
	// tx as the batch. Combined with --archive this achieves truly zero
	// silent data loss: every row from MySQL either lands in a structured
	// SDA table or in the archive.
	migrateLegacyCmd.Flags().Bool("archive-skips", false, "Archive rows the transform rejected into erp_legacy_archive (requires pipeline writer)")

	// Transform fan-out: when the writer can't keep the CPU busy (workstation
	// sat at 5% CPU with 24 writer workers), split the per-row Transform loop
	// across N goroutines. Keeps the structure single-reader so MySQL load is
	// unchanged; parallelises the CPU-bound transform step only.
	migrateLegacyCmd.Flags().Int("transform-workers", 1, "Parallel goroutines per batch for the Transform step (0/1 = sequential)")
	migrateLegacyCmd.Flags().Int("read-workers", 1, "Parallel MySQL readers per table via PK range partitioning (GenericReader-backed tables only; disables per-table resume when >1)")

	if err := migrateLegacyCmd.MarkFlagRequired("tenant"); err != nil {
		panic(err)
	}
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
	defer func() { _ = mysqlDB.Close() }()

	slog.Info("connecting to PostgreSQL", "dsn", maskDSN(pgDSN))
	pgCfg, err := pgxpool.ParseConfig(pgDSN)
	if err != nil {
		return fmt.Errorf("parse pg dsn: %w", err)
	}
	// Bulk import can fan out up to parallel-copy workers plus pipeline
	// overhead. Default pgxpool cap of 4 serializes the whole rig.
	pgCfg.MaxConns = 48
	pgCfg.MinConns = 8
	pgPool, err := pgxpool.NewWithConfig(ctx, pgCfg)
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
	if bs, _ := cmd.Flags().GetInt("batch-size"); bs > 0 {
		orch.SetBatchSize(bs)
	}

	// Fast-path writer selection. Pipeline+COPY are opt-in so existing
	// operator muscle memory keeps working.
	switch mode, _ := cmd.Flags().GetString("writer"); mode {
	case "copy":
		orch.UseCopyWriter()
		slog.Info("using COPY writer + pipeline")
	case "parallel-copy":
		workers, _ := cmd.Flags().GetInt("copy-workers")
		orch.UseParallelCopyWriter(workers)
		slog.Info("using parallel COPY writer + pipeline", "workers", workers)
	case "insert", "":
		// default — no change
	default:
		return fmt.Errorf("unknown --writer=%q (want insert|copy|parallel-copy)", mode)
	}
	if archiveSkips, _ := cmd.Flags().GetBool("archive-skips"); archiveSkips {
		orch.EnableSkipArchive()
		slog.Info("archiving skipped rows into erp_legacy_archive")
	}
	if tw, _ := cmd.Flags().GetInt("transform-workers"); tw > 1 {
		orch.SetTransformWorkers(tw)
		slog.Info("parallel transform enabled", "workers", tw)
	}
	if rw, _ := cmd.Flags().GetInt("read-workers"); rw > 1 {
		orch.SetReadWorkers(rw)
		slog.Info("multi-reader enabled", "workers", rw)
	}

	// Seed dry-run flag before initialising fallback so EnsureUnassignedWarehouse
	// uses the deterministic path instead of hitting PG.
	orch.GetMapper().SetDryRun(dryRun)

	// Fallback warehouse for STK_MOVIMIENTOS rows with stkdeposito_id=0 or
	// orphaned depot FKs. Must exist before the stock phase runs; called here
	// so both prod and dry-run (and resume) share the same code path.
	if err := orch.GetMapper().EnsureUnassignedWarehouse(ctx, pgPool); err != nil {
		return fmt.Errorf("init unassigned warehouse: %w", err)
	}
	// Fallback entity for rows whose ctacod/prvcod doesn't resolve in either
	// id_regcuenta or nro_cuenta. Required by RETACUMU, MOVDEMERITO, IVA*,
	// anything that carries the ambiguous Histrix entity code.
	if err := orch.GetMapper().EnsureUnknownEntity(ctx, pgPool); err != nil {
		return fmt.Errorf("init unknown entity: %w", err)
	}

	// Register all migrators in FK dependency order
	registerMigrators(orch, mysqlDB, tenantSlug)

	// --- Archive-only fast path ---
	// Skips the entire structured migration and goes straight to the
	// universal JSONB archive. Useful when the structured migration is
	// already complete (e.g. after a successful --resume) and you just
	// want to ship the non-modeled legacy tables.
	if archiveOnly, _ := cmd.Flags().GetBool("archive-only"); archiveOnly {
		fmt.Println("Archive-only mode: skipping structured migration, running universal archive...")
		covered := migration.CoveredLegacyTables(orch)
		if err := migration.ArchiveAll(ctx, mysqlDB, pgPool, tenantSlug, covered); err != nil {
			return fmt.Errorf("archive-only: %w", err)
		}
		fmt.Println("Archive-only done.")
		return nil
	}

	// --- Pre-validation (dry-run only) ---
	var dryRunID uuid.UUID
	if dryRun {
		dryRunID = uuid.New()
		if _, err := pgPool.Exec(ctx,
			`INSERT INTO erp_migration_runs (id, tenant_id, mode) VALUES ($1, $2, 'dry_run')`,
			dryRunID, tenantSlug,
		); err != nil {
			slog.Warn("could not record dry-run start", "err", err)
		}

		report, err := migration.RunPreValidation(ctx, mysqlDB, pgPool, tenantSlug, dryRunID)
		if err != nil {
			return fmt.Errorf("pre-validation: %w", err)
		}
		migration.PrintPreValidationReport(report)

		if report.FixManual > 0 {
			if _, err := pgPool.Exec(ctx,
				`UPDATE erp_migration_runs SET status = 'dry_run_failed', completed_at = now() WHERE id = $1`,
				dryRunID,
			); err != nil {
				slog.Warn("could not record dry-run failure", "err", err)
			}
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

	// --- Phase 17: Metadata enrichment (prod mode only) ---
	if !dryRun {
		fmt.Println("\nRunning Phase 17: metadata enrichment...")
		if err := migration.RunMetadataEnrichment(ctx, mysqlDB, pgPool, tenantSlug, orch.GetMapper()); err != nil {
			return fmt.Errorf("metadata enrichment: %w", err)
		}
		fmt.Println("Metadata enrichment completed.")
	}

	// --- Post-migration validation (prod mode) ---
	if !dryRun {
		fmt.Println("\nRunning post-migration validation...")
		if err := migration.PostMigrationValidation(ctx, mysqlDB, pgPool, tenantSlug); err != nil {
			slog.Warn("post-migration validation reported issues", "err", err)
		}
	}

	// --- Universal archive (opt-in, prod only) ---
	if !dryRun {
		if archive, _ := cmd.Flags().GetBool("archive"); archive {
			fmt.Println("\nArchiving every non-migrated, non-HTX legacy table → erp_legacy_archive ...")
			covered := migration.CoveredLegacyTables(orch)
			if err := migration.ArchiveAll(ctx, mysqlDB, pgPool, tenantSlug, covered); err != nil {
				slog.Warn("archive phase had issues", "err", err)
			}
		}
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

	// After-table hook: PERSONAL.legajo → entity UUID index.
	// Consumed by HR tables (FICHADADIA, RRHH_DESCUENTOS, RRHH_ADICIONALES) and
	// production inspections (legajo_realizo / legajo_personal in PROD_CONTROL_MOVIM).
	orch.AddAfterTableHook("PERSONAL", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildLegajoIndex(ctx, mysqlDB)
	})

	// After-table hook: REG_CUENTA.nro_cuenta → entity UUID index.
	// Consumed by RETACUMU / MOVDEMERITO / IVA* (ctacod ambiguity fallback).
	orch.AddAfterTableHook("REG_CUENTA", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildNroCuentaIndex(ctx, mysqlDB)
	})

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

	// Phase 3c: Legacy accounting line log (CTBREGIS → erp_accounting_registers).
	// Pareto #3 of the Phase 1 §Data migration gap post-2.0.9 (~604 K rows
	// live, ~28 % of remaining row volume). CTBREGIS is the pre-CTB_MOVIMIENTOS
	// debe/haber log; still read and written by live Histrix UI (libro_diario,
	// proveedores_loc, clientes_local, ordenpago, iva, anulaciones, etc.).
	// Runs here so the accounting code index (from AddSetupHook above) is
	// already loaded when each row resolves its ctbcod → erp_accounts.id.
	orch.RegisterMigrators(
		migration.NewAccountingRegisterMigrator(mysqlDB, tenantID),
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

	// Phase 4b: Tools (HERRAMIENTAS + HERRMOVS → erp_tools + erp_tool_movements).
	// Pareto #4 of the Phase 1 §Data migration gap post-2.0.9 (~400 K rows
	// combined). Despite the "herramientas" naming, HERRAMIENTAS is the
	// serialized inventory tag ledger (389 K items received with per-unit
	// barcode codes); HERRMOVS is the lending ledger (11.6 K check-out /
	// return entries). HERRAMIENTAS must run first so the AfterTableHook
	// can build the tools code index before HERRMOVS resolves its
	// tool_code → erp_tools.id.
	orch.RegisterMigrators(
		migration.NewToolMigrator(mysqlDB, tenantID),
	)
	orch.AddAfterTableHook("HERRAMIENTAS", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildCodeIndex(ctx, "tools", "erp_tools", "code")
	})
	orch.RegisterMigrators(
		migration.NewToolMovementMigrator(mysqlDB, tenantID),
	)

	// Phase 4c: Per-supplier article cost ledger (STKINSPR → erp_article_costs).
	// Pareto #5 of the Phase 1 §Data migration gap post-2.0.10 (~190 K rows).
	// Depends on both the stock article index (built by NewArticleMigrator above
	// in Phase 4) and the entity / nro_cuenta index (built by Phase 2's REG_CUENTA
	// + AddAfterTableHook("REG_CUENTA", BuildNroCuentaIndex)). No new hooks needed.
	orch.RegisterMigrators(
		migration.NewArticleSupplierCostMigrator(mysqlDB, tenantID),
	)

	// Phase 4e: Article cost history (STK_COSTO_HIST → erp_article_cost_history).
	// Pareto #8 (~104 K rows). Composite natural PK (article_code, year, month)
	// preserved via hashCode. Depends on the stock article index from Phase 4.
	orch.RegisterMigrators(
		migration.NewArticleCostHistoryMigrator(mysqlDB, tenantID),
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

	// After-table hook: build regmovim_id → invoice UUID index (FACDETAL needs it).
	// Fires right after IVACOMPRAS (the second of IVAVENTAS/IVACOMPRAS), so both
	// invoice masters are already cached in the mapper before FACDETAL runs.
	//
	// Previously registered via AddSetupHook — but setup hooks all fire together
	// at the first migrator that needs one (NewJournalLineMigrator, much earlier
	// in the phase order), which meant the regmovim index was built BEFORE the
	// invoice masters existed. The index came out empty and every FACDETAL row
	// (198K in prod saldivia) was silently skipped for missing parent invoice.
	orch.AddAfterTableHook("IVACOMPRAS", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildRegMovimIndex(ctx, mysqlDB)
	})

	// Phase 6b: Invoice lines + delivery notes (depends on invoice headers).
	// REMITO + REMITOINT must run before REMDETAL so both indexes are built in-between.
	orch.RegisterMigrators(
		migration.NewInvoiceLineMigrator(mysqlDB, tenantID),
		migration.NewDeliveryNoteMigrator(mysqlDB, tenantID),
		migration.NewDeliveryNoteAltMigrator(mysqlDB, tenantID),
		migration.NewInternalDeliveryNoteMigrator(mysqlDB, tenantID),
	)

	// After-table hook: REMITO.idRemito → REMITO UUID index (consumed by REMDETAL).
	orch.AddAfterTableHook("REMITO", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildRemitoIndex(ctx, mysqlDB)
	})

	// After-table hook: REMITOINT.idRemito → REMITOINT UUID index (consumed by REMDETAL).
	// Closes W-001: REMDETAL.idRemito references REMITOINT (not REMITO) on saldivia.
	orch.AddAfterTableHook("REMITOINT", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildRemitoIntIndex(ctx)
	})

	orch.RegisterMigrators(
		migration.NewDeliveryNoteLineMigrator(mysqlDB, tenantID),
	)

	// Phase 6c: Bank-statement imports (BCS_IMPORTACION → erp_bank_imports,
	// ~92 K rows live). Rank 1 of the remaining post-Pareto-#8 long tail.
	// Depends on BuildRegMovimIndex (Phase 6 AfterTableHook on IVACOMPRAS,
	// already fired above) + BuildNroCuentaIndex (Phase 2 hook). No new
	// indexes needed.
	orch.RegisterMigrators(
		migration.NewBankImportMigrator(mysqlDB, tenantID),
	)

	// Phase 7: Treasury movements — cash + bank + cash counts (depends on treasury catalogs)
	orch.RegisterMigrators(
		migration.NewCashMovementMigrator(mysqlDB, tenantID),
		migration.NewBankMovementMigrator(mysqlDB, tenantID),
		migration.NewCashCountMigrator(mysqlDB, tenantID),
	)

	// Phase 7b: Tax entries from IVAIMPORTES (depends on invoice headers)
	orch.RegisterMigrators(
		migration.NewTaxEntrySalesMigrator(mysqlDB, tenantID),
		migration.NewTaxEntryPurchasesMigrator(mysqlDB, tenantID),
	)

	// Phase 8: Current Accounts (depends on entities + treasury movements)
	orch.RegisterMigrators(
		migration.NewAccountMovementMigrator(mysqlDB, tenantID),
	)
	// CCTIMPUT references REG_MOVIMIENTOS legacy IDs that sometimes point at
	// records Histrix deleted. Before running the payment-allocation migrator,
	// seed ghost account_movements so orphan regmovim0_id / regmovim_id values
	// resolve instead of dropping the allocation.
	orch.AddAfterTableHook("REG_MOVIMIENTOS", func(ctx context.Context, mapper *migration.Mapper) error {
		return migration.RescueCCTIMPUTOrphanMovements(ctx, mysqlDB, orch.PGPool(), tenantID, mapper)
	})
	orch.RegisterMigrators(
		migration.NewPaymentAllocationMigrator(mysqlDB, tenantID),
	)

	// Phase 8b: Withholdings (depends on entities)
	orch.RegisterMigrators(
		migration.NewWithholdingGainsMigrator(mysqlDB, tenantID),
		migration.NewWithholdingIVAMigrator(mysqlDB, tenantID),
		migration.NewWithholding1598Migrator(mysqlDB, tenantID),
		migration.NewWithholdingIIBBMigrator(mysqlDB, tenantID),
	)

	// Phase 8c: Pareto tail Grupo A (2.0.11) — REG_CUENTA_CALIFICACION +
	// REG_MOVIMIENTO_OBS + CARCHEHI (~237 K rows combined). All three
	// resolve against indexes built earlier:
	//   - EntityCreditRating → ResolveEntityFlexible (id_regcuenta cache,
	//     nro_cuenta fallback) — both populated by Phase 2 hooks.
	//   - InvoiceNote → ResolveRegMovim (Phase 6 IVACOMPRAS hook).
	//   - CheckHistory → ResolveEntityFlexible via ctacod.
	// No new hooks needed; indexes are ready by this phase.
	orch.RegisterMigrators(
		migration.NewEntityCreditRatingMigrator(mysqlDB, tenantID),
		migration.NewInvoiceNoteMigrator(mysqlDB, tenantID),
		migration.NewCheckHistoryMigrator(mysqlDB, tenantID),
	)

	// Before the BOM history phase: seed ghost articles for orphan pieza_ids
	// so the 2.5M+ rows whose STKPIEZA parent no longer exists can still land
	// in SDA. See rescue.go for rationale + empirical numbers.
	orch.AddAfterTableHook("STKPIEZA", func(ctx context.Context, mapper *migration.Mapper) error {
		if err := migration.RescueBOMOrphanParents(ctx, mysqlDB, orch.PGPool(), tenantID, mapper); err != nil {
			return fmt.Errorf("rescue BOM orphan parents: %w", err)
		}
		return migration.RescueBOMOrphanChildren(ctx, mysqlDB, orch.PGPool(), tenantID, mapper)
	})

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
		migration.NewProductionInspectionDetailMigrator(mysqlDB, tenantID),
	)

	// Phase 11b: Homologations — vehicle model homologations + their cost/process
	// revision history. STK_ARTICULO_PROCESO_HIST_DETALLE is 2.6 M rows (Pareto #1
	// of the Phase 1 §Data migration gap, 42.7 % of uncovered row volume).
	// Order: HOMOLOGMOD (parent) → STK_ARTICULO_PROCESO_HIST (revisions) →
	// STK_ARTICULO_PROCESO_HIST_DETALLE (lines). Detail line resolves article_id
	// through the erp_articles cache produced by NewArticleMigrator in Phase 4.
	orch.RegisterMigrators(
		migration.NewHomologationMigrator(mysqlDB, tenantID),
		migration.NewHomologationRevisionMigrator(mysqlDB, tenantID),
		migration.NewHomologationRevisionLineMigrator(mysqlDB, tenantID),
	)

	// Phase 11c: Products domain (PRODUCTO_* cluster). Pareto #6 of the
	// Phase 1 §Data migration gap (PRODUCTO_ATRIB_VALORES, 354 K rows) plus
	// Pareto #18 (PRODUCTO_ATRIBUTO_HOMOLOGACION, 47 K rows, needs the
	// Phase 11b HOMOLOGMOD cache). Order: sections → products →
	// attributes → options → values → atrib_homologations, each parent
	// before children so ResolveOptional finds the mapped UUID.
	orch.RegisterMigrators(
		migration.NewProductSectionMigrator(mysqlDB, tenantID),
		migration.NewProductMigrator(mysqlDB, tenantID),
		migration.NewProductAttributeMigrator(mysqlDB, tenantID),
		migration.NewProductAttributeOptionMigrator(mysqlDB, tenantID),
		migration.NewProductAttributeValueMigrator(mysqlDB, tenantID),
		migration.NewProductAttributeHomologationMigrator(mysqlDB, tenantID),
	)

	// Phase 11d: Production inspection × homologation cross-reference
	// (PROD_CONTROL_HOMOLOG → erp_production_inspection_homologations).
	// Pareto #7 (~403 K rows live — scrape was 106 K, +282 % growth).
	// Depends on both PROD_CONTROLES (Phase 7/8) and HOMOLOGMOD
	// (Phase 11b) caches being populated. 0 orphans confirmed live.
	orch.RegisterMigrators(
		migration.NewProductionInspectionHomologationMigrator(mysqlDB, tenantID),
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

	// Phase 13b: Time clock — raw clock-punch events (FICHADAS) + the
	// card-to-employee assignment table they resolve through. Pareto #2 of
	// the Phase 1 §Data migration gap post-2.0.8 (1.46 M rows, 41 % of the
	// remaining uncovered row volume). PERSONAL_TARJETA must run before
	// FICHADAS so BuildTarjetaIndex can see every assignment; FICHADAS uses
	// ResolveByTarjetaAtDate to pick the right employee for each event date.
	orch.RegisterMigrators(
		migration.NewEmployeeCardMigrator(mysqlDB, tenantID),
	)
	orch.AddAfterTableHook("PERSONAL_TARJETA", func(ctx context.Context, mapper *migration.Mapper) error {
		return mapper.BuildTarjetaIndex(ctx, mysqlDB)
	})
	orch.RegisterMigrators(
		migration.NewTimeClockEventMigrator(mysqlDB, tenantID),
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

	// After users land: rescue the 23 HTXPROFILES → roles, 188 user_roles
	// assignments, and decorate each role with its legacy HTXPROFILE_AUTH menu
	// list as metadata. Keeps migrated users logged in with their original
	// Histrix authority instead of landing as role-less ghosts.
	orch.AddAfterTableHook("HTXUSERS", func(ctx context.Context, mapper *migration.Mapper) error {
		return migration.RescueLegacyAuthAll(ctx, mysqlDB, orch.PGPool(), tenantID, mapper)
	})
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
