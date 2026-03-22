import pytest
from saldivia.gateway import extract_page_count, classify_tier


def test_classify_tier_by_pages():
    assert classify_tier(10, 0) == "tiny"
    assert classify_tier(20, 0) == "tiny"
    assert classify_tier(21, 0) == "small"
    assert classify_tier(80, 0) == "small"
    assert classify_tier(81, 0) == "medium"
    assert classify_tier(250, 0) == "medium"
    assert classify_tier(251, 0) == "large"
    assert classify_tier(1000, 0) == "large"


def test_classify_tier_by_size_when_no_pages():
    assert classify_tier(None, 50_000) == "tiny"
    assert classify_tier(None, 99_999) == "tiny"
    assert classify_tier(None, 100_000) == "small"
    assert classify_tier(None, 499_999) == "small"
    assert classify_tier(None, 500_000) == "medium"
    assert classify_tier(None, 4_999_999) == "medium"
    assert classify_tier(None, 5_000_000) == "large"


def test_extract_page_count_non_pdf():
    assert extract_page_count(b"texto plano", "doc.txt") is None
    assert extract_page_count(b"markdown", "readme.md") is None
    assert extract_page_count(b"word", "doc.docx") is None


def test_extract_page_count_invalid_pdf():
    # Bytes que no son un PDF válido → debe devolver None, no lanzar excepción
    assert extract_page_count(b"not a pdf", "doc.pdf") is None
