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


def test_create_and_get_ingestion_job(db):
    area = db.create_area("Ops")
    _, hash_val = generate_api_key()
    user = db.create_user("ops@test.com", "Ops", area.id, Role.ADMIN, hash_val)

    job_id = db.create_ingestion_job(
        user_id=user.id,
        task_id="ingestor-task-abc",
        filename="contrato.pdf",
        collection="legal",
        tier="medium",
        page_count=180,
    )
    assert isinstance(job_id, str) and len(job_id) == 36  # UUID

    job = db.get_ingestion_job(job_id)
    assert job["user_id"] == user.id
    assert job["task_id"] == "ingestor-task-abc"
    assert job["filename"] == "contrato.pdf"
    assert job["tier"] == "medium"
    assert job["page_count"] == 180
    assert job["state"] == "pending"
    assert job["progress"] == 0


def test_get_active_ingestion_jobs(db):
    area = db.create_area("IT")
    _, hash_val = generate_api_key()
    user = db.create_user("it@test.com", "IT", area.id, Role.ADMIN, hash_val)

    id1 = db.create_ingestion_job(user.id, "t1", "a.pdf", "col", "tiny", 5)
    id2 = db.create_ingestion_job(user.id, "t2", "b.pdf", "col", "small", 50)

    # Completar uno
    db.update_ingestion_job(id2, "completed", 100)

    active = db.get_active_ingestion_jobs(user.id)
    assert len(active) == 1
    assert active[0]["id"] == id1


def test_update_ingestion_job(db):
    area = db.create_area("Dev")
    _, hash_val = generate_api_key()
    user = db.create_user("dev@test.com", "Dev", area.id, Role.ADMIN, hash_val)
    job_id = db.create_ingestion_job(user.id, "t3", "c.pdf", "col", "large", 300)

    db.update_ingestion_job(job_id, "running", 42)
    job = db.get_ingestion_job(job_id)
    assert job["state"] == "running"
    assert job["progress"] == 42
    assert job["completed_at"] is None

    db.update_ingestion_job(job_id, "completed", 100, completed_at="2026-03-22T12:00:00")
    job = db.get_ingestion_job(job_id)
    assert job["completed_at"] == "2026-03-22T12:00:00"


def test_ingestion_job_not_found(db):
    assert db.get_ingestion_job("nonexistent-uuid") is None


def test_active_jobs_only_own_user(db):
    area = db.create_area("Finance")
    _, h1 = generate_api_key()
    _, h2 = generate_api_key()
    u1 = db.create_user("u1@test.com", "U1", area.id, Role.USER, h1)
    u2 = db.create_user("u2@test.com", "U2", area.id, Role.USER, h2)

    db.create_ingestion_job(u1.id, "ta", "x.pdf", "col", "tiny", 5)
    db.create_ingestion_job(u2.id, "tb", "y.pdf", "col", "tiny", 5)

    assert len(db.get_active_ingestion_jobs(u1.id)) == 1
    assert len(db.get_active_ingestion_jobs(u2.id)) == 1
