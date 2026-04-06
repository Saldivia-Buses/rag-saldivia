package security

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// InitBlacklist connects to Redis and returns a TokenBlacklist.
// Returns nil if Redis is unavailable (graceful degradation).
// All services should call this on startup for consistent blacklist wiring.
func InitBlacklist(ctx context.Context, redisURL string) *TokenBlacklist {
	if redisURL == "" {
		return nil
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisURL})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("redis not available, token blacklist disabled", "error", err)
		return nil
	}

	slog.Info("token blacklist enabled", "redis", redisURL)
	return NewTokenBlacklist(rdb)
}
