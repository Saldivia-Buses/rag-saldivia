package security

import (
	"context"
	"log/slog"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/redis/go-redis/v9"
)

// InitBlacklist connects to Redis using the provided options and returns a
// TokenBlacklist. Returns nil if opts is nil, Addr is unset, or Redis is
// unreachable (graceful degradation). Callers build the full *redis.Options
// so auth / TLS / DB / retry settings stay co-located at the wire-up site.
func InitBlacklist(ctx context.Context, opts *redis.Options) *TokenBlacklist {
	if opts == nil || opts.Addr == "" {
		return nil
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("redis not available, token blacklist disabled", "error", err)
		return nil
	}

	slog.Info("token blacklist enabled", "redis", config.RedactURL(opts.Addr))
	return NewTokenBlacklist(rdb)
}
