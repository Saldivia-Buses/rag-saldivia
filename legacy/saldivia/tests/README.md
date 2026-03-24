# saldivia/tests/

Test suite for the RAG Saldivia Python SDK. Covers gateway, auth, config, mode manager, providers, and collections.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `conftest.py` | Pytest configuration: sets `BYPASS_AUTH=true` and `ENVIRONMENT=development` before importing modules | pytest |
| `test_gateway.py` | Tests for gateway.py: FastAPI routes, auth endpoints, JWT validation, RBAC, SSE proxy | pytest, httpx, gateway.py |
| `test_gateway_extended.py` | Extended gateway tests: additional edge cases, error handling, multi-collection scenarios | pytest, httpx, gateway.py |
| `test_auth.py` | Tests for auth/database.py and auth/models.py: user CRUD, API key generation, password hashing, area-collection permissions | pytest, auth/ |
| `test_config.py` | Tests for config.py: ConfigLoader, YAML profile loading, env var merging, validation | pytest, config.py |
| `test_mode_manager.py` | Tests for mode_manager.py: QUERY/INGEST mode switching, VRAM estimation, container management | pytest, unittest.mock, mode_manager.py |
| `test_providers.py` | Tests for providers.py: RAG server client, Milvus client, LLM provider routing | pytest, httpx, providers.py |
| `test_collections.py` | Tests for collections.py: collection CRUD via ingestor API and direct Milvus connection | pytest, httpx, collections.py |

## How to run

```bash
# From repo root
uv run pytest saldivia/tests/ -v

# Single file
uv run pytest saldivia/tests/test_auth.py -v

# With coverage
uv run pytest saldivia/tests/ --cov=saldivia -v
```

All tests use mocked dependencies (httpx, pymilvus, docker SDK) to avoid requiring a running RAG instance. The `conftest.py` fixture sets `BYPASS_AUTH=true` to allow gateway imports without a `JWT_SECRET` environment variable.
