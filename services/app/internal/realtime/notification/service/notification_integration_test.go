//go:build integration

// Integration tests for the notification service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/

package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("sda_test"),
		postgres.WithUsername("sda"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	// Apply base users schema + notification schema
	migration := `
		CREATE TABLE users (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		INSERT INTO users (id, email, name) VALUES ('u-1', 'alice@test.com', 'Alice');
		INSERT INTO users (id, email, name) VALUES ('u-2', 'bob@test.com', 'Bob');

		CREATE TABLE notifications (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			body TEXT NOT NULL DEFAULT '',
			data JSONB NOT NULL DEFAULT '{}',
			channel TEXT NOT NULL DEFAULT 'in_app',
			is_read BOOLEAN NOT NULL DEFAULT false,
			read_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX idx_notifications_user ON notifications (user_id, created_at DESC);

		CREATE TABLE notification_preferences (
			user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			email_enabled BOOLEAN NOT NULL DEFAULT true,
			in_app_enabled BOOLEAN NOT NULL DEFAULT true,
			quiet_start TIME,
			quiet_end TIME,
			muted_types TEXT[] NOT NULL DEFAULT '{}',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`
	if _, err := pool.Exec(ctx, migration); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
	return pool, cleanup
}

func TestCreate_and_List_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	notif, err := svc.Create(ctx, "u-1", "chat.new_message", "Nuevo mensaje", "Hola mundo", nil, "in_app")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if notif.ID == "" {
		t.Error("expected non-empty ID")
	}
	if notif.IsRead {
		t.Error("new notification should be unread")
	}

	list, err := svc.List(ctx, "u-1", false, 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 notification, got %d", len(list))
	}
}

func TestUnreadCount_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	svc.Create(ctx, "u-1", "test", "A", "", nil, "in_app")
	svc.Create(ctx, "u-1", "test", "B", "", nil, "in_app")
	svc.Create(ctx, "u-2", "test", "C", "", nil, "in_app") // different user

	count, err := svc.UnreadCount(ctx, "u-1")
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 unread for u-1, got %d", count)
	}
}

func TestMarkRead_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	notif, _ := svc.Create(ctx, "u-1", "test", "Read me", "", nil, "in_app")

	if err := svc.MarkRead(ctx, notif.ID, "u-1"); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	count, _ := svc.UnreadCount(ctx, "u-1")
	if count != 0 {
		t.Errorf("expected 0 unread after mark read, got %d", count)
	}
}

func TestMarkRead_WrongUser_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	notif, _ := svc.Create(ctx, "u-1", "test", "Private", "", nil, "in_app")

	err := svc.MarkRead(ctx, notif.ID, "u-2")
	if err == nil {
		t.Fatal("expected error marking someone else's notification as read")
	}
}

func TestMarkAllRead_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	svc.Create(ctx, "u-1", "test", "A", "", nil, "in_app")
	svc.Create(ctx, "u-1", "test", "B", "", nil, "in_app")
	svc.Create(ctx, "u-1", "test", "C", "", nil, "in_app")

	marked, err := svc.MarkAllRead(ctx, "u-1")
	if err != nil {
		t.Fatalf("mark all read: %v", err)
	}
	if marked != 3 {
		t.Errorf("expected 3 marked, got %d", marked)
	}

	count, _ := svc.UnreadCount(ctx, "u-1")
	if count != 0 {
		t.Errorf("expected 0 unread, got %d", count)
	}
}

func TestPreferences_DefaultsAndUpdate_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	// Defaults
	prefs, err := svc.GetPreferences(ctx, "u-1")
	if err != nil {
		t.Fatalf("get preferences: %v", err)
	}
	if !prefs.EmailEnabled {
		t.Error("expected email_enabled true by default")
	}

	// Update
	qs := "22:00"
	qe := "08:00"
	updated, err := svc.UpdatePreferences(ctx, "u-1", false, true, &qs, &qe, []string{"chat.new_message"})
	if err != nil {
		t.Fatalf("update preferences: %v", err)
	}
	if updated.EmailEnabled {
		t.Error("expected email_enabled false after update")
	}
	if len(updated.MutedTypes) != 1 || updated.MutedTypes[0] != "chat.new_message" {
		t.Errorf("expected muted_types [chat.new_message], got %v", updated.MutedTypes)
	}

	// Re-read
	prefs2, _ := svc.GetPreferences(ctx, "u-1")
	if prefs2.EmailEnabled {
		t.Error("expected email_enabled false on re-read")
	}
}

func TestList_UnreadFilter_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := New(pool)
	ctx := context.Background()

	n1, _ := svc.Create(ctx, "u-1", "test", "Unread", "", nil, "in_app")
	svc.Create(ctx, "u-1", "test", "Also unread", "", nil, "in_app")
	svc.MarkRead(ctx, n1.ID, "u-1")

	// All
	all, _ := svc.List(ctx, "u-1", false, 50)
	if len(all) != 2 {
		t.Errorf("expected 2 total, got %d", len(all))
	}

	// Unread only
	unread, _ := svc.List(ctx, "u-1", true, 50)
	if len(unread) != 1 {
		t.Errorf("expected 1 unread, got %d", len(unread))
	}
}
