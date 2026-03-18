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
