# saldivia/auth/

Authentication and authorization subsystem for RAG Saldivia — SQLite-backed user/area/RBAC management.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `database.py` | AuthDB class: SQLite database for users, areas, area-collection permissions, audit log | sqlite3, models.py |
| `models.py` | Auth dataclasses (User, Area, AreaCollection, AuditEntry), Role/Permission enums, API key generation, bcrypt password hashing | bcrypt, hashlib |

