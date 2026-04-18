package legacy

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Connect opens a MySQL connection pool to the Histrix legacy database.
// The DSN is augmented with safe defaults (timeouts, utf8mb4, parseTime) and
// every connection in the pool is initialised with transaction_read_only=ON so
// accidental writes are impossible even through a non-transactional query.
func Connect(dsn string) (*sql.DB, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse mysql dsn: %w", err)
	}

	// Safe defaults — do not override user-provided values.
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Minute // big tables need time
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 1 * time.Minute
	}
	if !cfg.ParseTime {
		cfg.ParseTime = true
	}
	// Use handshake-time collation field — not a DSN param. Modern go-sql-driver
	// treats unknown params (including "charset") as SET session.<name>=<val>,
	// and MySQL rejects `SET charset=...` with error 1193. Histrix (MySQL 5.7)
	// uses utf8, so default to utf8_general_ci when the caller did not choose.
	if cfg.Collation == "" {
		cfg.Collation = "utf8_general_ci"
	}
	// Session-level read-only on every new connection of the pool.
	cfg.Params = ensureParam(cfg.Params, "tx_read_only", "1")
	// Do not reuse autocommit-related settings from user DSN aggressively:
	// leave autocommit alone; tx_read_only covers writes inside and outside tx.

	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, fmt.Errorf("new mysql connector: %w", err)
	}

	// Wrap connector so each new connection is validated as read-only before use.
	db := sql.OpenDB(readOnlyConnector{Connector: connector})
	// Bulk import benefits from more MySQL connections when the pipeline
	// has multiple readers (currently just one per migrator, but the pool
	// also absorbs burst patterns from rescue scans).
	db.SetMaxOpenConns(32)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

func ensureParam(params map[string]string, key, value string) map[string]string {
	if params == nil {
		params = make(map[string]string)
	}
	if _, ok := params[key]; !ok {
		params[key] = value
	}
	return params
}

// readOnlyConnector asserts on every new driver connection that the session is
// running in read-only mode. Fails closed if the server silently ignored the
// tx_read_only param (permissions issue, proxy, etc.).
type readOnlyConnector struct{ driver.Connector }

func (c readOnlyConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	q, ok := conn.(driver.QueryerContext)
	if !ok {
		return conn, nil
	}
	rows, err := q.QueryContext(ctx, "SELECT @@session.transaction_read_only", nil)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("verify read-only: %w", err)
	}
	defer func() { _ = rows.Close() }()
	vals := make([]driver.Value, 1)
	if err := rows.Next(vals); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read read-only flag: %w", err)
	}
	if fmt.Sprintf("%v", vals[0]) != "1" {
		_ = conn.Close()
		return nil, fmt.Errorf("legacy session is not read-only (tx_read_only=%v) — refuse to connect", vals[0])
	}
	return conn, nil
}
