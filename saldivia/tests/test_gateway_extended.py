"""Tests for SDA gateway extensions."""
import pytest
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
