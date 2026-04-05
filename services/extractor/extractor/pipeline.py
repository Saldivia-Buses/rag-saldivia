"""Extraction pipeline — orchestrates OCR, image extraction, and vision analysis.

This is the core of the Extractor service. Given a PDF, it produces an
ExtractionResult with all content structured page by page.
"""

import logging
import time

import fitz  # pymupdf

from .images import ExtractedImage, extract_images_from_pdf, render_page_as_image
from .ocr import OCRClient
from .schema import (
    ExtractionJob,
    ExtractionMetadata,
    ExtractionResult,
    ImageResult,
    PageResult,
)
from .storage import StorageClient
from .vision import VisionClient

logger = logging.getLogger(__name__)


class ExtractionPipeline:
    def __init__(
        self,
        ocr: OCRClient,
        vision: VisionClient,
        storage: StorageClient,
    ):
        self.ocr = ocr
        self.vision = vision
        self.storage = storage

    def extract(self, job: ExtractionJob) -> ExtractionResult:
        """Run the full extraction pipeline on a document.

        Steps:
        1. Download PDF from MinIO
        2. OCR each page via PaddleOCR-VL (SGLang)
        3. Extract embedded images via pymupdf
        4. Analyze each image via Qwen3.5-9B (SGLang)
        5. Store extracted images in MinIO
        6. Return unified ExtractionResult
        """
        start = time.monotonic()
        logger.info("extracting document=%s file=%s", job.document_id, job.file_name)

        # 1. Download from MinIO
        pdf_bytes = self.storage.get(job.storage_key)
        doc = fitz.open(stream=pdf_bytes, filetype="pdf")
        total_pages = len(doc)
        doc.close()

        logger.info("document has %d pages", total_pages)

        # 2. OCR each page
        pages: list[PageResult] = []
        for page_num in range(1, total_pages + 1):
            page_image = render_page_as_image(pdf_bytes, page_num)
            text = self.ocr.extract_page(page_image)
            pages.append(PageResult(page_number=page_num, text=text))
            logger.debug("ocr page %d/%d done", page_num, total_pages)

        # 3. Extract embedded images
        extracted_images = extract_images_from_pdf(pdf_bytes)
        logger.info("found %d embedded images", len(extracted_images))

        # 4. Analyze each image + store in MinIO
        images_by_page: dict[int, list[ImageResult]] = {}
        for img in extracted_images:
            # Store image in MinIO
            img_key = (
                f"{job.tenant_slug}/{job.document_id}"
                f"/images/p{img.page_number}_img{img.index}.png"
            )
            self.storage.put(img_key, img.image_bytes, content_type="image/png")

            # Analyze via vision model
            result = self.vision.analyze_image(img.image_bytes)
            result.storage_key = img_key

            images_by_page.setdefault(img.page_number, []).append(result)
            logger.debug(
                "analyzed image p%d_img%d: type=%s",
                img.page_number,
                img.index,
                result.type,
            )

        # 5. Attach images to their pages
        for page in pages:
            page.images = images_by_page.get(page.page_number, [])

        elapsed_ms = int((time.monotonic() - start) * 1000)

        return ExtractionResult(
            document_id=job.document_id,
            file_name=job.file_name,
            total_pages=total_pages,
            pages=pages,
            metadata=ExtractionMetadata(
                models={
                    "ocr": self.ocr.model,
                    "vision": self.vision.model,
                },
                extraction_time_ms=elapsed_ms,
            ),
        )
