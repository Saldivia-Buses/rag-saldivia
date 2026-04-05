# Search Service

Tree-based document search, inspired by PageIndex. Navigates hierarchical
document trees using LLM reasoning, then extracts relevant pages.

## Architecture

3-phase search pipeline:
1. **Navigate** — LLM reads tree TOC (titles + summaries only), selects nodes
2. **Extract** — Code maps nodes → page ranges, reads from `document_pages`
3. **Return** — Assembles selections with citations (document, pages, sections)

## Endpoints

| Path | Method | Description |
|---|---|---|
| `/health` | GET | Health check |
| `/v1/search/query` | POST | Search documents |

### POST /v1/search/query

```json
{
  "query": "medida del disco de freno delantero",
  "collection_id": "optional-collection-id",
  "max_nodes": 5
}
```

## Environment

| Variable | Required | Description |
|---|---|---|
| `SEARCH_PORT` | No | Default: `8010` |
| `POSTGRES_TENANT_URL` | Yes | Tenant DB connection |
| `JWT_PUBLIC_KEY` | Yes | Ed25519 public key (base64) |
| `SGLANG_LLM_URL` | No | Default: `http://localhost:8102` |
| `SGLANG_LLM_MODEL` | No | Model ID for tree navigation |
