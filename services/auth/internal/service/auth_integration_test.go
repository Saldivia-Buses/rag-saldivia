//go:build integration

// Integration tests for the auth service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/

package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

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
			action TEXT NOT NULL,
			resource TEXT,
			details JSONB NOT NULL DEFAULT '{}',
			ip_address TEXT,
			user_agent TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
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

	jwtCfg := sdajwt.DefaultConfig("integration-test-secret-32chars!!")
	auth := NewAuth(pool, jwtCfg, "t-test", "test-tenant", nil)

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
	claims, err := sdajwt.Verify("integration-test-secret-32chars!!", tokens.AccessToken)
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

	auth := NewAuth(pool, sdajwt.DefaultConfig("secret-32-characters-long!!!!!!!!"), "t-1", "dev", nil)

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

	auth := NewAuth(pool, sdajwt.DefaultConfig("secret-32-characters-long!!!!!!!!"), "t-1", "dev", nil)

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

func TestLogin_NonexistentUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	auth := NewAuth(pool, sdajwt.DefaultConfig("secret-32-characters-long!!!!!!!!"), "t-1", "dev", nil)

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

	auth := NewAuth(pool, sdajwt.DefaultConfig("secret-32-characters-long!!!!!!!!"), "t-1", "dev", nil)

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

	auth := NewAuth(pool, sdajwt.DefaultConfig("secret-32-characters-long!!!!!!!!"), "t-1", "dev", nil)
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

	auth := NewAuth(pool, sdajwt.DefaultConfig("secret-32-characters-long!!!!!!!!"), "t-1", "dev", nil)
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
