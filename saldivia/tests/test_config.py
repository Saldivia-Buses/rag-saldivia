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
