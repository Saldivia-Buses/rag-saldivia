package jwt

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// generateTestKeys creates a fresh Ed25519 keypair for testing.
func generateTestKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate keys: %v", err)
	}
	return pub, priv
}

func testConfig(t *testing.T) Config {
	t.Helper()
	pub, priv := generateTestKeys(t)
	return DefaultConfig(priv, pub)
}

func TestCreateAccess_and_Verify(t *testing.T) {
	cfg := testConfig(t)

	claims := Claims{
		UserID:   "u-123",
		Email:    "admin@saldivia.com",
		Name:     "Admin",
		TenantID: "t-456",
		Slug:     "saldivia",
		Role:     "admin",
	}

	token, err := CreateAccess(cfg, claims)
	if err != nil {
		t.Fatalf("CreateAccess failed: %v", err)
	}

	got, err := Verify(cfg.PublicKey, token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if got.UserID != "u-123" {
		t.Errorf("expected UserID 'u-123', got %q", got.UserID)
	}
	if got.Email != "admin@saldivia.com" {
		t.Errorf("expected Email, got %q", got.Email)
	}
	if got.TenantID != "t-456" {
		t.Errorf("expected TenantID, got %q", got.TenantID)
	}
	if got.Slug != "saldivia" {
		t.Errorf("expected Slug, got %q", got.Slug)
	}
	if got.Role != "admin" {
		t.Errorf("expected Role, got %q", got.Role)
	}
	if got.Issuer != "sda" {
		t.Errorf("expected Issuer 'sda', got %q", got.Issuer)
	}
}

func TestCreateAccess_WithPermissions(t *testing.T) {
	cfg := testConfig(t)

	claims := Claims{
		UserID:      "u-1",
		TenantID:    "t-1",
		Slug:        "test",
		Role:        "user",
		Permissions: []string{"chat.read", "chat.write", "collections.read"},
	}

	token, err := CreateAccess(cfg, claims)
	if err != nil {
		t.Fatalf("CreateAccess failed: %v", err)
	}

	got, err := Verify(cfg.PublicKey, token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(got.Permissions) != 3 {
		t.Fatalf("expected 3 permissions, got %d", len(got.Permissions))
	}
	if got.Permissions[0] != "chat.read" {
		t.Errorf("expected chat.read, got %q", got.Permissions[0])
	}
}

func TestCreateRefresh_and_Verify(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{
		UserID:   "u-1",
		TenantID: "t-1",
		Slug:     "test",
		Role:     "user",
	}

	token, err := CreateRefresh(cfg, claims)
	if err != nil {
		t.Fatalf("CreateRefresh failed: %v", err)
	}

	got, err := Verify(cfg.PublicKey, token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	// Refresh token should expire in ~7 days
	expiry := got.ExpiresAt.Sub(got.IssuedAt.Time)
	if expiry < 6*24*time.Hour || expiry > 8*24*time.Hour {
		t.Errorf("expected ~7d expiry, got %v", expiry)
	}
}

func TestVerify_WrongKey(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"}

	token, _ := CreateAccess(cfg, claims)

	// Verify with a different public key
	otherPub, _ := generateTestKeys(t)
	_, err := Verify(otherPub, token)
	if err == nil {
		t.Fatal("expected error with wrong public key")
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	pub, priv := generateTestKeys(t)
	cfg := Config{
		PrivateKey:   priv,
		PublicKey:    pub,
		AccessExpiry: -1 * time.Hour, // already expired
		Issuer:       "sda",
	}
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"}

	token, _ := CreateAccess(cfg, claims)

	_, err := Verify(pub, token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerify_MissingClaims(t *testing.T) {
	cfg := testConfig(t)

	// Missing TenantID and Slug
	claims := Claims{UserID: "u-1"}
	token, _ := CreateAccess(cfg, claims)

	_, err := Verify(cfg.PublicKey, token)
	if err == nil {
		t.Fatal("expected ErrMissingClaim for missing tenant info")
	}
}

func TestVerify_InvalidString(t *testing.T) {
	pub, _ := generateTestKeys(t)
	_, err := Verify(pub, "not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid token string")
	}
}

func TestCreateAccess_NilPrivateKey(t *testing.T) {
	pub, _ := generateTestKeys(t)
	cfg := VerifyOnlyConfig(pub)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"}

	_, err := CreateAccess(cfg, claims)
	if err == nil {
		t.Fatal("expected error when signing without private key")
	}
}

func TestVerify_NilPublicKey(t *testing.T) {
	_, err := Verify(nil, "some.jwt.token")
	if err == nil {
		t.Fatal("expected error when verifying without public key")
	}
}

func TestCreateAccess_SetsSubject(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{UserID: "u-42", TenantID: "t-1", Slug: "x", Role: "user"}

	token, _ := CreateAccess(cfg, claims)
	got, _ := Verify(cfg.PublicKey, token)

	if got.Subject != "u-42" {
		t.Errorf("expected Subject 'u-42', got %q", got.Subject)
	}
}

func TestParseKeyPEM_Roundtrip(t *testing.T) {
	pub, priv := generateTestKeys(t)

	// Encode private key to PEM
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	// Encode public key to PEM
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	// Parse back
	parsedPriv, err := ParsePrivateKeyPEM(privPEM)
	if err != nil {
		t.Fatalf("ParsePrivateKeyPEM: %v", err)
	}
	parsedPub, err := ParsePublicKeyPEM(pubPEM)
	if err != nil {
		t.Fatalf("ParsePublicKeyPEM: %v", err)
	}

	// Sign with parsed private, verify with parsed public
	cfg := DefaultConfig(parsedPriv, parsedPub)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user"}
	token, err := CreateAccess(cfg, claims)
	if err != nil {
		t.Fatalf("CreateAccess with parsed keys: %v", err)
	}

	got, err := Verify(parsedPub, token)
	if err != nil {
		t.Fatalf("Verify with parsed keys: %v", err)
	}
	if got.UserID != "u-1" {
		t.Errorf("expected u-1, got %q", got.UserID)
	}
}

func TestParseKeyPEM_InvalidData(t *testing.T) {
	_, err := ParsePrivateKeyPEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}

	_, err = ParsePublicKeyPEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}

func TestVerifyOnlyConfig_CannotSign(t *testing.T) {
	pub, _ := generateTestKeys(t)
	cfg := VerifyOnlyConfig(pub)

	_, err := CreateAccess(cfg, Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"})
	if err == nil {
		t.Fatal("VerifyOnlyConfig should not be able to sign tokens")
	}
}

// --- Edge cases: empty required claim fields ---

// TestCreateAccess_EmptyUserID_Fails verifies that a token with an empty UserID
// is rejected at Verify time with ErrMissingClaim. CreateAccess itself doesn't
// validate claims — the guard lives in Verify so all code paths are protected.
func TestCreateAccess_EmptyUserID_Fails(t *testing.T) {
	cfg := testConfig(t)
	token, err := CreateAccess(cfg, Claims{TenantID: "t-1", Slug: "test", Role: "user"})
	if err != nil {
		t.Fatalf("CreateAccess: %v", err)
	}
	_, err = Verify(cfg.PublicKey, token)
	if err == nil {
		t.Fatal("expected error for empty UserID")
	}
	if !isErr(err, ErrMissingClaim) {
		t.Errorf("expected ErrMissingClaim, got %v", err)
	}
}

func TestCreateAccess_EmptyTenantID_Fails(t *testing.T) {
	cfg := testConfig(t)
	token, err := CreateAccess(cfg, Claims{UserID: "u-1", Slug: "test", Role: "user"})
	if err != nil {
		t.Fatalf("CreateAccess: %v", err)
	}
	_, err = Verify(cfg.PublicKey, token)
	if err == nil {
		t.Fatal("expected error for empty TenantID")
	}
	if !isErr(err, ErrMissingClaim) {
		t.Errorf("expected ErrMissingClaim, got %v", err)
	}
}

func TestCreateAccess_EmptySlug_Fails(t *testing.T) {
	cfg := testConfig(t)
	token, err := CreateAccess(cfg, Claims{UserID: "u-1", TenantID: "t-1", Role: "user"})
	if err != nil {
		t.Fatalf("CreateAccess: %v", err)
	}
	_, err = Verify(cfg.PublicKey, token)
	if err == nil {
		t.Fatal("expected error for empty Slug")
	}
	if !isErr(err, ErrMissingClaim) {
		t.Errorf("expected ErrMissingClaim, got %v", err)
	}
}

// TestVerify_NoneAlgorithm_Rejected ensures that a token crafted with alg:"none"
// (a known JWT attack vector) is always rejected, even if the payload is valid.
func TestVerify_NoneAlgorithm_Rejected(t *testing.T) {
	pub, _ := generateTestKeys(t)

	// Craft a "none" algorithm token manually: header.payload.empty-signature
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload, _ := json.Marshal(map[string]any{
		"uid":  "u-1",
		"tid":  "t-1",
		"slug": "test",
		"role": "user",
		"sub":  "u-1",
		"iss":  "sda",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	noneToken := header + "." + payloadB64 + "."

	_, err := Verify(pub, noneToken)
	if err == nil {
		t.Fatal("alg:none token must be rejected")
	}
}

// TestVerify_HS256Algorithm_Rejected ensures HMAC-signed tokens are not accepted.
// The middleware only accepts EdDSA (Ed25519). An attacker might craft an HS256
// token using the public key as the HMAC secret (algorithm confusion attack).
func TestVerify_HS256Algorithm_Rejected(t *testing.T) {
	pub, _ := generateTestKeys(t)

	// We can't easily sign with HS256 here without a shared secret, but we can
	// craft a well-formed HS256 header with a dummy signature and verify rejection.
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(map[string]any{
		"uid":  "u-1",
		"tid":  "t-1",
		"slug": "test",
		"role": "user",
		"sub":  "u-1",
		"iss":  "sda",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	fakeSig := base64.RawURLEncoding.EncodeToString([]byte("fake-hmac-signature"))
	hs256Token := header + "." + payloadB64 + "." + fakeSig

	_, err := Verify(pub, hs256Token)
	if err == nil {
		t.Fatal("HS256 token must be rejected — only EdDSA accepted")
	}
}

// TestVerify_TamperedPayload_Rejected verifies that flipping a byte in the payload
// section of a valid token causes signature validation failure.
func TestVerify_TamperedPayload_Rejected(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user"}
	token, err := CreateAccess(cfg, claims)
	if err != nil {
		t.Fatalf("CreateAccess: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token structure: %q", token)
	}

	// Decode, flip a byte, re-encode
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	payloadBytes[len(payloadBytes)/2] ^= 0xFF // flip bits in the middle
	parts[1] = base64.RawURLEncoding.EncodeToString(payloadBytes)
	tamperedToken := strings.Join(parts, ".")

	_, err = Verify(cfg.PublicKey, tamperedToken)
	if err == nil {
		t.Fatal("tampered payload must be rejected")
	}
}

// TestVerify_ExpiredToken_ReturnsErrInvalidToken confirms the sentinel error value
// for expired tokens — callers should handle ErrInvalidToken specifically.
func TestVerify_ExpiredToken_ReturnsErrInvalidToken(t *testing.T) {
	pub, priv := generateTestKeys(t)
	cfg := Config{
		PrivateKey:   priv,
		PublicKey:    pub,
		AccessExpiry: -1 * time.Hour,
		Issuer:       "sda",
	}
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user"}
	token, _ := CreateAccess(cfg, claims)

	_, err := Verify(pub, token)
	if !isErr(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

// TestVerify_WrongKey_ReturnsError already tested above but this variant
// explicitly checks error is ErrInvalidToken (not some other error).
func TestVerify_WrongKey_ReturnsErrInvalidToken(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user"}
	token, _ := CreateAccess(cfg, claims)

	otherPub, _ := generateTestKeys(t)
	_, err := Verify(otherPub, token)
	if !isErr(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for wrong key, got %v", err)
	}
}

// TestVerify_ValidToken_ExtractsAllClaims is a full happy-path test verifying
// every claim field is preserved through sign → verify.
func TestVerify_ValidToken_ExtractsAllClaims(t *testing.T) {
	cfg := testConfig(t)
	in := Claims{
		UserID:      "u-abc",
		Email:       "user@example.com",
		Name:        "Test User",
		TenantID:    "t-xyz",
		Slug:        "example",
		Role:        "manager",
		Permissions: []string{"erp.read", "chat.write"},
	}

	token, err := CreateAccess(cfg, in)
	if err != nil {
		t.Fatalf("CreateAccess: %v", err)
	}

	got, err := Verify(cfg.PublicKey, token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if got.UserID != in.UserID {
		t.Errorf("UserID: got %q, want %q", got.UserID, in.UserID)
	}
	if got.Email != in.Email {
		t.Errorf("Email: got %q, want %q", got.Email, in.Email)
	}
	if got.Name != in.Name {
		t.Errorf("Name: got %q, want %q", got.Name, in.Name)
	}
	if got.TenantID != in.TenantID {
		t.Errorf("TenantID: got %q, want %q", got.TenantID, in.TenantID)
	}
	if got.Slug != in.Slug {
		t.Errorf("Slug: got %q, want %q", got.Slug, in.Slug)
	}
	if got.Role != in.Role {
		t.Errorf("Role: got %q, want %q", got.Role, in.Role)
	}
	if len(got.Permissions) != 2 || got.Permissions[0] != "erp.read" {
		t.Errorf("Permissions: got %v, want [erp.read chat.write]", got.Permissions)
	}
	if got.Issuer != "sda" {
		t.Errorf("Issuer: got %q, want sda", got.Issuer)
	}
	if got.Subject != in.UserID {
		t.Errorf("Subject: got %q, want %q", got.Subject, in.UserID)
	}
	if got.ID == "" {
		t.Error("JTI (ID) must be auto-generated and non-empty")
	}
}

// TestCreateRefresh_DifferentExpiryFromAccess verifies that refresh tokens have
// a meaningfully longer lifetime than access tokens, caught at the Claims level.
func TestCreateRefresh_DifferentExpiryFromAccess(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "test", Role: "user"}

	accessToken, err := CreateAccess(cfg, claims)
	if err != nil {
		t.Fatalf("CreateAccess: %v", err)
	}
	refreshToken, err := CreateRefresh(cfg, claims)
	if err != nil {
		t.Fatalf("CreateRefresh: %v", err)
	}

	accessClaims, err := Verify(cfg.PublicKey, accessToken)
	if err != nil {
		t.Fatalf("Verify access: %v", err)
	}
	refreshClaims, err := Verify(cfg.PublicKey, refreshToken)
	if err != nil {
		t.Fatalf("Verify refresh: %v", err)
	}

	accessExpiry := accessClaims.ExpiresAt.Sub(accessClaims.IssuedAt.Time)
	refreshExpiry := refreshClaims.ExpiresAt.Sub(refreshClaims.IssuedAt.Time)

	if refreshExpiry <= accessExpiry {
		t.Errorf("refresh expiry (%v) must be longer than access expiry (%v)", refreshExpiry, accessExpiry)
	}
	// Access token should be ~15min
	if accessExpiry < 10*time.Minute || accessExpiry > 20*time.Minute {
		t.Errorf("access expiry %v not in expected range [10m, 20m]", accessExpiry)
	}
	// Refresh token should be ~7d
	if refreshExpiry < 6*24*time.Hour || refreshExpiry > 8*24*time.Hour {
		t.Errorf("refresh expiry %v not in expected range [6d, 8d]", refreshExpiry)
	}
}

// TestVerify_RefreshToken_ValidatesCorrectly checks that a refresh token passes
// verification and all identity claims survive the round-trip.
func TestVerify_RefreshToken_ValidatesCorrectly(t *testing.T) {
	cfg := testConfig(t)
	claims := Claims{UserID: "u-99", TenantID: "t-99", Slug: "refresh-test", Role: "user"}

	token, err := CreateRefresh(cfg, claims)
	if err != nil {
		t.Fatalf("CreateRefresh: %v", err)
	}

	got, err := Verify(cfg.PublicKey, token)
	if err != nil {
		t.Fatalf("Verify refresh token: %v", err)
	}
	if got.UserID != "u-99" {
		t.Errorf("UserID: got %q, want u-99", got.UserID)
	}
	if got.TenantID != "t-99" {
		t.Errorf("TenantID: got %q, want t-99", got.TenantID)
	}
	if got.Slug != "refresh-test" {
		t.Errorf("Slug: got %q, want refresh-test", got.Slug)
	}
	if got.ID == "" {
		t.Error("JTI must be non-empty on refresh token")
	}
}

// isErr wraps errors.Is for readable table-driven assertions.
func isErr(err, target error) bool {
	return errors.Is(err, target)
}

// encodeTestKeysPEM returns (public PEM bytes, private PEM bytes) for testing
// the env-file loading paths. Matches the encoding TestParseKeyPEM_Roundtrip uses.
func encodeTestKeysPEM(t *testing.T) (pubPEM, privPEM []byte) {
	t.Helper()
	pub, priv := generateTestKeys(t)
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	return pubPEM, privPEM
}

func TestMustLoadPublicKey_EnvVarMode(t *testing.T) {
	pubPEM, _ := encodeTestKeysPEM(t)
	b64 := base64.StdEncoding.EncodeToString(pubPEM)
	t.Setenv("TEST_PUB_KEY", b64)
	t.Setenv("TEST_PUB_KEY_FILE", "")
	got := MustLoadPublicKey("TEST_PUB_KEY")
	if len(got) != ed25519.PublicKeySize {
		t.Fatalf("expected Ed25519 public key of size %d, got %d", ed25519.PublicKeySize, len(got))
	}
}

func TestMustLoadPublicKey_FileMode(t *testing.T) {
	pubPEM, _ := encodeTestKeysPEM(t)
	dir := t.TempDir()
	path := dir + "/pub.pem"
	if err := os.WriteFile(path, pubPEM, 0o600); err != nil {
		t.Fatalf("write pub pem: %v", err)
	}
	t.Setenv("TEST_PUB_KEY", "")
	t.Setenv("TEST_PUB_KEY_FILE", path)
	got := MustLoadPublicKey("TEST_PUB_KEY")
	if len(got) != ed25519.PublicKeySize {
		t.Fatalf("expected Ed25519 public key of size %d, got %d", ed25519.PublicKeySize, len(got))
	}
}

func TestMustLoadPublicKey_EnvPreferredOverFile(t *testing.T) {
	pubPEM, _ := encodeTestKeysPEM(t)
	b64 := base64.StdEncoding.EncodeToString(pubPEM)
	t.Setenv("TEST_PUB_KEY", b64)
	// _FILE points at /dev/null — if env-first logic regresses we'd get a PEM parse panic.
	t.Setenv("TEST_PUB_KEY_FILE", "/dev/null")
	got := MustLoadPublicKey("TEST_PUB_KEY")
	if len(got) != ed25519.PublicKeySize {
		t.Fatalf("expected Ed25519 public key of size %d, got %d", ed25519.PublicKeySize, len(got))
	}
}

func TestMustLoadPublicKey_NeitherSet_Panics(t *testing.T) {
	t.Setenv("TEST_PUB_KEY", "")
	t.Setenv("TEST_PUB_KEY_FILE", "")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when neither env var is set")
		}
	}()
	_ = MustLoadPublicKey("TEST_PUB_KEY")
}

func TestMustLoadPrivateKey_EnvVarMode(t *testing.T) {
	_, privPEM := encodeTestKeysPEM(t)
	b64 := base64.StdEncoding.EncodeToString(privPEM)
	t.Setenv("TEST_PRIV_KEY", b64)
	t.Setenv("TEST_PRIV_KEY_FILE", "")
	got := MustLoadPrivateKey("TEST_PRIV_KEY")
	if len(got) != ed25519.PrivateKeySize {
		t.Fatalf("expected Ed25519 private key of size %d, got %d", ed25519.PrivateKeySize, len(got))
	}
}

func TestMustLoadPrivateKey_FileMode(t *testing.T) {
	_, privPEM := encodeTestKeysPEM(t)
	dir := t.TempDir()
	path := dir + "/priv.pem"
	if err := os.WriteFile(path, privPEM, 0o600); err != nil {
		t.Fatalf("write priv pem: %v", err)
	}
	t.Setenv("TEST_PRIV_KEY", "")
	t.Setenv("TEST_PRIV_KEY_FILE", path)
	got := MustLoadPrivateKey("TEST_PRIV_KEY")
	if len(got) != ed25519.PrivateKeySize {
		t.Fatalf("expected Ed25519 private key of size %d, got %d", ed25519.PrivateKeySize, len(got))
	}
}

func TestMustLoadPrivateKey_NeitherSet_Panics(t *testing.T) {
	t.Setenv("TEST_PRIV_KEY", "")
	t.Setenv("TEST_PRIV_KEY_FILE", "")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when neither env var is set")
		}
	}()
	_ = MustLoadPrivateKey("TEST_PRIV_KEY")
}
