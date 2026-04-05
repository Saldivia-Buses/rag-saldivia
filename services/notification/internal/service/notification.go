// Package service implements the notification business logic.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/services/notification/internal/repository"
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
	db      *pgxpool.Pool
	repo    *repository.Queries
	auditor *audit.Writer
}

// New creates a notification service.
func New(db *pgxpool.Pool) *NotificationService {
	return &NotificationService{
		db:      db,
		repo:    repository.New(db),
		auditor: audit.NewWriter(db),
	}
}

// Create persists a notification in the database.
func (s *NotificationService) Create(ctx context.Context, userID, notifType, title, body string, data json.RawMessage, channel string) (*Notification, error) {
	if channel == "" {
		channel = "in_app"
	}
	if data == nil {
		data = []byte("{}")
	}

	row, err := s.repo.CreateNotification(ctx, repository.CreateNotificationParams{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Body:    body,
		Data:    []byte(data),
		Channel: channel,
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return repoToNotification(row), nil
}

// List returns notifications for a user, most recent first.
// If unreadOnly is true, only returns unread notifications.
func (s *NotificationService) List(ctx context.Context, userID string, unreadOnly bool, limit int) ([]Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.repo.ListNotifications(ctx, repository.ListNotificationsParams{
		UserID:     userID,
		Limit:      int32(limit),
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	notifications := make([]Notification, 0, len(rows))
	for _, r := range rows {
		notifications = append(notifications, *repoToNotification(r))
	}
	return notifications, nil
}

// UnreadCount returns the number of unread notifications for a user.
func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return int(count), nil
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, notifID, userID string) error {
	rowsAffected, err := s.repo.MarkRead(ctx, repository.MarkReadParams{
		ID:     notifID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	if rowsAffected == 0 {
		// Check existence scoped to user — never leak other users' notification IDs
		exists, err := s.repo.NotificationExistsByUser(ctx, repository.NotificationExistsByUserParams{
			ID:     notifID,
			UserID: userID,
		})
		if err != nil || !exists {
			return ErrNotificationNotFound
		}
		// Exists but already read — no-op
	}
	return nil
}

// MarkAllRead marks all unread notifications as read for a user.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	count, err := s.repo.MarkAllRead(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("mark all read: %w", err)
	}
	return count, nil
}

// GetPreferences returns notification preferences for a user.
// Returns defaults if no preferences are saved.
func (s *NotificationService) GetPreferences(ctx context.Context, userID string) (*Preferences, error) {
	row, err := s.repo.GetPreferences(ctx, userID)
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
	return repoToPreferences(row), nil
}

// UpdatePreferences upserts notification preferences for a user.
func (s *NotificationService) UpdatePreferences(ctx context.Context, userID string, emailEnabled, inAppEnabled bool, quietStart, quietEnd *string, mutedTypes []string) (*Preferences, error) {
	if mutedTypes == nil {
		mutedTypes = []string{}
	}

	row, err := s.repo.UpsertPreferences(ctx, repository.UpsertPreferencesParams{
		UserID:       userID,
		EmailEnabled: emailEnabled,
		InAppEnabled: inAppEnabled,
		QuietStart:   stringToPgTime(quietStart),
		QuietEnd:     stringToPgTime(quietEnd),
		MutedTypes:   mutedTypes,
	})
	if err != nil {
		return nil, fmt.Errorf("update preferences: %w", err)
	}

	s.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "notification.preferences.update",
	})
	return repoToPreferences(row), nil
}

// --- type conversion helpers ---

// repoToNotification converts a sqlc-generated Notification to the domain type.
func repoToNotification(r repository.Notification) *Notification {
	n := &Notification{
		ID:      r.ID,
		UserID:  r.UserID,
		Type:    r.Type,
		Title:   r.Title,
		Body:    r.Body,
		Data:    json.RawMessage(r.Data),
		Channel: r.Channel,
		IsRead:  r.IsRead,
	}
	if r.ReadAt.Valid {
		t := r.ReadAt.Time
		n.ReadAt = &t
	}
	if r.CreatedAt.Valid {
		n.CreatedAt = r.CreatedAt.Time
	}
	return n
}

// repoToPreferences converts a sqlc-generated NotificationPreference to the domain type.
func repoToPreferences(r repository.NotificationPreference) *Preferences {
	p := &Preferences{
		UserID:       r.UserID,
		EmailEnabled: r.EmailEnabled,
		InAppEnabled: r.InAppEnabled,
		MutedTypes:   r.MutedTypes,
	}
	if r.UpdatedAt.Valid {
		p.UpdatedAt = r.UpdatedAt.Time
	}
	if r.QuietStart.Valid {
		s := fmt.Sprintf("%02d:%02d", r.QuietStart.Microseconds/3600000000, (r.QuietStart.Microseconds%3600000000)/60000000)
		p.QuietStart = &s
	}
	if r.QuietEnd.Valid {
		s := fmt.Sprintf("%02d:%02d", r.QuietEnd.Microseconds/3600000000, (r.QuietEnd.Microseconds%3600000000)/60000000)
		p.QuietEnd = &s
	}
	if p.MutedTypes == nil {
		p.MutedTypes = []string{}
	}
	return p
}

// stringToPgTime converts a "HH:MM" string pointer to pgtype.Time.
func stringToPgTime(s *string) pgtype.Time {
	if s == nil || *s == "" {
		return pgtype.Time{Valid: false}
	}
	t, err := time.Parse("15:04", *s)
	if err != nil {
		return pgtype.Time{Valid: false}
	}
	micros := int64(t.Hour())*3600000000 + int64(t.Minute())*60000000
	return pgtype.Time{Microseconds: micros, Valid: true}
}
