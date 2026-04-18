---
name: rag-pipeline
description: Use when working on the RAG stack — services/app/internal/rag/{ingest,search,agent}/, services/extractor/ (Python), tree-search retrieval, crossdoc decompose/synth flow, ingest tier pipeline. After ADR 025 the three Go services live in-process inside the app monolith; after ADR 026 the RAG layer answers to Phase 2 (hierarchical prompts + per-collection ACL + memory integration).
---

# rag-pipeline

Scope: `services/app/internal/rag/{ingest,search,agent}/`,
`services/extractor/` (Python sidecar), `scripts/smart_ingest.py`,
`apps/web/src/hooks/useCrossdoc*.ts`, the streaming path through
`services/app/internal/realtime/ws/`.

## Core decisions (read before changing anything)

- **ADR 016 — tree-RAG (no vectors)**: retrieval is tree-based. A new
  cosine-similarity lookup is a re-read-ADR-016 event.
- **ADR 025 — consolidation shape**: `ingest`, `search`, `agent` are no
  longer separate binaries. They are in-process packages under
  `services/app/internal/rag/`. `agent → search` and `agent → ingest`
  are Go function calls, not HTTP/gRPC. `pkg/grpc` was fully deleted
  in the realtime fusion; there is no proto layer to regenerate.
- **ADR 026 — SDA replaces Histrix**: RAG ingests Histrix-migrated
  data, emails, WhatsApp, and uploaded files. Per-collection ACL
  (Phase 2) is non-negotiable — engineering does not see purchasing
  trees unless they have the role.
- **ADR 027 — phased success criteria**: the Phase 2 "Chat as UI" and
  "Hierarchical prompts" items gate a lot of the product. Changes
  here should tick one of those.

## Layout inside the monolith

```
services/app/
└── internal/
    ├── rag/
    │   ├── ingest/          upload → extractor → chunk → tree → events
    │   ├── search/          tree-search retrieval; in-process only
    │   └── agent/           LLM-facing; tool dispatch; streaming
    └── realtime/ws/         WS hub; fans NATS events to the browser
```

`services/extractor/` stays standalone — it's Python, CPU-only, runs
SGLang HTTP calls for OCR (PaddleOCR-VL) and vision (Qwen).

## Data flow

```
upload
  → ingest handler (apps/web fetch)
  → ingest service
      → extractor (NATS request/reply or HTTP to Python sidecar)
      → chunker (tree builder)
      → Postgres (metadata + tree), MinIO (raw + artifacts)
      → NATS: tenant.{slug}.ingest.completed
  → agent query (chat)
      → agent.handler.Query
      → search.Service.Search (in-process)
      → LLM (SGLang / OpenAI-compatible)
      → WS hub stream (tokens + tool calls + final)
```

## Tree-search retrieval (not flat vectors)

Documents are trees of nodes: document → section → paragraph → chunk.
Retrieval flow:

1. Decompose the user query into sub-queries (crossdoc, below).
2. Tree traversal with learned relevance — not k-NN over a single
   embedding.
3. BM25 on leaves as a hybrid signal.
4. Rerank over candidate nodes.
5. Top-N returned with provenance (the tree path IS the citation).

### Why this matches current RAG literature

Aligns with "hierarchical agentic RAG" (A-RAG, GraphRAG). Principles
that still apply when extending retrieval:

- **Hybrid signals** — combine tree relevance with BM25 on leaves.
- **Contextual chunking** — chunks carry ancestor titles/summaries.
- **Reranking is mandatory** — never hand the agent raw retrieval.
- **Metadata filtering first** — tenant, collection, doc_id, date,
  owner — before any scoring.
- **Provenance is structural** — tree path is the citation.

### Per-collection ACL (Phase 2 — not yet implemented)

Collections are the access-control unit. Every tree node belongs to
exactly one collection. Retrieval filters collections by the user's
role **before** scoring, not after. A cross-area denial test lives in
`services/app/internal/rag/search/service/acl_test.go` (to be added
when this lands).

When you work on this, read ADR 026 §Phase 2 and ADR 027 §Phase 2
"Tree-RAG with ACL". Expect an ADR 028 (or similar) formalising the
collection model when the first implementation ships.

## Crossdoc: 4-phase pipeline

Implemented both client-side (`apps/web/src/hooks/useCrossdoc*.ts`)
and server-side (`services/app/internal/rag/agent/`).

1. **Decompose** — break the question into ≤ N sub-questions.
2. **Retrieve** — run search per sub-question in parallel.
3. **Synthesize** — combine evidence into an answer.
4. **Follow-up** — up to K retries if the answer is ungrounded.

Settings exposed in the UI:
`queryMode`, `maxSubQueries`, `synthesisModel`, `followUpRetries`,
`showDecomposition`, `vdbTopK`, `rerankerTopK`.

Any algorithm change must stay coherent across the hook and the
server — they share the same 4 phases. Desyncs → misaligned streaming.

## Hierarchical prompts (Phase 2 — not yet implemented)

The agent assembles context in layers before the user message:

```
system.md          (company-wide: jargon, org shape, policies)
 → area.md         (per-area: ingeniería.md, compras.md, ...)
  → user.md        (editable by the employee)
   → memories      (global + per-user, curated)
    → recent chat  (conversation so far)
     → user query  (this message)
```

Token budgets are enforced per layer so layers don't starve each
other. See the `prompt-layers` skill for the architecture and token
budget tables.

## Memory integration (Phase 2 — not yet implemented)

Memories are ranked alongside tree results in retrieval. Memory
tables (global + per-user) are written by a background curator
agent — see `background-agents` skill. From the RAG side the contract
is: memories participate in retrieval like any other node, filtered
by the same ACL, surfaced with their source (user / global).

## Smart ingest (`scripts/smart_ingest.py`)

Tier system — classifies by estimated page count:

| Tier | Pages | Strategy |
|---|---|---|
| tiny | ≤ 2 | inline, no chunking needed |
| small | 3–20 | normal pipeline |
| medium | 21–100 | batched, adaptive timeout |
| large | > 100 | streamed, checkpointed, resumable |

Features: deadlock detection, resume on failure, adaptive timeout
per tier. Tuning was done on real docs; do not rewrite without
measurements.

## Storage

- **Postgres**: metadata, tree structure, ingestion state, memories.
- **MinIO**: raw documents + extracted artifacts (the raw file is
  the source of truth; we re-extract on demand).
- **Redis**: ephemeral ingestion locks, reranker cache.
- **Milvus**: legacy/experimental — do not add new callers. If a
  cosine use-case appears, justify it in an ADR first.

## Events

Published via outbox (at-least-once):

- `tenant.{slug}.ingest.document.created|completed|failed`
- `tenant.{slug}.ingest.chunk.created`
- `tenant.{slug}.agent.answer.streaming|completed`

The WS hub (`services/app/internal/realtime/ws/`) fans these to the
browser. Ingest submission that publishes direct (not via outbox) is a
**bug** to fix — see ADR 027 Phase 0.

## Don't

- Don't add a new vector store. Read ADR 016.
- Don't bypass `ingest` to write chunks directly.
- Don't re-introduce gRPC between rag sub-modules. They are
  in-process Go calls post-ADR 025.
- Don't change the crossdoc phase count — load-bearing across hook +
  server.
- Don't log document contents at INFO. DEBUG with tenant + doc_id
  only.
- Don't skip the collection ACL filter in new retrieval paths once
  ACL lands (currently a TODO — but new code must not foreclose it).

## Before touching RAG code

1. `git log --oneline -5 -- services/app/internal/rag/`
2. Read ADR 016 + ADR 025 + ADR 026 §Phase 2.
3. If the change concerns ingest from Histrix data: consult
   `.intranet-scrape/` and `htx-parity` skill first.
4. If the change concerns prompts/memories: read `prompt-layers`.
5. If the change concerns a new tool or tool-callable flow: read
   `agent-tools`.
