// Package service implements the auth business logic.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountLocked       = errors.New("account is locked")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidMFACode      = errors.New("invalid MFA code")
	ErrMFARequired         = errors.New("MFA verification required")
)

const (
	bcryptCost              = 12
	maxFailedLogins         = 5  // temporary lockout threshold (15 min)
	permanentLockoutLogins  = 20 // permanent lockout threshold (admin reset required)
)

// dummyHash is used for timing-safe responses when the user doesn't exist.
// Prevents enumeration via response timing differences.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)

// EventPublisher can publish notification events. Optional — if nil, no events are published.
type EventPublisher interface {
	Notify(tenantSlug string, evt any) error
}

// Auth handles authentication operations for a single tenant.
type Auth struct {
	db      *pgxpool.Pool
	jwtCfg  sdajwt.Config
	events  EventPublisher
	auditor *audit.Writer
	tenant  struct {
		ID   string
		Slug string
	}
}

// NewAuth creates an auth service for a specific tenant.
func NewAuth(db *pgxpool.Pool, jwtCfg sdajwt.Config, tenantID, tenantSlug string, events EventPublisher) *Auth {
	a := &Auth{db: db, jwtCfg: jwtCfg, events: events, auditor: audit.NewWriter(db)}
	a.tenant.ID = tenantID
	a.tenant.Slug = tenantSlug
	return a
}

// LoginRequest holds the login input.
type LoginRequest struct {
	Email     string
	Password  string
	IP        string
	UserAgent string
}

// TokenPair holds access + refresh tokens.
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresIn        int       `json:"expires_in"`      // seconds
	RefreshExpiresAt time.Time `json:"-"`                // used by handler for cookie, not serialized
	MFARequired      bool      `json:"mfa_required,omitempty"` // true when MFA pending
	MFAToken         string    `json:"mfa_token,omitempty"`    // temp JWT for MFA verification
}

// UserInfo holds the current user's profile data.
type UserInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	TenantID   string `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
}

// Login authenticates a user and returns a token pair.
func (a *Auth) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var (
		userID       string
		name         string
		passwordHash string
		isActive     bool
		failedLogins int
		lockedUntil  *time.Time
	)

	err := a.db.QueryRow(ctx,
		`SELECT id, name, password_hash, is_active, failed_logins, locked_until
		 FROM users WHERE email = $1`, email,
	).Scan(&userID, &name, &passwordHash, &isActive, &failedLogins, &lockedUntil)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Timing-safe: run bcrypt even when user doesn't exist
			// to prevent enumeration via response time
			bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	// Disabled and locked accounts return the same error as invalid credentials
	// to prevent information leakage about account state
	if !isActive {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		return nil, ErrInvalidCredentials
	}

	if lockedUntil != nil && time.Now().Before(*lockedUntil) {
		return nil, ErrAccountLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		a.recordFailedLogin(ctx, userID)
		a.auditor.Write(ctx, audit.Entry{
			UserID: userID, Action: "user.login_failed", Resource: email,
			IP: req.IP, UserAgent: req.UserAgent,
		})
		return nil, ErrInvalidCredentials
	}

	// Success — reset failed logins, record login
	a.recordSuccessfulLogin(ctx, userID, req.IP)

	// Check MFA
	mfaRequired, err := a.CheckMFARequired(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check MFA: %w", err)
	}
	if mfaRequired {
		// Return a short-lived MFA token instead of real tokens.
		// The user must complete MFA verification to get access/refresh tokens.
		mfaClaims := sdajwt.Claims{
			UserID:   userID,
			Email:    email,
			Name:     name,
			TenantID: a.tenant.ID,
			Slug:     a.tenant.Slug,
			Role:     "mfa_pending", // not a real role — signals MFA is pending
		}
		mfaClaims.ID = uuid.New().String()
		mfaCfg := a.jwtCfg
		mfaCfg.AccessExpiry = 5 * time.Minute // short-lived
		mfaToken, err := sdajwt.CreateAccess(mfaCfg, mfaClaims)
		if err != nil {
			return nil, fmt.Errorf("create MFA token: %w", err)
		}
		return &TokenPair{MFARequired: true, MFAToken: mfaToken}, nil
	}

	// Get primary role
	role, err := a.getPrimaryRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	// Load permissions for RBAC
	permissions, err := a.getPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}

	// Create tokens with separate JTIs
	accessClaims := sdajwt.Claims{
		UserID:      userID,
		Email:       email,
		Name:        name,
		TenantID:    a.tenant.ID,
		Slug:        a.tenant.Slug,
		Role:        role,
		Permissions: permissions,
	}
	accessClaims.ID = uuid.New().String()

	refreshClaims := accessClaims // copy
	refreshClaims.ID = uuid.New().String()

	accessToken, err := sdajwt.CreateAccess(a.jwtCfg, accessClaims)
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}

	refreshToken, err := sdajwt.CreateRefresh(a.jwtCfg, refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	// Store refresh token hash (SHA-256, not bcrypt — tokens are high-entropy,
	// no rainbow table risk, and bcrypt truncates at 72 bytes which JWTs exceed)
	refreshHash := hashToken(refreshToken)

	// Revoke old refresh tokens for this user
	_, err = a.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		slog.Warn("failed to revoke old refresh tokens", "error", err, "user_id", userID)
	}

	// Store new refresh token
	_, err = a.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, refreshHash, time.Now().Add(a.jwtCfg.RefreshExpiry),
	)
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	// Audit log
	a.auditor.Write(ctx, audit.Entry{
		UserID: userID, Action: "user.login", Resource: email,
		IP: req.IP, UserAgent: req.UserAgent,
	})

	// Publish login event for notifications
	a.publishEvent("auth.login_success", userID, name, email, map[string]string{
		"ip": req.IP,
	})

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(a.jwtCfg.AccessExpiry.Seconds()),
		RefreshExpiresAt: time.Now().Add(a.jwtCfg.RefreshExpiry),
	}, nil
}

// CompleteMFALogin verifies the TOTP code and issues real tokens.
// Called after the user passes the MFA challenge.
func (a *Auth) CompleteMFALogin(ctx context.Context, mfaToken, code string) (*TokenPair, error) {
	// Verify the MFA temp token
	claims, err := sdajwt.Verify(a.jwtCfg.PublicKey, mfaToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if claims.Role != "mfa_pending" {
		return nil, ErrInvalidRefreshToken
	}

	// Verify TOTP code
	if err := a.VerifyMFA(ctx, claims.UserID, code); err != nil {
		return nil, err
	}

	// Issue real tokens (same as post-password login)
	role, err := a.getPrimaryRole(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	permissions, err := a.getPermissions(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get permissions: %w", err)
	}

	accessClaims := sdajwt.Claims{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Name:        claims.Name,
		TenantID:    a.tenant.ID,
		Slug:        a.tenant.Slug,
		Role:        role,
		Permissions: permissions,
	}
	accessClaims.ID = uuid.New().String()
	refreshClaims := accessClaims
	refreshClaims.ID = uuid.New().String()

	accessToken, err := sdajwt.CreateAccess(a.jwtCfg, accessClaims)
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}
	refreshToken, err := sdajwt.CreateRefresh(a.jwtCfg, refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	refreshHash := hashToken(refreshToken)
	_, err = a.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		claims.UserID, refreshHash, time.Now().Add(a.jwtCfg.RefreshExpiry),
	)
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	a.auditor.Write(ctx, audit.Entry{
		UserID: claims.UserID, Action: "user.login", Resource: claims.Email,
		Details: map[string]any{"mfa": true},
	})

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int(a.jwtCfg.AccessExpiry.Seconds()),
		RefreshExpiresAt: time.Now().Add(a.jwtCfg.RefreshExpiry),
	}, nil
}

// Refresh validates a refresh token and returns a new token pair (rotation).
func (a *Auth) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify the JWT signature and expiry
	claims, err := sdajwt.Verify(a.jwtCfg.PublicKey, refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Verify the token hash exists in DB and is not revoked
	tokenHash := hashToken(refreshToken)
	var exists bool
	err = a.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM refresh_tokens
		 WHERE token_hash = $1 AND user_id = $2 AND revoked_at IS NULL AND expires_at > now())`,
		tokenHash, claims.UserID,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("query refresh token: %w", err)
	}
	if !exists {
		return nil, ErrInvalidRefreshToken
	}

	// Revoke the old refresh token (rotation — each token is single-use)
	_, _ = a.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`,
		tokenHash,
	)

	// Re-fetch user data to get current role (may have changed since login)
	var name, email string
	var isActive bool
	err = a.db.QueryRow(ctx,
		`SELECT name, email, is_active FROM users WHERE id = $1`, claims.UserID,
	).Scan(&name, &email, &isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("query user for refresh: %w", err)
	}
	if !isActive {
		return nil, ErrInvalidRefreshToken
	}

	role, err := a.getPrimaryRole(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get role for refresh: %w", err)
	}

	// Load permissions for RBAC
	permissions, err := a.getPermissions(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get permissions for refresh: %w", err)
	}

	// Issue new token pair
	newAccessClaims := sdajwt.Claims{
		UserID:      claims.UserID,
		Email:       email,
		Name:        name,
		TenantID:    a.tenant.ID,
		Slug:        a.tenant.Slug,
		Role:        role,
		Permissions: permissions,
	}
	newAccessClaims.ID = uuid.New().String()

	newRefreshClaims := newAccessClaims
	newRefreshClaims.ID = uuid.New().String()

	accessToken, err := sdajwt.CreateAccess(a.jwtCfg, newAccessClaims)
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}

	newRefresh, err := sdajwt.CreateRefresh(a.jwtCfg, newRefreshClaims)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	// Store new refresh token
	newHash := hashToken(newRefresh)
	refreshExpiry := time.Now().Add(a.jwtCfg.RefreshExpiry)
	_, err = a.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		claims.UserID, newHash, refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	a.auditor.Write(ctx, audit.Entry{
		UserID: claims.UserID, Action: "user.refresh",
	})

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     newRefresh,
		ExpiresIn:        int(a.jwtCfg.AccessExpiry.Seconds()),
		RefreshExpiresAt: refreshExpiry,
	}, nil
}

// Logout revokes the given refresh token.
func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)

	// Get user_id from the token record for audit logging
	var userID *string
	_ = a.db.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	).Scan(&userID)

	_, err := a.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	uid := ""
	if userID != nil {
		uid = *userID
	}
	a.auditor.Write(ctx, audit.Entry{
		UserID: uid, Action: "user.logout",
	})
	return nil
}

// Me returns profile info for the authenticated user.
func (a *Auth) Me(ctx context.Context, userID string) (*UserInfo, error) {
	var email, name string
	err := a.db.QueryRow(ctx,
		`SELECT email, name FROM users WHERE id = $1 AND is_active = true`, userID,
	).Scan(&email, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	role, err := a.getPrimaryRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	return &UserInfo{
		ID:         userID,
		Email:      email,
		Name:       name,
		Role:       role,
		TenantID:   a.tenant.ID,
		TenantSlug: a.tenant.Slug,
	}, nil
}

// HashPassword hashes a password with bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// hashToken creates a SHA-256 hex digest of a token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (a *Auth) recordFailedLogin(ctx context.Context, userID string) {
	_, err := a.db.Exec(ctx,
		`UPDATE users SET failed_logins = failed_logins + 1,
		 locked_until = CASE WHEN failed_logins + 1 >= $2 THEN now() + interval '15 minutes' ELSE locked_until END,
		 is_active = CASE WHEN failed_logins + 1 >= $3 THEN false ELSE is_active END
		 WHERE id = $1`,
		userID, maxFailedLogins, permanentLockoutLogins,
	)
	if err != nil {
		slog.Error("failed to record failed login", "error", err, "user_id", userID)
	}
}

func (a *Auth) recordSuccessfulLogin(ctx context.Context, userID, ip string) {
	_, err := a.db.Exec(ctx,
		`UPDATE users SET failed_logins = 0, locked_until = NULL,
		 last_login_at = now(), last_login_ip = $2
		 WHERE id = $1`,
		userID, ip,
	)
	if err != nil {
		slog.Error("failed to record successful login", "error", err, "user_id", userID)
	}
}

func (a *Auth) getPrimaryRole(ctx context.Context, userID string) (string, error) {
	var role string
	err := a.db.QueryRow(ctx,
		`SELECT r.name FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1
		 ORDER BY CASE r.name WHEN 'admin' THEN 1 WHEN 'manager' THEN 2 WHEN 'user' THEN 3 ELSE 4 END
		 LIMIT 1`,
		userID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "user", nil // no role assigned, default
		}
		return "", fmt.Errorf("query role for user %s: %w", userID, err)
	}
	return role, nil
}

func (a *Auth) getPermissions(ctx context.Context, userID string) ([]string, error) {
	rows, err := a.db.Query(ctx,
		`SELECT DISTINCT p.id FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 JOIN user_roles ur ON ur.role_id = rp.role_id
		 WHERE ur.user_id = $1
		 ORDER BY p.id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query permissions for user %s: %w", userID, err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (a *Auth) publishEvent(eventType, userID, name, email string, extra map[string]string) {
	if a.events == nil {
		return
	}

	data, _ := json.Marshal(extra)
	err := a.events.Notify(a.tenant.Slug, map[string]any{
		"user_id": userID,
		"type":    eventType,
		"title":   formatEventTitle(eventType, name),
		"body":    "",
		"channel": "in_app",
		"data":    json.RawMessage(data),
	})
	if err != nil {
		slog.Warn("failed to publish auth event", "error", err, "type", eventType)
	}
}

func formatEventTitle(eventType, name string) string {
	switch eventType {
	case "auth.login_success":
		return name + " inicio sesion"
	case "auth.account_locked":
		return "Cuenta bloqueada: " + name
	default:
		return eventType
	}
}

