package migration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Mapper handles legacy INT → SDA UUID mapping with in-memory cache.
type Mapper struct {
	pool      *pgxpool.Pool
	tenantID  string
	dryRun    bool
	mu        sync.RWMutex
	cache     map[string]map[int64]uuid.UUID   // "domain:table" → legacy_id → uuid
	codeIndex map[string]map[string]uuid.UUID  // "domain:table" → code → uuid (for varchar PK tables)
	dateCache map[uuid.UUID]time.Time          // entry UUID → date (for journal line date resolution)
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

	// Use tx if available, otherwise pool
	var q querier = m.pool
	if tx != nil {
		q = tx
	}

	// Check DB
	var sdaID uuid.UUID
	err := q.QueryRow(ctx,
		`SELECT sda_id FROM erp_legacy_mapping
		 WHERE tenant_id = $1 AND domain = $2 AND legacy_table = $3 AND legacy_id = $4`,
		m.tenantID, domain, table, legacyID,
	).Scan(&sdaID)
	if err == nil {
		m.mu.Lock()
		if m.cache[key] == nil {
			m.cache[key] = make(map[int64]uuid.UUID)
		}
		m.cache[key][legacyID] = sdaID
		m.mu.Unlock()
		return sdaID, nil
	}

	// Generate new
	newID := uuid.New()
	_, err = q.Exec(ctx,
		`INSERT INTO erp_legacy_mapping (tenant_id, domain, legacy_table, legacy_id, sda_id, legacy_created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (tenant_id, domain, legacy_table, legacy_id) DO NOTHING`,
		m.tenantID, domain, table, legacyID, newID, legacyCreatedBy,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create mapping %s/%s/%d: %w", domain, table, legacyID, err)
	}

	m.mu.Lock()
	if m.cache[key] == nil {
		m.cache[key] = make(map[int64]uuid.UUID)
	}
	m.cache[key][legacyID] = newID
	m.mu.Unlock()

	return newID, nil
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
