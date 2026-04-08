# ADR 017: gRPC internal, REST external

**Date:** 2026-04-04
**Status:** Accepted
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
- Module tools (fleet, astro) use HTTP (simpler, compute-dominated, gRPC overhead negligible)
