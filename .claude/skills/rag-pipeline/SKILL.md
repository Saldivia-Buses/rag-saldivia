---
name: rag-pipeline
description: Use when working on the RAG stack — services/ingest, services/search, services/agent, services/extractor, the crossdoc decompose/synth flow, Milvus or embedding code, or the smart ingestion CLI. Covers the tree-search retrieval model, the crossdoc 4-phase pipeline, the tier-based smart ingest, and how the three services talk to each other via NATS + gRPC.
---

# rag-pipeline

Scope: `services/ingest/`, `services/search/`, `services/agent/`, `services/extractor/`,
`scripts/smart_ingest.py`, any hook or component named `useCrossdoc*`.

## Core decisions (read before changing anything)

Two ADRs govern this area — read them if you are about to deviate:

- `docs/decisions/016-tree-rag-no-vectors.md` — why retrieval is tree-based, not
  flat vector similarity.
- `docs/decisions/020-shared-pkg-traces-cache.md` — shared tracing / cache layer.

## Services

| Service | Responsibility |
|---|---|
| `extractor` | Python; takes a document (PDF, DOCX, …) → structured text + metadata |
| `ingest` | Go; orchestrates extraction, chunks, stores, publishes events |
| `search` | Go; tree-search retrieval, exposes gRPC on `:50051` and REST on `:8010` |
| `agent` | Go; LLM-facing; calls `search`, composes answers, streams via WS hub |

Data flow: `upload → ingest → extractor → ingest (chunk + store) → NATS event →
agent reacts → search → LLM → WS stream`.

## Tree-search retrieval (not flat vectors)

Documents are stored as a tree of nodes (sections → paragraphs → chunks). Retrieval:

1. Query decomposition (crossdoc, below) into sub-queries.
2. Tree traversal with learned relevance — not k-NN over a single embedding.
3. Reranker over candidate nodes.
4. Top-N nodes returned with provenance to the agent.

If you are about to add a cosine-similarity vector lookup, **stop** and re-read
ADR 016.

### Why this matches 2026 RAG literature

The tree-search approach aligns with what current research calls "hierarchical
agentic RAG" (see: A-RAG paper on arXiv, GraphRAG). Parent nodes hold context,
child nodes hold density; the model navigates instead of flat similarity. When
extending retrieval, the principles that still apply:

- **Hybrid signals** — combine tree relevance with BM25 scoring on leaves; don't rely
  on a single metric.
- **Contextual chunking** — chunks carry their ancestor titles/summaries so they
  stand alone when surfaced.
- **Reranking is mandatory** — never hand the agent raw retrieval output; rerank.
- **Metadata filtering first** — tenant, doc_id, date, owner — before any scoring.
- **Provenance is structural** — the tree path *is* the citation. Never return a
  chunk without its ancestry.

## Crossdoc: 4-phase pipeline

Implemented client-side in React (`apps/web/src/hooks/useCrossdoc*.ts`) and
server-side in `agent`:

1. **Decompose** — break the question into ≤ N sub-questions.
2. **Retrieve** — run `search` per sub-question in parallel.
3. **Synthesize** — combine evidence into an answer.
4. **Follow-up** — up to K retries if the answer is ungrounded.

Settings (exposed in the UI `SaldiviaSection`):
`queryMode`, `maxSubQueries`, `synthesisModel`, `followUpRetries`, `showDecomposition`,
`vdbTopK`, `rerankerTopK`.

Any change to the algorithm has to stay coherent across the hook and the server —
they share the same 4 phases. If you change one without the other, streaming will
misalign.

## Smart ingest (`scripts/smart_ingest.py`)

Tier system — classifies documents by estimated page count and routes them:

| Tier | Pages | Strategy |
|---|---|---|
| tiny | ≤ 2 | inline, no chunking needed |
| small | 3–20 | normal pipeline |
| medium | 21–100 | batched, adaptive timeout |
| large | > 100 | streamed, checkpointed, resumable |

Features: deadlock detection, resume on failure, adaptive timeout per tier.
Do not rewrite the tier logic without measurements — it was tuned on real docs.

## Storage

- Postgres: metadata, tree structure, ingestion state.
- MinIO: raw documents + extracted artifacts.
- Redis: ephemeral ingestion locks, reranker cache.
- Milvus (if enabled): legacy/experimental vector store — do not add new callers.

## gRPC boundary

- `search` exposes gRPC on `:50051` for `agent`. REST is for debugging and admin.
- Protos live in `pkg/grpc/proto/`. Regenerate with `make proto`.
- Streaming search results (the agent needs them live) uses server-streaming RPC.

## Events

- `tenant.{slug}.ingest.document.created|completed|failed`
- `tenant.{slug}.ingest.chunk.created`
- `tenant.{slug}.agent.answer.streaming|completed`

WebSocket hub fans these out to the UI.

## Don't

- Don't add a new vector store.
- Don't bypass `ingest` to write chunks directly.
- Don't change the crossdoc phase count — it is load-bearing across hook + server.
- Don't log document contents at INFO — use DEBUG with tenant+doc ID only.
