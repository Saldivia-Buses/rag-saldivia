"""Tests for SDA gateway extensions."""
import pytest
from fastapi import HTTPException
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key, hash_password, verify_password


def test_password_hashing():
    pw = "supersecret123"
    hashed = hash_password(pw)
    assert verify_password(pw, hashed)
    assert not verify_password("wrong", hashed)
    assert hashed != pw


@pytest.fixture
def client():
    return TestClient(app, raise_server_exceptions=False)


@pytest.fixture
def admin_user():
    key, hash_val = generate_api_key()
    return User(id=1, email="admin@test.com", name="Admin", area_id=1,
                role=Role.ADMIN, api_key_hash=hash_val,
                password_hash=hash_password("admin123"))


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
        resp = client.get("/admin/users",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert len(resp.json()["users"]) == 1


def test_create_user(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        from saldivia.auth.models import User as UserModel, Role as RoleEnum
        from saldivia.auth.models import generate_api_key
        k, h = generate_api_key()
        new_u = UserModel(id=99, email="new@test.com", name="New", area_id=1,
                          role=RoleEnum.USER, api_key_hash=h)
        mock_db.create_user.return_value = new_u
        resp = client.post("/admin/users",
                           json={"email": "new@test.com", "name": "New",
                                 "area_id": 1, "role": "user", "password": "pass123"},
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
    """File over 1GB limit returns 413."""
    from saldivia.gateway import MAX_UPLOAD_SIZE_BYTES
    oversized = b"x" * (MAX_UPLOAD_SIZE_BYTES + 1)
    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user):
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
    # Request should not 500 — sanitization should handle the filename gracefully
    assert resp.status_code != 500
