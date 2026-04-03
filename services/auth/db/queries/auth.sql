-- Auth service queries — generated code lives in internal/repository/

-- name: GetUserByEmail :one
SELECT id, name, password_hash, is_active, failed_logins, locked_until
FROM users WHERE email = $1;

-- name: GetActiveUserById :one
SELECT email, name
FROM users WHERE id = $1 AND is_active = true;

-- name: GetUserForRefresh :one
SELECT name, email, is_active
FROM users WHERE id = $1;

-- name: GetPrimaryRole :one
SELECT r.name FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY CASE r.name
    WHEN 'admin' THEN 1
    WHEN 'manager' THEN 2
    WHEN 'user' THEN 3
    ELSE 4
END
LIMIT 1;

-- name: GetPermissions :many
SELECT DISTINCT p.id FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1
ORDER BY p.id;

-- name: StoreRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: RevokeUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: ValidateRefreshToken :one
SELECT EXISTS(
    SELECT 1 FROM refresh_tokens
    WHERE token_hash = $1 AND user_id = $2
      AND revoked_at IS NULL AND expires_at > now()
);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = now()
WHERE token_hash = $1;

-- name: GetRefreshTokenOwner :one
SELECT user_id FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RecordFailedLogin :exec
UPDATE users SET
    failed_logins = failed_logins + 1,
    locked_until = CASE
        WHEN failed_logins + 1 >= @max_failed::int THEN now() + interval '15 minutes'
        ELSE locked_until
    END,
    is_active = CASE
        WHEN failed_logins + 1 >= @permanent_lockout::int THEN false
        ELSE is_active
    END
WHERE id = $1;

-- name: RecordSuccessfulLogin :exec
UPDATE users SET
    failed_logins = 0,
    locked_until = NULL,
    last_login_at = now(),
    last_login_ip = $2
WHERE id = $1;

-- name: GetUserEmail :one
SELECT email FROM users WHERE id = $1;

-- name: SetMFASecret :exec
UPDATE users SET mfa_secret = $2 WHERE id = $1;

-- name: EnableMFA :exec
UPDATE users SET mfa_enabled = true WHERE id = $1;

-- name: DisableMFA :exec
UPDATE users SET mfa_enabled = false, mfa_secret = NULL WHERE id = $1;

-- name: CheckMFAEnabled :one
SELECT COALESCE(mfa_enabled, false)::bool AS enabled FROM users WHERE id = $1;

-- name: GetMFASecret :one
SELECT mfa_secret FROM users WHERE id = $1;

-- name: UpdateUserName :exec
UPDATE users SET name = $2, updated_at = now()
WHERE id = $1 AND is_active = true;
