// Package audit provides a shared audit log writer for all SDA services.
// Every service that has access to a tenant DB can write audit entries.
// The audit_log table is created by auth migration 001.
//
// Writes are non-failing: errors are logged but not returned, so audit
// failures never break business logic.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Writer writes audit entries to the tenant's audit_log table.
type Writer struct {
	db *pgxpool.Pool
}

// NewWriter creates an audit writer for the given tenant database.
func NewWriter(db *pgxpool.Pool) *Writer {
	return &Writer{db: db}
}

// Entry represents one audit log row.
type Entry struct {
	UserID    string         // who did it (empty for system events)
	Action    string         // "user.login", "chat.session.create", etc.
	Resource  string         // what was affected (session ID, email, etc.)
	Details   map[string]any // action-specific data (stored as JSONB)
	IP        string         // client IP address
	UserAgent string         // client user-agent string
}

// Write inserts an audit entry. Non-failing — errors are logged, not returned.
// This ensures audit failures never break the business operation that triggered them.
func (w *Writer) Write(ctx context.Context, e Entry) {
	details, err := json.Marshal(e.Details)
	if err != nil {
		details = []byte("{}")
	}

	_, err = w.db.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, resource, details, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		nilIfEmpty(e.UserID), e.Action, nilIfEmpty(e.Resource), details, nilIfEmpty(e.IP), nilIfEmpty(e.UserAgent),
	)
	if err != nil {
		slog.Error("audit write failed", "error", err, "action", e.Action, "user_id", e.UserID)
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
