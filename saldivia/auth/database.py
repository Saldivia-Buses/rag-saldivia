# saldivia/auth/database.py
"""SQLite database for auth."""
import sqlite3
import uuid
from datetime import datetime
from pathlib import Path
from typing import Optional
from saldivia.auth.models import User, Area, AreaCollection, AuditEntry, Role, Permission

DB_PATH = Path("data/auth.db")

# Cacheado al importar el módulo — la versión no cambia en runtime
try:
    from importlib.metadata import version as _pkg_version
    _SALDIVIA_VERSION = _pkg_version("saldivia")
except Exception:
    _SALDIVIA_VERSION = "unknown"

# Estados de jobs para el stall checker (solo activos, sin stalled)
_STALL_CHECK_STATES = ("pending", "running")
# Estados visibles como "activos" en la UI del usuario (incluye stalled)
_ACTIVE_UI_STATES = ("pending", "running", "stalled")


def init_db_conn(conn: sqlite3.Connection):
    """Run all DDL migrations on an existing connection (supports :memory: and file DBs)."""
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS areas (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            description TEXT DEFAULT '',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT UNIQUE NOT NULL,
            name TEXT NOT NULL,
            area_id INTEGER NOT NULL REFERENCES areas(id),
            role TEXT NOT NULL DEFAULT 'user',
            api_key_hash TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_login TIMESTAMP,
            active BOOLEAN DEFAULT 1
        );

        CREATE TABLE IF NOT EXISTS area_collections (
            area_id INTEGER NOT NULL REFERENCES areas(id),
            collection_name TEXT NOT NULL,
            permission TEXT NOT NULL DEFAULT 'read',
            PRIMARY KEY (area_id, collection_name)
        );

        CREATE TABLE IF NOT EXISTS audit_log (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL REFERENCES users(id),
            action TEXT NOT NULL,
            collection TEXT,
            query_preview TEXT,
            ip_address TEXT,
            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id);
        CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);
        CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key_hash);
    """)

    # Migrate: add password_hash column if it doesn't exist
    try:
        conn.execute("ALTER TABLE users ADD COLUMN password_hash TEXT")
    except Exception:
        pass  # Column already exists

    # Migration: add chat tables if not present (idempotent)
    conn.executescript("""
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
        CREATE TABLE IF NOT EXISTS message_feedback (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            message_id  INTEGER NOT NULL,
            user_id     INTEGER NOT NULL,
            rating      TEXT    NOT NULL CHECK(rating IN ('up', 'down')),
            created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
            UNIQUE(message_id, user_id)
        );
        CREATE INDEX IF NOT EXISTS idx_chat_sessions_user ON chat_sessions(user_id);
        CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_updated ON chat_sessions(user_id, updated_at);
        CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id);
        CREATE INDEX IF NOT EXISTS idx_message_feedback_message ON message_feedback(message_id);
    """)

    # Migration: add ingestion_jobs table if not present (idempotent)
    conn.executescript("""
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
        );
        CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_user ON ingestion_jobs(user_id);
        CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_state ON ingestion_jobs(state);
    """)

    # Migration: add Fase 6 columns to ingestion_jobs if missing
    for col, definition in [
        ("file_hash",    "TEXT"),
        ("retry_count",  "INTEGER DEFAULT 0"),
        ("last_checked", "TEXT"),
    ]:
        try:
            conn.execute(f"ALTER TABLE ingestion_jobs ADD COLUMN {col} {definition}")
        except Exception:
            pass  # columna ya existe

    # Migration: add ingestion_alerts table (Fase 6)
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS ingestion_alerts (
            id                   TEXT PRIMARY KEY,
            job_id               TEXT NOT NULL,
            user_id              INTEGER NOT NULL,
            filename             TEXT NOT NULL,
            collection           TEXT NOT NULL,
            tier                 TEXT NOT NULL,
            page_count           INTEGER,
            file_hash            TEXT,
            error                TEXT,
            retry_count          INTEGER,
            progress_at_failure  INTEGER,
            gateway_version      TEXT,
            created_at           TEXT NOT NULL,
            resolved_at          TEXT,
            resolved_by          TEXT,
            notes                TEXT
        );
        CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON ingestion_alerts(resolved_at);
    """)


def init_db(db_path: Path = DB_PATH):
    """Initialize database schema."""
    if str(db_path) != ":memory:":
        Path(db_path).parent.mkdir(parents=True, exist_ok=True)

    conn = sqlite3.connect(db_path)
    init_db_conn(conn)
    conn.close()


class _MemConnContext:
    """Context manager wrapping a persistent in-memory sqlite3 connection.

    Behaves like the context manager returned by ``sqlite3.connect()``:
    commits on __exit__ without error, rolls back on exception, but never
    closes the underlying connection so that the in-memory database persists
    for the lifetime of the ``AuthDB`` instance.
    """

    def __init__(self, conn: sqlite3.Connection):
        self._conn = conn

    def __enter__(self) -> sqlite3.Connection:
        return self._conn

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type is None:
            self._conn.commit()
        else:
            self._conn.rollback()
        return False  # don't suppress exceptions

    # Proxy attribute access so callers can use the context object directly
    def execute(self, *args, **kwargs):
        return self._conn.execute(*args, **kwargs)

    def executescript(self, *args, **kwargs):
        return self._conn.executescript(*args, **kwargs)


class AuthDB:
    """Synchronous auth database operations."""

    def __init__(self, db_path: Path = DB_PATH):
        self.db_path = db_path
        if str(db_path) == ":memory:":
            # Maintain a single persistent connection for in-memory databases
            self._mem_conn = sqlite3.connect(":memory:", check_same_thread=False)
            self._mem_conn.isolation_level = None  # autocommit off; managed by context manager
            init_db_conn(self._mem_conn)
        else:
            self._mem_conn = None
            init_db(db_path)

    def _conn(self):
        if self._mem_conn is not None:
            return _MemConnContext(self._mem_conn)
        return sqlite3.connect(self.db_path)

    # Areas
    def create_area(self, name: str, description: str = "") -> Area:
        with self._conn() as conn:
            cur = conn.execute(
                "INSERT INTO areas (name, description) VALUES (?, ?)",
                (name, description)
            )
            return Area(id=cur.lastrowid, name=name, description=description)

    def get_area(self, area_id: int) -> Optional[Area]:
        with self._conn() as conn:
            row = conn.execute(
                "SELECT id, name, description, created_at FROM areas WHERE id = ?",
                (area_id,)
            ).fetchone()
            if row:
                return Area(id=row[0], name=row[1], description=row[2], created_at=row[3])
            return None

    def list_areas(self) -> list[Area]:
        with self._conn() as conn:
            rows = conn.execute("SELECT id, name, description, created_at FROM areas").fetchall()
            return [Area(id=r[0], name=r[1], description=r[2], created_at=r[3]) for r in rows]

    # Users
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
                            created_at=row[7], last_login=row[8], active=bool(row[9]))
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
                            created_at=row[7], last_login=row[8], active=bool(row[9]))
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

    def update_last_login(self, user_id: int):
        with self._conn() as conn:
            conn.execute("UPDATE users SET last_login = ? WHERE id = ?",
                         (datetime.now().isoformat(), user_id))

    def deactivate_user(self, user_id: int):
        with self._conn() as conn:
            conn.execute("UPDATE users SET active = 0 WHERE id = ?", (user_id,))

    # Permissions
    def grant_collection_access(self, area_id: int, collection: str, permission: Permission):
        with self._conn() as conn:
            conn.execute(
                "INSERT OR REPLACE INTO area_collections (area_id, collection_name, permission) "
                "VALUES (?, ?, ?)",
                (area_id, collection, permission.value)
            )

    def revoke_collection_access(self, area_id: int, collection: str):
        with self._conn() as conn:
            conn.execute(
                "DELETE FROM area_collections WHERE area_id = ? AND collection_name = ?",
                (area_id, collection)
            )

    def get_area_collections(self, area_id: int) -> list[AreaCollection]:
        with self._conn() as conn:
            rows = conn.execute(
                "SELECT area_id, collection_name, permission FROM area_collections WHERE area_id = ?",
                (area_id,)
            ).fetchall()
            return [AreaCollection(area_id=r[0], collection_name=r[1], permission=Permission(r[2])) for r in rows]

    def get_user_collections(self, user: User) -> list[str]:
        """Get list of collections a user can access."""
        if user.role == Role.ADMIN:
            # Admin can access all collections
            from saldivia.collections import CollectionManager
            return CollectionManager().list()

        return [ac.collection_name for ac in self.get_area_collections(user.area_id)]

    def can_access(self, user: User, collection: str, required: Permission) -> bool:
        """Check if user can perform action on collection."""
        if user.role == Role.ADMIN:
            return True

        for ac in self.get_area_collections(user.area_id):
            if ac.collection_name == collection:
                # admin > write > read
                if ac.permission == Permission.ADMIN:
                    return True
                if ac.permission == Permission.WRITE and required in (Permission.WRITE, Permission.READ):
                    return True
                if ac.permission == Permission.READ and required == Permission.READ:
                    return True
        return False

    # Audit
    def log_action(self, user_id: int, action: str, collection: str = None,
                   query_preview: str = None, ip_address: str = ""):
        with self._conn() as conn:
            conn.execute(
                "INSERT INTO audit_log (user_id, action, collection, query_preview, ip_address) "
                "VALUES (?, ?, ?, ?, ?)",
                (user_id, action, collection, query_preview[:100] if query_preview else None, ip_address)
            )

    def update_area(self, area_id: int, name: str = None, description: str = None):
        """Update area name and/or description."""
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
        return [AuditEntry(
            id=r[0], user_id=r[1], action=r[2], collection=r[3],
            query_preview=r[4], ip_address=r[5], timestamp=r[6]
        ) for r in rows]

    def list_chat_sessions(self, user_id: int, limit: int = 50) -> list:
        with self._conn() as conn:
            rows = conn.execute(
                "SELECT id, user_id, title, collection, crossdoc, created_at, updated_at "
                "FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC LIMIT ?",
                (user_id, limit)
            ).fetchall()
        from saldivia.auth.models import ChatSession
        return [ChatSession(id=r[0], user_id=r[1], title=r[2], collection=r[3],
                            crossdoc=bool(r[4]), created_at=r[5], updated_at=r[6])
                for r in rows]

    def get_chat_session(self, session_id: str, user_id: int):
        import json
        from saldivia.auth.models import ChatSession, ChatMessage
        with self._conn() as conn:
            row = conn.execute(
                "SELECT id, user_id, title, collection, crossdoc, created_at, updated_at "
                "FROM chat_sessions WHERE id = ? AND user_id = ?",
                (session_id, user_id)
            ).fetchone()
            if not row:
                return None
            msg_rows = conn.execute(
                "SELECT m.id, m.role, m.content, m.sources, m.timestamp "
                "FROM chat_messages m "
                "JOIN chat_sessions s ON m.session_id = s.id "
                "WHERE m.session_id = ? AND s.user_id = ? "
                "ORDER BY m.timestamp",
                (session_id, user_id)
            ).fetchall()
        messages = [ChatMessage(id=m[0], role=m[1], content=m[2],
                                 sources=json.loads(m[3]) if m[3] else None, timestamp=m[4])
                    for m in msg_rows]
        return ChatSession(id=row[0], user_id=row[1], title=row[2], collection=row[3],
                           crossdoc=bool(row[4]), created_at=row[5], updated_at=row[6],
                           messages=messages)

    def create_chat_session(self, user_id: int, collection: str, crossdoc: bool = False):
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
            # Verificar ownership antes de borrar hijos
            row = conn.execute(
                "SELECT 1 FROM chat_sessions WHERE id = ? AND user_id = ?",
                (session_id, user_id)
            ).fetchone()
            if not row:
                return  # No existe o no es del usuario — no hacer nada
            # Borrar hijos en orden correcto (feedback → messages → sesión)
            conn.execute(
                "DELETE FROM message_feedback WHERE message_id IN "
                "(SELECT id FROM chat_messages WHERE session_id = ?)",
                (session_id,)
            )
            conn.execute("DELETE FROM chat_messages WHERE session_id = ?", (session_id,))
            conn.execute("DELETE FROM chat_sessions WHERE id = ?", (session_id,))

    def rename_chat_session(self, session_id: str, user_id: int, title: str):
        with self._conn() as conn:
            conn.execute(
                "UPDATE chat_sessions SET title = ?, updated_at = ? "
                "WHERE id = ? AND user_id = ?",
                (title[:80], datetime.now().isoformat(), session_id, user_id)
            )

    def upsert_message_feedback(self, message_id: int, user_id: int, rating: str):
        with self._conn() as conn:
            conn.execute(
                "INSERT OR REPLACE INTO message_feedback (message_id, user_id, rating) "
                "VALUES (?, ?, ?)",
                (message_id, user_id, rating)
            )

    # Ingestion jobs
    def create_ingestion_job(
        self,
        user_id: int,
        task_id: str,
        filename: str,
        collection: str,
        tier: str,
        page_count: int | None,
        file_hash: str | None = None,
    ) -> str:
        job_id = str(uuid.uuid4())
        with self._conn() as conn:
            conn.execute(
                """INSERT INTO ingestion_jobs
                   (id, user_id, task_id, filename, collection, tier, page_count, created_at, file_hash)
                   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                (job_id, user_id, task_id, filename, collection, tier, page_count,
                 datetime.now().isoformat(), file_hash),
            )
        return job_id

    def get_ingestion_job(self, job_id: str) -> dict | None:
        with self._conn() as conn:
            cur = conn.execute(
                "SELECT * FROM ingestion_jobs WHERE id = ?", (job_id,)
            )
            row = cur.fetchone()
            if row is None:
                return None
            columns = [col[0] for col in cur.description]
            return dict(zip(columns, row))

    def get_active_ingestion_jobs(self, user_id: int) -> list[dict]:
        placeholders = ",".join("?" * len(_ACTIVE_UI_STATES))
        with self._conn() as conn:
            cur = conn.execute(
                f"SELECT * FROM ingestion_jobs WHERE user_id = ? AND state IN ({placeholders}) ORDER BY created_at DESC",
                (user_id, *_ACTIVE_UI_STATES),
            )
            rows = cur.fetchall()
            columns = [col[0] for col in cur.description]
        return [dict(zip(columns, r)) for r in rows]

    def update_ingestion_job(
        self,
        job_id: str,
        state: str,
        progress: int,
        completed_at: str | None = None,
    ) -> None:
        with self._conn() as conn:
            conn.execute(
                """UPDATE ingestion_jobs
                   SET state = ?, progress = ?, completed_at = ?
                   WHERE id = ?""",
                (state, progress, completed_at, job_id),
            )

    def check_file_hash(self, file_hash: str, collection: str) -> dict | None:
        """Busca el job más reciente con ese hash en esa colección."""
        with self._conn() as conn:
            cur = conn.execute(
                """SELECT id, filename, collection, tier, page_count, state,
                          created_at, completed_at
                   FROM ingestion_jobs
                   WHERE file_hash = ? AND collection = ?
                   ORDER BY created_at DESC LIMIT 1""",
                (file_hash, collection),
            )
            row = cur.fetchone()
            if row is None:
                return None
            cols = [c[0] for c in cur.description]
            return dict(zip(cols, row))

    def get_all_active_ingestion_jobs(self) -> list[dict]:
        """Lista todos los jobs activos (todos los usuarios) para el stall checker."""
        placeholders = ",".join("?" * len(_STALL_CHECK_STATES))
        with self._conn() as conn:
            cur = conn.execute(
                f"SELECT * FROM ingestion_jobs WHERE state IN ({placeholders}) ORDER BY created_at ASC",
                _STALL_CHECK_STATES,
            )
            cols = [c[0] for c in cur.description]
            return [dict(zip(cols, r)) for r in cur.fetchall()]

    def increment_ingestion_retry(self, job_id: str) -> None:
        """Incrementa retry_count y actualiza last_checked."""
        with self._conn() as conn:
            conn.execute(
                """UPDATE ingestion_jobs
                   SET retry_count = retry_count + 1, last_checked = ?
                   WHERE id = ?""",
                (datetime.now().isoformat(), job_id),
            )

    # Ingestion alerts
    def create_ingestion_alert(
        self,
        *,
        job_id: str,
        user_id: int,
        filename: str,
        collection: str,
        tier: str,
        page_count: int | None,
        file_hash: str | None,
        error: str | None,
        retry_count: int,
        progress_at_failure: int,
    ) -> str:
        alert_id = str(uuid.uuid4())
        with self._conn() as conn:
            conn.execute(
                """INSERT INTO ingestion_alerts
                   (id, job_id, user_id, filename, collection, tier, page_count,
                    file_hash, error, retry_count, progress_at_failure,
                    gateway_version, created_at)
                   VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)""",
                (alert_id, job_id, user_id, filename, collection, tier,
                 page_count, file_hash, error, retry_count, progress_at_failure,
                 _SALDIVIA_VERSION, datetime.now().isoformat()),
            )
        return alert_id

    def list_ingestion_alerts(self, resolved: bool | None = None) -> list[dict]:
        where = "" if resolved is None else (
            "WHERE resolved_at IS NOT NULL" if resolved else "WHERE resolved_at IS NULL"
        )
        with self._conn() as conn:
            cur = conn.execute(
                f"SELECT * FROM ingestion_alerts {where} ORDER BY created_at DESC"
            )
            cols = [c[0] for c in cur.description]
            return [dict(zip(cols, r)) for r in cur.fetchall()]

    def resolve_ingestion_alert(
        self, alert_id: str, resolved_by: str, notes: str | None = None
    ) -> None:
        with self._conn() as conn:
            conn.execute(
                """UPDATE ingestion_alerts
                   SET resolved_at = ?, resolved_by = ?, notes = ?
                   WHERE id = ?""",
                (datetime.now().isoformat(), resolved_by, notes, alert_id),
            )
