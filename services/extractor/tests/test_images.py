"""Tests for pymupdf image extraction.

Uses a minimal PDF generated in-memory — no external files needed.
"""

import fitz  # pymupdf

from extractor.images import extract_images_from_pdf, render_page_as_image


def _make_test_pdf(pages: int = 2, with_image: bool = False) -> bytes:
    """Create a minimal PDF in memory for testing."""
    doc = fitz.open()
    for i in range(pages):
        page = doc.new_page(width=595, height=842)  # A4
        page.insert_text((72, 72), f"Page {i + 1} content", fontsize=14)

        if with_image and i == 0:
            # Insert a 200x200 red square as an image
            pix = fitz.Pixmap(fitz.csRGB, fitz.IRect(0, 0, 200, 200), 1)
            pix.set_rect(pix.irect, (255, 0, 0, 255))
            img_bytes = pix.tobytes("png")
            rect = fitz.Rect(100, 100, 300, 300)
            page.insert_image(rect, stream=img_bytes)

    pdf_bytes = doc.tobytes()
    doc.close()
    return pdf_bytes


def test_render_page_as_image():
    pdf = _make_test_pdf(pages=3)
    png = render_page_as_image(pdf, page_number=1, dpi=72)
    assert len(png) > 0
    assert png[:8] == b"\x89PNG\r\n\x1a\n"  # PNG magic bytes


def test_render_all_pages():
    pdf = _make_test_pdf(pages=3)
    for i in range(1, 4):
        png = render_page_as_image(pdf, page_number=i)
        assert len(png) > 100


def test_extract_images_from_pdf_no_images():
    pdf = _make_test_pdf(pages=2, with_image=False)
    images = extract_images_from_pdf(pdf)
    assert images == []


def test_extract_images_from_pdf_with_image():
    pdf = _make_test_pdf(pages=2, with_image=True)
    images = extract_images_from_pdf(pdf, min_size_px=50)
    assert len(images) >= 1
    assert images[0].page_number == 1
    assert images[0].width >= 50
    assert len(images[0].image_bytes) > 0


def test_extract_images_min_size_filter():
    pdf = _make_test_pdf(pages=1, with_image=True)
    # Image is 200x200 — filter at 300 should skip it
    images = extract_images_from_pdf(pdf, min_size_px=300)
    assert images == []
