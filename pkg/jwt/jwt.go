// Package jwt provides JWT creation and verification for SDA Framework.
// Access tokens are short-lived (15min). Refresh tokens are long-lived (7d)
// and stored hashed in the tenant DB. Every service verifies JWTs locally
// using the shared secret — no round-trip to Auth Service per request.
package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrMissingClaim  = errors.New("missing required claim")
	ErrSecretTooShort = errors.New("JWT secret must be at least 32 bytes")
)

// Claims holds the custom claims for an SDA JWT.
type Claims struct {
	gojwt.RegisteredClaims
	UserID   string `json:"uid"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	TenantID string `json:"tid"`   // tenant UUID
	Slug     string `json:"slug"`  // tenant subdomain slug
	Role     string `json:"role"`  // primary role name
}

// Config holds JWT signing configuration.
type Config struct {
	Secret        string
	AccessExpiry  time.Duration // default 15 min
	RefreshExpiry time.Duration // default 7 days
	Issuer        string        // default "sda"
}

// DefaultConfig returns sensible defaults.
func DefaultConfig(secret string) Config {
	return Config{
		Secret:        secret,
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "sda",
	}
}

// CreateAccess creates a short-lived access token.
func CreateAccess(cfg Config, claims Claims) (string, error) {
	if len(cfg.Secret) < 32 {
		return "", ErrSecretTooShort
	}

	now := time.Now()
	claims.RegisteredClaims = gojwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   claims.UserID, // RFC 7519: sub = principal identifier
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(cfg.AccessExpiry)),
		ID:        claims.ID,
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// CreateRefresh creates a long-lived refresh token.
func CreateRefresh(cfg Config, claims Claims) (string, error) {
	if len(cfg.Secret) < 32 {
		return "", ErrSecretTooShort
	}

	now := time.Now()
	claims.RegisteredClaims = gojwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   claims.UserID,
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(cfg.RefreshExpiry)),
		ID:        claims.ID,
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// Verify parses and validates a JWT, returning the claims.
func Verify(secret string, tokenString string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.UserID == "" || claims.TenantID == "" || claims.Slug == "" {
		return nil, ErrMissingClaim
	}

	return claims, nil
}
