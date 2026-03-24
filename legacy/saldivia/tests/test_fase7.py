"""Tests para Fase 7 — Chat Sesiones Pro (backend)."""
import pytest
from saldivia.auth.database import AuthDB
from fastapi.testclient import TestClient
from unittest.mock import patch
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key


@pytest.fixture
def db():
    """AuthDB en memoria para tests."""
    import bcrypt
    d = AuthDB(":memory:")
    pw = bcrypt.hashpw(b"pass", bcrypt.gensalt()).decode()
    with d._conn() as conn:
        conn.execute("INSERT INTO areas (id, name) VALUES (1, 'Test')")
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (1, 'u@test.com', 'U', ?, 'user', 1, 'dummy_hash')", (pw,)
        )
    return d


def test_chat_message_has_id(db):
    """get_chat_session debe retornar mensajes con id INTEGER."""
    session = db.create_chat_session(user_id=1, collection="col")
    db.add_chat_message(session.id, "user", "Hola")
    s = db.get_chat_session(session.id, 1)
    assert len(s.messages) == 1
    assert s.messages[0].id is not None
    assert isinstance(s.messages[0].id, int)


def test_rename_chat_session(db):
    """rename_chat_session actualiza el título."""
    session = db.create_chat_session(user_id=1, collection="col")
    db.rename_chat_session(session.id, user_id=1, title="Nuevo título")
    s = db.get_chat_session(session.id, 1)
    assert s.title == "Nuevo título"


def test_rename_session_wrong_user(db):
    """rename_chat_session no afecta sesiones de otro usuario."""
    session = db.create_chat_session(user_id=1, collection="col")
    db.rename_chat_session(session.id, user_id=99, title="Hackeado")
    s = db.get_chat_session(session.id, 1)
    assert s.title == "Nueva consulta"


def test_upsert_message_feedback(db):
    """upsert_message_feedback guarda y actualiza el voto."""
    session = db.create_chat_session(user_id=1, collection="col")
    db.add_chat_message(session.id, "assistant", "Respuesta")
    s = db.get_chat_session(session.id, 1)
    msg_id = s.messages[0].id

    db.upsert_message_feedback(msg_id, user_id=1, rating="up")
    with db._conn() as conn:
        row = conn.execute(
            "SELECT rating FROM message_feedback WHERE message_id=? AND user_id=?",
            (msg_id, 1)
        ).fetchone()
    assert row[0] == "up"

    # Cambiar voto (upsert)
    db.upsert_message_feedback(msg_id, user_id=1, rating="down")
    with db._conn() as conn:
        row = conn.execute(
            "SELECT rating FROM message_feedback WHERE message_id=? AND user_id=?",
            (msg_id, 1)
        ).fetchone()
    assert row[0] == "down"


def test_delete_session_cleans_feedback(db):
    """delete_chat_session debe borrar también el feedback de sus mensajes."""
    session = db.create_chat_session(user_id=1, collection="col")
    db.add_chat_message(session.id, "assistant", "Respuesta")
    s = db.get_chat_session(session.id, 1)
    msg_id = s.messages[0].id
    db.upsert_message_feedback(msg_id, user_id=1, rating="up")

    db.delete_chat_session(session.id, user_id=1)

    with db._conn() as conn:
        row = conn.execute(
            "SELECT COUNT(*) FROM message_feedback WHERE message_id=?", (msg_id,)
        ).fetchone()
    assert row[0] == 0


# ---------------------------------------------------------------------------
# Gateway endpoint tests
# ---------------------------------------------------------------------------

@pytest.fixture
def gw_client():
    return TestClient(app, raise_server_exceptions=False)


@pytest.fixture
def admin_user_gw():
    key, hash_val = generate_api_key()
    return User(id=1, email="admin@test.com", name="Admin", area_id=1,
                role=Role.ADMIN, api_key_hash=hash_val)


def test_patch_rename_session(gw_client, admin_user_gw):
    """PATCH /chat/sessions/{id} renombra la sesión."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user_gw
        mock_db.rename_chat_session.return_value = None
        resp = gw_client.patch(
            "/chat/sessions/abc-123?user_id=1",
            json={"title": "Nuevo nombre"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["ok"] is True
    mock_db.rename_chat_session.assert_called_once_with("abc-123", user_id=1, title="Nuevo nombre")


def test_patch_rename_requires_auth(gw_client):
    """PATCH sin auth devuelve 401."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = None
        resp = gw_client.patch(
            "/chat/sessions/abc-123?user_id=1",
            json={"title": "X"},
            headers={"Authorization": "Bearer invalid"}
        )
    assert resp.status_code == 401


def test_post_message_feedback(gw_client, admin_user_gw):
    """POST /chat/sessions/{id}/messages/{msgId}/feedback guarda el voto."""
    from saldivia.auth.models import ChatSession, ChatMessage
    from datetime import datetime
    fake_msg = ChatMessage(id=42, role="assistant", content="resp")
    fake_session = ChatSession(id="abc-123", user_id=1, title="T",
                               collection="col", crossdoc=False,
                               created_at=datetime.now(), updated_at=datetime.now(),
                               messages=[fake_msg])
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user_gw
        mock_db.get_chat_session.return_value = fake_session
        mock_db.upsert_message_feedback.return_value = None
        resp = gw_client.post(
            "/chat/sessions/abc-123/messages/42/feedback?user_id=1",
            json={"rating": "up"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["ok"] is True
    mock_db.upsert_message_feedback.assert_called_once_with(42, user_id=1, rating="up")


def test_post_feedback_session_not_found(gw_client, admin_user_gw):
    """POST feedback con sesión inexistente devuelve 404."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user_gw
        mock_db.get_chat_session.return_value = None
        resp = gw_client.post(
            "/chat/sessions/nonexistent/messages/42/feedback?user_id=1",
            json={"rating": "up"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 404


def test_post_feedback_message_not_in_session(gw_client, admin_user_gw):
    """POST feedback con message_id que no pertenece a la sesión devuelve 404."""
    from saldivia.auth.models import ChatSession, ChatMessage
    from datetime import datetime
    other_msg = ChatMessage(id=99, role="assistant", content="otro")
    fake_session = ChatSession(id="abc-123", user_id=1, title="T",
                               collection="col", crossdoc=False,
                               created_at=datetime.now(), updated_at=datetime.now(),
                               messages=[other_msg])
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user_gw
        mock_db.get_chat_session.return_value = fake_session
        resp = gw_client.post(
            "/chat/sessions/abc-123/messages/42/feedback?user_id=1",
            json={"rating": "down"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 404


def test_post_feedback_wrong_user(gw_client):
    """POST feedback con user_id distinto al del token devuelve 403."""
    non_admin = User(id=2, email="b@test.com", name="B", role=Role.USER,
                     area_id=1, api_key_hash="rsk_dummy")
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = non_admin
        resp = gw_client.post(
            "/chat/sessions/abc-123/messages/42/feedback?user_id=1",
            json={"rating": "up"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 403


def test_post_feedback_invalid_rating(gw_client, admin_user_gw):
    """POST feedback con rating inválido devuelve 422."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user_gw
        resp = gw_client.post(
            "/chat/sessions/abc-123/messages/42/feedback?user_id=1",
            json={"rating": "meh"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 422


def test_get_session_includes_message_id(gw_client, admin_user_gw):
    """GET /chat/sessions/{id} debe incluir el id de cada mensaje."""
    from saldivia.auth.models import ChatSession, ChatMessage
    mock_session = ChatSession(
        id="ses-1", user_id=1, title="Test", collection="col",
        messages=[ChatMessage(id=42, role="user", content="Hola")]
    )
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user_gw
        mock_db.get_chat_session.return_value = mock_session
        resp = gw_client.get("/chat/sessions/ses-1?user_id=1",
                             headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert resp.json()["messages"][0]["id"] == 42
