# config/

Configuration files for the RAG Saldivia overlay on NVIDIA RAG Blueprint v2.5.0.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `compose-openrouter-proxy.yaml` | Docker Compose service definition for OpenRouter proxy (optional, for routing LLM calls via OpenRouter) | None |
| `compose-overrides-workstation.yaml` | Compose overrides for 1-GPU workstation deployment: disables LLM service, adds mode manager | compose-overrides.yaml |
| `compose-overrides.yaml` | Saldivia service overrides: adds auth gateway (port 9000), mode manager, ingestion worker | platform.yaml |
| `compose-platform-services.yaml` | Platform services: Milvus vector DB, Redis, NIMs (embedding, reranking, OCR) | None |
| `guardrails.yaml` | Guardrails configuration for LLM output validation (NVIDIA Blueprint feature) | None |
| `milvus.yaml` | Milvus vector database configuration: collection settings, index parameters | None |
| `models.yaml` | Model configuration: LLM, embedding, reranker, OCR model names and endpoints | None |
| `observability.yaml` | Observability configuration: logging, metrics, tracing (NVIDIA Blueprint feature) | None |
| `platform.yaml` | Base platform configuration: RAG server, ingestor, LLM service settings | None |
| `prompt.yaml` | System and RAG prompts for the LLM | None |
| `profiles/` | Deployment profiles: YAML + env files for workstation-1gpu | See profiles/README.md |

## Design Notes

### Compose File Layering Strategy

The Docker Compose deployment uses a multi-layer approach:

1. **Base layer**: `platform.yaml` — NVIDIA Blueprint's base configuration (RAG server, ingestor, LLM service, NIMs)
2. **Saldivia overlay**: `compose-overrides.yaml` — adds auth gateway (port 9000), mode manager, ingestion worker
3. **Deployment-specific overrides**: `compose-overrides-workstation.yaml` — further adjusts for 1-GPU deployments (disables local LLM, configures mode switching)

The `deploy.sh` script merges these layers using `docker compose -f platform.yaml -f compose-overrides.yaml -f compose-overrides-workstation.yaml up`, allowing profile-specific customization without duplicating the base configuration. Environment variables from `profiles/*.env` are loaded before starting services.
