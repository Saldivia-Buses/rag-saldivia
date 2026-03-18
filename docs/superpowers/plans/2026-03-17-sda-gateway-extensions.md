# SDA Gateway Extensions — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the existing FastAPI gateway with JWT session auth, admin CRUD endpoints, chat history, and collection stats — the 20 new endpoints required by the SDA frontend.

**Architecture:** Add JWT-based session auth alongside existing Bearer token auth. The BFF (SvelteKit) uses the Bearer `SYSTEM_API_KEY`; JWT is issued to the browser session and decoded by the BFF, never reaching the gateway. New endpoints for user/area management use the existing `AuthDB` class with a few new methods. Chat sessions get their own table in `auth.db`.

**Tech Stack:** FastAPI, PyJWT, bcrypt, SQLite (existing), pytest

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `pyproject.toml` | Modify | Add PyJWT, bcrypt dependencies |
| `saldivia/auth/models.py` | Modify | Add `password_hash` to User; add `ChatSession`, `ChatMessage` dataclasses |
| `saldivia/auth/database.py` | Modify | Add `get_user_by_id`, `update_user`, `update_api_key`, `verify_password`, `set_password`, `get_chat_sessions`, `get_chat_session`, `create_chat_session`, `delete_chat_session`, `add_chat_message`; add schema migration for `password_hash` and `chat_sessions`/`chat_messages` tables |
| `saldivia/gateway.py` | Modify | Add 20 new endpoints across 5 groups: JWT auth, admin/users, admin/areas, chat/sessions, collection stats + extended audit |
| `saldivia/tests/test_gateway_extended.py` | Create | Tests for all new endpoints |

---

### Task 1: Password field + JWT dependency

**Files:**
- Modify: `pyproject.toml`
- Modify: `saldivia/auth/models.py`
- Modify: `saldivia/auth/database.py`

The `User` dataclass needs `password_hash`. The schema needs `password_hash TEXT` on the `users` table and a migration that adds it if missing (SQLite `ALTER TABLE ADD COLUMN` is safe when using `IF NOT EXISTS`-style migration).

- [ ] **Add dependencies**

Edit `pyproject.toml` to add `PyJWT>=2.8.0` and `bcrypt>=4.1.0` to the `dependencies` list.

- [ ] **Update User dataclass**

In `saldivia/auth/models.py`, add `password_hash: Optional[str] = None` to the `User` dataclass after `api_key_hash`:

```python
@dataclass
class User:
    id: int
    email: str
    name: str
    area_id: int
    role: Role
    api_key_hash: str
    password_hash: Optional[str] = None
    created_at: datetime = field(default_factory=datetime.now)
    last_login: Optional[datetime] = None
    active: bool = True
```

Add password helpers at the bottom of `models.py`:

```python
def hash_password(password: str) -> str:
    """Hash a password with bcrypt."""
    import bcrypt
    return bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()


def verify_password(password: str, password_hash: str) -> bool:
    """Verify a password against its bcrypt hash."""
    import bcrypt
    return bcrypt.checkpw(password.encode(), password_hash.encode())
```

- [ ] **Migrate schema**

In `saldivia/auth/database.py`, inside `init_db()`, add after the existing `executescript`:

```python
# Migration: add password_hash column if not present (idempotent)
conn.execute("ALTER TABLE users ADD COLUMN password_hash TEXT")
```

Wrap the ALTER in a try/except since SQLite raises if the column already exists:

```python
try:
    conn.execute("ALTER TABLE users ADD COLUMN password_hash TEXT")
except Exception:
    pass  # Column already exists
conn.close()
```

- [ ] **Add DB helper methods**

In `saldivia/auth/database.py`, add to the `AuthDB` class:

```python
def get_user_by_id(self, user_id: int) -> Optional[User]:
    with self._conn() as conn:
        row = conn.execute(
            "SELECT id, email, name, area_id, role, api_key_hash, password_hash, "
            "created_at, last_login, active FROM users WHERE id = ? AND active = 1",
            (user_id,)
        ).fetchone()
        if row:
            return User(id=row[0], email=row[1], name=row[2], area_id=row[3],
                        role=Role(row[4]), api_key_hash=row[5], password_hash=row[6],
                        created_at=row[7], last_login=row[8], active=row[9])
        return None

def get_user_by_email(self, email: str) -> Optional[User]:
    with self._conn() as conn:
        row = conn.execute(
            "SELECT id, email, name, area_id, role, api_key_hash, password_hash, "
            "created_at, last_login, active FROM users WHERE email = ? AND active = 1",
            (email,)
        ).fetchone()
        if row:
            return User(id=row[0], email=row[1], name=row[2], area_id=row[3],
                        role=Role(row[4]), api_key_hash=row[5], password_hash=row[6],
                        created_at=row[7], last_login=row[8], active=row[9])
        return None

def update_user(self, user_id: int, **fields):
    """Update user fields. Allowed: name, area_id, role, active."""
    allowed = {"name", "area_id", "role", "active"}
    updates = {k: v for k, v in fields.items() if k in allowed}
    if not updates:
        return
    # Convert Role enum to value if needed
    if "role" in updates and isinstance(updates["role"], Role):
        updates["role"] = updates["role"].value
    set_clause = ", ".join(f"{k} = ?" for k in updates)
    values = list(updates.values()) + [user_id]
    with self._conn() as conn:
        conn.execute(f"UPDATE users SET {set_clause} WHERE id = ?", values)

def update_api_key(self, user_id: int, api_key_hash: str):
    with self._conn() as conn:
        conn.execute("UPDATE users SET api_key_hash = ? WHERE id = ?", (api_key_hash, user_id))

def set_password(self, user_id: int, password_hash: str):
    with self._conn() as conn:
        conn.execute("UPDATE users SET password_hash = ? WHERE id = ?", (password_hash, user_id))

def update_last_login(self, user_id: int):
    from datetime import datetime
    with self._conn() as conn:
        conn.execute("UPDATE users SET last_login = ? WHERE id = ?",
                     (datetime.now().isoformat(), user_id))
```

Also update `create_user()` to accept optional `password_hash`:

```python
def create_user(self, email: str, name: str, area_id: int, role: Role,
                api_key_hash: str, password_hash: Optional[str] = None) -> User:
    with self._conn() as conn:
        cur = conn.execute(
            "INSERT INTO users (email, name, area_id, role, api_key_hash, password_hash) "
            "VALUES (?, ?, ?, ?, ?, ?)",
            (email, name, area_id, role.value, api_key_hash, password_hash)
        )
        return User(id=cur.lastrowid, email=email, name=name, area_id=area_id,
                    role=role, api_key_hash=api_key_hash, password_hash=password_hash)
```

Also update `list_users()` and `get_user_by_api_key_hash()` to SELECT `password_hash`. Here are the updated methods:

```python
def list_users(self, area_id: int = None, active_only: bool = True) -> list[User]:
    query = ("SELECT id, email, name, area_id, role, api_key_hash, password_hash, "
             "created_at, last_login, active FROM users WHERE 1=1")
    params = []
    if area_id is not None:
        query += " AND area_id = ?"
        params.append(area_id)
    if active_only:
        query += " AND active = 1"
    with self._conn() as conn:
        rows = conn.execute(query, params).fetchall()
    return [User(id=r[0], email=r[1], name=r[2], area_id=r[3], role=Role(r[4]),
                 api_key_hash=r[5], password_hash=r[6],
                 created_at=r[7], last_login=r[8], active=bool(r[9]))
            for r in rows]

def get_user_by_api_key_hash(self, api_key_hash: str) -> Optional[User]:
    with self._conn() as conn:
        row = conn.execute(
            "SELECT id, email, name, area_id, role, api_key_hash, password_hash, "
            "created_at, last_login, active FROM users WHERE api_key_hash = ? AND active = 1",
            (api_key_hash,)
        ).fetchone()
    if not row:
        return None
    return User(id=row[0], email=row[1], name=row[2], area_id=row[3], role=Role(row[4]),
                api_key_hash=row[5], password_hash=row[6],
                created_at=row[7], last_login=row[8], active=bool(row[9]))
```

**Important:** These replace (not add alongside) the existing `list_users()` and `get_user_by_api_key_hash()` methods in `database.py`. The indices now include `password_hash` at position `row[6]`, shifting `created_at`, `last_login`, `active` to `row[7]`, `row[8]`, `row[9]`.

- [ ] **Write failing test**

Create `saldivia/tests/test_gateway_extended.py`:

```python
"""Tests for SDA gateway extensions."""
import pytest
from saldivia.auth.models import hash_password, verify_password

def test_password_hashing():
    pw = "supersecret123"
    hashed = hash_password(pw)
    assert verify_password(pw, hashed)
    assert not verify_password("wrong", hashed)
    assert hashed != pw
```

- [ ] **Run test to verify it fails**

```bash
cd ~/rag-saldivia && pip install -e ".[dev]" -q && pytest saldivia/tests/test_gateway_extended.py -v
```

Expected: FAIL — `hash_password` not defined yet (or ImportError until models.py is updated).

- [ ] **Run test after implementing**

```bash
pytest saldivia/tests/test_gateway_extended.py -v
```

Expected: PASS

- [ ] **Commit**

```bash
git add pyproject.toml saldivia/auth/models.py saldivia/auth/database.py saldivia/tests/test_gateway_extended.py
git commit -m "feat: add password hashing + user DB helpers for JWT auth"
```

---

### Task 2: JWT auth endpoints

**Files:**
- Modify: `saldivia/gateway.py`
- Modify: `saldivia/tests/test_gateway_extended.py`

Adds `POST /auth/session`, `DELETE /auth/session`, `GET /auth/me`, `POST /auth/refresh-key`.

The gateway issues the JWT but never validates it on incoming requests — that's the BFF's job. The gateway only validates the `SYSTEM_API_KEY` Bearer token for all requests (including the new ones).

- [ ] **Add JWT config to gateway.py**

After the existing configuration block in `gateway.py`:

```python
import jwt as pyjwt
from datetime import timedelta

JWT_SECRET = os.getenv("JWT_SECRET", "")
JWT_ALGORITHM = "HS256"
JWT_EXPIRE_HOURS = 8
```

Add a helper to create tokens:

```python
def create_jwt(user: User) -> str:
    from datetime import datetime
    payload = {
        "user_id": user.id,
        "email": user.email,
        "role": user.role.value,
        "area_id": user.area_id,
        "exp": datetime.utcnow() + timedelta(hours=JWT_EXPIRE_HOURS),
    }
    return pyjwt.encode(payload, JWT_SECRET, algorithm=JWT_ALGORITHM)
```

- [ ] **Add auth endpoints**

Add to `gateway.py`:

```python
from pydantic import BaseModel

class LoginRequest(BaseModel):
    email: str
    password: str

@app.post("/auth/session")
async def login(body: LoginRequest, user: User = Depends(get_user_from_token)):
    """Issue JWT for a valid email+password. Caller must be authenticated (BFF uses SYSTEM_API_KEY)."""
    from saldivia.auth.models import verify_password
    target = db.get_user_by_email(body.email)
    if not target or not target.password_hash:
        raise HTTPException(status_code=401, detail="Invalid credentials")
    if not verify_password(body.password, target.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    if not target.active:
        raise HTTPException(status_code=403, detail="Account disabled")
    db.update_last_login(target.id)
    token = create_jwt(target)
    return {"token": token, "user": {"id": target.id, "email": target.email,
                                      "name": target.name, "role": target.role.value,
                                      "area_id": target.area_id}}

@app.delete("/auth/session")
async def logout(user: User = Depends(get_user_from_token)):
    """Logout endpoint (stateless — BFF clears the cookie)."""
    return {"ok": True}

@app.get("/auth/me")
async def me(user_id: int, user: User = Depends(get_user_from_token)):
    """Get profile for a user_id (BFF passes user_id from JWT)."""
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    return {"id": target.id, "email": target.email, "name": target.name,
            "role": target.role.value, "area_id": target.area_id,
            "last_login": target.last_login.isoformat() if target.last_login else None}

@app.post("/auth/refresh-key")
async def refresh_my_key(user_id: int, user: User = Depends(get_user_from_token)):
    """Regenerate API key for a user (user regenerates their own, or admin for any)."""
    from saldivia.auth.models import generate_api_key
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    # Only the user themselves or an admin can refresh
    if user and user.id != user_id and user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Not allowed")
    new_key, new_hash = generate_api_key()
    db.update_api_key(user_id, new_hash)
    return {"api_key": new_key}
```

- [ ] **Write failing tests**

Append to `saldivia/tests/test_gateway_extended.py`:

```python
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key, hash_password

@pytest.fixture
def client():
    return TestClient(app)

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
```

- [ ] **Run tests**

```bash
pytest saldivia/tests/test_gateway_extended.py -v -k "login"
```

Expected: PASS for all 3 login tests

- [ ] **Commit**

```bash
git add saldivia/gateway.py saldivia/tests/test_gateway_extended.py
git commit -m "feat: add JWT auth endpoints (session, me, refresh-key)"
```

---

### Task 3: Admin user management endpoints

**Files:**
- Modify: `saldivia/gateway.py`
- Modify: `saldivia/tests/test_gateway_extended.py`

Adds CRUD for users: `GET /admin/users`, `POST /admin/users`, `PUT /admin/users/{id}`, `DELETE /admin/users/{id}`, `POST /admin/users/{id}/reset-key`.

- [ ] **Add helper: admin_required**

Add to `gateway.py` after `get_user_from_token`:

```python
def admin_required(user: User = Depends(get_user_from_token)) -> User:
    """Require ADMIN role."""
    if user is None and not BYPASS_AUTH:
        raise HTTPException(status_code=401, detail="Auth required")
    if user and user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin only")
    return user

def admin_or_manager_required(user: User = Depends(get_user_from_token)) -> User:
    """Require ADMIN or AREA_MANAGER role."""
    if user is None and not BYPASS_AUTH:
        raise HTTPException(status_code=401, detail="Auth required")
    if user and user.role == Role.USER:
        raise HTTPException(status_code=403, detail="Insufficient permissions")
    return user
```

- [ ] **Add user CRUD endpoints**

```python
class CreateUserRequest(BaseModel):
    email: str
    name: str
    area_id: int
    role: str = "user"
    password: Optional[str] = None

class UpdateUserRequest(BaseModel):
    name: Optional[str] = None
    area_id: Optional[int] = None
    role: Optional[str] = None
    active: Optional[bool] = None

@app.get("/admin/users")
async def list_users_endpoint(user: User = Depends(admin_required)):
    users = db.list_users()
    return {"users": [{"id": u.id, "email": u.email, "name": u.name,
                        "area_id": u.area_id, "role": u.role.value,
                        "active": u.active,
                        "last_login": u.last_login.isoformat() if u.last_login else None}
                       for u in users]}

@app.post("/admin/users", status_code=201)
async def create_user_endpoint(body: CreateUserRequest, user: User = Depends(admin_required)):
    from saldivia.auth.models import generate_api_key, hash_password, Role as RoleEnum
    new_key, new_hash = generate_api_key()
    pw_hash = hash_password(body.password) if body.password else None
    try:
        new_user = db.create_user(
            email=body.email, name=body.name, area_id=body.area_id,
            role=RoleEnum(body.role), api_key_hash=new_hash, password_hash=pw_hash
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
    return {"id": new_user.id, "email": new_user.email, "api_key": new_key}

@app.put("/admin/users/{user_id}")
async def update_user_endpoint(user_id: int, body: UpdateUserRequest,
                                user: User = Depends(admin_required)):
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    updates = {k: v for k, v in body.dict().items() if v is not None}
    if "role" in updates:
        from saldivia.auth.models import Role as RoleEnum
        updates["role"] = RoleEnum(updates["role"])
    db.update_user(user_id, **updates)
    return {"ok": True}

@app.delete("/admin/users/{user_id}")
async def delete_user_endpoint(user_id: int, user: User = Depends(admin_required)):
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    db.deactivate_user(user_id)
    return {"ok": True}

@app.post("/admin/users/{user_id}/reset-key")
async def reset_user_key(user_id: int, user: User = Depends(admin_required)):
    from saldivia.auth.models import generate_api_key
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    new_key, new_hash = generate_api_key()
    db.update_api_key(user_id, new_hash)
    return {"api_key": new_key}
```

- [ ] **Write failing tests**

```python
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
        new_u = User(id=99, email="new@test.com", name="New", area_id=1,
                     role=Role.USER, api_key_hash="hash")
        mock_db.create_user.return_value = new_u
        resp = client.post("/admin/users",
                           json={"email": "new@test.com", "name": "New",
                                 "area_id": 1, "role": "user", "password": "pass123"},
                           headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 201
    assert "api_key" in resp.json()
```

- [ ] **Run tests**

```bash
pytest saldivia/tests/test_gateway_extended.py -v -k "user"
```

Expected: PASS

- [ ] **Commit**

```bash
git add saldivia/gateway.py saldivia/tests/test_gateway_extended.py
git commit -m "feat: add admin user CRUD endpoints"
```

---

### Task 4: Admin areas + extended audit + collection stats

**Files:**
- Modify: `saldivia/gateway.py`
- Modify: `saldivia/auth/database.py`
- Modify: `saldivia/tests/test_gateway_extended.py`

Adds areas CRUD, collection permission grant/revoke, `GET /v1/collections/{name}/stats`, and extended audit filters.

- [ ] **Add area delete + update to DB**

In `database.py`, add:

```python
def update_area(self, area_id: int, name: str = None, description: str = None):
    updates = {}
    if name is not None:
        updates["name"] = name
    if description is not None:
        updates["description"] = description
    if not updates:
        return
    set_clause = ", ".join(f"{k} = ?" for k in updates)
    with self._conn() as conn:
        conn.execute(f"UPDATE areas SET {set_clause} WHERE id = ?",
                     list(updates.values()) + [area_id])

def delete_area(self, area_id: int):
    """Delete area. Raises ValueError if area has active users."""
    with self._conn() as conn:
        count = conn.execute(
            "SELECT COUNT(*) FROM users WHERE area_id = ? AND active = 1", (area_id,)
        ).fetchone()[0]
        if count > 0:
            raise ValueError(f"Area has {count} active users")
        conn.execute("DELETE FROM area_collections WHERE area_id = ?", (area_id,))
        conn.execute("DELETE FROM areas WHERE id = ?", (area_id,))

def get_audit_log_filtered(self, user_id: int = None, action: str = None,
                            collection: str = None, from_ts: str = None,
                            to_ts: str = None, limit: int = 100) -> list[AuditEntry]:
    """Audit log with optional filters."""
    conditions = []
    params = []
    if user_id:
        conditions.append("user_id = ?")
        params.append(user_id)
    if action:
        conditions.append("action = ?")
        params.append(action)
    if collection:
        conditions.append("collection = ?")
        params.append(collection)
    if from_ts:
        conditions.append("timestamp >= ?")
        params.append(from_ts)
    if to_ts:
        conditions.append("timestamp <= ?")
        params.append(to_ts)
    where = f"WHERE {' AND '.join(conditions)}" if conditions else ""
    params.append(limit)
    with self._conn() as conn:
        rows = conn.execute(
            f"SELECT id, user_id, action, collection, query_preview, ip_address, timestamp "
            f"FROM audit_log {where} ORDER BY timestamp DESC LIMIT ?", params
        ).fetchall()
    return [AuditEntry(id=r[0], user_id=r[1], action=r[2], collection=r[3],
                       query_preview=r[4], ip_address=r[5], timestamp=r[6]) for r in rows]
```

- [ ] **Add area endpoints**

```python
class CreateAreaRequest(BaseModel):
    name: str
    description: str = ""

class UpdateAreaRequest(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None

class GrantCollectionRequest(BaseModel):
    collection_name: str
    permission: str = "read"

@app.get("/admin/areas")
async def list_areas_endpoint(user: User = Depends(admin_or_manager_required)):
    areas = db.list_areas()
    return {"areas": [{"id": a.id, "name": a.name, "description": a.description} for a in areas]}

@app.post("/admin/areas", status_code=201)
async def create_area_endpoint(body: CreateAreaRequest, user: User = Depends(admin_required)):
    area = db.create_area(body.name, body.description)
    return {"id": area.id, "name": area.name}

@app.put("/admin/areas/{area_id}")
async def update_area_endpoint(area_id: int, body: UpdateAreaRequest,
                                user: User = Depends(admin_required)):
    db.update_area(area_id, name=body.name, description=body.description)
    return {"ok": True}

@app.delete("/admin/areas/{area_id}")
async def delete_area_endpoint(area_id: int, user: User = Depends(admin_required)):
    try:
        db.delete_area(area_id)
    except ValueError as e:
        raise HTTPException(status_code=409, detail=str(e))
    return {"ok": True}

@app.get("/admin/areas/{area_id}/collections")
async def get_area_collections_endpoint(area_id: int, user: User = Depends(admin_or_manager_required)):
    # AREA_MANAGER can only see their own area
    if user and user.role == Role.AREA_MANAGER and user.area_id != area_id:
        raise HTTPException(status_code=403, detail="Can only view your own area")
    collections = db.get_area_collections(area_id)
    return {"collections": [{"name": c.collection_name, "permission": c.permission.value}
                              for c in collections]}

@app.post("/admin/areas/{area_id}/collections")
async def grant_collection_endpoint(area_id: int, body: GrantCollectionRequest,
                                     user: User = Depends(admin_or_manager_required)):
    if user and user.role == Role.AREA_MANAGER and user.area_id != area_id:
        raise HTTPException(status_code=403, detail="Can only modify your own area")
    from saldivia.auth.models import Permission as PermEnum
    db.grant_collection_access(area_id, body.collection_name, PermEnum(body.permission))
    return {"ok": True}

@app.delete("/admin/areas/{area_id}/collections/{collection_name}")
async def revoke_collection_endpoint(area_id: int, collection_name: str,
                                      user: User = Depends(admin_or_manager_required)):
    if user and user.role == Role.AREA_MANAGER and user.area_id != area_id:
        raise HTTPException(status_code=403, detail="Can only modify your own area")
    db.revoke_collection_access(area_id, collection_name)
    return {"ok": True}
```

- [ ] **Add collection stats + extended audit**

```python
@app.get("/v1/collections/{collection_name}/stats")
async def collection_stats(collection_name: str, user: User = Depends(get_user_from_token)):
    """Stats for a specific collection (entity count, doc count, last ingest)."""
    # Check access
    if user and user.role != Role.ADMIN:
        if not db.can_access(user, collection_name, Permission.READ):
            raise HTTPException(status_code=403, detail="No access to collection")
    try:
        from saldivia.collections import CollectionManager
        stats = CollectionManager().stats(collection_name)
        return stats
    except Exception as e:
        raise HTTPException(status_code=404, detail=str(e))

@app.get("/admin/audit")
async def get_audit(
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    collection: Optional[str] = None,
    from_ts: Optional[str] = None,
    to_ts: Optional[str] = None,
    limit: int = 100,
    user: User = Depends(get_user_from_token)
):
    """Get audit log with optional filters (admin only)."""
    if user and user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin only")
    entries = db.get_audit_log_filtered(
        user_id=user_id, action=action, collection=collection,
        from_ts=from_ts, to_ts=to_ts, limit=limit
    )
    return {"entries": [{"id": e.id, "user_id": e.user_id, "action": e.action,
                          "collection": e.collection, "query_preview": e.query_preview,
                          "ip_address": e.ip_address,
                          "timestamp": e.timestamp.isoformat() if e.timestamp else None}
                         for e in entries]}
```

Note: the existing `get_audit` function at the bottom of `gateway.py` needs to be **replaced** (not added alongside) to avoid route conflicts.

- [ ] **Write failing tests**

```python
def test_list_areas(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        from saldivia.auth.models import Area
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.list_areas.return_value = [Area(id=1, name="Mantenimiento")]
        resp = client.get("/admin/areas", headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    assert resp.json()["areas"][0]["name"] == "Mantenimiento"

def test_area_manager_cannot_see_other_area(client):
    key, hash_val = generate_api_key()
    manager = User(id=2, email="mgr@test.com", name="Mgr", area_id=1,
                   role=Role.AREA_MANAGER, api_key_hash=hash_val)
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = manager
        resp = client.get("/admin/areas/99/collections",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 403

def test_audit_filters(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.get_audit_log_filtered.return_value = []
        resp = client.get("/admin/audit?action=query&limit=10",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 200
    mock_db.get_audit_log_filtered.assert_called_once_with(
        user_id=None, action="query", collection=None,
        from_ts=None, to_ts=None, limit=10
    )
```

- [ ] **Run tests**

```bash
pytest saldivia/tests/test_gateway_extended.py -v -k "area or audit"
```

Expected: PASS

- [ ] **Commit**

```bash
git add saldivia/auth/database.py saldivia/gateway.py saldivia/tests/test_gateway_extended.py
git commit -m "feat: add areas CRUD, collection perms, stats, extended audit"
```

---

### Task 5: Chat sessions

**Files:**
- Modify: `saldivia/auth/models.py`
- Modify: `saldivia/auth/database.py`
- Modify: `saldivia/gateway.py`
- Modify: `saldivia/tests/test_gateway_extended.py`

New `chat_sessions` + `chat_messages` tables. Endpoints: `GET/POST /chat/sessions`, `GET/DELETE /chat/sessions/{id}`.

- [ ] **Add models**

In `saldivia/auth/models.py`:

```python
@dataclass
class ChatMessage:
    role: str        # "user" | "assistant"
    content: str
    sources: Optional[list] = None   # list of source dicts from RAG
    timestamp: datetime = field(default_factory=datetime.now)

@dataclass
class ChatSession:
    id: str          # UUID
    user_id: int
    title: str       # First 60 chars of first user message
    collection: str
    crossdoc: bool = False
    messages: list = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)
```

- [ ] **Add schema migration**

In `init_db()`, after the existing tables (inside the executescript, before `conn.close()`):

```sql
CREATE TABLE IF NOT EXISTS chat_sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    collection TEXT NOT NULL,
    crossdoc INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES chat_sessions(id),
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    sources TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_user ON chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id);
```

- [ ] **Add DB methods**

```python
def list_chat_sessions(self, user_id: int, limit: int = 50) -> list[ChatSession]:
    with self._conn() as conn:
        rows = conn.execute(
            "SELECT id, user_id, title, collection, crossdoc, created_at, updated_at "
            "FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC LIMIT ?",
            (user_id, limit)
        ).fetchall()
    return [ChatSession(id=r[0], user_id=r[1], title=r[2], collection=r[3],
                        crossdoc=bool(r[4]), created_at=r[5], updated_at=r[6])
            for r in rows]

def get_chat_session(self, session_id: str, user_id: int) -> Optional[ChatSession]:
    from saldivia.auth.models import ChatMessage
    import json
    with self._conn() as conn:
        row = conn.execute(
            "SELECT id, user_id, title, collection, crossdoc, created_at, updated_at "
            "FROM chat_sessions WHERE id = ? AND user_id = ?",
            (session_id, user_id)
        ).fetchone()
        if not row:
            return None
        msg_rows = conn.execute(
            "SELECT role, content, sources, timestamp FROM chat_messages "
            "WHERE session_id = ? ORDER BY timestamp",
            (session_id,)
        ).fetchall()
    messages = [ChatMessage(role=m[0], content=m[1],
                             sources=json.loads(m[2]) if m[2] else None, timestamp=m[3])
                for m in msg_rows]
    return ChatSession(id=row[0], user_id=row[1], title=row[2], collection=row[3],
                       crossdoc=bool(row[4]), created_at=row[5], updated_at=row[6],
                       messages=messages)

def create_chat_session(self, user_id: int, collection: str,
                         crossdoc: bool = False) -> ChatSession:
    import uuid
    from saldivia.auth.models import ChatSession
    session_id = str(uuid.uuid4())
    with self._conn() as conn:
        conn.execute(
            "INSERT INTO chat_sessions (id, user_id, title, collection, crossdoc) "
            "VALUES (?, ?, ?, ?, ?)",
            (session_id, user_id, "Nueva consulta", collection, int(crossdoc))
        )
    return ChatSession(id=session_id, user_id=user_id, title="Nueva consulta",
                       collection=collection, crossdoc=crossdoc)

def add_chat_message(self, session_id: str, role: str, content: str, sources=None):
    import json
    from datetime import datetime
    sources_json = json.dumps(sources) if sources else None
    with self._conn() as conn:
        conn.execute(
            "INSERT INTO chat_messages (session_id, role, content, sources) VALUES (?, ?, ?, ?)",
            (session_id, role, content, sources_json)
        )
        conn.execute(
            "UPDATE chat_sessions SET updated_at = ?, title = CASE "
            "WHEN title = 'Nueva consulta' AND ? = 'user' THEN SUBSTR(?, 1, 60) "
            "ELSE title END WHERE id = ?",
            (datetime.now().isoformat(), role, content, session_id)
        )

def delete_chat_session(self, session_id: str, user_id: int):
    with self._conn() as conn:
        conn.execute("DELETE FROM chat_messages WHERE session_id = ?", (session_id,))
        conn.execute("DELETE FROM chat_sessions WHERE id = ? AND user_id = ?",
                     (session_id, user_id))
```

- [ ] **Add chat session endpoints**

```python
class CreateSessionRequest(BaseModel):
    collection: str
    crossdoc: bool = False

@app.get("/chat/sessions")
async def list_sessions(user_id: int, limit: int = 50,
                         user: User = Depends(get_user_from_token)):
    sessions = db.list_chat_sessions(user_id=user_id, limit=limit)
    return {"sessions": [{"id": s.id, "title": s.title, "collection": s.collection,
                           "crossdoc": s.crossdoc,
                           "updated_at": s.updated_at.isoformat() if hasattr(s.updated_at, 'isoformat') else s.updated_at}
                          for s in sessions]}

@app.post("/chat/sessions", status_code=201)
async def create_session(body: CreateSessionRequest, user_id: int,
                          user: User = Depends(get_user_from_token)):
    session = db.create_chat_session(user_id=user_id, collection=body.collection,
                                     crossdoc=body.crossdoc)
    return {"id": session.id, "title": session.title, "collection": session.collection}

@app.get("/chat/sessions/{session_id}")
async def get_session(session_id: str, user_id: int,
                       user: User = Depends(get_user_from_token)):
    session = db.get_chat_session(session_id=session_id, user_id=user_id)
    if not session:
        raise HTTPException(status_code=404, detail="Session not found")
    return {"id": session.id, "title": session.title, "collection": session.collection,
            "crossdoc": session.crossdoc,
            "messages": [{"role": m.role, "content": m.content, "sources": m.sources,
                           "timestamp": m.timestamp.isoformat() if hasattr(m.timestamp, 'isoformat') else m.timestamp}
                          for m in session.messages]}

@app.delete("/chat/sessions/{session_id}")
async def delete_session(session_id: str, user_id: int,
                          user: User = Depends(get_user_from_token)):
    db.delete_chat_session(session_id=session_id, user_id=user_id)
    return {"ok": True}
```

- [ ] **Write failing tests**

```python
def test_create_session(client, admin_user):
    with patch("saldivia.gateway.db") as mock_db:
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
    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = admin_user
        mock_db.get_chat_session.return_value = None
        resp = client.get("/chat/sessions/nonexistent?user_id=1",
                          headers={"Authorization": "Bearer rsk_dummy"})
    assert resp.status_code == 404
```

- [ ] **Run all tests**

```bash
pytest saldivia/tests/test_gateway_extended.py -v
```

Expected: all PASS

- [ ] **Run existing tests to confirm no regressions**

```bash
pytest saldivia/tests/ -v
```

Expected: all PASS (including pre-existing gateway and auth tests)

- [ ] **Commit**

```bash
git add saldivia/auth/models.py saldivia/auth/database.py saldivia/gateway.py saldivia/tests/test_gateway_extended.py
git commit -m "feat: add chat sessions endpoints + SQLite schema"
```

---

### Task 6: Integration smoke test + deploy

**Files:**
- Modify: `config/compose-platform-services.yaml`
- Modify: `config/profiles/brev-2gpu.yaml`
- Modify: `config/profiles/workstation-1gpu.yaml`

Add `JWT_SECRET` and `SYSTEM_API_KEY` env vars to the gateway service, then deploy and smoke test.

- [ ] **Add env vars to gateway compose**

In `config/compose-platform-services.yaml`, update the `auth-gateway` service environment:

```yaml
  auth-gateway:
    # ...existing config...
    environment:
      - RAG_SERVER_URL=http://rag-server:8081
      - INGESTOR_URL=http://ingestor-server:8082
      - MILVUS_HOST=milvus-standalone
      - BYPASS_AUTH=${BYPASS_AUTH:-false}
      - ENVIRONMENT=${ENVIRONMENT:-production}
      - JWT_SECRET=${JWT_SECRET}        # ADD
      - SYSTEM_API_KEY=${SYSTEM_API_KEY}  # ADD (for gateway to validate its own caller)
```

- [ ] **Add env vars to profiles**

In `config/.env.saldivia`, add (generate real values with `python3 -c "import secrets; print(secrets.token_hex(32))"`):

```bash
JWT_SECRET=<generated>
SYSTEM_API_KEY=<use existing admin API key from data/auth.db>
```

- [ ] **Deploy to Brev and smoke test**

```bash
# On Brev server
ssh nvidia-enterprise-rag-deb106
cd ~/rag-saldivia && git pull && make deploy PROFILE=brev-2gpu

# Smoke test: health
curl http://localhost:9000/health

# Smoke test: list users (with admin API key)
curl -H "Authorization: Bearer <SYSTEM_API_KEY>" http://localhost:9000/admin/users

# Smoke test: login (must have a user with password set first)
curl -X POST http://localhost:9000/auth/session \
  -H "Authorization: Bearer <SYSTEM_API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@tecpia.com","password":"changeme"}'
```

Expected: all return 200 with valid JSON.

Note: To set the first admin password, use the CLI:
```bash
cd ~/rag-saldivia && python3 -c "
from saldivia.auth.database import AuthDB
from saldivia.auth.models import hash_password
db = AuthDB()
db.set_password(1, hash_password('changeme'))
print('Password set')
"
```

- [ ] **Commit**

```bash
git add config/compose-platform-services.yaml config/.env.saldivia
git commit -m "chore: add JWT_SECRET + SYSTEM_API_KEY to gateway compose config"
```

---

**Gateway extensions complete.** All 20 new endpoints are live. Proceed to Plan B: SDA Frontend.
