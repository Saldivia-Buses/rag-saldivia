---
name: PR #86 Standardize cmd/main.go
description: Plan 07 Phases 6+7 -- replaced copy-pasted env() and loadPublicKey() with config.Env() and sdajwt.MustLoadPublicKey(). APPROVED with cosmetic suggestions.
type: project
---

PR #86 replaces local env() with config.Env() and local loadPublicKey() with sdajwt.MustLoadPublicKey() across all 9 non-auth services. Auth correctly excluded (needs both private+public keys).

**Why:** Plan 07 consolidation -- 10 copies of env() and 9 copies of loadPublicKey() were copy-pasted across services.

**How to apply:** This is the first clean APPROVED review with no blockers. The consolidation pattern is now: config.Env() for env vars with defaults, sdajwt.MustLoadPublicKey() for JWT public key loading. Auth service remains the exception with its own env() and loadJWTKeys(). Minor cosmetic issues: double blank lines in import blocks (chat, notification), trailing blank lines at EOF (all modified files).
