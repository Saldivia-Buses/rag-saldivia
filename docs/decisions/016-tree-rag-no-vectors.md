# ADR 016: Tree-based RAG (no vectors)

**Date:** 2026-04-02
**Status:** Accepted
**Plan:** 06

## Decision

Use tree-based document search (PageIndex-inspired) instead of vector embeddings for RAG.

## Context

Traditional RAG systems use embedding models to vectorize documents and do similarity search. This requires:
- An embedding model (GPU memory)
- A vector database (Milvus, Qdrant, etc.)
- Chunking strategies that lose document structure

## Choice

The LLM navigates document trees using structural metadata (titles, summaries, page numbers). No vectors, no embedding model, no vector DB.

3-phase pipeline:
1. Navigate: LLM reads tree TOC → selects relevant nodes
2. Extract: code maps nodes → page ranges → reads from PostgreSQL
3. Cite: assembles selections with source citations

## Consequences

- No embedding model needed (saves GPU VRAM)
- No vector database (simpler infra, just PostgreSQL)
- Document structure preserved (sections, chapters, tables)
- LLM reasoning is more accurate than cosine similarity for complex queries
- Requires document tree generation during ingest (more upfront work)
