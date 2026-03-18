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
