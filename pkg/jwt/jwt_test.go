package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-at-least-32-chars-long!!"

func TestCreateAccess_and_Verify(t *testing.T) {
	cfg := DefaultConfig(testSecret)

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

	got, err := Verify(testSecret, token)
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

func TestCreateRefresh_and_Verify(t *testing.T) {
	cfg := DefaultConfig(testSecret)
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

	got, err := Verify(testSecret, token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	// Refresh token should expire in ~7 days
	expiry := got.ExpiresAt.Time.Sub(got.IssuedAt.Time)
	if expiry < 6*24*time.Hour || expiry > 8*24*time.Hour {
		t.Errorf("expected ~7d expiry, got %v", expiry)
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	cfg := DefaultConfig(testSecret)
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"}

	token, _ := CreateAccess(cfg, claims)

	_, err := Verify("wrong-secret", token)
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	cfg := Config{
		Secret:       testSecret,
		AccessExpiry: -1 * time.Hour, // already expired
		Issuer:       "sda",
	}
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"}

	token, _ := CreateAccess(cfg, claims)

	_, err := Verify(testSecret, token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerify_MissingClaims(t *testing.T) {
	cfg := DefaultConfig(testSecret)

	// Missing TenantID and Slug
	claims := Claims{UserID: "u-1"}
	token, _ := CreateAccess(cfg, claims)

	_, err := Verify(testSecret, token)
	if err == nil {
		t.Fatal("expected ErrMissingClaim for missing tenant info")
	}
}

func TestVerify_InvalidString(t *testing.T) {
	_, err := Verify(testSecret, "not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid token string")
	}
}

func TestCreateAccess_SecretTooShort(t *testing.T) {
	cfg := DefaultConfig("short")
	claims := Claims{UserID: "u-1", TenantID: "t-1", Slug: "x", Role: "user"}

	_, err := CreateAccess(cfg, claims)
	if err != ErrSecretTooShort {
		t.Fatalf("expected ErrSecretTooShort, got %v", err)
	}
}

func TestCreateAccess_SetsSubject(t *testing.T) {
	cfg := DefaultConfig(testSecret)
	claims := Claims{UserID: "u-42", TenantID: "t-1", Slug: "x", Role: "user"}

	token, _ := CreateAccess(cfg, claims)
	got, _ := Verify(testSecret, token)

	if got.Subject != "u-42" {
		t.Errorf("expected Subject 'u-42', got %q", got.Subject)
	}
}
