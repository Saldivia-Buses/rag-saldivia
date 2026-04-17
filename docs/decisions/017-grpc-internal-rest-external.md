# ADR 017: gRPC internal, REST external

**Date:** 2026-04-04
**Status:** Deprecated (2026-04-17 — superseded by ADR 025)
**Plan:** 09

## Decision

Use gRPC for internal service-to-service communication and REST for external (frontend → backend) APIs.

## Context

Services need to communicate. Options: all REST, all gRPC, or hybrid.

## Choice

- **External (frontend → Traefik → service):** REST/JSON over HTTP. Browsers speak HTTP natively.
- **Internal (agent → search, ws → chat):** gRPC with protobuf. Type-safe, fast, streaming.
- **JWT forwarding:** gRPC interceptors forward the user's JWT from REST to gRPC calls.

## Consequences

- Proto files in `proto/` define service contracts
- `pkg/grpc/` provides server helpers
- Agent Runtime uses gRPC for search (falls back to HTTP if unavailable)
- WS Hub uses gRPC for chat mutations
- Module tools (fleet, erp) use HTTP (simpler, compute-dominated, gRPC overhead negligible)

## Deprecation (2026-04-17)

ADR 025 collapses the 13 Go services into a single binary. With every
internal caller living in the same process, the two gRPC seams this ADR
named both became in-process method calls:

- **agent → search:** inlined during the rag fusion. Search's gRPC
  server (`:50051`, `searchv1.RegisterSearchServiceServer`, the agent's
  `GRPCSearchClient` fallback) all deleted.
- **ws → chat:** inlined during the realtime fusion. Chat's gRPC
  server (`:50052`, `chatv1.RegisterChatServiceServer`, WS's
  `MutationHandler` gRPC client + `ForwardJWT`) all deleted.

No other gRPC callers existed inside the workspace. That left
`pkg/grpc/`, `gen/go/`, and `proto/` with zero consumers, so the
realtime fusion deleted all three outright together with the `make
proto` target and the CI build/vet/test steps that touched
`./gen/go/...`. Regenerating protobufs is no longer part of the build.

If cross-process communication comes back as a real requirement
later (e.g., extracting `erp` into its own deployable), the question
will be re-opened in a new ADR — likely preferring HTTP+JSON over
reviving gRPC, since the browser-facing surface already speaks HTTP
and the performance argument for gRPC between same-kernel processes
on a vertical-scale box doesn't hold.
