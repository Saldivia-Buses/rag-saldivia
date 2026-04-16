//go:build integration

// Integration tests for the auth service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/

package service

import (
	"context"
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

// testKeys generates a fresh Ed25519 keypair for each test run.
// Integration tests use real asymmetric signing (same as production).
func testKeys(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate test keys: %v", err)
	}
	return priv, pub
}

// testJWTCfg returns a JWT config with fresh ephemeral keys for testing.
func testJWTCfg(t *testing.T) sdajwt.Config {
	t.Helper()
	priv, pub := testKeys(t)
	return sdajwt.DefaultConfig(priv, pub)
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("sda_test"),
		postgres.WithUsername("sda"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	// Apply auth migration
	migration := `
		CREATE TABLE users (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			avatar_url TEXT,
			mfa_secret TEXT,
			mfa_enabled BOOLEAN NOT NULL DEFAULT false,
			is_active BOOLEAN NOT NULL DEFAULT true,
			failed_logins INTEGER NOT NULL DEFAULT 0,
			locked_until TIMESTAMPTZ,
			last_login_at TIMESTAMPTZ,
			last_login_ip TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE roles (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			is_system BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE user_roles (
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			PRIMARY KEY (user_id, role_id)
		);
		CREATE TABLE refresh_tokens (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			revoked_at TIMESTAMPTZ
		);
		CREATE TABLE audit_log (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			user_id TEXT REFERENCES users(id),
			tenant_id TEXT,
			action TEXT NOT NULL,
			resource TEXT,
			details JSONB NOT NULL DEFAULT '{}',
			ip_address TEXT,
			user_agent TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE permissions (
			id TEXT PRIMARY KEY,
			description TEXT
		);
		CREATE TABLE role_permissions (
			role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		);
		INSERT INTO roles (id, name, is_system) VALUES ('role-admin', 'admin', true);
		INSERT INTO roles (id, name, is_system) VALUES ('role-user', 'user', true);
	`
	_, err = pool.Exec(ctx, migration)
	if err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return pool, cleanup
}

func seedTestUser(t *testing.T, pool *pgxpool.Pool, email, password, roleID string) string {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	err = pool.QueryRow(context.Background(),
		`INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		email, "Test User", hash,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	_, err = pool.Exec(context.Background(),
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
		userID, roleID,
	)
	if err != nil {
		t.Fatalf("seed user role: %v", err)
	}

	return userID
}

func TestLogin_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "admin@test.com", "correctpassword", "role-admin")

	jwtCfg := testJWTCfg(t)
	auth := NewAuth(pool, jwtCfg, "t-test", "test-tenant")

	tokens, err := auth.Login(context.Background(), LoginRequest{
		Email:    "admin@test.com",
		Password: "correctpassword",
		IP:       "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if tokens.ExpiresIn != 900 { // 15 min = 900s
		t.Errorf("expected ExpiresIn=900, got %d", tokens.ExpiresIn)
	}

	// Verify the access token contains correct claims
	claims, err := sdajwt.Verify(jwtCfg.PublicKey, tokens.AccessToken)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Email != "admin@test.com" {
		t.Errorf("expected email admin@test.com, got %q", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %q", claims.Role)
	}
	if claims.Slug != "test-tenant" {
		t.Errorf("expected slug test-tenant, got %q", claims.Slug)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "user@test.com", "realpassword", "role-user")

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")

	_, err := auth.Login(context.Background(), LoginRequest{
		Email:    "user@test.com",
		Password: "wrongpassword",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}

	// Verify failed_logins was incremented
	var failedLogins int
	pool.QueryRow(context.Background(),
		`SELECT failed_logins FROM users WHERE email = $1`, "user@test.com",
	).Scan(&failedLogins)
	if failedLogins != 1 {
		t.Errorf("expected failed_logins=1, got %d", failedLogins)
	}
}

func TestLogin_BruteForce_Lockout(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "victim@test.com", "secret123", "role-user")

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")

	// 5 failed attempts
	for i := 0; i < 5; i++ {
		auth.Login(context.Background(), LoginRequest{
			Email: "victim@test.com", Password: "wrong",
		})
	}

	// 6th attempt — even with correct password — should be locked
	_, err := auth.Login(context.Background(), LoginRequest{
		Email:    "victim@test.com",
		Password: "secret123",
	})
	if err != ErrAccountLocked {
		t.Fatalf("expected ErrAccountLocked after 5 failures, got: %v", err)
	}
}

func TestLogin_AfterLockout_CorrectPassword_StillFails(t *testing.T) {
	// Separate from the lockout test above: explicitly verify that even the
	// correct password is rejected while the account is in temporary lockout.
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "locked@test.com", "realpassword", "role-user")

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")

	for i := 0; i < 5; i++ {
		auth.Login(context.Background(), LoginRequest{
			Email: "locked@test.com", Password: "wrong",
		})
	}

	// Correct password but account is locked
	_, err := auth.Login(context.Background(), LoginRequest{
		Email:    "locked@test.com",
		Password: "realpassword",
	})
	if err != ErrAccountLocked {
		t.Fatalf("expected ErrAccountLocked with correct password while locked, got: %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")

	_, err := auth.Login(context.Background(), LoginRequest{
		Email: "nobody@test.com", Password: "anything",
	})
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_DisabledUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "disabled@test.com", "pass123", "role-user")
	pool.Exec(context.Background(),
		`UPDATE users SET is_active = false WHERE email = $1`, "disabled@test.com",
	)

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")

	_, err := auth.Login(context.Background(), LoginRequest{
		Email: "disabled@test.com", Password: "pass123",
	})
	// Disabled users get the same error as invalid credentials
	// to prevent information leakage about account state
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for disabled user, got: %v", err)
	}
}

func TestLogin_AuditLog(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "audited@test.com", "password", "role-admin")

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")
	auth.Login(context.Background(), LoginRequest{
		Email: "audited@test.com", Password: "password", IP: "10.0.0.1",
	})

	var count int
	pool.QueryRow(context.Background(),
		`SELECT count(*) FROM audit_log WHERE action = 'user.login' AND ip_address = '10.0.0.1'`,
	).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 audit log entry, got %d", count)
	}
}

func TestLogin_RefreshTokenStored(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "refresh@test.com", "password", "role-user")

	auth := NewAuth(pool, testJWTCfg(t), "t-1", "dev")
	auth.Login(context.Background(), LoginRequest{
		Email: "refresh@test.com", Password: "password",
	})

	var count int
	pool.QueryRow(context.Background(),
		`SELECT count(*) FROM refresh_tokens WHERE user_id = (SELECT id FROM users WHERE email = 'refresh@test.com')`,
	).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 refresh token stored, got %d", count)
	}
}

// TestRefreshRotation_OldTokenInvalidated verifies the token rotation invariant:
// after a successful refresh, the old refresh token is revoked and cannot be reused.
func TestRefreshRotation_OldTokenInvalidated(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "rotate@test.com", "password", "role-user")

	jwtCfg := testJWTCfg(t)
	auth := NewAuth(pool, jwtCfg, "t-1", "dev")

	// Initial login — get first refresh token
	tokens1, err := auth.Login(context.Background(), LoginRequest{
		Email: "rotate@test.com", Password: "password",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	firstRefresh := tokens1.RefreshToken

	// Use first refresh token — get new pair
	tokens2, err := auth.Refresh(context.Background(), firstRefresh)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if tokens2.RefreshToken == firstRefresh {
		t.Error("expected new refresh token after rotation, got the same token")
	}

	// Attempt to reuse the original refresh token — must fail
	_, err = auth.Refresh(context.Background(), firstRefresh)
	if err != ErrInvalidRefreshToken {
		t.Fatalf("expected ErrInvalidRefreshToken reusing old token after rotation, got: %v", err)
	}
}

// TestLogin_MultipleActiveSessions_AllValid checks that two separate logins
// produce two independent valid token pairs.
// Note: the current implementation revokes old refresh tokens on each new login
// (RevokeUserRefreshTokens), so only the latest session's refresh token is valid.
// This test verifies both access tokens are independently valid (access tokens
// are not revoked on new login — only at explicit logout or expiry).
func TestLogin_MultipleActiveSessions_AllValid(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "multisession@test.com", "password", "role-user")

	jwtCfg := testJWTCfg(t)
	auth := NewAuth(pool, jwtCfg, "t-1", "dev")

	// First login
	tokens1, err := auth.Login(context.Background(), LoginRequest{
		Email: "multisession@test.com", Password: "password",
	})
	if err != nil {
		t.Fatalf("first login: %v", err)
	}

	// Second login (e.g., from a different device)
	tokens2, err := auth.Login(context.Background(), LoginRequest{
		Email: "multisession@test.com", Password: "password",
	})
	if err != nil {
		t.Fatalf("second login: %v", err)
	}

	// Both access tokens must be independently verifiable
	claims1, err := sdajwt.Verify(jwtCfg.PublicKey, tokens1.AccessToken)
	if err != nil {
		t.Fatalf("verify first access token: %v", err)
	}
	claims2, err := sdajwt.Verify(jwtCfg.PublicKey, tokens2.AccessToken)
	if err != nil {
		t.Fatalf("verify second access token: %v", err)
	}

	// Both must have the same identity claims
	if claims1.Email != "multisession@test.com" {
		t.Errorf("session1 email: %q", claims1.Email)
	}
	if claims2.Email != "multisession@test.com" {
		t.Errorf("session2 email: %q", claims2.Email)
	}

	// Each login must produce a distinct access token (different JTIs)
	if claims1.ID == claims2.ID {
		t.Error("expected distinct JTI per session, got duplicate")
	}
}

// TestTenantIsolation_TokenSlugMatchesService verifies that the tenant slug
// embedded in JWT claims matches the service's configured slug.
// A token issued for tenant A cannot be used as a valid token for tenant B
// because Verify would be called with tenant B's public key in a real deployment,
// but here we test that the slug claim is correctly set.
func TestTenantIsolation_TokenSlugMatchesService(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	seedTestUser(t, pool, "tenantuser@test.com", "password", "role-user")

	jwtCfg := testJWTCfg(t)

	// Service configured for tenant A
	tenantAAuth := NewAuth(pool, jwtCfg, "tenant-a-id", "tenant-a")
	tokens, err := tenantAAuth.Login(context.Background(), LoginRequest{
		Email: "tenantuser@test.com", Password: "password",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	// The token's slug must be tenant A's slug, not tenant B
	claims, err := sdajwt.Verify(jwtCfg.PublicKey, tokens.AccessToken)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Slug != "tenant-a" {
		t.Errorf("expected slug tenant-a, got %q", claims.Slug)
	}
	if claims.TenantID != "tenant-a-id" {
		t.Errorf("expected tenant_id tenant-a-id, got %q", claims.TenantID)
	}

	// Attempting to use this token for a service configured for tenant B:
	// The middleware enforces JWT slug == X-Tenant-Slug (Traefik-injected).
	// Simulate: tenant B's auth service would be NewAuth(pool, jwtCfg, "tenant-b-id", "tenant-b").
	// If a request comes in with X-Tenant-Slug: tenant-b but JWT.slug == tenant-a,
	// the middleware returns 403 (tested in pkg/middleware/auth_test.go).
	// At the service layer, a direct slug mismatch would be a 403 before hitting service code.
	// Here we verify the invariant: the service encodes its own slug, not an arbitrary value.
	if claims.Slug == "tenant-b" {
		t.Error("tenant isolation violated: token slug should be tenant-a, not tenant-b")
	}
}
