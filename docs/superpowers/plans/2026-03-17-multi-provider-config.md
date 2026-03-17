# RAG Saldivia — Complete Platform Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-ready RAG platform with 1-GPU optimization, multi-provider routing, collection management, MCP integration, and commercial-ready features.

**Architecture:**
```
┌─────────────────────────────────────────────────────────────────┐
│                      RAG Saldivia Platform                       │
├─────────────────────────────────────────────────────────────────┤
│  CLI (rag-saldivia)  │  MCP Server  │  Watch Folder  │  API    │
├─────────────────────────────────────────────────────────────────┤
│                     Mode Manager (1-GPU)                         │
│         ┌─────────────┐              ┌─────────────┐            │
│         │ QUERY MODE  │◄────────────►│ INGEST MODE │            │
│         │ NIMs: 46 GB │              │ NIMs + VLM  │            │
│         │ LLM: API    │              │   ~90 GB    │            │
│         └─────────────┘              └─────────────┘            │
├─────────────────────────────────────────────────────────────────┤
│  Collections  │  Ingestion Queue  │  Cache  │  Config Loader    │
├─────────────────────────────────────────────────────────────────┤
│                    NVIDIA RAG Blueprint v2.5.0                   │
└─────────────────────────────────────────────────────────────────┘
```

**Tech Stack:** Python 3.11+, PyYAML, httpx, FastAPI, Redis (queue/cache), Click (CLI), MCP SDK

**Prerequisites:**
- `pip install pytest pyyaml httpx click redis mcp`
- Blueprint cloned at `./blueprint/` (done by setup.sh)
- Docker + Docker Compose
- RTX PRO 6000 Blackwell (98 GB) or Brev 2-GPU instance

---

## File Structure

### New Files

```
saldivia/                              # SDK package
├── __init__.py                        # Package init, re-exports
├── providers.py                       # ProviderClient, ModelConfig
├── config.py                          # ConfigLoader, env generation
├── collections.py                     # Collection management
├── mode_manager.py                    # 1-GPU dynamic model loading
├── ingestion_queue.py                 # Redis-backed job queue
├── cache.py                           # Query result caching
├── mcp_server.py                      # MCP server implementation
└── tests/
    ├── __init__.py
    ├── test_providers.py
    ├── test_config.py
    ├── test_collections.py
    ├── test_mode_manager.py
    └── test_cache.py

config/
├── models.yaml                        # Service definitions
├── guardrails.yaml                    # Guardrails config
├── observability.yaml                 # Observability stack config
├── platform.yaml                      # Platform settings (modes, cache, queue)
├── profiles/
│   ├── brev-2gpu.yaml                 # 2-GPU: all local
│   ├── workstation-1gpu.yaml          # 1-GPU: dynamic loading
│   └── full-cloud.yaml                # No GPU: everything via API
├── compose-guardrails-cloud.yaml
├── compose-observability.yaml
├── compose-openrouter-proxy.yaml
└── compose-platform-services.yaml     # Redis, mode-manager

services/
├── openrouter-proxy/
│   ├── proxy.py
│   ├── Dockerfile
│   └── requirements.txt
└── mode-manager/                      # VLM loader/unloader service
    ├── manager.py
    ├── Dockerfile
    └── requirements.txt

cli/
├── __init__.py
├── main.py                            # CLI entry point
├── collections.py                     # rag-saldivia collections ...
├── ingest.py                          # rag-saldivia ingest ...
├── query.py                           # rag-saldivia query ...
└── status.py                          # rag-saldivia status

watch/                                 # Watch folder for auto-ingestion
└── .gitkeep
```

### Modified Files

```
scripts/deploy.sh                      # Add mode detection, platform services
scripts/crossdoc_client.py             # Use SDK, support caching
Makefile                               # Add CLI targets
.gitignore                             # Add watch/, cache files
pyproject.toml                         # Package definition (new)
```

---

## Phase 1: Core SDK (Tasks 0-2)

### Task 0: GitHub Setup

**Files:** git only

- [ ] **Step 1: Push current commits**
```bash
git push origin main
```

- [ ] **Step 2: Create feature branch**
```bash
git checkout -b feat/platform-v1
```

- [ ] **Step 3: Create pyproject.toml**
```toml
# pyproject.toml
[project]
name = "rag-saldivia"
version = "0.1.0"
description = "RAG platform overlay for NVIDIA Blueprint"
requires-python = ">=3.11"
dependencies = [
    "pyyaml>=6.0",
    "httpx>=0.26.0",
    "click>=8.1.0",
    "redis>=5.0.0",
    "pymilvus>=2.4.0",
    "watchdog>=4.0.0",
    "mcp>=1.0.0",
]

[project.optional-dependencies]
dev = ["pytest>=8.0.0", "pytest-asyncio>=0.23.0"]

[project.scripts]
rag-saldivia = "cli.main:cli"

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

- [ ] **Step 4: Commit**
```bash
git add pyproject.toml
git commit -m "chore: add pyproject.toml for package management"
```

---

### Task 1: SDK — ModelConfig and ProviderClient

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
import os
from unittest.mock import patch, MagicMock

def test_model_config_defaults():
    from saldivia.providers import ModelConfig
    cfg = ModelConfig(provider="local", model="test-model")
    assert cfg.provider == "local"
    assert cfg.model == "test-model"
    assert cfg.temperature == 0.1
    assert cfg.max_tokens == 2048

def test_model_config_with_endpoint():
    from saldivia.providers import ModelConfig
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
```bash
python -m pytest saldivia/tests/test_providers.py::test_model_config_defaults -v
```
Expected: FAIL with "cannot import name 'ModelConfig'"

- [ ] **Step 4: Implement ModelConfig and ProviderClient**
```python
# saldivia/providers.py
"""Provider SDK for RAG Saldivia."""
from dataclasses import dataclass, field
from typing import Optional, Iterator
import os
import json
import httpx

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
                        delta = chunk["choices"][0].get("delta", {})
                        if delta.get("content"):
                            yield delta["content"]
```

- [ ] **Step 5: Run tests**
```bash
python -m pytest saldivia/tests/test_providers.py -v
```
Expected: PASS

- [ ] **Step 6: Update __init__.py**
```python
# saldivia/__init__.py
"""RAG Saldivia SDK."""
from saldivia.providers import ModelConfig, ProviderClient

__all__ = ["ModelConfig", "ProviderClient"]
```

- [ ] **Step 7: Commit**
```bash
git add saldivia/
git commit -m "feat(sdk): add ModelConfig and ProviderClient"
```

---

### Task 2: SDK — ConfigLoader

**Files:**
- Create: `saldivia/config.py`
- Create: `saldivia/tests/test_config.py`
- Modify: `saldivia/__init__.py`

- [ ] **Step 1: Write failing test**
```python
# saldivia/tests/test_config.py
import pytest
from pathlib import Path

@pytest.fixture
def config_dir(tmp_path):
    """Create temporary config directory."""
    models = tmp_path / "models.yaml"
    models.write_text("""
providers:
  local: {}
  nvidia-api:
    base_url: https://integrate.api.nvidia.com/v1

services:
  llm:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron
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
""")

    profiles = tmp_path / "profiles"
    profiles.mkdir()
    (profiles / "brev-2gpu.yaml").write_text("{}")
    (profiles / "workstation-1gpu.yaml").write_text("""
services:
  llm:
    provider: nvidia-api
    model: nvidia/llama-3.3-nemotron-super-49b-v1.5
""")

    return tmp_path

def test_config_loader_load_default(config_dir):
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    config = loader.load()
    assert config["services"]["llm"]["provider"] == "local"

def test_config_loader_load_with_profile(config_dir):
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    config = loader.load(profile="workstation-1gpu")
    assert config["services"]["llm"]["provider"] == "nvidia-api"
```

- [ ] **Step 2: Run test to verify it fails**
```bash
python -m pytest saldivia/tests/test_config.py::test_config_loader_load_default -v
```

- [ ] **Step 3: Implement ConfigLoader**
```python
# saldivia/config.py
"""Configuration loader for RAG Saldivia."""
import os
from pathlib import Path
from typing import Optional
import yaml

from saldivia.providers import ModelConfig


def deep_merge(base: dict, override: dict) -> dict:
    """Deep merge two dicts, override wins."""
    result = base.copy()
    for key, value in override.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = value
    return result


class ConfigLoader:
    """Loads and merges configuration from YAMLs."""

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
        ("services", "vlm", "endpoint"): "APP_VLM_SERVERURL",
        ("services", "vlm", "model"): "APP_VLM_MODELNAME",
        ("guardrails", "enabled"): "ENABLE_GUARDRAILS",
        ("guardrails", "config_id"): "DEFAULT_CONFIG",
        ("observability", "opentelemetry", "endpoint"): "OTEL_EXPORTER_OTLP_ENDPOINT",
    }

    def __init__(self, config_dir: str = "config"):
        self.config_dir = Path(config_dir)

    def load(self, profile: str = None) -> dict:
        """Load configuration with optional profile overrides."""
        config = {}

        for name in ["models", "guardrails", "observability", "platform"]:
            path = self.config_dir / f"{name}.yaml"
            if path.exists():
                with open(path) as f:
                    data = yaml.safe_load(f) or {}
                    config = deep_merge(config, data)

        if profile:
            profile_path = self.config_dir / "profiles" / f"{profile}.yaml"
            if profile_path.exists():
                with open(profile_path) as f:
                    override = yaml.safe_load(f) or {}
                    config = deep_merge(config, override)

        return config

    def get_service(self, name: str, profile: str = None) -> ModelConfig:
        """Get ModelConfig for a service."""
        config = self.load(profile)
        service = config.get("services", {}).get(name, {})

        return ModelConfig(
            provider=service.get("provider", "local"),
            model=service.get("model", ""),
            endpoint=service.get("endpoint"),
            temperature=service.get("parameters", {}).get("temperature", 0.1),
            max_tokens=service.get("parameters", {}).get("max_tokens", 2048),
        )

    def _get_nested(self, data: dict, keys: tuple):
        """Get nested value."""
        for key in keys:
            if isinstance(data, dict):
                data = data.get(key)
            else:
                return None
        return data

    def generate_env(self, profile: str = None) -> dict:
        """Generate environment variables dict."""
        config = self.load(profile)
        env = {}

        for yaml_path, env_var in self.ENV_MAPPING.items():
            value = self._get_nested(config, yaml_path)
            if value is not None:
                env[env_var] = str(value)

        # OTEL_SDK_DISABLED is inverted
        if self._get_nested(config, ("observability", "enabled")) is False:
            env["OTEL_SDK_DISABLED"] = "true"

        # API key for nvidia-api provider
        if self._get_nested(config, ("services", "llm", "provider")) == "nvidia-api":
            if os.environ.get("NVIDIA_API_KEY"):
                env["APP_LLM_APIKEY"] = os.environ["NVIDIA_API_KEY"]

        return env

    def write_env_file(self, path: str, profile: str = None):
        """Write .env file."""
        env = self.generate_env(profile)
        with open(path, "w") as f:
            for key, value in sorted(env.items()):
                f.write(f"{key}={value}\n")


def validate_config(config: dict) -> list[str]:
    """Validate configuration, return errors."""
    errors = []
    services = config.get("services", {})

    for svc in ["llm", "embeddings", "reranker"]:
        if svc not in services:
            errors.append(f"Missing required service: {svc}")
        elif not services[svc].get("model"):
            errors.append(f"Service '{svc}' missing 'model'")

    valid_providers = {"local", "nvidia-api", "openrouter", "openai", "openrouter-proxy"}
    for name, svc in services.items():
        if svc.get("provider", "local") not in valid_providers:
            errors.append(f"Invalid provider for '{name}'")

    return errors
```

- [ ] **Step 4: Run tests**
```bash
python -m pytest saldivia/tests/test_config.py -v
```

- [ ] **Step 5: Update __init__.py**
```python
# saldivia/__init__.py
"""RAG Saldivia SDK."""
from saldivia.providers import ModelConfig, ProviderClient
from saldivia.config import ConfigLoader, validate_config

__all__ = ["ModelConfig", "ProviderClient", "ConfigLoader", "validate_config"]
```

- [ ] **Step 6: Commit**
```bash
git add saldivia/
git commit -m "feat(sdk): add ConfigLoader with env generation"
```

---

## Phase 2: 1-GPU Mode Manager (Tasks 3-4)

### Task 3: Mode Manager — Core Logic

**Files:**
- Create: `saldivia/mode_manager.py`
- Create: `saldivia/tests/test_mode_manager.py`

- [ ] **Step 1: Write failing test**
```python
# saldivia/tests/test_mode_manager.py
import pytest
from unittest.mock import patch, MagicMock

def test_mode_manager_initial_state():
    from saldivia.mode_manager import ModeManager, Mode
    manager = ModeManager(gpu_memory_gb=98)
    assert manager.current_mode == Mode.QUERY

def test_mode_manager_can_switch_to_ingest():
    from saldivia.mode_manager import ModeManager, Mode
    manager = ModeManager(gpu_memory_gb=98)
    assert manager.can_switch_to(Mode.INGEST) == True

def test_mode_manager_memory_requirements():
    from saldivia.mode_manager import ModeManager, Mode, MEMORY_REQUIREMENTS
    assert MEMORY_REQUIREMENTS[Mode.QUERY] < 50  # NIMs only
    assert MEMORY_REQUIREMENTS[Mode.INGEST] < 95  # NIMs + VLM
```

- [ ] **Step 2: Implement ModeManager**
```python
# saldivia/mode_manager.py
"""1-GPU Mode Manager for dynamic model loading."""
import enum
import subprocess
import time
import logging
from dataclasses import dataclass
from typing import Optional

logger = logging.getLogger(__name__)


class Mode(enum.Enum):
    QUERY = "query"      # NIMs loaded, VLM unloaded
    INGEST = "ingest"    # NIMs + VLM loaded
    TRANSITION = "transition"


# VRAM requirements in GB (from Brev measurements)
MEMORY_REQUIREMENTS = {
    Mode.QUERY: 46,    # Triton NIMs only
    Mode.INGEST: 90,   # NIMs (46) + VLM (44)
}

# Container names
VLM_CONTAINER = "qwen3-vl-8b"
NIMS_CONTAINERS = [
    "nemotron-embedding-ms",
    "nemotron-ranking-ms",
    "compose-nv-ingest-ms-runtime-1",
]


@dataclass
class ModeStatus:
    mode: Mode
    vlm_loaded: bool
    nims_loaded: bool
    gpu_memory_used_gb: float
    pending_ingestion_jobs: int


class ModeManager:
    """Manages GPU memory by loading/unloading models based on workload."""

    def __init__(self, gpu_memory_gb: float = 98):
        self.gpu_memory_gb = gpu_memory_gb
        self.current_mode = Mode.QUERY
        self._vlm_loaded = False

    def can_switch_to(self, target: Mode) -> bool:
        """Check if we have enough VRAM for target mode."""
        required = MEMORY_REQUIREMENTS.get(target, 0)
        return required <= self.gpu_memory_gb

    def get_status(self) -> ModeStatus:
        """Get current mode status."""
        return ModeStatus(
            mode=self.current_mode,
            vlm_loaded=self._vlm_loaded,
            nims_loaded=self._check_nims_running(),
            gpu_memory_used_gb=self._get_gpu_memory_used(),
            pending_ingestion_jobs=self._get_pending_jobs(),
        )

    def switch_to_ingest_mode(self) -> bool:
        """Load VLM for ingestion. Returns True if successful."""
        if self.current_mode == Mode.INGEST:
            return True

        if not self.can_switch_to(Mode.INGEST):
            logger.error(f"Not enough VRAM for ingest mode")
            return False

        self.current_mode = Mode.TRANSITION
        logger.info("Switching to INGEST mode - loading VLM...")

        try:
            self._start_vlm()
            self._wait_for_vlm_healthy()
            self._vlm_loaded = True
            self.current_mode = Mode.INGEST
            logger.info("INGEST mode active")
            return True
        except Exception as e:
            logger.error(f"Failed to switch to ingest mode: {e}")
            self.current_mode = Mode.QUERY
            return False

    def switch_to_query_mode(self) -> bool:
        """Unload VLM for query-only mode. Returns True if successful."""
        if self.current_mode == Mode.QUERY:
            return True

        self.current_mode = Mode.TRANSITION
        logger.info("Switching to QUERY mode - unloading VLM...")

        try:
            self._stop_vlm()
            self._vlm_loaded = False
            self.current_mode = Mode.QUERY
            logger.info("QUERY mode active")
            return True
        except Exception as e:
            logger.error(f"Failed to switch to query mode: {e}")
            return False

    def _start_vlm(self):
        """Start VLM container."""
        subprocess.run(
            ["docker", "start", VLM_CONTAINER],
            check=True,
            capture_output=True
        )

    def _stop_vlm(self):
        """Stop VLM container."""
        subprocess.run(
            ["docker", "stop", VLM_CONTAINER],
            check=True,
            capture_output=True
        )

    def _wait_for_vlm_healthy(self, timeout: int = 120):
        """Wait for VLM to be healthy."""
        import httpx
        start = time.time()
        while time.time() - start < timeout:
            try:
                resp = httpx.get("http://localhost:8000/health", timeout=5)
                if resp.status_code == 200:
                    return
            except:
                pass
            time.sleep(3)
        raise TimeoutError("VLM failed to become healthy")

    def _check_nims_running(self) -> bool:
        """Check if NIM containers are running."""
        for container in NIMS_CONTAINERS:
            result = subprocess.run(
                ["docker", "inspect", "-f", "{{.State.Running}}", container],
                capture_output=True, text=True
            )
            if result.returncode != 0 or result.stdout.strip() != "true":
                return False
        return True

    def _get_gpu_memory_used(self) -> float:
        """Get GPU memory usage in GB."""
        try:
            result = subprocess.run(
                ["nvidia-smi", "--query-gpu=memory.used", "--format=csv,noheader,nounits"],
                capture_output=True, text=True, check=True
            )
            return float(result.stdout.strip()) / 1024
        except:
            return 0.0

    def _get_pending_jobs(self) -> int:
        """Get number of pending ingestion jobs from queue."""
        # Will be implemented with Redis queue
        return 0
```

- [ ] **Step 3: Run tests**
```bash
python -m pytest saldivia/tests/test_mode_manager.py -v
```

- [ ] **Step 4: Commit**
```bash
git add saldivia/mode_manager.py saldivia/tests/test_mode_manager.py
git commit -m "feat(mode): add ModeManager for 1-GPU dynamic loading"
```

---

### Task 4: Mode Manager — Service and Auto-Switch

**Files:**
- Create: `services/mode-manager/manager.py`
- Create: `services/mode-manager/Dockerfile`
- Create: `services/mode-manager/requirements.txt`

- [ ] **Step 1: Create service directory**
```bash
mkdir -p services/mode-manager
```

- [ ] **Step 2: Create manager service**
```python
# services/mode-manager/manager.py
"""Mode Manager Service - monitors queue and switches modes automatically."""
import os
import time
import redis
import logging
from saldivia.mode_manager import ModeManager, Mode

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")
QUEUE_NAME = "ingestion_queue"
IDLE_TIMEOUT = int(os.getenv("IDLE_TIMEOUT", "300"))  # 5 min default
CHECK_INTERVAL = int(os.getenv("CHECK_INTERVAL", "10"))  # 10 sec

def main():
    manager = ModeManager(gpu_memory_gb=float(os.getenv("GPU_MEMORY_GB", "98")))
    r = redis.from_url(REDIS_URL)

    last_job_time = time.time()
    logger.info(f"Mode Manager started. IDLE_TIMEOUT={IDLE_TIMEOUT}s")

    while True:
        try:
            queue_length = r.llen(QUEUE_NAME)

            if queue_length > 0:
                last_job_time = time.time()
                if manager.current_mode != Mode.INGEST:
                    logger.info(f"Jobs pending ({queue_length}), switching to INGEST mode")
                    manager.switch_to_ingest_mode()

            else:
                idle_time = time.time() - last_job_time
                if manager.current_mode == Mode.INGEST and idle_time > IDLE_TIMEOUT:
                    logger.info(f"Idle for {idle_time:.0f}s, switching to QUERY mode")
                    manager.switch_to_query_mode()

            # Publish status
            status = manager.get_status()
            r.set("mode_manager:status", f"{status.mode.value}")
            r.set("mode_manager:vlm_loaded", str(status.vlm_loaded).lower())

        except Exception as e:
            logger.error(f"Error in main loop: {e}")

        time.sleep(CHECK_INTERVAL)

if __name__ == "__main__":
    main()
```

- [ ] **Step 3: Create Dockerfile**
```dockerfile
# services/mode-manager/Dockerfile
FROM python:3.11-slim

RUN apt-get update && apt-get install -y docker.io && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# saldivia package is installed via pip from local, not COPY
# Build with: docker build -f services/mode-manager/Dockerfile .
COPY services/mode-manager/manager.py .

CMD ["python", "manager.py"]
```

**Note:** Build from repo root: `docker build -f services/mode-manager/Dockerfile -t mode-manager .`

- [ ] **Step 4: Create requirements.txt**
```
# services/mode-manager/requirements.txt
redis>=5.0.0
httpx>=0.26.0
pyyaml>=6.0
```

- [ ] **Step 5: Create compose file**
```yaml
# config/compose-platform-services.yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes

  mode-manager:
    build:
      context: ../services/mode-manager
    environment:
      - REDIS_URL=redis://redis:6379
      - GPU_MEMORY_GB=${GPU_MEMORY_GB:-98}
      - IDLE_TIMEOUT=${MODE_IDLE_TIMEOUT:-300}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    depends_on:
      - redis

volumes:
  redis-data:

networks:
  default:
    name: nvidia-rag
    external: true
```

- [ ] **Step 6: Commit**
```bash
git add services/mode-manager/ config/compose-platform-services.yaml
git commit -m "feat(mode): add mode-manager service with auto-switch"
```

---

## Phase 3: Collections Management (Task 5)

### Task 5: Collections CLI and API

**Files:**
- Create: `saldivia/collections.py`
- Create: `saldivia/tests/test_collections.py`
- Create: `cli/collections.py`

- [ ] **Step 1: Write failing test**
```python
# saldivia/tests/test_collections.py
import pytest
from unittest.mock import patch, MagicMock

def test_collection_manager_list():
    from saldivia.collections import CollectionManager
    with patch('saldivia.collections.httpx') as mock_httpx:
        mock_httpx.get.return_value.json.return_value = {
            "collections": ["tecpia", "docs"]
        }
        manager = CollectionManager()
        collections = manager.list()
        assert "tecpia" in collections

def test_collection_manager_stats():
    from saldivia.collections import CollectionManager
    manager = CollectionManager()
    # Will test with mock
```

- [ ] **Step 2: Implement CollectionManager**
```python
# saldivia/collections.py
"""Collection management for RAG Saldivia."""
import httpx
from dataclasses import dataclass
from typing import Optional
from pymilvus import connections, utility, Collection


@dataclass
class CollectionStats:
    name: str
    entity_count: int
    index_type: str
    has_sparse: bool


class CollectionManager:
    """Manages Milvus collections via ingestor API and direct connection."""

    def __init__(
        self,
        ingestor_url: str = "http://localhost:8082",
        milvus_host: str = "localhost",
        milvus_port: int = 19530,
    ):
        self.ingestor_url = ingestor_url
        self.milvus_host = milvus_host
        self.milvus_port = milvus_port
        self._connected = False

    def _connect_milvus(self):
        """Connect to Milvus if not connected."""
        if not self._connected:
            connections.connect(host=self.milvus_host, port=self.milvus_port)
            self._connected = True

    def list(self) -> list[str]:
        """List all collections."""
        self._connect_milvus()
        return utility.list_collections()

    def create(self, name: str, schema: str = "hybrid") -> bool:
        """Create a new collection via ingestor API."""
        with httpx.Client(timeout=30) as client:
            resp = client.post(
                f"{self.ingestor_url}/v1/collections",
                json={"collection_name": name, "schema_type": schema}
            )
            return resp.status_code == 200

    def delete(self, name: str) -> bool:
        """Delete a collection."""
        self._connect_milvus()
        if name in self.list():
            utility.drop_collection(name)
            return True
        return False

    def stats(self, name: str) -> Optional[CollectionStats]:
        """Get collection statistics."""
        self._connect_milvus()
        if name not in self.list():
            return None

        col = Collection(name)
        col.load()

        # Check for sparse field
        has_sparse = any(f.name == "sparse" for f in col.schema.fields)

        return CollectionStats(
            name=name,
            entity_count=col.num_entities,
            index_type=col.indexes[0].params.get("index_type", "unknown") if col.indexes else "none",
            has_sparse=has_sparse,
        )

    def health(self) -> dict:
        """Check Milvus health."""
        try:
            self._connect_milvus()
            return {
                "status": "healthy",
                "collections": len(self.list()),
            }
        except Exception as e:
            return {"status": "unhealthy", "error": str(e)}
```

- [ ] **Step 3: Create CLI commands**
```python
# cli/collections.py
"""CLI commands for collection management."""
import click
from saldivia.collections import CollectionManager


@click.group()
def collections():
    """Manage document collections."""
    pass


@collections.command()
def list():
    """List all collections."""
    manager = CollectionManager()
    cols = manager.list()
    if not cols:
        click.echo("No collections found")
        return
    for col in cols:
        stats = manager.stats(col)
        if stats:
            click.echo(f"  {col}: {stats.entity_count} entities, {stats.index_type}")
        else:
            click.echo(f"  {col}")


@collections.command()
@click.argument("name")
@click.option("--schema", default="hybrid", help="Schema type: hybrid or dense")
def create(name: str, schema: str):
    """Create a new collection."""
    manager = CollectionManager()
    if manager.create(name, schema):
        click.echo(f"Created collection: {name}")
    else:
        click.echo(f"Failed to create collection: {name}", err=True)


@collections.command()
@click.argument("name")
@click.option("--confirm", is_flag=True, help="Confirm deletion")
def delete(name: str, confirm: bool):
    """Delete a collection."""
    if not confirm:
        click.echo("Add --confirm to delete")
        return
    manager = CollectionManager()
    if manager.delete(name):
        click.echo(f"Deleted collection: {name}")
    else:
        click.echo(f"Collection not found: {name}", err=True)


@collections.command()
@click.argument("name")
def stats(name: str):
    """Show collection statistics."""
    manager = CollectionManager()
    s = manager.stats(name)
    if not s:
        click.echo(f"Collection not found: {name}", err=True)
        return
    click.echo(f"Collection: {s.name}")
    click.echo(f"  Entities: {s.entity_count}")
    click.echo(f"  Index: {s.index_type}")
    click.echo(f"  Hybrid: {s.has_sparse}")
```

- [ ] **Step 4: Create CLI main**
```python
# cli/main.py
"""RAG Saldivia CLI."""
import click
from cli.collections import collections


@click.group()
@click.version_option(version="0.1.0")
def cli():
    """RAG Saldivia - Document RAG Platform"""
    pass


cli.add_command(collections)


@cli.command()
def status():
    """Show platform status."""
    from saldivia.collections import CollectionManager
    from saldivia.mode_manager import ModeManager

    click.echo("RAG Saldivia Status")
    click.echo("=" * 40)

    # Collections
    cm = CollectionManager()
    health = cm.health()
    click.echo(f"Milvus: {health['status']}")
    if health['status'] == 'healthy':
        click.echo(f"  Collections: {health['collections']}")

    # Mode (if available)
    try:
        import redis
        r = redis.from_url("redis://localhost:6379")
        mode = r.get("mode_manager:status")
        if mode:
            click.echo(f"Mode: {mode.decode()}")
    except:
        click.echo("Mode: unknown (redis not available)")


if __name__ == "__main__":
    cli()
```

- [ ] **Step 5: Create cli __init__.py**
```python
# cli/__init__.py
"""RAG Saldivia CLI."""
```

- [ ] **Step 6: Run tests**
```bash
python -m pytest saldivia/tests/test_collections.py -v
```

- [ ] **Step 7: Test CLI**
```bash
python -m cli.main --help
python -m cli.main collections --help
```

- [ ] **Step 8: Commit**
```bash
git add saldivia/collections.py saldivia/tests/test_collections.py cli/
git commit -m "feat(cli): add collection management CLI"
```

---

## Phase 4: Ingestion Queue and Watch Folder (Task 6)

### Task 6: Ingestion Queue with Watch Folder

**Files:**
- Create: `saldivia/ingestion_queue.py`
- Create: `saldivia/watch.py`
- Modify: `scripts/smart_ingest.py`

- [ ] **Step 1: Implement ingestion queue**
```python
# saldivia/ingestion_queue.py
"""Redis-backed ingestion job queue."""
import json
import redis
import uuid
from dataclasses import dataclass, asdict
from typing import Optional
from datetime import datetime


@dataclass
class IngestionJob:
    id: str
    file_path: str
    collection: str
    status: str  # pending, processing, completed, failed
    created_at: str
    started_at: Optional[str] = None
    completed_at: Optional[str] = None
    error: Optional[str] = None
    pages: Optional[int] = None


class IngestionQueue:
    """Manages ingestion jobs via Redis."""

    QUEUE_KEY = "ingestion_queue"
    JOBS_KEY = "ingestion_jobs"

    def __init__(self, redis_url: str = "redis://localhost:6379"):
        self.redis = redis.from_url(redis_url)

    def enqueue(self, file_path: str, collection: str) -> IngestionJob:
        """Add a file to the ingestion queue."""
        job = IngestionJob(
            id=str(uuid.uuid4())[:8],
            file_path=file_path,
            collection=collection,
            status="pending",
            created_at=datetime.now().isoformat(),
        )
        self.redis.lpush(self.QUEUE_KEY, job.id)
        self.redis.hset(self.JOBS_KEY, job.id, json.dumps(asdict(job)))
        return job

    def dequeue(self) -> Optional[IngestionJob]:
        """Get next job from queue."""
        job_id = self.redis.rpop(self.QUEUE_KEY)
        if not job_id:
            return None
        job_data = self.redis.hget(self.JOBS_KEY, job_id)
        if not job_data:
            return None
        return IngestionJob(**json.loads(job_data))

    def update_status(self, job_id: str, status: str, error: str = None):
        """Update job status."""
        job_data = self.redis.hget(self.JOBS_KEY, job_id)
        if not job_data:
            return
        job = json.loads(job_data)
        job["status"] = status
        if status == "processing":
            job["started_at"] = datetime.now().isoformat()
        elif status in ("completed", "failed"):
            job["completed_at"] = datetime.now().isoformat()
        if error:
            job["error"] = error
        self.redis.hset(self.JOBS_KEY, job_id, json.dumps(job))

    def pending_count(self) -> int:
        """Get number of pending jobs."""
        return self.redis.llen(self.QUEUE_KEY)

    def list_jobs(self, status: str = None) -> list[IngestionJob]:
        """List all jobs, optionally filtered by status."""
        jobs = []
        for job_id in self.redis.hkeys(self.JOBS_KEY):
            job_data = self.redis.hget(self.JOBS_KEY, job_id)
            if job_data:
                job = IngestionJob(**json.loads(job_data))
                if status is None or job.status == status:
                    jobs.append(job)
        return sorted(jobs, key=lambda j: j.created_at, reverse=True)

    def clear_completed(self):
        """Remove completed/failed jobs from history."""
        for job_id in self.redis.hkeys(self.JOBS_KEY):
            job_data = self.redis.hget(self.JOBS_KEY, job_id)
            if job_data:
                job = json.loads(job_data)
                if job["status"] in ("completed", "failed"):
                    self.redis.hdel(self.JOBS_KEY, job_id)
```

- [ ] **Step 2: Implement watch folder**
```python
# saldivia/watch.py
"""Watch folder for automatic ingestion."""
import os
import time
import logging
from pathlib import Path
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler, FileCreatedEvent
from saldivia.ingestion_queue import IngestionQueue

logger = logging.getLogger(__name__)


class IngestionHandler(FileSystemEventHandler):
    """Handler for new files in watch folder."""

    SUPPORTED_EXTENSIONS = {".pdf", ".docx", ".txt", ".md"}

    def __init__(self, queue: IngestionQueue, collection: str):
        self.queue = queue
        self.collection = collection
        self._processed = set()

    def on_created(self, event: FileCreatedEvent):
        if event.is_directory:
            return

        path = Path(event.src_path)
        if path.suffix.lower() not in self.SUPPORTED_EXTENSIONS:
            return

        # Avoid duplicates
        if str(path) in self._processed:
            return
        self._processed.add(str(path))

        # Wait for file to be fully written
        time.sleep(1)

        logger.info(f"New file detected: {path}")
        job = self.queue.enqueue(str(path), self.collection)
        logger.info(f"Queued job {job.id} for {path.name}")


def start_watcher(
    watch_dir: str,
    collection: str,
    redis_url: str = "redis://localhost:6379"
):
    """Start watching a directory for new files."""
    queue = IngestionQueue(redis_url)
    handler = IngestionHandler(queue, collection)

    observer = Observer()
    observer.schedule(handler, watch_dir, recursive=True)
    observer.start()

    logger.info(f"Watching {watch_dir} for new files -> collection '{collection}'")

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()


if __name__ == "__main__":
    import sys
    logging.basicConfig(level=logging.INFO)
    if len(sys.argv) < 3:
        print("Usage: python -m saldivia.watch <directory> <collection>")
        sys.exit(1)
    start_watcher(sys.argv[1], sys.argv[2])
```

- [ ] **Step 3: Add CLI commands for ingestion**
```python
# cli/ingest.py
"""CLI commands for ingestion."""
import click
from pathlib import Path


@click.group()
def ingest():
    """Manage document ingestion."""
    pass


@ingest.command()
@click.argument("path")
@click.argument("collection")
def add(path: str, collection: str):
    """Add a file or directory to ingestion queue."""
    from saldivia.ingestion_queue import IngestionQueue

    queue = IngestionQueue()
    p = Path(path)

    if p.is_file():
        job = queue.enqueue(str(p), collection)
        click.echo(f"Queued: {p.name} (job {job.id})")
    elif p.is_dir():
        count = 0
        for f in p.glob("**/*.pdf"):
            queue.enqueue(str(f), collection)
            count += 1
        click.echo(f"Queued {count} files")
    else:
        click.echo(f"Path not found: {path}", err=True)


@ingest.command()
def queue():
    """Show ingestion queue status."""
    from saldivia.ingestion_queue import IngestionQueue

    q = IngestionQueue()
    click.echo(f"Pending: {q.pending_count()}")

    jobs = q.list_jobs()[:10]
    for job in jobs:
        status_icon = {"pending": "⏳", "processing": "🔄", "completed": "✅", "failed": "❌"}.get(job.status, "?")
        click.echo(f"  {status_icon} {job.id}: {Path(job.file_path).name} -> {job.collection}")


@ingest.command()
@click.argument("directory")
@click.argument("collection")
def watch(directory: str, collection: str):
    """Watch a directory for new files and auto-ingest."""
    from saldivia.watch import start_watcher
    start_watcher(directory, collection)


@ingest.command()
def clear():
    """Clear completed jobs from history."""
    from saldivia.ingestion_queue import IngestionQueue
    q = IngestionQueue()
    q.clear_completed()
    click.echo("Cleared completed jobs")
```

- [ ] **Step 4: Update cli/main.py**
```python
# Add to cli/main.py
from cli.ingest import ingest
cli.add_command(ingest)
```

- [ ] **Step 5: Create ingestion worker**
```python
# saldivia/ingestion_worker.py
"""Ingestion worker - processes jobs from queue."""
import os
import time
import logging
import httpx
from pathlib import Path
from saldivia.ingestion_queue import IngestionQueue

logger = logging.getLogger(__name__)

INGESTOR_URL = os.getenv("INGESTOR_URL", "http://localhost:8082")


def process_job(job) -> bool:
    """Process a single ingestion job."""
    logger.info(f"Processing job {job.id}: {job.file_path}")

    file_path = Path(job.file_path)
    if not file_path.exists():
        logger.error(f"File not found: {file_path}")
        return False

    try:
        with open(file_path, "rb") as f:
            files = {"documents": (file_path.name, f, "application/pdf")}
            data = {"data": f'{{"collection_name": "{job.collection}"}}'}

            with httpx.Client(timeout=600) as client:
                resp = client.post(
                    f"{INGESTOR_URL}/v1/documents",
                    files=files,
                    data=data,
                )
                resp.raise_for_status()

        logger.info(f"Job {job.id} completed successfully")
        return True

    except Exception as e:
        logger.error(f"Job {job.id} failed: {e}")
        return False


def run_worker(redis_url: str = "redis://localhost:6379"):
    """Run the ingestion worker loop."""
    queue = IngestionQueue(redis_url)
    logger.info("Ingestion worker started")

    while True:
        job = queue.dequeue()

        if job is None:
            time.sleep(5)
            continue

        queue.update_status(job.id, "processing")

        success = process_job(job)

        if success:
            queue.update_status(job.id, "completed")
        else:
            queue.update_status(job.id, "failed", error="See logs for details")


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    run_worker()
```

- [ ] **Step 6: Add worker to compose**
```yaml
# Add to config/compose-platform-services.yaml under services:
  ingestion-worker:
    build:
      context: ..
      dockerfile: services/mode-manager/Dockerfile
    command: ["python", "-m", "saldivia.ingestion_worker"]
    environment:
      - REDIS_URL=redis://redis:6379
      - INGESTOR_URL=http://ingestor-server:8082
    depends_on:
      - redis
    restart: unless-stopped
```

- [ ] **Step 7: Commit**
```bash
git add saldivia/ingestion_queue.py saldivia/watch.py cli/ingest.py
git commit -m "feat(ingest): add queue and watch folder support"
```

---

## Phase 5: MCP Server (Task 7)

### Task 7: MCP Server Implementation

**Files:**
- Create: `saldivia/mcp_server.py`

- [ ] **Step 1: Implement MCP server**
```python
# saldivia/mcp_server.py
"""MCP Server for RAG Saldivia."""
import asyncio
from mcp.server import Server
from mcp.types import Tool, TextContent
from saldivia.collections import CollectionManager
from saldivia.ingestion_queue import IngestionQueue
import httpx

server = Server("rag-saldivia")


@server.list_tools()
async def list_tools():
    return [
        Tool(
            name="search_documents",
            description="Search documents in a RAG collection",
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {"type": "string", "description": "Search query"},
                    "collection": {"type": "string", "description": "Collection name"},
                    "top_k": {"type": "integer", "description": "Number of results", "default": 10},
                },
                "required": ["query", "collection"],
            },
        ),
        Tool(
            name="ask_question",
            description="Ask a question using RAG with cross-document synthesis",
            inputSchema={
                "type": "object",
                "properties": {
                    "question": {"type": "string", "description": "Question to answer"},
                    "collection": {"type": "string", "description": "Collection name"},
                },
                "required": ["question", "collection"],
            },
        ),
        Tool(
            name="list_collections",
            description="List all document collections",
            inputSchema={"type": "object", "properties": {}},
        ),
        Tool(
            name="collection_stats",
            description="Get statistics for a collection",
            inputSchema={
                "type": "object",
                "properties": {
                    "collection": {"type": "string", "description": "Collection name"},
                },
                "required": ["collection"],
            },
        ),
        Tool(
            name="ingest_document",
            description="Queue a document for ingestion",
            inputSchema={
                "type": "object",
                "properties": {
                    "file_path": {"type": "string", "description": "Path to document"},
                    "collection": {"type": "string", "description": "Target collection"},
                },
                "required": ["file_path", "collection"],
            },
        ),
        Tool(
            name="ingestion_status",
            description="Check ingestion queue status",
            inputSchema={"type": "object", "properties": {}},
        ),
    ]


@server.call_tool()
async def call_tool(name: str, arguments: dict):
    if name == "search_documents":
        return await search_documents(**arguments)
    elif name == "ask_question":
        return await ask_question(**arguments)
    elif name == "list_collections":
        return await list_collections_tool()
    elif name == "collection_stats":
        return await collection_stats_tool(**arguments)
    elif name == "ingest_document":
        return await ingest_document(**arguments)
    elif name == "ingestion_status":
        return await ingestion_status()
    else:
        raise ValueError(f"Unknown tool: {name}")


async def search_documents(query: str, collection: str, top_k: int = 10):
    """Search documents via RAG API."""
    async with httpx.AsyncClient(timeout=60) as client:
        resp = await client.post(
            "http://localhost:8081/v1/search",
            json={
                "query": query,
                "collection_names": [collection],
                "top_k": top_k,
            }
        )
        results = resp.json()
        return [TextContent(type="text", text=str(results))]


async def ask_question(question: str, collection: str):
    """Answer question via RAG API with streaming."""
    async with httpx.AsyncClient(timeout=120) as client:
        resp = await client.post(
            "http://localhost:8081/v1/generate",
            json={
                "messages": [{"role": "user", "content": question}],
                "collection_names": [collection],
                "use_knowledge_base": True,
            }
        )
        return [TextContent(type="text", text=resp.text)]


async def list_collections_tool():
    """List all collections."""
    manager = CollectionManager()
    collections = manager.list()
    result = "\n".join(f"- {c}" for c in collections) if collections else "No collections"
    return [TextContent(type="text", text=result)]


async def collection_stats_tool(collection: str):
    """Get collection stats."""
    manager = CollectionManager()
    stats = manager.stats(collection)
    if not stats:
        return [TextContent(type="text", text=f"Collection '{collection}' not found")]
    result = f"""Collection: {stats.name}
Entities: {stats.entity_count}
Index: {stats.index_type}
Hybrid: {stats.has_sparse}"""
    return [TextContent(type="text", text=result)]


async def ingest_document(file_path: str, collection: str):
    """Queue document for ingestion."""
    queue = IngestionQueue()
    job = queue.enqueue(file_path, collection)
    return [TextContent(type="text", text=f"Queued job {job.id} for {file_path}")]


async def ingestion_status():
    """Get ingestion queue status."""
    queue = IngestionQueue()
    pending = queue.pending_count()
    jobs = queue.list_jobs()[:5]
    lines = [f"Pending: {pending}"]
    for job in jobs:
        lines.append(f"- {job.id}: {job.status} - {job.file_path}")
    return [TextContent(type="text", text="\n".join(lines))]


def main():
    """Run MCP server."""
    import sys
    from mcp.server.stdio import stdio_server

    async def run():
        async with stdio_server() as (read, write):
            await server.run(read, write, server.create_initialization_options())

    asyncio.run(run())


if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Add MCP CLI command**
```python
# Add to cli/main.py
@cli.command()
def mcp():
    """Run MCP server for AI assistant integration."""
    from saldivia.mcp_server import main
    main()
```

- [ ] **Step 3: Create MCP config for Claude Code**
```json
// Add to ~/.claude.json mcp servers section:
{
  "rag-saldivia": {
    "command": "python",
    "args": ["-m", "saldivia.mcp_server"],
    "cwd": "/path/to/rag-saldivia"
  }
}
```

- [ ] **Step 4: Test MCP server**
```bash
python -m saldivia.mcp_server
# In another terminal:
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | python -m saldivia.mcp_server
```

- [ ] **Step 5: Commit**
```bash
git add saldivia/mcp_server.py
git commit -m "feat(mcp): add MCP server for AI assistant integration"
```

---

## Phase 6: Config Files and Profiles (Task 8)

### Task 8: YAML Configs and Profiles

**Files:**
- Create: `config/models.yaml`
- Create: `config/guardrails.yaml`
- Create: `config/observability.yaml`
- Create: `config/platform.yaml`
- Create: `config/profiles/*.yaml`

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

- [ ] **Step 2: Create platform.yaml**
```yaml
# config/platform.yaml
mode:
  # GPU configuration
  gpu_memory_gb: 98
  # Auto-switch to query mode after idle (seconds)
  idle_timeout: 300

queue:
  redis_url: redis://localhost:6379

cache:
  enabled: true
  ttl_seconds: 3600
  max_entries: 1000

watch:
  enabled: false
  directory: ./watch
  default_collection: default
```

- [ ] **Step 3: Create profiles**
```yaml
# config/profiles/brev-2gpu.yaml
# 2-GPU setup: everything local, no mode switching needed
mode:
  gpu_memory_gb: 196  # 2x 98GB
  idle_timeout: 0     # Never switch modes

services:
  llm:
    provider: local
    endpoint: nim-llm:8000
    model: nvidia/nemotron-3-super-120b-a12b
```

```yaml
# config/profiles/workstation-1gpu.yaml
# 1-GPU setup: LLM via API, dynamic VLM loading
mode:
  gpu_memory_gb: 98
  idle_timeout: 300

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

```yaml
# config/profiles/full-cloud.yaml
# No GPU: everything via API
mode:
  gpu_memory_gb: 0
  idle_timeout: 0

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
```

- [ ] **Step 4: Create guardrails.yaml and observability.yaml**
```yaml
# config/guardrails.yaml
enabled: false
provider: nvidia-api
config_id: nemoguard_cloud
```

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

grafana:
  enabled: true
  port: 3000
```

- [ ] **Step 5: Delete old .env profiles**
```bash
rm -f config/profiles/brev-2gpu.env config/profiles/workstation-1gpu.env
```

- [ ] **Step 6: Commit**
```bash
git add config/
git commit -m "feat(config): add YAML configs and GPU profiles"
```

---

## Phase 7: Services and Integration (Tasks 9-10)

### Task 9: OpenRouter Proxy

**Files:**
- Create: `services/openrouter-proxy/`

- [ ] **Step 1: Create proxy files**
```python
# services/openrouter-proxy/proxy.py
"""OpenRouter proxy with header injection."""
import os
from fastapi import FastAPI, Request, Response
import httpx

app = FastAPI(title="OpenRouter Proxy")

OPENROUTER_URL = os.getenv("OPENROUTER_URL", "https://openrouter.ai/api/v1")
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY", "")

@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE"])
async def proxy(path: str, request: Request):
    headers = dict(request.headers)
    headers["HTTP-Referer"] = "https://rag-saldivia.local"
    headers["X-Title"] = "RAG Saldivia"
    if OPENROUTER_API_KEY:
        headers["Authorization"] = f"Bearer {OPENROUTER_API_KEY}"
    headers.pop("host", None)

    async with httpx.AsyncClient(timeout=120) as client:
        resp = await client.request(
            method=request.method,
            url=f"{OPENROUTER_URL}/{path}",
            headers=headers,
            content=await request.body(),
        )
        return Response(content=resp.content, status_code=resp.status_code)

@app.get("/health")
async def health():
    return {"status": "ok"}
```

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

```
# services/openrouter-proxy/requirements.txt
fastapi>=0.109.0
uvicorn>=0.27.0
httpx>=0.26.0
```

- [ ] **Step 2: Create compose file**
```yaml
# config/compose-openrouter-proxy.yaml
services:
  openrouter-proxy:
    build:
      context: ../services/openrouter-proxy
    environment:
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
    ports:
      - "8080:8080"

networks:
  default:
    name: nvidia-rag
    external: true
```

- [ ] **Step 3: Commit**
```bash
git add services/openrouter-proxy/ config/compose-openrouter-proxy.yaml
git commit -m "feat(proxy): add OpenRouter proxy service"
```

---

### Task 10: Deploy Script Update

**Files:**
- Modify: `scripts/deploy.sh`
- Modify: `Makefile`

- [ ] **Step 1: Update deploy.sh**
Add after PROFILE parsing:
```bash
# Add to Python path
export PYTHONPATH="${PYTHONPATH}:$(pwd)"

# Detect GPU configuration
GPU_COUNT=$(nvidia-smi -L 2>/dev/null | wc -l || echo "0")
echo "Detected $GPU_COUNT GPU(s)"

# Generate .env.merged from YAML config
echo "Generating environment from config..."
python3 -c "
from saldivia.config import ConfigLoader
loader = ConfigLoader('config')
loader.write_env_file('.env.merged', profile='${PROFILE}')
print('Generated .env.merged')
"

# Build compose command
COMPOSE_CMD="docker compose"
COMPOSE_FILES="-f blueprint/deploy/compose/docker-compose-rag-server.yaml"
COMPOSE_FILES="$COMPOSE_FILES -f config/compose-overrides.yaml"
COMPOSE_FILES="$COMPOSE_FILES -f config/compose-platform-services.yaml"

# Add optional services
if python3 -c "from saldivia.config import ConfigLoader; c = ConfigLoader('config').load('${PROFILE}'); exit(0 if c.get('services',{}).get('llm',{}).get('provider') == 'openrouter-proxy' else 1)" 2>/dev/null; then
    COMPOSE_FILES="$COMPOSE_FILES -f config/compose-openrouter-proxy.yaml"
fi

if python3 -c "from saldivia.config import ConfigLoader; c = ConfigLoader('config').load('${PROFILE}'); exit(0 if c.get('guardrails',{}).get('enabled') else 1)" 2>/dev/null; then
    COMPOSE_FILES="$COMPOSE_FILES -f config/compose-guardrails-cloud.yaml"
fi

if python3 -c "from saldivia.config import ConfigLoader; c = ConfigLoader('config').load('${PROFILE}'); exit(0 if c.get('observability',{}).get('enabled') else 1)" 2>/dev/null; then
    COMPOSE_FILES="$COMPOSE_FILES -f blueprint/deploy/compose/observability.yaml"
fi

echo "Compose files: $COMPOSE_FILES"
```

- [ ] **Step 2: Update Makefile**
```makefile
# Add to Makefile

.PHONY: validate
validate:
	@python3 -c "from saldivia.config import ConfigLoader, validate_config; \
		c = ConfigLoader('config').load('$(PROFILE)'); \
		errors = validate_config(c); \
		print('OK' if not errors else '\n'.join(errors))"

.PHONY: show-env
show-env:
	@python3 -c "from saldivia.config import ConfigLoader; \
		env = ConfigLoader('config').generate_env('$(PROFILE)'); \
		print('\n'.join(f'{k}={v}' for k,v in sorted(env.items())))"

.PHONY: mcp
mcp:
	python -m saldivia.mcp_server

.PHONY: watch
watch:
	python -m saldivia.watch ./watch $(COLLECTION)

.PHONY: cli
cli:
	python -m cli.main $(ARGS)
```

- [ ] **Step 3: Commit**
```bash
git add scripts/deploy.sh Makefile
git commit -m "feat(deploy): integrate config loader and platform services"
```

---

## Phase 8: Caching and Crossdoc Update (Tasks 11-12)

### Task 11: Query Caching

**Files:**
- Create: `saldivia/cache.py`

- [ ] **Step 1: Implement cache**
```python
# saldivia/cache.py
"""Query result caching."""
import hashlib
import json
import redis
from typing import Optional
from dataclasses import dataclass


@dataclass
class CacheConfig:
    enabled: bool = True
    ttl_seconds: int = 3600
    max_entries: int = 1000


class QueryCache:
    """Redis-backed query cache."""

    PREFIX = "rag_cache:"

    def __init__(self, redis_url: str = "redis://localhost:6379", config: CacheConfig = None):
        self.redis = redis.from_url(redis_url)
        self.config = config or CacheConfig()

    def _key(self, query: str, collection: str) -> str:
        """Generate cache key."""
        content = f"{query}:{collection}"
        hash_val = hashlib.md5(content.encode()).hexdigest()[:16]
        return f"{self.PREFIX}{hash_val}"

    def get(self, query: str, collection: str) -> Optional[str]:
        """Get cached result."""
        if not self.config.enabled:
            return None
        key = self._key(query, collection)
        result = self.redis.get(key)
        return result.decode() if result else None

    def set(self, query: str, collection: str, result: str):
        """Cache a result."""
        if not self.config.enabled:
            return
        key = self._key(query, collection)
        self.redis.setex(key, self.config.ttl_seconds, result)

    def invalidate(self, collection: str = None):
        """Invalidate cache entries."""
        pattern = f"{self.PREFIX}*"
        for key in self.redis.scan_iter(pattern):
            self.redis.delete(key)

    def stats(self) -> dict:
        """Get cache statistics."""
        pattern = f"{self.PREFIX}*"
        count = sum(1 for _ in self.redis.scan_iter(pattern))
        return {"entries": count, "enabled": self.config.enabled}
```

- [ ] **Step 2: Commit**
```bash
git add saldivia/cache.py
git commit -m "feat(cache): add query result caching"
```

---

### Task 12: Update Crossdoc Client

**Files:**
- Modify: `scripts/crossdoc_client.py`

- [ ] **Step 1: Add SDK and cache integration**
Add near top of file:
```python
# SDK integration
try:
    from saldivia import ConfigLoader, ProviderClient, ModelConfig
    from saldivia.cache import QueryCache, CacheConfig
    SDK_AVAILABLE = True
except ImportError:
    SDK_AVAILABLE = False

# Initialize cache if available
_cache = None
def get_cache():
    global _cache
    if _cache is None and SDK_AVAILABLE:
        _cache = QueryCache()
    return _cache
```

- [ ] **Step 2: Add profile support**
Add CrossdocConfig class:
```python
class CrossdocConfig:
    """Profile-based configuration for crossdoc."""

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
                ))

            # Setup synthesis client
            synth = crossdoc.get("synthesis", {})
            if not synth.get("use_rag_server", True):
                self.synth_client = ProviderClient(ModelConfig(
                    provider=synth["provider"],
                    model=synth.get("model", ""),
                    max_tokens=synth.get("parameters", {}).get("max_tokens", 4096),
                ))
```

- [ ] **Step 3: Add --profile argument**
```python
parser.add_argument("--profile", type=str, help="Config profile")
parser.add_argument("--no-cache", action="store_true", help="Disable caching")
```

- [ ] **Step 4: Commit**
```bash
git add scripts/crossdoc_client.py
git commit -m "feat(crossdoc): add SDK integration with caching"
```

---

## Phase 9: Final Integration (Tasks 13-14)

### Task 13: Testing and Validation

- [ ] **Step 1: Run all unit tests**
```bash
python -m pytest saldivia/tests/ -v
```

- [ ] **Step 2: Validate configs**
```bash
make validate PROFILE=brev-2gpu
make validate PROFILE=workstation-1gpu
make validate PROFILE=full-cloud
```

- [ ] **Step 3: Test CLI**
```bash
python -m cli.main --help
python -m cli.main status
python -m cli.main collections list
```

- [ ] **Step 4: Test MCP server**
```bash
python -m saldivia.mcp_server &
# Test with simple JSON-RPC
```

---

### Task 14: Documentation and PR

- [ ] **Step 1: Update README.md**
Add sections for:
- CLI usage
- MCP integration
- 1-GPU mode
- Profile selection

- [ ] **Step 2: Push feature branch**
```bash
git push -u origin feat/platform-v1
```

- [ ] **Step 3: Create PR**
```bash
gh pr create --title "feat: RAG Saldivia Platform v1" --body "$(cat <<'EOF'
## Summary

Complete platform implementation with:

- **SDK**: ConfigLoader, ProviderClient, CollectionManager
- **1-GPU Mode**: Dynamic VLM loading/unloading based on workload
- **CLI**: `rag-saldivia collections|ingest|status`
- **MCP Server**: 6 tools for AI assistant integration
- **Queue**: Redis-backed ingestion with watch folder
- **Cache**: Query result caching
- **Profiles**: brev-2gpu, workstation-1gpu, full-cloud

## Architecture

- 1-GPU mode uses ~46 GB for queries (NIMs only)
- Loads VLM (+44 GB) only when ingestion pending
- LLM always via API (NVIDIA/OpenRouter)

## Test Plan

- [ ] Unit tests: `pytest saldivia/tests/ -v`
- [ ] CLI: `python -m cli.main status`
- [ ] MCP: `python -m saldivia.mcp_server`
- [ ] Deploy workstation-1gpu profile
- [ ] Test mode switching on 1-GPU

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | 0-2 | Core SDK (providers, config) |
| 2 | 3-4 | 1-GPU Mode Manager |
| 3 | 5 | Collections CLI |
| 4 | 6 | Ingestion Queue + Watch |
| 5 | 7 | MCP Server |
| 6 | 8 | YAML Configs |
| 7 | 9-10 | Services + Deploy |
| 8 | 11-12 | Cache + Crossdoc |
| 9 | 13-14 | Test + PR |

**Total: 14 tasks, ~4-5 hours**
