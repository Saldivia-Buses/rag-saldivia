package crypto

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func randomKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return key
}

func TestEncryptorRoundtrip(t *testing.T) {
	enc, err := NewEncryptorFromBytes(randomKey(t))
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("ssh-rsa AAAA... secret key material")
	aad := []byte("cred-123||device-456||tenant-789")

	encDEK, encData, err := enc.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}

	got, err := enc.Decrypt(encDEK, encData, aad)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip failed: got %q, want %q", got, plaintext)
	}
}

func TestEncryptorAADMismatch(t *testing.T) {
	enc, err := NewEncryptorFromBytes(randomKey(t))
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("secret")
	aad := []byte("cred-123||device-456||tenant-789")
	wrongAAD := []byte("cred-123||device-WRONG||tenant-789")

	encDEK, encData, err := enc.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}

	_, err = enc.Decrypt(encDEK, encData, wrongAAD)
	if err == nil {
		t.Fatal("expected error for AAD mismatch, got nil")
	}
}

func TestEncryptorWrongKEK(t *testing.T) {
	enc1, _ := NewEncryptorFromBytes(randomKey(t))
	enc2, _ := NewEncryptorFromBytes(randomKey(t))

	plaintext := []byte("secret")
	aad := []byte("context")

	encDEK, encData, err := enc1.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}

	_, err = enc2.Decrypt(encDEK, encData, aad)
	if err == nil {
		t.Fatal("expected error for wrong KEK, got nil")
	}
}

func TestEncryptorInvalidKEKLength(t *testing.T) {
	_, err := NewEncryptorFromBytes([]byte("too-short"))
	if err == nil {
		t.Fatal("expected error for short KEK")
	}

	_, err = NewEncryptorFromBytes(make([]byte, 64))
	if err == nil {
		t.Fatal("expected error for long KEK")
	}
}

func TestEncryptorFromFile(t *testing.T) {
	dir := t.TempDir()
	kekPath := filepath.Join(dir, "kek")
	kek := randomKey(t)
	if err := os.WriteFile(kekPath, kek, 0600); err != nil {
		t.Fatal(err)
	}

	enc, err := NewEncryptor(kekPath)
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("from-file-test")
	aad := []byte("ctx")
	encDEK, encData, err := enc.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatal(err)
	}
	got, err := enc.Decrypt(encDEK, encData, aad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("got %q, want %q", got, plaintext)
	}
}

func TestEncryptorFromFileMissing(t *testing.T) {
	_, err := NewEncryptor("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEncryptorDifferentPlaintextsSameAAD(t *testing.T) {
	enc, _ := NewEncryptorFromBytes(randomKey(t))
	aad := []byte("same-context")

	encDEK1, encData1, _ := enc.Encrypt([]byte("secret-1"), aad)
	encDEK2, encData2, _ := enc.Encrypt([]byte("secret-2"), aad)

	// Different DEKs per encryption
	if bytes.Equal(encDEK1, encDEK2) {
		t.Fatal("expected different encrypted DEKs")
	}
	if bytes.Equal(encData1, encData2) {
		t.Fatal("expected different encrypted data")
	}

	// Both decrypt correctly
	p1, _ := enc.Decrypt(encDEK1, encData1, aad)
	p2, _ := enc.Decrypt(encDEK2, encData2, aad)
	if string(p1) != "secret-1" || string(p2) != "secret-2" {
		t.Fatalf("got %q and %q", p1, p2)
	}
}

func TestEncryptorEmptyPlaintext(t *testing.T) {
	enc, _ := NewEncryptorFromBytes(randomKey(t))
	aad := []byte("ctx")

	encDEK, encData, err := enc.Encrypt([]byte{}, aad)
	if err != nil {
		t.Fatal(err)
	}
	got, err := enc.Decrypt(encDEK, encData, aad)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestEncryptorNilAAD(t *testing.T) {
	enc, _ := NewEncryptorFromBytes(randomKey(t))

	encDEK, encData, err := enc.Encrypt([]byte("secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := enc.Decrypt(encDEK, encData, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "secret" {
		t.Fatalf("got %q", got)
	}

	// nil AAD during encrypt, non-nil during decrypt should fail
	_, err = enc.Decrypt(encDEK, encData, []byte("wrong"))
	if err == nil {
		t.Fatal("expected error for nil vs non-nil AAD")
	}
}
