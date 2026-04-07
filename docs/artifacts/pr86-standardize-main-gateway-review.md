# Gateway Review -- PR #86 Standardize cmd/main.go (Plan 07 Phases 6+7)

**Fecha:** 2026-04-05
**Resultado:** APROBADO (con sugerencias menores)

## Bloqueantes

Ninguno.

## Debe corregirse

Ninguno.

## Sugerencias

1. **[services/chat/cmd/main.go:11, services/notification/cmd/main.go:11] Double blank line in import block.**
   The sed deletion of `"crypto/ed25519"` left behind an extra empty line in the import grouping separator. Two files affected: chat and notification. Not a compile error, but `goimports` would flag it.
   Fix: remove the extra blank line (one blank between stdlib and external groups, not two).

2. **[All 9 non-auth services] Trailing blank lines at EOF.**
   Every modified file has 2-3 trailing blank lines at the end where the deleted `func env()` and/or `func loadPublicKey()` used to be. Cosmetic only.
   Fix: `truncate` or `sed -e :a -e '/^\n*$/{$d;N;ba' -e '}'` to strip trailing newlines, or just let `gofmt` handle it on next format pass.

3. **[pkg/jwt/jwt.go:195-205] MustLoadPublicKey uses panic() vs old pattern's slog.Error + os.Exit(1).**
   The old `loadPublicKey()` in each service produced a structured JSON log line via slog before calling `os.Exit(1)`. The new `MustLoadPublicKey` uses `panic()` which produces an unstructured stack trace on stderr and exits with status 2. The `Must*` prefix convention is idiomatic Go, so this is acceptable, but it changes the failure UX for operators debugging missing env vars.
   Suggestion: if structured logging on startup failure matters, consider `slog.Error(...)` before the panic, or provide a companion `LoadPublicKey(envVar) (ed25519.PublicKey, error)` that callers can use when they want graceful error handling.

4. **[services/auth/cmd/main.go] Auth service still uses local env() -- consider partial migration.**
   Auth correctly keeps its own `env()` and `loadJWTKeys()` since it needs both private and public keys. However, the `env()` calls (port, DB URLs, NATS URL, Redis URL, tenant ID/slug, OTEL endpoint) could use `config.Env()` to complete the consolidation. The `loadJWTKeys()` can stay since it has different logic (loads two keys, uses slog+os.Exit rather than panic).
   This is a follow-up, not blocking.

5. **[services/traces/cmd/main.go:8] Unsorted stdlib imports (pre-existing).**
   `"strings"` appears between `"os"` and `"os/signal"` -- not caused by this PR but visible in the diff context. `goimports` would sort it.

## Lo que esta bien

1. **MustLoadPublicKey is correctly implemented.** Reads from the env var name passed as parameter (not hardcoded), uses `ParsePublicKeyEnv` for base64-decode + PEM parse, panics with a descriptive message including the env var name. The function is 10 lines replacing ~15 lines of copy-paste in each service.

2. **All 9 non-auth services consistently migrated.** Every one uses `sdajwt.MustLoadPublicKey("JWT_PUBLIC_KEY")` and `config.Env()`. No service was missed.

3. **All local `loadPublicKey()` functions deleted.** Zero residual copies found via grep. Clean removal.

4. **Auth service correctly excluded.** It retains its own `env()` and `loadJWTKeys()` because it needs the private key for signing. No `config` import was forced where it doesn't belong.

5. **No orphaned imports.** `"crypto/ed25519"`, `"crypto/x509"`, `"encoding/base64"`, `"encoding/pem"` -- all removed from non-auth services. No compile errors from unused imports.

6. **No env var names broken by sed.** All `config.Env("ENV_VAR", "default")` calls verified -- every env var key is correctly preserved from the original `env("ENV_VAR", "default")` pattern.

7. **config.Env() is semantically identical to the old env().** Both check `os.Getenv`, return fallback on empty. No behavior change for config reading.

8. **Scaffold template not affected.** `.scaffold/cmd/main.go` is a minimal skeleton without env/loadPublicKey, so it was correctly left alone.

## Verification matrix

| Service | env() removed | loadPublicKey() removed | config.Env() added | MustLoadPublicKey() added | Imports clean | Compiles |
|---------|:---:|:---:|:---:|:---:|:---:|:---:|
| chat | Y | Y | Y | Y | double blank line | Y |
| ws | Y | Y | Y | Y | OK | Y |
| notification | Y | Y | Y | Y | double blank line | Y |
| platform | Y | Y | Y | Y | OK | Y |
| ingest | Y | Y | Y | Y | OK | Y |
| feedback | Y | Y | Y | Y | OK | Y |
| agent | Y | Y | Y | Y | OK | Y |
| search | Y | Y | Y | Y | OK | Y |
| traces | Y | Y | Y | Y | OK | Y |
| auth | N (correct) | N (correct) | N (correct) | N (correct) | OK | Y |
