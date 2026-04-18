package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	permissionsKey contextKey = "permissions"
	roleKey        contextKey = "role"
	userIDKey      contextKey = "user_id"
	userEmailKey   contextKey = "user_email"
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

// WithUserID stores the user's ID in context.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext returns the user's ID from context.
func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// WithUserEmail stores the user's email in context.
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailKey, email)
}

// UserEmailFromContext returns the user's email from context.
func UserEmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(userEmailKey).(string)
	return email
}

// RequirePermission returns middleware that checks if the authenticated user
// has the given permission. Permissions are loaded from the JWT claims by the
// Auth middleware and stored in context.
//
// Supports wildcard permissions: a user with "erp.accounting.*" satisfies
// RequirePermission("erp.accounting.read"). The wildcard must be the last
// segment (e.g., "erp.*" matches "erp.anything.read").
//
// Admin role bypasses all permission checks.
//
// Usage:
//
//	r.With(middleware.RequirePermission("chat.read")).Get("/sessions", h.ListSessions)
//	r.With(middleware.RequirePermission("erp.accounting.write")).Post("/entries", h.Create)
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if HasPermission(r.Context(), permission) {
				next.ServeHTTP(w, r)
				return
			}
			writeJSONError(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

// HasPermission returns true when the context's identity satisfies the given
// permission. Admin role bypasses the check. Same wildcard semantics as
// RequirePermission ("erp.accounting.*" satisfies "erp.accounting.read",
// bare "*" satisfies everything).
//
// Non-HTTP callers (agent tool dispatcher, NATS consumers, background
// agents) use this to share a single RBAC implementation with the HTTP
// middleware.
func HasPermission(ctx context.Context, permission string) bool {
	if RoleFromContext(ctx) == "admin" {
		return true
	}
	for _, p := range PermissionsFromContext(ctx) {
		if matchPermission(p, permission) {
			return true
		}
	}
	return false
}

// RequireModule returns middleware that checks if the request's tenant has
// the given module enabled. Reads enabled modules from context (set by auth
// middleware or a preceding middleware that queries tenant_modules).
func RequireModule(moduleID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			modules := EnabledModulesFromContext(r.Context())
			for _, m := range modules {
				if m == moduleID {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeJSONError(w, http.StatusForbidden, "module not enabled")
		})
	}
}

// matchPermission checks if a user's permission p satisfies the required
// permission. Supports exact match and wildcard: "erp.*" matches "erp.stock.read".
func matchPermission(userPerm, required string) bool {
	if userPerm == required {
		return true
	}
	// Wildcard: "erp.accounting.*" → prefix "erp.accounting."
	// Also handles bare "*" (superuser wildcard)
	if len(userPerm) >= 1 && userPerm[len(userPerm)-1] == '*' {
		prefix := userPerm[:len(userPerm)-1] // "erp.accounting." or "" for bare "*"
		if len(required) >= len(prefix) && (prefix == "" || required[:len(prefix)] == prefix) {
			return true
		}
	}
	return false
}

const enabledModulesKey contextKey = "enabled_modules"

// WithEnabledModules stores the tenant's enabled module IDs in context.
func WithEnabledModules(ctx context.Context, modules []string) context.Context {
	return context.WithValue(ctx, enabledModulesKey, modules)
}

// EnabledModulesFromContext returns the tenant's enabled module IDs.
func EnabledModulesFromContext(ctx context.Context) []string {
	modules, _ := ctx.Value(enabledModulesKey).([]string)
	return modules
}
