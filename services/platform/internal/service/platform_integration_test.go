//go:build integration

// Integration tests for the platform service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/

package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Camionerou/rag-saldivia/services/platform/db"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("sda_platform_test"),
		postgres.WithUsername("sda"),
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

	// Apply platform migration (plans, tenants, modules, etc.)
	migration := `
		CREATE TABLE plans (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			max_users INTEGER NOT NULL DEFAULT 10,
			max_storage_mb INTEGER NOT NULL DEFAULT 5120,
			ai_credits_monthly INTEGER NOT NULL DEFAULT 1000,
			price_usd NUMERIC(10,2) NOT NULL DEFAULT 0,
			features JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE tenants (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			slug TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			plan_id TEXT NOT NULL REFERENCES plans(id),
			postgres_url TEXT NOT NULL,
			redis_url TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT true,
			logo_url TEXT,
			domain TEXT,
			settings JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE modules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			category TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			version TEXT NOT NULL DEFAULT '0.1.0',
			requires TEXT[] DEFAULT '{}',
			tier_min TEXT NOT NULL DEFAULT 'starter',
			enabled BOOLEAN NOT NULL DEFAULT true
		);
		CREATE TABLE tenant_modules (
			tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			module_id TEXT NOT NULL REFERENCES modules(id),
			enabled BOOLEAN NOT NULL DEFAULT true,
			config JSONB NOT NULL DEFAULT '{}',
			enabled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			enabled_by TEXT NOT NULL,
			PRIMARY KEY (tenant_id, module_id)
		);
		CREATE TABLE feature_flags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			tenant_id TEXT REFERENCES tenants(id) ON DELETE CASCADE,
			enabled BOOLEAN NOT NULL DEFAULT false,
			config JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE global_config (
			key TEXT PRIMARY KEY,
			value JSONB NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_by TEXT NOT NULL
		);

		INSERT INTO plans (id, name, max_users, price_usd) VALUES
			('starter', 'Starter', 10, 49),
			('business', 'Business', 50, 299);

		INSERT INTO modules (id, name, category) VALUES
			('chat', 'Chat + RAG', 'core'),
			('docs', 'Gestion Documental', 'platform'),
			('fleet', 'Transporte/Logistica', 'vertical');
	`
	if _, err := pool.Exec(ctx, migration); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
	return pool, cleanup
}

func TestCreateTenant_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug:        "saldivia",
		Name:        "Saldivia Buses",
		PlanID:      "starter",
		PostgresUrl: "postgres://localhost/saldivia",
		RedisUrl:    "redis://localhost/0",
		Settings:    []byte("{}"),
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if tenant.Slug != "saldivia" {
		t.Errorf("expected slug 'saldivia', got %q", tenant.Slug)
	}
	if tenant.ID == "" {
		t.Error("expected non-empty tenant ID")
	}
}

func TestCreateTenant_DuplicateSlug_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "acme", Name: "Acme", PlanID: "starter",
		PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
	})

	_, err := svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "acme", Name: "Acme Dup", PlanID: "starter",
		PostgresUrl: "pg://y", RedisUrl: "redis://y", Settings: []byte("{}"),
	})
	if err != ErrSlugTaken {
		t.Fatalf("expected ErrSlugTaken, got: %v", err)
	}
}

func TestCreateTenant_InvalidSlug_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	slugs := []string{"BAD", "has spaces", "-starts-dash", "a", "../etc"}
	for _, slug := range slugs {
		t.Run(slug, func(t *testing.T) {
			_, err := svc.CreateTenant(ctx, db.CreateTenantParams{
				Slug: slug, Name: "Test", PlanID: "starter",
				PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
			})
			if err != ErrInvalidSlug {
				t.Errorf("expected ErrInvalidSlug for %q, got: %v", slug, err)
			}
		})
	}
}

func TestListTenants_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "tenant-a", Name: "A", PlanID: "starter",
		PostgresUrl: "pg://a", RedisUrl: "redis://a", Settings: []byte("{}"),
	})
	svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "tenant-b", Name: "B", PlanID: "business",
		PostgresUrl: "pg://b", RedisUrl: "redis://b", Settings: []byte("{}"),
	})

	tenants, err := svc.ListTenants(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(tenants))
	}
}

func TestGetTenant_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "findme", Name: "Find Me", PlanID: "starter",
		PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
	})

	tenant, err := svc.GetTenant(ctx, "findme")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if tenant.Name != "Find Me" {
		t.Errorf("expected name 'Find Me', got %q", tenant.Name)
	}
}

func TestGetTenant_NotFound_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	_, err := svc.GetTenant(context.Background(), "nonexistent")
	if err != ErrTenantNotFound {
		t.Fatalf("expected ErrTenantNotFound, got: %v", err)
	}
}

func TestEnableModule_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "modtest", Name: "Mod Test", PlanID: "starter",
		PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
	})

	err := svc.EnableModule(ctx, db.EnableModuleForTenantParams{
		TenantID:  tenant.ID,
		ModuleID:  "fleet",
		Config:    []byte(`{"vehicles":100}`),
		EnabledBy: "admin",
	})
	if err != nil {
		t.Fatalf("enable module: %v", err)
	}

	modules, err := svc.GetTenantModules(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("get tenant modules: %v", err)
	}
	if len(modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(modules))
	}
}

func TestDisableModule_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "dismod", Name: "Dis Mod", PlanID: "starter",
		PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
	})

	svc.EnableModule(ctx, db.EnableModuleForTenantParams{
		TenantID: tenant.ID, ModuleID: "docs", Config: []byte("{}"), EnabledBy: "admin",
	})

	err := svc.DisableModule(ctx, tenant.ID, "docs")
	if err != nil {
		t.Fatalf("disable module: %v", err)
	}

	modules, _ := svc.GetTenantModules(ctx, tenant.ID)
	if len(modules) != 0 {
		t.Errorf("expected 0 modules after disable, got %d", len(modules))
	}
}

func TestUpdateTenant_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "updateme", Name: "Original", PlanID: "starter",
		PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
	})

	err := svc.UpdateTenant(ctx, db.UpdateTenantParams{
		ID:       tenant.ID,
		Name:     "Updated Name",
		PlanID:   "business",
		Settings: []byte(`{"theme":"dark"}`),
	})
	if err != nil {
		t.Fatalf("update tenant: %v", err)
	}

	got, _ := svc.GetTenant(ctx, "updateme")
	if got.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", got.Name)
	}
	if got.PlanID != "business" {
		t.Errorf("expected plan_id 'business', got %q", got.PlanID)
	}
}

func TestDisableTenant_and_EnableTenant_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, db.CreateTenantParams{
		Slug: "toggleme", Name: "Toggle", PlanID: "starter",
		PostgresUrl: "pg://x", RedisUrl: "redis://x", Settings: []byte("{}"),
	})

	// Disable
	if err := svc.DisableTenant(ctx, tenant.ID); err != nil {
		t.Fatalf("disable: %v", err)
	}
	// GetTenant filters by enabled=true, so disabled tenant returns ErrTenantNotFound
	_, err := svc.GetTenant(ctx, "toggleme")
	if err != ErrTenantNotFound {
		t.Fatalf("expected ErrTenantNotFound for disabled tenant, got: %v", err)
	}
	// Verify directly in DB that enabled=false
	var enabled bool
	pool.QueryRow(ctx, `SELECT enabled FROM tenants WHERE id = $1`, tenant.ID).Scan(&enabled)
	if enabled {
		t.Error("expected enabled=false in DB after DisableTenant")
	}

	// Re-enable
	if err := svc.EnableTenant(ctx, tenant.ID); err != nil {
		t.Fatalf("enable: %v", err)
	}
	got, err := svc.GetTenant(ctx, "toggleme")
	if err != nil {
		t.Fatalf("expected tenant visible after re-enable: %v", err)
	}
	if !got.Enabled {
		t.Error("expected tenant enabled after EnableTenant")
	}
}

func TestListModules_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	modules, err := svc.ListModules(context.Background())
	if err != nil {
		t.Fatalf("list modules: %v", err)
	}
	// Seed has 3 modules: chat, docs, fleet
	if len(modules) != 3 {
		t.Errorf("expected 3 seeded modules, got %d", len(modules))
	}
}

func TestToggleFeatureFlag_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	// Seed a flag
	pool.Exec(ctx, `INSERT INTO feature_flags (id, name, enabled) VALUES ('ff-1', 'dark_mode', false)`)

	// Toggle on
	if err := svc.ToggleFeatureFlag(ctx, "ff-1", true); err != nil {
		t.Fatalf("toggle on: %v", err)
	}

	flags, _ := svc.ListFeatureFlags(ctx)
	found := false
	for _, f := range flags {
		if f.ID == "ff-1" {
			found = true
			if !f.Enabled {
				t.Error("expected flag enabled after toggle on")
			}
		}
	}
	if !found {
		t.Error("flag ff-1 not found in list")
	}

	// Toggle off
	svc.ToggleFeatureFlag(ctx, "ff-1", false)
	flags, _ = svc.ListFeatureFlags(ctx)
	for _, f := range flags {
		if f.ID == "ff-1" && f.Enabled {
			t.Error("expected flag disabled after toggle off")
		}
	}
}

func TestToggleFeatureFlag_NotFound_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	err := svc.ToggleFeatureFlag(context.Background(), "nonexistent", true)
	if err != ErrFlagNotFound {
		t.Fatalf("expected ErrFlagNotFound, got: %v", err)
	}
}

func TestSetConfig_and_GetConfig_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	err := svc.SetConfig(ctx, "maintenance_mode", []byte(`true`), "admin")
	if err != nil {
		t.Fatalf("set config: %v", err)
	}

	config, err := svc.GetConfig(ctx, "maintenance_mode")
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if string(config.Value) != "true" {
		t.Errorf("expected value 'true', got %q", string(config.Value))
	}
}

func TestGetConfig_NotFound_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	_, err := svc.GetConfig(context.Background(), "nonexistent")
	if err != ErrConfigNotFound {
		t.Fatalf("expected ErrConfigNotFound, got: %v", err)
	}
}
