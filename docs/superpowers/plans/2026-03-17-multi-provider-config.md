# Multi-Provider Configuration Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add per-service LLM provider routing with OpenRouter proxy, guardrails, and observability to RAG Saldivia.

**Architecture:** YAML configs → config_loader.py → .env.merged → Blueprint. SDK provides `ProviderClient` for crossdoc and scripts. OpenRouter proxy enables any model for Blueprint LLM.

**Tech Stack:** Python 3.11+, PyYAML, httpx, FastAPI (proxy), pytest

**Prerequisites:**
- Install test dependencies: `pip install pytest pyyaml httpx`
- Blueprint must be cloned at `./blueprint/` (done by setup.sh)

---

## File Structure

### New Files

```
saldivia/                          # SDK package
├── __init__.py                    # Package init, re-exports
├── providers.py                   # ProviderClient, ModelConfig
├── config.py                      # ConfigLoader, env generation
└── tests/
    ├── __init__.py
    ├── test_providers.py          # ProviderClient unit tests
    └── test_config.py             # ConfigLoader unit tests

config/
├── models.yaml                    # Service definitions
├── guardrails.yaml                # Guardrails config
├── observability.yaml             # Observability stack config
├── profiles/
│   ├── brev-2gpu.yaml             # All local profile
│   ├── workstation-hybrid.yaml    # Hybrid API+local profile
│   └── full-cloud.yaml            # Everything via API
├── compose-guardrails-cloud.yaml  # Guardrails compose override
├── compose-observability.yaml     # Observability compose override
└── compose-openrouter-proxy.yaml  # Proxy compose override

services/openrouter-proxy/
├── proxy.py                       # FastAPI proxy (~40 lines)
├── Dockerfile                     # Container build
└── requirements.txt               # httpx, fastapi, uvicorn
```

### Modified Files

```
scripts/deploy.sh                  # Add config loader call
scripts/crossdoc_client.py         # Use SDK for decomposition
Makefile                           # Add validate-config target
.gitignore                         # Already correct
```

### Deleted Files

```
config/profiles/brev-2gpu.env      # Replaced by YAML
config/profiles/workstation-1gpu.env  # Replaced by YAML
```

---

## Task 0: GitHub Housekeeping

**Files:**
- None (git operations only)

- [ ] **Step 1: Push existing commits**

```bash
git push origin main
```

Expected: 2 commits pushed (spec commits)

- [ ] **Step 2: Create feature branch**

```bash
git checkout -b feat/multi-provider-config
```

- [ ] **Step 3: Verify branch**

```bash
git branch --show-current
```

Expected: `feat/multi-provider-config`

---

## Task 1: SDK Foundation — ModelConfig and ProviderClient

**Files:**
- Create: `saldivia/__init__.py`
- Create: `saldivia/providers.py`
- Create: `saldivia/tests/__init__.py`
- Create: `saldivia/tests/test_providers.py`

- [ ] **Step 1: Create package structure**

```bash
mkdir -p saldivia/tests
touch saldivia/__init__.py saldivia/tests/__init__.py
```

- [ ] **Step 2: Write failing test for ModelConfig**

```python
# saldivia/tests/test_providers.py
import pytest
from saldivia.providers import ModelConfig

def test_model_config_defaults():
    cfg = ModelConfig(provider="local", model="test-model")
    assert cfg.provider == "local"
    assert cfg.model == "test-model"
    assert cfg.temperature == 0.1
    assert cfg.max_tokens == 2048
    assert cfg.extra_headers == {}

def test_model_config_with_endpoint():
    cfg = ModelConfig(
        provider="nvidia-api",
        model="nvidia/nemotron",
        endpoint="https://api.nvidia.com/v1",
        api_key="test-key"
    )
    assert cfg.endpoint == "https://api.nvidia.com/v1"
    assert cfg.api_key == "test-key"
```

- [ ] **Step 3: Run test to verify it fails**

Run: `python -m pytest saldivia/tests/test_providers.py -v`
Expected: FAIL with "cannot import name 'ModelConfig'"

- [ ] **Step 4: Implement ModelConfig**

```python
# saldivia/providers.py
"""Provider SDK for RAG Saldivia."""
from dataclasses import dataclass, field
from typing import Optional, Iterator
import os

@dataclass
class ModelConfig:
    """Configuration for a model/provider."""
    provider: str  # local | nvidia-api | openrouter | openai | openrouter-proxy
    model: str
    endpoint: Optional[str] = None
    api_key: Optional[str] = None
    temperature: float = 0.1
    max_tokens: int = 2048
    extra_headers: dict = field(default_factory=dict)
```

- [ ] **Step 5: Run test to verify it passes**

Run: `python -m pytest saldivia/tests/test_providers.py -v`
Expected: PASS

- [ ] **Step 6: Write failing test for ProviderClient**

```python
# saldivia/tests/test_providers.py (append)
from saldivia.providers import ProviderClient
from unittest.mock import patch, MagicMock

def test_provider_client_init_local():
    cfg = ModelConfig(provider="local", model="test", endpoint="http://localhost:8000")
    client = ProviderClient(cfg)
    assert client.config == cfg
    assert client.base_url == "http://localhost:8000"

def test_provider_client_init_openrouter():
    cfg = ModelConfig(provider="openrouter", model="anthropic/claude-sonnet-4")
    with patch.dict(os.environ, {"OPENROUTER_API_KEY": "test-key"}):
        client = ProviderClient(cfg)
    assert client.base_url == "https://openrouter.ai/api/v1"
    assert "HTTP-Referer" in client.headers

def test_provider_client_chat_sync():
    cfg = ModelConfig(provider="local", model="test", endpoint="http://localhost:8000")
    client = ProviderClient(cfg)

    mock_response = MagicMock()
    mock_response.json.return_value = {
        "choices": [{"message": {"content": "Hello!"}}]
    }

    with patch("httpx.Client.post", return_value=mock_response):
        result = client.chat_sync([{"role": "user", "content": "Hi"}])

    assert result == "Hello!"
```

- [ ] **Step 7: Run test to verify it fails**

Run: `python -m pytest saldivia/tests/test_providers.py::test_provider_client_init_local -v`
Expected: FAIL with "cannot import name 'ProviderClient'"

- [ ] **Step 8: Implement ProviderClient**

```python
# saldivia/providers.py (append after ModelConfig)
import json
import httpx

PROVIDER_URLS = {
    "nvidia-api": "https://integrate.api.nvidia.com/v1",
    "openrouter": "https://openrouter.ai/api/v1",
    "openai": "https://api.openai.com/v1",
}

PROVIDER_KEY_ENVS = {
    "nvidia-api": "NVIDIA_API_KEY",
    "openrouter": "OPENROUTER_API_KEY",
    "openai": "OPENAI_API_KEY",
}

class ProviderClient:
    """Unified client for any LLM provider."""

    def __init__(self, config: ModelConfig):
        self.config = config

        # Determine base URL
        if config.endpoint:
            self.base_url = config.endpoint.rstrip("/")
        elif config.provider in PROVIDER_URLS:
            self.base_url = PROVIDER_URLS[config.provider]
        else:
            raise ValueError(f"Unknown provider: {config.provider}")

        # Determine API key
        if config.api_key:
            self.api_key = config.api_key
        elif config.provider in PROVIDER_KEY_ENVS:
            self.api_key = os.environ.get(PROVIDER_KEY_ENVS[config.provider])
        else:
            self.api_key = None

        # Build headers
        self.headers = {"Content-Type": "application/json"}
        if self.api_key:
            self.headers["Authorization"] = f"Bearer {self.api_key}"

        # OpenRouter requires extra headers
        if config.provider == "openrouter":
            self.headers["HTTP-Referer"] = "https://rag-saldivia.local"
            self.headers["X-Title"] = "RAG Saldivia"

        # Merge custom headers
        self.headers.update(config.extra_headers)

    def chat_sync(self, messages: list[dict]) -> str:
        """Synchronous chat completion."""
        payload = {
            "model": self.config.model,
            "messages": messages,
            "temperature": self.config.temperature,
            "max_tokens": self.config.max_tokens,
        }

        with httpx.Client(timeout=120) as client:
            resp = client.post(
                f"{self.base_url}/chat/completions",
                headers=self.headers,
                json=payload
            )
            resp.raise_for_status()
            return resp.json()["choices"][0]["message"]["content"]

    def chat(self, messages: list[dict], stream: bool = True) -> Iterator[str]:
        """Streaming chat completion."""
        payload = {
            "model": self.config.model,
            "messages": messages,
            "temperature": self.config.temperature,
            "max_tokens": self.config.max_tokens,
            "stream": stream,
        }

        with httpx.Client(timeout=120) as client:
            with client.stream(
                "POST",
                f"{self.base_url}/chat/completions",
                headers=self.headers,
                json=payload
            ) as resp:
                resp.raise_for_status()
                for line in resp.iter_lines():
                    if line.startswith("data: ") and line != "data: [DONE]":
                        chunk = json.loads(line[6:])
                        if chunk["choices"][0].get("delta", {}).get("content"):
                            yield chunk["choices"][0]["delta"]["content"]
```

- [ ] **Step 9: Run all provider tests**

Run: `python -m pytest saldivia/tests/test_providers.py -v`
Expected: PASS (3 tests)

- [ ] **Step 10: Update package __init__.py**

```python
# saldivia/__init__.py
"""RAG Saldivia SDK."""
from saldivia.providers import ModelConfig, ProviderClient

__all__ = ["ModelConfig", "ProviderClient"]
```

- [ ] **Step 11: Commit**

```bash
git add saldivia/
git commit -m "feat(sdk): add ModelConfig and ProviderClient"
```

---

## Task 2: SDK — ConfigLoader

**Files:**
- Create: `saldivia/config.py`
- Create: `saldivia/tests/test_config.py`
- Modify: `saldivia/__init__.py`

- [ ] **Step 1: Write failing test for ConfigLoader.load()**

```python
# saldivia/tests/test_config.py
import pytest
import tempfile
import os
from pathlib import Path
from saldivia.config import ConfigLoader

@pytest.fixture
def config_dir(tmp_path):
    """Create a temporary config directory with test files."""
    # models.yaml
    models = tmp_path / "models.yaml"
    models.write_text("""
providers:
  local: {}
  nvidia-api:
    base_url: https://integrate.api.nvidia.com/v1
    api_key_env: NVIDIA_API_KEY

services:
  llm:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron
    parameters:
      temperature: 0.1
      max_tokens: 2048
""")

    # profiles directory
    profiles = tmp_path / "profiles"
    profiles.mkdir()

    # brev-2gpu.yaml profile (empty - uses defaults)
    (profiles / "brev-2gpu.yaml").write_text("{}")

    # workstation-hybrid.yaml profile
    (profiles / "workstation-hybrid.yaml").write_text("""
services:
  llm:
    provider: nvidia-api
    model: nvidia/llama-3.3-nemotron-super-49b-v1.5
""")

    return tmp_path

def test_config_loader_load_default(config_dir):
    loader = ConfigLoader(str(config_dir))
    config = loader.load()

    assert config["services"]["llm"]["provider"] == "local"
    assert config["services"]["llm"]["model"] == "nvidia/nemotron"

def test_config_loader_load_with_profile(config_dir):
    loader = ConfigLoader(str(config_dir))
    config = loader.load(profile="workstation-hybrid")

    # Profile overrides
    assert config["services"]["llm"]["provider"] == "nvidia-api"
    assert config["services"]["llm"]["model"] == "nvidia/llama-3.3-nemotron-super-49b-v1.5"
    # Base values preserved
    assert config["services"]["llm"]["endpoint"] == "nim-llm:8000"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `python -m pytest saldivia/tests/test_config.py::test_config_loader_load_default -v`
Expected: FAIL with "cannot import name 'ConfigLoader'"

- [ ] **Step 3: Implement ConfigLoader.load()**

```python
# saldivia/config.py
"""Configuration loader for RAG Saldivia."""
import os
from pathlib import Path
from typing import Optional
import yaml

from saldivia.providers import ModelConfig

def deep_merge(base: dict, override: dict) -> dict:
    """Deep merge two dicts, override wins on conflicts."""
    result = base.copy()
    for key, value in override.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = value
    return result

class ConfigLoader:
    """Loads and merges configuration from YAMLs."""

    def __init__(self, config_dir: str = "config"):
        self.config_dir = Path(config_dir)

    def load(self, profile: str = None) -> dict:
        """Load configuration, optionally with profile overrides."""
        config = {}

        # Load base configs
        for name in ["models", "guardrails", "observability"]:
            path = self.config_dir / f"{name}.yaml"
            if path.exists():
                with open(path) as f:
                    data = yaml.safe_load(f) or {}
                    config = deep_merge(config, data)

        # Load profile overrides
        if profile:
            profile_path = self.config_dir / "profiles" / f"{profile}.yaml"
            if profile_path.exists():
                with open(profile_path) as f:
                    override = yaml.safe_load(f) or {}
                    config = deep_merge(config, override)

        return config

    def get_service(self, name: str, profile: str = None) -> ModelConfig:
        """Get ModelConfig for a specific service."""
        config = self.load(profile)
        service = config.get("services", {}).get(name, {})

        return ModelConfig(
            provider=service.get("provider", "local"),
            model=service.get("model", ""),
            endpoint=service.get("endpoint"),
            temperature=service.get("parameters", {}).get("temperature", 0.1),
            max_tokens=service.get("parameters", {}).get("max_tokens", 2048),
        )
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `python -m pytest saldivia/tests/test_config.py -v`
Expected: PASS (2 tests)

- [ ] **Step 5: Write failing test for generate_env()**

```python
# saldivia/tests/test_config.py (append)

def test_config_loader_generate_env(config_dir):
    loader = ConfigLoader(str(config_dir))
    env = loader.generate_env()

    assert env["APP_LLM_SERVERURL"] == "nim-llm:8000"
    assert env["APP_LLM_MODELNAME"] == "nvidia/nemotron"
    assert env["LLM_TEMPERATURE"] == "0.1"
    assert env["LLM_MAX_TOKENS"] == "2048"

def test_config_loader_generate_env_with_profile(config_dir):
    loader = ConfigLoader(str(config_dir))
    env = loader.generate_env(profile="workstation-hybrid")

    assert env["APP_LLM_MODELNAME"] == "nvidia/llama-3.3-nemotron-super-49b-v1.5"
```

- [ ] **Step 6: Run test to verify it fails**

Run: `python -m pytest saldivia/tests/test_config.py::test_config_loader_generate_env -v`
Expected: FAIL with "has no attribute 'generate_env'"

- [ ] **Step 7: Implement generate_env()**

```python
# saldivia/config.py (add to ConfigLoader class)

    # Mapping from YAML paths to env var names
    ENV_MAPPING = {
        ("services", "llm", "endpoint"): "APP_LLM_SERVERURL",
        ("services", "llm", "model"): "APP_LLM_MODELNAME",
        ("services", "llm", "parameters", "temperature"): "LLM_TEMPERATURE",
        ("services", "llm", "parameters", "max_tokens"): "LLM_MAX_TOKENS",
        ("services", "embeddings", "endpoint"): "APP_EMBEDDINGS_SERVERURL",
        ("services", "embeddings", "model"): "APP_EMBEDDINGS_MODELNAME",
        ("services", "reranker", "endpoint"): "APP_RANKING_SERVERURL",
        ("services", "reranker", "model"): "APP_RANKING_MODELNAME",
        ("services", "query_rewriter", "enabled"): "ENABLE_QUERYREWRITER",
        ("services", "query_rewriter", "endpoint"): "APP_QUERYREWRITER_SERVERURL",
        ("services", "filter_generator", "enabled"): "ENABLE_FILTER_GENERATOR",
        ("services", "filter_generator", "endpoint"): "APP_FILTEREXPRESSIONGENERATOR_SERVERURL",
        ("services", "filter_generator", "model"): "APP_FILTEREXPRESSIONGENERATOR_MODELNAME",
        ("services", "summarizer", "endpoint"): "SUMMARY_LLM_SERVERURL",
        ("services", "summarizer", "model"): "SUMMARY_LLM",
        ("services", "vlm", "endpoint"): "APP_VLM_SERVERURL",
        ("services", "vlm", "model"): "APP_VLM_MODELNAME",
        ("observability", "opentelemetry", "endpoint"): "OTEL_EXPORTER_OTLP_ENDPOINT",
        ("guardrails", "enabled"): "ENABLE_GUARDRAILS",
        ("guardrails", "config_id"): "DEFAULT_CONFIG",
    }

    def _get_nested(self, data: dict, keys: tuple):
        """Get nested value from dict."""
        for key in keys:
            if isinstance(data, dict):
                data = data.get(key)
            else:
                return None
        return data

    def generate_env(self, profile: str = None) -> dict:
        """Generate environment variables dict from config."""
        config = self.load(profile)
        env = {}

        for yaml_path, env_var in self.ENV_MAPPING.items():
            value = self._get_nested(config, yaml_path)
            if value is not None:
                env[env_var] = str(value)

        # Handle observability.enabled -> OTEL_SDK_DISABLED (inverted)
        obs_enabled = self._get_nested(config, ("observability", "enabled"))
        if obs_enabled is False:
            env["OTEL_SDK_DISABLED"] = "true"

        # Handle API keys based on provider
        llm_provider = self._get_nested(config, ("services", "llm", "provider"))
        if llm_provider == "nvidia-api":
            nvidia_key = os.environ.get("NVIDIA_API_KEY")
            if nvidia_key:
                env["APP_LLM_APIKEY"] = nvidia_key

        return env

    def write_env_file(self, path: str, profile: str = None):
        """Write .env file from config."""
        env = self.generate_env(profile)
        with open(path, "w") as f:
            for key, value in sorted(env.items()):
                f.write(f"{key}={value}\n")
```

- [ ] **Step 8: Run all config tests**

Run: `python -m pytest saldivia/tests/test_config.py -v`
Expected: PASS (4 tests)

- [ ] **Step 9: Add validate_config() function**

```python
# saldivia/config.py (add at end of file)

def validate_config(config: dict) -> list[str]:
    """Validate configuration, return list of errors."""
    errors = []

    services = config.get("services", {})

    # Check required services
    for svc in ["llm", "embeddings", "reranker"]:
        if svc not in services:
            errors.append(f"Missing required service: {svc}")
        elif not services[svc].get("model"):
            errors.append(f"Service '{svc}' missing 'model' field")

    # Check provider validity
    valid_providers = {"local", "nvidia-api", "openrouter", "openai", "openrouter-proxy"}
    for name, svc in services.items():
        provider = svc.get("provider", "local")
        if provider not in valid_providers:
            errors.append(f"Service '{name}' has invalid provider: {provider}")

    # Check crossdoc config
    crossdoc = services.get("crossdoc", {})
    if crossdoc:
        decomp = crossdoc.get("decomposition", {})
        if decomp and not decomp.get("model"):
            errors.append("crossdoc.decomposition missing 'model' field")

        synth = crossdoc.get("synthesis", {})
        if synth and not synth.get("use_rag_server", True):
            if not synth.get("model"):
                errors.append("crossdoc.synthesis with use_rag_server=false requires 'model' field")

    return errors
```

- [ ] **Step 10: Add test for validate_config()**

```python
# saldivia/tests/test_config.py (append)

from saldivia.config import validate_config

def test_validate_config_valid(config_dir):
    loader = ConfigLoader(str(config_dir))
    config = loader.load()
    errors = validate_config(config)
    assert errors == []

def test_validate_config_missing_model():
    config = {"services": {"llm": {"provider": "local"}}}  # Missing model
    errors = validate_config(config)
    assert "missing 'model' field" in errors[0]

def test_validate_config_invalid_provider():
    config = {"services": {"llm": {"provider": "invalid", "model": "test"}}}
    errors = validate_config(config)
    assert "invalid provider" in errors[0]
```

- [ ] **Step 11: Run validation tests**

Run: `python -m pytest saldivia/tests/test_config.py -v`
Expected: PASS (7 tests)

- [ ] **Step 12: Update package __init__.py**

```python
# saldivia/__init__.py
"""RAG Saldivia SDK."""
from saldivia.providers import ModelConfig, ProviderClient
from saldivia.config import ConfigLoader, validate_config

__all__ = ["ModelConfig", "ProviderClient", "ConfigLoader", "validate_config"]
```

- [ ] **Step 13: Commit**

```bash
git add saldivia/
git commit -m "feat(sdk): add ConfigLoader with env generation and validation"
```

---

## Task 3: Config YAML Files

**Files:**
- Create: `config/models.yaml`
- Create: `config/guardrails.yaml`
- Create: `config/observability.yaml`
- Create: `config/profiles/brev-2gpu.yaml`
- Create: `config/profiles/workstation-hybrid.yaml`
- Create: `config/profiles/full-cloud.yaml`
- Delete: `config/profiles/brev-2gpu.env`
- Delete: `config/profiles/workstation-1gpu.env`

- [ ] **Step 1: Create models.yaml**

```yaml
# config/models.yaml
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
    enabled: false

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

- [ ] **Step 2: Create guardrails.yaml**

```yaml
# config/guardrails.yaml
enabled: false
provider: nvidia-api
config_id: nemoguard_cloud
```

- [ ] **Step 3: Create observability.yaml**

```yaml
# config/observability.yaml
enabled: false

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

- [ ] **Step 4: Create brev-2gpu.yaml profile**

```yaml
# config/profiles/brev-2gpu.yaml
# All local deployment (default settings)
# No overrides needed - uses models.yaml defaults
{}
```

- [ ] **Step 5: Create workstation-hybrid.yaml profile**

```yaml
# config/profiles/workstation-hybrid.yaml
# LLM via NVIDIA API, NIMs local

services:
  llm:
    provider: nvidia-api
    model: nvidia/llama-3.3-nemotron-super-49b-v1.5
    parameters:
      max_tokens: 4096

  crossdoc:
    decomposition:
      provider: openrouter
      model: anthropic/claude-sonnet-4
```

- [ ] **Step 6: Create full-cloud.yaml profile**

```yaml
# config/profiles/full-cloud.yaml
# Everything via API (no local GPUs needed)

services:
  llm:
    provider: nvidia-api
    model: nvidia/llama-3.3-nemotron-super-49b-v1.5

  embeddings:
    provider: nvidia-api
    endpoint: https://integrate.api.nvidia.com/v1
    model: nvidia/nv-embedqa-e5-v5

  reranker:
    provider: nvidia-api
    endpoint: https://integrate.api.nvidia.com/v1
    model: nvidia/nv-rerankqa-mistral-4b-v3

  crossdoc:
    decomposition:
      provider: openrouter
      model: anthropic/claude-sonnet-4
    synthesis:
      use_rag_server: false
      provider: openrouter
      model: anthropic/claude-sonnet-4

guardrails:
  enabled: true

observability:
  enabled: true
```

- [ ] **Step 7: Delete old .env profiles**

```bash
rm config/profiles/brev-2gpu.env config/profiles/workstation-1gpu.env
```

- [ ] **Step 8: Verify config loads**

```bash
python -c "from saldivia import ConfigLoader; c = ConfigLoader('config'); print(c.load('workstation-hybrid')['services']['llm'])"
```

Expected: Shows LLM config with nvidia-api provider

- [ ] **Step 9: Commit**

```bash
git add config/
git commit -m "feat(config): add YAML configs and profiles

- models.yaml with all 7 services
- guardrails.yaml and observability.yaml
- 3 profiles: brev-2gpu, workstation-hybrid, full-cloud
- Remove old .env profiles"
```

---

## Task 4: OpenRouter Proxy Service

**Files:**
- Create: `services/openrouter-proxy/proxy.py`
- Create: `services/openrouter-proxy/Dockerfile`
- Create: `services/openrouter-proxy/requirements.txt`
- Create: `config/compose-openrouter-proxy.yaml`

- [ ] **Step 1: Create proxy directory**

```bash
mkdir -p services/openrouter-proxy
```

- [ ] **Step 2: Create requirements.txt**

```
# services/openrouter-proxy/requirements.txt
fastapi>=0.109.0
uvicorn>=0.27.0
httpx>=0.26.0
```

- [ ] **Step 3: Create proxy.py**

```python
# services/openrouter-proxy/proxy.py
"""OpenRouter proxy that adds required headers."""
import os
from fastapi import FastAPI, Request, Response
import httpx

app = FastAPI(title="OpenRouter Proxy")

OPENROUTER_URL = os.getenv("OPENROUTER_URL", "https://openrouter.ai/api/v1")
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY", "")
HEADERS = {
    "HTTP-Referer": os.getenv("HTTP_REFERER", "https://rag-saldivia.local"),
    "X-Title": os.getenv("X_TITLE", "RAG Saldivia"),
}

@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE"])
async def proxy(path: str, request: Request):
    """Proxy requests to OpenRouter with required headers."""
    # Build headers
    headers = dict(request.headers)
    headers.update(HEADERS)

    # Add API key
    if OPENROUTER_API_KEY:
        headers["Authorization"] = f"Bearer {OPENROUTER_API_KEY}"

    # Remove host header (will be set by httpx)
    headers.pop("host", None)

    async with httpx.AsyncClient(timeout=120) as client:
        resp = await client.request(
            method=request.method,
            url=f"{OPENROUTER_URL}/{path}",
            headers=headers,
            content=await request.body(),
        )

        # Forward response
        return Response(
            content=resp.content,
            status_code=resp.status_code,
            headers=dict(resp.headers),
        )

@app.get("/health")
async def health():
    return {"status": "ok"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
```

- [ ] **Step 4: Create Dockerfile**

```dockerfile
# services/openrouter-proxy/Dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY proxy.py .

EXPOSE 8080

CMD ["uvicorn", "proxy:app", "--host", "0.0.0.0", "--port", "8080"]
```

- [ ] **Step 5: Create compose-openrouter-proxy.yaml**

```yaml
# config/compose-openrouter-proxy.yaml
services:
  openrouter-proxy:
    build:
      context: ../services/openrouter-proxy
      dockerfile: Dockerfile
    environment:
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  default:
    name: nvidia-rag
    external: true
```

- [ ] **Step 6: Test proxy locally**

```bash
cd services/openrouter-proxy
pip install -r requirements.txt
python -c "from proxy import app; print('Proxy imports OK')"
```

Expected: "Proxy imports OK"

- [ ] **Step 7: Commit**

```bash
git add services/ config/compose-openrouter-proxy.yaml
git commit -m "feat(proxy): add OpenRouter proxy service

Enables using any OpenRouter model with Blueprint LLM by
adding required HTTP-Referer and X-Title headers."
```

---

## Task 5: Compose Integration Files

**Files:**
- Create: `config/compose-guardrails-cloud.yaml`
- Create: `config/compose-observability.yaml`

- [ ] **Step 1: Create compose-guardrails-cloud.yaml**

```yaml
# config/compose-guardrails-cloud.yaml
# NeMo Guardrails via NVIDIA API (no local GPU needed)
services:
  nemo-guardrails:
    image: nvcr.io/nvidia/nemo-guardrails:0.11.0
    environment:
      - NVIDIA_API_KEY=${NVIDIA_API_KEY}
      - DEFAULT_CONFIG=${DEFAULT_CONFIG:-nemoguard_cloud}
    ports:
      - "8085:8085"
    volumes:
      - ../blueprint/deploy/compose/nemo-guardrails/config-store:/app/config-store:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8085/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  default:
    name: nvidia-rag
    external: true
```

- [ ] **Step 2: Create compose-observability.yaml**

```yaml
# config/compose-observability.yaml
# Observability stack - copied from Blueprint with port customization
services:
  otel-collector:
    image: otel/opentelemetry-collector-contrib:0.140.0
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ../blueprint/deploy/config/otel-collector-config.yaml:/etc/otel-collector-config.yaml:ro
    ports:
      - "4317:4317"
      - "4318:4318"
    depends_on:
      - zipkin

  zipkin:
    image: openzipkin/zipkin:3.5.0
    environment:
      JAVA_OPTS: "-Xms2g -Xmx4g -XX:+ExitOnOutOfMemoryError"
    ports:
      - "${ZIPKIN_PORT:-9411}:9411"

  prometheus:
    image: prom/prometheus:latest
    command:
      - --config.file=/etc/prometheus/prometheus-config.yaml
      - --storage.tsdb.retention.time=${PROMETHEUS_RETENTION:-1h}
      - --web.enable-lifecycle
    volumes:
      - ../blueprint/deploy/config/prometheus.yaml:/etc/prometheus/prometheus-config.yaml:ro
    ports:
      - "${PROMETHEUS_PORT:-9090}:9090"

  grafana:
    image: grafana/grafana:latest
    ports:
      - "${GRAFANA_PORT:-3000}:3000"
    volumes:
      - ../blueprint/deploy/config/grafana.yaml:/etc/grafana/provisioning/datasources/grafana.yaml:ro

networks:
  default:
    name: nvidia-rag
    external: true
```

- [ ] **Step 3: Commit**

```bash
git add config/compose-*.yaml
git commit -m "feat(compose): add guardrails and observability compose files"
```

---

## Task 6: Deploy Script Integration

**Files:**
- Modify: `scripts/deploy.sh`
- Modify: `Makefile`

- [ ] **Step 1: Read current deploy.sh**

```bash
head -50 scripts/deploy.sh
```

- [ ] **Step 2: Update deploy.sh to use config loader**

Add after the profile parsing section (around line 30):

```bash
# --- Add this section after PROFILE is set ---

# Add repo root to Python path for saldivia imports
export PYTHONPATH="${PYTHONPATH}:$(pwd)"

# Generate .env.merged from YAML config
echo "Generating environment from config..."
python3 -c "
from saldivia import ConfigLoader
loader = ConfigLoader('config')
loader.write_env_file('.env.merged', profile='${PROFILE}')
print('Generated .env.merged')
"

# Determine which compose files to include
COMPOSE_FILES="-f blueprint/deploy/compose/docker-compose-rag-server.yaml"
COMPOSE_FILES="$COMPOSE_FILES -f config/compose-overrides.yaml"

# Check if proxy is needed
if grep -q "provider: openrouter-proxy" config/models.yaml config/profiles/${PROFILE}.yaml 2>/dev/null; then
    echo "Including OpenRouter proxy..."
    COMPOSE_FILES="$COMPOSE_FILES -f config/compose-openrouter-proxy.yaml"
fi

# Check if guardrails enabled
if python3 -c "from saldivia import ConfigLoader; c = ConfigLoader('config').load('${PROFILE}'); exit(0 if c.get('guardrails',{}).get('enabled') else 1)" 2>/dev/null; then
    echo "Including guardrails..."
    COMPOSE_FILES="$COMPOSE_FILES -f config/compose-guardrails-cloud.yaml"
fi

# Check if observability enabled
if python3 -c "from saldivia import ConfigLoader; c = ConfigLoader('config').load('${PROFILE}'); exit(0 if c.get('observability',{}).get('enabled') else 1)" 2>/dev/null; then
    echo "Including observability..."
    COMPOSE_FILES="$COMPOSE_FILES -f config/compose-observability.yaml"
fi
```

- [ ] **Step 3: Update Makefile**

```makefile
# Add to Makefile

.PHONY: validate-config
validate-config:
	@python3 -c "from saldivia import ConfigLoader; \
		loader = ConfigLoader('config'); \
		config = loader.load('$(PROFILE)'); \
		print('Config loaded successfully'); \
		env = loader.generate_env('$(PROFILE)'); \
		print(f'Generated {len(env)} env vars')"

.PHONY: show-env
show-env:
	@python3 -c "from saldivia import ConfigLoader; \
		loader = ConfigLoader('config'); \
		env = loader.generate_env('$(PROFILE)'); \
		for k,v in sorted(env.items()): print(f'{k}={v}')"
```

- [ ] **Step 4: Test config validation**

```bash
make validate-config PROFILE=workstation-hybrid
```

Expected: "Config loaded successfully" + env var count

- [ ] **Step 5: Commit**

```bash
git add scripts/deploy.sh Makefile
git commit -m "feat(deploy): integrate config loader into deploy flow

- deploy.sh generates .env.merged from YAML config
- Conditionally includes proxy/guardrails/observability compose files
- Add validate-config and show-env Makefile targets"
```

---

## Task 7: Crossdoc Client SDK Integration

**Files:**
- Modify: `scripts/crossdoc_client.py`

- [ ] **Step 1: Read current crossdoc_client.py structure**

```bash
head -100 scripts/crossdoc_client.py
```

- [ ] **Step 2: Add SDK import and CrossdocClient class**

Add near the top of the file after existing imports:

```python
# SDK integration (optional - falls back to direct RAG calls if not available)
try:
    from saldivia import ConfigLoader, ProviderClient, ModelConfig
    SDK_AVAILABLE = True
except ImportError:
    SDK_AVAILABLE = False
    ConfigLoader = None
```

- [ ] **Step 3: Add profile-aware initialization**

Add a new class or modify existing code to support profile-based configuration:

```python
class CrossdocConfig:
    """Configuration for crossdoc client."""

    def __init__(self, profile: str = None):
        self.profile = profile
        self.decomp_client = None
        self.synth_client = None

        if SDK_AVAILABLE and profile:
            loader = ConfigLoader("config")
            config = loader.load(profile)
            crossdoc = config.get("services", {}).get("crossdoc", {})

            # Setup decomposition client
            decomp = crossdoc.get("decomposition", {})
            if decomp.get("provider") and decomp.get("provider") != "local":
                self.decomp_client = ProviderClient(ModelConfig(
                    provider=decomp["provider"],
                    model=decomp.get("model", ""),
                    temperature=decomp.get("parameters", {}).get("temperature", 0.1),
                    max_tokens=decomp.get("parameters", {}).get("max_tokens", 2048),
                ))

            # Setup synthesis client (if not using RAG server)
            synth = crossdoc.get("synthesis", {})
            if not synth.get("use_rag_server", True):
                self.synth_client = ProviderClient(ModelConfig(
                    provider=synth["provider"],
                    model=synth.get("model", ""),
                    temperature=synth.get("parameters", {}).get("temperature", 0.1),
                    max_tokens=synth.get("parameters", {}).get("max_tokens", 4096),
                ))
```

- [ ] **Step 4: Update decompose function to use SDK client**

Modify the `decompose_question` function to optionally use the SDK client:

```python
def decompose_question(question: str, config: CrossdocConfig = None) -> list[str]:
    """Decompose question into sub-queries using configured provider."""

    # Use SDK client if available
    if config and config.decomp_client:
        messages = [
            {"role": "system", "content": DECOMPOSITION_PROMPT},
            {"role": "user", "content": question}
        ]
        response = config.decomp_client.chat_sync(messages)
        return parse_sub_queries(response)

    # Fall back to RAG server
    # ... existing implementation ...
```

- [ ] **Step 5: Add --profile CLI argument**

Add to argument parser:

```python
parser.add_argument("--profile", type=str, default=None,
                    help="Config profile (e.g., workstation-hybrid)")
```

And in main():

```python
config = CrossdocConfig(profile=args.profile) if args.profile else None
```

- [ ] **Step 6: Test SDK integration**

```bash
python -c "
from scripts.crossdoc_client import CrossdocConfig
config = CrossdocConfig(profile='workstation-hybrid')
print(f'Decomp client: {config.decomp_client}')
print(f'Using provider: {config.decomp_client.config.provider if config.decomp_client else \"RAG server\"}')
"
```

Expected: Shows "Using provider: openrouter"

- [ ] **Step 7: Commit**

```bash
git add scripts/crossdoc_client.py
git commit -m "feat(crossdoc): add SDK integration for provider routing

- Optional SDK import with fallback to direct RAG calls
- CrossdocConfig class for profile-based provider selection
- --profile CLI argument for selecting configuration
- Decomposition can now use OpenRouter or other providers"
```

---

## Task 8: Final Integration and PR

**Files:**
- None (git operations and testing)

- [ ] **Step 1: Run all tests**

```bash
python -m pytest saldivia/tests/ -v
```

Expected: All tests pass

- [ ] **Step 2: Verify config loading**

```bash
make validate-config PROFILE=brev-2gpu
make validate-config PROFILE=workstation-hybrid
make validate-config PROFILE=full-cloud
```

Expected: All profiles load successfully

- [ ] **Step 3: Show env vars for each profile**

```bash
make show-env PROFILE=brev-2gpu | head -10
make show-env PROFILE=workstation-hybrid | head -10
```

Expected: Different configs for each profile

- [ ] **Step 4: Push feature branch**

```bash
git push -u origin feat/multi-provider-config
```

- [ ] **Step 5: Create PR**

```bash
gh pr create --title "feat: multi-provider configuration system" --body "$(cat <<'EOF'
## Summary

- Add per-service LLM provider routing (local, NVIDIA API, OpenRouter, OpenAI)
- YAML-based configuration with profile overrides
- OpenRouter proxy for using any model with Blueprint LLM
- Optional guardrails (cloud) and observability stack
- SDK for provider-agnostic LLM calls (ProviderClient, ConfigLoader)

## Changes

- New: `saldivia/` SDK package with tests
- New: `config/models.yaml`, `guardrails.yaml`, `observability.yaml`
- New: `config/profiles/` with 3 deployment profiles
- New: `services/openrouter-proxy/` for header injection
- Modified: `scripts/deploy.sh` to use config loader
- Modified: `scripts/crossdoc_client.py` with SDK integration

## Test Plan

- [ ] Unit tests pass: `pytest saldivia/tests/ -v`
- [ ] Config validation: `make validate-config PROFILE=workstation-hybrid`
- [ ] Deploy with brev-2gpu profile (local)
- [ ] Deploy with workstation-hybrid profile (API LLM)
- [ ] Test crossdoc with OpenRouter decomposition

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 6: Verify PR created**

```bash
gh pr view --web
```

Expected: Opens PR in browser

---

## Summary

| Task | Files | Estimated Time |
|------|-------|----------------|
| 0. GitHub Housekeeping | git only | 2 min |
| 1. SDK ModelConfig + ProviderClient | 4 files | 15 min |
| 2. SDK ConfigLoader | 2 files | 15 min |
| 3. Config YAML Files | 6 files | 10 min |
| 4. OpenRouter Proxy | 4 files | 10 min |
| 5. Compose Integration | 2 files | 5 min |
| 6. Deploy Script Integration | 2 files | 10 min |
| 7. Crossdoc Client Integration | 1 file | 15 min |
| 8. Final Integration + PR | git only | 10 min |

**Total: ~90 minutes**
