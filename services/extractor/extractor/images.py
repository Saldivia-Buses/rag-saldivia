"""Image extraction from PDFs using pymupdf (fitz).

Functions accept an open fitz.Document to avoid re-opening the PDF repeatedly.
"""

import logging
from dataclasses import dataclass

import fitz

logger = logging.getLogger(__name__)


@dataclass
class ExtractedImage:
    page_number: int
    index: int  # image index within the page
    image_bytes: bytes
    width: int
    height: int
    ext: str  # "png", "jpeg"


def extract_images_from_doc(
    doc: fitz.Document,
    min_size_px: int = 100,
    max_per_page: int = 10,
) -> list[ExtractedImage]:
    """Extract embedded images from an open PDF document.

    Args:
        doc: An already-open fitz.Document.
        min_size_px: Skip images smaller than this in either dimension.
        max_per_page: Max images to extract per page.

    Returns:
        List of ExtractedImage with PNG bytes.
    """
    results: list[ExtractedImage] = []

    for page_num in range(len(doc)):
        page = doc[page_num]
        image_list = page.get_images(full=True)
        count = 0

        for img_index, img_info in enumerate(image_list):
            if count >= max_per_page:
                break

            xref = img_info[0]
            try:
                base_image = doc.extract_image(xref)
            except Exception:
                logger.debug("skip unextractable image xref=%d page=%d", xref, page_num + 1)
                continue

            width = base_image["width"]
            height = base_image["height"]

            if width < min_size_px or height < min_size_px:
                continue

            # Convert to PNG if not already
            image_bytes = base_image["image"]
            ext = base_image["ext"]
            if ext != "png":
                pix = fitz.Pixmap(image_bytes)
                image_bytes = pix.tobytes("png")
                ext = "png"

            results.append(
                ExtractedImage(
                    page_number=page_num + 1,
                    index=img_index,
                    image_bytes=image_bytes,
                    width=width,
                    height=height,
                    ext=ext,
                )
            )
            count += 1

    return results


def extract_images_from_pdf(
    pdf_bytes: bytes,
    min_size_px: int = 100,
    max_per_page: int = 10,
) -> list[ExtractedImage]:
    """Convenience wrapper that opens a PDF from bytes."""
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    results = extract_images_from_doc(doc, min_size_px, max_per_page)
    doc.close()
    return results


def render_page_as_image_from_doc(doc: fitz.Document, page_number: int, dpi: int = 200) -> bytes:
    """Render a single page from an open document as PNG bytes.

    Args:
        doc: An already-open fitz.Document.
        page_number: 1-based page number.
        dpi: Resolution for rendering.

    Returns:
        PNG bytes of the rendered page.
    """
    page = doc[page_number - 1]
    mat = fitz.Matrix(dpi / 72, dpi / 72)
    pix = page.get_pixmap(matrix=mat)
    return pix.tobytes("png")


def render_page_as_image(pdf_bytes: bytes, page_number: int, dpi: int = 200) -> bytes:
    """Convenience wrapper that opens a PDF from bytes."""
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    result = render_page_as_image_from_doc(doc, page_number, dpi)
    doc.close()
    return result
