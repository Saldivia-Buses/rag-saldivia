package legacy

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a read-only MySQL connection pool to the Histrix legacy database.
// DSN must include charset=utf8mb4&parseTime=true for proper encoding/date handling.
// The session is set to READ ONLY to prevent accidental writes.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	// Enforce read-only session to prevent accidental writes to legacy DB
	if _, err := db.Exec("SET SESSION TRANSACTION READ ONLY"); err != nil {
		return nil, fmt.Errorf("set read-only: %w", err)
	}
	return db, nil
}
