"""Tests for SDA gateway extensions."""
import pytest
from fastapi import HTTPException
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock, AsyncMock
import httpx
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key, hash_password, verify_password


def test_password_hashing():
    pw = "supersecret123"
    hashed = hash_password(pw)
    assert verify_password(pw, hashed)
    assert not verify_password("wrong", hashed)
    assert hashed != pw


def test_login_success(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_email.return_value = admin_user
        mock_db.update_last_login.return_value = None
        # Provide SYSTEM_API_KEY via Bearer
        with patch("saldivia.gateway.BYPASS_AUTH", True):
            resp = client.post("/auth/session",
                               json={"email": "admin@test.com", "password": "admin123"})
    assert resp.status_code == 200
    assert "token" in resp.json()


def test_login_wrong_password(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_email.return_value = admin_user
        with patch("saldivia.gateway.BYPASS_AUTH", True):
            resp = client.post("/auth/session",
                               json={"email": "admin@test.com", "password": "wrong"})
    assert resp.status_code == 401


def test_login_user_not_found(client):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_email.return_value = None
        with patch("saldivia.gateway.BYPASS_AUTH", True):
            resp = client.post("/auth/session",
                               json={"email": "noone@test.com", "password": "x"})
    assert resp.status_code == 401


def test_list_users_requires_admin(client):
    with patch("saldivia.gateway.BYPASS_AUTH", False):
        resp = client.get("/admin/users")
    assert resp.status_code == 401


def test_list_users_as_admin(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.list_users.return_value = [admin_user]
        mock_db.get_user_areas.return_value = []
        resp = client.get("/admin/users",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert len(resp.json()["users"]) == 1
    assert "areas" in resp.json()["users"][0]


def test_create_user(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        from saldivia.auth.models import User as UserModel, Role as RoleEnum
        from saldivia.auth.models import generate_api_key
        k, h = generate_api_key()
        new_u = UserModel(id=99, email="new@test.com", name="New", area_id=1,
                          role=RoleEnum.USER, api_key_hash=h)
        mock_db.create_user.return_value = new_u
        mock_db.add_user_area.return_value = None
        resp = client.post("/admin/users",
                           json={"email": "new@test.com", "name": "New",
                                 "area_ids": [1], "role": "user", "password": "pass123"},
                           headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 201
    assert "api_key" in resp.json()


def test_update_user(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.get_user_by_id.return_value = admin_user
        mock_db.update_user.return_value = None
        resp = client.put("/admin/users/1",
                          json={"name": "Updated Name"},
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert resp.json()["ok"] is True


def test_delete_user(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.get_user_by_id.return_value = admin_user
        mock_db.update_user.return_value = None
        resp = client.delete("/admin/users/1",
                             headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert resp.json()["ok"] is True


def test_reset_user_key(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.get_user_by_id.return_value = admin_user
        mock_db.update_api_key.return_value = None
        resp = client.post("/admin/users/1/reset-key",
                           headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert "api_key" in resp.json()


def test_non_admin_forbidden(client):
    from saldivia.auth.models import User as UserModel, Role as RoleEnum
    from saldivia.auth.models import generate_api_key
    k, h = generate_api_key()
    regular_user = UserModel(id=2, email="user@test.com", name="User", area_id=1,
                              role=RoleEnum.USER, api_key_hash=h)
    with patch("saldivia.gateway.db") as mock_db:
        with patch("saldivia.gateway.BYPASS_AUTH", False):
            mock_db.get_user_by_api_key_hash.return_value = regular_user
            resp = client.get("/admin/users",
                              headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 403


def test_list_areas(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        from saldivia.auth.models import Area
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.list_areas.return_value = [Area(id=1, name="Mantenimiento", description="Area de mantenimiento")]
        resp = client.get("/admin/areas", headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert resp.json()["areas"][0]["name"] == "Mantenimiento"


def test_area_manager_cannot_see_other_area(client):
    from saldivia.auth.models import generate_api_key, User as UserModel, Role as RoleModel
    key, hash_val = generate_api_key()
    manager = UserModel(id=2, email="mgr@test.com", name="Mgr", area_id=1,
                        role=RoleModel.AREA_MANAGER, api_key_hash=hash_val)
    with patch("saldivia.gateway.db") as mock_db:
        with patch("saldivia.gateway.BYPASS_AUTH", False):
            mock_db.get_user_by_api_key_hash.return_value = manager
            resp = client.get("/admin/areas/99/collections",
                              headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 403


def test_audit_filters(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        with patch("saldivia.gateway.BYPASS_AUTH", False):
            mock_db.get_user_by_api_key_hash.return_value = admin_user
            mock_db.get_audit_log_filtered.return_value = []
            resp = client.get("/admin/audit?action=query&limit=10",
                              headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    mock_db.get_audit_log_filtered.assert_called_once_with(
        user_id=None, action="query", collection=None,
        from_ts=None, to_ts=None, limit=10
    )


def test_create_session(client, admin_user):
    with patch("saldivia.gateway.BYPASS_AUTH", False), patch("saldivia.gateway.db") as mock_db:
        from saldivia.auth.models import ChatSession
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.create_chat_session.return_value = ChatSession(
            id="abc-123", user_id=1, title="Nueva consulta", collection="tecpia_test"
        )
        resp = client.post("/chat/sessions?user_id=1",
                           json={"collection": "tecpia_test"},
                           headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 201
    assert resp.json()["collection"] == "tecpia_test"


def test_get_session_not_found(client, admin_user):
    with patch("saldivia.gateway.BYPASS_AUTH", False), patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.get_chat_session.return_value = None
        resp = client.get("/chat/sessions/nonexistent?user_id=1",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 404


def test_delete_session(client, admin_user):
    with patch("saldivia.gateway.BYPASS_AUTH", False), patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.delete_chat_session.return_value = None
        resp = client.delete("/chat/sessions/abc-123?user_id=1",
                             headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert resp.json()["ok"] is True


def test_login_rate_limit_blocks_after_5_failures(client, admin_user):
    """When _check_login_rate_limit raises 429, login returns 429."""
    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.BYPASS_AUTH", True), \
         patch("saldivia.gateway._check_login_rate_limit",
               side_effect=HTTPException(status_code=429,
                                         detail="Too many failed login attempts. Try again in 60 seconds.")), \
         patch("saldivia.gateway._record_failed_login"), \
         patch("saldivia.gateway._reset_login_rate_limit"):
        mock_db.get_user_by_email.return_value = admin_user
        resp = client.post("/auth/session", json={"email": "admin@test.com", "password": "wrong"})

    assert resp.status_code == 429


def test_login_rate_limit_resets_on_success(client, admin_user):
    """Successful login calls _reset_login_rate_limit."""
    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.BYPASS_AUTH", True), \
         patch("saldivia.gateway._check_login_rate_limit"), \
         patch("saldivia.gateway._record_failed_login"), \
         patch("saldivia.gateway._reset_login_rate_limit") as mock_reset:
        mock_db.get_user_by_email.return_value = admin_user
        mock_db.update_last_login.return_value = None
        resp = client.post("/auth/session", json={"email": "admin@test.com", "password": "admin123"})

    assert resp.status_code == 200
    mock_reset.assert_called_once_with("admin@test.com")


def test_login_rate_limit_records_on_wrong_password(client, admin_user):
    """Failed login calls _record_failed_login."""
    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.BYPASS_AUTH", True), \
         patch("saldivia.gateway._check_login_rate_limit"), \
         patch("saldivia.gateway._record_failed_login") as mock_record, \
         patch("saldivia.gateway._reset_login_rate_limit"):
        mock_db.get_user_by_email.return_value = admin_user
        resp = client.post("/auth/session", json={"email": "admin@test.com", "password": "wrongpw"})

    assert resp.status_code == 401
    mock_record.assert_called_once_with("admin@test.com")


def test_upload_file_too_large_returns_413(client, admin_user):
    """File over upload limit returns 413."""
    small_limit = 10  # bytes, for test only
    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.MAX_UPLOAD_SIZE_BYTES", small_limit):
        oversized = b"x" * (small_limit + 1)
        resp = client.post(
            "/v1/documents",
            data={"data": '{"collection_name": "test"}'},
            files={"file": ("big.pdf", oversized, "application/pdf")},
        )
    assert resp.status_code == 413


def test_upload_filename_path_traversal_sanitized(client, admin_user):
    """Filename with path traversal is sanitized before processing."""
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json.return_value = {"task_id": "task-abc"}

    mock_async_client = MagicMock()
    mock_async_client.__aenter__ = MagicMock(return_value=mock_async_client)
    mock_async_client.__aexit__ = MagicMock(return_value=False)
    mock_async_client.post = MagicMock(return_value=mock_response)

    import asyncio

    async def fake_aenter(self):
        return mock_async_client

    async def fake_aexit(self, *args):
        return False

    async def fake_post(*args, **kwargs):
        return mock_response

    mock_async_client.__aenter__ = fake_aenter
    mock_async_client.__aexit__ = fake_aexit
    mock_async_client.post = fake_post

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.extract_page_count", return_value=1), \
         patch("saldivia.gateway.classify_tier", return_value="tiny"), \
         patch("httpx.AsyncClient", return_value=mock_async_client):
        mock_db.can_access.return_value = True
        mock_db.create_ingestion_job.return_value = 1
        mock_db.log_action.return_value = None

        resp = client.post(
            "/v1/documents",
            data={"data": '{"collection_name": "test"}'},
            files={"file": ("../../etc/passwd", b"content", "text/plain")},
        )
    assert resp.status_code == 200
    body = resp.json()
    assert ".." not in body.get("filename", "")
    assert "/" not in body.get("filename", "")


def test_cors_headers_present(client):
    """CORS headers are present in responses when Origin is provided."""
    resp = client.options(
        "/health",
        headers={"Origin": "http://localhost:3000", "Access-Control-Request-Method": "GET"}
    )
    assert resp.headers.get("access-control-allow-origin") == "http://localhost:3000"


# ---------------------------------------------------------------------------
# /v1/search tests
# ---------------------------------------------------------------------------

def _make_async_client_ctx(mock_resp):
    """Return a mock httpx.AsyncClient usable as async context manager."""
    mock_client = MagicMock()

    async def fake_aenter(self):
        return mock_client

    async def fake_aexit(self, *args):
        return False

    async def fake_post(*args, **kwargs):
        return mock_resp

    mock_client.__aenter__ = fake_aenter
    mock_client.__aexit__ = fake_aexit
    mock_client.post = fake_post
    return mock_client


def test_search_happy_path(client, admin_user):
    """Happy path: RAG server returns results → 200 with results list."""
    rag_body = {"results": [{"text": "chunk1", "score": 0.9}]}
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = rag_body

    mock_client = _make_async_client_ctx(mock_resp)

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_client):
        mock_db.log_action.return_value = None
        resp = client.post(
            "/v1/search",
            json={"query": "what is RAG?", "collection_names": ["docs"]},
            headers={"Authorization": "Bearer rsk_dummy"},
        )

    assert resp.status_code == 200
    assert resp.json() == rag_body


def test_search_empty_results(client, admin_user):
    """RAG server returns empty results list → 200 with results: []."""
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {"results": []}

    mock_client = _make_async_client_ctx(mock_resp)

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_client):
        mock_db.log_action.return_value = None
        resp = client.post(
            "/v1/search",
            json={"query": "nothing here", "collection_names": ["docs"]},
            headers={"Authorization": "Bearer rsk_dummy"},
        )

    assert resp.status_code == 200
    assert resp.json()["results"] == []


def test_search_rag_server_500(client, admin_user):
    """RAG server returns 500 → httpx propagates the raw response; gateway re-raises ≥400."""
    mock_resp = MagicMock()
    mock_resp.status_code = 500
    mock_resp.json.side_effect = Exception("upstream 500 — no valid JSON")

    mock_client = _make_async_client_ctx(mock_resp)

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_client):
        mock_db.log_action.return_value = None
        # The search endpoint calls resp.json() directly; when that throws, FastAPI
        # catches the unhandled exception and returns 500 to the client.
        resp = client.post(
            "/v1/search",
            json={"query": "fail", "collection_names": ["docs"]},
            headers={"Authorization": "Bearer rsk_dummy"},
        )

    assert resp.status_code >= 400


def test_search_rag_server_timeout(client, admin_user):
    """RAG server timeout raises httpx.TimeoutException → gateway returns ≥400."""
    mock_client = MagicMock()

    async def fake_aenter(self):
        return mock_client

    async def fake_aexit(self, *args):
        return False

    async def fake_post_timeout(*args, **kwargs):
        raise httpx.TimeoutException("upstream timed out")

    mock_client.__aenter__ = fake_aenter
    mock_client.__aexit__ = fake_aexit
    mock_client.post = fake_post_timeout

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_client):
        mock_db.log_action.return_value = None
        resp = client.post(
            "/v1/search",
            json={"query": "slow query", "collection_names": ["docs"]},
            headers={"Authorization": "Bearer rsk_dummy"},
        )

    assert resp.status_code >= 400


# ---------------------------------------------------------------------------
# /v1/generate tests
# ---------------------------------------------------------------------------

def _make_streaming_response(status_code: int, chunks: list[bytes] = None, error_body: bytes = b""):
    """Build a mock httpx response for send(stream=True)."""
    if chunks is None:
        chunks = [b"data: {}\n\n"]

    mock_resp = MagicMock()
    mock_resp.status_code = status_code
    mock_resp.aread = AsyncMock(return_value=error_body)
    mock_resp.aclose = AsyncMock()

    async def _aiter_bytes():
        for chunk in chunks:
            yield chunk

    mock_resp.aiter_bytes = _aiter_bytes
    return mock_resp


def test_generate_happy_path(client, admin_user):
    """Happy path: RAG server streams SSE → gateway returns 200 StreamingResponse."""
    sse_chunks = [b"data: {\"delta\": \"hello\"}\n\n", b"data: [DONE]\n\n"]
    mock_resp = _make_streaming_response(status_code=200, chunks=sse_chunks)

    mock_client = MagicMock()
    mock_client.build_request = MagicMock(return_value=MagicMock())
    mock_client.send = AsyncMock(return_value=mock_resp)
    mock_client.aclose = AsyncMock()

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_client):
        mock_db.log_action.return_value = None
        resp = client.post(
            "/v1/generate",
            json={
                "messages": [{"role": "user", "content": "hello"}],
                "collection_names": ["docs"],
            },
            headers={"Authorization": "Bearer rsk_dummy"},
        )

    assert resp.status_code == 200
    content = resp.content
    assert b"hello" in content


def test_generate_rag_server_500(client, admin_user):
    """RAG server returns 500 before streaming starts → gateway raises HTTPException ≥400."""
    mock_resp = _make_streaming_response(
        status_code=500,
        chunks=[],
        error_body=b'{"detail": "Internal Server Error"}',
    )

    mock_client = MagicMock()
    mock_client.build_request = MagicMock(return_value=MagicMock())
    mock_client.send = AsyncMock(return_value=mock_resp)
    mock_client.aclose = AsyncMock()

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_client):
        mock_db.log_action.return_value = None
        resp = client.post(
            "/v1/generate",
            json={
                "messages": [{"role": "user", "content": "fail me"}],
                "collection_names": ["docs"],
            },
            headers={"Authorization": "Bearer rsk_dummy"},
        )

    assert resp.status_code >= 400


# ── Task 4: File cache durante upload ─────────────────────────────────────────

def test_upload_creates_file_cache(client, admin_user, tmp_path, monkeypatch):
    """POST /v1/documents guarda el archivo en disco para retry server-side."""
    import saldivia.gateway as gw_module
    monkeypatch.setattr(gw_module, "INGEST_CACHE_DIR", tmp_path)

    mock_http = AsyncMock()
    mock_http.__aenter__ = AsyncMock(return_value=mock_http)
    mock_http.__aexit__ = AsyncMock(return_value=None)
    mock_http.post.return_value = MagicMock(status_code=200, json=lambda: {"task_id": "t1"})

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient", return_value=mock_http):
        mock_db.create_ingestion_job.return_value = "job-123"
        mock_db.log_action.return_value = None
        resp = client.post(
            "/v1/documents",
            files={"file": ("test.pdf", b"%PDF-1.4 test content", "application/pdf")},
            data={"data": '{"collection_name": "col"}'},
        )

    assert resp.status_code == 200
    cached = list(tmp_path.glob("**/*.pdf"))
    assert len(cached) == 1


def test_cleanup_ingest_cache(tmp_path):
    """_cleanup_ingest_cache elimina el directorio del job."""
    from saldivia.gateway import _cleanup_ingest_cache
    import saldivia.gateway as gw_module
    original = gw_module.INGEST_CACHE_DIR
    gw_module.INGEST_CACHE_DIR = tmp_path

    job_dir = tmp_path / "job-abc"
    job_dir.mkdir()
    (job_dir / "file.pdf").write_bytes(b"content")

    _cleanup_ingest_cache("job-abc")
    assert not job_dir.exists()

    gw_module.INGEST_CACHE_DIR = original


# ── Task 5: GET /v1/documents/check ───────────────────────────────────────────

def test_check_file_hash_not_found(client):
    """GET /v1/documents/check devuelve exists=false si no hay match."""
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.check_file_hash.return_value = None
        resp = client.get("/v1/documents/check?hash=abc123&collection=col")
    assert resp.status_code == 200
    assert resp.json() == {"exists": False}


def test_check_file_hash_completed(client):
    """GET /v1/documents/check devuelve exists=true con metadata si ya está indexado."""
    job = {
        "id": "job-1", "filename": "doc.pdf", "collection": "col",
        "tier": "small", "page_count": 42, "state": "completed",
        "created_at": "2026-03-20T10:00:00", "completed_at": "2026-03-20T10:05:00",
    }
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.check_file_hash.return_value = job
        resp = client.get("/v1/documents/check?hash=abc123&collection=col")
    assert resp.status_code == 200
    data = resp.json()
    assert data["exists"] is True
    assert data["state"] == "completed"
    assert data["filename"] == "doc.pdf"
    assert data["pages"] == 42


def test_check_file_hash_requires_auth(client):
    """GET /v1/documents/check requiere autenticación con BYPASS_AUTH=False."""
    with patch("saldivia.gateway.BYPASS_AUTH", False):
        resp = client.get("/v1/documents/check?hash=abc&collection=col")
    assert resp.status_code == 401


# ── Task 6: Endpoints de alertas ──────────────────────────────────────────────

def test_list_alerts_admin_only(client):
    """GET /v1/admin/alerts accesible con BYPASS_AUTH=True (dev mode)."""
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.list_ingestion_alerts.return_value = []
        resp = client.get("/v1/admin/alerts")
    assert resp.status_code == 200
    assert resp.json() == {"alerts": []}


def test_list_alerts_forbidden_for_non_admin(client):
    """GET /v1/admin/alerts devuelve 403 para usuarios no-admin."""
    from saldivia.auth.models import User as UModel, Role as RModel
    _, h = generate_api_key()
    area_manager = UModel(id=2, email="m@t.com", name="M", area_id=1,
                          role=RModel.AREA_MANAGER, api_key_hash=h)
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = area_manager
        resp = client.get("/v1/admin/alerts",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 403


def test_resolve_alert(client, admin_user):
    """PATCH /v1/admin/alerts/{id}/resolve marca la alerta como resuelta."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        resp = client.patch(
            "/v1/admin/alerts/alert-1/resolve",
            json={"notes": "fixed by restarting ingestor"},
            headers={"Authorization": "Bearer rsk_dummy"},
        )
    assert resp.status_code == 200
    mock_db.resolve_ingestion_alert.assert_called_once_with(
        "alert-1", resolved_by=admin_user.email, notes="fixed by restarting ingestor"
    )
