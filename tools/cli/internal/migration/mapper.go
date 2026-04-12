package migration

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Mapper handles legacy INT → SDA UUID mapping with in-memory cache.
type Mapper struct {
	pool     *pgxpool.Pool
	tenantID string
	mu       sync.RWMutex
	cache    map[string]map[int64]uuid.UUID // "domain:table" → legacy_id → uuid
}

// NewMapper creates a new ID mapper for a tenant.
func NewMapper(pool *pgxpool.Pool, tenantID string) *Mapper {
	return &Mapper{
		pool:     pool,
		tenantID: tenantID,
		cache:    make(map[string]map[int64]uuid.UUID),
	}
}

func (m *Mapper) cacheKey(domain, table string) string {
	return domain + ":" + table
}

// Map returns the UUID for a legacy ID. If no mapping exists, generates a new one.
// Must be called within a transaction.
func (m *Mapper) Map(ctx context.Context, tx pgx.Tx, domain, table string, legacyID int64, legacyCreatedBy *string) (uuid.UUID, error) {
	key := m.cacheKey(domain, table)

	m.mu.RLock()
	if cached, ok := m.cache[key][legacyID]; ok {
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	// Check DB
	var sdaID uuid.UUID
	err := tx.QueryRow(ctx,
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
	_, err = tx.Exec(ctx,
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
