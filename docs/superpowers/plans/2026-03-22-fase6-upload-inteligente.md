# Fase 6 — Upload Inteligente: Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reemplazar el upload blocking de 600s con un sistema asíncrono real: non-blocking, progreso verdadero basado en datos del ingestor, persistencia server-side de jobs en SQLite, y recovery al recargar la página.

**Architecture:** El gateway extrae page count con pypdf, clasifica tier, llama al ingestor con `blocking=false` y registra el job en SQLite. El frontend hace polling adaptativo por tier, calculando progress = extraction×60% + indexing×40% con datos reales. Al recargar, `+page.server.ts` hidrata jobs activos desde el servidor.

**Tech Stack:** Python `pypdf`, SQLite (AuthDB existente), FastAPI (gateway), SvelteKit 5 runes, Vitest, pytest + TestClient

**Design doc:** `docs/superpowers/specs/2026-03-22-fase6-upload-inteligente-design.md`

---

## Task 1: Agregar pypdf y helpers de clasificación en gateway

**Files:**
- Modify: `pyproject.toml`
- Modify: `saldivia/gateway.py`
- Test: `saldivia/tests/test_gateway_ingest.py` (crear)

**Step 1: Escribir tests para extract_page_count y classify_tier**

Crear `saldivia/tests/test_gateway_ingest.py`:

```python
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
```

**Step 2: Correr tests — deben fallar**

```bash
cd ~/rag-saldivia && uv run pytest saldivia/tests/test_gateway_ingest.py -v
```

Esperado: `ImportError` o `AttributeError` — las funciones no existen todavía.

**Step 3: Agregar pypdf a pyproject.toml**

En la sección `dependencies`, agregar:
```toml
"pypdf>=4.0.0",
```

Instalar:
```bash
cd ~/rag-saldivia && uv sync
```

**Step 4: Agregar helpers en gateway.py**

Después de los imports existentes en `saldivia/gateway.py`, agregar:

```python
def extract_page_count(file_bytes: bytes, filename: str) -> int | None:
    """Extrae page count de un PDF. Devuelve None para no-PDFs o PDFs inválidos."""
    if not filename.lower().endswith('.pdf'):
        return None
    try:
        from pypdf import PdfReader
        import io
        reader = PdfReader(io.BytesIO(file_bytes))
        return len(reader.pages)
    except Exception:
        return None


def classify_tier(page_count: int | None, file_size: int) -> str:
    """Clasifica el tier por páginas (PDF) o tamaño de archivo (otros formatos)."""
    if page_count is not None:
        if page_count <= 20:  return "tiny"
        if page_count <= 80:  return "small"
        if page_count <= 250: return "medium"
        return "large"
    if file_size < 100_000:   return "tiny"
    if file_size < 500_000:   return "small"
    if file_size < 5_000_000: return "medium"
    return "large"
```

**Step 5: Correr tests — deben pasar**

```bash
uv run pytest saldivia/tests/test_gateway_ingest.py -v
```

Esperado: 4 tests PASS.

**Step 6: Commit**

```bash
git add pyproject.toml saldivia/gateway.py saldivia/tests/test_gateway_ingest.py
git commit -m "feat(gateway): add pypdf dependency and tier classification helpers"
```

---

## Task 2: Nueva tabla ingestion_jobs en AuthDB

**Files:**
- Modify: `saldivia/auth/database.py`
- Test: `saldivia/tests/test_auth.py` (agregar tests al final)

**Step 1: Escribir tests para los nuevos métodos**

Agregar al final de `saldivia/tests/test_auth.py`:

```python
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
```

**Step 2: Correr tests — deben fallar**

```bash
uv run pytest saldivia/tests/test_auth.py::test_create_and_get_ingestion_job -v
```

Esperado: `AttributeError: 'AuthDB' object has no attribute 'create_ingestion_job'`

**Step 3: Agregar tabla y métodos en database.py**

En `saldivia/auth/database.py`, en el método `__init__` donde se crean las tablas (buscar `CREATE TABLE IF NOT EXISTS`), agregar:

```python
conn.execute("""
    CREATE TABLE IF NOT EXISTS ingestion_jobs (
        id           TEXT PRIMARY KEY,
        user_id      INTEGER NOT NULL,
        task_id      TEXT NOT NULL,
        filename     TEXT NOT NULL,
        collection   TEXT NOT NULL,
        tier         TEXT NOT NULL,
        page_count   INTEGER,
        state        TEXT DEFAULT 'pending',
        progress     INTEGER DEFAULT 0,
        created_at   TEXT NOT NULL,
        completed_at TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id)
    )
""")
```

Luego agregar los 4 métodos en la clase `AuthDB`:

```python
def create_ingestion_job(
    self,
    user_id: int,
    task_id: str,
    filename: str,
    collection: str,
    tier: str,
    page_count: int | None,
) -> str:
    import uuid
    job_id = str(uuid.uuid4())
    with self._connect() as conn:
        conn.execute(
            """INSERT INTO ingestion_jobs
               (id, user_id, task_id, filename, collection, tier, page_count, created_at)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?)""",
            (job_id, user_id, task_id, filename, collection, tier, page_count,
             datetime.now().isoformat()),
        )
    return job_id

def get_ingestion_job(self, job_id: str) -> dict | None:
    with self._connect() as conn:
        row = conn.execute(
            "SELECT * FROM ingestion_jobs WHERE id = ?", (job_id,)
        ).fetchone()
    if row is None:
        return None
    return dict(row)

def get_active_ingestion_jobs(self, user_id: int) -> list[dict]:
    with self._connect() as conn:
        rows = conn.execute(
            """SELECT * FROM ingestion_jobs
               WHERE user_id = ? AND state IN ('pending','running','stalled')
               ORDER BY created_at DESC""",
            (user_id,),
        ).fetchall()
    return [dict(r) for r in rows]

def update_ingestion_job(
    self,
    job_id: str,
    state: str,
    progress: int,
    completed_at: str | None = None,
) -> None:
    with self._connect() as conn:
        conn.execute(
            """UPDATE ingestion_jobs
               SET state = ?, progress = ?, completed_at = ?
               WHERE id = ?""",
            (state, progress, completed_at, job_id),
        )
```

**Nota:** Verificar que `_connect()` usa `row_factory = sqlite3.Row` para que `dict(row)` funcione. Si no, usar `conn.row_factory = sqlite3.Row` o `dict(zip([c[0] for c in cursor.description], row))`.

**Step 4: Correr todos los tests de auth**

```bash
uv run pytest saldivia/tests/test_auth.py -v
```

Esperado: todos PASS (incluyendo los 5 nuevos).

**Step 5: Commit**

```bash
git add saldivia/auth/database.py saldivia/tests/test_auth.py
git commit -m "feat(auth): add ingestion_jobs table and CRUD methods"
```

---

## Task 3: Modificar POST /v1/documents en gateway — non-blocking

**Files:**
- Modify: `saldivia/gateway.py`
- Test: `saldivia/tests/test_gateway_ingest.py` (agregar)

**Step 1: Agregar tests del endpoint modificado**

Agregar en `saldivia/tests/test_gateway_ingest.py`:

```python
import json
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from saldivia.gateway import app
from saldivia.auth import User, Role


@pytest.fixture
def admin_user():
    return User(id=1, email="admin@test.com", name="Admin",
                area_id=1, role=Role.ADMIN, api_key_hash="hash")


def test_ingest_returns_job_id_and_tier(admin_user):
    """POST /v1/documents devuelve job_id, tier, page_count."""
    client = TestClient(app)

    mock_ingestor_resp = MagicMock()
    mock_ingestor_resp.json.return_value = {"task_id": "ingestor-task-xyz", "message": "queued"}

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_httpx, \
         patch("saldivia.gateway.db") as mock_db:

        mock_db.create_ingestion_job.return_value = "job-uuid-123"
        mock_client = AsyncMock()
        mock_client.post.return_value = mock_ingestor_resp
        mock_httpx.return_value.__aenter__.return_value = mock_client

        # Crear un PDF mínimo válido de 1 página (header de PDF)
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

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_httpx, \
         patch("saldivia.gateway.db") as mock_db:

        mock_db.create_ingestion_job.return_value = "job-123"
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
```

**Step 2: Correr tests — deben fallar**

```bash
uv run pytest saldivia/tests/test_gateway_ingest.py::test_ingest_returns_job_id_and_tier -v
```

Esperado: falla porque el endpoint retorna el formato del ingestor, no `{ job_id, tier }`.

**Step 3: Modificar POST /v1/documents en gateway.py**

Reemplazar la función `ingest` existente (líneas ~259-302):

```python
@app.post("/v1/documents")
async def ingest(request: Request, user: User = Depends(get_user_from_token)):
    """Upload de documento: non-blocking, devuelve job_id + tier."""
    if user and user.role == Role.USER:
        raise HTTPException(status_code=403, detail="Users cannot ingest documents directly")

    form = await request.form()
    file = form.get("file")
    data_str = form.get("data", "{}")

    if not file:
        raise HTTPException(status_code=400, detail="Se requiere un archivo.")

    try:
        data = json.loads(data_str)
    except json.JSONDecodeError:
        data = {}

    collection_name = data.get("collection_name", "")

    if user and user.role == Role.AREA_MANAGER and collection_name:
        if not db.can_access(user, collection_name, Permission.WRITE):
            raise HTTPException(
                status_code=403,
                detail=f"No write access to collection: {collection_name}"
            )

    file_bytes = await file.read()
    page_count = extract_page_count(file_bytes, file.filename)
    tier = classify_tier(page_count, len(file_bytes))

    # Payload al ingestor con blocking=False
    ingestor_data = {**data, "blocking": False}

    headers = dict(request.headers)
    headers.pop("host", None)
    headers.pop("content-length", None)
    headers.pop("content-type", None)

    async with httpx.AsyncClient(timeout=30) as client:
        resp = await client.post(
            f"{INGESTOR_URL}/v1/documents",
            files={"file": (file.filename, file_bytes, file.content_type or "application/octet-stream")},
            data={"data": json.dumps(ingestor_data)},
            headers={k: v for k, v in headers.items() if k.lower() not in ("content-type", "content-length")},
        )

    if resp.status_code >= 400:
        raise HTTPException(status_code=resp.status_code, detail=resp.text)

    ingestor_body = resp.json()
    task_id = ingestor_body.get("task_id", "")

    job_id = db.create_ingestion_job(
        user_id=user.id if user else 0,
        task_id=task_id,
        filename=file.filename,
        collection=collection_name,
        tier=tier,
        page_count=page_count,
    )

    if user:
        db.log_action(
            user_id=user.id,
            action="ingest",
            ip_address=request.client.host if request.client else ""
        )

    return {
        "job_id": job_id,
        "tier": tier,
        "page_count": page_count,
        "filename": file.filename,
    }
```

**Step 4: Correr tests**

```bash
uv run pytest saldivia/tests/test_gateway_ingest.py -v
```

Esperado: todos PASS.

**Step 5: Correr suite completa para verificar no hay regresiones**

```bash
uv run pytest saldivia/tests/ -v
```

Esperado: todos los tests existentes siguen pasando.

**Step 6: Commit**

```bash
git add saldivia/gateway.py saldivia/tests/test_gateway_ingest.py
git commit -m "feat(gateway): make POST /v1/documents non-blocking, return job_id and tier"
```

---

## Task 4: Nuevos endpoints GET /v1/jobs y GET /v1/jobs/{job_id}/status

**Files:**
- Modify: `saldivia/gateway.py`
- Test: `saldivia/tests/test_gateway_ingest.py` (agregar)

**Step 1: Agregar tests**

```python
def test_list_jobs_returns_active_only(admin_user):
    client = TestClient(app)

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db:

        mock_db.get_active_ingestion_jobs.return_value = [
            {"id": "j1", "filename": "a.pdf", "tier": "tiny", "state": "pending", "progress": 0}
        ]
        resp = client.get("/v1/jobs", headers={"Authorization": "Bearer test-key"})

    assert resp.status_code == 200
    assert len(resp.json()["jobs"]) == 1


def test_job_status_calculates_progress(admin_user):
    client = TestClient(app)

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_httpx:

        mock_db.get_ingestion_job.return_value = {
            "id": "j1", "user_id": 1, "task_id": "t1",
            "filename": "doc.pdf", "collection": "col",
            "tier": "medium", "page_count": 100,
            "state": "pending", "progress": 0,
            "created_at": "2026-03-22T10:00:00", "completed_at": None,
        }

        mock_ingestor = MagicMock()
        mock_ingestor.json.return_value = {
            "state": "PENDING",
            "nv_ingest_status": {"extraction_completed": 1},
            "result": {"total_documents": 2, "documents_completed": 0},
        }
        mock_client = AsyncMock()
        mock_client.get.return_value = mock_ingestor
        mock_httpx.return_value.__aenter__.return_value = mock_client

        resp = client.get("/v1/jobs/j1/status", headers={"Authorization": "Bearer test-key"})

    assert resp.status_code == 200
    body = resp.json()
    # extraction_completed=1, total=2 → 1/2 * 60 = 30. documents_completed=0 → 0. Total = 30
    assert body["progress"] == 30
    assert body["state"] == "running"


def test_job_status_finished_is_100(admin_user):
    client = TestClient(app)

    with patch("saldivia.gateway.get_user_from_token", return_value=admin_user), \
         patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_httpx:

        mock_db.get_ingestion_job.return_value = {
            "id": "j2", "user_id": 1, "task_id": "t2",
            "filename": "big.pdf", "collection": "col",
            "tier": "large", "page_count": 300,
            "state": "running", "progress": 60,
            "created_at": "2026-03-22T10:00:00", "completed_at": None,
        }

        mock_ingestor = MagicMock()
        mock_ingestor.json.return_value = {
            "state": "FINISHED",
            "nv_ingest_status": {"extraction_completed": 1},
            "result": {"total_documents": 1, "documents_completed": 1},
        }
        mock_client = AsyncMock()
        mock_client.get.return_value = mock_ingestor
        mock_httpx.return_value.__aenter__.return_value = mock_client

        resp = client.get("/v1/jobs/j2/status", headers={"Authorization": "Bearer test-key"})

    assert resp.json()["progress"] == 100
    assert resp.json()["state"] == "completed"


def test_job_status_403_for_other_user(admin_user):
    other_user = User(id=99, email="other@test.com", name="Other",
                      area_id=2, role=Role.USER, api_key_hash="h")
    client = TestClient(app)

    with patch("saldivia.gateway.get_user_from_token", return_value=other_user), \
         patch("saldivia.gateway.db") as mock_db:

        # Job pertenece a user_id=1, el requester es user_id=99
        mock_db.get_ingestion_job.return_value = {
            "id": "j3", "user_id": 1, "task_id": "t3",
            "filename": "secret.pdf", "collection": "col",
            "tier": "tiny", "page_count": 5,
            "state": "pending", "progress": 0,
            "created_at": "2026-03-22T10:00:00", "completed_at": None,
        }
        resp = client.get("/v1/jobs/j3/status", headers={"Authorization": "Bearer test-key"})

    assert resp.status_code == 403
```

**Step 2: Correr tests — deben fallar**

```bash
uv run pytest saldivia/tests/test_gateway_ingest.py::test_list_jobs_returns_active_only -v
```

Esperado: `404 Not Found` — endpoints no existen todavía.

**Step 3: Agregar endpoints en gateway.py**

Al final de `saldivia/gateway.py`, antes del bloque `if __name__ == "__main__"` (si existe), agregar:

```python
@app.get("/v1/jobs")
async def list_jobs(request: Request, user: User = Depends(get_user_from_token)):
    """Lista jobs de ingesta activos del usuario autenticado."""
    if not user:
        raise HTTPException(status_code=401, detail="Auth required")
    jobs = db.get_active_ingestion_jobs(user.id)
    return {"jobs": jobs}


@app.get("/v1/jobs/{job_id}/status")
async def job_status(job_id: str, request: Request, user: User = Depends(get_user_from_token)):
    """Devuelve progreso real de un job de ingesta."""
    if not user:
        raise HTTPException(status_code=401, detail="Auth required")

    job = db.get_ingestion_job(job_id)
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    if job["user_id"] != user.id:
        raise HTTPException(status_code=403, detail="Access denied")

    async with httpx.AsyncClient(timeout=10) as client:
        try:
            resp = await client.get(f"{INGESTOR_URL}/v1/status?task_id={job['task_id']}")
            result = resp.json()
        except Exception:
            # Si el ingestor no responde, devolvemos el estado guardado en SQLite
            return {
                "job_id": job_id,
                "state": job["state"],
                "progress": job["progress"],
                "tier": job["tier"],
                "page_count": job["page_count"],
                "filename": job["filename"],
                "collection": job["collection"],
                "created_at": job["created_at"],
            }

    ingestor_state = result.get("state", "UNKNOWN")
    nv = result.get("nv_ingest_status", {})
    res = result.get("result", {})
    total = max(res.get("total_documents", 1), 1)
    extracted = nv.get("extraction_completed", 0)
    completed = res.get("documents_completed", 0)

    if ingestor_state == "FINISHED":
        progress = 100
        new_state = "completed"
    elif ingestor_state == "FAILED":
        progress = job["progress"]
        new_state = "failed"
    else:
        progress = int((extracted / total * 60) + (completed / total * 40))
        new_state = "running" if progress > 0 else "pending"

    completed_at = None
    if new_state in ("completed", "failed"):
        from datetime import datetime
        completed_at = datetime.now().isoformat()

    db.update_ingestion_job(job_id, new_state, progress, completed_at)

    return {
        "job_id": job_id,
        "state": new_state,
        "progress": progress,
        "tier": job["tier"],
        "page_count": job["page_count"],
        "filename": job["filename"],
        "collection": job["collection"],
        "created_at": job["created_at"],
    }
```

**Step 4: Correr todos los tests del gateway**

```bash
uv run pytest saldivia/tests/test_gateway_ingest.py -v
uv run pytest saldivia/tests/ -v
```

Esperado: todos PASS.

**Step 5: Commit**

```bash
git add saldivia/gateway.py saldivia/tests/test_gateway_ingest.py
git commit -m "feat(gateway): add GET /v1/jobs and GET /v1/jobs/{job_id}/status endpoints"
```

---

## Task 5: Modificar BFF gateway.ts y api/upload

**Files:**
- Modify: `services/sda-frontend/src/lib/server/gateway.ts`
- Modify: `services/sda-frontend/src/routes/api/upload/+server.ts`
- Test: `services/sda-frontend/src/routes/api/upload/upload.test.ts` (actualizar)

**Step 1: Actualizar el test de upload para la nueva respuesta**

En `upload.test.ts`, modificar el test `'forwards multipart al gateway y retorna su respuesta'`:

```typescript
it('devuelve job_id, tier y page_count del gateway', async () => {
    process.env.GATEWAY_URL = 'http://gateway:9000';
    process.env.SYSTEM_API_KEY = 'test-key';

    const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({
            job_id: 'abc-123',
            tier: 'medium',
            page_count: 180,
            filename: 'contrato.pdf'
        }),
    });
    vi.stubGlobal('fetch', mockFetch);

    const { POST } = await import('./+server.js');

    const formData = new FormData();
    formData.append('file', new File(['pdf content'], 'contrato.pdf', { type: 'application/pdf' }));
    formData.append('collection', 'legal');

    const event = {
        request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
        locals: { user: { id: 42, email: 'test@test.com', role: 'admin', area_id: 1, name: 'Test' } },
    } as any;

    const res = await POST(event);
    const body = await res.json();

    expect(res.status).toBe(200);
    expect(body.job_id).toBe('abc-123');
    expect(body.tier).toBe('medium');
    expect(body.page_count).toBe(180);
});
```

**Step 2: Correr tests — verificar estado actual**

```bash
cd ~/rag-saldivia/services/sda-frontend && pnpm run test:run
```

**Step 3: Agregar gatewayListActiveJobs en gateway.ts**

Al final de `src/lib/server/gateway.ts`, agregar:

```typescript
export interface ActiveJobResponse {
    id: string;
    filename: string;
    collection: string;
    tier: string;
    page_count: number | null;
    state: string;
    progress: number;
    created_at: string;
}

export async function gatewayListActiveJobs(token: string): Promise<ActiveJobResponse[]> {
    const res = await gw<{ jobs: ActiveJobResponse[] }>('/v1/jobs', {
        headers: { 'Authorization': `Bearer ${token}` },
    });
    return res.jobs ?? [];
}

export async function gatewayJobStatus(token: string, jobId: string) {
    return gw<{
        job_id: string;
        state: string;
        progress: number;
        tier: string;
        page_count: number | null;
        filename: string;
        collection: string;
        created_at: string;
    }>(`/v1/jobs/${jobId}/status`, {
        headers: { 'Authorization': `Bearer ${token}` },
    });
}
```

**Nota:** El token del usuario viene en `locals.user.token` (verificar cómo se guarda en `hooks.server.ts`).

**Step 4: Actualizar api/upload/+server.ts**

Reemplazar el contenido de `src/routes/api/upload/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);

    const formData = await request.formData();
    const file = formData.get('file');
    const collection = formData.get('collection');

    if (!file || !(file instanceof File)) throw error(400, 'Se requiere un archivo.');
    if (!collection || typeof collection !== 'string' || !collection.trim()) {
        throw error(400, 'Se requiere seleccionar una colección.');
    }

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    const gw = new FormData();
    gw.append('file', file);
    gw.append('data', JSON.stringify({ collection_name: collection.trim() }));

    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 30_000); // 30s (non-blocking)

    let resp: Response;
    try {
        resp = await fetch(`${gatewayUrl}/v1/documents`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'X-User-Id': String(locals.user.id),
            },
            body: gw,
            signal: controller.signal,
        });
    } catch (err) {
        if ((err as any)?.name === 'AbortError') throw error(504, 'Gateway timeout al subir el documento.');
        throw error(502, 'Gateway inalcanzable.');
    } finally {
        clearTimeout(timer);
    }

    const body = await resp.json().catch(() => ({}));
    return json(body, { status: resp.status });
};
```

**Step 5: Correr tests**

```bash
pnpm run test:run
```

Esperado: todos PASS.

**Step 6: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/server/gateway.ts
git add services/sda-frontend/src/routes/api/upload/+server.ts
git add services/sda-frontend/src/routes/api/upload/upload.test.ts
git commit -m "feat(bff): update upload endpoint for non-blocking response, add job status helpers"
```

---

## Task 6: Nueva ruta BFF api/ingestion/[jobId]/status

**Files:**
- Create: `services/sda-frontend/src/routes/api/ingestion/[jobId]/status/+server.ts`
- Create: `services/sda-frontend/src/routes/api/ingestion/[jobId]/status/status.test.ts`

**Step 1: Escribir el test primero**

Crear `src/routes/api/ingestion/[jobId]/status/status.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('GET /api/ingestion/[jobId]/status', () => {
    beforeEach(() => vi.resetAllMocks());

    it('proxea al gateway y devuelve estado del job', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const mockResponse = {
            job_id: 'j1', state: 'running', progress: 42,
            tier: 'medium', page_count: 180, filename: 'doc.pdf',
            collection: 'col', created_at: '2026-03-22T10:00:00'
        };

        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            status: 200,
            json: () => Promise.resolve(mockResponse),
        }));

        const { GET } = await import('./+server.js');
        const event = {
            params: { jobId: 'j1' },
            locals: { user: { id: 1, token: 'user-jwt' } },
        } as any;

        const res = await GET(event);
        const body = await res.json();

        expect(res.status).toBe(200);
        expect(body.progress).toBe(42);
        expect(body.state).toBe('running');
    });

    it('retorna 401 sin sesión', async () => {
        const { GET } = await import('./+server.js');
        const event = { params: { jobId: 'j1' }, locals: { user: null } } as any;
        await expect(GET(event)).rejects.toMatchObject({ status: 401 });
    });

    it('propaga 404 cuando el job no existe', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: false, status: 404,
            text: () => Promise.resolve('Not found'),
            json: () => Promise.resolve({ detail: 'Job not found' }),
        }));

        const { GET } = await import('./+server.js');
        const event = {
            params: { jobId: 'nonexistent' },
            locals: { user: { id: 1, token: 'jwt' } },
        } as any;

        const res = await GET(event);
        expect(res.status).toBe(404);
    });
});
```

**Step 2: Correr test — debe fallar**

```bash
cd ~/rag-saldivia/services/sda-frontend && pnpm run test:run
```

Esperado: `Cannot find module './+server.js'`

**Step 3: Crear el endpoint**

Crear `src/routes/api/ingestion/[jobId]/status/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    let resp: Response;
    try {
        resp = await fetch(`${gatewayUrl}/v1/jobs/${params.jobId}/status`, {
            headers: {
                'Authorization': `Bearer ${apiKey}`,
                'X-User-Id': String(locals.user.id),
            },
        });
    } catch {
        throw error(502, 'Gateway inalcanzable.');
    }

    const body = await resp.json().catch(() => ({}));
    return json(body, { status: resp.status });
};
```

**Step 4: Correr tests**

```bash
pnpm run test:run
```

Esperado: todos PASS.

**Step 5: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/routes/api/ingestion/
git commit -m "feat(bff): add GET /api/ingestion/[jobId]/status route"
```

---

## Task 7: Modificar upload/+page.server.ts para cargar jobs activos

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/upload/+page.server.ts`

**Step 1: Leer el archivo actual**

```bash
cat services/sda-frontend/src/routes/\(app\)/upload/+page.server.ts
```

**Step 2: Actualizar el load**

Reemplazar el contenido:

```typescript
import type { PageServerLoad } from './$types';
import { gatewayListCollections, gatewayListActiveJobs } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    const token = locals.user?.token ?? '';

    const [collectionsResult, activeJobs] = await Promise.all([
        gatewayListCollections().catch(() => ({ collections: [] })),
        token ? gatewayListActiveJobs(token).catch(() => []) : Promise.resolve([]),
    ]);

    return {
        collections: collectionsResult.collections ?? [],
        activeJobs,
    };
};
```

**Nota:** Si `locals.user.token` no existe (el token se guarda con otro nombre), revisar `hooks.server.ts` y ajustar.

**Step 3: Verificar que el tipo `locals.user` tiene `token`**

```bash
grep -n "token\|locals.user" services/sda-frontend/src/hooks.server.ts | head -20
```

Ajustar el acceso al token según lo que encuentres.

**Step 4: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/routes/\(app\)/upload/+page.server.ts
git commit -m "feat(bff): load active ingestion jobs on upload page for recovery"
```

---

## Task 8: Crear src/lib/ingestion/types.ts

**Files:**
- Create: `services/sda-frontend/src/lib/ingestion/types.ts`
- Create: `services/sda-frontend/src/lib/ingestion/types.test.ts`

**Step 1: Escribir tests primero**

Crear `src/lib/ingestion/types.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { classifyTier, classifyTierBySize, estimateETA, TIER_CONFIG } from './types.js';

describe('classifyTier', () => {
    it('clasifica correctamente por páginas', () => {
        expect(classifyTier(1)).toBe('tiny');
        expect(classifyTier(20)).toBe('tiny');
        expect(classifyTier(21)).toBe('small');
        expect(classifyTier(80)).toBe('small');
        expect(classifyTier(81)).toBe('medium');
        expect(classifyTier(250)).toBe('medium');
        expect(classifyTier(251)).toBe('large');
        expect(classifyTier(5000)).toBe('large');
    });
});

describe('classifyTierBySize', () => {
    it('clasifica por tamaño de archivo', () => {
        expect(classifyTierBySize(50_000)).toBe('tiny');
        expect(classifyTierBySize(99_999)).toBe('tiny');
        expect(classifyTierBySize(100_000)).toBe('small');
        expect(classifyTierBySize(499_999)).toBe('small');
        expect(classifyTierBySize(500_000)).toBe('medium');
        expect(classifyTierBySize(4_999_999)).toBe('medium');
        expect(classifyTierBySize(5_000_000)).toBe('large');
    });
});

describe('estimateETA', () => {
    it('devuelve expectedMaxDuration cuando progress es 0', () => {
        expect(estimateETA('medium', 0, 0)).toBe(TIER_CONFIG.medium.expectedMaxDuration);
    });

    it('devuelve 0 cuando progress es 100', () => {
        expect(estimateETA('tiny', 100, 30)).toBe(0);
    });

    it('calcula ETA correctamente con progreso real', () => {
        // 50% en 60s → total estimado = 120s → restante = 60s
        expect(estimateETA('small', 50, 60)).toBe(60);
    });

    it('nunca devuelve negativo', () => {
        expect(estimateETA('tiny', 99, 10000)).toBe(0);
    });
});

describe('TIER_CONFIG', () => {
    it('todos los tiers tienen los campos requeridos', () => {
        for (const tier of ['tiny', 'small', 'medium', 'large'] as const) {
            const config = TIER_CONFIG[tier];
            expect(config.pollInterval).toBeGreaterThan(0);
            expect(config.deadlockThreshold).toBeGreaterThan(0);
            expect(config.expectedMaxDuration).toBeGreaterThan(0);
            expect(['green', 'blue', 'amber', 'red']).toContain(config.color);
        }
    });
});
```

**Step 2: Correr tests — deben fallar**

```bash
pnpm run test:run -- src/lib/ingestion/types.test.ts
```

**Step 3: Crear types.ts**

Crear `src/lib/ingestion/types.ts` con el contenido exacto del design doc (sección 4, `types.ts`).

**Step 4: Correr tests — deben pasar**

```bash
pnpm run test:run -- src/lib/ingestion/types.test.ts
```

**Step 5: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/ingestion/
git commit -m "feat(frontend): add ingestion types, tier config, and classification helpers"
```

---

## Task 9: Crear src/lib/ingestion/poller.ts

**Files:**
- Create: `services/sda-frontend/src/lib/ingestion/poller.ts`
- Create: `services/sda-frontend/src/lib/ingestion/poller.test.ts`

**Step 1: Escribir tests**

Crear `src/lib/ingestion/poller.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { IngestPoller } from './poller.js';

function sleep(ms: number) { return new Promise(r => setTimeout(r, ms)); }

describe('IngestPoller', () => {
    beforeEach(() => vi.resetAllMocks());

    it('llama onUpdate con el estado del servidor', async () => {
        vi.useFakeTimers();

        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ state: 'completed', progress: 100 }),
        }));

        const updates: any[] = [];
        const poller = new IngestPoller('job-1', 'tiny');
        await poller.poll((s) => updates.push(s));

        expect(updates.length).toBeGreaterThan(0);
        expect(updates.at(-1)?.state).toBe('completed');

        vi.useRealTimers();
    });

    it('para el loop cuando state es completed', async () => {
        let callCount = 0;
        vi.stubGlobal('fetch', vi.fn().mockImplementation(() => {
            callCount++;
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve({ state: 'completed', progress: 100 }),
            });
        }));

        const poller = new IngestPoller('job-2', 'tiny');
        await poller.poll(() => {});

        expect(callCount).toBe(1); // Solo 1 call porque ya completó
    });

    it('detecta deadlock y emite state=stalled', async () => {
        vi.useFakeTimers();

        // Siempre devuelve progress=10 (sin cambio)
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ state: 'running', progress: 10 }),
        }));

        const updates: any[] = [];
        const poller = new IngestPoller('job-3', 'tiny'); // deadlockThreshold = 30s

        const pollPromise = poller.poll((s) => updates.push(s));

        // Simular que pasan 31 segundos sin progreso
        await vi.advanceTimersByTimeAsync(31_000);
        await pollPromise.catch(() => {});

        const stalled = updates.find(u => u.state === 'stalled');
        expect(stalled).toBeDefined();

        vi.useRealTimers();
    });

    it('stop() detiene el loop', async () => {
        let callCount = 0;
        vi.stubGlobal('fetch', vi.fn().mockImplementation(() => {
            callCount++;
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve({ state: 'running', progress: callCount * 10 }),
            });
        }));

        const poller = new IngestPoller('job-4', 'tiny');
        const pollPromise = poller.poll(() => {});
        poller.stop();
        await pollPromise.catch(() => {});

        expect(callCount).toBeLessThanOrEqual(2);
    });
});
```

**Step 2: Correr tests — deben fallar**

```bash
pnpm run test:run -- src/lib/ingestion/poller.test.ts
```

**Step 3: Crear poller.ts**

Crear `src/lib/ingestion/poller.ts`:

```typescript
import { TIER_CONFIG, estimateETA, type Tier } from './types.js';

function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export class IngestPoller {
    private jobId: string;
    private tier: Tier;
    private lastProgress: number = -1;
    private lastProgressAt: number = Date.now();
    private startedAt: number = Date.now();
    private stopped = false;

    constructor(jobId: string, tier: Tier) {
        this.jobId = jobId;
        this.tier = tier;
    }

    stop(): void {
        this.stopped = true;
    }

    async poll(onUpdate: (update: {
        state: string;
        progress: number;
        eta: number | null;
    }) => void): Promise<void> {
        const config = TIER_CONFIG[this.tier];
        this.startedAt = Date.now();

        while (!this.stopped) {
            let data: { state: string; progress: number };

            try {
                const resp = await fetch(`/api/ingestion/${this.jobId}/status`);
                if (!resp.ok) {
                    onUpdate({ state: 'failed', progress: 0, eta: null });
                    break;
                }
                data = await resp.json();
            } catch {
                onUpdate({ state: 'failed', progress: 0, eta: null });
                break;
            }

            const now = Date.now();
            const elapsedSinceStart = (now - this.startedAt) / 1000;
            const elapsedSinceProgress = (now - this.lastProgressAt) / 1000;
            const eta = estimateETA(this.tier, data.progress, elapsedSinceStart);

            // Deadlock detection
            if (data.progress === this.lastProgress) {
                if (elapsedSinceProgress > config.deadlockThreshold) {
                    onUpdate({ state: 'stalled', progress: data.progress, eta: null });
                    break;
                }
            } else {
                this.lastProgress = data.progress;
                this.lastProgressAt = now;
            }

            onUpdate({ state: data.state, progress: data.progress, eta });

            if (data.state === 'completed' || data.state === 'failed') break;

            await sleep(config.pollInterval * 1000);
        }
    }
}
```

**Step 4: Correr tests**

```bash
pnpm run test:run -- src/lib/ingestion/poller.test.ts
```

**Step 5: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/ingestion/poller.ts
git add services/sda-frontend/src/lib/ingestion/poller.test.ts
git commit -m "feat(frontend): add IngestPoller with deadlock detection"
```

---

## Task 10: Crear stores/ingestion.svelte.ts

**Files:**
- Create: `services/sda-frontend/src/lib/stores/ingestion.svelte.ts`
- Create: `services/sda-frontend/src/lib/stores/ingestion.test.ts`

**Step 1: Escribir tests**

Crear `src/lib/stores/ingestion.test.ts`:

```typescript
import { describe, it, expect, beforeEach } from 'vitest';

// Reimportar el módulo fresco en cada test
async function freshStore() {
    vi.resetModules();
    return import('./ingestion.svelte.js');
}

import { vi } from 'vitest';

describe('ingestionStore', () => {
    it('addJob agrega un job al store', async () => {
        const { ingestionStore } = await freshStore();
        const job = {
            jobId: 'j1', filename: 'a.pdf', collection: 'col',
            tier: 'tiny' as const, pageCount: 10, state: 'pending' as const,
            progress: 0, eta: null, startedAt: Date.now(), lastProgressAt: Date.now(),
        };
        ingestionStore.addJob(job);
        expect(ingestionStore.jobs.find(j => j.jobId === 'j1')).toBeDefined();
    });

    it('updateJob actualiza solo el job correcto', async () => {
        const { ingestionStore } = await freshStore();
        ingestionStore.addJob({
            jobId: 'j2', filename: 'b.pdf', collection: 'col',
            tier: 'small' as const, pageCount: 50, state: 'pending' as const,
            progress: 0, eta: null, startedAt: Date.now(), lastProgressAt: Date.now(),
        });
        ingestionStore.updateJob('j2', { progress: 42, state: 'running' });
        const job = ingestionStore.jobs.find(j => j.jobId === 'j2');
        expect(job?.progress).toBe(42);
        expect(job?.state).toBe('running');
    });

    it('removeJob elimina el job', async () => {
        const { ingestionStore } = await freshStore();
        ingestionStore.addJob({
            jobId: 'j3', filename: 'c.pdf', collection: 'col',
            tier: 'medium' as const, pageCount: 100, state: 'completed' as const,
            progress: 100, eta: 0, startedAt: Date.now(), lastProgressAt: Date.now(),
        });
        ingestionStore.removeJob('j3');
        expect(ingestionStore.jobs.find(j => j.jobId === 'j3')).toBeUndefined();
    });

    it('hydrateFromServer no duplica jobs existentes', async () => {
        const { ingestionStore } = await freshStore();
        const existing = {
            jobId: 'j4', filename: 'd.pdf', collection: 'col',
            tier: 'tiny' as const, pageCount: 5, state: 'running' as const,
            progress: 30, eta: null, startedAt: Date.now(), lastProgressAt: Date.now(),
        };
        ingestionStore.addJob(existing);
        ingestionStore.hydrateFromServer([
            { id: 'j4', filename: 'd.pdf', collection: 'col', tier: 'tiny',
              page_count: 5, state: 'running', progress: 30, created_at: '' }
        ]);
        expect(ingestionStore.jobs.filter(j => j.jobId === 'j4').length).toBe(1);
    });
});
```

**Step 2: Correr — debe fallar**

```bash
pnpm run test:run -- src/lib/stores/ingestion.test.ts
```

**Step 3: Crear el store**

Crear `src/lib/stores/ingestion.svelte.ts`:

```typescript
import type { Tier } from '$lib/ingestion/types.js';

export type JobState = 'pending' | 'running' | 'completed' | 'failed' | 'stalled';

export interface IngestionJob {
    jobId: string;
    filename: string;
    collection: string;
    tier: Tier;
    pageCount: number | null;
    state: JobState;
    progress: number;
    eta: number | null;
    startedAt: number;
    lastProgressAt: number;
}

export interface ServerJob {
    id: string;
    filename: string;
    collection: string;
    tier: string;
    page_count: number | null;
    state: string;
    progress: number;
    created_at: string;
}

let jobs = $state<IngestionJob[]>([]);

export const ingestionStore = {
    get jobs(): IngestionJob[] {
        return jobs;
    },

    addJob(job: IngestionJob): void {
        jobs = [...jobs, job];
    },

    updateJob(jobId: string, updates: Partial<IngestionJob>): void {
        jobs = jobs.map(j => j.jobId === jobId ? { ...j, ...updates } : j);
    },

    removeJob(jobId: string): void {
        jobs = jobs.filter(j => j.jobId !== jobId);
    },

    hydrateFromServer(serverJobs: ServerJob[]): void {
        for (const sj of serverJobs) {
            if (jobs.find(j => j.jobId === sj.id)) continue;
            jobs = [...jobs, {
                jobId: sj.id,
                filename: sj.filename,
                collection: sj.collection,
                tier: sj.tier as Tier,
                pageCount: sj.page_count,
                state: sj.state as JobState,
                progress: sj.progress,
                eta: null,
                startedAt: new Date(sj.created_at).getTime(),
                lastProgressAt: Date.now(),
            }];
        }
    },
};
```

**Step 4: Correr tests**

```bash
pnpm run test:run -- src/lib/stores/ingestion.test.ts
```

**Step 5: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/stores/ingestion.svelte.ts
git add services/sda-frontend/src/lib/stores/ingestion.test.ts
git commit -m "feat(frontend): add ingestion store with hydration support"
```

---

## Task 11: Crear componentes TierBadge y JobCard

**Files:**
- Create: `services/sda-frontend/src/lib/components/upload/TierBadge.svelte`
- Create: `services/sda-frontend/src/lib/components/upload/JobCard.svelte`
- Create: `services/sda-frontend/src/lib/components/upload/JobCard.test.ts`

**Step 1: Crear TierBadge.svelte**

```svelte
<script lang="ts">
    import { TIER_CONFIG, type Tier } from '$lib/ingestion/types.js';

    let { tier }: { tier: Tier } = $props();

    const config = $derived(TIER_CONFIG[tier]);

    const colorClasses: Record<string, string> = {
        green: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
        blue:  'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
        amber: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
        red:   'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    };
</script>

<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {colorClasses[config.color]}">
    {config.label}
</span>
```

**Step 2: Crear JobCard.svelte**

```svelte
<script lang="ts">
    import { FileText, CheckCircle, XCircle, AlertTriangle, RefreshCw } from 'lucide-svelte';
    import TierBadge from './TierBadge.svelte';
    import type { IngestionJob } from '$lib/stores/ingestion.svelte.js';

    let { job, onRetry }: {
        job: IngestionJob;
        onRetry?: (jobId: string) => void;
    } = $props();

    const stateLabel: Record<string, string> = {
        pending:   'En cola',
        running:   job.progress < 60 ? 'Extrayendo' : 'Indexando',
        completed: 'Completado',
        failed:    'Error',
        stalled:   'Sin progreso',
    };

    const isFinished = $derived(job.state === 'completed');
    const isError = $derived(job.state === 'failed' || job.state === 'stalled');
    const isActive = $derived(job.state === 'pending' || job.state === 'running');
</script>

<div class="flex items-center gap-3 py-3 border-b border-[var(--border)] last:border-0">
    <div class="shrink-0">
        {#if isFinished}
            <CheckCircle size={18} class="text-green-500" />
        {:else if isError}
            <XCircle size={18} class="text-[var(--danger)]" />
        {:else}
            <FileText size={18} class="text-[var(--text-faint)]" />
        {/if}
    </div>

    <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-0.5">
            <span class="text-sm font-medium text-[var(--text)] truncate">{job.filename}</span>
            <TierBadge tier={job.tier} />
        </div>

        <div class="flex items-center gap-2 text-xs text-[var(--text-muted)]">
            <span>{stateLabel[job.state] ?? job.state}</span>
            {#if job.pageCount}
                <span>·</span>
                <span>{job.pageCount} págs</span>
            {/if}
            {#if isActive && job.eta !== null && job.eta > 0}
                <span>·</span>
                <span>~{job.eta}s</span>
            {/if}
        </div>

        {#if isActive || isError}
            <div class="mt-1.5 h-1.5 bg-[var(--bg-hover)] rounded-full overflow-hidden">
                <div
                    class="h-full transition-all duration-500 rounded-full {isError ? 'bg-[var(--danger)]' : 'bg-[var(--accent)]'}"
                    style="width: {job.progress}%"
                ></div>
            </div>
        {/if}
    </div>

    <div class="shrink-0 text-right">
        {#if isActive}
            <span class="text-sm font-medium text-[var(--text-muted)]">{job.progress}%</span>
        {:else if isError && onRetry}
            <button
                onclick={() => onRetry?.(job.jobId)}
                class="flex items-center gap-1 text-xs text-[var(--accent)] hover:underline"
            >
                <RefreshCw size={12} />
                Reintentar
            </button>
        {/if}
    </div>
</div>
```

**Step 3: Escribir test de JobCard**

Crear `src/lib/components/upload/JobCard.test.ts`:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/svelte';
import JobCard from './JobCard.svelte';

const baseJob = {
    jobId: 'j1', filename: 'contrato.pdf', collection: 'legal',
    tier: 'medium' as const, pageCount: 180,
    state: 'running' as const, progress: 75,
    eta: 30, startedAt: Date.now(), lastProgressAt: Date.now(),
};

describe('JobCard', () => {
    it('muestra el nombre del archivo', () => {
        const { getByText } = render(JobCard, { props: { job: baseJob } });
        expect(getByText('contrato.pdf')).toBeTruthy();
    });

    it('barra de progreso al 75%', () => {
        const { container } = render(JobCard, { props: { job: baseJob } });
        const bar = container.querySelector('[style*="width: 75%"]');
        expect(bar).toBeTruthy();
    });

    it('muestra botón reintentar cuando state=stalled', () => {
        const stalledJob = { ...baseJob, state: 'stalled' as const };
        const onRetry = vi.fn();
        const { getByText } = render(JobCard, { props: { job: stalledJob, onRetry } });
        expect(getByText('Reintentar')).toBeTruthy();
    });

    it('no muestra botón reintentar cuando state=running', () => {
        const { queryByText } = render(JobCard, { props: { job: baseJob } });
        expect(queryByText('Reintentar')).toBeNull();
    });

    it('muestra checkmark cuando state=completed', () => {
        const done = { ...baseJob, state: 'completed' as const, progress: 100 };
        const { container } = render(JobCard, { props: { job: done } });
        // El ícono CheckCircle renderiza un SVG
        expect(container.querySelector('svg')).toBeTruthy();
    });
});
```

**Step 4: Correr tests**

```bash
pnpm run test:run -- src/lib/components/upload/JobCard.test.ts
```

**Step 5: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/components/upload/
git commit -m "feat(frontend): add TierBadge and JobCard components"
```

---

## Task 12: Crear DropZone e IngestionQueue

**Files:**
- Create: `services/sda-frontend/src/lib/components/upload/DropZone.svelte`
- Create: `services/sda-frontend/src/lib/components/upload/IngestionQueue.svelte`

**Step 1: Crear DropZone.svelte**

```svelte
<script lang="ts">
    import { Upload, File as FileIcon, X } from 'lucide-svelte';
    import TierBadge from './TierBadge.svelte';
    import { classifyTierBySize, type Tier } from '$lib/ingestion/types.js';

    let {
        onFile,
        disabled = false,
    }: {
        onFile: (file: File) => void;
        disabled?: boolean;
    } = $props();

    const ACCEPTED_EXTENSIONS = ['.pdf', '.txt', '.md', '.docx'];
    const ACCEPTED_MIME = [
        'application/pdf', 'text/plain', 'text/markdown',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ];

    let selectedFile = $state<File | null>(null);
    let dragging = $state(false);
    let fileError = $state('');
    let estimatedTier = $derived<Tier | null>(
        selectedFile ? classifyTierBySize(selectedFile.size) : null
    );

    function validateFile(file: File): string {
        const ext = '.' + (file.name.split('.').pop()?.toLowerCase() ?? '');
        if (!ACCEPTED_EXTENSIONS.includes(ext))
            return `Tipo no soportado. Acepta: ${ACCEPTED_EXTENSIONS.join(', ')}`;
        if (file.size > 50 * 1024 * 1024)
            return 'El archivo no puede superar 50 MB.';
        return '';
    }

    function handleFileSelect(files: FileList | null) {
        if (!files || files.length === 0) return;
        const file = files[0];
        fileError = validateFile(file);
        if (!fileError) {
            selectedFile = file;
            onFile(file);
        } else {
            selectedFile = null;
        }
    }

    function clearFile(e: MouseEvent) {
        e.stopPropagation();
        selectedFile = null;
        fileError = '';
    }
</script>

<div
    role="button"
    tabindex={disabled ? -1 : 0}
    ondragenter={(e) => { e.preventDefault(); dragging = true; }}
    ondragleave={(e) => { e.preventDefault(); dragging = false; }}
    ondragover={(e) => e.preventDefault()}
    ondrop={(e) => { e.preventDefault(); dragging = false; handleFileSelect(e.dataTransfer?.files ?? null); }}
    onclick={() => !disabled && document.getElementById('file-input-dz')?.click()}
    onkeydown={(e) => e.key === 'Enter' && !disabled && document.getElementById('file-input-dz')?.click()}
    class="border-2 border-dashed rounded-[var(--radius-lg)] p-8 text-center transition-colors
           {disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
           {dragging ? 'border-[var(--accent)] bg-[var(--accent)]/5'
                     : 'border-[var(--border)] hover:border-[var(--text-faint)]'}"
>
    <input
        id="file-input-dz"
        type="file"
        class="hidden"
        accept={ACCEPTED_MIME.join(',')}
        onchange={(e) => handleFileSelect((e.target as HTMLInputElement).files)}
        {disabled}
    />

    {#if selectedFile}
        <div class="flex items-center justify-center gap-3">
            <FileIcon size={20} class="text-[var(--accent)]" />
            <span class="text-sm font-medium text-[var(--text)]">{selectedFile.name}</span>
            {#if estimatedTier}
                <TierBadge tier={estimatedTier} />
            {/if}
            <button onclick={clearFile} class="text-[var(--text-faint)] hover:text-[var(--text)]" aria-label="Quitar archivo">
                <X size={16} />
            </button>
        </div>
        <p class="text-xs text-[var(--text-faint)] mt-1.5">
            {(selectedFile.size / 1024).toFixed(1)} KB
            {#if estimatedTier}<span class="ml-1">(tamaño estimado — tier exacto disponible tras subir)</span>{/if}
        </p>
    {:else}
        <Upload size={24} class="text-[var(--text-faint)] mx-auto mb-3" />
        <p class="text-sm text-[var(--text-muted)]">
            Arrastrá un archivo o <span class="text-[var(--accent)]">hacé click para elegir</span>
        </p>
        <p class="text-xs text-[var(--text-faint)] mt-1">
            {ACCEPTED_EXTENSIONS.join(', ')} · máx 50 MB
        </p>
    {/if}
</div>

{#if fileError}
    <p class="text-xs text-[var(--danger)] mt-2">{fileError}</p>
{/if}
```

**Step 2: Crear IngestionQueue.svelte**

```svelte
<script lang="ts">
    import JobCard from './JobCard.svelte';
    import type { IngestionJob } from '$lib/stores/ingestion.svelte.js';

    let { jobs, onRetry }: {
        jobs: IngestionJob[];
        onRetry?: (jobId: string) => void;
    } = $props();

    // Activos primero (por progreso desc), luego completados (últimos 5), luego fallados
    const sortedJobs = $derived([
        ...jobs
            .filter(j => j.state === 'pending' || j.state === 'running')
            .sort((a, b) => b.progress - a.progress),
        ...jobs
            .filter(j => j.state === 'completed')
            .slice(0, 5),
        ...jobs
            .filter(j => j.state === 'failed' || j.state === 'stalled'),
    ]);
</script>

{#if sortedJobs.length > 0}
    <div class="mt-6">
        <h2 class="text-xs font-medium text-[var(--text-muted)] uppercase tracking-wide mb-2">
            Cola de ingesta
        </h2>
        <div class="border border-[var(--border)] rounded-[var(--radius-lg)] divide-y divide-[var(--border)] px-3">
            {#each sortedJobs as job (job.jobId)}
                <JobCard {job} {onRetry} />
            {/each}
        </div>
    </div>
{/if}
```

**Step 3: Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/components/upload/DropZone.svelte
git add services/sda-frontend/src/lib/components/upload/IngestionQueue.svelte
git commit -m "feat(frontend): add DropZone and IngestionQueue components"
```

---

## Task 13: Reescribir upload/+page.svelte

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/upload/+page.svelte`

**Step 1: Leer el archivo actual** (ya leído en exploración, pero confirmar línea 1)

**Step 2: Reescribir +page.svelte**

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import { ingestionStore } from '$lib/stores/ingestion.svelte.js';
    import { IngestPoller } from '$lib/ingestion/poller.js';
    import DropZone from '$lib/components/upload/DropZone.svelte';
    import IngestionQueue from '$lib/components/upload/IngestionQueue.svelte';
    import type { Tier } from '$lib/ingestion/types.js';

    let { data } = $props();

    let selectedFile = $state<File | null>(null);
    let selectedCollection = $state(data.collections[0] ?? '');
    let uploading = $state(false);

    // Hidratar jobs activos desde el servidor al montar
    onMount(() => {
        if (data.activeJobs?.length) {
            ingestionStore.hydrateFromServer(data.activeJobs);
            // Reanudar pollers para jobs activos recuperados
            for (const job of ingestionStore.jobs) {
                if (job.state === 'pending' || job.state === 'running') {
                    startPoller(job.jobId, job.tier);
                }
            }
        }
    });

    function startPoller(jobId: string, tier: Tier) {
        const poller = new IngestPoller(jobId, tier);
        poller.poll(({ state, progress, eta }) => {
            ingestionStore.updateJob(jobId, {
                state: state as any,
                progress,
                eta,
            });
            if (state === 'completed') {
                toastStore.success(`"${ingestionStore.jobs.find(j => j.jobId === jobId)?.filename}" indexado correctamente.`);
                // Remover del store después de 5 segundos
                setTimeout(() => ingestionStore.removeJob(jobId), 5_000);
            }
            if (state === 'failed') {
                toastStore.error(`Error al ingestar "${ingestionStore.jobs.find(j => j.jobId === jobId)?.filename}".`);
            }
        });
    }

    async function handleUpload() {
        if (!selectedFile || !selectedCollection || uploading) return;
        uploading = true;

        try {
            const formData = new FormData();
            formData.append('file', selectedFile);
            formData.append('collection', selectedCollection);

            const res = await fetch('/api/upload', { method: 'POST', body: formData });
            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                throw new Error((body as any).message ?? `Error ${res.status}`);
            }

            const { job_id, tier, page_count, filename } = await res.json();

            // Registrar en el store e iniciar poller
            ingestionStore.addJob({
                jobId: job_id,
                filename,
                collection: selectedCollection,
                tier: tier as Tier,
                pageCount: page_count,
                state: 'pending',
                progress: 0,
                eta: null,
                startedAt: Date.now(),
                lastProgressAt: Date.now(),
            });

            startPoller(job_id, tier as Tier);
            toastStore.success(`"${filename}" enviado a ingesta.`);
            selectedFile = null;

        } catch (e: any) {
            toastStore.error(e.message ?? 'Error al subir el archivo.');
        } finally {
            uploading = false;
        }
    }

    function handleRetry(jobId: string) {
        const job = ingestionStore.jobs.find(j => j.jobId === jobId);
        if (!job) return;
        ingestionStore.updateJob(jobId, { state: 'pending', progress: 0, eta: null });
        startPoller(jobId, job.tier);
    }
</script>

<div class="p-6 max-w-xl">
    <h1 class="text-lg font-semibold text-[var(--text)] mb-6">Subir documentos</h1>

    <DropZone onFile={(f) => (selectedFile = f)} disabled={uploading} />

    <!-- Collection selector -->
    <div class="mt-5">
        <label for="collection-select" class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
            Colección destino
        </label>
        {#if data.collections.length === 0}
            <p class="text-sm text-[var(--text-faint)]">
                No hay colecciones disponibles.
                <a href="/collections" class="text-[var(--accent)] hover:underline">Creá una primero.</a>
            </p>
        {:else}
            <select
                id="collection-select"
                bind:value={selectedCollection}
                class="w-full bg-[var(--bg-base)] border border-[var(--border)] rounded-lg
                       px-3 py-2 text-sm text-[var(--text)] focus:outline-none
                       focus:border-[var(--accent)] transition-colors"
            >
                {#each data.collections as col (col)}
                    <option value={col}>{col}</option>
                {/each}
            </select>
        {/if}
    </div>

    <!-- Upload button -->
    <button
        onclick={handleUpload}
        disabled={!selectedFile || !selectedCollection || uploading}
        class="mt-5 w-full py-2.5 px-4 text-sm font-medium text-white bg-[var(--accent)]
               rounded-lg hover:opacity-90 transition-opacity
               disabled:opacity-40 disabled:cursor-not-allowed"
    >
        {uploading ? 'Enviando...' : 'Subir documento'}
    </button>

    <!-- Queue de jobs -->
    <IngestionQueue jobs={ingestionStore.jobs} onRetry={handleRetry} />
</div>
```

**Step 3: Verificar que TypeScript compila**

```bash
cd ~/rag-saldivia/services/sda-frontend && pnpm exec tsc --noEmit
```

Corregir cualquier error de tipos antes de continuar.

**Step 4: Correr todos los tests del frontend**

```bash
pnpm run test:run
```

Esperado: todos PASS.

**Step 5: Correr suite completa de Python para verificar no hay regresiones**

```bash
cd ~/rag-saldivia && uv run pytest saldivia/tests/ -v
```

**Step 6: Commit final**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/routes/\(app\)/upload/+page.svelte
git commit -m "feat(frontend): rewrite upload page with real-time ingestion queue and job recovery"
```

---

## Verificación final

```bash
# Todos los tests Python
cd ~/rag-saldivia && uv run pytest saldivia/tests/ -v

# Todos los tests frontend
cd ~/rag-saldivia/services/sda-frontend && pnpm run test:run

# Type check
pnpm exec tsc --noEmit
```

Todos deben pasar antes de considerar la Fase 6 completa.
