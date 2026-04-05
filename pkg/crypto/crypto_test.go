package crypto

import (
	"crypto/rand"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	key := testKey(t)
	plaintext := "postgres://user:pass@host:5432/db?sslmode=disable"

	encrypted, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if encrypted == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}

	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := testKey(t)
	key2 := testKey(t)

	encrypted, _ := Encrypt(key1, "secret data")

	_, err := Decrypt(key2, encrypted)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	_, err := Encrypt([]byte("short"), "data")
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestDecrypt_InvalidKeyLength(t *testing.T) {
	_, err := Decrypt([]byte("short"), "data")
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	key := testKey(t)
	_, err := Decrypt(key, "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecrypt_TruncatedCiphertext(t *testing.T) {
	key := testKey(t)
	_, err := Decrypt(key, "YWJj") // "abc" in base64 — too short
	if err == nil {
		t.Fatal("expected error for truncated ciphertext")
	}
}

func TestEncrypt_DifferentNonce(t *testing.T) {
	key := testKey(t)
	plaintext := "same input"

	enc1, _ := Encrypt(key, plaintext)
	enc2, _ := Encrypt(key, plaintext)

	if enc1 == enc2 {
		t.Error("two encryptions of same plaintext should produce different ciphertext (random nonce)")
	}
}
