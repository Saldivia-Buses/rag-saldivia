package migration

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pendingMapping is a mapping waiting to be flushed to PG in batch.
type pendingMapping struct {
	domain, table string
	legacyID      int64
	sdaID         uuid.UUID
}

// Mapper handles legacy INT → SDA UUID mapping with in-memory cache.
type Mapper struct {
	pool                  *pgxpool.Pool
	tenantID              string
	dryRun                bool
	mu                    sync.RWMutex
	cache                 map[string]map[int64]uuid.UUID  // "domain:table" → legacy_id → uuid
	codeIndex             map[string]map[string]uuid.UUID // "domain:table" → code → uuid (for varchar PK tables)
	dateCache             map[uuid.UUID]time.Time         // entry UUID → date (for journal line date resolution)
	regMovimIndex         map[int64]uuid.UUID             // regmovim_id → invoice UUID (for FACDETAL FK resolution)
	legajoIndex           map[int64]uuid.UUID             // PERSONAL.legajo → entity UUID (HR FK resolution)
	nroCuentaIndex        map[int64]uuid.UUID             // REG_CUENTA.nro_cuenta → entity UUID (ctacod ambiguity)
	remitoByID            map[int64]uuid.UUID             // REMITO.idRemito → REMITO UUID (REMDETAL FK resolution)
	unassignedWarehouseID uuid.UUID                       // fallback warehouse for STK_MOVIMIENTOS with stkdeposito_id=0
	unknownEntityID       uuid.UUID                       // fallback entity for rows whose ctacod cannot be resolved
	pending               []pendingMapping                // batch of mappings to flush
}

// NewMapper creates a new ID mapper for a tenant.
func NewMapper(pool *pgxpool.Pool, tenantID string) *Mapper {
	return &Mapper{
		pool:        pool,
		tenantID:    tenantID,
		cache:       make(map[string]map[int64]uuid.UUID),
		codeIndex:   make(map[string]map[string]uuid.UUID),
		dateCache:   make(map[uuid.UUID]time.Time),
		legajoIndex: make(map[int64]uuid.UUID),
		remitoByID:  make(map[int64]uuid.UUID),
	}
}

// SetDryRun enables dry-run mode: all lookups return deterministic UUIDs without hitting PG.
func (m *Mapper) SetDryRun(v bool) { m.dryRun = v }

// EnsureUnassignedWarehouse creates (or finds) a fallback warehouse used for
// STK_MOVIMIENTOS rows whose stkdeposito_id is 0 or points to a deleted depot.
// Histrix let users record stock movements without a depot, so blindly requiring
// a FK would drop ~1M rows in saldivia. The fallback preserves the movement
// (quantity, article, date) and marks it in notes so it can be reassigned later
// via an admin UI. Idempotent — safe to call from prod, dry-run, and resume.
func (m *Mapper) EnsureUnassignedWarehouse(ctx context.Context, pgPool *pgxpool.Pool) error {
	if m.dryRun {
		m.unassignedWarehouseID = m.deterministicUUID("stock", "UNASSIGNED_WAREHOUSE", "0")
		return nil
	}
	const code = "UNASSIGNED"
	var id uuid.UUID
	err := pgPool.QueryRow(ctx,
		`INSERT INTO erp_warehouses (tenant_id, code, name, location, active)
		 VALUES ($1, $2, 'Sin Asignar (Legacy)', 'Movimientos Histrix sin deposito — reasignar manualmente', true)
		 ON CONFLICT (tenant_id, code) DO UPDATE SET code = EXCLUDED.code
		 RETURNING id`, m.tenantID, code).Scan(&id)
	if err != nil {
		return fmt.Errorf("ensure unassigned warehouse: %w", err)
	}
	m.unassignedWarehouseID = id
	return nil
}

// UnassignedWarehouseID returns the fallback warehouse UUID for legacy stock
// movements without a depot. Returns uuid.Nil if EnsureUnassignedWarehouse was
// not called — callers must treat that as "skip row" rather than insert NULL.
func (m *Mapper) UnassignedWarehouseID() uuid.UUID {
	return m.unassignedWarehouseID
}

// EnsureUnknownEntity is the entity-side counterpart to EnsureUnassignedWarehouse.
// Several Histrix tables (RETACUMU, MOVDEMERITO, IVA*) reference entities through
// `ctacod` which can be either REG_CUENTA.id_regcuenta OR REG_CUENTA.nro_cuenta;
// the legacy codebase never normalised which one. When neither lookup succeeds
// we fall back to this synthetic "UNKNOWN" entity so the row still lands —
// flagged in the caller's notes with "[legacy:unknown_entity:<ctacod>]" so ops
// can reassign them from the admin UI.
func (m *Mapper) EnsureUnknownEntity(ctx context.Context, pgPool *pgxpool.Pool) error {
	if m.dryRun {
		m.unknownEntityID = m.deterministicUUID("entity", "UNKNOWN", "0")
		return nil
	}
	const code = "UNKNOWN-LEGACY"
	var id uuid.UUID
	err := pgPool.QueryRow(ctx,
		`INSERT INTO erp_entities (tenant_id, type, code, name, active)
		 VALUES ($1, 'customer', $2, 'Desconocido (Legacy)', false)
		 ON CONFLICT (tenant_id, type, code) DO UPDATE SET code = EXCLUDED.code
		 RETURNING id`, m.tenantID, code).Scan(&id)
	if err != nil {
		return fmt.Errorf("ensure unknown entity: %w", err)
	}
	m.unknownEntityID = id
	return nil
}

// UnknownEntityID returns the synthetic "UNKNOWN" entity UUID created by
// EnsureUnknownEntity, or uuid.Nil if that hook never ran (treat as skip).
func (m *Mapper) UnknownEntityID() uuid.UUID {
	return m.unknownEntityID
}

// BuildNroCuentaIndex builds nro_cuenta → entity UUID so migrators that carry
// ctacod (which is `nro_cuenta`, NOT `id_regcuenta`) can still resolve. Runs
// after NewEntityMigrator finishes. Cheap: REG_CUENTA is small (~5.5K rows
// on saldivia) and we already have the id_regcuenta → UUID cache.
func (m *Mapper) BuildNroCuentaIndex(ctx context.Context, mysqlDB *sql.DB) error {
	if m.dryRun {
		return nil
	}
	// Resume-safe: cache may be empty if this invocation didn't freshly
	// migrate REG_CUENTA. PreloadDomain rebuilds it from erp_legacy_mapping.
	if err := m.PreloadDomain(ctx, "entity"); err != nil {
		return fmt.Errorf("preload entity for nro_cuenta index: %w", err)
	}
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT id_regcuenta, nro_cuenta FROM REG_CUENTA WHERE nro_cuenta > 0`)
	if err != nil {
		return fmt.Errorf("build nro_cuenta index: %w", err)
	}
	defer func() { _ = rows.Close() }()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.nroCuentaIndex == nil {
		m.nroCuentaIndex = make(map[int64]uuid.UUID)
	}
	cacheKey := m.cacheKey("entity", "REG_CUENTA")
	count, orphans := 0, 0
	for rows.Next() {
		var idReg, nroCta int64
		if err := rows.Scan(&idReg, &nroCta); err != nil {
			return fmt.Errorf("scan nro_cuenta index: %w", err)
		}
		uid, ok := m.cache[cacheKey][idReg]
		if !ok {
			orphans++
			continue
		}
		m.nroCuentaIndex[nroCta] = uid
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate nro_cuenta index: %w", err)
	}
	fmt.Printf("  nro_cuenta index: loaded %d mappings (%d orphans)\n", count, orphans)
	return nil
}

// ResolveEntityFlexible tries to resolve an entity UUID from a ctacod field
// whose semantics vary by source table. Order: id_regcuenta → nro_cuenta →
// unknown entity fallback. Returns (uuid, "matched") where matched is one of
// "id_regcuenta", "nro_cuenta", "unknown". Caller uses the tag for tracing
// and can stash it in notes when the fallback fires.
func (m *Mapper) ResolveEntityFlexible(ctx context.Context, ctacod int64) (uuid.UUID, string) {
	if ctacod > 0 {
		if uid, err := m.ResolveOptional(ctx, "entity", "REG_CUENTA", ctacod); err == nil && uid != uuid.Nil {
			return uid, "id_regcuenta"
		}
		m.mu.RLock()
		if uid, ok := m.nroCuentaIndex[ctacod]; ok {
			m.mu.RUnlock()
			return uid, "nro_cuenta"
		}
		m.mu.RUnlock()
	}
	if m.unknownEntityID != uuid.Nil {
		return m.unknownEntityID, "unknown"
	}
	return uuid.Nil, ""
}

func (m *Mapper) cacheKey(domain, table string) string {
	return domain + ":" + table
}

// querier abstracts pgx.Tx and pgxpool.Pool for query execution.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// dryRunNamespace is a fixed UUIDv5 namespace for deterministic dry-run UUID
// generation. Keeping it stable across runs lets us diff two dry-runs or compare
// a dry-run against a prod migration for the same inputs.
var dryRunNamespace = uuid.MustParse("b7f9a3d4-2c1e-4b0a-9d8e-0f1a2b3c4d5e")

// deterministicUUID produces a v5 UUID from the parts, joined with '|'. Used
// exclusively in dry-run mode so outputs are reproducible.
func (m *Mapper) deterministicUUID(parts ...string) uuid.UUID {
	return uuid.NewSHA1(dryRunNamespace, []byte(m.tenantID+"|"+strings.Join(parts, "|")))
}

// Map returns the UUID for a legacy ID. If no mapping exists, generates a new one.
// When tx is non-nil, uses the transaction; otherwise falls back to the pool.
func (m *Mapper) Map(ctx context.Context, tx pgx.Tx, domain, table string, legacyID int64, legacyCreatedBy *string) (uuid.UUID, error) {
	if m.dryRun {
		return m.deterministicUUID(domain, table, fmt.Sprintf("%d", legacyID)), nil
	}

	key := m.cacheKey(domain, table)

	m.mu.RLock()
	if cached, ok := m.cache[key][legacyID]; ok {
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	// Generate new UUID, cache it, and queue for batch flush
	newID := uuid.New()
	m.mu.Lock()
	if m.cache[key] == nil {
		m.cache[key] = make(map[int64]uuid.UUID)
	}
	m.cache[key][legacyID] = newID
	m.pending = append(m.pending, pendingMapping{domain: domain, table: table, legacyID: legacyID, sdaID: newID})
	m.mu.Unlock()

	return newID, nil
}

// FlushPending bulk-inserts all pending mappings into erp_legacy_mapping.
// Call once per batch, inside the transaction.
//
// PG's extended-protocol hard cap is 65535 bind parameters per statement.
// Each mapping contributes 5 params → chunk at 12_000 rows for 64k headroom.
// Before this chunk loop, batches of ≥ 13107 mappings (not uncommon in
// CTB_MOVIMIENTOS or STK_BOM_HIST) failed with "extended protocol limited
// to 65535 parameters".
func (m *Mapper) FlushPending(ctx context.Context, q querier) error {
	m.mu.Lock()
	batch := m.pending
	m.pending = nil
	m.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}

	const chunkSize = 12_000
	for start := 0; start < len(batch); start += chunkSize {
		end := start + chunkSize
		if end > len(batch) {
			end = len(batch)
		}
		sub := batch[start:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id) VALUES ")
		args := make([]any, 0, len(sub)*5)
		for i, p := range sub {
			if i > 0 {
				sb.WriteString(",")
			}
			n := i * 5
			sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5))
			args = append(args, m.tenantID, p.domain, p.table, p.legacyID, p.sdaID)
		}
		sb.WriteString(" ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING")

		if _, err := q.Exec(ctx, sb.String(), args...); err != nil {
			return fmt.Errorf("flush %d mappings (chunk %d-%d): %w", len(batch), start, end, err)
		}
	}
	return nil
}

// Resolve looks up an existing mapping (for foreign keys). Returns error if not found.
func (m *Mapper) Resolve(ctx context.Context, domain, table string, legacyID int64) (uuid.UUID, error) {
	if legacyID == 0 {
		return uuid.Nil, fmt.Errorf("resolve %s/%s: legacy_id is 0", domain, table)
	}

	if m.dryRun {
		return m.deterministicUUID(domain, table, fmt.Sprintf("%d", legacyID)), nil
	}

	key := m.cacheKey(domain, table)

	m.mu.RLock()
	if cached, ok := m.cache[key][legacyID]; ok {
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	var sdaID uuid.UUID
	err := m.pool.QueryRow(ctx,
		`SELECT sda_id FROM erp_legacy_mapping
		 WHERE tenant_id = $1 AND domain = $2 AND legacy_table = $3 AND legacy_id = $4`,
		m.tenantID, domain, table, legacyID,
	).Scan(&sdaID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve %s/%s/%d: not found (migrate dependency first): %w", domain, table, legacyID, err)
	}

	m.mu.Lock()
	if m.cache[key] == nil {
		m.cache[key] = make(map[int64]uuid.UUID)
	}
	m.cache[key][legacyID] = sdaID
	m.mu.Unlock()

	return sdaID, nil
}

// ResolveOptional looks up a mapping, returning uuid.Nil if legacy ID is 0 or NULL.
func (m *Mapper) ResolveOptional(ctx context.Context, domain, table string, legacyID int64) (uuid.UUID, error) {
	if legacyID == 0 {
		return uuid.Nil, nil
	}
	return m.Resolve(ctx, domain, table, legacyID)
}

// BuildCodeIndex builds a code→UUID index for a target SDA table.
// Used for tables like CTB_CUENTAS where legacy FK is a varchar code, not an int.
func (m *Mapper) BuildCodeIndex(ctx context.Context, domain, sdaTable, codeColumn string) error {
	rows, err := m.pool.Query(ctx,
		fmt.Sprintf(`SELECT id, %s FROM %s WHERE tenant_id = $1`, codeColumn, sdaTable),
		m.tenantID,
	)
	if err != nil {
		return fmt.Errorf("build code index %s: %w", sdaTable, err)
	}
	defer rows.Close()

	key := m.cacheKey(domain, sdaTable)
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.codeIndex[key] == nil {
		m.codeIndex[key] = make(map[string]uuid.UUID)
	}
	for rows.Next() {
		var id uuid.UUID
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return fmt.Errorf("build code index scan: %w", err)
		}
		m.codeIndex[key][code] = id
	}
	return rows.Err()
}

// ResolveByCode looks up a UUID by string code in the code index.
func (m *Mapper) ResolveByCode(domain, sdaTable, code string) (uuid.UUID, error) {
	if m.dryRun {
		return m.deterministicUUID(domain, sdaTable, code), nil
	}
	key := m.cacheKey(domain, sdaTable)
	m.mu.RLock()
	defer m.mu.RUnlock()

	if idx, ok := m.codeIndex[key]; ok {
		if id, ok := idx[code]; ok {
			return id, nil
		}
	}
	return uuid.Nil, fmt.Errorf("code %q not found in %s index", code, sdaTable)
}

// BuildEntryDateCache loads journal entry dates from PostgreSQL for journal line migration.
func (m *Mapper) BuildEntryDateCache(ctx context.Context) error {
	rows, err := m.pool.Query(ctx,
		`SELECT id, date FROM erp_journal_entries WHERE tenant_id = $1`,
		m.tenantID,
	)
	if err != nil {
		return fmt.Errorf("build entry date cache: %w", err)
	}
	defer rows.Close()

	m.mu.Lock()
	defer m.mu.Unlock()

	for rows.Next() {
		var id uuid.UUID
		var date time.Time
		if err := rows.Scan(&id, &date); err != nil {
			return fmt.Errorf("scan entry date: %w", err)
		}
		m.dateCache[id] = date
	}
	return rows.Err()
}

// GetEntryDate returns the date for a journal entry UUID, or zero time if not found.
func (m *Mapper) GetEntryDate(entryID uuid.UUID) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dateCache[entryID]
}

// BuildRegMovimIndex builds an in-memory index: MySQL regmovim_id → SDA invoice UUID.
// Used to link FACDETAL rows to their parent invoice (IVAVENTAS/IVACOMPRAS share regmovim_id).
// Must be called after IVAVENTAS and IVACOMPRAS are migrated.
func (m *Mapper) BuildRegMovimIndex(ctx context.Context, mysqlDB *sql.DB) error {
	// Load invoice mappings from erp_legacy_mapping before looking them up.
	// In a fresh run Map() has already populated the cache; in a resume or
	// hook-replay the cache is empty and the lookup below would miss every
	// row, producing a silently empty index and 100% skip on FACDETAL.
	if err := m.PreloadDomain(ctx, "invoicing"); err != nil {
		return fmt.Errorf("preload invoicing for regmovim index: %w", err)
	}

	m.mu.Lock()
	m.regMovimIndex = make(map[int64]uuid.UUID)
	m.mu.Unlock()

	// Load from IVAVENTAS: regmovim_id → legacy_id, then legacy_id → UUID from cache
	for _, spec := range []struct {
		table  string
		pk     string
		query  string
	}{
		{"IVAVENTAS", "id_ivaventa", "SELECT id_ivaventa, regmovim_id FROM IVAVENTAS WHERE regmovim_id IS NOT NULL AND regmovim_id > 0"},
		{"IVACOMPRAS", "id_ivacompra", "SELECT id_ivacompra, regmovim_id FROM IVACOMPRAS WHERE regmovim_id IS NOT NULL AND regmovim_id > 0"},
	} {
		rows, err := mysqlDB.QueryContext(ctx, spec.query)
		if err != nil {
			return fmt.Errorf("build regmovim index from %s: %w", spec.table, err)
		}
		count := 0
		for rows.Next() {
			var legacyID, regMovimID int64
			if err := rows.Scan(&legacyID, &regMovimID); err != nil {
				_ = rows.Close()
				return fmt.Errorf("scan regmovim index: %w", err)
			}
			// Look up the invoice UUID from cache
			key := m.cacheKey("invoicing", spec.table)
			m.mu.RLock()
			invoiceUUID, ok := m.cache[key][legacyID]
			m.mu.RUnlock()
			if ok && regMovimID > 0 {
				m.mu.Lock()
				m.regMovimIndex[regMovimID] = invoiceUUID
				m.mu.Unlock()
				count++
			}
		}
		_ = rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate regmovim index %s: %w", spec.table, err)
		}
		fmt.Printf("  regmovim index: loaded %d mappings from %s\n", count, spec.table)
	}
	return nil
}

// ResolveRegMovim looks up an invoice UUID by regmovim_id.
func (m *Mapper) ResolveRegMovim(regMovimID int64) (uuid.UUID, bool) {
	if m.dryRun {
		return m.deterministicUUID("regmovim", fmt.Sprintf("%d", regMovimID)), true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.regMovimIndex[regMovimID]
	return id, ok
}

// BuildLegajoIndex builds an in-memory index: PERSONAL.legajo → entity UUID.
// Must be called after entity phase is complete so cache["entity:PERSONAL"] is populated.
// HR tables (FICHADADIA, RRHH_DESCUENTOS, RRHH_ADICIONALES, CTR_CALIDAD*) use legajo as FK,
// but PERSONAL's PK is IdPersona. Without this index those FKs silently fail.
func (m *Mapper) BuildLegajoIndex(ctx context.Context, mysqlDB *sql.DB) error {
	if m.dryRun {
		return nil
	}
	// Load entity mappings from erp_legacy_mapping before looking them up.
	// In a fresh run Map() has already populated the cache; in a resume or
	// hook-replay the cache is empty and the lookup below would miss every
	// row, silently dropping every HR migrator (FICHADADIA et al) to 0 writes.
	if err := m.PreloadDomain(ctx, "entity"); err != nil {
		return fmt.Errorf("preload entity for legajo index: %w", err)
	}
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT IdPersona, legajo FROM PERSONAL WHERE legajo IS NOT NULL AND legajo > 0`)
	if err != nil {
		return fmt.Errorf("build legajo index: %w", err)
	}
	defer func() { _ = rows.Close() }()

	key := m.cacheKey("entity", "PERSONAL")
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.legajoIndex == nil {
		m.legajoIndex = make(map[int64]uuid.UUID)
	}
	count := 0
	orphans := 0
	for rows.Next() {
		var idPersona, legajo int64
		if err := rows.Scan(&idPersona, &legajo); err != nil {
			return fmt.Errorf("scan legajo index: %w", err)
		}
		uid, ok := m.cache[key][idPersona]
		if !ok {
			orphans++
			continue
		}
		m.legajoIndex[legajo] = uid
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate legajo index: %w", err)
	}
	fmt.Printf("  legajo index: loaded %d mappings (%d PERSONAL rows without cached UUID)\n", count, orphans)
	return nil
}

// ResolveByLegajo looks up an entity UUID by PERSONAL.legajo. Returns (uuid.Nil, false) if
// legajo <= 0 or not in index. Caller should treat that as "skip row" (orphan employee FK).
func (m *Mapper) ResolveByLegajo(legajo int64) (uuid.UUID, bool) {
	if legajo <= 0 {
		return uuid.Nil, false
	}
	if m.dryRun {
		return m.deterministicUUID("legajo", fmt.Sprintf("%d", legajo)), true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.legajoIndex[legajo]
	return id, ok
}

// BuildRemitoIndex builds an in-memory index: REMITO.idRemito → REMITO UUID.
// REMITO's legacy_id in the mapper is hash("REMITO:<numero>:<puesto>"), because
// the logical PK is composite. REMDETAL.idRemito references the physical auto-inc
// column, so without this index REMDETAL loses its parent FK.
// Must be called after REMITO migrator finishes.
func (m *Mapper) BuildRemitoIndex(ctx context.Context, mysqlDB *sql.DB) error {
	if m.dryRun {
		return nil
	}
	// Same re-run safety as BuildLegajoIndex / BuildRegMovimIndex: make sure
	// the invoicing cache is loaded from erp_legacy_mapping before we look up
	// REMITO's UUIDs, so a hook replay on resume still builds a populated
	// index instead of silently emitting zero mappings.
	if err := m.PreloadDomain(ctx, "invoicing"); err != nil {
		return fmt.Errorf("preload invoicing for remito index: %w", err)
	}
	// REMITO's schema varies across saldivia snapshots: some have the
	// auto-increment idRemito column REMDETAL references, most don't. When
	// absent we log a warning and leave the index empty — REMDETAL rows with
	// orphan parents will be archived by --archive-skips instead of blocking
	// the entire run.
	var hasIDRemito int
	if err := mysqlDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'REMITO' AND column_name = 'idRemito'
	`).Scan(&hasIDRemito); err != nil {
		return fmt.Errorf("detect REMITO.idRemito: %w", err)
	}
	if hasIDRemito == 0 {
		fmt.Println("  remito index: REMITO.idRemito column not present — REMDETAL FKs will flow through archive-skips")
		return nil
	}
	rows, err := mysqlDB.QueryContext(ctx,
		`SELECT idRemito, numero, puesto FROM REMITO WHERE idRemito IS NOT NULL AND idRemito > 0`)
	if err != nil {
		return fmt.Errorf("build remito index: %w", err)
	}
	defer func() { _ = rows.Close() }()

	key := m.cacheKey("invoicing", "REMITO")
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.remitoByID == nil {
		m.remitoByID = make(map[int64]uuid.UUID)
	}
	count := 0
	orphans := 0
	for rows.Next() {
		var idRemito, numero, puesto int64
		if err := rows.Scan(&idRemito, &numero, &puesto); err != nil {
			return fmt.Errorf("scan remito index: %w", err)
		}
		compositeKey := fmt.Sprintf("REMITO:%d:%d", numero, puesto)
		hashKey := hashCode(compositeKey)
		if hashKey == 0 {
			hashKey = 1
		}
		uid, ok := m.cache[key][hashKey]
		if !ok {
			orphans++
			continue
		}
		m.remitoByID[idRemito] = uid
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate remito index: %w", err)
	}
	fmt.Printf("  remito index: loaded %d mappings (%d REMITO rows without cached UUID)\n", count, orphans)
	return nil
}

// ResolveByRemitoID looks up a REMITO UUID by its idRemito auto-inc PK.
// Returns (uuid.Nil, false) if the REMITO parent is not found (orphan REMDETAL line).
func (m *Mapper) ResolveByRemitoID(idRemito int64) (uuid.UUID, bool) {
	if idRemito <= 0 {
		return uuid.Nil, false
	}
	if m.dryRun {
		return m.deterministicUUID("remito_id", fmt.Sprintf("%d", idRemito)), true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.remitoByID[idRemito]
	return id, ok
}

// cacheCount returns the number of (legacy_id → uuid) mappings the in-memory
// cache holds for (domain, legacy_table). Used by tests to verify that
// PreloadDomain populated the bucket — there is no production use.
func (m *Mapper) cacheCount(domain, legacyTable string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cache[m.cacheKey(domain, legacyTable)])
}

// legajoIndexCount returns the size of the PERSONAL.legajo → entity UUID
// index. Used by tests to confirm BuildLegajoIndex produced non-zero mappings;
// not used in production code.
func (m *Mapper) legajoIndexCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.legajoIndex)
}

// PreloadDomain loads all existing mappings for a domain into cache (for FK resolution performance).
func (m *Mapper) PreloadDomain(ctx context.Context, domain string) error {
	rows, err := m.pool.Query(ctx,
		`SELECT legacy_table, legacy_id, sda_id FROM erp_legacy_mapping
		 WHERE tenant_id = $1 AND domain = $2`,
		m.tenantID, domain,
	)
	if err != nil {
		return fmt.Errorf("preload %s: %w", domain, err)
	}
	defer rows.Close()

	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for rows.Next() {
		var table string
		var legacyID int64
		var sdaID uuid.UUID
		if err := rows.Scan(&table, &legacyID, &sdaID); err != nil {
			return fmt.Errorf("preload scan %s: %w", domain, err)
		}
		key := m.cacheKey(domain, table)
		if m.cache[key] == nil {
			m.cache[key] = make(map[int64]uuid.UUID)
		}
		m.cache[key][legacyID] = sdaID
		count++
	}
	return rows.Err()
}
