"""Tests for OCR client — request formatting and response parsing."""

import json
from unittest.mock import MagicMock, patch

import httpx
import pytest

from extractor.ocr import OCRClient


def _mock_response(content: str, status_code: int = 200) -> httpx.Response:
    return httpx.Response(
        status_code=status_code,
        json={
            "choices": [{"message": {"content": content}}],
            "usage": {"prompt_tokens": 100, "completion_tokens": 50},
        },
        request=httpx.Request("POST", "http://test/v1/chat/completions"),
    )


def test_extract_page_sends_correct_payload():
    """OCR client sends base64 image in correct OpenAI format."""
    client = OCRClient("http://localhost:8100")
    mock_client = MagicMock()
    mock_client.post.return_value = _mock_response("# Heading\n\nSome text")
    client._client = mock_client

    result = client.extract_page(b"\x89PNG\r\n\x1a\nfakeimage")

    assert result == "# Heading\n\nSome text"
    call_args = mock_client.post.call_args
    assert call_args[0][0] == "http://localhost:8100/v1/chat/completions"
    body = call_args[1]["json"]
    assert body["model"] == "PaddlePaddle/PaddleOCR-VL-1.5"
    assert body["temperature"] == 0.0
    assert body["max_tokens"] == 4096
    # Verify base64 encoding
    content = body["messages"][0]["content"]
    assert content[0]["type"] == "image_url"
    assert content[0]["image_url"]["url"].startswith("data:image/png;base64,")


def test_extract_page_strips_trailing_slash():
    """Endpoint URL trailing slash is stripped."""
    client = OCRClient("http://localhost:8100/")
    assert client.endpoint == "http://localhost:8100"


def test_extract_page_raises_on_http_error():
    """OCR client raises on non-200 responses."""
    client = OCRClient("http://localhost:8100")
    mock_client = MagicMock()
    mock_client.post.return_value = httpx.Response(
        status_code=500,
        text="Internal Server Error",
        request=httpx.Request("POST", "http://test"),
    )
    client._client = mock_client

    with pytest.raises(httpx.HTTPStatusError):
        client.extract_page(b"image")


def test_extract_page_returns_empty_string_from_model():
    """OCR client handles empty model response."""
    client = OCRClient("http://localhost:8100")
    mock_client = MagicMock()
    mock_client.post.return_value = _mock_response("")
    client._client = mock_client

    result = client.extract_page(b"image")
    assert result == ""
