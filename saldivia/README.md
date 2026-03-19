# saldivia/

Python SDK for RAG Saldivia — provides auth gateway, config loader, provider clients, mode management, collection management, and ingestion queue.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `gateway.py` | FastAPI auth gateway with JWT, RBAC, SSE streaming proxy to RAG server | FastAPI, pyjwt, httpx, auth/database.py |
| `config.py` | ConfigLoader: loads YAML profiles, merges env vars, validates deployment config | PyYAML, providers.py |
| `providers.py` | HTTP clients for RAG server, Milvus, LLM providers (OpenRouter, NVIDIA API, OpenAI) | httpx, pymilvus |
| `mode_manager.py` | 1-GPU mode manager: switches between QUERY (NIMs only) and INGEST (NIMs + VLM) | docker SDK |
| `collections.py` | CollectionManager: CRUD operations for Milvus collections via ingestor API | httpx, pymilvus |
| `ingestion_queue.py` | Redis-backed ingestion job queue with status tracking | redis |
| `ingestion_worker.py` | Background worker that processes jobs from ingestion queue | ingestion_queue.py |
| `cache.py` | QueryCache: Redis-backed query result caching with TTL | redis |
| `watch.py` | File watcher for auto-ingestion of new documents in watched directories | watchdog |
| `mcp_server.py` | MCP server for RAG Saldivia (experimental, requires running RAG instance) | mcp SDK |

## Design Notes

### Gateway Timestamp Serialization (`_ts()` helper)

The gateway defines a `_ts(timestamp)` helper function (in `gateway.py`) for safe JSON serialization of timestamp values in HTTP responses. It handles both `datetime` objects (converting via `.isoformat()`) and plain strings, ensuring consistent ISO format timestamps in all API responses.

This is used throughout the gateway's endpoints (e.g., `/auth/me`, `/admin/users`, audit log) to prevent JSON serialization errors when returning timestamp fields from the database.
