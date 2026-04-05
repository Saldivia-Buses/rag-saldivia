package security_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Camionerou/rag-saldivia/pkg/security"
	"github.com/redis/go-redis/v9"
)

func newTestBlacklist(t *testing.T) *security.TokenBlacklist {
	t.Helper()
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisURL})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	t.Cleanup(func() { rdb.FlushDB(ctx); rdb.Close() })
	return security.NewTokenBlacklist(rdb)
}

func TestRevoke_And_IsRevoked(t *testing.T) {
	bl := newTestBlacklist(t)
	ctx := context.Background()
	jti := "test-token-123"

	revoked, _ := bl.IsRevoked(ctx, jti)
	if revoked {
		t.Fatal("should not be revoked initially")
	}

	bl.Revoke(ctx, jti, time.Now().Add(1*time.Hour))

	revoked, _ = bl.IsRevoked(ctx, jti)
	if !revoked {
		t.Fatal("should be revoked after Revoke()")
	}
}

func TestRevoke_ExpiredToken_NoOp(t *testing.T) {
	bl := newTestBlacklist(t)
	ctx := context.Background()
	bl.Revoke(ctx, "expired", time.Now().Add(-1*time.Hour))

	revoked, _ := bl.IsRevoked(ctx, "expired")
	if revoked {
		t.Fatal("expired tokens should not be blacklisted")
	}
}

func TestRevokeAll(t *testing.T) {
	bl := newTestBlacklist(t)
	ctx := context.Background()
	jtis := []string{"a", "b", "c"}
	bl.RevokeAll(ctx, jtis, time.Now().Add(1*time.Hour))

	for _, jti := range jtis {
		revoked, _ := bl.IsRevoked(ctx, jti)
		if !revoked {
			t.Fatalf("%s should be revoked", jti)
		}
	}
}
