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
	pool          *pgxpool.Pool
	tenantID      string
	dryRun        bool
	mu            sync.RWMutex
	cache         map[string]map[int64]uuid.UUID   // "domain:table" → legacy_id → uuid
	codeIndex     map[string]map[string]uuid.UUID  // "domain:table" → code → uuid (for varchar PK tables)
	dateCache     map[uuid.UUID]time.Time          // entry UUID → date (for journal line date resolution)
	regMovimIndex map[int64]uuid.UUID              // regmovim_id → invoice UUID (for FACDETAL FK resolution)
	pending       []pendingMapping                  // batch of mappings to flush
}

// NewMapper creates a new ID mapper for a tenant.
func NewMapper(pool *pgxpool.Pool, tenantID string) *Mapper {
	return &Mapper{
		pool:      pool,
		tenantID:  tenantID,
		cache:     make(map[string]map[int64]uuid.UUID),
		codeIndex: make(map[string]map[string]uuid.UUID),
		dateCache: make(map[uuid.UUID]time.Time),
	}
}

// SetDryRun enables dry-run mode: all lookups return deterministic UUIDs without hitting PG.
func (m *Mapper) SetDryRun(v bool) { m.dryRun = v }

func (m *Mapper) cacheKey(domain, table string) string {
	return domain + ":" + table
}

// querier abstracts pgx.Tx and pgxpool.Pool for query execution.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Map returns the UUID for a legacy ID. If no mapping exists, generates a new one.
// When tx is non-nil, uses the transaction; otherwise falls back to the pool.
func (m *Mapper) Map(ctx context.Context, tx pgx.Tx, domain, table string, legacyID int64, legacyCreatedBy *string) (uuid.UUID, error) {
	if m.dryRun {
		return uuid.New(), nil
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
func (m *Mapper) FlushPending(ctx context.Context, q querier) error {
	m.mu.Lock()
	batch := m.pending
	m.pending = nil
	m.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id) VALUES ")
	args := make([]any, 0, len(batch)*5)
	for i, p := range batch {
		if i > 0 {
			sb.WriteString(",")
		}
		n := i * 5
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5))
		args = append(args, m.tenantID, p.domain, p.table, p.legacyID, p.sdaID)
	}
	sb.WriteString(" ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING")

	_, err := q.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("flush %d mappings: %w", len(batch), err)
	}
	return nil
}

// Resolve looks up an existing mapping (for foreign keys). Returns error if not found.
func (m *Mapper) Resolve(ctx context.Context, domain, table string, legacyID int64) (uuid.UUID, error) {
	if legacyID == 0 {
		return uuid.Nil, fmt.Errorf("resolve %s/%s: legacy_id is 0", domain, table)
	}

	if m.dryRun {
		return uuid.New(), nil
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
		return uuid.New(), nil
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
				rows.Close()
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
		rows.Close()
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
		return uuid.New(), true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.regMovimIndex[regMovimID]
	return id, ok
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
