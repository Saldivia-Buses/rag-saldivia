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


import json
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from saldivia.gateway import app
from saldivia.auth.models import User, Role


@pytest.fixture
def admin_user():
    return User(id=1, email="admin@test.com", name="Admin",
                area_id=1, role=Role.ADMIN, api_key_hash="hash")


def test_ingest_returns_job_id_and_tier(admin_user):
    """POST /v1/documents devuelve job_id, tier, page_count."""
    client = TestClient(app)

    mock_ingestor_resp = MagicMock()
    mock_ingestor_resp.json.return_value = {"task_id": "ingestor-task-xyz", "message": "queued"}
    mock_ingestor_resp.status_code = 200

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_httpx, \
         patch("saldivia.gateway.db") as mock_db:

        mock_db.create_ingestion_job.return_value = "job-uuid-123"
        mock_db.log_action = MagicMock()
        mock_client = AsyncMock()
        mock_client.post.return_value = mock_ingestor_resp
        mock_httpx.return_value.__aenter__.return_value = mock_client

        fake_pdf = b"%PDF-1.4\n%fake"

        resp = client.post(
            "/v1/documents",
            files={"file": ("test.pdf", fake_pdf, "application/pdf")},
            data={"data": json.dumps({"collection_name": "mi-col"})},
            headers={"Authorization": "Bearer test-key"},
        )

    assert resp.status_code == 200
    body = resp.json()
    assert "job_id" in body
    assert body["job_id"] == "job-uuid-123"
    assert "tier" in body
    assert body["tier"] in ("tiny", "small", "medium", "large")
    assert "filename" in body


def test_ingest_non_pdf_uses_size_tier(admin_user):
    """Archivos no-PDF deben clasificar tier por tamaño."""
    client = TestClient(app)

    mock_resp = MagicMock()
    mock_resp.json.return_value = {"task_id": "t1"}
    mock_resp.status_code = 200

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_httpx, \
         patch("saldivia.gateway.db") as mock_db:

        mock_db.create_ingestion_job.return_value = "job-123"
        mock_db.log_action = MagicMock()
        mock_client = AsyncMock()
        mock_client.post.return_value = mock_resp
        mock_httpx.return_value.__aenter__.return_value = mock_client

        tiny_txt = b"x" * 50_000  # 50KB → tiny

        resp = client.post(
            "/v1/documents",
            files={"file": ("doc.txt", tiny_txt, "text/plain")},
            data={"data": json.dumps({"collection_name": "col"})},
            headers={"Authorization": "Bearer test-key"},
        )

    assert resp.json()["tier"] == "tiny"
    assert resp.json()["page_count"] is None
