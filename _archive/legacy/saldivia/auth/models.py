# saldivia/auth/models.py
"""Auth models for RAG Saldivia."""
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Optional
import secrets
import hashlib


class Role(str, Enum):
    ADMIN = "admin"           # Full access to all areas and collections
    AREA_MANAGER = "area_manager"  # Manage users and collections in their area
    USER = "user"             # Query collections assigned to their area


class Permission(str, Enum):
    READ = "read"             # Can query collection
    WRITE = "write"           # Can ingest to collection
    ADMIN = "admin"           # Can delete from collection


@dataclass
class Area:
    id: int
    name: str                 # e.g., "Mantenimiento", "Producción"
    description: str = ""
    created_at: datetime = field(default_factory=datetime.now)


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


@dataclass
class AreaCollection:
    """Permission for an area to access a collection."""
    area_id: int
    collection_name: str
    permission: Permission


@dataclass
class AuditEntry:
    id: int
    user_id: int
    action: str               # query, ingest, create_collection, delete, etc.
    collection: Optional[str]
    query_preview: Optional[str]  # First 100 chars of query
    ip_address: str
    timestamp: datetime = field(default_factory=datetime.now)


def generate_api_key() -> tuple[str, str]:
    """Generate API key and its hash. Returns (key, hash)."""
    key = f"rsk_{secrets.token_urlsafe(32)}"
    hash_val = hashlib.sha256(key.encode()).hexdigest()
    return key, hash_val


def verify_api_key(key: str, hash_val: str) -> bool:
    """Verify an API key against its hash."""
    return hashlib.sha256(key.encode()).hexdigest() == hash_val


def hash_password(password: str) -> str:
    """Hash a password with bcrypt."""
    import bcrypt
    return bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()


def verify_password(password: str, password_hash: str) -> bool:
    """Verify a password against its bcrypt hash."""
    import bcrypt
    return bcrypt.checkpw(password.encode(), password_hash.encode())


@dataclass
class ChatMessage:
    role: str
    content: str
    sources: Optional[list] = None   # list of source dicts from RAG
    timestamp: datetime = field(default_factory=datetime.now)
    id: Optional[int] = None         # DB primary key — al final para no forzar defaults en role/content


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
