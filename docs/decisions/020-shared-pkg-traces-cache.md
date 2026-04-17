# ADR 020: Shared packages — pkg/traces and pkg/cache

**Date:** 2026-04-08
**Status:** Accepted
**Plan:** 12

## Decision

Extract trace publishing and Redis caching to shared `pkg/` packages so multiple services can reuse them.

## Context

The agent service needed a `TracePublisher` for NATS event publishing. Extracted to `pkg/` early so new services could reuse the pattern without duplication.

## Choice

- `pkg/traces/publisher.go` — shared Publisher with methods: Start, End, Event, Feedback, Notify, Broadcast. Used by agent. Agent wraps it with backward-compatible method names.
- `pkg/cache/redis.go` — generic Redis JSON cache with graceful degradation (nil client = no-op). JSONCache with Get/Set/Del. *Note: as of 2026-04-17, pkg/cache has zero importers and is slated for deletion in the dead-package cleanup pass.*

## Consequences

- Any new service gets trace publishing by importing `pkg/traces`
- Agent's `service/traces.go` is a thin wrapper (backward compat)
- pkg/traces has zero external dependencies beyond nats.go
