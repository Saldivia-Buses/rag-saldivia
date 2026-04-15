//go:build integration

// Integration tests for the chat service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/

package service

import (
	"context"
	"fmt"
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
			thinking TEXT,
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

	// Non-owner gets ErrSessionNotFound (SQL-level filtering, not post-fetch check)
	_, err = svc.GetSession(ctx, session.ID, "u-2")
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
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

	sessions, err := svc.ListSessions(ctx, "u-1", 50, 0)
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

	msg, err := svc.AddMessage(ctx, session.ID, "u-1", "user", "Hola", nil, nil, nil)
	if err != nil {
		t.Fatalf("add message: %v", err)
	}
	if msg.Content != "Hola" {
		t.Errorf("expected content 'Hola', got %q", msg.Content)
	}

	svc.AddMessage(ctx, session.ID, "u-1", "assistant", "Hola! En que puedo ayudarte?", nil, nil, nil)

	messages, err := svc.GetMessages(ctx, session.ID, 100)
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
	svc.AddMessage(ctx, session.ID, "u-1", "user", "msg1", nil, nil, nil)
	svc.AddMessage(ctx, session.ID, "u-1", "assistant", "msg2", nil, nil, nil)

	svc.DeleteSession(ctx, session.ID, "u-1")

	// Messages should be cascade deleted
	var count int
	pool.QueryRow(ctx, `SELECT count(*) FROM messages WHERE session_id = $1`, session.ID).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 messages after cascade delete, got %d", count)
	}
}

// TestListSessions_Pagination_Integration seeds 5 sessions with distinct timestamps
// and verifies that limit + offset slicing returns the expected windows in
// created_at DESC order (most recent first).
//
// TDD-ANCHOR: ListSessions uses created_at DESC (see idx_sessions_user index).
// Sessions are inserted directly via SQL with explicit timestamps to guarantee
// deterministic ordering — CreateSession uses DEFAULT now() which can collide
// within the same millisecond in a fast test loop.
func TestListSessions_Pagination_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	svc := NewChat(pool, "dev", nil)
	ctx := context.Background()

	// Insert 5 sessions with explicit, evenly-spaced timestamps so ordering is
	// deterministic regardless of test execution speed.
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	var sessionIDs []string
	for i := 0; i < 5; i++ {
		ts := base.Add(time.Duration(i) * time.Minute)
		var id string
		err := pool.QueryRow(ctx,
			`INSERT INTO sessions (user_id, title, created_at, updated_at)
			 VALUES ($1, $2, $3, $3) RETURNING id`,
			"u-1",
			fmt.Sprintf("Session %d", i+1),
			ts,
		).Scan(&id)
		if err != nil {
			t.Fatalf("seed session %d: %v", i, err)
		}
		sessionIDs = append(sessionIDs, id)
	}
	// sessionIDs[4] has the latest timestamp → should appear first in DESC order.

	// Page 1: limit=2, offset=0 → 2 most recent sessions.
	page1, err := svc.ListSessions(ctx, "u-1", 2, 0)
	if err != nil {
		t.Fatalf("list sessions page 1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page 1: expected 2 sessions, got %d", len(page1))
	}
	// Most recent first → session index 4, then 3.
	if page1[0].ID != sessionIDs[4] {
		t.Errorf("page1[0]: expected session %s (newest), got %s", sessionIDs[4], page1[0].ID)
	}
	if page1[1].ID != sessionIDs[3] {
		t.Errorf("page1[1]: expected session %s, got %s", sessionIDs[3], page1[1].ID)
	}

	// Page 2: limit=2, offset=2 → next 2 sessions.
	page2, err := svc.ListSessions(ctx, "u-1", 2, 2)
	if err != nil {
		t.Fatalf("list sessions page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page 2: expected 2 sessions, got %d", len(page2))
	}
	if page2[0].ID != sessionIDs[2] {
		t.Errorf("page2[0]: expected session %s, got %s", sessionIDs[2], page2[0].ID)
	}
	if page2[1].ID != sessionIDs[1] {
		t.Errorf("page2[1]: expected session %s, got %s", sessionIDs[1], page2[1].ID)
	}

	// Page 3: limit=2, offset=4 → last session (only 1 remaining).
	page3, err := svc.ListSessions(ctx, "u-1", 2, 4)
	if err != nil {
		t.Fatalf("list sessions page 3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page 3: expected 1 session (last), got %d", len(page3))
	}
	if page3[0].ID != sessionIDs[0] {
		t.Errorf("page3[0]: expected session %s (oldest), got %s", sessionIDs[0], page3[0].ID)
	}

	// Boundary: offset beyond total returns empty slice, not error.
	beyond, err := svc.ListSessions(ctx, "u-1", 2, 100)
	if err != nil {
		t.Fatalf("list sessions beyond: %v", err)
	}
	if len(beyond) != 0 {
		t.Errorf("offset beyond total: expected 0 sessions, got %d", len(beyond))
	}
}
