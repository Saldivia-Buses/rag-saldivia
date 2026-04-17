# ADR 019: Astro intelligence layer — deterministic, no LLM

**Date:** 2026-04-08
**Status:** Superseded 2026-04-17 — astro service removed entirely from scope. Product pivot: Saldivia RAG for bus company.
**Plan:** 12

## Decision

The astro intelligence layer (domain routing, technique gating, cross-references, quality audit) is fully deterministic. No LLM calls in the intelligence pipeline.

## Context

The Python astro-v2 used a Semantic Router (FastEmbed ONNX embeddings) for intent classification. The Go port needs intent parsing and domain routing.

Options:
1. Port the Semantic Router (requires ONNX runtime, embedding model, GPU)
2. LLM-based classification (expensive per query, adds latency)
3. Keyword + regex matching (deterministic, zero cost, <1ms)

## Choice

Keyword + regex matching for intent parsing. Struct field inspection for technique gating (not regex on text like Python). Set-intersection algorithms for cross-reference detection.

The LLM's job is NARRATION ONLY. Everything else is deterministic:
- Domain routing: keyword match → domain → inheritance resolution
- Technique gating: `len(ctx.Transits) > 0` → validated (not regex on brief text)
- Cross-references: planet overlap across technique results
- Quality audit: technique coverage scoring, ghost detection, precaution checking
- All under 10ms combined

## Consequences

- No embedding model needed (saves GPU VRAM + dependency)
- No per-query LLM cost for routing
- Keyword parser handles ~90% of queries correctly
- `WithLLM()` option reserved for future upgrade on ambiguous queries
- Quality audit runs post-LLM, checks response against computed facts
- Zero chance of LLM hallucinating the routing/gating layer
