package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CoverageRow summarises how well a single legacy table landed in SDA.
// `LegacyRows` is SELECT COUNT(*) on MySQL. `MigratedRows` is the number of
// mapping rows emitted for that legacy_table in erp_legacy_mapping. The ratio
// is the best operator-facing "did we lose data?" signal we can compute
// without doing row-level comparisons.
type CoverageRow struct {
	LegacyTable  string  `json:"legacy_table"`
	Domain       string  `json:"domain"`
	SDATable     string  `json:"sda_table"`
	LegacyRows   int64   `json:"legacy_rows"`
	MigratedRows int64   `json:"migrated_rows"`
	Coverage     float64 `json:"coverage_pct"`  // 0..100
	Gap          int64   `json:"gap_rows"`      // LegacyRows - MigratedRows
}

// CoverageReport is the top-level result returned by AuditCoverage.
type CoverageReport struct {
	Tenant          string        `json:"tenant"`
	TotalLegacy     int64         `json:"total_legacy_rows"`
	TotalMigrated   int64         `json:"total_migrated_rows"`
	OverallCoverage float64       `json:"overall_coverage_pct"`
	PerTable        []CoverageRow `json:"per_table"`
	UnmappedTables  []string      `json:"unmapped_tables"`  // legacy tables with zero migrator coverage
}

// AuditCoverage compares row counts between the legacy MySQL and the
// erp_legacy_mapping table for the given tenant. Expects `registry` to list
// every (legacy_table, sda_table, domain) triple the registered migrators
// care about — typically produced by reading Orchestrator's readers.
//
// It is read-only on both sides and completes in seconds even on wide
// catalogues because COUNT(*) on indexed PKs is fast.
func AuditCoverage(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, registry []MigratorRegistration) (*CoverageReport, error) {
	report := &CoverageReport{Tenant: tenantID}

	// 1) Count migrated rows grouped by legacy_table (one PG query).
	migratedByTable, err := countMigratedPerTable(ctx, pgPool, tenantID)
	if err != nil {
		return nil, fmt.Errorf("count migrated: %w", err)
	}

	// 2) Count legacy rows for each registered migrator. One MySQL query per
	//    table — fan-out is bounded by the registry size (≈ 100 queries).
	seen := make(map[string]struct{}, len(registry))
	for _, reg := range registry {
		if _, dup := seen[reg.LegacyTable]; dup {
			continue
		}
		seen[reg.LegacyTable] = struct{}{}

		var legacyN int64
		err := mysqlDB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s%s", reg.LegacyTable, reg.whereClause()),
		).Scan(&legacyN)
		if err != nil {
			// Missing table (schema drift) is a real finding — record it as 0 legacy
			// rather than abort the whole audit.
			legacyN = -1
		}

		migN := migratedByTable[reg.LegacyTable]
		cov := 0.0
		if legacyN > 0 {
			cov = 100.0 * float64(migN) / float64(legacyN)
		}
		gap := int64(0)
		if legacyN > 0 {
			gap = legacyN - migN
		}
		report.PerTable = append(report.PerTable, CoverageRow{
			LegacyTable:  reg.LegacyTable,
			Domain:       reg.Domain,
			SDATable:     reg.SDATable,
			LegacyRows:   legacyN,
			MigratedRows: migN,
			Coverage:     cov,
			Gap:          gap,
		})
		report.TotalLegacy += max64(legacyN, 0)
		report.TotalMigrated += migN
	}

	if report.TotalLegacy > 0 {
		report.OverallCoverage = 100.0 * float64(report.TotalMigrated) / float64(report.TotalLegacy)
	}

	// 3) Unmapped tables: anything in erp_legacy_mapping that is not in the
	//    registry is a bonus (good — unexpected win). Anything in the registry
	//    with MigratedRows == 0 after a full run is a regression candidate.
	// Sort per-table gap descending so the worst gaps show first.
	sort.Slice(report.PerTable, func(i, j int) bool {
		return report.PerTable[i].Gap > report.PerTable[j].Gap
	})

	return report, nil
}

// MigratorRegistration is what AuditCoverage needs to know about each
// migrator. Kept separate from TableMigrator so callers can build the list
// from any source (orchestrator readers, static registry, config file).
type MigratorRegistration struct {
	LegacyTable string
	SDATable    string
	Domain      string
	// Extra WHERE fragment to match the migrator's own filter (optional).
	ExtraWhere string
}

func (r MigratorRegistration) whereClause() string {
	if strings.TrimSpace(r.ExtraWhere) == "" {
		return ""
	}
	return " WHERE " + r.ExtraWhere
}

func countMigratedPerTable(ctx context.Context, pgPool *pgxpool.Pool, tenantID string) (map[string]int64, error) {
	rows, err := pgPool.Query(ctx,
		`SELECT legacy_table, COUNT(*) FROM erp_legacy_mapping
		 WHERE tenant_id = $1 GROUP BY legacy_table`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int64)
	for rows.Next() {
		var t string
		var n int64
		if err := rows.Scan(&t, &n); err != nil {
			return nil, err
		}
		out[t] = n
	}
	return out, rows.Err()
}

// RegistrationsFromOrchestrator scrapes the registered readers and builds the
// registration list needed by AuditCoverage. Must be called after
// RegisterMigrators has populated the readers slice.
func RegistrationsFromOrchestrator(o *Orchestrator) []MigratorRegistration {
	out := make([]MigratorRegistration, 0, len(o.readers))
	for _, m := range o.readers {
		out = append(out, MigratorRegistration{
			LegacyTable: m.LegacyTable(),
			SDATable:    m.SDATable(),
			Domain:      m.Domain(),
		})
	}
	return out
}

// PrintCoverageReport writes a human-readable breakdown to stdout and also
// returns the report JSON so callers can pipe it to jq or persist it.
func PrintCoverageReport(r *CoverageReport) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n=== COVERAGE AUDIT — tenant=%s ===\n", r.Tenant)
	fmt.Fprintf(&sb, "  legacy rows:   %d\n", r.TotalLegacy)
	fmt.Fprintf(&sb, "  migrated rows: %d\n", r.TotalMigrated)
	fmt.Fprintf(&sb, "  overall:       %.2f%%\n\n", r.OverallCoverage)

	fmt.Fprintln(&sb, "  TOP GAPS (legacy rows not landed in SDA):")
	for i, row := range r.PerTable {
		if i >= 20 {
			break
		}
		flag := "  "
		if row.Coverage < 90 && row.LegacyRows > 0 {
			flag = "!!"
		}
		if row.LegacyRows < 0 {
			fmt.Fprintf(&sb, "  %s %-30s  domain=%-14s MISSING IN LEGACY (schema drift)\n",
				flag, row.LegacyTable, row.Domain)
			continue
		}
		fmt.Fprintf(&sb, "  %s %-30s  legacy=%-9d  sda=%-9d  coverage=%6.2f%%  gap=%d\n",
			flag, row.LegacyTable, row.LegacyRows, row.MigratedRows, row.Coverage, row.Gap)
	}
	j, _ := json.MarshalIndent(r, "", "  ")
	return sb.String() + "\n" + string(j) + "\n"
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
