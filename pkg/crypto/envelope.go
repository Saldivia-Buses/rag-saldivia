// Envelope encryption provides KEK/DEK two-layer encryption with AAD.
//
// Each secret gets its own random DEK (Data Encryption Key). The DEK encrypts
// the data, and the KEK (Key Encryption Key) encrypts the DEK. This allows
// key rotation by re-encrypting only the DEKs (fast), not the data.
//
// AAD (Additional Authenticated Data) binds the ciphertext to its context.
// If someone swaps encrypted blobs between rows, aead.Open() fails because
// the AAD won't match. Callers provide the AAD (e.g., "credential_id||device_id||tenant_id").
//
// Usage:
//
//	enc, err := crypto.NewEncryptor("/run/secrets/bb_kek")
//	encDEK, encData, err := enc.Encrypt(plaintext, aad)
//	plaintext, err := enc.Decrypt(encDEK, encData, aad)
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrInvalidKEK   = errors.New("KEK must be exactly 32 bytes")
	ErrDecryptFailed = errors.New("decryption failed")
)

// Encryptor provides KEK/DEK envelope encryption with AAD.
// Thread-safe: the KEK is read-only after construction.
type Encryptor struct {
	kek []byte
}

// NewEncryptor loads a KEK from a file and validates it is exactly 32 bytes.
// The file should contain raw bytes (not hex-encoded).
func NewEncryptor(kekPath string) (*Encryptor, error) {
	raw, err := os.ReadFile(kekPath)
	if err != nil {
		return nil, fmt.Errorf("read KEK: %w", err)
	}
	// KEK must be exactly 32 raw bytes. No trimming — file must not have
	// trailing newlines. Generate with: openssl rand 32 > kek-file
	if len(raw) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes (file must be exactly 32 raw bytes, no trailing newline)", ErrInvalidKEK, len(raw))
	}
	kek := make([]byte, 32)
	copy(kek, raw)
	clear(raw)
	return &Encryptor{kek: kek}, nil
}

// NewEncryptorFromBytes creates an Encryptor from a raw KEK slice.
// Useful for testing. The slice is copied internally.
func NewEncryptorFromBytes(kek []byte) (*Encryptor, error) {
	if len(kek) != 32 {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKEK, len(kek))
	}
	k := make([]byte, 32)
	copy(k, kek)
	return &Encryptor{kek: k}, nil
}

// Encrypt encrypts plaintext with a fresh random DEK, then encrypts the DEK
// with the KEK. Both use AES-256-GCM. AAD is bound to both layers.
//
// Returns (encryptedDEK, encryptedData) as raw bytes for storage.
// Caller is responsible for zeroing plaintext after use if sensitive.
func (e *Encryptor) Encrypt(plaintext, aad []byte) (encDEK, encData []byte, err error) {
	// Generate random DEK
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, nil, fmt.Errorf("generate DEK: %w", err)
	}
	defer clear(dek)

	// Encrypt data with DEK + AAD
	encData, err = aesGCMSeal(dek, plaintext, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt data: %w", err)
	}

	// Encrypt DEK with KEK + AAD
	encDEK, err = aesGCMSeal(e.kek, dek, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt DEK: %w", err)
	}

	return encDEK, encData, nil
}

// Decrypt decrypts the DEK with the KEK, then decrypts data with the DEK.
// Both use AES-256-GCM with the provided AAD. If AAD doesn't match what was
// used during encryption, decryption fails (prevents blob swapping).
//
// Caller is responsible for rate limiting decryption operations.
func (e *Encryptor) Decrypt(encDEK, encData, aad []byte) ([]byte, error) {
	// Decrypt DEK with KEK + AAD
	dek, err := aesGCMOpen(e.kek, encDEK, aad)
	if err != nil {
		return nil, fmt.Errorf("%w: decrypt DEK: %v", ErrDecryptFailed, err)
	}
	defer clear(dek)

	// Decrypt data with DEK + AAD
	plaintext, err := aesGCMOpen(dek, encData, aad)
	if err != nil {
		return nil, fmt.Errorf("%w: decrypt data: %v", ErrDecryptFailed, err)
	}

	return plaintext, nil
}

// aesGCMSeal encrypts plaintext using AES-256-GCM with random nonce and AAD.
// Returns nonce prepended to ciphertext.
func aesGCMSeal(key, plaintext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, aad), nil
}

// aesGCMOpen decrypts ciphertext (nonce prepended) using AES-256-GCM with AAD.
func aesGCMOpen(key, ciphertext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, sealed := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, sealed, aad)
}
