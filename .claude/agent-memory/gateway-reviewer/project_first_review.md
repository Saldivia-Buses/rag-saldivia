---
name: First review context
description: PR #34 Platform Service was the first gateway review on 2.0.x branch -- establishes baseline patterns for future reviews
type: project
---

PR #34 (Platform Service) was reviewed 2026-04-02. Key patterns observed:
- Auth middleware uses `X-Platform-Admin` header set by Traefik/gateway
- Services use chi router, slog, pgxpool, sqlc-generated queries
- Auth service pattern: MaxBytesReader, RequestID, RealIP, Recoverer, Timeout
- No Traefik config yet exists for Platform Service routing or X-Platform-Admin injection
- Feature flags and global config use raw pgx queries instead of sqlc

**Why:** Establishes what "good" looks like for this codebase so future reviews can reference consistency.
**How to apply:** Compare new service PRs against these established patterns.
