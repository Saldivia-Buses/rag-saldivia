package jwt

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
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
	expiry := got.ExpiresAt.Time.Sub(got.IssuedAt.Time)
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
