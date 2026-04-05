"""OCR client — calls SGLang (PaddleOCR-VL) via OpenAI-compatible API.

Sends a page image and receives structured text (Markdown) with tables.
"""

import base64
import logging
from pathlib import Path

import httpx

logger = logging.getLogger(__name__)


class OCRClient:
    def __init__(self, endpoint: str, model: str = "PaddlePaddle/PaddleOCR-VL-1.5"):
        self.endpoint = endpoint.rstrip("/")
        self.model = model
        self._client = httpx.Client(timeout=120.0)

    def extract_page(self, image_bytes: bytes) -> str:
        """Send a page image to PaddleOCR-VL and return Markdown text.

        Args:
            image_bytes: PNG/JPEG bytes of a single page.

        Returns:
            Extracted text in Markdown format (includes tables).
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
                                "text": "Extract all text from this document page. "
                                "Output in Markdown format. Preserve tables as "
                                "Markdown tables. Preserve headings hierarchy.",
                            },
                        ],
                    }
                ],
                "max_tokens": 4096,
                "temperature": 0.0,
            },
        )
        resp.raise_for_status()
        data = resp.json()
        return data["choices"][0]["message"]["content"]

    def close(self):
        self._client.close()
