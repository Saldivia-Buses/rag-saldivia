# ADR-013: Go microservices (migration from 1.0.x)

**Fecha:** 2026-04-01
**Estado:** Aceptado
**Reemplaza:** ADR-012 (stack definitivo para 1.0.x)

---

## Contexto

The 1.0.x series used Next.js + SQLite + Bun as a monolith (ADR-012). While functional for prototyping, the product vision grew to require multi-tenant SaaS with per-tenant databases, background processing (document ingestion, NATS events), real-time WebSocket push, and dedicated AI service proxies (RAG Blueprint, agents).

Key limitations of the 1.0.x stack:

- SQLite cannot support multi-tenant isolation (one DB per tenant)
- Next.js Server Actions mix UI and backend logic -- impossible to scale services independently
- No message broker support (needed for async pipelines, real-time events)
- TypeScript runtime performance insufficient for high-throughput proxy services (RAG streaming)
- Single process model incompatible with 8+ independent services

The hardware target is a single workstation (RTX PRO 6000, 96GB VRAM) running all services -- memory efficiency matters.

## Opciones consideradas

- **Opcion A -- Keep Next.js, add microservices in TypeScript (Nest.js/Fastify):** TypeScript end-to-end. Pros: familiar language, shared types. Contras: Node.js memory overhead per service (~80-150MB each), no native concurrency, weaker performance for streaming proxies.

- **Opcion B -- Go microservices + Next.js frontend:** Go for all backend services, Next.js stays as frontend-only. Pros: ~10-20MB per service, native goroutines for concurrency, single binary deploys, strong stdlib (net/http, crypto, encoding). Contras: separate language for frontend/backend, no shared types.

- **Opcion C -- Rust microservices:** Maximum performance. Pros: zero-cost abstractions, memory safety. Contras: steep learning curve, slow iteration, ecosystem less mature for web services, team of 1.

## Decision

**Opcion B -- Go microservices + Next.js frontend.** The combination gives:

- 8 services running in ~160MB total RAM vs ~800MB+ for Node.js equivalent
- goroutines for concurrent NATS consumers, WebSocket hubs, streaming proxies
- Single static binary per service (no node_modules, no runtime)
- `chi` for HTTP routing (lightweight, stdlib-compatible)
- `sqlc` for type-safe SQL (no ORM, no reflection)
- `slog` for structured logging (stdlib)
- Next.js stays as a pure frontend (React + TanStack Query), deployed to CDN

## Consecuencias

**Positivas:**
- 8 independent services, each deployable and testable in isolation
- Per-service PostgreSQL migrations, clean separation of tenant and platform data
- NATS JetStream for guaranteed delivery (notification, ingest, feedback pipelines)
- WebSocket Hub as a dedicated service with NATS bridge
- Memory footprint allows all services to coexist with GPU workloads on single machine

**Negativas / trade-offs:**
- Complete rewrite from 1.0.x (no code reuse except frontend patterns)
- Two languages in the repo (Go backend, TypeScript frontend)
- No shared type definitions between backend and frontend
- Go error handling is verbose compared to TypeScript
- Team of 1 must maintain Go and TypeScript

## Referencias

- `services/` -- all 8 Go microservices
- `apps/web/` -- Next.js frontend
- `go.work` -- Go workspace configuration
- `docs/plans/2.0.x-plan01-sda-framework.md` -- full system spec
