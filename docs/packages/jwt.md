---
title: Package: pkg/jwt
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/auth-jwt.md
  - ./middleware.md
  - ./security.md
---

## Purpose

JWT creation and verification primitives for SDA Framework. Uses Ed25519
asymmetric signing — Auth Service signs with the private key, every other
service verifies with the public key only, so a compromised non-auth service
cannot forge tokens. See `architecture/auth-jwt.md` for the auth model. Import
this when bootstrapping the public key, signing tokens (Auth only), or
verifying them outside the standard middleware.

## Public API

Source: `pkg/jwt/jwt.go:8`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Claims` | struct | `RegisteredClaims` + `UserID`, `Email`, `Name`, `TenantID`, `Slug`, `Role`, `Permissions` |
| `Config` | struct | `PrivateKey`, `PublicKey`, `AccessExpiry`, `RefreshExpiry`, `Issuer` |
| `DefaultConfig(privateKey, publicKey)` | func | 15min access / 7day refresh / `Issuer="sda"` |
| `VerifyOnlyConfig(publicKey)` | func | For services that only verify |
| `CreateAccess(cfg, claims)` | func | Sign 15min token (auto-fills JTI) |
| `CreateRefresh(cfg, claims)` | func | Sign 7d token |
| `Verify(publicKey, token)` | func | Parses + validates, returns `*Claims` |
| `ParsePrivateKeyPEM` / `ParsePublicKeyPEM` | func | PEM → Ed25519 |
| `ParsePrivateKeyEnv` / `ParsePublicKeyEnv` | func | Base64 PEM (env var) → Ed25519 |
| `MustLoadPublicKey(envVar)` | func | Panics if env var missing/invalid — for `cmd/main.go` |
| `ErrInvalidToken`/`ErrMissingClaim`/`ErrInvalidKey` | var | Sentinel errors |

## Usage

```go
// In every cmd/main.go
pubKey := jwt.MustLoadPublicKey("AUTH_PUBLIC_KEY")
r.Use(middleware.Auth(pubKey))

// Auth service: sign tokens
cfg := jwt.DefaultConfig(privateKey, pubKey)
access, err := jwt.CreateAccess(cfg, jwt.Claims{
    UserID: u.ID, Email: u.Email, TenantID: tid, Slug: slug, Role: "admin",
})
```

## Invariants

- Only Auth Service holds the private key. Every other service only has the
  public key — verification is local, no network round trip.
- `Verify` REQUIRES `UserID`, `TenantID`, and `Slug` to be present
  (`pkg/jwt/jwt.go:139`) — returns `ErrMissingClaim` otherwise.
- Tokens are auto-assigned a UUID `JTI` if `claims.ID` is empty
  (`pkg/jwt/jwt.go:78`) so the blacklist (`pkg/security`) can revoke them.
- Signing method is locked to `EdDSA`. Any other algorithm in the token header
  is rejected (`pkg/jwt/jwt.go:125`).
- The auth middleware also rejects tokens with `Role == "mfa_pending"`.

## Importers

24+ files: `auth`, `agent`, `astro`, `bigbrother`, `chat`, `erp`, `feedback`,
`healthwatch`, `ingest`, `notification`, `platform`, `search`, `traces`,
`ws` — primarily in `cmd/main.go` plus auth-related handlers.
