// Package tenant provides multi-tenant context propagation and database resolution.
//
// Every request in SDA carries a tenant context extracted from the JWT claim
// and/or the X-Tenant-ID header (injected by Traefik from subdomain).
// This package provides helpers to store/retrieve tenant info from context.Context
// and resolve the correct PostgreSQL + Redis connections for a given tenant.
package tenant

import (
	"context"
	"errors"
)

// Info holds the tenant identification for a request.
type Info struct {
	ID   string // tenant UUID from platform DB
	Slug string // subdomain slug (e.g., "saldivia")
}

type ctxKey struct{}

var (
	ErrNoTenant      = errors.New("no tenant in context")
	ErrTenantUnknown = errors.New("unknown tenant")
)

// WithInfo stores tenant info in the context.
func WithInfo(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, ctxKey{}, info)
}

// FromContext extracts tenant info from the context.
// Returns ErrNoTenant if the context has no tenant.
func FromContext(ctx context.Context) (Info, error) {
	info, ok := ctx.Value(ctxKey{}).(Info)
	if !ok {
		return Info{}, ErrNoTenant
	}
	return info, nil
}

// SlugFromContext is a convenience function that returns just the tenant slug.
// Panics if there is no tenant in context — use only in code paths where
// tenant is guaranteed (e.g., after middleware).
func SlugFromContext(ctx context.Context) string {
	info, err := FromContext(ctx)
	if err != nil {
		panic("tenant.SlugFromContext called without tenant in context")
	}
	return info.Slug
}
