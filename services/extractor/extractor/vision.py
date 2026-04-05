"""Vision client — calls SGLang (Qwen3.5-9B) via OpenAI-compatible API.

Analyzes extracted images and returns structured descriptions.
"""

import base64
import logging

import httpx

from .schema import ImageResult

logger = logging.getLogger(__name__)


class VisionClient:
    def __init__(self, endpoint: str, model: str = "Qwen/Qwen3.5-9B"):
        self.endpoint = endpoint.rstrip("/")
        self.model = model
        self._client = httpx.Client(timeout=120.0)

    def analyze_image(self, image_bytes: bytes) -> ImageResult:
        """Send an image to Qwen3.5-9B and return a structured description.

        Args:
            image_bytes: PNG/JPEG bytes of the image.

        Returns:
            ImageResult with description, type, and extracted data.
        """
        b64 = base64.b64encode(image_bytes).decode()

        resp = self._client.post(
            f"{self.endpoint}/v1/chat/completions",
            json={
                "model": self.model,
                "messages": [
                    {
                        "role": "user",
                        "content": [
                            {
                                "type": "image_url",
                                "image_url": {"url": f"data:image/png;base64,{b64}"},
                            },
                            {
                                "type": "text",
                                "text": (
                                    "Analyze this image from a technical document. "
                                    "Respond in JSON with these fields:\n"
                                    '- "description": detailed description of the image\n'
                                    '- "type": one of "technical_diagram", "photo", '
                                    '"chart", "table_image", "logo", "other"\n'
                                    '- "extracted_data": list of key data points, '
                                    "labels, or measurements visible in the image\n"
                                    "Respond ONLY with valid JSON, no markdown fences."
                                ),
                            },
                        ],
                    }
                ],
                "max_tokens": 1024,
                "temperature": 0.0,
            },
        )
        resp.raise_for_status()
        content = resp.json()["choices"][0]["message"]["content"]

        return self._parse_response(content)

    def _parse_response(self, content: str) -> ImageResult:
        """Parse LLM JSON response into ImageResult, with fallback."""
        import json

        try:
            data = json.loads(content)
            return ImageResult(
                description=data.get("description", content),
                type=data.get("type", "unknown"),
                extracted_data=data.get("extracted_data", []),
            )
        except (json.JSONDecodeError, KeyError):
            logger.warning("failed to parse vision response as JSON, using raw text")
            return ImageResult(description=content)

    def close(self):
        self._client.close()
