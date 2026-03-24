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
    model: nvidia/llama-nemotron-embed-1b-v2
  reranker:
    provider: local
    endpoint: nemotron-ranking-ms:8000
    model: nvidia/llama-nemotron-rerank-1b-v2
""")

    profiles = tmp_path / "profiles"
    profiles.mkdir()
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

def test_generate_env(config_dir):
    """Test that generate_env() returns expected ENV_MAPPING keys."""
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    env = loader.generate_env()

    # At least these keys should be present
    assert "APP_LLM_MODELNAME" in env
    assert "APP_EMBEDDINGS_MODELNAME" in env
    assert env["APP_LLM_MODELNAME"] == "nvidia/nemotron"
    assert env["APP_EMBEDDINGS_MODELNAME"] == "nvidia/llama-nemotron-embed-1b-v2"

def test_validate_config_ok(config_dir):
    """Test that a valid config returns no errors."""
    from saldivia.config import ConfigLoader, validate_config
    loader = ConfigLoader(str(config_dir))
    config = loader.load()
    errors = validate_config(config)
    assert errors == []

def test_validate_config_missing_service(config_dir):
    """Test that missing required service returns errors."""
    from saldivia.config import validate_config
    config = {"services": {"embeddings": {"model": "x"}, "reranker": {"model": "y"}}}
    errors = validate_config(config)
    assert len(errors) > 0
    assert any("llm" in err.lower() for err in errors)

def test_ingestion_config_defaults():
    """ingestion_config() devuelve defaults cuando no hay YAML cargado."""
    from saldivia.config import ConfigLoader
    loader = ConfigLoader()
    cfg = loader.ingestion_config()
    assert cfg["parallel_slots_small"] == 2
    assert cfg["parallel_slots_large"] == 1
    assert cfg["client_max_retries"] == 3
    assert cfg["server_max_retries"] == 3
    assert cfg["stall_check_interval"] == 60
    assert set(cfg["tiers"].keys()) == {"tiny", "small", "medium", "large"}
    assert cfg["tiers"]["tiny"]["timeout"] == 300
    assert cfg["tiers"]["large"]["timeout"] == 7200


def test_ingestion_config_profile_override(tmp_path):
    """Valores del profile YAML overridean los defaults."""
    profiles = tmp_path / "profiles"
    profiles.mkdir()
    (profiles / "test.yaml").write_text(
        "ingestion:\n  stall_check_interval: 120\n  server_max_retries: 5\n"
    )
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(tmp_path))
    loader.load(profile="test")
    cfg = loader.ingestion_config()
    assert cfg["stall_check_interval"] == 120
    assert cfg["server_max_retries"] == 5
    assert cfg["parallel_slots_small"] == 2  # default mantenido


def test_env_merged_includes_saldivia_vars(config_dir):
    """Test that .env.saldivia vars are not in ENV_MAPPING output.

    This documents that .env.saldivia should be the base before generated vars.
    APP_VECTORSTORE_SEARCHTYPE is a saldivia-specific var, not in ENV_MAPPING.
    """
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    env = loader.generate_env()

    # APP_VECTORSTORE_SEARCHTYPE is not in ENV_MAPPING
    assert "APP_VECTORSTORE_SEARCHTYPE" not in env


def test_get_rag_params_returns_expected_keys(config_dir):
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    loader.load()
    params = loader.get_rag_params()
    assert "temperature" in params
    assert "max_tokens" in params
    assert "llm_model" in params
    assert params["temperature"] == 0.1
    assert params["llm_model"] == "nvidia/nemotron"


def test_update_rag_params_creates_overrides(config_dir):
    from saldivia.config import ConfigLoader
    import yaml
    loader = ConfigLoader(str(config_dir))
    loader.load()
    loader.update_rag_params({"temperature": 0.9})
    overrides_path = config_dir / "admin-overrides.yaml"
    assert overrides_path.exists()
    data = yaml.safe_load(overrides_path.read_text())
    assert data["services"]["llm"]["parameters"]["temperature"] == 0.9


def test_update_rag_params_merges(config_dir):
    from saldivia.config import ConfigLoader
    import yaml
    loader = ConfigLoader(str(config_dir))
    loader.load()
    loader.update_rag_params({"temperature": 0.9})
    loader.update_rag_params({"max_tokens": 4096})
    overrides_path = config_dir / "admin-overrides.yaml"
    data = yaml.safe_load(overrides_path.read_text())
    assert data["services"]["llm"]["parameters"]["temperature"] == 0.9
    assert data["services"]["llm"]["parameters"]["max_tokens"] == 4096


def test_reset_rag_params(config_dir):
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    loader.load()
    loader.update_rag_params({"temperature": 0.9})
    loader.reset_rag_params()
    assert not (config_dir / "admin-overrides.yaml").exists()


def test_switch_profile_inmemory(config_dir):
    from saldivia.config import ConfigLoader
    loader = ConfigLoader(str(config_dir))
    loader.load()
    assert loader._config["services"]["llm"]["provider"] == "local"
    loader.switch_profile("workstation-1gpu")
    assert loader._config["services"]["llm"]["provider"] == "nvidia-api"
    assert loader._active_profile == "workstation-1gpu"
    assert not (config_dir / "admin-overrides.yaml").exists()
