//go:build integration

// Integration tests for the chat service.
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

	// Apply auth base schema (users table needed for FK) + chat schema
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

		CREATE TABLE sessions (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT NOT NULL DEFAULT 'Nueva conversacion',
			collection TEXT,
			is_saved BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX idx_sessions_user ON sessions (user_id, updated_at DESC);

		CREATE TABLE messages (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
			content TEXT NOT NULL,
			sources JSONB,
			metadata JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE INDEX idx_messages_session ON messages (session_id, created_at ASC);
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

func TestCreateSession_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	session, err := svc.CreateSession(ctx, "u-1", "Mi primer chat", nil)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}
	if session.Title != "Mi primer chat" {
		t.Errorf("expected title 'Mi primer chat', got %q", session.Title)
	}
	if session.UserID != "u-1" {
		t.Errorf("expected user_id u-1, got %q", session.UserID)
	}
}

func TestGetSession_Ownership_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, "u-1", "Private chat", nil)

	// Owner can access
	got, err := svc.GetSession(ctx, session.ID, "u-1")
	if err != nil {
		t.Fatalf("owner should access: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("expected session %s, got %s", session.ID, got.ID)
	}

	// Non-owner gets ErrNotOwner
	_, err = svc.GetSession(ctx, session.ID, "u-2")
	if err != ErrNotOwner {
		t.Fatalf("expected ErrNotOwner, got: %v", err)
	}
}

func TestListSessions_FiltersbyUser_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	svc.CreateSession(ctx, "u-1", "Alice chat 1", nil)
	svc.CreateSession(ctx, "u-1", "Alice chat 2", nil)
	svc.CreateSession(ctx, "u-2", "Bob chat", nil)

	sessions, err := svc.ListSessions(ctx, "u-1")
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions for u-1, got %d", len(sessions))
	}
}

func TestDeleteSession_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, "u-1", "To delete", nil)

	if err := svc.DeleteSession(ctx, session.ID, "u-1"); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	_, err := svc.GetSession(ctx, session.ID, "u-1")
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound after delete, got: %v", err)
	}
}

func TestAddMessage_and_GetMessages_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, "u-1", "Chat", nil)

	msg, err := svc.AddMessage(ctx, session.ID, "u-1", "user", "Hola", nil, nil)
	if err != nil {
		t.Fatalf("add message: %v", err)
	}
	if msg.Content != "Hola" {
		t.Errorf("expected content 'Hola', got %q", msg.Content)
	}

	svc.AddMessage(ctx, session.ID, "u-1", "assistant", "Hola! En que puedo ayudarte?", nil, nil)

	messages, err := svc.GetMessages(ctx, session.ID)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
	// Messages should be ordered by created_at ASC
	if messages[0].Role != "user" {
		t.Errorf("first message should be user, got %q", messages[0].Role)
	}
	if messages[1].Role != "assistant" {
		t.Errorf("second message should be assistant, got %q", messages[1].Role)
	}
}

func TestRenameSession_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, "u-1", "Original", nil)

	if err := svc.RenameSession(ctx, session.ID, "u-1", "Renamed"); err != nil {
		t.Fatalf("rename: %v", err)
	}

	got, _ := svc.GetSession(ctx, session.ID, "u-1")
	if got.Title != "Renamed" {
		t.Errorf("expected title 'Renamed', got %q", got.Title)
	}
}

func TestDeleteSession_CascadesMessages_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	session, _ := svc.CreateSession(ctx, "u-1", "Chat", nil)
	svc.AddMessage(ctx, session.ID, "u-1", "user", "msg1", nil, nil)
	svc.AddMessage(ctx, session.ID, "u-1", "assistant", "msg2", nil, nil)

	svc.DeleteSession(ctx, session.ID, "u-1")

	// Messages should be cascade deleted
	var count int
	pool.QueryRow(ctx, `SELECT count(*) FROM messages WHERE session_id = $1`, session.ID).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", count)
	}
}
