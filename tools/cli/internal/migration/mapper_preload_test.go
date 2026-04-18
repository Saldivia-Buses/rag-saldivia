package migration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	_ "github.com/go-sql-driver/mysql"
)

// mustHavePG spins up the bench pool or skips. Same gate as bench_test.go so a
// single BENCH_PG_DSN covers every integration-style migration test.
func mustHavePG(t *testing.T) (context.Context, string) {
	t.Helper()
	dsn := os.Getenv("BENCH_PG_DSN")
	if dsn == "" {
		t.Skip("BENCH_PG_DSN not set — skipping PG integration test")
	}
	return context.Background(), dsn
}

// TestPreloadDomainRepopulatesEmptyCache is the Phase 0 migration-integrity
// regression test for the silent-skip bug fixed in this commit.
//
// The bug: every Build*Index (BuildLegajoIndex / BuildRegMovimIndex /
// BuildRemitoIndex / BuildNroCuentaIndex) looked up its source UUIDs via the
// mapper's in-memory cache. In a fresh run the cache was populated by the
// parent migrator's Map() calls; in a resume or after-table-hook replay the
// cache started empty and the indexes came out empty — causing 5 migrators
// (FICHADADIA, RHDESCUENTOS, RRHH_ADICIONALES, FACDETAL, REMDETAL) to mark
// 100% of their rows as skipped while the run still reported status=completed.
//
// The fix is a PreloadDomain() call at the top of each Build*Index. This test
// seeds erp_legacy_mapping directly, confirms the cache starts empty, calls
// PreloadDomain, and confirms the cache is now populated. If this test
// passes, every downstream Build*Index will find its parent UUIDs.
func TestPreloadDomainRepopulatesEmptyCache(t *testing.T) {
	ctx, dsn := mustHavePG(t)
	pool := openBenchPool(t, dsn)
	tenantID := fmt.Sprintf("test-preload-%s", uuid.NewString()[:8])

	persona1, persona2 := int64(42), int64(99)
	uuid1, uuid2 := uuid.New(), uuid.New()
	_, err := pool.Exec(ctx,
		`INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
		 VALUES ($1,'entity','PERSONAL',$2,$3),($1,'entity','PERSONAL',$4,$5)`,
		tenantID, persona1, uuid1, persona2, uuid2)
	if err != nil {
		t.Fatalf("seed erp_legacy_mapping: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM erp_legacy_mapping WHERE tenant_id = $1`, tenantID)
	})

	m := NewMapper(pool, tenantID)
	if got := m.cacheCount("entity", "PERSONAL"); got != 0 {
		t.Fatalf("fresh mapper: expected empty cache, got %d entries", got)
	}

	if err := m.PreloadDomain(ctx, "entity"); err != nil {
		t.Fatalf("preload: %v", err)
	}
	if got := m.cacheCount("entity", "PERSONAL"); got != 2 {
		t.Errorf("after preload: expected 2 cached mappings, got %d", got)
	}
}

// TestBuildLegajoIndexEndToEnd is the evidence that a run with an empty
// in-memory cache still produces a populated legajo index — the behaviour
// every one of FICHADADIA / RHDESCUENTOS / RRHH_ADICIONALES depends on to
// stop skipping 100% of their rows in prod.
//
// Setup:
//  1. PG has `erp_legacy_mapping` rows for two PERSONAL entries (IdPersona
//     100 → UUID-A, IdPersona 101 → UUID-B). That models the state after
//     the PERSONAL migrator successfully ran and flushed its mappings.
//  2. MySQL has three PERSONAL rows: (100, legajo=1001), (101, legajo=1002),
//     (102, legajo=0). The third is the "skip — legajo missing" case.
//  3. The mapper is freshly constructed — cache is empty, as it would be on
//     a resume or hook replay.
//
// Then BuildLegajoIndex runs. Assertions:
//  - cache[entity:PERSONAL] is populated (by the PreloadDomain call this
//    fix added).
//  - ResolveByLegajo(1001) returns UUID-A.
//  - ResolveByLegajo(1002) returns UUID-B.
//  - ResolveByLegajo(0) returns (Nil, false).
//
// Before the fix, step 3 left the cache empty — so BuildLegajoIndex's
// `cache[key][idPersona]` lookup missed every row and legajoIndex ended up
// with zero entries. This test would have asserted (Nil, false) for 1001
// and 1002 (because legajoIndex doesn't have them) — matching the silent-
// skip prod behaviour. After the fix it returns the correct UUIDs.
func TestBuildLegajoIndexEndToEnd(t *testing.T) {
	ctx, dsn := mustHavePG(t)
	mysqlDSN := os.Getenv("TEST_MYSQL_DSN")
	if mysqlDSN == "" {
		t.Skip("TEST_MYSQL_DSN not set — skipping end-to-end BuildLegajoIndex test")
	}
	pool := openBenchPool(t, dsn)
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	t.Cleanup(func() { _ = mysqlDB.Close() })
	if err := mysqlDB.PingContext(ctx); err != nil {
		t.Fatalf("mysql ping: %v", err)
	}

	tenantID := fmt.Sprintf("test-bli-e2e-%s", uuid.NewString()[:8])
	sdaA, sdaB := uuid.New(), uuid.New()
	_, err = pool.Exec(ctx,
		`INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
		 VALUES ($1,'entity','PERSONAL',100,$2),($1,'entity','PERSONAL',101,$3)`,
		tenantID, sdaA, sdaB)
	if err != nil {
		t.Fatalf("seed mapping: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM erp_legacy_mapping WHERE tenant_id = $1`, tenantID)
	})

	m := NewMapper(pool, tenantID)
	if got := m.cacheCount("entity", "PERSONAL"); got != 0 {
		t.Fatalf("expected empty cache before BuildLegajoIndex; got %d", got)
	}

	if err := m.BuildLegajoIndex(ctx, mysqlDB); err != nil {
		t.Fatalf("BuildLegajoIndex: %v", err)
	}

	if got := m.cacheCount("entity", "PERSONAL"); got != 2 {
		t.Errorf("cache should hold 2 PERSONAL mappings after Preload; got %d", got)
	}
	if got, ok := m.ResolveByLegajo(1001); !ok || got != sdaA {
		t.Errorf("ResolveByLegajo(1001) = (%v, %v), want (%v, true)", got, ok, sdaA)
	}
	if got, ok := m.ResolveByLegajo(1002); !ok || got != sdaB {
		t.Errorf("ResolveByLegajo(1002) = (%v, %v), want (%v, true)", got, ok, sdaB)
	}
	if _, ok := m.ResolveByLegajo(0); ok {
		t.Errorf("ResolveByLegajo(0) should be (Nil, false)")
	}
}

// TestBuildLegajoIndexPreloadsEntityCache confirms BuildLegajoIndex itself
// invokes PreloadDomain before it starts looking up PERSONAL UUIDs from the
// cache — the specific code path that used to fail silently when the cache
// was empty on a resume.
//
// MySQL is nil so the PERSONAL scan will error out, but by the time it does
// PreloadDomain has already populated the entity cache from erp_legacy_mapping.
// Asserting the cache is filled proves the PreloadDomain call happens first.
func TestBuildLegajoIndexPreloadsEntityCache(t *testing.T) {
	ctx, dsn := mustHavePG(t)
	pool := openBenchPool(t, dsn)
	tenantID := fmt.Sprintf("test-bli-%s", uuid.NewString()[:8])

	persona := int64(7)
	sdaID := uuid.New()
	if _, err := pool.Exec(ctx,
		`INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
		 VALUES ($1, 'entity', 'PERSONAL', $2, $3)`, tenantID, persona, sdaID); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`DELETE FROM erp_legacy_mapping WHERE tenant_id = $1`, tenantID)
	})

	m := NewMapper(pool, tenantID)

	// nil *sql.DB will panic or error inside BuildLegajoIndex once it tries
	// to scan PERSONAL. Either way, PreloadDomain ran first — so the entity
	// cache should be populated with the seeded mapping.
	func() {
		defer func() { _ = recover() }()
		var mysqlDB *sql.DB
		_ = m.BuildLegajoIndex(ctx, mysqlDB)
	}()

	if got := m.cacheCount("entity", "PERSONAL"); got != 1 {
		t.Errorf("entity cache should be populated by PreloadDomain before BuildLegajoIndex hits MySQL; got %d entries", got)
	}
}
