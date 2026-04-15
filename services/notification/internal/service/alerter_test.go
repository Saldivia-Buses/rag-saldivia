//go:build integration

// Integration tests for PgAlertStore.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/ -run TestAlert
//
// NOTE: alerter.go does not exist yet — alert_store.go is a pure persistence
// layer with no threshold, dedup or aggregation logic of its own. These tests
// verify the upsert semantics of SaveAlert, which are the only non-trivial
// behavior in this layer.
package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// infraAlertSchema is the minimal DDL needed to run alert store tests.
// Mirrors the production migration without foreign keys to other tables.
const infraAlertSchema = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE infra_alerts (
	id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
	fingerprint   TEXT NOT NULL,
	status        TEXT NOT NULL CHECK (status IN ('firing', 'resolved')),
	severity      TEXT NOT NULL DEFAULT 'info',
	alertname     TEXT NOT NULL,
	service       TEXT NOT NULL DEFAULT '',
	summary       TEXT NOT NULL DEFAULT '',
	description   TEXT NOT NULL DEFAULT '',
	labels        JSONB NOT NULL DEFAULT '{}',
	annotations   JSONB NOT NULL DEFAULT '{}',
	starts_at     TIMESTAMPTZ NOT NULL,
	ends_at       TIMESTAMPTZ,
	received_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (fingerprint, starts_at)
);
`

// setupAlertTestDB spins up an isolated PostgreSQL container and applies the
// infra_alerts schema. Returns the pool and a cleanup function.
func setupAlertTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("sda_alert_test"),
		postgres.WithUsername("sda"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "start postgres container")
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "get connection string")

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "create pool")
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, infraAlertSchema)
	require.NoError(t, err, "apply infra_alerts schema")

	return pool
}

// sampleAlert returns a fully populated InfraAlert for a given fingerprint and status.
func sampleAlert(fingerprint, status string, startsAt time.Time) InfraAlert {
	return InfraAlert{
		Fingerprint: fingerprint,
		Status:      status,
		Severity:    "warning",
		AlertName:   "HighMemoryUsage",
		Service:     "auth",
		Summary:     "Memory above 80%",
		Description: "The auth service is using more than 80% memory",
		Labels:      json.RawMessage(`{"env":"prod","service":"auth"}`),
		Annotations: json.RawMessage(`{"runbook":"https://wiki/runbook/memory"}`),
		StartsAt:    startsAt,
		EndsAt:      nil,
	}
}

// TestAlertStore_SaveAlert_Persists verifies that a new alert is stored with all fields.
func TestAlertStore_SaveAlert_Persists(t *testing.T) {
	pool := setupAlertTestDB(t)
	store := NewPgAlertStore(pool)
	ctx := t.Context()

	startsAt := time.Now().Truncate(time.Microsecond).UTC()
	alert := sampleAlert("fp-001", "firing", startsAt)

	err := store.SaveAlert(ctx, alert)
	require.NoError(t, err)

	// Verify the row was persisted
	var (
		gotFingerprint string
		gotStatus      string
		gotSeverity    string
		gotAlertName   string
		gotService     string
	)
	row := pool.QueryRow(ctx,
		`SELECT fingerprint, status, severity, alertname, service FROM infra_alerts WHERE fingerprint = $1`,
		"fp-001",
	)
	err = row.Scan(&gotFingerprint, &gotStatus, &gotSeverity, &gotAlertName, &gotService)
	require.NoError(t, err, "alert must exist after SaveAlert")

	require.Equal(t, "fp-001", gotFingerprint)
	require.Equal(t, "firing", gotStatus)
	require.Equal(t, "warning", gotSeverity)
	require.Equal(t, "HighMemoryUsage", gotAlertName)
	require.Equal(t, "auth", gotService)
}

// TestAlertStore_SaveAlert_Upsert_UpdatesStatus verifies the ON CONFLICT upsert:
// saving the same (fingerprint, starts_at) pair twice with different status
// should update the existing row, not insert a duplicate.
func TestAlertStore_SaveAlert_Upsert_UpdatesStatus(t *testing.T) {
	pool := setupAlertTestDB(t)
	store := NewPgAlertStore(pool)
	ctx := t.Context()

	startsAt := time.Now().Truncate(time.Microsecond).UTC()

	// First save: status=firing
	err := store.SaveAlert(ctx, sampleAlert("fp-upsert", "firing", startsAt))
	require.NoError(t, err, "initial firing save")

	// Second save: same fingerprint + starts_at, but now resolved
	endsAt := startsAt.Add(5 * time.Minute)
	resolved := sampleAlert("fp-upsert", "resolved", startsAt)
	resolved.EndsAt = &endsAt

	err = store.SaveAlert(ctx, resolved)
	require.NoError(t, err, "upsert to resolved")

	// Must still be exactly one row
	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM infra_alerts WHERE fingerprint = $1`,
		"fp-upsert",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "upsert must not create a second row")

	// Status must be updated to resolved
	var gotStatus string
	err = pool.QueryRow(ctx,
		`SELECT status FROM infra_alerts WHERE fingerprint = $1`,
		"fp-upsert",
	).Scan(&gotStatus)
	require.NoError(t, err)
	require.Equal(t, "resolved", gotStatus, "upsert must update status to resolved")
}

// TestAlertStore_SaveAlert_DifferentStartsAt_InsertsSeparateRow verifies that
// the same fingerprint with a different starts_at is treated as a new alert
// (not a conflict), since re-fires are distinct events.
func TestAlertStore_SaveAlert_DifferentStartsAt_InsertsSeparateRow(t *testing.T) {
	pool := setupAlertTestDB(t)
	store := NewPgAlertStore(pool)
	ctx := t.Context()

	startsAt1 := time.Now().Truncate(time.Microsecond).UTC()
	startsAt2 := startsAt1.Add(1 * time.Hour) // same alert, re-fired 1h later

	err := store.SaveAlert(ctx, sampleAlert("fp-refire", "firing", startsAt1))
	require.NoError(t, err)

	err = store.SaveAlert(ctx, sampleAlert("fp-refire", "firing", startsAt2))
	require.NoError(t, err)

	// Both rows must exist (different starts_at = different incident)
	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM infra_alerts WHERE fingerprint = $1`,
		"fp-refire",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count, "different starts_at must create separate alert rows")
}

// TestAlertStore_SaveAlert_MultipleTenants verifies that alerts from different
// services are stored independently and do not interfere with each other.
func TestAlertStore_SaveAlert_MultipleAlerts(t *testing.T) {
	pool := setupAlertTestDB(t)
	store := NewPgAlertStore(pool)
	ctx := t.Context()

	startsAt := time.Now().Truncate(time.Microsecond).UTC()
	alerts := []InfraAlert{
		{
			Fingerprint: "fp-auth-mem",
			Status:      "firing",
			Severity:    "critical",
			AlertName:   "HighMemory",
			Service:     "auth",
			Labels:      json.RawMessage(`{}`),
			Annotations: json.RawMessage(`{}`),
			StartsAt:    startsAt,
		},
		{
			Fingerprint: "fp-chat-cpu",
			Status:      "firing",
			Severity:    "warning",
			AlertName:   "HighCPU",
			Service:     "chat",
			Labels:      json.RawMessage(`{}`),
			Annotations: json.RawMessage(`{}`),
			StartsAt:    startsAt,
		},
		{
			Fingerprint: "fp-ingest-disk",
			Status:      "firing",
			Severity:    "info",
			AlertName:   "DiskUsage",
			Service:     "ingest",
			Labels:      json.RawMessage(`{}`),
			Annotations: json.RawMessage(`{}`),
			StartsAt:    startsAt,
		},
	}

	for _, a := range alerts {
		require.NoError(t, store.SaveAlert(ctx, a))
	}

	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM infra_alerts`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 3, count, "all three alerts must be persisted")
}
