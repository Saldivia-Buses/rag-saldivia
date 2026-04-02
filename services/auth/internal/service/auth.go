// Package service implements the auth business logic.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked")
)

const (
	bcryptCost      = 12
	maxFailedLogins = 5
)

// dummyHash is used for timing-safe responses when the user doesn't exist.
// Prevents enumeration via response timing differences.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)

// Auth handles authentication operations for a single tenant.
type Auth struct {
	db     *pgxpool.Pool
	jwtCfg sdajwt.Config
	tenant struct {
		ID   string
		Slug string
	}
}

// NewAuth creates an auth service for a specific tenant.
func NewAuth(db *pgxpool.Pool, jwtCfg sdajwt.Config, tenantID, tenantSlug string) *Auth {
	a := &Auth{db: db, jwtCfg: jwtCfg}
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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
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
		return nil, ErrInvalidCredentials
	}

	// Success — reset failed logins, record login
	a.recordSuccessfulLogin(ctx, userID, req.IP)

	// Get primary role
	role, err := a.getPrimaryRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	// Create tokens with separate JTIs
	accessClaims := sdajwt.Claims{
		UserID:   userID,
		Email:    email,
		Name:     name,
		TenantID: a.tenant.ID,
		Slug:     a.tenant.Slug,
		Role:     role,
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
	a.audit(ctx, userID, "user.login", email, req.IP, req.UserAgent)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(a.jwtCfg.AccessExpiry.Seconds()),
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
		 locked_until = CASE WHEN failed_logins + 1 >= $2 THEN now() + interval '15 minutes' ELSE locked_until END
		 WHERE id = $1`,
		userID, maxFailedLogins,
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

func (a *Auth) audit(ctx context.Context, userID, action, resource, ip, ua string) {
	_, err := a.db.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, resource, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, action, resource, ip, ua,
	)
	if err != nil {
		slog.Error("failed to write audit log", "error", err, "action", action)
	}
}
