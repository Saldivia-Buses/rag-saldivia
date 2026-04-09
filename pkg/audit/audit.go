// Package audit provides a shared audit log writer for all SDA services.
// Every service that has access to a tenant DB can write audit entries.
// The audit_log table is created by auth migration 001.
//
// Two interfaces:
//   - Logger: non-failing writes (errors logged, not returned). Use for operations
//     where audit failure should not break business logic.
//   - StrictLogger: fail-closed writes (errors returned). Use for operations that
//     MUST be audited before proceeding (PLC writes, remote exec, credential access).
package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Logger is the interface for non-failing audit logging. Services should depend
// on this interface, not on *Writer directly, to enable mocking in tests.
type Logger interface {
	Write(ctx context.Context, e Entry)
}

// StrictLogger is the interface for fail-closed audit logging.
// Use for operations that MUST be audited (PLC writes, remote exec).
// If the audit write fails, the caller MUST abort the operation.
type StrictLogger interface {
	WriteStrict(ctx context.Context, e Entry) error
}

// Writer writes audit entries to the tenant's audit_log table.
type Writer struct {
	db *pgxpool.Pool
}

// Ensure Writer implements both interfaces at compile time.
var _ Logger = (*Writer)(nil)
var _ StrictLogger = (*Writer)(nil)

// NewWriter creates an audit writer for the given tenant database.
func NewWriter(db *pgxpool.Pool) *Writer {
	return &Writer{db: db}
}

// Entry represents one audit log row.
type Entry struct {
	TenantID  string         // tenant context (empty for backward compat)
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
		`INSERT INTO audit_log (tenant_id, user_id, action, resource, details, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nilIfEmpty(e.TenantID), nilIfEmpty(e.UserID), e.Action, nilIfEmpty(e.Resource),
		details, nilIfEmpty(e.IP), nilIfEmpty(e.UserAgent),
	)
	if err != nil {
		slog.Error("audit write failed", "error", err, "action", e.Action, "user_id", e.UserID)
	}
}

// WriteStrict inserts an audit entry and returns the error.
// If audit fails, the caller MUST abort the operation.
func (w *Writer) WriteStrict(ctx context.Context, e Entry) error {
	details, err := json.Marshal(e.Details)
	if err != nil {
		details = []byte("{}")
	}

	_, err = w.db.Exec(ctx,
		`INSERT INTO audit_log (tenant_id, user_id, action, resource, details, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nilIfEmpty(e.TenantID), nilIfEmpty(e.UserID), e.Action, nilIfEmpty(e.Resource),
		details, nilIfEmpty(e.IP), nilIfEmpty(e.UserAgent),
	)
	if err != nil {
		slog.Error("STRICT audit write failed — operation will be aborted",
			"error", err, "action", e.Action, "user_id", e.UserID)
	}
	return err
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
