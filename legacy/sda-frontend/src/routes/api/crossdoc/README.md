# Crossdoc API Routes

BFF endpoints for the crossdoc pipeline. These three endpoints are called sequentially by the client to orchestrate a cross-document query.

## Endpoints

| Endpoint | Method | Input | Output | Description |
|----------|--------|-------|--------|-------------|
| `/api/crossdoc/decompose` | POST | `{ question: string, maxSubQueries: number }` | `{ subQueries: string[] }` | Decomposes the user's question into focused sub-queries using an LLM. |
| `/api/crossdoc/subquery` | POST | `{ query: string, collection: string, vdbTopK: number, rerankerTopK: number }` | `{ content: string, sources: Source[] }` | Executes a single sub-query against the RAG system. Returns the LLM's answer and source documents. |
| `/api/crossdoc/synthesize` | POST | `{ question: string, results: SubResult[], synthesisModel?: string }` | `{ synthesis: string }` | Synthesizes the results from all sub-queries into a coherent final answer. |

## Client-driven sequence

The crossdoc pipeline is orchestrated **client-side** by the `CrossdocStore` (see `src/lib/stores/crossdoc.svelte.ts`). The client calls these endpoints in sequence:

```
Client                  BFF (/api/crossdoc/*)         Gateway (port 9000)
  |                            |                              |
  |-- POST /decompose -------->|                              |
  |                            |-- POST /crossdoc/decompose ->|
  |<-- { subQueries: [...] } --|<-- { subQueries: [...] } ----|
  |                            |                              |
  | (for each subQuery):       |                              |
  |-- POST /subquery --------->|                              |
  |                            |-- POST /crossdoc/subquery -->|
  |<-- { content, sources } ---|<-- { content, sources } -----|
  |                            |                              |
  |-- POST /synthesize ------->|                              |
  |                            |-- POST /crossdoc/synthesize->|
  |<-- { synthesis } ----------|<-- { synthesis } ------------|
```

## Design notes

### Why client-driven?

The pipeline is orchestrated client-side (not server-side) to enable:

1. **Real-time progress updates** — The `CrossdocStore` updates `progress` state after each phase, allowing the UI to show live progress indicators via `CrossdocProgress.svelte`.

2. **Parallel batching** — The client executes sub-queries in parallel (up to 6 at a time) using `Promise.allSettled()`, maximizing throughput while avoiding overload.

3. **Early abort** — The client can cancel the pipeline mid-execution (e.g., if the user clicks "Stop") by calling `abortController.abort()`.

### BFF role

Each BFF endpoint is a thin proxy:
- Validates user's JWT cookie
- Proxies request to gateway using `SYSTEM_API_KEY`
- Returns response (no additional processing)

This keeps the BFF stateless and simple, while the client handles orchestration complexity.
