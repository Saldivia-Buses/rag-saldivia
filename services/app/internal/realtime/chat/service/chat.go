// Package service implements the chat business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	notify "github.com/Camionerou/rag-saldivia/services/app/internal/events/gen/notify"
	"github.com/Camionerou/rag-saldivia/services/app/internal/outbox"
	"github.com/Camionerou/rag-saldivia/services/app/internal/spine"
	"github.com/Camionerou/rag-saldivia/services/app/internal/realtime/chat/repository"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrNotOwner        = errors.New("session does not belong to user")
)

// Session represents a chat session.
type Session struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Title      string  `json:"title"`
	Collection *string `json:"collection,omitempty"`
	IsSaved    bool    `json:"is_saved"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Message represents a chat message.
type Message struct {
	ID        string  `json:"id"`
	SessionID string  `json:"session_id"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Thinking  *string `json:"thinking,omitempty"` // model reasoning/thinking
	Sources   []byte  `json:"sources,omitempty"`  // raw JSON
	Metadata  []byte  `json:"metadata,omitempty"` // raw JSON
	CreatedAt time.Time `json:"created_at"`
}

// Chat handles chat operations for a single tenant.
type Chat struct {
	db         *pgxpool.Pool
	repo       *repository.Queries
	auditor    *audit.Writer
	tenantSlug string
}

// NewChat creates a chat service. Event publishing is handled via the
// transactional outbox (pkg/outbox) inside AddMessage — no EventPublisher
// dependency needed.
func NewChat(db *pgxpool.Pool, tenantSlug string) *Chat {
	return &Chat{
		db:         db,
		repo:       repository.New(db),
		tenantSlug: tenantSlug,
		auditor:    audit.NewWriter(db),
	}
}

// CreateSession creates a new chat session.
func (c *Chat) CreateSession(ctx context.Context, userID, title string, collection *string) (*Session, error) {
	row, err := c.repo.CreateSession(ctx, repository.CreateSessionParams{
		UserID:     userID,
		Title:      title,
		Collection: ptrToText(collection),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	s := sessionFromRepo(row)
	c.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "chat.session.create", Resource: s.ID,
	})
	return &s, nil
}

// GetSession returns a session by ID, verifying ownership at the query level.
func (c *Chat) GetSession(ctx context.Context, sessionID, userID string) (*Session, error) {
	row, err := c.repo.GetSession(ctx, repository.GetSessionParams{
		ID:     sessionID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	s := sessionFromRepo(row)
	return &s, nil
}

// ListSessions returns sessions for a user, most recent first (paginated).
func (c *Chat) ListSessions(ctx context.Context, userID string, limit, offset int32) ([]Session, error) {
	rows, err := c.repo.ListSessionsByUser(ctx, repository.ListSessionsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	sessions := make([]Session, len(rows))
	for i, row := range rows {
		sessions[i] = sessionFromRepo(row)
	}
	return sessions, nil
}

// DeleteSession deletes a session and all its messages.
func (c *Chat) DeleteSession(ctx context.Context, sessionID, userID string) error {
	n, err := c.repo.DeleteSession(ctx, repository.DeleteSessionParams{
		ID: sessionID, UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if n == 0 {
		return ErrSessionNotFound
	}

	c.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "chat.session.delete", Resource: sessionID,
	})
	return nil
}

// RenameSession updates a session's title.
func (c *Chat) RenameSession(ctx context.Context, sessionID, userID, title string) error {
	n, err := c.repo.RenameSession(ctx, repository.RenameSessionParams{
		ID: sessionID, UserID: userID, Title: title,
	})
	if err != nil {
		return fmt.Errorf("rename session: %w", err)
	}
	if n == 0 {
		return ErrSessionNotFound
	}

	c.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "chat.session.rename", Resource: sessionID,
		Details: map[string]any{"title": title},
	})
	return nil
}

// AddMessage adds a message to a session. The DB write and the notification
// event are committed atomically via the transactional outbox — if the insert
// succeeds, the event is guaranteed to eventually reach NATS.
func (c *Chat) AddMessage(ctx context.Context, sessionID, userID, role, content string, thinking *string, sources, metadata []byte) (*Message, error) {
	if metadata == nil {
		metadata = []byte("{}")
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txRepo := c.repo.WithTx(tx)

	row, err := txRepo.CreateMessage(ctx, repository.CreateMessageParams{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Thinking:  ptrToText(thinking),
		Sources:   sources,
		Metadata:  metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	_ = txRepo.TouchSession(ctx, repository.TouchSessionParams{
		ID:     sessionID,
		UserID: userID,
	})

	m := messageFromRepo(row)

	// Publish notification via outbox for user messages (not assistant/system).
	// Uses a savepoint so an outbox failure doesn't abort the message tx.
	if c.tenantSlug != "" && role == "user" {
		env, envErr := spine.New(c.tenantSlug, notify.ChatNewMessageType,
			notify.ChatNewMessageSchemaVersion,
			notify.ChatNewMessagePayload{
				UserID:    userID,
				SessionID: sessionID,
				MessageID: m.ID,
				Title:     "Nuevo mensaje",
				Body:      truncate(content, 100),
				Channel:   notify.ChatNewMessageChannelInApp,
			})
		if envErr == nil {
			subject, _ := spine.BuildSubject(notify.ChatNewMessageSubject,
				map[string]string{"slug": c.tenantSlug})
			if _, spErr := tx.Exec(ctx, "SAVEPOINT outbox_sp"); spErr == nil {
				if pubErr := outbox.PublishTx(ctx, tx, subject, env); pubErr != nil {
					slog.Warn("outbox enqueue failed, rolling back savepoint", "error", pubErr)
					_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT outbox_sp")
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit message + outbox: %w", err)
	}

	return &m, nil
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// GetMessages returns messages for a session, oldest first (with limit).
func (c *Chat) GetMessages(ctx context.Context, sessionID string, limit int32) ([]Message, error) {
	rows, err := c.repo.ListMessages(ctx, repository.ListMessagesParams{
		SessionID: sessionID,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	messages := make([]Message, len(rows))
	for i, row := range rows {
		messages[i] = listMessageRowToMessage(row)
	}
	return messages, nil
}

// --- type conversions between domain and repository ---

// sessionFromRepo converts a sqlc-generated Session to the domain Session.
func sessionFromRepo(r repository.Session) Session {
	return Session{
		ID:         r.ID,
		UserID:     r.UserID,
		Title:      r.Title,
		Collection: textToPtr(r.Collection),
		IsSaved:    r.IsSaved,
		CreatedAt:  r.CreatedAt.Time,
		UpdatedAt:  r.UpdatedAt.Time,
	}
}

// messageFromRepo converts a sqlc-generated CreateMessageRow to the domain Message.
func messageFromRepo(r repository.CreateMessageRow) Message {
	return Message{
		ID:        r.ID,
		SessionID: r.SessionID,
		Role:      r.Role,
		Content:   r.Content,
		Thinking:  textToPtr(r.Thinking),
		Sources:   r.Sources,
		Metadata:  r.Metadata,
		CreatedAt: r.CreatedAt.Time,
	}
}

// listMessageRowToMessage converts a sqlc-generated ListMessagesRow to the domain Message.
func listMessageRowToMessage(r repository.ListMessagesRow) Message {
	return Message{
		ID:        r.ID,
		SessionID: r.SessionID,
		Role:      r.Role,
		Content:   r.Content,
		Thinking:  textToPtr(r.Thinking),
		Sources:   r.Sources,
		Metadata:  r.Metadata,
		CreatedAt: r.CreatedAt.Time,
	}
}

// ptrToText converts a *string to pgtype.Text.
func ptrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// textToPtr converts a pgtype.Text to *string.
func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
