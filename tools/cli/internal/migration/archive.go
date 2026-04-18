package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ArchiveAll universally captures every legacy MySQL row the explicit migrator
// set did not land into a first-class SDA table, writing the row as JSONB into
// erp_legacy_archive. This is the safety net that lets us claim "no silent
// data loss": if a row exists in legacy and doesn't belong to the HTX-system
// / deprecated skip set, it lives in the archive.
//
// Two filtering rules:
//   1. Skip tables already handled by a real migrator (their names come from
//      the orchestrator's registry — passed in as `covered`).
//   2. Skip tables that match the HTX / calendar / system-noise prefixes —
//      these are Histrix UI/access-log tables with no business value, the
//      user explicitly called them out as safe to drop.
//
// Everything else is archived. For tables with a single BIGINT PK we populate
// legacy_pk_num for indexed lookups; composite / varchar PKs go into legacy_pk
// as a canonical string.
func ArchiveAll(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, covered map[string]struct{}) error {
	tables, err := listLegacyDataTables(ctx, mysqlDB)
	if err != nil {
		return fmt.Errorf("list legacy tables: %w", err)
	}
	slog.Info("archive: starting", "total_tables", len(tables))

	archived := 0
	rowsTotal := int64(0)
	for _, t := range tables {
		if _, ok := covered[t.name]; ok {
			continue
		}
		if isArchiveSkip(t.name) {
			continue
		}
		n, err := archiveOne(ctx, mysqlDB, pgPool, tenantID, t)
		if err != nil {
			// Archive is best-effort — log and keep going. The bigger the
			// coverage the better, but a single bad table must not fail
			// the whole run.
			slog.Warn("archive: table failed", "table", t.name, "err", err)
			continue
		}
		rowsTotal += n
		archived++
	}
	slog.Info("archive: done", "tables_archived", archived, "rows_archived", rowsTotal)
	return nil
}

// isArchiveSkip returns true for tables that carry no business value (HTX
// access logs, sabredav calendar, system mail queues). Everything else is
// a migration candidate.
func isArchiveSkip(name string) bool {
	// Exclude HTX* legacy-portal tables — login, UI, mail queues, logs.
	if strings.HasPrefix(name, "HTX") || strings.HasPrefix(name, "Htx") || strings.HasPrefix(name, "htx") {
		return true
	}
	// sabre/dav CalDAV stuff
	lower := strings.ToLower(name)
	switch lower {
	case "calendar", "calendars", "calendarchanges", "calendarobjects",
		"calendarsubscriptions", "cards", "aperturas", "addressbooks",
		"htxaddressbook", "principals", "schedulingobjects", "locks",
		"propertystorage", "users", "groupmembers":
		return true
	}
	// Deprecated "_OLD" snapshots
	if strings.HasSuffix(name, "_OLD") {
		return true
	}
	// Asterisk CDR (VoIP logs)
	if strings.HasPrefix(name, "ASTERIX") || strings.HasPrefix(name, "asterisk") {
		return true
	}
	return false
}

type legacyTable struct {
	name    string
	pk      string // single-column BIGINT PK if any; "" otherwise
	pkIsInt bool
}

// listLegacyDataTables asks information_schema for every base table in the
// current database and detects the primary key shape. When the PK is a single
// BIGINT we record that column so archiving can populate legacy_pk_num.
func listLegacyDataTables(ctx context.Context, mysqlDB *sql.DB) ([]legacyTable, error) {
	rows, err := mysqlDB.QueryContext(ctx, `
		SELECT t.table_name
		FROM information_schema.tables t
		WHERE t.table_schema = DATABASE() AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]legacyTable, 0, len(names))
	for _, n := range names {
		pk, isInt, err := detectPK(ctx, mysqlDB, n)
		if err != nil {
			// Unknown PK shape: treat as composite, archiver will hash.
			pk, isInt = "", false
		}
		out = append(out, legacyTable{name: n, pk: pk, pkIsInt: isInt})
	}
	return out, nil
}

// detectPK returns the single-column PRIMARY KEY column name + whether it is
// an integer type. Composite PKs return ("", false).
func detectPK(ctx context.Context, db *sql.DB, table string) (string, bool, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT k.column_name, c.data_type
		FROM information_schema.key_column_usage k
		JOIN information_schema.columns c
		  ON c.table_schema = k.table_schema
		 AND c.table_name = k.table_name
		 AND c.column_name = k.column_name
		WHERE k.table_schema = DATABASE()
		  AND k.table_name = ?
		  AND k.constraint_name = 'PRIMARY'
		ORDER BY k.ordinal_position
	`, table)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = rows.Close() }()

	type col struct {
		name string
		typ  string
	}
	var pk []col
	for rows.Next() {
		var c col
		if err := rows.Scan(&c.name, &c.typ); err != nil {
			return "", false, err
		}
		pk = append(pk, c)
	}
	if err := rows.Err(); err != nil {
		return "", false, err
	}
	if len(pk) != 1 {
		return "", false, nil
	}
	isInt := strings.Contains(pk[0].typ, "int")
	return pk[0].name, isInt, nil
}

// archiveOne streams rows from a single legacy table into erp_legacy_archive
// using COPY from a staging temp table, then INSERT ... SELECT ... ON CONFLICT
// DO NOTHING for idempotency. Returns the number of rows persisted.
func archiveOne(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, t legacyTable) (int64, error) {
	// Learn the columns so we can build JSONB payloads from the full row.
	colNames, err := listTableColumns(ctx, mysqlDB, t.name)
	if err != nil {
		return 0, fmt.Errorf("list cols %s: %w", t.name, err)
	}
	if len(colNames) == 0 {
		return 0, nil
	}

	// Select every column; MySQL driver surface them as []byte / strings /
	// numbers and we let the JSON encoder normalise.
	sel := "`" + strings.Join(colNames, "`, `") + "`"
	order := "1"
	if t.pk != "" {
		order = "`" + t.pk + "`"
	}

	// Keyset pagination on single-int PK, plain offset pagination otherwise.
	const batchSize = 5000
	lastInt := int64(0)
	offset := 0
	total := int64(0)

	for {
		var q string
		var args []any
		switch {
		case t.pkIsInt:
			q = fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s` > ? ORDER BY `%s` LIMIT ?",
				sel, t.name, t.pk, t.pk)
			args = []any{lastInt, batchSize}
		default:
			q = fmt.Sprintf("SELECT %s FROM `%s` ORDER BY %s LIMIT ? OFFSET ?",
				sel, t.name, order)
			args = []any{batchSize, offset}
		}

		rows, err := mysqlDB.QueryContext(ctx, q, args...)
		if err != nil {
			return total, fmt.Errorf("scan %s: %w", t.name, err)
		}

		batchRows, err := readBatchAsJSON(rows, colNames)
		_ = rows.Close()
		if err != nil {
			return total, fmt.Errorf("read batch %s: %w", t.name, err)
		}
		if len(batchRows) == 0 {
			break
		}

		// Build COPY rows: (tenant_id, legacy_table, legacy_pk, legacy_pk_num, data, migrated_at).
		copyRows := make([][]any, len(batchRows))
		for i, r := range batchRows {
			pkStr := ""
			var pkNum *int64
			if t.pk != "" {
				pkStr = fmt.Sprintf("%v", r[t.pk])
				if t.pkIsInt {
					if v, ok := toInt64(r[t.pk]); ok {
						pkNum = &v
						if v > lastInt {
							lastInt = v
						}
					}
				}
			} else {
				pkStr = canonicalCompositePK(colNames, r)
			}
			jsonPayload, _ := json.Marshal(r)
			copyRows[i] = []any{tenantID, t.name, pkStr, pkNum, string(jsonPayload), time.Now()}
		}

		if err := copyIntoArchive(ctx, pgPool, tenantID, copyRows); err != nil {
			return total, fmt.Errorf("archive %s: %w", t.name, err)
		}
		total += int64(len(batchRows))
		offset += len(batchRows)
		if len(batchRows) < batchSize {
			break
		}
	}

	slog.Info("archive: table done", "table", t.name, "rows", total)
	return total, nil
}

func listTableColumns(ctx context.Context, mysqlDB *sql.DB, table string) ([]string, error) {
	rows, err := mysqlDB.QueryContext(ctx, `
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
		ORDER BY ordinal_position
	`, table)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// readBatchAsJSON turns a *sql.Rows into []map[col]any suitable for JSON
// encoding. Bytes-slice columns are decoded to string; everything else we
// pass through.
func readBatchAsJSON(rows *sql.Rows, cols []string) ([]map[string]any, error) {
	var out []map[string]any
	values := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			switch v := values[i].(type) {
			case []byte:
				row[c] = string(v)
			default:
				row[c] = v
			}
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int32:
		return int64(x), true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		var n int64
		_, err := fmt.Sscanf(x, "%d", &n)
		return n, err == nil
	case []byte:
		var n int64
		_, err := fmt.Sscanf(string(x), "%d", &n)
		return n, err == nil
	}
	return 0, false
}

func canonicalCompositePK(cols []string, row map[string]any) string {
	// Stable canonical rendering: col1=val|col2=val|...
	parts := make([]string, 0, len(cols))
	for _, c := range cols {
		parts = append(parts, fmt.Sprintf("%s=%v", c, row[c]))
	}
	return strings.Join(parts, "|")
}

// copyIntoArchive writes one batch of archive rows using COPY into a temp
// table then INSERT ... SELECT ... ON CONFLICT DO NOTHING into the real one.
// The conflict key is (tenant_id, legacy_table, legacy_pk).
func copyIntoArchive(ctx context.Context, pgPool *pgxpool.Pool, tenantID string, rows [][]any) error {
	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin archive tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		return fmt.Errorf("set tenant: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		CREATE TEMP TABLE IF NOT EXISTS staging_erp_legacy_archive
		(LIKE erp_legacy_archive INCLUDING DEFAULTS) ON COMMIT DROP
	`); err != nil {
		return fmt.Errorf("create staging: %w", err)
	}
	if _, err := tx.Exec(ctx, "TRUNCATE staging_erp_legacy_archive"); err != nil {
		return fmt.Errorf("truncate staging: %w", err)
	}

	if _, err := tx.CopyFrom(ctx,
		pgx.Identifier{"staging_erp_legacy_archive"},
		[]string{"tenant_id", "legacy_table", "legacy_pk", "legacy_pk_num", "data", "migrated_at"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return fmt.Errorf("copy into staging: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO erp_legacy_archive
			(tenant_id, legacy_table, legacy_pk, legacy_pk_num, data, migrated_at)
		SELECT tenant_id, legacy_table, legacy_pk, legacy_pk_num, data, migrated_at
		FROM staging_erp_legacy_archive
		ON CONFLICT (tenant_id, legacy_table, legacy_pk) DO NOTHING
	`); err != nil {
		return fmt.Errorf("insert from staging: %w", err)
	}

	return tx.Commit(ctx)
}

// CoveredLegacyTables returns the set of legacy tables the registered
// migrators already own, so ArchiveAll can skip them. Feeds on the
// orchestrator's reader list.
func CoveredLegacyTables(o *Orchestrator) map[string]struct{} {
	covered := make(map[string]struct{}, len(o.readers))
	for _, m := range o.readers {
		covered[m.LegacyTable()] = struct{}{}
	}
	return covered
}

// archiveSkippedRows persists the raw MySQL rows that a migrator transform
// rejected (returned nil). Each row lands in erp_legacy_archive with its
// original column payload as JSONB, keyed by the legacy table name. Used by
// runTablePipeline when archiveSkipped is enabled so no row is ever silently
// lost — if the structured migrator can't find a home for it, the archive
// holds on to the full source data for later reprocessing.
func archiveSkippedRows(ctx context.Context, tx pgx.Tx, tenantID, legacyTable string, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	// One multi-valued INSERT per batch — fewer than ~500 skipped rows per
	// batch in practice, so parameter limit is not a concern.
	for _, row := range rows {
		payload, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("marshal skip: %w", err)
		}
		// Stable canonical PK: when the transform exposes a single-int PK we
		// prefer it; otherwise fall back to a composite hash. Collisions are
		// handled by ON CONFLICT DO NOTHING.
		pkStr, pkNum := firstPKCandidate(row)
		if pkStr == "" {
			pkStr = fmt.Sprintf("skip_%d", time.Now().UnixNano())
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO erp_legacy_archive
			  (tenant_id, legacy_table, legacy_pk, legacy_pk_num, data)
			VALUES ($1, $2, $3, $4, $5::jsonb)
			ON CONFLICT (tenant_id, legacy_table, legacy_pk) DO NOTHING
		`, tenantID, legacyTable, pkStr, pkNum, string(payload)); err != nil {
			return fmt.Errorf("archive skip row: %w", err)
		}
	}
	return nil
}

// firstPKCandidate picks the most likely PK column/value from a legacy row.
// Heuristic: any int column whose name starts with "id_" or equals "id" or
// ends with "_id" is treated as a PK candidate. Composite PKs fall back to
// a canonical string representation of all fields.
func firstPKCandidate(row map[string]any) (string, *int64) {
	for _, k := range pkPreferredNames {
		if v, ok := row[k]; ok && v != nil {
			if n, ok := toInt64(v); ok && n > 0 {
				return fmt.Sprintf("%d", n), &n
			}
			s := fmt.Sprintf("%v", v)
			if s != "" && s != "0" {
				return s, nil
			}
		}
	}
	// Fallback: canonical composite render of all columns.
	return canonicalCompositeRow(row), nil
}

var pkPreferredNames = []string{
	"id", "Id_usuario", "Id_perfil", "Id_curso", "IdPersona",
	"id_regcuenta", "id_stkarticulo", "id_stkmovimiento", "id_cpsmovimiento",
	"id_cpsdetalle", "id_ivaventa", "id_ivacompra", "id_ivaimporte",
	"id_retganan", "id_retiva", "id_ret1598", "id_detal", "id_pieza",
	"id_movimiento", "id_detalle", "id_proceso", "id_ejercicio",
	"id_ctbcuenta", "id_ctbcentro", "id_carbanco", "id_cajpuesto",
	"id_cajpuestoarqueo", "id_carmovimiento", "id_controlmovim",
	"id_prodcontrol", "id_centro_productivo", "id_mrporden",
	"id_stkdeposito", "id_stklista", "idCotiz", "idRemito", "idRemdet",
	"idDesc", "id_movdemerito", "id_retacumu", "id_stkbomhist",
	"id_accidente", "id_riesgo", "id_carmedico",
}

func canonicalCompositeRow(row map[string]any) string {
	// Sort keys so output is deterministic.
	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sortStrings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, row[k]))
	}
	return strings.Join(parts, "|")
}

// sortStrings is a tiny sort helper kept local so archive.go doesn't pull in
// sort just for this one call.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
