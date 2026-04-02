"""Tests para Fase 9 — Admin Pro (multi-área, CRUD áreas, permisos)."""
import pytest
from fastapi.testclient import TestClient
from saldivia.gateway import app
from saldivia.auth.database import AuthDB
from saldivia.auth.models import Permission


@pytest.fixture
def client(monkeypatch):
    import bcrypt
    db = AuthDB(":memory:")
    pw = bcrypt.hashpw(b"admin123", bcrypt.gensalt()).decode()
    with db._conn() as conn:
        conn.execute("INSERT INTO areas (id, name, description) VALUES (1, 'Default', 'Área base')")
        conn.execute("INSERT INTO areas (id, name, description) VALUES (2, 'Ingeniería', 'Equipo técnico')")
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (1, 'admin@test.com', 'Admin', ?, 'admin', 1, 'sys-key')", (pw,)
        )
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (2, 'user@test.com', 'User', ?, 'user', 1, 'user-key')", (pw,)
        )
        # Populate user_areas so count_users_in_area works correctly
        conn.execute("INSERT OR IGNORE INTO user_areas (user_id, area_id) VALUES (1, 1)")
        conn.execute("INSERT OR IGNORE INTO user_areas (user_id, area_id) VALUES (2, 1)")
    import saldivia.gateway as gw_module
    monkeypatch.setattr(gw_module, "db", db)
    monkeypatch.setenv("SYSTEM_API_KEY", "sys-key")
    with TestClient(app) as c:
        yield c


AUTH = {"Authorization": "Bearer sys-key"}


# --- Áreas CRUD ---

def test_list_areas(client):
    r = client.get("/admin/areas", headers=AUTH)
    assert r.status_code == 200
    assert len(r.json()["areas"]) == 2


def test_create_area(client):
    r = client.post("/admin/areas", json={"name": "Legal", "description": "Área legal"}, headers=AUTH)
    assert r.status_code == 201
    assert r.json()["name"] == "Legal"


def test_update_area(client):
    r = client.put("/admin/areas/1", json={"name": "Default Actualizado"}, headers=AUTH)
    assert r.status_code == 200
    assert r.json()["ok"] is True


def test_delete_area_empty(client):
    r = client.post("/admin/areas", json={"name": "Temporal"}, headers=AUTH)
    area_id = r.json()["id"]
    r = client.delete(f"/admin/areas/{area_id}", headers=AUTH)
    assert r.status_code == 200


def test_delete_area_with_collections_cleans_up(client):
    """Área sin usuarios pero con colecciones asignadas puede eliminarse — limpia area_collections."""
    r = client.post("/admin/areas", json={"name": "ConColecciones"}, headers=AUTH)
    area_id = r.json()["id"]
    client.post(f"/admin/areas/{area_id}/collections",
                json={"collection_name": "docs", "permission": "read"}, headers=AUTH)
    r = client.delete(f"/admin/areas/{area_id}", headers=AUTH)
    assert r.status_code == 200


def test_delete_area_with_users_fails(client):
    """Área con usuarios activos no se puede eliminar — retorna 409."""
    r = client.delete("/admin/areas/1", headers=AUTH)
    assert r.status_code == 409


# --- Multi-área ---

def test_add_user_area(client):
    r = client.post("/admin/users/2/areas", json={"area_id": 2}, headers=AUTH)
    assert r.status_code == 200
    assert r.json()["ok"] is True


def test_remove_user_area(client):
    client.post("/admin/users/2/areas", json={"area_id": 2}, headers=AUTH)
    r = client.delete("/admin/users/2/areas/2", headers=AUTH)
    assert r.status_code == 200


def test_list_users_returns_areas_list(client):
    """GET /admin/users retorna `areas` como lista, no `area_id`."""
    r = client.get("/admin/users", headers=AUTH)
    assert r.status_code == 200
    users = r.json()["users"]
    assert all("areas" in u for u in users)
    assert all("area_id" not in u for u in users)


def test_create_user_without_area(client):
    """Crear usuario sin area_ids es válido."""
    r = client.post("/admin/users", json={
        "email": "nuevo@test.com", "name": "Nuevo",
        "role": "user", "password": "pass123"
    }, headers=AUTH)
    assert r.status_code == 201


def test_create_user_with_multiple_areas(client):
    r = client.post("/admin/users", json={
        "email": "multi@test.com", "name": "Multi",
        "role": "user", "password": "pass123",
        "area_ids": [1, 2]
    }, headers=AUTH)
    assert r.status_code == 201


# --- can_access multi-área ---

def test_can_access_union_of_areas():
    """Usuario con 2 áreas puede acceder a colecciones de ambas."""
    import bcrypt
    db = AuthDB(":memory:")
    pw = bcrypt.hashpw(b"pass", bcrypt.gensalt()).decode()
    with db._conn() as conn:
        conn.execute("INSERT INTO areas (id, name) VALUES (1, 'A1')")
        conn.execute("INSERT INTO areas (id, name) VALUES (2, 'A2')")
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (1, 'u@t.com', 'U', ?, 'user', NULL, 'k')", (pw,)
        )
        conn.execute("INSERT INTO user_areas VALUES (1, 1)")
        conn.execute("INSERT INTO user_areas VALUES (1, 2)")
        conn.execute("INSERT INTO area_collections (area_id, collection_name, permission) VALUES (1, 'colA', 'read')")
        conn.execute("INSERT INTO area_collections (area_id, collection_name, permission) VALUES (2, 'colB', 'read')")
    user = db.get_user_by_id(1)
    assert db.can_access(user, "colA", Permission.READ) is True
    assert db.can_access(user, "colB", Permission.READ) is True
    assert db.can_access(user, "colC", Permission.READ) is False


def test_can_access_no_areas_denied():
    """Usuario sin áreas no puede acceder a ninguna colección."""
    import bcrypt
    db = AuthDB(":memory:")
    pw = bcrypt.hashpw(b"pass", bcrypt.gensalt()).decode()
    with db._conn() as conn:
        conn.execute("INSERT INTO areas (id, name) VALUES (1, 'A')")
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (1, 'u@t.com', 'U', ?, 'user', NULL, 'k')", (pw,)
        )
        conn.execute("INSERT INTO area_collections (area_id, collection_name, permission) VALUES (1, 'col', 'read')")
    user = db.get_user_by_id(1)
    assert db.can_access(user, "col", Permission.READ) is False


# --- Permisos colecciones ---

def test_grant_and_revoke_collection(client):
    client.post("/admin/areas/1/collections",
                json={"collection_name": "docs", "permission": "read"}, headers=AUTH)
    r = client.get("/admin/areas/1/collections", headers=AUTH)
    assert any(c["name"] == "docs" for c in r.json()["collections"])
    client.delete("/admin/areas/1/collections/docs", headers=AUTH)
    r = client.get("/admin/areas/1/collections", headers=AUTH)
    assert not any(c["name"] == "docs" for c in r.json()["collections"])
