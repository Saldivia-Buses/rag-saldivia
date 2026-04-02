// Package service implements the notification business logic.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrNotOwner             = errors.New("notification does not belong to user")
)

// Notification represents an in-app notification.
type Notification struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Data      json.RawMessage `json:"data"`
	Channel   string          `json:"channel"`
	IsRead    bool            `json:"is_read"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// Preferences represents a user's notification preferences.
type Preferences struct {
	UserID       string   `json:"user_id"`
	EmailEnabled bool     `json:"email_enabled"`
	InAppEnabled bool     `json:"in_app_enabled"`
	QuietStart   *string  `json:"quiet_start,omitempty"`
	QuietEnd     *string  `json:"quiet_end,omitempty"`
	MutedTypes   []string `json:"muted_types"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Notification service handles notification operations for a single tenant.
type NotificationService struct {
	db *pgxpool.Pool
}

// New creates a notification service.
func New(db *pgxpool.Pool) *NotificationService {
	return &NotificationService{db: db}
}

// Create persists a notification in the database.
func (s *NotificationService) Create(ctx context.Context, userID, notifType, title, body string, data json.RawMessage, channel string) (*Notification, error) {
	if channel == "" {
		channel = "in_app"
	}
	if data == nil {
		data = []byte("{}")
	}

	var n Notification
	err := s.db.QueryRow(ctx,
		`INSERT INTO notifications (user_id, type, title, body, data, channel)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, type, title, body, data, channel, is_read, read_at, created_at`,
		userID, notifType, title, body, data, channel,
	).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.Channel, &n.IsRead, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return &n, nil
}

// List returns notifications for a user, most recent first.
// If unreadOnly is true, only returns unread notifications.
func (s *NotificationService) List(ctx context.Context, userID string, unreadOnly bool, limit int) ([]Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `SELECT id, user_id, type, title, body, data, channel, is_read, read_at, created_at
		 FROM notifications WHERE user_id = $1`
	if unreadOnly {
		query += ` AND is_read = false`
	}
	query += ` ORDER BY created_at DESC LIMIT $2`

	rows, err := s.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.Channel, &n.IsRead, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	if notifications == nil {
		notifications = []Notification{}
	}
	return notifications, nil
}

// UnreadCount returns the number of unread notifications for a user.
func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, notifID, userID string) error {
	result, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_read = true, read_at = now()
		 WHERE id = $1 AND user_id = $2 AND is_read = false`,
		notifID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	if result.RowsAffected() == 0 {
		// Check existence scoped to user — never leak other users' notification IDs
		var exists bool
		s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1 AND user_id = $2)`,
			notifID, userID,
		).Scan(&exists)
		if !exists {
			return ErrNotificationNotFound
		}
		// Exists but already read — no-op
	}
	return nil
}

// MarkAllRead marks all unread notifications as read for a user.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	result, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_read = true, read_at = now()
		 WHERE user_id = $1 AND is_read = false`,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("mark all read: %w", err)
	}
	return result.RowsAffected(), nil
}

// GetPreferences returns notification preferences for a user.
// Returns defaults if no preferences are saved.
func (s *NotificationService) GetPreferences(ctx context.Context, userID string) (*Preferences, error) {
	var p Preferences
	var quietStart, quietEnd *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT user_id, email_enabled, in_app_enabled, quiet_start, quiet_end, muted_types, updated_at
		 FROM notification_preferences WHERE user_id = $1`,
		userID,
	).Scan(&p.UserID, &p.EmailEnabled, &p.InAppEnabled, &quietStart, &quietEnd, &p.MutedTypes, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &Preferences{
				UserID:       userID,
				EmailEnabled: true,
				InAppEnabled: true,
				MutedTypes:   []string{},
				UpdatedAt:    time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("get preferences: %w", err)
	}
	if quietStart != nil {
		s := quietStart.Format("15:04")
		p.QuietStart = &s
	}
	if quietEnd != nil {
		s := quietEnd.Format("15:04")
		p.QuietEnd = &s
	}
	if p.MutedTypes == nil {
		p.MutedTypes = []string{}
	}
	return &p, nil
}

// UpdatePreferences upserts notification preferences for a user.
func (s *NotificationService) UpdatePreferences(ctx context.Context, userID string, emailEnabled, inAppEnabled bool, quietStart, quietEnd *string, mutedTypes []string) (*Preferences, error) {
	if mutedTypes == nil {
		mutedTypes = []string{}
	}

	var p Preferences
	var qs, qe *time.Time
	err := s.db.QueryRow(ctx,
		`INSERT INTO notification_preferences (user_id, email_enabled, in_app_enabled, quiet_start, quiet_end, muted_types, updated_at)
		 VALUES ($1, $2, $3, $4::time, $5::time, $6, now())
		 ON CONFLICT (user_id) DO UPDATE SET
		   email_enabled = EXCLUDED.email_enabled,
		   in_app_enabled = EXCLUDED.in_app_enabled,
		   quiet_start = EXCLUDED.quiet_start,
		   quiet_end = EXCLUDED.quiet_end,
		   muted_types = EXCLUDED.muted_types,
		   updated_at = now()
		 RETURNING user_id, email_enabled, in_app_enabled, quiet_start, quiet_end, muted_types, updated_at`,
		userID, emailEnabled, inAppEnabled, quietStart, quietEnd, mutedTypes,
	).Scan(&p.UserID, &p.EmailEnabled, &p.InAppEnabled, &qs, &qe, &p.MutedTypes, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update preferences: %w", err)
	}
	if qs != nil {
		s := qs.Format("15:04")
		p.QuietStart = &s
	}
	if qe != nil {
		s := qe.Format("15:04")
		p.QuietEnd = &s
	}
	if p.MutedTypes == nil {
		p.MutedTypes = []string{}
	}
	return &p, nil
}
