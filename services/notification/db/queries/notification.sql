-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body, data, channel)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, type, title, body, data, channel, is_read, read_at, created_at;

-- name: ListNotifications :many
SELECT id, user_id, type, title, body, data, channel, is_read, read_at, created_at
FROM notifications
WHERE user_id = $1 AND (NOT sqlc.arg(unread_only)::bool OR is_read = false)
ORDER BY created_at DESC
LIMIT $2;

-- name: UnreadCount :one
SELECT COUNT(*)::int FROM notifications
WHERE user_id = $1 AND is_read = false;

-- name: MarkRead :execrows
UPDATE notifications SET is_read = true, read_at = now()
WHERE id = $1 AND user_id = $2 AND is_read = false;

-- name: MarkAllRead :execrows
UPDATE notifications SET is_read = true, read_at = now()
WHERE user_id = $1 AND is_read = false;

-- name: NotificationExistsByUser :one
SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1 AND user_id = $2);

-- name: GetPreferences :one
SELECT user_id, email_enabled, in_app_enabled, quiet_start, quiet_end, muted_types, updated_at
FROM notification_preferences
WHERE user_id = $1;

-- name: UpsertPreferences :one
INSERT INTO notification_preferences (user_id, email_enabled, in_app_enabled, quiet_start, quiet_end, muted_types, updated_at)
VALUES (
  sqlc.arg(user_id),
  sqlc.arg(email_enabled),
  sqlc.arg(in_app_enabled),
  sqlc.arg(quiet_start)::time,
  sqlc.arg(quiet_end)::time,
  sqlc.arg(muted_types),
  now()
)
ON CONFLICT (user_id) DO UPDATE SET
  email_enabled = EXCLUDED.email_enabled,
  in_app_enabled = EXCLUDED.in_app_enabled,
  quiet_start = EXCLUDED.quiet_start,
  quiet_end = EXCLUDED.quiet_end,
  muted_types = EXCLUDED.muted_types,
  updated_at = now()
RETURNING user_id, email_enabled, in_app_enabled, quiet_start, quiet_end, muted_types, updated_at;

-- name: PurgeOldNotifications :execrows
DELETE FROM notifications
WHERE is_read = true AND created_at < now() - interval '90 days';
