---
title: Package: pkg/crypto
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./tenant.md
---

## Purpose

Two encryption primitives for sensitive data at rest. `Encrypt`/`Decrypt`
provide single-key AES-256-GCM (used to encrypt tenant connection strings in
the Platform DB). `Encryptor` provides envelope encryption (KEK + per-secret
DEK) with AAD binding (used for credentials, OAuth tokens, anything that needs
fast key rotation or per-row context binding). Import this whenever data
must be encrypted before storage.

## Public API

Sources: `pkg/crypto/crypto.go:3`, `pkg/crypto/envelope.go:16`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Encrypt(masterKey, plaintext)` | func | Single-key AES-256-GCM, base64-encoded output with prepended nonce |
| `Decrypt(masterKey, ciphertext)` | func | Inverse of `Encrypt` |
| `Encryptor` | struct | KEK/DEK envelope encrypter, thread-safe (KEK read-only after init) |
| `NewEncryptor(kekPath)` | func | Loads KEK from file (must be exactly 32 raw bytes) |
| `NewEncryptorFromBytes(kek)` | func | Same but takes a slice (testing) |
| `Encryptor.Encrypt(plaintext, aad)` | method | Returns `(encDEK, encData)` |
| `Encryptor.Decrypt(encDEK, encData, aad)` | method | Inverse — fails if AAD mismatches |
| `ErrInvalidKey`/`ErrInvalidCiphertext` | var | From simple AES path |
| `ErrInvalidKEK`/`ErrDecryptFailed` | var | From envelope path |

## Usage

```go
// Envelope encryption with AAD binding
enc, _ := crypto.NewEncryptor("/run/secrets/bb_kek")
aad := []byte(credentialID + "||" + deviceID + "||" + tenantID)
encDEK, encData, err := enc.Encrypt(plaintext, aad)
// store both blobs in DB
// later:
plaintext, err := enc.Decrypt(encDEK, encData, aad) // fails if blobs swapped
```

## Invariants

- KEK file MUST be exactly 32 raw bytes — no trailing newline. Generate with
  `openssl rand 32 > kek-file` (`pkg/crypto/envelope.go:48`).
- AAD MUST be reconstructed identically at decrypt time. Swapping AAD between
  rows is the attack this defends against.
- `Encryptor.Encrypt` zeroes the ephemeral DEK with `defer clear(dek)`
  (`pkg/crypto/envelope.go:79`). Callers should similarly zero plaintext.
- Caller is responsible for rate-limiting decryption operations (defense
  against ciphertext-bashing).
- The simple `Encrypt`/`Decrypt` path lacks AAD — only suitable when
  ciphertext binding to context is not required (e.g., column-level encryption
  of a unique URL).

## Importers

`services/auth/internal/service/mfa.go` (TOTP secrets), `services/bigbrother`
(credentials & cmd/main.go for envelope), `pkg/tenant/resolver.go:16` (decrypt
tenant connection URLs).
