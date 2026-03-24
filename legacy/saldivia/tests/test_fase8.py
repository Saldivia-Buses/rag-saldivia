"""Tests para Fase 8 — Settings Pro (backend)."""
import pytest
from saldivia.auth.database import AuthDB


@pytest.fixture
def db():
    import bcrypt
    d = AuthDB(":memory:")
    pw = bcrypt.hashpw(b"pass", bcrypt.gensalt()).decode()
    with d._conn() as conn:
        conn.execute("INSERT INTO areas (id, name) VALUES (1, 'Test')")
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (1, 'u@test.com', 'U', ?, 'user', 1, 'dummy')", (pw,)
        )
    return d


def test_get_user_preferences_defaults(db):
    """Usuario sin prefs retorna dict con defaults."""
    prefs = db.get_user_preferences(1)
    assert prefs["default_query_mode"] == "standard"
    assert prefs["vdb_top_k"] == 10
    assert prefs["avatar_color"] == "#6366f1"
    assert prefs["ui_language"] == "es"
    assert prefs["notify_ingestion_done"] is True


def test_update_user_preferences_merge(db):
    """update_user_preferences hace merge parcial — preserva campos no incluidos."""
    db.update_user_preferences(1, {"vdb_top_k": 20})
    prefs = db.get_user_preferences(1)
    assert prefs["vdb_top_k"] == 20
    assert prefs["reranker_top_k"] == 5  # campo no tocado conservado


def test_update_user_name(db):
    """update_user actualiza el nombre del usuario."""
    db.update_user(1, name="Nuevo Nombre")
    with db._conn() as conn:
        row = conn.execute("SELECT name FROM users WHERE id=1").fetchone()
    assert row[0] == "Nuevo Nombre"


def test_update_user_password_correct(db):
    """update_user_password retorna True cuando la contraseña actual es correcta."""
    result = db.update_user_password(1, "pass", "nueva123")
    assert result is True
    # Verificar que la nueva contraseña funciona
    import bcrypt
    with db._conn() as conn:
        row = conn.execute("SELECT password_hash FROM users WHERE id=1").fetchone()
    assert bcrypt.checkpw(b"nueva123", row[0].encode())


def test_update_user_password_wrong(db):
    """update_user_password retorna False cuando la contraseña actual es incorrecta."""
    result = db.update_user_password(1, "incorrecta", "nueva123")
    assert result is False
    # La contraseña original sigue siendo válida
    import bcrypt
    with db._conn() as conn:
        row = conn.execute("SELECT password_hash FROM users WHERE id=1").fetchone()
    assert bcrypt.checkpw(b"pass", row[0].encode())


# --- Gateway endpoint tests ---

from fastapi.testclient import TestClient
from unittest.mock import patch
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key


@pytest.fixture
def gw_client():
    return TestClient(app, raise_server_exceptions=False)


@pytest.fixture
def user_gw():
    key, hash_val = generate_api_key()
    return User(id=1, email="u@test.com", name="U", area_id=1,
                role=Role.USER, api_key_hash=hash_val)


def test_get_preferences_endpoint(gw_client, user_gw):
    """GET /auth/me/preferences retorna las prefs del usuario."""
    prefs = {"vdb_top_k": 15, "ui_language": "en"}
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.get_user_preferences.return_value = prefs
        resp = gw_client.get(
            "/auth/me/preferences?user_id=1",
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["vdb_top_k"] == 15


def test_patch_preferences_endpoint(gw_client, user_gw):
    """PATCH /auth/me/preferences actualiza prefs."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_preferences.return_value = None
        resp = gw_client.patch(
            "/auth/me/preferences?user_id=1",
            json={"vdb_top_k": 20},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["ok"] is True
    mock_db.update_user_preferences.assert_called_once_with(1, {"vdb_top_k": 20})


def test_patch_profile_endpoint(gw_client, user_gw):
    """PATCH /auth/me/profile actualiza el nombre."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user.return_value = None
        resp = gw_client.patch(
            "/auth/me/profile?user_id=1",
            json={"name": "Nuevo Nombre"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["ok"] is True


def test_patch_password_wrong(gw_client, user_gw):
    """PATCH /auth/me/password retorna 400 si contraseña incorrecta."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_password.return_value = False
        resp = gw_client.patch(
            "/auth/me/password?user_id=1",
            json={"current_password": "wrong", "new_password": "nueva123"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 400


def test_patch_password_correct(gw_client, user_gw):
    """PATCH /auth/me/password retorna 200 si contraseña correcta."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_password.return_value = True
        resp = gw_client.patch(
            "/auth/me/password?user_id=1",
            json={"current_password": "pass", "new_password": "nueva123"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200


def test_preferences_cross_user_forbidden(gw_client):
    """Un user no puede ver ni modificar las prefs de otro user."""
    key, hash_val = generate_api_key()
    user_2 = User(id=2, email="other@test.com", name="Other", area_id=1,
                  role=Role.USER, api_key_hash=hash_val)
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_2
        # GET preferences de user_id=1 con token de user_id=2
        resp = gw_client.get(
            "/auth/me/preferences?user_id=1",
            headers={"Authorization": "Bearer rsk_dummy"}
        )
        assert resp.status_code == 403
        # PATCH preferences
        resp = gw_client.patch(
            "/auth/me/preferences?user_id=1",
            json={"vdb_top_k": 99},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
        assert resp.status_code == 403
        # PATCH profile
        resp = gw_client.patch(
            "/auth/me/profile?user_id=1",
            json={"name": "Hacker"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
        assert resp.status_code == 403
        # PATCH password
        resp = gw_client.patch(
            "/auth/me/password?user_id=1",
            json={"current_password": "pass", "new_password": "hackeado1"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
        assert resp.status_code == 403
