// Package crypto provides AES-256-GCM encryption for sensitive data at rest.
// Used to encrypt tenant connection strings in the Platform DB.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey        = errors.New("encryption key must be 32 bytes (AES-256)")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Returns base64-encoded ciphertext with nonce prepended.
func Encrypt(masterKey []byte, plaintext string) (string, error) {
	if len(masterKey) != 32 {
		return "", ErrInvalidKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	// nonce is prepended to ciphertext
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts base64-encoded ciphertext produced by Encrypt.
func Decrypt(masterKey []byte, ciphertext string) (string, error) {
	if len(masterKey) != 32 {
		return "", ErrInvalidKey
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: base64 decode failed", ErrInvalidCiphertext)
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, sealed := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("%w: decryption failed", ErrInvalidCiphertext)
	}

	return string(plaintext), nil
}
