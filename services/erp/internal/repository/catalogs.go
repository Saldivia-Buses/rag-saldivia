package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Catalog represents a row in erp_catalogs.
type Catalog struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  string          `json:"tenant_id"`
	Type      string          `json:"type"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	ParentID  *uuid.UUID      `json:"parent_id,omitempty"`
	SortOrder int             `json:"sort_order"`
	Active    bool            `json:"active"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ListCatalogs returns catalogs filtered by type.
func (q *Queries) ListCatalogs(ctx context.Context, tenantID, catalogType string, activeOnly bool) ([]Catalog, error) {
	query := `SELECT id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at
		 FROM erp_catalogs
		 WHERE tenant_id = $1 AND type = $2`
	args := []any{tenantID, catalogType}

	if activeOnly {
		query += ` AND active = true`
	}
	query += ` ORDER BY sort_order, name`

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list catalogs: %w", err)
	}
	defer rows.Close()

	return scanCatalogs(rows)
}

// ListCatalogTypes returns distinct catalog types for a tenant.
func (q *Queries) ListCatalogTypes(ctx context.Context, tenantID string) ([]string, error) {
	rows, err := q.db.Query(ctx,
		`SELECT DISTINCT type FROM erp_catalogs WHERE tenant_id = $1 ORDER BY type`,
		tenantID)
	if err != nil {
		return nil, fmt.Errorf("list catalog types: %w", err)
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	if types == nil {
		types = []string{}
	}
	return types, rows.Err()
}

// GetCatalog returns a single catalog entry by ID.
func (q *Queries) GetCatalog(ctx context.Context, id uuid.UUID, tenantID string) (*Catalog, error) {
	var c Catalog
	err := q.db.QueryRow(ctx,
		`SELECT id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at
		 FROM erp_catalogs
		 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID).Scan(&c.ID, &c.TenantID, &c.Type, &c.Code, &c.Name,
		&c.ParentID, &c.SortOrder, &c.Active, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get catalog: %w", err)
	}
	return &c, nil
}

// CreateCatalog inserts a new catalog entry.
func (q *Queries) CreateCatalog(ctx context.Context, tenantID, catalogType, code, name string, parentID *uuid.UUID, sortOrder int, metadata json.RawMessage) (*Catalog, error) {
	if metadata == nil {
		metadata = json.RawMessage(`{}`)
	}
	var c Catalog
	err := q.db.QueryRow(ctx,
		`INSERT INTO erp_catalogs (tenant_id, type, code, name, parent_id, sort_order, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at`,
		tenantID, catalogType, code, name, parentID, sortOrder, metadata).Scan(
		&c.ID, &c.TenantID, &c.Type, &c.Code, &c.Name,
		&c.ParentID, &c.SortOrder, &c.Active, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create catalog: %w", err)
	}
	return &c, nil
}

// UpdateCatalog updates an existing catalog entry.
func (q *Queries) UpdateCatalog(ctx context.Context, id uuid.UUID, tenantID, code, name string, parentID *uuid.UUID, sortOrder int, active bool, metadata json.RawMessage) (*Catalog, error) {
	if metadata == nil {
		metadata = json.RawMessage(`{}`)
	}
	var c Catalog
	err := q.db.QueryRow(ctx,
		`UPDATE erp_catalogs
		 SET code = $3, name = $4, parent_id = $5, sort_order = $6, active = $7, metadata = $8, updated_at = now()
		 WHERE id = $1 AND tenant_id = $2
		 RETURNING id, tenant_id, type, code, name, parent_id, sort_order, active, metadata, created_at, updated_at`,
		id, tenantID, code, name, parentID, sortOrder, active, metadata).Scan(
		&c.ID, &c.TenantID, &c.Type, &c.Code, &c.Name,
		&c.ParentID, &c.SortOrder, &c.Active, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update catalog: %w", err)
	}
	return &c, nil
}

// DeleteCatalog soft-deletes a catalog entry (sets active=false).
func (q *Queries) DeleteCatalog(ctx context.Context, id uuid.UUID, tenantID string) error {
	tag, err := q.db.Exec(ctx,
		`UPDATE erp_catalogs SET active = false, updated_at = now()
		 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID)
	if err != nil {
		return fmt.Errorf("delete catalog: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("catalog not found")
	}
	return nil
}

func scanCatalogs(rows pgx.Rows) ([]Catalog, error) {
	var catalogs []Catalog
	for rows.Next() {
		var c Catalog
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Type, &c.Code, &c.Name,
			&c.ParentID, &c.SortOrder, &c.Active, &c.Metadata, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		catalogs = append(catalogs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if catalogs == nil {
		catalogs = []Catalog{}
	}
	return catalogs, nil
}
