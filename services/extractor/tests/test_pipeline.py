"""Tests for the extraction pipeline with mocked SGLang calls.

These tests verify the pipeline orchestration without requiring
actual SGLang model servers to be running.
"""

from unittest.mock import MagicMock, patch

import fitz

from extractor.pipeline import ExtractionPipeline
from extractor.schema import ExtractionJob, ImageResult


def _make_test_pdf(with_image: bool = True) -> bytes:
    doc = fitz.open()
    page = doc.new_page(width=595, height=842)
    page.insert_text((72, 72), "Test document content", fontsize=14)
    page.insert_text((72, 100), "| Col1 | Col2 |", fontsize=10)

    if with_image:
        pix = fitz.Pixmap(fitz.csRGB, fitz.IRect(0, 0, 200, 200), 1)
        pix.set_rect(pix.irect, (0, 0, 255, 255))
        page.insert_image(fitz.Rect(100, 200, 300, 400), stream=pix.tobytes("png"))

    pdf_bytes = doc.tobytes()
    doc.close()
    return pdf_bytes


def test_pipeline_extracts_pages():
    """Pipeline produces correct page count and OCR text."""
    pdf = _make_test_pdf(with_image=False)

    mock_storage = MagicMock()
    mock_storage.get.return_value = pdf

    mock_ocr = MagicMock()
    mock_ocr.model = "test-ocr"
    mock_ocr.extract_page.return_value = "# Test\n\nExtracted text from OCR"

    mock_vision = MagicMock()
    mock_vision.model = "test-vision"

    pipeline = ExtractionPipeline(mock_ocr, mock_vision, mock_storage)

    job = ExtractionJob(
        document_id="doc-1",
        tenant_slug="test-tenant",
        storage_key="test-tenant/doc-1/original.pdf",
        file_name="test.pdf",
        file_type="pdf",
    )

    result = pipeline.extract(job)

    assert result.document_id == "doc-1"
    assert result.total_pages == 1
    assert len(result.pages) == 1
    assert result.pages[0].text == "# Test\n\nExtracted text from OCR"
    assert mock_ocr.extract_page.call_count == 1


def test_pipeline_extracts_images():
    """Pipeline detects images, stores them, and calls vision model."""
    pdf = _make_test_pdf(with_image=True)

    mock_storage = MagicMock()
    mock_storage.get.return_value = pdf

    mock_ocr = MagicMock()
    mock_ocr.model = "test-ocr"
    mock_ocr.extract_page.return_value = "Page text"

    mock_vision = MagicMock()
    mock_vision.model = "test-vision"
    mock_vision.analyze_image.return_value = ImageResult(
        description="A blue square",
        type="other",
        extracted_data=[],
    )

    pipeline = ExtractionPipeline(mock_ocr, mock_vision, mock_storage)

    job = ExtractionJob(
        document_id="doc-2",
        tenant_slug="tenant",
        storage_key="tenant/doc-2/original.pdf",
        file_name="with-image.pdf",
        file_type="pdf",
    )

    result = pipeline.extract(job)

    assert result.total_pages == 1
    # Image was found and analyzed
    assert mock_vision.analyze_image.call_count >= 1
    # Image was stored in MinIO
    assert mock_storage.put.call_count >= 1
    # Image is attached to page 1
    assert len(result.pages[0].images) >= 1
    assert result.pages[0].images[0].description == "A blue square"


def test_pipeline_metadata():
    """Pipeline produces correct metadata."""
    pdf = _make_test_pdf(with_image=False)

    mock_storage = MagicMock()
    mock_storage.get.return_value = pdf

    mock_ocr = MagicMock()
    mock_ocr.model = "PaddleOCR-VL-1.5"
    mock_ocr.extract_page.return_value = "text"

    mock_vision = MagicMock()
    mock_vision.model = "Qwen3.5-9B"

    pipeline = ExtractionPipeline(mock_ocr, mock_vision, mock_storage)

    job = ExtractionJob(
        document_id="d",
        tenant_slug="t",
        storage_key="t/d/original.pdf",
        file_name="f.pdf",
        file_type="pdf",
    )

    result = pipeline.extract(job)

    assert result.metadata.models["ocr"] == "PaddleOCR-VL-1.5"
    assert result.metadata.models["vision"] == "Qwen3.5-9B"
    assert result.metadata.extraction_time_ms >= 0
