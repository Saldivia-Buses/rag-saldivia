---
title: RAG via Tree Reasoning
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/ingest.md
  - ../services/search.md
  - ../services/agent.md
  - ../services/extractor.md
  - llm-sglang.md
  - storage-minio.md
---

This document describes how SDA does retrieval-augmented generation without
vector embeddings. Read it before changing the document tree shape, the
ingest pipeline, the search prompt, or the agent's `search_documents` tool
contract — these are the only paths through which the agent obtains grounded
text.

## Why no vectors

SDA RAG is **PageIndex-inspired**: documents are converted to a
hierarchical TOC ("tree") of titled sections with summaries, and an LLM
navigates the tree by title and summary to pick the relevant nodes. There
is no embedding model, no vector database, no nearest-neighbour search.
This trades recall on tiny chunks for explainable, citation-friendly
selections aligned with the document's own structure.

## Tree shape

The canonical node (`services/ingest/internal/tree/types.go:6`):

| Field        | Meaning                                       |
|--------------|-----------------------------------------------|
| `title`      | Section heading (LLM-derived)                 |
| `node_id`    | Sequential id assigned post-build             |
| `start_index`| First page (1-based)                          |
| `end_index`  | Last page                                     |
| `summary`    | Short LLM summary used during navigation      |
| `nodes`      | Children (recursive)                          |

Trees live in the tenant DB `document_trees` table as JSONB
(`db/tenant/migrations/006_ingest_init.up.sql`). Each tree row references
one document.

## Ingest pipeline

1. **Upload.** `POST /v1/ingest/upload` accepts a file. The Ingest service
   computes a content hash, deduplicates against `documents.file_hash`, and
   stores the bytes in MinIO under `{tenant}/{docID}/original.{ext}`
   (`services/ingest/internal/service/documents.go:89`). See
   storage-minio.md.
2. **Trigger extraction.** Ingest publishes
   `tenant.{slug}.extractor.job` with the storage key
   (`documents.go:114`). See nats-events.md.
3. **Extract pages.** The Python `extractor` service downloads the file,
   runs PaddleOCR-VL via SGLang for text/tables and Qwen3.5-9B for image
   descriptions (`services/extractor/main.py:5`), and publishes the
   `ExtractionResult` back over NATS.
4. **Build tree.** The Ingest worker consumes the result and calls
   `tree.Generator.Generate` (`services/ingest/internal/tree/generate.go:33`)
   which: builds page-indexed unified text → asks the LLM for a flat list of
   sections with `physical_index` → assigns start/end indices → converts the
   flat list into a nested tree → numbers nodes → generates per-node
   summaries in parallel → generates a one-line `doc_description`.
5. **Persist.** The tree JSONB and `doc_description` are written to
   `document_trees`; `documents.status` becomes `ready`.

The Generator uses a versioned prompt set (`Prompts` in `generate.go:73`)
loaded from the `prompt_versions` table — never inline new prompt strings.

## Search pipeline

`SearchDocuments` (`services/search/internal/service/search.go:62`) runs
the query in three phases:

- **Phase A — navigate.** Validate the query through guardrails, load all
  trees for the requested collection (or all ready documents), build a
  compact `[node_id] title — summary (pages X-Y)` view, and ask the LLM to
  return up to `maxNodes` comma-separated `node_id`s. The default cap is 5
  with a hard ceiling of 20 (`search.go:69`).
- **Phase B — extract.** Pure code: for each selected node, group by
  document, pull pages between `start_index` and `end_index` from the
  `document_pages` table, and assemble a `Selection` with text, tables,
  images, and section titles.
- **Phase C — return.** A `SearchResult` with one selection per document,
  each carrying citations (`document_id`, `node_ids`, `pages`,
  `sections`).

The agent calls this via the `search_documents` tool wired in
`services/agent/cmd/main.go:63` and treats the response text as grounded
context for its next LLM turn.

## What you must never do

- Add an embedding step or vector store.
- Skip Phase A and feed full documents to the LLM — the tree is the budget.
- Mutate `document_trees.tree` outside the Ingest service.
- Persist a tree without `node_id`s — search depends on them for citations.
