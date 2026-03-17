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
      use_rag_server: true
```

### Profile Override Example

```yaml
# config/profiles/workstation-hybrid.yaml
# Only overrides what changes from defaults

services:
  llm:
    provider: openrouter
    model: anthropic/claude-sonnet-4
    parameters:
      max_tokens: 4096

  crossdoc:
    decomposition:
      provider: openrouter
      model: anthropic/claude-sonnet-4
```

### guardrails.yaml

```yaml
enabled: true
provider: nvidia-api  # Always cloud

content_safety:
  enabled: true
  model: llama-3.1-nemoguard-8b-content-safety

topic_control:
  enabled: true
  model: llama-3.1-nemoguard-8b-topic-control
  allowed_topics:
    - technical documentation
    - industrial equipment
    - engineering calculations
```

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
- `saldivia/__init__.py`
- `saldivia/providers.py`
- `saldivia/config.py`

### Modified Files
- `scripts/deploy.sh` — Call config loader before compose
- `scripts/crossdoc_client.py` — Use SDK for decomposition
- `Makefile` — Add validate-config target, update deploy
- `.gitignore` — Ensure .env.local is ignored

### Deleted Files
- `config/profiles/brev-2gpu.env` — Replaced by YAML
- `config/profiles/workstation-1gpu.env` — Replaced by YAML

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

All the `APP_*` variables the Blueprint expects, derived from models.yaml + profile.

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

| Provider | Chat Completions | Streaming | Tool Use | Notes |
|----------|-----------------|-----------|----------|-------|
| Local NIM | Yes | Yes | No | OpenAI-compatible |
| NVIDIA API | Yes | Yes | No | Requires NVIDIA_API_KEY |
| OpenRouter | Yes | Yes | Yes | Requires extra headers |
| OpenAI | Yes | Yes | Yes | Standard API |
