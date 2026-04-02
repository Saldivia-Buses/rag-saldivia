// Package service implements the auth business logic.
package service

import (
	"context"
	"errors"
	"fmt"
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
	ErrAccountDisabled    = errors.New("account is disabled")
)

const (
	bcryptCost       = 12
	maxFailedLogins  = 5
	lockoutDuration  = 15 * time.Minute
)

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
	Email    string
	Password string
	IP       string
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
		 FROM users WHERE email = $1`, req.Email,
	).Scan(&userID, &name, &passwordHash, &isActive, &failedLogins, &lockedUntil)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if !isActive {
		return nil, ErrAccountDisabled
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
	role := a.getPrimaryRole(ctx, userID)

	// Create tokens
	jti := uuid.New().String()
	claims := sdajwt.Claims{
		UserID:   userID,
		Email:    req.Email,
		Name:     name,
		TenantID: a.tenant.ID,
		Slug:     a.tenant.Slug,
		Role:     role,
	}
	claims.ID = jti

	accessToken, err := sdajwt.CreateAccess(a.jwtCfg, claims)
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}

	refreshToken, err := sdajwt.CreateRefresh(a.jwtCfg, claims)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	// Store refresh token hash
	refreshHash, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	_, _ = a.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, string(refreshHash), time.Now().Add(a.jwtCfg.RefreshExpiry),
	)

	// Audit log
	a.audit(ctx, userID, "user.login", req.Email, req.IP, req.UserAgent)

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

func (a *Auth) recordFailedLogin(ctx context.Context, userID string) {
	_, _ = a.db.Exec(ctx,
		`UPDATE users SET failed_logins = failed_logins + 1,
		 locked_until = CASE WHEN failed_logins + 1 >= $2 THEN now() + interval '15 minutes' ELSE locked_until END
		 WHERE id = $1`,
		userID, maxFailedLogins,
	)
}

func (a *Auth) recordSuccessfulLogin(ctx context.Context, userID, ip string) {
	_, _ = a.db.Exec(ctx,
		`UPDATE users SET failed_logins = 0, locked_until = NULL,
		 last_login_at = now(), last_login_ip = $2
		 WHERE id = $1`,
		userID, ip,
	)
}

func (a *Auth) getPrimaryRole(ctx context.Context, userID string) string {
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
		return "user" // default
	}
	return role
}

func (a *Auth) audit(ctx context.Context, userID, action, resource, ip, ua string) {
	_, _ = a.db.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, resource, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, action, resource, ip, ua,
	)
}
