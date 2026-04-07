---
name: Plan 09 — gRPC inter-service
description: gRPC plan written 2026-04-05, 8 phases (0-7), pragmatic scope — only agent→search and agent→ingest wired, rest deferred
type: project
---

Plan 09 implements gRPC for inter-service communication. Written 2026-04-05.

**Key findings during research:**
- 6 proto files exist (1187 lines, 41 RPCs) but ZERO generated code, ZERO buf config
- `make proto` target exists but doesn't work (no buf.yaml)
- All inter-service calls are HTTP REST, all from agent service (4 calls: search, ingest x2, notification)
- WS Hub does NOT call other services — mutations are stub ("not yet implemented")
- grpc and protobuf are already indirect deps in pkg/go.mod

**Scope decisions:**
- Only wire agent→search (gRPC) and agent→ingest (gRPC for read, HTTP for upload)
- search proto doesn't exist yet — needs to be created
- WS Hub mutations, auth, platform, notification deferred
- Fallback HTTP maintained if gRPC unavailable

**Why:** Closes gap between spec (says gRPC) and reality (all HTTP). Plan 08 prerequisite.
**How to apply:** This plan is sequential after plan 08. 7 implementation phases + phase 0 (security design).
