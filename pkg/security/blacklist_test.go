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

	t.Cleanup(func() {
		rdb.FlushDB(ctx)
		rdb.Close()
	})

	return security.NewTokenBlacklist(rdb)
}

func TestRevoke_And_IsRevoked(t *testing.T) {
	bl := newTestBlacklist(t)
	ctx := context.Background()

	jti := "test-token-123"
	expires := time.Now().Add(1 * time.Hour)

	// Not revoked initially
	revoked, err := bl.IsRevoked(ctx, jti)
	if err != nil {
		t.Fatal(err)
	}
	if revoked {
		t.Fatal("should not be revoked initially")
	}

	// Revoke
	if err := bl.Revoke(ctx, jti, expires); err != nil {
		t.Fatal(err)
	}

	// Now revoked
	revoked, err = bl.IsRevoked(ctx, jti)
	if err != nil {
		t.Fatal(err)
	}
	if !revoked {
		t.Fatal("should be revoked after Revoke()")
	}
}

func TestRevoke_ExpiredToken_NoOp(t *testing.T) {
	bl := newTestBlacklist(t)
	ctx := context.Background()

	// Already expired — should not be stored
	err := bl.Revoke(ctx, "expired-token", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	revoked, _ := bl.IsRevoked(ctx, "expired-token")
	if revoked {
		t.Fatal("expired tokens should not be blacklisted")
	}
}

func TestRevokeAll(t *testing.T) {
	bl := newTestBlacklist(t)
	ctx := context.Background()

	jtis := []string{"token-a", "token-b", "token-c"}
	expires := time.Now().Add(1 * time.Hour)

	if err := bl.RevokeAll(ctx, jtis, expires); err != nil {
		t.Fatal(err)
	}

	for _, jti := range jtis {
		revoked, err := bl.IsRevoked(ctx, jti)
		if err != nil {
			t.Fatal(err)
		}
		if !revoked {
			t.Fatalf("%s should be revoked", jti)
		}
	}
}
