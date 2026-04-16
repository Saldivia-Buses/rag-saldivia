//go:build integration

// MFA integration tests for the auth service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v ./internal/service/ -timeout 120s

package service

import (
	"context"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
)

// TestSetupMFA_GeneratesSecretAndURI verifies that SetupMFA returns a valid
// base32 TOTP secret and a well-formed otpauth:// URI.
func TestSetupMFA_GeneratesSecretAndURI(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-setup@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)
	require.NotEmpty(t, result.Secret, "expected non-empty TOTP secret")
	require.NotEmpty(t, result.URI, "expected non-empty otpauth:// URI")

	// URI must start with otpauth://totp/
	require.True(t, len(result.URI) > 0 && result.URI[:15] == "otpauth://totp/",
		"URI must start with otpauth://totp/, got: %s", result.URI)

	// Secret must be a valid TOTP secret — validate by generating a code
	code, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)
	require.Len(t, code, 6, "TOTP code must be 6 digits")
}

// TestSetupMFA_StoredSecret_IsValidForTOTP verifies that the secret stored in DB
// during SetupMFA can produce a valid code that passes VerifyMFA.
func TestSetupMFA_StoredSecret_IsValidForTOTP(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-stored@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)

	// Generate a valid TOTP code from the returned secret
	code, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)

	// VerifyMFA must accept this code (reads secret from DB)
	err = auth.VerifyMFA(context.Background(), userID, code)
	require.NoError(t, err)
}

// TestVerifySetup_ValidTOTP_EnablesMFA verifies that calling VerifySetup with a
// valid TOTP code sets mfa_enabled = true in the database.
func TestVerifySetup_ValidTOTP_EnablesMFA(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-verify@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	// Setup first to get a secret stored in DB
	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)

	// Generate valid code from secret
	code, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)

	// VerifySetup must enable MFA
	err = auth.VerifySetup(context.Background(), userID, code)
	require.NoError(t, err)

	// Verify DB: mfa_enabled must be true
	var mfaEnabled bool
	err = pool.QueryRow(context.Background(),
		`SELECT mfa_enabled FROM users WHERE id = $1`, userID,
	).Scan(&mfaEnabled)
	require.NoError(t, err)
	require.True(t, mfaEnabled, "mfa_enabled must be true after VerifySetup")
}

// TestVerifySetup_InvalidCode_ReturnsError verifies that VerifySetup with an
// invalid code returns ErrInvalidMFACode and does NOT enable MFA.
func TestVerifySetup_InvalidCode_ReturnsError(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-invalid@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	// Setup to store a secret
	_, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)

	// Use a clearly wrong code
	err = auth.VerifySetup(context.Background(), userID, "000000")
	require.ErrorIs(t, err, ErrInvalidMFACode)

	// mfa_enabled must still be false in DB
	var mfaEnabled bool
	err = pool.QueryRow(context.Background(),
		`SELECT mfa_enabled FROM users WHERE id = $1`, userID,
	).Scan(&mfaEnabled)
	require.NoError(t, err)
	require.False(t, mfaEnabled, "mfa_enabled must remain false after invalid code")
}

// TestDisableMFA_WithValidTOTP_DisablesMFA verifies that DisableMFA with a valid
// TOTP code disables MFA and clears the secret from the database.
// Note: DisableMFA(ctx, userID, code) takes a TOTP code, NOT a password.
// The production flow requires re-verifying possession of the authenticator
// before disabling MFA (same security pattern as "confirm via 2FA to disable 2FA").
func TestDisableMFA_WithValidTOTP_DisablesMFA(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-disable@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	// Enable MFA first
	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)
	code, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)
	err = auth.VerifySetup(context.Background(), userID, code)
	require.NoError(t, err)

	// Generate a fresh code to disable MFA
	// (use the same second to avoid TOTP clock window issues)
	code, err = totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)

	err = auth.DisableMFA(context.Background(), userID, code)
	require.NoError(t, err)

	// DB: mfa_enabled must be false, mfa_secret must be NULL
	var mfaEnabled bool
	var mfaSecret *string
	err = pool.QueryRow(context.Background(),
		`SELECT mfa_enabled, mfa_secret FROM users WHERE id = $1`, userID,
	).Scan(&mfaEnabled, &mfaSecret)
	require.NoError(t, err)
	require.False(t, mfaEnabled, "mfa_enabled must be false after DisableMFA")
	require.Nil(t, mfaSecret, "mfa_secret must be NULL after DisableMFA")
}

// TestDisableMFA_WrongCode_ReturnsError verifies that an invalid TOTP code
// is rejected by DisableMFA.
func TestDisableMFA_WrongCode_ReturnsError(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-wrong@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	// Enable MFA
	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)
	code, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)
	err = auth.VerifySetup(context.Background(), userID, code)
	require.NoError(t, err)

	// Try to disable with wrong code
	err = auth.DisableMFA(context.Background(), userID, "000000")
	require.ErrorIs(t, err, ErrInvalidMFACode)
}

// TestCompleteMFALogin_ValidCode_ReturnsTokens verifies the full MFA login flow:
// 1. Login returns MFARequired=true + mfa_token
// 2. CompleteMFALogin with a valid TOTP code returns real access + refresh tokens
func TestCompleteMFALogin_ValidCode_ReturnsTokens(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-complete@test.com", "password123", "role-user")
	jwtCfg := testJWTCfg(t)
	auth := NewAuth(pool, jwtCfg, "t-mfa", "mfa-tenant")

	// Enable MFA for this user
	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)
	setupCode, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)
	err = auth.VerifySetup(context.Background(), userID, setupCode)
	require.NoError(t, err)

	// Login — must return MFARequired=true
	tokens, err := auth.Login(context.Background(), LoginRequest{
		Email:    "mfa-complete@test.com",
		Password: "password123",
		IP:       "127.0.0.1",
	})
	require.NoError(t, err)
	require.True(t, tokens.MFARequired, "expected MFARequired=true for MFA-enabled user")
	require.NotEmpty(t, tokens.MFAToken, "expected non-empty MFA token")
	require.Empty(t, tokens.AccessToken, "access token must be empty before MFA completion")

	// Complete MFA with a fresh code
	loginCode, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)
	finalTokens, err := auth.CompleteMFALogin(context.Background(), tokens.MFAToken, loginCode)
	require.NoError(t, err)
	require.NotEmpty(t, finalTokens.AccessToken, "expected non-empty access token after MFA")
	require.NotEmpty(t, finalTokens.RefreshToken, "expected non-empty refresh token after MFA")

	// Access token must carry correct claims
	claims, err := sdajwt.Verify(jwtCfg.PublicKey, finalTokens.AccessToken)
	require.NoError(t, err)
	require.Equal(t, "mfa-complete@test.com", claims.Email)
	require.Equal(t, "user", claims.Role, "real role, not mfa_pending")
}

// TestCompleteMFALogin_InvalidCode_ReturnsError verifies that an invalid TOTP
// code is rejected by CompleteMFALogin.
func TestCompleteMFALogin_InvalidCode_ReturnsError(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userID := seedTestUser(t, pool, "mfa-badcode@test.com", "password123", "role-user")
	auth := NewAuth(pool, testJWTCfg(t), "t-mfa", "mfa-tenant")

	// Enable MFA
	result, err := auth.SetupMFA(context.Background(), userID)
	require.NoError(t, err)
	setupCode, err := totp.GenerateCode(result.Secret, time.Now())
	require.NoError(t, err)
	err = auth.VerifySetup(context.Background(), userID, setupCode)
	require.NoError(t, err)

	// Login to get mfa_token
	tokens, err := auth.Login(context.Background(), LoginRequest{
		Email: "mfa-badcode@test.com", Password: "password123",
	})
	require.NoError(t, err)
	require.True(t, tokens.MFARequired)

	// Complete MFA with wrong code
	_, err = auth.CompleteMFALogin(context.Background(), tokens.MFAToken, "000000")
	require.ErrorIs(t, err, ErrInvalidMFACode)
}
