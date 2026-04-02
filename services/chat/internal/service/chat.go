// Package service implements the chat business logic.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrNotOwner        = errors.New("session does not belong to user")
)

// Session represents a chat session.
type Session struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Title      string     `json:"title"`
	Collection *string    `json:"collection,omitempty"`
	IsSaved    bool       `json:"is_saved"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Message represents a chat message.
type Message struct {
	ID        string          `json:"id"`
	SessionID string          `json:"session_id"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Sources   []byte          `json:"sources,omitempty"`  // raw JSON
	Metadata  []byte          `json:"metadata,omitempty"` // raw JSON
	CreatedAt time.Time       `json:"created_at"`
}

// EventPublisher can publish notification events. Optional.
type EventPublisher interface {
	Notify(tenantSlug string, evt any) error
}

// Chat handles chat operations for a single tenant.
type Chat struct {
	db         *pgxpool.Pool
	events     EventPublisher
	tenantSlug string
}

// NewChat creates a chat service.
func NewChat(db *pgxpool.Pool, tenantSlug string, events EventPublisher) *Chat {
	return &Chat{db: db, tenantSlug: tenantSlug, events: events}
}

// CreateSession creates a new chat session.
func (c *Chat) CreateSession(ctx context.Context, userID, title string, collection *string) (*Session, error) {
	var s Session
	err := c.db.QueryRow(ctx,
		`INSERT INTO sessions (user_id, title, collection)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, title, collection, is_saved, created_at, updated_at`,
		userID, title, collection,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Collection, &s.IsSaved, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &s, nil
}

// GetSession returns a session by ID, verifying ownership.
func (c *Chat) GetSession(ctx context.Context, sessionID, userID string) (*Session, error) {
	var s Session
	err := c.db.QueryRow(ctx,
		`SELECT id, user_id, title, collection, is_saved, created_at, updated_at
		 FROM sessions WHERE id = $1`,
		sessionID,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Collection, &s.IsSaved, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	if s.UserID != userID {
		return nil, ErrNotOwner
	}
	return &s, nil
}

// ListSessions returns all sessions for a user, most recent first.
func (c *Chat) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	rows, err := c.db.Query(ctx,
		`SELECT id, user_id, title, collection, is_saved, created_at, updated_at
		 FROM sessions WHERE user_id = $1
		 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.Collection, &s.IsSaved, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []Session{} // never return nil slice
	}
	return sessions, nil
}

// DeleteSession deletes a session and all its messages.
func (c *Chat) DeleteSession(ctx context.Context, sessionID, userID string) error {
	result, err := c.db.Exec(ctx,
		`DELETE FROM sessions WHERE id = $1 AND user_id = $2`,
		sessionID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// RenameSession updates a session's title.
func (c *Chat) RenameSession(ctx context.Context, sessionID, userID, title string) error {
	result, err := c.db.Exec(ctx,
		`UPDATE sessions SET title = $3, updated_at = now()
		 WHERE id = $1 AND user_id = $2`,
		sessionID, userID, title,
	)
	if err != nil {
		return fmt.Errorf("rename session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// AddMessage adds a message to a session.
func (c *Chat) AddMessage(ctx context.Context, sessionID, role, content string, sources, metadata []byte) (*Message, error) {
	var m Message
	err := c.db.QueryRow(ctx,
		`INSERT INTO messages (session_id, role, content, sources, metadata)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, session_id, role, content, sources, metadata, created_at`,
		sessionID, role, content, sources, metadata,
	).Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Sources, &m.Metadata, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	// Touch session updated_at
	c.db.Exec(ctx, `UPDATE sessions SET updated_at = now() WHERE id = $1`, sessionID)

	// Broadcast new message to WS Hub for real-time updates
	if c.events != nil && c.tenantSlug != "" {
		data, _ := json.Marshal(map[string]string{"session_id": sessionID, "message_id": m.ID})
		err := c.events.Notify(c.tenantSlug, map[string]any{
			"user_id": "",
			"type":    "chat.new_message",
			"title":   "Nuevo mensaje",
			"body":    truncate(content, 100),
			"channel": "in_app",
			"data":    json.RawMessage(data),
		})
		if err != nil {
			slog.Warn("failed to publish chat event", "error", err)
		}
	}

	return &m, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetMessages returns all messages for a session, oldest first.
func (c *Chat) GetMessages(ctx context.Context, sessionID string) ([]Message, error) {
	rows, err := c.db.Query(ctx,
		`SELECT id, session_id, role, content, sources, metadata, created_at
		 FROM messages WHERE session_id = $1
		 ORDER BY created_at`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Sources, &m.Metadata, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []Message{}
	}
	return messages, nil
}
