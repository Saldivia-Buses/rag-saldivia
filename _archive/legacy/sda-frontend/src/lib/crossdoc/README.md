# Crossdoc

Cross-document query pipeline. Decomposes complex questions into sub-queries, executes them in parallel against the RAG system, deduplicates results, and synthesizes a final answer.

## What is crossdoc?

Crossdoc is a 4-phase pipeline that improves RAG quality for complex, multi-faceted questions by:

1. **Decomposing** the user's question into focused sub-queries (e.g., "What are the benefits and drawbacks of X?" → ["What are the benefits of X?", "What are the drawbacks of X?"])
2. **Executing** sub-queries in parallel (with batching to avoid overload)
3. **Deduplicating** results using Jaccard similarity to remove redundant content
4. **Synthesizing** a coherent answer from the combined results

This approach retrieves more relevant context from different documents, leading to more comprehensive answers.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `types.ts` | TypeScript types for crossdoc: `CrossdocOptions`, `SubResult`, `CrossdocProgress`, and `DEFAULT_CROSSDOC_OPTIONS`. | None |
| `pipeline.ts` | Utility functions for the pipeline: `jaccard()` (similarity), `dedup()` (remove duplicate queries), `parseSubQueries()` (extract sub-queries from LLM output), `hasUsefulData()` (detect empty responses), `truncateIfRepetitive()` (prevent repetitive LLM loops). | None |

## Design notes

### 4-phase pipeline

```
┌───────────────┐
│ User Question │
└───────┬───────┘
        │
        ▼
    Decompose       → LLM generates 3-5 sub-queries
        │
        ▼
 Parallel Subquery  → Execute in batches (max 6 parallel)
        │
        ▼
   Dedup/Rerank     → Jaccard dedup (threshold 0.65)
        │
        ▼
   Synthesize       → LLM combines results into final answer
        │
        ▼
   ┌─────────┐
   │  Answer │
   └─────────┘
```

### Jaccard deduplication

The `jaccard()` function computes word-level similarity between queries (normalized, accent-insensitive). Queries with similarity ≥ 0.65 are considered duplicates and filtered out. This prevents asking the same question multiple times in different phrasings.

### Early exit via `hasUsefulData`

After executing sub-queries, the pipeline checks if any returned useful content. If all responses are empty (e.g., "I cannot answer", "No information available"), the pipeline stops early and returns a message to the user instead of wasting tokens on synthesis.

### Orchestration

The pipeline is orchestrated in **`stores/crossdoc.svelte.ts`** (frontend state management) and **`routes/api/crossdoc/*`** (BFF endpoints). The `pipeline.ts` utilities are used by both the store and the API routes.
