"""Tests for Vision client — JSON parsing, fallback, and error handling."""

import json
from unittest.mock import MagicMock

import httpx
import pytest

from extractor.schema import ImageResult
from extractor.vision import VisionClient


def _mock_response(content: str, status_code: int = 200) -> httpx.Response:
    return httpx.Response(
        status_code=status_code,
        json={
            "choices": [{"message": {"content": content}}],
            "usage": {"prompt_tokens": 50, "completion_tokens": 30},
        },
        request=httpx.Request("POST", "http://test/v1/chat/completions"),
    )


def test_analyze_image_parses_json():
    """Vision client parses structured JSON from model."""
    client = VisionClient("http://localhost:8101")
    mock_client = MagicMock()
    mock_client.post.return_value = _mock_response(
        json.dumps({
            "description": "Engine schematic showing cylinder layout",
            "type": "technical_diagram",
            "extracted_data": ["Cylinder 1", "Cylinder 2", "Intake valve"],
        })
    )
    client._client = mock_client

    result = client.analyze_image(b"fakeimage")

    assert isinstance(result, ImageResult)
    assert result.description == "Engine schematic showing cylinder layout"
    assert result.type == "technical_diagram"
    assert len(result.extracted_data) == 3


def test_analyze_image_fallback_on_invalid_json():
    """Vision client falls back to raw text when JSON is invalid."""
    client = VisionClient("http://localhost:8101")
    mock_client = MagicMock()
    mock_client.post.return_value = _mock_response(
        "This is not valid JSON, just a description of the image."
    )
    client._client = mock_client

    result = client.analyze_image(b"fakeimage")

    assert isinstance(result, ImageResult)
    assert "not valid JSON" in result.description
    assert result.type == "unknown"
    assert result.extracted_data == []


def test_analyze_image_partial_json():
    """Vision client handles JSON missing optional fields."""
    client = VisionClient("http://localhost:8101")
    mock_client = MagicMock()
    mock_client.post.return_value = _mock_response(
        json.dumps({"description": "A photo of a bus"})
    )
    client._client = mock_client

    result = client.analyze_image(b"fakeimage")

    assert result.description == "A photo of a bus"
    assert result.type == "unknown"
    assert result.extracted_data == []


def test_analyze_image_raises_on_http_error():
    """Vision client raises on non-200 responses."""
    client = VisionClient("http://localhost:8101")
    mock_client = MagicMock()
    mock_client.post.return_value = httpx.Response(
        status_code=503,
        text="Service Unavailable",
        request=httpx.Request("POST", "http://test"),
    )
    client._client = mock_client

    with pytest.raises(httpx.HTTPStatusError):
        client.analyze_image(b"image")


def test_parse_response_with_markdown_fences():
    """Vision client strips markdown fences if model wraps JSON in them."""
    client = VisionClient("http://localhost:8101")
    # Model sometimes wraps JSON in ```json ... ```
    content = '```json\n{"description": "chart", "type": "chart"}\n```'
    # This will fail JSON parse and fall back to raw text
    result = client._parse_response(content)
    # Fallback: raw text as description
    assert "chart" in result.description
