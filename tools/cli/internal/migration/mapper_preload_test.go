package migration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
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
