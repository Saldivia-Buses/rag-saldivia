# cli/

Command-line interface for RAG Saldivia — collection management, user management, ingestion, and platform status.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `main.py` | CLI entry point: registers all command groups (collections, users, areas, audit, ingest) + defines status and mcp commands | click, saldivia/ |
| `areas.py` | Area management commands: create, list, delete areas | click, saldivia/auth/database.py |
| `audit.py` | Audit log commands: view audit entries, filter by user/collection/action | click, saldivia/auth/database.py |
| `collections.py` | Collection management commands: create, list, delete, stats, grant/revoke area access | click, saldivia/collections.py |
| `ingest.py` | Ingestion commands: add documents to queue, view queue status | click, saldivia/ingestion_queue.py |
| `users.py` | User management commands (admin only): create, list, delete, update users | click, saldivia/auth/database.py |

## Usage Examples

### Collections

```bash
# List all collections
rag-saldivia collections list

# Create a new collection
rag-saldivia collections create my-collection

# Delete a collection
rag-saldivia collections delete my-collection

# Grant area access to a collection
rag-saldivia collections grant my-collection --area 1 --permission write
```

### Users (admin only)

```bash
# List all users
rag-saldivia users list

# Create a new user
rag-saldivia users create username --email user@example.com --role admin --area 1

# Delete a user
rag-saldivia users delete username
```

### Ingestion

```bash
# Add documents to ingestion queue
rag-saldivia ingest add /path/to/docs/ --collection my-collection

# View ingestion queue
rag-saldivia ingest queue

# View ingestion job status
rag-saldivia ingest status <job-id>
```

### Audit

```bash
# View recent audit log entries
rag-saldivia audit log --limit 50

# Filter by user
rag-saldivia audit log --user-id 1

# Filter by collection
rag-saldivia audit log --collection my-collection
```

### Platform Status

```bash
# Show platform status (services, collections, mode)
rag-saldivia status
```

The CLI is installed as `rag-saldivia` when the package is installed with `pip install -e .` or `uv pip install -e .`.
