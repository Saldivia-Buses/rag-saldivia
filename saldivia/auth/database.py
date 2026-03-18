# saldivia/auth/database.py
"""SQLite database for auth."""
import sqlite3
from pathlib import Path
from typing import Optional
from saldivia.auth.models import User, Area, AreaCollection, AuditEntry, Role, Permission

DB_PATH = Path("data/auth.db")


def init_db(db_path: Path = DB_PATH):
    """Initialize database schema."""
    db_path.parent.mkdir(parents=True, exist_ok=True)

    conn = sqlite3.connect(db_path)
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

    conn.close()


class AuthDB:
    """Synchronous auth database operations."""

    def __init__(self, db_path: Path = DB_PATH):
        self.db_path = db_path
        init_db(db_path)

    def _conn(self):
        return sqlite3.connect(self.db_path, detect_types=sqlite3.PARSE_DECLTYPES)

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

    def set_password(self, user_id: int, password_hash: str):
        with self._conn() as conn:
            conn.execute("UPDATE users SET password_hash = ? WHERE id = ?", (password_hash, user_id))

    def update_last_login(self, user_id: int):
        from datetime import datetime
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

    def get_audit_log(self, user_id: int = None, limit: int = 100) -> list[AuditEntry]:
        with self._conn() as conn:
            if user_id:
                rows = conn.execute(
                    "SELECT id, user_id, action, collection, query_preview, ip_address, timestamp "
                    "FROM audit_log WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?",
                    (user_id, limit)
                ).fetchall()
            else:
                rows = conn.execute(
                    "SELECT id, user_id, action, collection, query_preview, ip_address, timestamp "
                    "FROM audit_log ORDER BY timestamp DESC LIMIT ?",
                    (limit,)
                ).fetchall()
            return [AuditEntry(
                id=r[0], user_id=r[1], action=r[2], collection=r[3],
                query_preview=r[4], ip_address=r[5], timestamp=r[6]
            ) for r in rows]

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
