# scripts/

Utility scripts for deployment, ingestion, testing, and operations.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `smart_ingest.py` | Adaptive PDF ingestion with tier system, deadlock detection, resume support | requests, subprocess |
| `crossdoc_client.py` | Cross-document RAG query tool: decompose complex queries into sub-queries, retrieve in parallel, synthesize | httpx, concurrent.futures |
| `stress_test.py` | Stress test for crossdoc RAG in maximum quality mode | crossdoc_client.py |
| `deploy.sh` | Main deployment script: starts services with compose + profile-specific overrides | bash, docker compose |
| `health_check.sh` | Health check for all RAG services: RAG server, ingestor, gateway, Milvus, NIMs | bash, curl |
| `setup.sh` | First-time setup: installs dependencies, applies patches, creates data directories | bash, git, docker |

## Design Notes

### `smart_ingest.py` — Tier System, Deadlock Detection, Resume

`smart_ingest.py` is an adaptive ingestion script that replaces the blueprint's basic ingestion tool. It was developed during production use on Brev to handle large document sets (500+ PDFs) with varying sizes and complexity.

#### Tier System

Classifies PDFs by page count and adapts polling and batch restart behavior:

- **Tiny**: ≤20 pages → 2s poll interval, restart after 30 files
- **Small**: 21-80 pages → 3s poll interval, restart after 15 files
- **Medium**: 81-250 pages → 5s poll interval, restart after 5 files
- **Large**: 251+ pages → 10s poll interval, restart after 3 files

**Timeouts**: Dynamic per file via `calc_timeout(pages)` function (roughly 2x expected processing time + overhead). Example: 20-page PDF gets ~35s timeout, 100-page PDF gets ~80s timeout.

**Batch restarts**: After processing N files (tier-specific), NV-Ingest is restarted to clear accumulated memory leaks and prevent slowdown.

This prevents small files from timing out too early, large files from blocking the queue indefinitely, and accumulated memory issues from degrading performance over time.

#### Deadlock Detection

Monitors ingestion progress by polling `/health` and checking the `extraction_count` field. If `extraction_count` does not increase for 45 seconds (configurable `DEADLOCK_THRESHOLD`), the script assumes the ingestor is stuck and either:
- Restarts the ingestor container (if restart count < tier limit)
- Skips the file and moves to the next (if restart limit exceeded)

This prevents a single problematic PDF from blocking the entire queue.

#### Resume Support

Tracks ingested file hashes in a `.ingested_files.json` file. When run with `--resume`, the script skips files that have already been successfully ingested. This allows resuming a large ingestion job after a crash or manual interruption without re-processing already-ingested files.

#### Usage

```bash
# Basic ingestion
python3 smart_ingest.py tecpia_test /path/to/docs/

# Fresh ingestion (delete collection first)
python3 smart_ingest.py tecpia_test /path/to/docs/ --fresh --confirm-delete

# Resume after interruption
python3 smart_ingest.py tecpia_test /path/to/docs/ --resume

# Dry run (show what would be ingested)
python3 smart_ingest.py tecpia_test /path/to/docs/ --dry-run
```
