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
    def create_user(self, email: str, name: str, area_id: int, role: Role, api_key_hash: str) -> User:
        with self._conn() as conn:
            cur = conn.execute(
                "INSERT INTO users (email, name, area_id, role, api_key_hash) VALUES (?, ?, ?, ?, ?)",
                (email, name, area_id, role.value, api_key_hash)
            )
            return User(
                id=cur.lastrowid, email=email, name=name,
                area_id=area_id, role=role, api_key_hash=api_key_hash
            )

    def get_user_by_api_key_hash(self, api_key_hash: str) -> Optional[User]:
        with self._conn() as conn:
            row = conn.execute(
                "SELECT id, email, name, area_id, role, api_key_hash, created_at, last_login, active "
                "FROM users WHERE api_key_hash = ? AND active = 1",
                (api_key_hash,)
            ).fetchone()
            if row:
                return User(
                    id=row[0], email=row[1], name=row[2], area_id=row[3],
                    role=Role(row[4]), api_key_hash=row[5], created_at=row[6],
                    last_login=row[7], active=row[8]
                )
            return None

    def list_users(self, area_id: int = None) -> list[User]:
        with self._conn() as conn:
            if area_id is not None:
                rows = conn.execute(
                    "SELECT id, email, name, area_id, role, api_key_hash, created_at, last_login, active "
                    "FROM users WHERE area_id = ?", (area_id,)
                ).fetchall()
            else:
                rows = conn.execute(
                    "SELECT id, email, name, area_id, role, api_key_hash, created_at, last_login, active "
                    "FROM users"
                ).fetchall()
            return [User(
                id=r[0], email=r[1], name=r[2], area_id=r[3],
                role=Role(r[4]), api_key_hash=r[5], created_at=r[6],
                last_login=r[7], active=r[8]
            ) for r in rows]

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
