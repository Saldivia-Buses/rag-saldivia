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
