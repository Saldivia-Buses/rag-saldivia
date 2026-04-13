//go:build integration

// Integration tests for the ERP service layer.
// Requires Docker — testcontainers-go spins up PostgreSQL automatically.
// Run: go test -tags=integration -v ./internal/service/ -count=1 -timeout=120s

package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// ============================================================
// Test DB setup
// ============================================================

func setupERPTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("erp_test"),
		postgres.WithUsername("erp"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	if err := applyERPSchema(t, pool); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}
	return pool, cleanup
}

// applyERPSchema creates all tables required by the ERP service tests.
// Derived from the SQL queries in services/erp/db/queries/.
func applyERPSchema(t *testing.T, pool *pgxpool.Pool) error {
	t.Helper()
	ctx := context.Background()

	schema := `
-- Audit log (required by audit.Writer)
CREATE TABLE IF NOT EXISTS audit_log (
	id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	tenant_id   TEXT,
	user_id     TEXT,
	action      TEXT NOT NULL,
	resource    TEXT,
	details     JSONB DEFAULT '{}',
	ip_address  TEXT,
	user_agent  TEXT,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Catalogs / articles / warehouses
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_catalogs (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	type        TEXT NOT NULL,
	code        TEXT NOT NULL,
	name        TEXT NOT NULL,
	parent_id   UUID,
	sort_order  INT NOT NULL DEFAULT 0,
	active      BOOLEAN NOT NULL DEFAULT true,
	metadata    JSONB DEFAULT '{}',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_articles (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	code            TEXT NOT NULL,
	name            TEXT NOT NULL,
	family_id       UUID,
	category_id     UUID,
	unit_id         UUID,
	article_type    TEXT NOT NULL DEFAULT 'product',
	min_stock       NUMERIC(16,4) NOT NULL DEFAULT 0,
	max_stock       NUMERIC(16,4) NOT NULL DEFAULT 0,
	reorder_point   NUMERIC(16,4) NOT NULL DEFAULT 0,
	last_cost       NUMERIC(16,4) NOT NULL DEFAULT 0,
	avg_cost        NUMERIC(16,4) NOT NULL DEFAULT 0,
	metadata        JSONB DEFAULT '{}',
	active          BOOLEAN NOT NULL DEFAULT true,
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_warehouses (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	code        TEXT NOT NULL,
	name        TEXT NOT NULL,
	location    TEXT NOT NULL DEFAULT '',
	active      BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS erp_stock_levels (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	article_id      UUID NOT NULL REFERENCES erp_articles(id),
	warehouse_id    UUID NOT NULL REFERENCES erp_warehouses(id),
	quantity        NUMERIC(16,4) NOT NULL DEFAULT 0,
	reserved        NUMERIC(16,4) NOT NULL DEFAULT 0,
	updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (tenant_id, article_id, warehouse_id)
);

CREATE TABLE IF NOT EXISTS erp_stock_movements (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	article_id      UUID NOT NULL REFERENCES erp_articles(id),
	warehouse_id    UUID NOT NULL REFERENCES erp_warehouses(id),
	movement_type   TEXT NOT NULL,
	quantity        NUMERIC(16,4) NOT NULL,
	unit_cost       NUMERIC(16,4) NOT NULL DEFAULT 0,
	reference_type  TEXT,
	reference_id    UUID,
	concept_id      UUID,
	user_id         TEXT NOT NULL DEFAULT '',
	notes           TEXT NOT NULL DEFAULT '',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Entities (customers, suppliers, etc.)
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_entities (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	type        TEXT NOT NULL,
	code        TEXT NOT NULL DEFAULT '',
	name        TEXT NOT NULL,
	tax_id_hash TEXT,
	email       TEXT,
	phone       TEXT,
	address     TEXT,
	metadata    JSONB DEFAULT '{}',
	active      BOOLEAN NOT NULL DEFAULT true,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS erp_entity_contacts (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	entity_id   UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
	type        TEXT NOT NULL,
	label       TEXT NOT NULL DEFAULT '',
	value       TEXT NOT NULL,
	metadata    JSONB DEFAULT '{}',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_entity_documents (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	entity_id   UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
	name        TEXT NOT NULL,
	doc_type    TEXT NOT NULL DEFAULT '',
	file_key    TEXT NOT NULL,
	uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_entity_notes (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	entity_id   UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
	user_id     TEXT NOT NULL DEFAULT '',
	type        TEXT NOT NULL DEFAULT 'note',
	body        TEXT NOT NULL,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_entity_relations (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	from_id     UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
	to_id       UUID NOT NULL REFERENCES erp_entities(id) ON DELETE CASCADE,
	type        TEXT NOT NULL,
	metadata    JSONB DEFAULT '{}',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Accounting
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_cost_centers (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	code        TEXT NOT NULL,
	name        TEXT NOT NULL,
	parent_id   UUID,
	active      BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS erp_accounts (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	code            TEXT NOT NULL,
	name            TEXT NOT NULL,
	parent_id       UUID,
	account_type    TEXT NOT NULL,
	is_detail       BOOLEAN NOT NULL DEFAULT true,
	cost_center_id  UUID,
	active          BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS erp_fiscal_years (
	id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id           TEXT NOT NULL,
	year                INT NOT NULL,
	start_date          DATE NOT NULL,
	end_date            DATE NOT NULL,
	status              TEXT NOT NULL DEFAULT 'open',
	result_account_id   UUID REFERENCES erp_accounts(id),
	closed_by           TEXT,
	closed_at           TIMESTAMPTZ,
	closing_entry_id    UUID,
	opening_entry_id    UUID
);

CREATE TABLE IF NOT EXISTS erp_journal_entries (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	number          TEXT NOT NULL,
	date            DATE NOT NULL,
	fiscal_year_id  UUID NOT NULL REFERENCES erp_fiscal_years(id),
	concept         TEXT NOT NULL,
	entry_type      TEXT NOT NULL DEFAULT 'manual',
	reference_type  TEXT,
	reference_id    UUID,
	reversed_by     UUID,
	user_id         TEXT NOT NULL DEFAULT '',
	status          TEXT NOT NULL DEFAULT 'draft',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_journal_lines (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	entry_id        UUID NOT NULL REFERENCES erp_journal_entries(id) ON DELETE CASCADE,
	account_id      UUID NOT NULL REFERENCES erp_accounts(id),
	cost_center_id  UUID,
	entry_date      DATE NOT NULL,
	debit           NUMERIC(16,2) NOT NULL DEFAULT 0,
	credit          NUMERIC(16,2) NOT NULL DEFAULT 0,
	description     TEXT NOT NULL DEFAULT '',
	sort_order      INT NOT NULL DEFAULT 0
);

-- ============================================================
-- Invoicing
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_invoices (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	number          TEXT NOT NULL,
	date            DATE,
	due_date        DATE,
	invoice_type    TEXT NOT NULL,
	direction       TEXT NOT NULL,
	entity_id       UUID NOT NULL REFERENCES erp_entities(id),
	currency_id     UUID,
	subtotal        NUMERIC(16,2) NOT NULL DEFAULT 0,
	tax_amount      NUMERIC(16,2) NOT NULL DEFAULT 0,
	total           NUMERIC(16,2) NOT NULL DEFAULT 0,
	order_id        UUID,
	journal_entry_id UUID,
	afip_cae        TEXT,
	afip_cae_due    DATE,
	voided_by       UUID,
	void_reason     TEXT,
	status          TEXT NOT NULL DEFAULT 'draft',
	user_id         TEXT NOT NULL DEFAULT '',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_invoice_lines (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	invoice_id  UUID NOT NULL REFERENCES erp_invoices(id) ON DELETE CASCADE,
	article_id  UUID,
	description TEXT NOT NULL DEFAULT '',
	quantity    NUMERIC(16,4) NOT NULL DEFAULT 0,
	unit_price  NUMERIC(16,4) NOT NULL DEFAULT 0,
	tax_rate    NUMERIC(6,2) NOT NULL DEFAULT 21.00,
	tax_amount  NUMERIC(16,2) NOT NULL DEFAULT 0,
	line_total  NUMERIC(16,2) NOT NULL DEFAULT 0,
	sort_order  INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS erp_tax_entries (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	invoice_id  UUID NOT NULL REFERENCES erp_invoices(id),
	period      TEXT NOT NULL,
	direction   TEXT NOT NULL,
	net_amount  NUMERIC(16,2) NOT NULL DEFAULT 0,
	tax_rate    NUMERIC(6,2) NOT NULL DEFAULT 0,
	tax_amount  NUMERIC(16,2) NOT NULL DEFAULT 0,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_withholdings (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	invoice_id      UUID,
	movement_id     UUID,
	entity_id       UUID NOT NULL REFERENCES erp_entities(id),
	type            TEXT NOT NULL,
	rate            NUMERIC(6,2) NOT NULL DEFAULT 0,
	base_amount     NUMERIC(16,2) NOT NULL DEFAULT 0,
	amount          NUMERIC(16,2) NOT NULL DEFAULT 0,
	certificate_num TEXT,
	date            DATE NOT NULL DEFAULT CURRENT_DATE,
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Account movements (required by VoidInvoice)
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_account_movements (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	entity_id       UUID NOT NULL REFERENCES erp_entities(id),
	date            DATE NOT NULL,
	movement_type   TEXT NOT NULL,
	direction       TEXT NOT NULL DEFAULT 'debit',
	amount          NUMERIC(16,2) NOT NULL DEFAULT 0,
	balance         NUMERIC(16,2) NOT NULL DEFAULT 0,
	invoice_id      UUID,
	treasury_id     UUID,
	journal_entry_id UUID,
	notes           TEXT NOT NULL DEFAULT '',
	user_id         TEXT NOT NULL DEFAULT '',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Treasury
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_bank_accounts (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	bank_name       TEXT NOT NULL,
	branch          TEXT NOT NULL DEFAULT '',
	account_number  TEXT NOT NULL,
	cbu             TEXT,
	alias           TEXT,
	currency_id     UUID,
	account_id      UUID,
	active          BOOLEAN NOT NULL DEFAULT true,
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_cash_registers (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	name        TEXT NOT NULL,
	account_id  UUID,
	active      BOOLEAN NOT NULL DEFAULT true,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_treasury_movements (
	id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id           TEXT NOT NULL,
	date                DATE NOT NULL,
	number              TEXT NOT NULL DEFAULT '',
	movement_type       TEXT NOT NULL,
	amount              NUMERIC(16,2) NOT NULL,
	currency_id         UUID,
	bank_account_id     UUID REFERENCES erp_bank_accounts(id),
	cash_register_id    UUID REFERENCES erp_cash_registers(id),
	entity_id           UUID REFERENCES erp_entities(id),
	concept_id          UUID,
	payment_method      TEXT,
	reference_type      TEXT,
	reference_id        UUID,
	journal_entry_id    UUID,
	user_id             TEXT NOT NULL DEFAULT '',
	notes               TEXT NOT NULL DEFAULT '',
	status              TEXT NOT NULL DEFAULT 'confirmed',
	reconciled          BOOLEAN NOT NULL DEFAULT false,
	reconciliation_id   UUID,
	created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_checks (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	direction   TEXT NOT NULL,
	number      TEXT NOT NULL,
	bank_name   TEXT NOT NULL DEFAULT '',
	amount      NUMERIC(16,2) NOT NULL,
	issue_date  DATE,
	due_date    DATE,
	entity_id   UUID REFERENCES erp_entities(id),
	status      TEXT NOT NULL DEFAULT 'in_portfolio',
	movement_id UUID REFERENCES erp_treasury_movements(id),
	notes       TEXT NOT NULL DEFAULT '',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_cash_counts (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	cash_register_id UUID REFERENCES erp_cash_registers(id),
	date            DATE NOT NULL,
	expected        NUMERIC(16,2) NOT NULL DEFAULT 0,
	counted         NUMERIC(16,2) NOT NULL DEFAULT 0,
	difference      NUMERIC(16,2) NOT NULL DEFAULT 0,
	user_id         TEXT NOT NULL DEFAULT '',
	notes           TEXT NOT NULL DEFAULT '',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_bank_reconciliations (
	id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id           TEXT NOT NULL,
	bank_account_id     UUID NOT NULL REFERENCES erp_bank_accounts(id),
	period              TEXT NOT NULL,
	statement_balance   NUMERIC(16,2) NOT NULL DEFAULT 0,
	book_balance        NUMERIC(16,2) NOT NULL DEFAULT 0,
	status              TEXT NOT NULL DEFAULT 'draft',
	user_id             TEXT NOT NULL DEFAULT '',
	confirmed_at        TIMESTAMPTZ,
	created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_bank_statement_lines (
	id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id           TEXT NOT NULL,
	reconciliation_id   UUID NOT NULL REFERENCES erp_bank_reconciliations(id) ON DELETE CASCADE,
	date                DATE NOT NULL,
	description         TEXT NOT NULL DEFAULT '',
	amount              NUMERIC(16,2) NOT NULL,
	reference           TEXT NOT NULL DEFAULT '',
	matched             BOOLEAN NOT NULL DEFAULT false,
	movement_id         UUID REFERENCES erp_treasury_movements(id),
	created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Receipts
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_receipts (
	id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id           TEXT NOT NULL,
	number              TEXT NOT NULL,
	date                DATE,
	receipt_type        TEXT NOT NULL,
	entity_id           UUID NOT NULL REFERENCES erp_entities(id),
	total               NUMERIC(16,2) NOT NULL DEFAULT 0,
	journal_entry_id    UUID,
	user_id             TEXT NOT NULL DEFAULT '',
	notes               TEXT NOT NULL DEFAULT '',
	status              TEXT NOT NULL DEFAULT 'confirmed',
	created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_receipt_payments (
	id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id               TEXT NOT NULL,
	receipt_id              UUID NOT NULL REFERENCES erp_receipts(id) ON DELETE CASCADE,
	payment_method          TEXT NOT NULL,
	amount                  NUMERIC(16,2) NOT NULL,
	treasury_movement_id    UUID,
	check_id                UUID,
	bank_account_id         UUID,
	notes                   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS erp_receipt_allocations (
	id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id           TEXT NOT NULL,
	receipt_id          UUID NOT NULL REFERENCES erp_receipts(id) ON DELETE CASCADE,
	invoice_id          UUID NOT NULL REFERENCES erp_invoices(id),
	amount              NUMERIC(16,2) NOT NULL,
	account_movement_id UUID
);

CREATE TABLE IF NOT EXISTS erp_receipt_withholdings (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	receipt_id      UUID NOT NULL REFERENCES erp_receipts(id) ON DELETE CASCADE,
	withholding_id  UUID NOT NULL REFERENCES erp_withholdings(id)
);

-- ============================================================
-- Purchasing
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_purchase_orders (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	number      TEXT NOT NULL,
	date        DATE NOT NULL,
	supplier_id UUID NOT NULL REFERENCES erp_entities(id),
	status      TEXT NOT NULL DEFAULT 'draft',
	currency_id UUID,
	total       NUMERIC(16,2) NOT NULL DEFAULT 0,
	notes       TEXT NOT NULL DEFAULT '',
	user_id     TEXT NOT NULL DEFAULT '',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_purchase_order_lines (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	order_id    UUID NOT NULL REFERENCES erp_purchase_orders(id) ON DELETE CASCADE,
	article_id  UUID NOT NULL REFERENCES erp_articles(id),
	quantity    NUMERIC(16,4) NOT NULL,
	unit_price  NUMERIC(16,4) NOT NULL DEFAULT 0,
	received_qty NUMERIC(16,4) NOT NULL DEFAULT 0,
	sort_order  INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS erp_purchase_receipts (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	order_id    UUID NOT NULL REFERENCES erp_purchase_orders(id),
	date        DATE NOT NULL,
	number      TEXT NOT NULL,
	user_id     TEXT NOT NULL DEFAULT '',
	notes       TEXT NOT NULL DEFAULT '',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_purchase_receipt_lines (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	receipt_id      UUID NOT NULL REFERENCES erp_purchase_receipts(id) ON DELETE CASCADE,
	order_line_id   UUID NOT NULL REFERENCES erp_purchase_order_lines(id),
	article_id      UUID NOT NULL REFERENCES erp_articles(id),
	quantity        NUMERIC(16,4) NOT NULL
);

CREATE TABLE IF NOT EXISTS erp_qc_inspections (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	receipt_id      UUID NOT NULL REFERENCES erp_purchase_receipts(id),
	receipt_line_id UUID NOT NULL REFERENCES erp_purchase_receipt_lines(id),
	article_id      UUID NOT NULL REFERENCES erp_articles(id),
	quantity        NUMERIC(16,4) NOT NULL,
	accepted_qty    NUMERIC(16,4) NOT NULL DEFAULT 0,
	rejected_qty    NUMERIC(16,4) NOT NULL DEFAULT 0,
	status          TEXT NOT NULL DEFAULT 'pending',
	inspector_id    TEXT NOT NULL DEFAULT '',
	notes           TEXT NOT NULL DEFAULT '',
	completed_at    TIMESTAMPTZ,
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS erp_supplier_demerits (
	id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id       TEXT NOT NULL,
	supplier_id     UUID NOT NULL REFERENCES erp_entities(id),
	inspection_id   UUID NOT NULL REFERENCES erp_qc_inspections(id),
	points          INT NOT NULL DEFAULT 1,
	reason          TEXT NOT NULL DEFAULT '',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- BOM (bill of materials) — referenced by stock queries
-- ============================================================

CREATE TABLE IF NOT EXISTS erp_bom (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   TEXT NOT NULL,
	parent_id   UUID NOT NULL REFERENCES erp_articles(id),
	child_id    UUID NOT NULL REFERENCES erp_articles(id),
	quantity    NUMERIC(16,4) NOT NULL,
	unit_id     UUID,
	sort_order  INT NOT NULL DEFAULT 0,
	notes       TEXT NOT NULL DEFAULT ''
);
`

	// Execute in a single batch for speed
	statements := strings.Split(schema, ";\n\n")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("exec schema statement: %w\nSQL: %s", err, stmt[:min(100, len(stmt))])
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================
// Helpers
// ============================================================

const testTenantID = "t-erp-test"
const testUserID = "u-erp-test"
const testIP = "127.0.0.1"

func pgDate(s string) pgtype.Date {
	var d pgtype.Date
	_ = d.Scan(s)
	return d
}

func mustParseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		panic(fmt.Sprintf("mustParseUUID(%q): %v", s, err))
	}
	return u
}

// makeServices creates all ERP services wired to the test pool.
// publisher is nil-safe (traces.NewPublisher(nil) no-ops all broadcasts).
func makeAccounting(pool *pgxpool.Pool) *Accounting {
	repo := repository.New(pool)
	aw := audit.NewWriter(pool)
	pub := traces.NewPublisher(nil)
	return NewAccounting(repo, pool, aw, pub)
}

func makeInvoicing(pool *pgxpool.Pool) *Invoicing {
	repo := repository.New(pool)
	aw := audit.NewWriter(pool)
	pub := traces.NewPublisher(nil)
	return NewInvoicing(repo, pool, aw, pub)
}

func makeTreasury(pool *pgxpool.Pool) *Treasury {
	repo := repository.New(pool)
	aw := audit.NewWriter(pool)
	pub := traces.NewPublisher(nil)
	return NewTreasury(repo, pool, aw, pub)
}

func makePurchasing(pool *pgxpool.Pool) *Purchasing {
	repo := repository.New(pool)
	aw := audit.NewWriter(pool)
	pub := traces.NewPublisher(nil)
	return NewPurchasing(repo, pool, aw, pub)
}

// ============================================================
// Test 1: Accounting — CreateFiscalYear + SetResultAccount + CloseFiscalYear
// ============================================================

func TestAccounting_CloseFiscalYear_Integration(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)
	repo := repository.New(pool)

	// Create a result-type equity account (is_detail = true so it appears in balance queries)
	resultAcct, err := svc.CreateAccount(ctx,
		testTenantID, "3-01-001", "Resultado del ejercicio",
		pgtype.UUID{}, "equity", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create result account: %v", err)
	}

	// Create fiscal year 2024
	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2024, "2024-01-01", "2024-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}
	if !fy.ID.Valid {
		t.Fatal("expected valid fiscal year ID")
	}

	// Set the result account on the fiscal year
	if err := svc.SetFiscalYearResultAccount(ctx, testTenantID, fy.ID, resultAcct.ID, testUserID, testIP); err != nil {
		t.Fatalf("set result account: %v", err)
	}

	// Close the fiscal year (no entries → no balance movements → closing entry has only result line)
	result, err := svc.CloseFiscalYear(ctx, testTenantID, fy.ID, testUserID, testIP)
	if err != nil {
		t.Fatalf("close fiscal year: %v", err)
	}
	if !result.ClosingEntryID.Valid {
		t.Error("expected non-null closing_entry_id")
	}
	if !result.NewYearID.Valid {
		t.Error("expected non-null new_year_id")
	}

	// Verify status = 'closed' in DB
	closed, err := repo.GetFiscalYear(ctx, repository.GetFiscalYearParams{
		ID: fy.ID, TenantID: testTenantID,
	})
	if err != nil {
		t.Fatalf("get closed fiscal year: %v", err)
	}
	if closed.Status != "closed" {
		t.Errorf("expected status 'closed', got %q", closed.Status)
	}
	if !closed.ClosingEntryID.Valid {
		t.Error("expected closing_entry_id to be set after close")
	}

	// Verify new year was created
	newFY, err := repo.GetFiscalYear(ctx, repository.GetFiscalYearParams{
		ID: result.NewYearID, TenantID: testTenantID,
	})
	if err != nil {
		t.Fatalf("get new fiscal year: %v", err)
	}
	if newFY.Year != 2025 {
		t.Errorf("expected new year to be 2025, got %d", newFY.Year)
	}
}

// ============================================================
// Test 2: Accounting — CloseFiscalYear blocks on draft entries
// ============================================================

func TestAccounting_CloseFiscalYear_BlocksOnDraft_Integration(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)

	// Create equity account
	resultAcct, err := svc.CreateAccount(ctx,
		testTenantID, "3-01-001", "Resultado",
		pgtype.UUID{}, "equity", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create result account: %v", err)
	}

	// Create expense account (needed to add a journal line)
	expenseAcct, err := svc.CreateAccount(ctx,
		testTenantID, "4-01-001", "Gastos generales",
		pgtype.UUID{}, "expense", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create expense account: %v", err)
	}

	// Create fiscal year
	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2024, "2024-01-01", "2024-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}

	if err := svc.SetFiscalYearResultAccount(ctx, testTenantID, fy.ID, resultAcct.ID, testUserID, testIP); err != nil {
		t.Fatalf("set result account: %v", err)
	}

	// Add a DRAFT journal entry in the fiscal year
	_, err = svc.CreateEntry(ctx, CreateEntryRequest{
		TenantID:     testTenantID,
		Number:       "J-001",
		Date:         pgDate("2024-06-15"),
		FiscalYearID: fy.ID,
		Concept:      "Gasto de prueba",
		EntryType:    "manual",
		UserID:       testUserID,
		IP:           testIP,
		Lines: []CreateLineRequest{
			{AccountID: expenseAcct.ID, Debit: "1000.00", Credit: "0", Description: "Gasto"},
			{AccountID: resultAcct.ID, Debit: "0", Credit: "1000.00", Description: "Resultado"},
		},
	})
	if err != nil {
		t.Fatalf("create draft entry: %v", err)
	}

	// Attempt to close — should fail with "draft" in the error message
	_, err = svc.CloseFiscalYear(ctx, testTenantID, fy.ID, testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when closing with draft entries, got nil")
	}
	if !strings.Contains(err.Error(), "draft") {
		t.Errorf("expected error to mention 'draft', got: %v", err)
	}
}

// ============================================================
// Test 3: Invoicing — CreateInvoice + PostInvoice + VoidPreview
// ============================================================

func TestInvoicing_CreatePostVoidPreview_Integration(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeInvoicing(pool)
	repo := repository.New(pool)

	// Create a customer entity
	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-001",
		Name:     "Test Customer SA",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create invoice
	inv, err := svc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-0001",
		Date:        pgDate("2024-06-15"),
		DueDate:     pgDate("2024-07-15"),
		InvoiceType: "invoice_a",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Producto A", Quantity: "10", UnitPrice: "100.00", TaxRate: "21.00"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if inv.Invoice.Status != "draft" {
		t.Errorf("expected status 'draft' after create, got %q", inv.Invoice.Status)
	}

	// Post the invoice
	if err := svc.PostInvoice(ctx, inv.Invoice.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post invoice: %v", err)
	}

	// Verify status after post
	posted, err := repo.GetInvoice(ctx, repository.GetInvoiceParams{
		ID: inv.Invoice.ID, TenantID: testTenantID,
	})
	if err != nil {
		t.Fatalf("get posted invoice: %v", err)
	}
	if posted.Status != "posted" {
		t.Errorf("expected status 'posted', got %q", posted.Status)
	}

	// VoidPreview
	preview, err := svc.VoidPreview(ctx, inv.Invoice.ID, testTenantID)
	if err != nil {
		t.Fatalf("void preview: %v", err)
	}
	if preview == nil {
		t.Fatal("expected non-nil void preview result")
	}
	// Should have tax entries from PostInvoice
	if preview.TaxEntryCount == 0 {
		t.Error("expected at least one tax entry in void preview")
	}
}

// ============================================================
// Test 4: Invoicing — VoidInvoice (without CAE)
// ============================================================

func TestInvoicing_VoidInvoice_Integration(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeInvoicing(pool)
	repo := repository.New(pool)

	// Create entity
	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-002",
		Name:     "Customer To Void",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create and post invoice
	inv, err := svc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-0002",
		Date:        pgDate("2024-07-01"),
		DueDate:     pgDate("2024-08-01"),
		InvoiceType: "invoice_a",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Servicio B", Quantity: "5", UnitPrice: "200.00", TaxRate: "21.00"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	if err := svc.PostInvoice(ctx, inv.Invoice.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post invoice: %v", err)
	}

	// Void the invoice
	voidResult, err := svc.VoidInvoice(ctx, inv.Invoice.ID, testTenantID, "Error en facturación", testUserID, testIP)
	if err != nil {
		t.Fatalf("void invoice: %v", err)
	}
	if voidResult == nil {
		t.Fatal("expected non-nil void result")
	}

	// Verify invoice status = 'cancelled'
	voided, err := repo.GetInvoice(ctx, repository.GetInvoiceParams{
		ID: inv.Invoice.ID, TenantID: testTenantID,
	})
	if err != nil {
		t.Fatalf("get voided invoice: %v", err)
	}
	if voided.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got %q", voided.Status)
	}
	// VoidInvoice sets voided_by (UUID from userID — it's actually text in this service, passed as VoidedBy pgtype.UUID)
	// The void_reason should be set (checked via direct SQL since GetInvoice doesn't return void_reason)
	var voidReason string
	if err := pool.QueryRow(ctx,
		`SELECT COALESCE(void_reason, '') FROM erp_invoices WHERE id = $1 AND tenant_id = $2`,
		inv.Invoice.ID, testTenantID,
	).Scan(&voidReason); err != nil {
		t.Fatalf("query void_reason: %v", err)
	}
	if voidReason != "Error en facturación" {
		t.Errorf("expected void_reason 'Error en facturación', got %q", voidReason)
	}

	// Tax entries should have been reversed (doubled in count: original + reversal)
	var taxCount int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM erp_tax_entries WHERE tenant_id = $1 AND invoice_id = $2`,
		testTenantID, inv.Invoice.ID,
	).Scan(&taxCount); err != nil {
		t.Fatalf("count tax entries: %v", err)
	}
	if taxCount < 2 {
		t.Errorf("expected at least 2 tax entries (original + reversal), got %d", taxCount)
	}
}

// ============================================================
// Test 5: Purchasing — InspectReceipt (accept all)
// ============================================================

func TestPurchasing_InspectReceipt_Integration(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svcPurch := makePurchasing(pool)
	repo := repository.New(pool)

	// Create supplier entity
	supplier, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "supplier",
		Code:     "PROV-001",
		Name:     "Proveedor Test SRL",
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}

	// Create article
	article, err := repo.CreateArticle(ctx, repository.CreateArticleParams{
		TenantID:    testTenantID,
		Code:        "ART-001",
		Name:        "Artículo Test",
		ArticleType: "product",
		Metadata:    []byte("{}"),
	})
	if err != nil {
		t.Fatalf("create article: %v", err)
	}

	// Create warehouse
	warehouse, err := repo.CreateWarehouse(ctx, repository.CreateWarehouseParams{
		TenantID: testTenantID,
		Code:     "DEP-01",
		Name:     "Depósito Central",
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	// Create purchase order
	order, err := svcPurch.CreateOrder(ctx, CreateOrderRequest{
		TenantID:   testTenantID,
		Number:     "OC-0001",
		Date:       pgDate("2024-05-01"),
		SupplierID: supplier.ID,
		UserID:     testUserID,
		IP:         testIP,
		Lines: []CreateOrderLineRequest{
			{ArticleID: article.ID, Quantity: "100", UnitPrice: "50.00"},
		},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	// Approve the order
	if err := svcPurch.ApproveOrder(ctx, order.Order.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("approve order: %v", err)
	}

	// Receive the order
	orderLines, err := repo.ListPurchaseOrderLines(ctx, repository.ListPurchaseOrderLinesParams{
		OrderID: order.Order.ID, TenantID: testTenantID,
	})
	if err != nil || len(orderLines) == 0 {
		t.Fatalf("list order lines: %v (count=%d)", err, len(orderLines))
	}
	orderLineID := orderLines[0].ID

	receiveErr := svcPurch.Receive(ctx, ReceiveRequest{
		TenantID: testTenantID,
		OrderID:  order.Order.ID,
		Date:     pgDate("2024-05-10"),
		Number:   "REC-0001",
		UserID:   testUserID,
		IP:       testIP,
		Lines: []ReceiveLineRequest{
			{OrderLineID: orderLineID, ArticleID: article.ID, Quantity: "100"},
		},
	})
	if receiveErr != nil {
		t.Fatalf("receive order: %v", receiveErr)
	}

	// Fetch receipt
	receipts, err := repo.ListPurchaseReceipts(ctx, repository.ListPurchaseReceiptsParams{
		TenantID: testTenantID, Limit: 10, Offset: 0,
	})
	if err != nil || len(receipts) == 0 {
		t.Fatalf("list receipts: %v (count=%d)", err, len(receipts))
	}
	receiptID := receipts[0].ID

	// Fetch receipt lines
	receiptLines, err := repo.ListPurchaseReceiptLines(ctx, repository.ListPurchaseReceiptLinesParams{
		TenantID: testTenantID, ReceiptID: receiptID,
	})
	if err != nil || len(receiptLines) == 0 {
		t.Fatalf("list receipt lines: %v (count=%d)", err, len(receiptLines))
	}

	// Inspect receipt — accept all
	inspections, err := svcPurch.InspectReceipt(ctx, testTenantID, receiptID,
		[]InspectionInput{
			{
				ReceiptLineID: uuidStr(receiptLines[0].ID),
				ArticleID:     uuidStr(article.ID),
				WarehouseID:   uuidStr(warehouse.ID),
				Quantity:      "100",
				AcceptedQty:   "100",
				RejectedQty:   "0",
				Notes:         "All accepted",
			},
		},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("inspect receipt: %v", err)
	}
	if len(inspections) != 1 {
		t.Fatalf("expected 1 inspection, got %d", len(inspections))
	}
	if inspections[0].Status != "completed" {
		t.Errorf("expected inspection status 'completed', got %q", inspections[0].Status)
	}
	if !isPositiveNumeric(inspections[0].AcceptedQty) {
		t.Error("expected accepted_qty > 0")
	}

	// Verify stock movement was created
	var movCount int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM erp_stock_movements WHERE tenant_id = $1 AND article_id = $2 AND movement_type = 'in'`,
		testTenantID, article.ID,
	).Scan(&movCount); err != nil {
		t.Fatalf("count stock movements: %v", err)
	}
	if movCount == 0 {
		t.Error("expected stock movement 'in' for accepted qty")
	}
}

// isPositiveNumeric returns true if a pgtype.Numeric represents a value > 0.
func isPositiveNumeric(n pgtype.Numeric) bool {
	if !n.Valid || n.Int == nil {
		return false
	}
	return n.Int.Sign() > 0
}

// ============================================================
// Test 6: Treasury — CreateReconciliation + ImportStatementLines + AutoMatch
// ============================================================

func TestTreasury_ReconciliationAutoMatch_Integration(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeTreasury(pool)

	// Create a bank account
	ba, err := svc.CreateBankAccount(ctx, repository.CreateBankAccountParams{
		TenantID:      testTenantID,
		BankName:      "Banco Test",
		AccountNumber: "0000-1234-5678",
	}, testUserID, testIP)
	if err != nil {
		t.Fatalf("create bank account: %v", err)
	}

	// Create a confirmed treasury movement for 2024-03 that the statement line will match
	period := "2024-03"
	mov, err := svc.CreateMovement(ctx, CreateTreasuryMovementRequest{
		CreateTreasuryMovementParams: repository.CreateTreasuryMovementParams{
			TenantID:      testTenantID,
			Date:          pgDate("2024-03-15"),
			Number:        "MOV-0001",
			MovementType:  "bank_deposit",
			Amount:        pgNumeric("1000.00"),
			BankAccountID: ba.ID,
			UserID:        testUserID,
		},
		UserIDVal: testUserID,
		IP:        testIP,
	})
	if err != nil {
		t.Fatalf("create treasury movement: %v", err)
	}
	_ = mov

	// Create reconciliation
	recon, err := svc.CreateReconciliation(ctx,
		testTenantID, ba.ID, period,
		"1000.00", "1000.00",
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create reconciliation: %v", err)
	}
	if !recon.ID.Valid {
		t.Fatal("expected valid reconciliation ID")
	}

	// Import a statement line that matches the movement (same amount, within 2 days)
	count, err := svc.ImportStatementLines(ctx, testTenantID, recon.ID, []StatementLineInput{
		{Date: "2024-03-15", Description: "Deposito banco", Amount: "1000.00", Reference: "REF-001"},
		{Date: "2024-03-20", Description: "Movimiento sin match", Amount: "999.00", Reference: "REF-002"},
	})
	if err != nil {
		t.Fatalf("import statement lines: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 lines imported, got %d", count)
	}

	// AutoMatch
	result, err := svc.AutoMatch(ctx, testTenantID, recon.ID)
	if err != nil {
		t.Fatalf("auto match: %v", err)
	}
	if result.Matched == 0 {
		t.Error("expected at least one line matched by AutoMatch")
	}
	if result.Matched+result.Unmatched != 2 {
		t.Errorf("matched(%d) + unmatched(%d) should equal 2 total lines", result.Matched, result.Unmatched)
	}

	// Verify reconciliation still exists
	detail, err := svc.GetReconciliation(ctx, testTenantID, recon.ID)
	if err != nil {
		t.Fatalf("get reconciliation: %v", err)
	}
	if len(detail.Lines) != 2 {
		t.Errorf("expected 2 statement lines, got %d", len(detail.Lines))
	}
}

// ============================================================
// Test 7: Accounting — CloseFiscalYear fails without result account
// ============================================================

func TestCloseFiscalYear_WithoutResultAccount_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)

	// Create fiscal year without setting result account
	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2025, "2025-01-01", "2025-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}

	// Attempt close — must fail because result_account_id is not set
	_, err = svc.CloseFiscalYear(ctx, testTenantID, fy.ID, testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when closing without result_account_id, got nil")
	}
	if !strings.Contains(err.Error(), "result_account_id") {
		t.Errorf("expected error to mention 'result_account_id', got: %v", err)
	}
}

// ============================================================
// Test 8: Accounting — CloseFiscalYear fails on already-closed year
// ============================================================

func TestCloseFiscalYear_AlreadyClosed_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)

	// Full close setup
	resultAcct, err := svc.CreateAccount(ctx,
		testTenantID, "3-01-002", "Resultado cierre doble",
		pgtype.UUID{}, "equity", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create result account: %v", err)
	}

	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2020, "2020-01-01", "2020-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}

	if err := svc.SetFiscalYearResultAccount(ctx, testTenantID, fy.ID, resultAcct.ID, testUserID, testIP); err != nil {
		t.Fatalf("set result account: %v", err)
	}

	// First close — must succeed
	if _, err := svc.CloseFiscalYear(ctx, testTenantID, fy.ID, testUserID, testIP); err != nil {
		t.Fatalf("first close: %v", err)
	}

	// Second close on the same (now closed) year — must fail
	_, err = svc.CloseFiscalYear(ctx, testTenantID, fy.ID, testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when closing an already-closed fiscal year, got nil")
	}
	// Service returns "fiscal year is not open"
	if !strings.Contains(err.Error(), "not open") && !strings.Contains(err.Error(), "closed") {
		t.Errorf("expected error to mention 'not open' or 'closed', got: %v", err)
	}
}

// ============================================================
// Test 9: Accounting — PreviewClose on empty fiscal year
// ============================================================

func TestPreviewClose_EmptyFiscalYear(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)

	// Create result account and fiscal year with no entries
	resultAcct, err := svc.CreateAccount(ctx,
		testTenantID, "3-01-003", "Resultado preview",
		pgtype.UUID{}, "equity", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create result account: %v", err)
	}

	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2021, "2021-01-01", "2021-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}

	if err := svc.SetFiscalYearResultAccount(ctx, testTenantID, fy.ID, resultAcct.ID, testUserID, testIP); err != nil {
		t.Fatalf("set result account: %v", err)
	}

	preview, err := svc.PreviewClose(ctx, testTenantID, fy.ID)
	if err != nil {
		t.Fatalf("preview close: %v", err)
	}
	if preview == nil {
		t.Fatal("expected non-nil preview result")
	}
	if !preview.CanClose {
		t.Errorf("expected CanClose=true for empty fiscal year, blocked_reason=%q", preview.BlockedReason)
	}
	if len(preview.DraftEntries) != 0 {
		t.Errorf("expected 0 draft entries, got %d", len(preview.DraftEntries))
	}
	if len(preview.Balances) != 0 {
		t.Errorf("expected 0 account balances for empty fiscal year, got %d", len(preview.Balances))
	}
}

// ============================================================
// Test 10: Accounting — PostEntry happy path (draft → posted)
// ============================================================

func TestPostEntry_HappyPath(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)
	repo := repository.New(pool)

	// Create accounts
	debitAcct, err := svc.CreateAccount(ctx,
		testTenantID, "4-02-001", "Gasto post test",
		pgtype.UUID{}, "expense", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create debit account: %v", err)
	}
	creditAcct, err := svc.CreateAccount(ctx,
		testTenantID, "2-02-001", "Pasivo post test",
		pgtype.UUID{}, "liability", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create credit account: %v", err)
	}

	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2022, "2022-01-01", "2022-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}

	// Create a draft entry
	detail, err := svc.CreateEntry(ctx, CreateEntryRequest{
		TenantID:     testTenantID,
		Number:       "J-POST-001",
		Date:         pgDate("2022-03-15"),
		FiscalYearID: fy.ID,
		Concept:      "Asiento de prueba",
		EntryType:    "manual",
		UserID:       testUserID,
		IP:           testIP,
		Lines: []CreateLineRequest{
			{AccountID: debitAcct.ID, Debit: "500.00", Credit: "0", Description: "Debe"},
			{AccountID: creditAcct.ID, Debit: "0", Credit: "500.00", Description: "Haber"},
		},
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if detail.Entry.Status != "draft" {
		t.Errorf("expected status 'draft' after create, got %q", detail.Entry.Status)
	}

	// Post it
	if err := svc.PostEntry(ctx, detail.Entry.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post entry: %v", err)
	}

	// Verify status in DB
	posted, err := repo.GetJournalEntry(ctx, repository.GetJournalEntryParams{
		ID: detail.Entry.ID, TenantID: testTenantID,
	})
	if err != nil {
		t.Fatalf("get posted entry: %v", err)
	}
	if posted.Status != "posted" {
		t.Errorf("expected status 'posted', got %q", posted.Status)
	}
}

// ============================================================
// Test 11: Accounting — PostEntry on already-posted entry fails
// ============================================================

func TestPostEntry_AlreadyPosted_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeAccounting(pool)

	debitAcct, err := svc.CreateAccount(ctx,
		testTenantID, "4-03-001", "Gasto double post",
		pgtype.UUID{}, "expense", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create debit account: %v", err)
	}
	creditAcct, err := svc.CreateAccount(ctx,
		testTenantID, "2-03-001", "Pasivo double post",
		pgtype.UUID{}, "liability", true, pgtype.UUID{},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("create credit account: %v", err)
	}

	fy, err := svc.CreateFiscalYear(ctx, testTenantID, 2023, "2023-01-01", "2023-12-31", testUserID, testIP)
	if err != nil {
		t.Fatalf("create fiscal year: %v", err)
	}

	detail, err := svc.CreateEntry(ctx, CreateEntryRequest{
		TenantID:     testTenantID,
		Number:       "J-DPOST-001",
		Date:         pgDate("2023-01-10"),
		FiscalYearID: fy.ID,
		Concept:      "Asiento double post",
		EntryType:    "manual",
		UserID:       testUserID,
		IP:           testIP,
		Lines: []CreateLineRequest{
			{AccountID: debitAcct.ID, Debit: "200.00", Credit: "0", Description: "D"},
			{AccountID: creditAcct.ID, Debit: "0", Credit: "200.00", Description: "C"},
		},
	})
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}

	// First post — must succeed
	if err := svc.PostEntry(ctx, detail.Entry.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("first post: %v", err)
	}

	// Second post on the same (already posted) entry — must fail
	err = svc.PostEntry(ctx, detail.Entry.ID, testTenantID, testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when posting an already-posted entry, got nil")
	}
	// Service returns "entry not found or already posted"
	if !strings.Contains(err.Error(), "already posted") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to mention 'already posted', got: %v", err)
	}
}

// ============================================================
// Test 12: Invoicing — VoidInvoice on draft invoice fails
// ============================================================

func TestVoidInvoice_OnDraftInvoice_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeInvoicing(pool)
	repo := repository.New(pool)

	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-VOID-DRAFT",
		Name:     "Customer Draft Void",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create invoice but do NOT post it — stays draft
	inv, err := svc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-DRAFT-VOID",
		Date:        pgDate("2024-09-01"),
		InvoiceType: "invoice_a",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Producto draft", Quantity: "1", UnitPrice: "100.00", TaxRate: "21.00"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	// VoidInvoice on a draft must fail
	_, err = svc.VoidInvoice(ctx, inv.Invoice.ID, testTenantID, "intento de anular borrador", testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when voiding a draft invoice, got nil")
	}
	// Service returns "solo se pueden anular facturas posted/paid"
	if !strings.Contains(err.Error(), "posted") && !strings.Contains(err.Error(), "draft") {
		t.Errorf("expected error to mention status restriction, got: %v", err)
	}
}

// ============================================================
// Test 13: Invoicing — VoidInvoice on already-voided invoice fails
// ============================================================

func TestVoidInvoice_AlreadyVoided_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeInvoicing(pool)
	repo := repository.New(pool)

	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-DOUBLE-VOID",
		Name:     "Customer Double Void",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create → post → void
	inv, err := svc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-DOUBLE-VOID",
		Date:        pgDate("2024-10-01"),
		InvoiceType: "invoice_b",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Item doble void", Quantity: "2", UnitPrice: "50.00", TaxRate: "21.00"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	if err := svc.PostInvoice(ctx, inv.Invoice.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post invoice: %v", err)
	}

	// First void — must succeed
	if _, err := svc.VoidInvoice(ctx, inv.Invoice.ID, testTenantID, "primer void", testUserID, testIP); err != nil {
		t.Fatalf("first void: %v", err)
	}

	// Second void on the same (already cancelled) invoice — must fail
	_, err = svc.VoidInvoice(ctx, inv.Invoice.ID, testTenantID, "segundo void", testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when voiding an already-voided invoice, got nil")
	}
	if !strings.Contains(err.Error(), "posted") && !strings.Contains(err.Error(), "cancelled") &&
		!strings.Contains(err.Error(), "anular") {
		t.Errorf("expected error to mention status restriction, got: %v", err)
	}
}

// ============================================================
// Test 14: Treasury — CreateReceipt fails when payments don't balance allocations
// ============================================================

func TestCreateReceipt_UnbalancedPayments_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeTreasury(pool)
	repo := repository.New(pool)

	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-UNBALANCED",
		Name:     "Customer Unbalanced",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create a posted invoice to allocate to
	invSvc := makeInvoicing(pool)
	inv, err := invSvc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-UNBAL-001",
		Date:        pgDate("2024-11-01"),
		InvoiceType: "invoice_a",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Item", Quantity: "1", UnitPrice: "1000.00", TaxRate: "0"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if err := invSvc.PostInvoice(ctx, inv.Invoice.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post invoice: %v", err)
	}

	invIDStr := uuidStr(inv.Invoice.ID)

	// Payment = 500, allocation = 1000 — diff = 500, exceeds the 0.01 tolerance
	_, err = svc.CreateReceipt(ctx, testTenantID, ReceiptInput{
		ReceiptType: "collection",
		EntityID:    uuidStr(entity.ID),
		Date:        "2024-11-05",
		Payments: []ReceiptPaymentInput{
			{PaymentMethod: "cash", Amount: "500.00"},
		},
		Allocations: []ReceiptAllocInput{
			{InvoiceID: invIDStr, Amount: "1000.00"},
		},
	}, testUserID, testIP)
	if err == nil {
		t.Fatal("expected error for unbalanced receipt, got nil")
	}
	if !strings.Contains(err.Error(), "balance") && !strings.Contains(err.Error(), "balanc") {
		t.Errorf("expected ErrReceiptUnbalanced, got: %v", err)
	}
}

// ============================================================
// Test 15: Treasury — CreateReceipt happy path
// ============================================================

func TestCreateReceipt_HappyPath(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeTreasury(pool)
	repo := repository.New(pool)

	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-REC-HP",
		Name:     "Customer Receipt Happy",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create + post an invoice to allocate against
	invSvc := makeInvoicing(pool)
	inv, err := invSvc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-REC-HP-001",
		Date:        pgDate("2024-08-01"),
		InvoiceType: "invoice_a",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Servicio", Quantity: "1", UnitPrice: "2000.00", TaxRate: "0"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if err := invSvc.PostInvoice(ctx, inv.Invoice.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post invoice: %v", err)
	}

	invIDStr := uuidStr(inv.Invoice.ID)

	// Create receipt: payment = allocation = 2000
	detail, err := svc.CreateReceipt(ctx, testTenantID, ReceiptInput{
		ReceiptType: "collection",
		EntityID:    uuidStr(entity.ID),
		Date:        "2024-08-10",
		Notes:       "Cobro FA-REC-HP-001",
		Payments: []ReceiptPaymentInput{
			{PaymentMethod: "cash", Amount: "2000.00"},
		},
		Allocations: []ReceiptAllocInput{
			{InvoiceID: invIDStr, Amount: "2000.00"},
		},
	}, testUserID, testIP)
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if detail == nil {
		t.Fatal("expected non-nil receipt detail")
	}

	// Verify receipt exists in DB with confirmed status
	var status string
	if err := pool.QueryRow(ctx,
		`SELECT status FROM erp_receipts WHERE id = $1 AND tenant_id = $2`,
		detail.Receipt.ID, testTenantID,
	).Scan(&status); err != nil {
		t.Fatalf("query receipt status: %v", err)
	}
	if status != "confirmed" {
		t.Errorf("expected receipt status 'confirmed', got %q", status)
	}

	// Verify one receipt payment was created
	if len(detail.Payments) != 1 {
		t.Errorf("expected 1 payment, got %d", len(detail.Payments))
	}

	// Verify one allocation was created
	if len(detail.Allocations) != 1 {
		t.Errorf("expected 1 allocation, got %d", len(detail.Allocations))
	}
}

// ============================================================
// Test 16: Treasury — VoidReceipt fails when receipt is not confirmed
// ============================================================

func TestVoidReceipt_AlreadyVoided_Fails(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := makeTreasury(pool)
	repo := repository.New(pool)

	entity, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "customer",
		Code:     "CLI-VOID-REC",
		Name:     "Customer Void Receipt",
	})
	if err != nil {
		t.Fatalf("create entity: %v", err)
	}

	// Create + post invoice
	invSvc := makeInvoicing(pool)
	inv, err := invSvc.CreateInvoice(ctx, CreateInvoiceRequest{
		TenantID:    testTenantID,
		Number:      "FA-VREC-001",
		Date:        pgDate("2024-09-15"),
		InvoiceType: "invoice_a",
		Direction:   "issued",
		EntityID:    entity.ID,
		UserID:      testUserID,
		IP:          testIP,
		Lines: []CreateInvoiceLineRequest{
			{Description: "Item void rec", Quantity: "1", UnitPrice: "300.00", TaxRate: "0"},
		},
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if err := invSvc.PostInvoice(ctx, inv.Invoice.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("post invoice: %v", err)
	}

	// Create receipt
	detail, err := svc.CreateReceipt(ctx, testTenantID, ReceiptInput{
		ReceiptType: "collection",
		EntityID:    uuidStr(entity.ID),
		Date:        "2024-09-20",
		Payments: []ReceiptPaymentInput{
			{PaymentMethod: "cash", Amount: "300.00"},
		},
		Allocations: []ReceiptAllocInput{
			{InvoiceID: uuidStr(inv.Invoice.ID), Amount: "300.00"},
		},
	}, testUserID, testIP)
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}

	// First void — must succeed
	if err := svc.VoidReceipt(ctx, testTenantID, detail.Receipt.ID, testUserID, testIP); err != nil {
		t.Fatalf("first void: %v", err)
	}

	// Verify receipt status is now cancelled
	var status string
	if err := pool.QueryRow(ctx,
		`SELECT status FROM erp_receipts WHERE id = $1 AND tenant_id = $2`,
		detail.Receipt.ID, testTenantID,
	).Scan(&status); err != nil {
		t.Fatalf("query receipt status: %v", err)
	}
	if status != "cancelled" {
		t.Errorf("expected status 'cancelled' after void, got %q", status)
	}

	// Second void on the already-cancelled receipt — must fail
	err = svc.VoidReceipt(ctx, testTenantID, detail.Receipt.ID, testUserID, testIP)
	if err == nil {
		t.Fatal("expected error when voiding an already-cancelled receipt, got nil")
	}
	// Service returns "receipt is not confirmed (status: cancelled)"
	if !strings.Contains(err.Error(), "confirmed") && !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected error to mention status restriction, got: %v", err)
	}
}

// ============================================================
// Test 17: Purchasing — InspectReceipt with rejection creates demerit record
// ============================================================

func TestInspectReceipt_Reject_CreatesDemeritRecord(t *testing.T) {
	pool, cleanup := setupERPTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svcPurch := makePurchasing(pool)
	repo := repository.New(pool)

	// Create supplier
	supplier, err := repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID: testTenantID,
		Type:     "supplier",
		Code:     "PROV-DEMERIT",
		Name:     "Proveedor Demerit SRL",
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}

	// Create article + warehouse
	article, err := repo.CreateArticle(ctx, repository.CreateArticleParams{
		TenantID:    testTenantID,
		Code:        "ART-DEMERIT",
		Name:        "Artículo Demerit",
		ArticleType: "product",
		Metadata:    []byte("{}"),
	})
	if err != nil {
		t.Fatalf("create article: %v", err)
	}

	warehouse, err := repo.CreateWarehouse(ctx, repository.CreateWarehouseParams{
		TenantID: testTenantID,
		Code:     "DEP-DEMERIT",
		Name:     "Depósito Demerit",
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	// Create + approve order
	order, err := svcPurch.CreateOrder(ctx, CreateOrderRequest{
		TenantID:   testTenantID,
		Number:     "OC-DEMERIT",
		Date:       pgDate("2024-04-01"),
		SupplierID: supplier.ID,
		UserID:     testUserID,
		IP:         testIP,
		Lines: []CreateOrderLineRequest{
			{ArticleID: article.ID, Quantity: "50", UnitPrice: "10.00"},
		},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := svcPurch.ApproveOrder(ctx, order.Order.ID, testTenantID, testUserID, testIP); err != nil {
		t.Fatalf("approve order: %v", err)
	}

	// Receive
	orderLines, err := repo.ListPurchaseOrderLines(ctx, repository.ListPurchaseOrderLinesParams{
		OrderID: order.Order.ID, TenantID: testTenantID,
	})
	if err != nil || len(orderLines) == 0 {
		t.Fatalf("list order lines: %v (count=%d)", err, len(orderLines))
	}

	if err := svcPurch.Receive(ctx, ReceiveRequest{
		TenantID: testTenantID,
		OrderID:  order.Order.ID,
		Date:     pgDate("2024-04-10"),
		Number:   "REC-DEMERIT",
		UserID:   testUserID,
		IP:       testIP,
		Lines: []ReceiveLineRequest{
			{OrderLineID: orderLines[0].ID, ArticleID: article.ID, Quantity: "50"},
		},
	}); err != nil {
		t.Fatalf("receive order: %v", err)
	}

	// Fetch receipt and its lines
	receipts, err := repo.ListPurchaseReceipts(ctx, repository.ListPurchaseReceiptsParams{
		TenantID: testTenantID, Limit: 10, Offset: 0,
	})
	if err != nil || len(receipts) == 0 {
		t.Fatalf("list receipts: %v (count=%d)", err, len(receipts))
	}
	receiptID := receipts[len(receipts)-1].ID

	receiptLines, err := repo.ListPurchaseReceiptLines(ctx, repository.ListPurchaseReceiptLinesParams{
		TenantID: testTenantID, ReceiptID: receiptID,
	})
	if err != nil || len(receiptLines) == 0 {
		t.Fatalf("list receipt lines: %v (count=%d)", err, len(receiptLines))
	}

	// Count demerits BEFORE inspection
	var demeritsBefore int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM erp_supplier_demerits WHERE tenant_id = $1 AND supplier_id = $2`,
		testTenantID, supplier.ID,
	).Scan(&demeritsBefore); err != nil {
		t.Fatalf("count demerits before: %v", err)
	}

	// Inspect: accept 30, reject 20
	_, err = svcPurch.InspectReceipt(ctx, testTenantID, receiptID,
		[]InspectionInput{
			{
				ReceiptLineID: uuidStr(receiptLines[0].ID),
				ArticleID:     uuidStr(article.ID),
				WarehouseID:   uuidStr(warehouse.ID),
				Quantity:      "50",
				AcceptedQty:   "30",
				RejectedQty:   "20",
				Notes:         "20 units defective",
			},
		},
		testUserID, testIP,
	)
	if err != nil {
		t.Fatalf("inspect receipt: %v", err)
	}

	// Verify demerit record was created for the supplier
	var demeritsAfter int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM erp_supplier_demerits WHERE tenant_id = $1 AND supplier_id = $2`,
		testTenantID, supplier.ID,
	).Scan(&demeritsAfter); err != nil {
		t.Fatalf("count demerits after: %v", err)
	}
	if demeritsAfter <= demeritsBefore {
		t.Errorf("expected demerit count to increase after rejection (before=%d, after=%d)",
			demeritsBefore, demeritsAfter)
	}

	// Verify demerit points = rejected quantity (1 point per unit)
	var points int
	if err := pool.QueryRow(ctx,
		`SELECT points FROM erp_supplier_demerits WHERE tenant_id = $1 AND supplier_id = $2 ORDER BY created_at DESC LIMIT 1`,
		testTenantID, supplier.ID,
	).Scan(&points); err != nil {
		t.Fatalf("query demerit points: %v", err)
	}
	if points != 20 {
		t.Errorf("expected 20 demerit points (1 per rejected unit), got %d", points)
	}

	// Verify stock level reflects accepted qty only (30, not 50)
	var stockQty string
	if err := pool.QueryRow(ctx,
		`SELECT quantity::text FROM erp_stock_levels WHERE tenant_id = $1 AND article_id = $2 AND warehouse_id = $3`,
		testTenantID, article.ID, warehouse.ID,
	).Scan(&stockQty); err != nil {
		t.Fatalf("query stock level: %v", err)
	}
	if stockQty != "30" {
		t.Errorf("expected stock quantity 30 (accepted only), got %q", stockQty)
	}
}
