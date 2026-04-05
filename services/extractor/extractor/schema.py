"""Output schema for the extraction pipeline.

Matches the JSON structure defined in Plan 06 Phase 2.
This is the contract between the Extractor and the Ingest Service.
"""

from pydantic import BaseModel, Field


class TableResult(BaseModel):
    markdown: str
    caption: str = ""


class ImageResult(BaseModel):
    description: str
    type: str = "unknown"  # technical_diagram, photo, chart, logo, etc.
    extracted_data: list[str] = Field(default_factory=list)
    storage_key: str = ""  # MinIO key for the extracted image file


class PageResult(BaseModel):
    page_number: int
    text: str = ""
    tables: list[TableResult] = Field(default_factory=list)
    images: list[ImageResult] = Field(default_factory=list)


class ExtractionMetadata(BaseModel):
    language: str = "auto"
    has_toc: bool = False
    toc_pages: list[int] = Field(default_factory=list)
    models: dict[str, str] = Field(default_factory=dict)
    extraction_time_ms: int = 0


class ExtractionResult(BaseModel):
    document_id: str
    file_name: str
    total_pages: int
    pages: list[PageResult]
    metadata: ExtractionMetadata


class ExtractionJob(BaseModel):
    """Message received via NATS to trigger extraction."""

    document_id: str
    tenant_slug: str
    storage_key: str  # MinIO key: {tenant}/{doc_id}/original.pdf
    file_name: str
    file_type: str  # pdf, png, jpg
