"""Tests for the extraction pipeline with mocked SGLang calls.

These tests verify the pipeline orchestration without requiring
actual SGLang model servers to be running.
"""

from unittest.mock import MagicMock, patch

import fitz
import pytest

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


def test_pipeline_ocr_failure_continues():
    """Pipeline records empty text when OCR fails on a page."""
    pdf = _make_test_pdf(with_image=False)

    mock_storage = MagicMock()
    mock_storage.get.return_value = pdf

    mock_ocr = MagicMock()
    mock_ocr.model = "test-ocr"
    mock_ocr.extract_page.side_effect = Exception("SGLang timeout")

    mock_vision = MagicMock()
    mock_vision.model = "test-vision"

    pipeline = ExtractionPipeline(mock_ocr, mock_vision, mock_storage)

    job = ExtractionJob(
        document_id="doc-fail",
        tenant_slug="t",
        storage_key="t/doc-fail/original.pdf",
        file_name="fail.pdf",
        file_type="pdf",
    )

    result = pipeline.extract(job)

    assert result.total_pages == 1
    assert result.pages[0].text == ""  # fallback to empty


def test_pipeline_vision_failure_records_placeholder():
    """Pipeline records placeholder when vision analysis fails."""
    pdf = _make_test_pdf(with_image=True)

    mock_storage = MagicMock()
    mock_storage.get.return_value = pdf

    mock_ocr = MagicMock()
    mock_ocr.model = "test-ocr"
    mock_ocr.extract_page.return_value = "text"

    mock_vision = MagicMock()
    mock_vision.model = "test-vision"
    mock_vision.analyze_image.side_effect = Exception("Vision model down")

    pipeline = ExtractionPipeline(mock_ocr, mock_vision, mock_storage)

    job = ExtractionJob(
        document_id="doc-v",
        tenant_slug="t",
        storage_key="t/doc-v/original.pdf",
        file_name="v.pdf",
        file_type="pdf",
    )

    result = pipeline.extract(job)

    # Pipeline should still complete
    assert result.total_pages == 1
    # Image should have error placeholder
    if result.pages[0].images:
        assert result.pages[0].images[0].description == "[extraction failed]"
        assert result.pages[0].images[0].type == "error"


def test_pipeline_rejects_cross_tenant_storage_key():
    """Pipeline rejects storage keys that don't match tenant slug."""
    mock_storage = MagicMock()
    mock_ocr = MagicMock()
    mock_vision = MagicMock()

    pipeline = ExtractionPipeline(mock_ocr, mock_vision, mock_storage)

    job = ExtractionJob(
        document_id="doc-1",
        tenant_slug="tenant-a",
        storage_key="tenant-b/doc-1/original.pdf",  # wrong tenant!
        file_name="test.pdf",
        file_type="pdf",
    )

    with pytest.raises(ValueError, match="does not match tenant"):
        pipeline.extract(job)

    # Storage should never be called
    mock_storage.get.assert_not_called()
