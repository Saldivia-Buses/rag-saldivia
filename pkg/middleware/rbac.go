package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	permissionsKey contextKey = "permissions"
	roleKey        contextKey = "role"
)

// WithPermissions stores the user's permissions in context.
func WithPermissions(ctx context.Context, perms []string) context.Context {
	return context.WithValue(ctx, permissionsKey, perms)
}

// PermissionsFromContext returns the user's permissions from context.
func PermissionsFromContext(ctx context.Context) []string {
	perms, _ := ctx.Value(permissionsKey).([]string)
	return perms
}

// WithRole stores the user's role in context.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

// RoleFromContext returns the user's role from context.
func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(roleKey).(string)
	return role
}

// RequirePermission returns middleware that checks if the authenticated user
// has the given permission. Permissions are loaded from the JWT claims by the
// Auth middleware and stored in context.
//
// Admin role bypasses all permission checks.
//
// Usage:
//
//	r.With(middleware.RequirePermission("chat.read")).Get("/sessions", h.ListSessions)
//	r.With(middleware.RequirePermission("ingest.write")).Post("/upload", h.Upload)
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFromContext(r.Context())

			// Admin bypasses all permission checks
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			perms := PermissionsFromContext(r.Context())
			for _, p := range perms {
				if p == permission {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeJSONError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}
