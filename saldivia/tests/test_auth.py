# saldivia/tests/test_auth.py
import pytest
from saldivia.auth import AuthDB, Role, Permission, generate_api_key, verify_api_key


@pytest.fixture
def db(tmp_path):
    from saldivia.auth.database import DB_PATH
    import saldivia.auth.database as db_module
    db_module.DB_PATH = tmp_path / "test_auth.db"
    return AuthDB(db_module.DB_PATH)


def test_generate_api_key():
    key, hash_val = generate_api_key()
    assert key.startswith("rsk_")
    assert len(key) > 40
    assert verify_api_key(key, hash_val)


def test_create_area(db):
    area = db.create_area("Mantenimiento", "Equipo de mantenimiento")
    assert area.id == 1
    assert area.name == "Mantenimiento"


def test_create_user(db):
    area = db.create_area("IT")
    key, hash_val = generate_api_key()
    user = db.create_user("admin@empresa.com", "Admin", area.id, Role.ADMIN, hash_val)
    assert user.id == 1
    assert user.role == Role.ADMIN


def test_collection_permissions(db):
    area = db.create_area("Producción")
    key, hash_val = generate_api_key()
    user = db.create_user("juan@empresa.com", "Juan", area.id, Role.USER, hash_val)

    # No access initially
    assert not db.can_access(user, "tecpia", Permission.READ)

    # Grant read access
    db.grant_collection_access(area.id, "tecpia", Permission.READ)
    assert db.can_access(user, "tecpia", Permission.READ)
    assert not db.can_access(user, "tecpia", Permission.WRITE)

    # Upgrade to write
    db.grant_collection_access(area.id, "tecpia", Permission.WRITE)
    assert db.can_access(user, "tecpia", Permission.WRITE)
