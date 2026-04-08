# ADR 020: Shared packages — pkg/traces and pkg/cache

**Date:** 2026-04-08
**Status:** Accepted
**Plan:** 12

## Decision

Extract trace publishing and Redis caching to shared `pkg/` packages so multiple services can reuse them.

## Context

The agent service had a `TracePublisher` for NATS event publishing. The astro service needed the same pattern. Duplicating would create drift.

## Choice

- `pkg/traces/publisher.go` — shared Publisher with methods: Start, End, Event, Feedback, Notify, Broadcast. Used by agent + astro. Agent wraps it with backward-compatible method names.
- `pkg/cache/redis.go` — generic Redis JSON cache with graceful degradation (nil client = no-op). JSONCache with Get/Set/Del.

## Consequences

- Any new service gets trace publishing by importing `pkg/traces`
- Redis caching available to search, agent, and future services
- Agent's `service/traces.go` is a thin wrapper (backward compat)
- Both packages have zero external dependencies beyond nats.go
