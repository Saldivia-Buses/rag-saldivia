// Package security provides security primitives for SDA services.
// TokenBlacklist stores revoked JWT token IDs in Redis for logout/revocation.
package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist stores revoked JWT token IDs in Redis.
// When a user logs out, the token's JTI is added with a TTL matching
// the token's remaining lifetime. Auth middleware checks IsRevoked()
// on every request to reject revoked tokens.
type TokenBlacklist struct {
	rdb    *redis.Client
	prefix string
}

// NewTokenBlacklist creates a blacklist backed by Redis.
func NewTokenBlacklist(rdb *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{
		rdb:    rdb,
		prefix: "sda:token:blacklist:",
	}
}

// Revoke adds a token ID to the blacklist. Expires when the token would
// have expired naturally, so the blacklist doesn't grow forever.
func (b *TokenBlacklist) Revoke(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return b.rdb.Set(ctx, b.prefix+jti, "1", ttl).Err()
}

// IsRevoked checks if a token ID has been blacklisted.
func (b *TokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	result, err := b.rdb.Exists(ctx, b.prefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return result > 0, nil
}

// RevokeAll blacklists multiple token IDs (e.g., on password change).
func (b *TokenBlacklist) RevokeAll(ctx context.Context, jtis []string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	pipe := b.rdb.Pipeline()
	for _, jti := range jtis {
		pipe.Set(ctx, b.prefix+jti, "1", ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}
