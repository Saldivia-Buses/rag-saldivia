// Package jwt provides JWT creation and verification for SDA Framework.
// Access tokens are short-lived (15min). Refresh tokens are long-lived (7d)
// and stored hashed in the tenant DB.
//
// Uses Ed25519 asymmetric signing: Auth Service signs with the private key,
// all other services verify with the public key only. A compromised service
// cannot forge tokens.
package jwt

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrMissingClaim = errors.New("missing required claim")
	ErrInvalidKey   = errors.New("invalid Ed25519 key")
)

// Claims holds the custom claims for an SDA JWT.
type Claims struct {
	gojwt.RegisteredClaims
	UserID      string   `json:"uid"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	TenantID    string   `json:"tid"`            // tenant UUID
	Slug        string   `json:"slug"`           // tenant subdomain slug
	Role        string   `json:"role"`           // primary role name
	Permissions []string `json:"perms,omitempty"` // RBAC permissions (Phase 2)
}

// Config holds JWT signing configuration.
type Config struct {
	PrivateKey    ed25519.PrivateKey // only needed by Auth Service (signing)
	PublicKey     ed25519.PublicKey  // needed by all services (verification)
	AccessExpiry  time.Duration     // default 15 min
	RefreshExpiry time.Duration     // default 7 days
	Issuer        string            // default "sda"
}

// DefaultConfig returns sensible defaults for the Auth Service (signs + verifies).
func DefaultConfig(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) Config {
	return Config{
		PrivateKey:    privateKey,
		PublicKey:     publicKey,
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "sda",
	}
}

// VerifyOnlyConfig returns config for services that only verify (no signing).
func VerifyOnlyConfig(publicKey ed25519.PublicKey) Config {
	return Config{
		PublicKey: publicKey,
		Issuer:   "sda",
	}
}

// CreateAccess creates a short-lived access token. Requires private key.
func CreateAccess(cfg Config, claims Claims) (string, error) {
	if cfg.PrivateKey == nil {
		return "", fmt.Errorf("%w: private key required for signing", ErrInvalidKey)
	}

	now := time.Now()
	claims.RegisteredClaims = gojwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   claims.UserID,
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(cfg.AccessExpiry)),
		ID:        claims.ID,
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, claims)
	return token.SignedString(cfg.PrivateKey)
}

// CreateRefresh creates a long-lived refresh token. Requires private key.
func CreateRefresh(cfg Config, claims Claims) (string, error) {
	if cfg.PrivateKey == nil {
		return "", fmt.Errorf("%w: private key required for signing", ErrInvalidKey)
	}

	now := time.Now()
	claims.RegisteredClaims = gojwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   claims.UserID,
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(cfg.RefreshExpiry)),
		ID:        claims.ID,
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, claims)
	return token.SignedString(cfg.PrivateKey)
}

// Verify parses and validates a JWT using the public key, returning the claims.
func Verify(publicKey ed25519.PublicKey, tokenString string) (*Claims, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("%w: public key required for verification", ErrInvalidKey)
	}

	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
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

// ParsePrivateKeyPEM parses a PEM-encoded Ed25519 private key.
func ParsePrivateKeyPEM(pemData []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("%w: no PEM block found", ErrInvalidKey)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: not an Ed25519 private key", ErrInvalidKey)
	}

	return edKey, nil
}

// ParsePublicKeyPEM parses a PEM-encoded Ed25519 public key.
func ParsePublicKeyPEM(pemData []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("%w: no PEM block found", ErrInvalidKey)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: not an Ed25519 public key", ErrInvalidKey)
	}

	return edKey, nil
}

// ParsePrivateKeyEnv parses a base64-encoded PEM Ed25519 private key (for env vars).
func ParsePrivateKeyEnv(b64 string) (ed25519.PrivateKey, error) {
	pemData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode failed: %v", ErrInvalidKey, err)
	}
	return ParsePrivateKeyPEM(pemData)
}

// ParsePublicKeyEnv parses a base64-encoded PEM Ed25519 public key (for env vars).
func ParsePublicKeyEnv(b64 string) (ed25519.PublicKey, error) {
	pemData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode failed: %v", ErrInvalidKey, err)
	}
	return ParsePublicKeyPEM(pemData)
}
