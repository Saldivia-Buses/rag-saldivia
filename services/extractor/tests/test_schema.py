"""Tests for extraction schema serialization."""

from extractor.schema import (
    ExtractionJob,
    ExtractionMetadata,
    ExtractionResult,
    ImageResult,
    PageResult,
    TableResult,
)


def test_extraction_result_roundtrip():
    result = ExtractionResult(
        document_id="doc-123",
        file_name="manual.pdf",
        total_pages=2,
        pages=[
            PageResult(
                page_number=1,
                text="# Chapter 1\n\nSome text",
                tables=[TableResult(markdown="| A | B |\n|---|---|\n| 1 | 2 |", caption="Table 1")],
                images=[
                    ImageResult(
                        description="Technical diagram of engine",
                        type="technical_diagram",
                        extracted_data=["Part A", "Part B"],
                        storage_key="tenant/doc-123/images/p1_img0.png",
                    )
                ],
            ),
            PageResult(page_number=2, text="# Chapter 2"),
        ],
        metadata=ExtractionMetadata(
            language="es",
            has_toc=True,
            toc_pages=[1],
            models={"ocr": "PaddleOCR-VL-1.5", "vision": "Qwen3.5-9B"},
            extraction_time_ms=5000,
        ),
    )

    # Serialize and deserialize
    json_str = result.model_dump_json()
    parsed = ExtractionResult.model_validate_json(json_str)

    assert parsed.document_id == "doc-123"
    assert parsed.total_pages == 2
    assert len(parsed.pages) == 2
    assert len(parsed.pages[0].tables) == 1
    assert len(parsed.pages[0].images) == 1
    assert parsed.pages[0].images[0].type == "technical_diagram"
    assert parsed.metadata.has_toc is True


def test_extraction_job_from_json():
    raw = (
        '{"document_id":"d1","tenant_slug":"saldivia",'
        '"storage_key":"saldivia/d1/original.pdf",'
        '"file_name":"manual.pdf","file_type":"pdf"}'
    )
    job = ExtractionJob.model_validate_json(raw)
    assert job.tenant_slug == "saldivia"
    assert job.file_type == "pdf"
