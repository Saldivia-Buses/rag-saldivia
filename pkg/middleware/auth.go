// Package middleware provides shared HTTP middleware for SDA services.
package middleware

import (
	"crypto/ed25519"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"
)

// AuthConfig holds optional dependencies for the auth middleware.
type AuthConfig struct {
	Blacklist *security.TokenBlacklist // nil = no blacklist checking
	FailOpen  bool                     // true = allow on Redis error, false = reject (default)
}

// Auth returns a chi middleware that verifies the JWT from the Authorization
// header using an Ed25519 public key and injects identity into the request:
//   - X-User-ID header (user UUID)
//   - X-User-Email header
//   - X-User-Role header
//   - X-Tenant-ID header (tenant UUID)
//   - X-Tenant-Slug header
//   - tenant.Info in context (via pkg/tenant)
//
// Requests without a valid JWT get a 401 response.
// The /health endpoint is excluded from auth.
func Auth(publicKey ed25519.PublicKey) func(http.Handler) http.Handler {
	return AuthWithConfig(publicKey, AuthConfig{})
}

// AuthWithConfig is like Auth but accepts optional configuration (blacklist, etc).
func AuthWithConfig(publicKey ed25519.PublicKey, cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Save Traefik-injected slug before stripping (Traefik re-injects after client spoofing is stripped)
			traefikSlug := r.Header.Get("X-Tenant-Slug")

			// Strip any client-spoofed identity headers before processing
			r.Header.Del("X-User-ID")
			r.Header.Del("X-User-Email")
			r.Header.Del("X-User-Role")
			r.Header.Del("X-Tenant-ID")
			r.Header.Del("X-Tenant-Slug")

			// Skip health checks (exact path, with or without trailing slash)
			path := strings.TrimRight(r.URL.Path, "/")
			if path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			token := extractBearer(r)
			if token == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing authorization")
				return
			}

			claims, err := sdajwt.Verify(publicKey, token)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			// Check token blacklist (revoked on logout/password change)
			if cfg.Blacklist != nil {
				if claims.ID == "" {
					writeJSONError(w, http.StatusUnauthorized, "invalid token")
					return
				}
				revoked, err := cfg.Blacklist.IsRevoked(r.Context(), claims.ID)
				if err != nil {
					slog.Error("blacklist check failed", "error", err)
					if !cfg.FailOpen {
						writeJSONError(w, http.StatusServiceUnavailable, "auth check unavailable")
						return
					}
				} else if revoked {
					writeJSONError(w, http.StatusUnauthorized, "token revoked")
					return
				}
			}

			// Reject MFA-pending tokens — they're only valid for /v1/auth/mfa/verify
			if claims.Role == "mfa_pending" {
				writeJSONError(w, http.StatusUnauthorized, "mfa verification required")
				return
			}

			// Inject identity headers for downstream handlers
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)
			r.Header.Set("X-Tenant-ID", claims.TenantID)
			r.Header.Set("X-Tenant-Slug", claims.Slug)

			// Set tenant context for pkg/tenant consumers
			ctx := tenant.WithInfo(r.Context(), tenant.Info{
				ID:   claims.TenantID,
				Slug: claims.Slug,
			})

			// Inject role + permissions into context for RBAC middleware
			ctx = WithRole(ctx, claims.Role)
			ctx = WithPermissions(ctx, claims.Permissions)

			// Cross-validate: JWT slug must match subdomain-derived slug (prevents token replay)
			if traefikSlug != "" && claims.Slug != traefikSlug {
				writeJSONError(w, http.StatusForbidden, "tenant mismatch")
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
