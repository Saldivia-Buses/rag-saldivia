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


# ── Fase 6: file_hash, retry_count, ingestion_alerts ──────────────────────────

def test_ingestion_job_has_hash_and_retry_fields(db):
    """ingestion_jobs acepta file_hash y expone retry_count=0 por defecto."""
    area = db.create_area("Test")
    _, h = generate_api_key()
    user = db.create_user("u@t.com", "U", area.id, Role.ADMIN, h)
    job_id = db.create_ingestion_job(
        user.id, "t1", "doc.pdf", "col", "small", 10, file_hash="abc123"
    )
    job = db.get_ingestion_job(job_id)
    assert job["file_hash"] == "abc123"
    assert job["retry_count"] == 0
    assert job["last_checked"] is None


def test_check_file_hash_not_found(db):
    assert db.check_file_hash("nonexistent", "col") is None


def test_check_file_hash_completed(db):
    area = db.create_area("Test")
    _, h = generate_api_key()
    user = db.create_user("u@t.com", "U", area.id, Role.ADMIN, h)
    job_id = db.create_ingestion_job(
        user.id, "t1", "doc.pdf", "col", "small", 10, file_hash="sha256abc"
    )
    db.update_ingestion_job(job_id, "completed", 100)
    result = db.check_file_hash("sha256abc", "col")
    assert result is not None
    assert result["state"] == "completed"
    assert result["filename"] == "doc.pdf"


def test_check_file_hash_different_collection_returns_none(db):
    """Hash en colección distinta no da match."""
    area = db.create_area("Test")
    _, h = generate_api_key()
    user = db.create_user("u@t.com", "U", area.id, Role.ADMIN, h)
    job_id = db.create_ingestion_job(
        user.id, "t1", "doc.pdf", "col-a", "small", 10, file_hash="sha256xyz"
    )
    db.update_ingestion_job(job_id, "completed", 100)
    assert db.check_file_hash("sha256xyz", "col-b") is None


def test_create_and_list_ingestion_alerts(db):
    area = db.create_area("Test")
    _, h = generate_api_key()
    user = db.create_user("u@t.com", "U", area.id, Role.ADMIN, h)
    alert_id = db.create_ingestion_alert(
        job_id="job1", user_id=user.id, filename="doc.pdf",
        collection="col", tier="large", page_count=300,
        file_hash="sha256abc", error="timeout after 7200s",
        retry_count=3, progress_at_failure=42,
    )
    alerts = db.list_ingestion_alerts()
    assert len(alerts) == 1
    assert alerts[0]["id"] == alert_id
    assert alerts[0]["resolved_at"] is None


def test_resolve_ingestion_alert(db):
    area = db.create_area("Test")
    _, h = generate_api_key()
    user = db.create_user("u@t.com", "U", area.id, Role.ADMIN, h)
    alert_id = db.create_ingestion_alert(
        job_id="job1", user_id=user.id, filename="f.pdf",
        collection="col", tier="tiny", page_count=5,
        file_hash="h", error="err", retry_count=1, progress_at_failure=0,
    )
    db.resolve_ingestion_alert(alert_id, resolved_by="admin@test.com", notes="fixed")
    resolved = db.list_ingestion_alerts(resolved=True)
    assert len(resolved) == 1
    assert resolved[0]["resolved_at"] is not None
    assert resolved[0]["notes"] == "fixed"
    assert db.list_ingestion_alerts(resolved=False) == []


def test_increment_ingestion_retry(db):
    area = db.create_area("Test")
    _, h = generate_api_key()
    user = db.create_user("u@t.com", "U", area.id, Role.ADMIN, h)
    job_id = db.create_ingestion_job(user.id, "t1", "doc.pdf", "col", "medium", 100)
    db.increment_ingestion_retry(job_id)
    job = db.get_ingestion_job(job_id)
    assert job["retry_count"] == 1
    assert job["last_checked"] is not None


def test_get_all_active_ingestion_jobs(db):
    """get_all_active_ingestion_jobs devuelve jobs de todos los usuarios."""
    area = db.create_area("Test")
    _, h1 = generate_api_key()
    _, h2 = generate_api_key()
    u1 = db.create_user("u1@t.com", "U1", area.id, Role.USER, h1)
    u2 = db.create_user("u2@t.com", "U2", area.id, Role.USER, h2)
    db.create_ingestion_job(u1.id, "ta", "a.pdf", "col", "tiny", 5)
    db.create_ingestion_job(u2.id, "tb", "b.pdf", "col", "tiny", 5)
    all_jobs = db.get_all_active_ingestion_jobs()
    assert len(all_jobs) == 2
