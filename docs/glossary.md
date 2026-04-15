---
title: Glossary
audience: ai
last_reviewed: 2026-04-15
related:
  - ./README.md
  - ./architecture/multi-tenant.md
---

## Purpose

Domain terms used across SDA Framework, ordered alphabetically. One or two
sentences each. Link to the deeper doc when the term needs more than that.

## Terms

**Agent Runtime** — `services/agent/`. The Go service that executes LLM
calls plus tool invocations on behalf of the chat session. Replaces the
deprecated `services/rag/` proxy.

**AFIP** — Administración Federal de Ingresos Públicos. Argentine federal
tax authority. Source of CAE issuance and CUIT validation rules used by the
ERP service.

**CAE** — Código de Autorización Electrónico. The authorization code AFIP
returns when an electronic invoice (factura) is approved. Required to issue
the invoice legally.

**CUIT** — Código Único de Identificación Tributaria. Argentine
taxpayer ID, 11 digits with check digit. Used to identify any business or
person on a fiscal document.

**Fiscal Year** — The accounting year used for journal entries and
financial statements. Configurable per tenant; defaults to the calendar
year. Used to gate journal entry posting.

**Histrix** — The legacy Saldivia Buses intranet system being replaced by
SDA Framework. References to "Histrix scrape" point to the captured MySQL
schema and tables under `.intranet-scrape/` used as a migration source.

**JetStream** — The persistence layer of NATS. Provides streams, durable
consumers, and at-least-once delivery. The `NOTIFICATIONS` stream is the
canonical example.

**Journal Entry** — A double-entry accounting record: one or more debit
lines balanced by one or more credit lines, posted against a fiscal year.
Owned by the ERP service.

**JWT Access** — Short-lived signed token (15 min) carrying user identity
and tenant claims (`uid`, `email`, `tid`, `slug`, `role`). Sent in
`Authorization: Bearer ...`. The single source of identity for all
services.

**JWT Refresh** — Long-lived token (7 days) used only to mint a new
access token. Stored hashed; rotated on each use.

**NATS Subject** — Topic name on the NATS bus. SDA convention:
`tenant.{slug}.{service}.{entity}[.{action}]`. Always tenant-namespaced;
consumers subscribe with `tenant.*.{service}.>`.

**Permission** — A capability identifier (e.g., `chat:write`,
`platform:admin`) granted to a role. Checked by handler middleware against
the JWT role claim.

**Platform DB** — The single PostgreSQL database shared by all tenants.
Holds the tenant registry (`tenants` table), platform-level config,
modules, feature flags, and traces.

**Receipt** — A document acknowledging payment received against an
invoice. Owned by the ERP service; ties one or more journal entries to a
customer transaction.

**Refresh Rotation** — Security policy: each refresh token can be used
exactly once. The auth service issues a new refresh token every time and
revokes the previous one, so a leaked token is detectable.

**Role** — A named bundle of permissions assigned to a user (`platform_admin`,
`tenant_admin`, `user`). Encoded in the JWT `role` claim.

**SGLang** — The LLM serving runtime running on the workstation GPU.
Exposes an OpenAI-compatible HTTP API. One SGLang instance per model
(LLM, OCR, vision); pipeline steps occupy slots on the same instance.

**Slug** — The URL-safe short name of a tenant (e.g., `saldivia`). Used in
NATS subjects, DB naming (`sda_tenant_{slug}`), and the subdomain
(`{slug}.sda.example.com`). Resolved from the JWT `slug` claim.

**Tenant** — A logically isolated customer of the platform. Has its own
PostgreSQL database, its own Redis namespace, and its own slice of NATS
subjects. Rows in tenant DBs always carry a `tenant_id` defensively, even
though the DB is per-tenant.

**Tenant DB** — The PostgreSQL database scoped to a single tenant
(`sda_tenant_{slug}`). Holds users, sessions, chat messages, documents,
journal entries — everything tenant-specific.

**Trace (execution trace)** — A persisted record of one agent invocation:
the LLM calls, tool calls, latencies, token counts, cost, and final
output. Stored by the traces service for cost reporting and replay.

**Tree Reasoning** — The retrieval approach used in place of vectors.
Documents are pre-built into a hierarchical tree (PageIndex-inspired);
the search service walks the tree top-down at query time. No embedding
store.
