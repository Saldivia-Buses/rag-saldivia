// Package database provides pgxpool helpers for SDA services.
//
// Currently a thin wrapper. When otelpgx is added to go.mod, NewPool
// will automatically add OpenTelemetry query tracing to all connections.
//
// To enable tracing (requires internet for go get):
//   1. go get github.com/exaring/otelpgx
//   2. Uncomment the tracer line in NewPool below
//   3. All SQL queries appear as spans in Tempo
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pgxpool connection pool.
// When otelpgx is available, uncomment the tracer to enable query tracing.
func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// TODO(plan14): Enable DB query tracing when otelpgx is added to go.mod:
	//   import "github.com/exaring/otelpgx"
	//   cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	//
	// This makes every SQL query visible as a span in Tempo traces:
	//   [HTTP POST /v1/chat 450ms]
	//     └── [pgx.query GetMessages 3ms]
	//     └── [pgx.query CreateMessage 2ms]

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	return pool, nil
}

// SetTenantID sets the app.tenant_id session variable on a connection.
// This enables PostgreSQL Row-Level Security (RLS) policies that filter
// by current_setting('app.tenant_id'). Call this at the start of each
// request-scoped transaction.
//
// Usage:
//
//	tx, _ := pool.Begin(ctx)
//	database.SetTenantID(ctx, tx, tenantSlug)
//	// all queries in tx are now RLS-filtered
func SetTenantID(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	return err
}
