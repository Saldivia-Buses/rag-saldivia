# RAG Saldivia — Multi-Provider Configuration Design

## Overview

Extend RAG Saldivia to support granular, per-service model routing with multiple LLM providers (local NIMs, NVIDIA API, OpenRouter, OpenAI). Add cloud-based guardrails and full observability stack.

## Goals

1. **Per-service configuration** — Each of the 7 RAG services can use a different provider/model
2. **Multiple providers** — Support local deployment, NVIDIA API, OpenRouter, and OpenAI
3. **Profile-based overrides** — Predefined configurations for different deployment scenarios
4. **Cloud guardrails** — NeMo Guardrails via NVIDIA API (no additional GPUs)
5. **Full observability** — OpenTelemetry, Zipkin, Prometheus, Grafana (toggleable)
6. **Configurable crossdoc** — Decomposition and synthesis can use different providers

## Non-Goals

- Modifying NVIDIA RAG Blueprint source code (maintain overlay pattern)
- Local guardrails deployment (requires 2 additional GPUs)
- Custom provider implementations beyond OpenAI-compatible APIs

## Constraints

### Blueprint LLM Provider Limitation

The Blueprint's RAG server only supports local NIMs and NVIDIA API for the core LLM service. It does NOT support arbitrary OpenAI-compatible providers (like OpenRouter or OpenAI directly) because:

1. The Blueprint uses `ChatNVIDIA` from `langchain_nvidia_ai_endpoints` for all LLM calls
2. `ChatNVIDIA` doesn't support custom headers required by OpenRouter (`HTTP-Referer`, `X-Title`)
3. The `_is_nvidia_endpoint()` function only controls NVIDIA-specific parameters (like `min_tokens`), not client selection

**Workaround:** For the `llm` service specifically, use NVIDIA API when local NIM isn't available. OpenRouter and OpenAI can still be used for:
- Crossdoc decomposition/synthesis (via our SDK, not Blueprint)
- Any other custom scripts

This limitation does NOT apply to services we control directly (crossdoc_client.py).

### Solution: OpenRouter Proxy

To enable ANY OpenAI-compatible provider for the Blueprint's LLM service, we add an optional lightweight proxy:

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Blueprint  │────▶│ openrouter-proxy │────▶│   OpenRouter    │
│  RAG Server │     │   (adds headers) │     │   (any model)   │
└─────────────┘     └──────────────────┘     └─────────────────┘
```

**Proxy implementation** (`services/openrouter-proxy/`):

```python
# proxy.py (~40 lines, FastAPI)
from fastapi import FastAPI, Request, Response
import httpx

app = FastAPI()
OPENROUTER_URL = "https://openrouter.ai/api/v1"
HEADERS = {
    "HTTP-Referer": "https://rag-saldivia.local",
    "X-Title": "RAG Saldivia"
}

@app.api_route("/{path:path}", methods=["GET", "POST"])
async def proxy(path: str, request: Request):
    async with httpx.AsyncClient() as client:
        resp = await client.request(
            method=request.method,
            url=f"{OPENROUTER_URL}/{path}",
            headers={**dict(request.headers), **HEADERS},
            content=await request.body(),
        )
        return Response(content=resp.content, status_code=resp.status_code)
```

**Configuration when using proxy:**

```yaml
# models.yaml
services:
  llm:
    provider: openrouter-proxy  # Uses local proxy
    endpoint: openrouter-proxy:8080/v1
    model: anthropic/claude-sonnet-4  # Any OpenRouter model
```

**Compose file** (`compose-openrouter-proxy.yaml`):

```yaml
services:
  openrouter-proxy:
    build: ./services/openrouter-proxy
    environment:
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
    ports:
      - "8080:8080"
```

The proxy is optional — only included when `services.llm.provider: openrouter-proxy` is set.

## Architecture

### Approach: Overlay + SDK

```
config/*.yaml → config_loader.py → .env.merged → Blueprint (unmodified)
                      ↓
              saldivia.providers → crossdoc_client.py, scripts
```

This maintains the overlay pattern, keeping the Blueprint updatable while adding a reusable SDK.

## Configuration Structure

### Directory Layout

```
config/
├── models.yaml              # Service definitions and provider defaults
├── guardrails.yaml          # Guardrails configuration (cloud)
├── observability.yaml       # Observability stack configuration
├── profiles/
│   ├── brev-2gpu.yaml       # All local (2 GPUs)
│   ├── workstation-hybrid.yaml  # LLM via API + local NIMs
│   └── full-cloud.yaml      # Everything via API
├── .env.saldivia            # Non-secret config (committed)
└── compose-overrides.yaml   # Existing compose overrides

.env.local                   # Secrets only (gitignored)
```

### models.yaml Schema

```yaml
providers:
  nvidia-api:
    base_url: https://integrate.api.nvidia.com/v1
    api_key_env: NVIDIA_API_KEY

  openrouter:
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_API_KEY
    headers:
      HTTP-Referer: https://rag-saldivia.local
      X-Title: RAG Saldivia

  openai:
    base_url: https://api.openai.com/v1
    api_key_env: OPENAI_API_KEY

  local:
    # Endpoints defined per-service

services:
  llm:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron-3-super-120b-a12b
    parameters:
      temperature: 0.1
      max_tokens: 2048
      enable_thinking: true
      reasoning_budget: 256

  embeddings:
    provider: local
    endpoint: nemotron-embedding-ms:8000/v1
    model: nvidia/nv-embedqa-e5-v5

  reranker:
    provider: local
    endpoint: nemotron-ranking-ms:8000
    model: nvidia/nv-rerankqa-mistral-4b-v3

  query_rewriter:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron-3-super-120b-a12b
    enabled: false

  filter_generator:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron-3-super-120b-a12b

  summarizer:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron-3-super-120b-a12b

  vlm:
    provider: local
    endpoint: qwen3-vl-8b:8000
    model: qwen3-vl-8b

  crossdoc:
    decomposition:
      provider: local
      model: nvidia/nemotron-3-super-120b-a12b
    synthesis:
      use_rag_server: true  # Uses RAG server's LLM for final answer
      # Alternative: call provider directly (bypasses RAG server for synthesis)
      # use_rag_server: false
      # provider: openrouter
      # model: anthropic/claude-sonnet-4
      # parameters:
      #   max_tokens: 4096
```

### Profile Override Example

```yaml
# config/profiles/workstation-hybrid.yaml
# Only overrides what changes from defaults

services:
  llm:
    # Option A: NVIDIA API (direct, no proxy needed)
    provider: nvidia-api
    model: nvidia/llama-3.3-nemotron-super-49b-v1.5
    # Option B: OpenRouter via proxy (any model)
    # provider: openrouter-proxy
    # model: anthropic/claude-sonnet-4
    parameters:
      max_tokens: 4096

  crossdoc:
    # Crossdoc can use OpenRouter directly (we control the client)
    decomposition:
      provider: openrouter
      model: anthropic/claude-sonnet-4
```

### guardrails.yaml

```yaml
enabled: true
provider: nvidia-api  # Always cloud (NVIDIA API key required)

# Config bundle selection (maps to DEFAULT_CONFIG env var)
config_id: nemoguard_cloud  # Options: nemoguard, nemoguard_cloud
```

**Note:** Guardrails behavior (content safety, topic control, etc.) is configured via NeMo Guardrails config bundles in `deploy/compose/nemo-guardrails/config-store/`. The `config_id` selects which bundle to use. Our config loader only controls on/off and bundle selection — fine-grained guardrail tuning requires modifying the Blueprint's config files directly.

### observability.yaml

```yaml
enabled: true

opentelemetry:
  endpoint: otel-collector:4317

zipkin:
  enabled: true
  port: 9411

prometheus:
  enabled: true
  port: 9090
  retention: 1h

grafana:
  enabled: true
  port: 3000
```

**Generated env vars:**

| YAML Field | Env Var | Description |
|------------|---------|-------------|
| `enabled: false` | `OTEL_SDK_DISABLED=true` | Disables all OpenTelemetry instrumentation |
| `opentelemetry.endpoint` | `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP collector endpoint |
| `zipkin.enabled` | Controls whether Zipkin container is included in compose |
| `prometheus.enabled` | Controls whether Prometheus container is included |
| `prometheus.retention` | `--storage.tsdb.retention.time` flag in Prometheus |
| `grafana.enabled` | Controls whether Grafana container is included |

When `enabled: true`, the config loader includes `compose-observability.yaml` in the compose command.

## Provider SDK

### Module: `saldivia/providers.py`

```python
@dataclass
class ModelConfig:
    provider: str           # local | nvidia-api | openrouter | openai
    model: str
    endpoint: Optional[str] = None
    api_key: Optional[str] = None
    temperature: float = 0.1
    max_tokens: int = 2048
    extra_headers: dict = field(default_factory=dict)

class ProviderClient:
    """Unified client for any LLM provider."""

    def __init__(self, config: ModelConfig): ...
    def chat(self, messages: list[dict], stream: bool = True) -> Iterator[str]: ...
    def chat_sync(self, messages: list[dict]) -> str: ...

class ConfigLoader:
    """Loads and merges configuration from YAMLs."""

    def __init__(self, config_dir: str = "config"): ...
    def load(self, profile: str = None) -> dict: ...
    def get_service(self, name: str, profile: str = None) -> ModelConfig: ...
    def generate_env(self, profile: str = None) -> dict: ...
    def write_env_file(self, path: str, profile: str = None): ...

# Convenience functions
def get_llm(profile: str = None) -> ProviderClient: ...
def get_service(name: str, profile: str = None) -> ProviderClient: ...
```

### Module: `saldivia/config.py`

```python
def load_config(profile: str = None) -> dict:
    """Load full merged configuration."""

def generate_env_merged(profile: str, output_path: str):
    """Generate .env.merged file for docker compose."""

def validate_config(config: dict) -> list[str]:
    """Validate configuration, return list of errors."""
```

## Compose Integration

### deploy.sh Changes

```bash
# Current flow
deploy.sh PROFILE → merge env files → docker compose up

# New flow
deploy.sh PROFILE → config_loader.py generates .env.merged → docker compose up
```

### New Compose Files

- `compose-guardrails-cloud.yaml` — NeMo Guardrails microservice pointing to NVIDIA API
- `compose-observability.yaml` — Full observability stack (already exists in Blueprint)

### Makefile Changes

```makefile
deploy:
    @python3 $(SALDIVIA_ROOT)/saldivia/config.py generate-env $(PROFILE)
    @./scripts/deploy.sh $(PROFILE)

validate-config:
    @python3 $(SALDIVIA_ROOT)/saldivia/config.py validate $(PROFILE)
```

## Crossdoc Client Changes

### Current Behavior
- Always calls RAG server at `localhost:8081`
- Uses RAG server's LLM for decomposition and synthesis

### New Behavior
- Decomposition: Uses configured provider (can be different from RAG server)
- Retrieval: Always via RAG server
- Synthesis: Configurable (RAG server or direct provider call)

```python
# crossdoc_client.py changes

from saldivia.providers import ConfigLoader, ProviderClient

class CrossdocClient:
    def __init__(self, profile: str = None):
        loader = ConfigLoader()
        crossdoc_cfg = loader.load(profile).get("services", {}).get("crossdoc", {})

        # Decomposition can use any provider
        decomp_cfg = crossdoc_cfg.get("decomposition", {})
        self.decomp_client = ProviderClient(ModelConfig(**decomp_cfg))

        # Synthesis either uses RAG server or configured provider
        if crossdoc_cfg.get("synthesis", {}).get("use_rag_server", True):
            self.synth_client = None  # Use RAG server
        else:
            synth_cfg = crossdoc_cfg.get("synthesis", {})
            self.synth_client = ProviderClient(ModelConfig(**synth_cfg))
```

## File Changes Summary

### New Files
- `config/models.yaml`
- `config/guardrails.yaml`
- `config/observability.yaml`
- `config/profiles/brev-2gpu.yaml`
- `config/profiles/workstation-hybrid.yaml`
- `config/profiles/full-cloud.yaml`
- `config/compose-guardrails-cloud.yaml`
- `config/compose-observability.yaml`
- `config/compose-openrouter-proxy.yaml`
- `services/openrouter-proxy/proxy.py`
- `services/openrouter-proxy/Dockerfile`
- `services/openrouter-proxy/requirements.txt`
- `saldivia/__init__.py`
- `saldivia/providers.py`
- `saldivia/config.py`

### Modified Files
- `scripts/deploy.sh` — Call config loader before compose
- `scripts/crossdoc_client.py` — Use SDK for decomposition
- `Makefile` — Add validate-config target, update deploy
- `.gitignore` — Ensure .env.local is ignored

### Deleted Files
- `config/profiles/brev-2gpu.env` — Replaced by `brev-2gpu.yaml`
- `config/profiles/workstation-1gpu.env` — Replaced by `workstation-hybrid.yaml` (renamed: the profile now uses hybrid local+API approach, not just "1 GPU")

## Environment Variables

### Required in .env.local (secrets)

```bash
# At least one of these, depending on providers used
NVIDIA_API_KEY=nvapi-...
OPENROUTER_API_KEY=sk-or-v1-...
OPENAI_API_KEY=sk-...
NGC_API_KEY=...  # For guardrails
```

### Generated in .env.merged (by config_loader)

The config loader translates YAML fields to Blueprint env vars:

| YAML Path | Blueprint Env Var |
|-----------|-------------------|
| `services.llm.endpoint` | `APP_LLM_SERVERURL` |
| `services.llm.model` | `APP_LLM_MODELNAME` |
| `services.llm.parameters.temperature` | `LLM_TEMPERATURE` |
| `services.llm.parameters.max_tokens` | `LLM_MAX_TOKENS` |
| `services.embeddings.endpoint` | `APP_EMBEDDINGS_SERVERURL` |
| `services.embeddings.model` | `APP_EMBEDDINGS_MODELNAME` |
| `services.reranker.endpoint` | `APP_RANKING_SERVERURL` |
| `services.reranker.model` | `APP_RANKING_MODELNAME` |
| `services.query_rewriter.enabled` | `ENABLE_QUERYREWRITER` |
| `services.query_rewriter.endpoint` | `APP_QUERYREWRITER_SERVERURL` |
| `services.filter_generator.enabled` | `ENABLE_FILTER_GENERATOR` |
| `services.filter_generator.endpoint` | `APP_FILTEREXPRESSIONGENERATOR_SERVERURL` |
| `services.filter_generator.model` | `APP_FILTEREXPRESSIONGENERATOR_MODELNAME` |
| `services.summarizer.endpoint` | `SUMMARY_LLM_SERVERURL` |
| `services.summarizer.model` | `SUMMARY_LLM` |
| `services.vlm.endpoint` (RAG) | `APP_VLM_SERVERURL` |
| `services.vlm.endpoint` (ingestion) | `APP_NVINGEST_CAPTIONENDPOINTURL` |
| `services.vlm.model` | `APP_VLM_MODELNAME` / `APP_NVINGEST_CAPTIONMODELNAME` |
| `observability.enabled` | `OTEL_SDK_DISABLED` (inverted) |
| `observability.opentelemetry.endpoint` | `OTEL_EXPORTER_OTLP_ENDPOINT` |
| `guardrails.enabled` | `ENABLE_GUARDRAILS` |

**Note:** Guardrails configuration beyond on/off is controlled via NeMo Guardrails config files, not env vars. The config loader sets `DEFAULT_CONFIG_ID` to select the appropriate config bundle.

For API providers, the loader also sets:
- `APP_LLM_APIKEY` from `NVIDIA_API_KEY` env var (when provider is `nvidia-api`)

## Testing Strategy

1. **Unit tests** for ConfigLoader and ProviderClient
2. **Integration test** — Deploy with each profile, verify services respond
3. **Crossdoc test** — Verify decomposition works with OpenRouter
4. **Guardrails test** — Verify content safety blocks inappropriate input
5. **Observability test** — Verify traces appear in Zipkin

## Migration Path

1. Existing `.env.saldivia` values move to `models.yaml`
2. Profile `.env` files become `.yaml` overrides
3. `deploy.sh` updated to use config loader
4. Backward compatible: if no `models.yaml`, fall back to current behavior

## Open Questions

None — all decisions made during brainstorming.

## Appendix: Provider Compatibility

| Provider | Chat Completions | Streaming | Tool Use | Blueprint LLM | Notes |
|----------|-----------------|-----------|----------|---------------|-------|
| Local NIM | Yes | Yes | No | ✅ Direct | OpenAI-compatible |
| NVIDIA API | Yes | Yes | No | ✅ Direct | Requires NVIDIA_API_KEY |
| OpenRouter | Yes | Yes | Yes | ⚠️ Via proxy | Requires extra headers |
| OpenAI | Yes | Yes | Yes | ⚠️ Via proxy | Could work direct, proxy recommended for consistency |

**Blueprint LLM column:** Whether the provider can be used for the `services.llm` (RAG server's main LLM).
- ✅ Direct — Works out of the box
- ⚠️ Via proxy — Requires `openrouter-proxy` container
