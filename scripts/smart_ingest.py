#!/usr/bin/env python3
"""
Smart ingestion for NVIDIA RAG Blueprint v5 — Adaptive.
Auto-classifies PDFs by size, adapts poll/timeout/restart frequency.
Deadlock detection, resume support, smart restart wait.

Usage:
    python3 smart_ingest.py <collection> <pdf_dir_or_file> [pdf_dir_or_file ...]
    python3 smart_ingest.py tecpia_test /path/to/docs/
    python3 smart_ingest.py tecpia_test /path/to/single.pdf
    python3 smart_ingest.py tecpia_test /path/to/docs/ --fresh --confirm-delete
    python3 smart_ingest.py tecpia_test /path/to/docs/ --dry-run
    python3 smart_ingest.py tecpia_test /path/to/docs/ --resume
"""

import subprocess
import requests
import time
import json
import signal
import sys
import os
import shutil
import argparse
from dataclasses import dataclass, field, asdict

# === DEFAULTS ===
INGESTOR_URL = os.environ.get("INGESTOR_URL", "http://localhost:8082")
REPORT_EVERY = 5
DEADLOCK_THRESHOLD = 45  # seconds without extraction progress -> deadlock

# === TIER SYSTEM ===
from saldivia.tier import classify_tier as _classify_tier, TierConfig, TIERS as _TIERS


@dataclass
class Tier:
    """Local wrapper that adds a display name to saldivia.tier.TierConfig."""
    name: str
    poll_interval: float
    restart_after: int


TIERS: dict[str, Tier] = {
    k: Tier(name=k, poll_interval=v.poll_interval, restart_after=v.restart_after)
    for k, v in _TIERS.items()
}


def classify_tier(pages: int) -> Tier:
    """Classify a document by page count into a processing tier."""
    return TIERS[_classify_tier(page_count=pages)]

def calc_timeout(pages):
    """Adaptive timeout: ~2x expected time + overhead. Min 20s."""
    if pages <= 20:
        return max(20, int(pages * 1.0 + 15))
    elif pages <= 80:
        return int(pages * 0.5 + 30)
    else:
        return int(pages * 0.5 + 60)

def calc_max_pages(pages):
    """Adaptive split threshold. Large PDFs split to medium chunks."""
    if pages <= 250:
        return pages  # no split needed
    return 200  # split into ~200-page chunks


# === HELPERS ===

def log(msg, level="INFO"):
    ts = time.strftime("%H:%M:%S")
    print(f"[{ts}] [{level}] {msg}", flush=True)

def fmt_duration(seconds):
    if seconds < 60:
        return f"{int(seconds)}s"
    m, s = divmod(int(seconds), 60)
    if m < 60:
        return f"{m}m {s}s"
    h, m = divmod(m, 60)
    return f"{h}h {m}m {s}s"

def fmt_size(bytes_val):
    for unit in ['B', 'KB', 'MB', 'GB']:
        if bytes_val < 1024:
            return f"{bytes_val:.1f} {unit}"
        bytes_val /= 1024
    return f"{bytes_val:.1f} TB"


# === MILVUS (persistent connection) ===

class MilvusClient:
    """Persistent Milvus connection. Opens once, reuses, reconnects on error."""

    def __init__(self, host="localhost", port="19530"):
        self._host = host
        self._port = port
        self._connected = False

    def _ensure_connected(self):
        if not self._connected:
            from pymilvus import connections
            try:
                connections.disconnect("default")
            except:
                pass
            connections.connect(host=self._host, port=self._port)
            self._connected = True

    def _reconnect(self):
        self._connected = False
        self._ensure_connected()

    def get_entity_count(self, collection):
        from pymilvus import Collection
        for attempt in range(2):
            try:
                self._ensure_connected()
                col = Collection(collection)
                col.flush()
                return col.num_entities
            except Exception as e:
                if attempt == 0:
                    self._reconnect()
                else:
                    log(f"Error getting entity count: {e}", "WARN")
                    return -1

    def get_indexed_documents(self, collection):
        """Get indexed document names with pagination for 10K+ scale."""
        from pymilvus import Collection
        for attempt in range(2):
            try:
                self._ensure_connected()
                doc_info = Collection("document_info")
                doc_info.load()

                # Paginate to handle >16384 documents (Milvus limit)
                all_docs = set()
                batch_size = 10000
                offset = 0

                while True:
                    results = doc_info.query(
                        expr=f'collection_name == "{collection}" and info_type == "document"',
                        output_fields=["document_name"],
                        limit=batch_size,
                        offset=offset
                    )
                    if not results:
                        break
                    all_docs.update(r["document_name"] for r in results)
                    if len(results) < batch_size:
                        break
                    offset += batch_size

                return all_docs
            except Exception as e:
                if attempt == 0:
                    self._reconnect()
                else:
                    log(f"Error getting indexed docs: {e}", "WARN")
                    return set()

    def delete_collection(self, collection):
        from pymilvus import Collection, utility
        log(f"Deleting collection '{collection}' and metadata...")
        try:
            self._ensure_connected()
            for col_name in [collection, "document_info", "meta", "metadata_schema"]:
                if utility.has_collection(col_name):
                    col = Collection(col_name)
                    col.release()
                    col.drop()
                    log(f"  Dropped: {col_name}")
            return True
        except Exception as e:
            log(f"Error deleting: {e}", "ERROR")
            return False

    def verify_sparse_field(self, collection):
        from pymilvus import Collection
        try:
            self._ensure_connected()
            col = Collection(collection)
            fields = {f.name: f.dtype for f in col.schema.fields}
            has_sparse = "sparse" in fields
            log(f"  Fields: {list(fields.keys())}, sparse={has_sparse}")
            return has_sparse
        except Exception as e:
            log(f"Error verifying schema: {e}", "ERROR")
            return False

    def collection_exists(self, collection):
        from pymilvus import utility
        try:
            self._ensure_connected()
            return utility.has_collection(collection)
        except:
            return False

    def close(self):
        if self._connected:
            try:
                from pymilvus import connections
                connections.disconnect("default")
            except:
                pass
            self._connected = False


# === INFRASTRUCTURE ===

def create_collection(collection):
    log(f"Creating collection '{collection}' via ingestor API...")
    try:
        r = requests.post(
            f"{INGESTOR_URL}/v1/collections",
            json=[collection],
            timeout=30
        )
        if r.status_code == 200:
            log(f"  Created: {collection}")
            return True
        else:
            log(f"  Failed: {r.status_code} {r.text}", "ERROR")
            return False
    except Exception as e:
        log(f"  Error: {e}", "ERROR")
        return False

def check_nvingest_health():
    """Quick health check. Returns True if NV-Ingest is responding."""
    try:
        r = requests.get(f"{INGESTOR_URL}/v1/health", timeout=5)
        return r.status_code == 200
    except:
        return False

def restart_nvingest():
    """Restart NV-Ingest + Ingestor + flush Redis. Uses smart wait."""
    log("Restarting NV-Ingest + Ingestor + flushing Redis...")
    t0 = time.time()
    subprocess.run(
        ["docker", "restart", "compose-nv-ingest-ms-runtime-1", "ingestor-server"],
        capture_output=True, timeout=60
    )
    subprocess.run(
        ["docker", "exec", "redis", "redis-cli", "FLUSHALL"],
        capture_output=True, timeout=10
    )
    # Smart wait: poll health every 3s instead of sleeping 30s
    healthy = smart_restart_wait(max_wait=90)
    dur = int(time.time() - t0)
    if healthy:
        log(f"  Containers healthy (restart took {dur}s)")
    else:
        log(f"  Containers not healthy after {dur}s!", "ERROR")
    return healthy, dur

def smart_restart_wait(max_wait=90):
    """Poll health every 3s until healthy or max_wait. Returns True if healthy."""
    # Give containers a moment to start shutting down/restarting
    time.sleep(5)
    for _ in range(max_wait // 3):
        if check_nvingest_health():
            return True
        time.sleep(3)
    return False


# === PDF HANDLING ===

def get_page_count(filepath):
    """Get page count from PDF. Returns -1 if unreadable (corrupted/encrypted/invalid)."""
    try:
        import fitz
        doc = fitz.open(filepath)

        # Check for encryption
        if doc.is_encrypted:
            log(f"  SKIP (encrypted): {os.path.basename(filepath)}", "WARN")
            doc.close()
            return -1

        pages = len(doc)
        doc.close()
        return pages
    except Exception as e:
        error_msg = str(e).lower()
        if "password" in error_msg or "encrypt" in error_msg:
            log(f"  SKIP (password-protected): {os.path.basename(filepath)}", "WARN")
        elif "corrupt" in error_msg or "invalid" in error_msg or "damage" in error_msg:
            log(f"  SKIP (corrupted): {os.path.basename(filepath)} - {e}", "WARN")
        else:
            log(f"  SKIP (unreadable): {os.path.basename(filepath)} - {e}", "WARN")
        return -1


def check_docker_running():
    """Check if Docker is running. Returns (ok, error_msg)."""
    try:
        result = subprocess.run(
            ["docker", "info"],
            capture_output=True,
            timeout=10
        )
        if result.returncode == 0:
            return True, None
        return False, "Docker not responding"
    except FileNotFoundError:
        return False, "Docker CLI not found"
    except subprocess.TimeoutExpired:
        return False, "Docker info timed out"
    except Exception as e:
        return False, str(e)


def check_containers_exist():
    """Check if required containers exist. Returns (ok, missing_list)."""
    required = ["compose-nv-ingest-ms-runtime-1", "ingestor-server", "redis"]
    try:
        result = subprocess.run(
            ["docker", "ps", "-a", "--format", "{{.Names}}"],
            capture_output=True,
            text=True,
            timeout=10
        )
        existing = set(result.stdout.strip().split("\n"))
        missing = [c for c in required if c not in existing]
        return len(missing) == 0, missing
    except:
        return False, required  # Assume all missing if can't check

def check_disk_space(path, required_mb=500):
    """Check if there's enough disk space. Returns (ok, available_mb)."""
    try:
        stat = os.statvfs(path)
        available_mb = (stat.f_bavail * stat.f_frsize) / (1024 * 1024)
        return available_mb >= required_mb, available_mb
    except:
        return True, -1  # Can't check, assume OK


def split_pdf(filepath, max_pages, split_dir):
    import fitz
    doc = fitz.open(filepath)
    total = len(doc)
    base = os.path.splitext(os.path.basename(filepath))[0]

    if total <= max_pages:
        doc.close()
        return [filepath]

    # Check disk space before splitting (estimate: 2x original file size)
    file_size_mb = os.path.getsize(filepath) / (1024 * 1024)
    required_mb = file_size_mb * 2 + 100  # 2x file + 100MB buffer
    ok, available = check_disk_space(os.path.dirname(split_dir), required_mb)
    if not ok:
        log(f"  WARNING: Low disk space ({available:.0f}MB available, {required_mb:.0f}MB needed)", "WARN")

    n_chunks = (total + max_pages - 1) // max_pages
    chunk_size = (total + n_chunks - 1) // n_chunks

    os.makedirs(split_dir, exist_ok=True)
    chunks = []

    for i in range(n_chunks):
        start = i * chunk_size
        end = min((i + 1) * chunk_size, total)
        chunk_path = os.path.join(split_dir, f"{base}_part{i+1}of{n_chunks}.pdf")

        try:
            new_doc = fitz.open()
            new_doc.insert_pdf(doc, from_page=start, to_page=end - 1)
            new_doc.save(chunk_path)
            new_doc.close()
            chunks.append(chunk_path)
            log(f"    Part {i+1}/{n_chunks}: pages {start+1}-{end} -> {os.path.basename(chunk_path)}")
        except Exception as e:
            log(f"    SPLIT ERROR part {i+1}: {e}", "ERROR")
            doc.close()
            return chunks  # Return what we have so far

    doc.close()
    return chunks

def collect_pdfs(paths, skip_set):
    """Collect PDF file paths from arguments. Returns (pdf_files, total_bytes)."""
    pdf_files = []
    total_bytes = 0
    for path in paths:
        if os.path.isfile(path) and path.lower().endswith(".pdf"):
            pdf_files.append(path)
            total_bytes += os.path.getsize(path)
        elif os.path.isdir(path):
            for f in sorted(os.listdir(path)):
                if f.lower().endswith(".pdf") and f not in skip_set:
                    fpath = os.path.join(path, f)
                    pdf_files.append(fpath)
                    total_bytes += os.path.getsize(fpath)
        else:
            log(f"Skipping invalid path: {path}", "WARN")
    return pdf_files, total_bytes

def scan_pdfs(pdf_files, indexed):
    """Scan PDFs, split large ones adaptively, build work items.
    Returns (work_items, total_pages, split_dir)."""
    work_items = []  # (original_name, filepath, pages, file_size_bytes, tier)
    total_pages = 0
    split_dir = os.path.join(os.path.dirname(pdf_files[0]), "auto-split") if pdf_files else "/tmp/auto-split"

    for pdf_path in pdf_files:
        pdf_name = os.path.basename(pdf_path)
        pages = get_page_count(pdf_path)
        if pages <= 0:
            log(f"  SKIP ({pages} pages): {pdf_name}", "WARN")
            continue

        file_size = os.path.getsize(pdf_path)
        max_pages = calc_max_pages(pages)

        if pages > max_pages:
            log(f"  {pdf_name}: {pages} pages ({fmt_size(file_size)}) -> splitting at {max_pages}p...")
            chunks = split_pdf(pdf_path, max_pages, split_dir)
            for chunk in chunks:
                cp = get_page_count(chunk)
                cs = os.path.getsize(chunk)
                tier = classify_tier(cp)
                work_items.append((pdf_name, chunk, cp, cs, tier))
                total_pages += cp
        else:
            if pdf_name in indexed:
                log(f"  SKIP (indexed): {pdf_name}")
                continue
            tier = classify_tier(pages)
            log(f"  {pdf_name}: {pages} pages ({fmt_size(file_size)}) -> {tier.name}")
            work_items.append((pdf_name, pdf_path, pages, file_size, tier))
            total_pages += pages

    return work_items, total_pages, split_dir


# === STATE PERSISTENCE ===

def state_file_path(collection):
    return f"/tmp/ingest_state_{collection}.json"

def save_state(collection, completed_files, failed_files):
    """Save progress after each successful PDF. Uses atomic write to prevent corruption."""
    state = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "collection": collection,
        "completed": list(completed_files),
        "failed": list(failed_files),
    }
    path = state_file_path(collection)
    tmp_path = path + ".tmp"

    # Atomic write: write to temp file, then rename
    try:
        with open(tmp_path, "w") as fp:
            json.dump(state, fp, indent=2)
        os.replace(tmp_path, path)  # atomic on POSIX
    except Exception as e:
        log(f"Warning: failed to save state: {e}", "WARN")
        # Try to remove tmp file if it exists
        try:
            os.remove(tmp_path)
        except:
            pass

def load_state(collection):
    """Load previous state. Returns (completed_set, failed_set) or (set(), set())."""
    path = state_file_path(collection)
    if not os.path.exists(path):
        return set(), set()
    try:
        with open(path, "r") as fp:
            state = json.load(fp)
        completed = set(state.get("completed", []))
        failed = set(state.get("failed", []))
        log(f"Resumed state: {len(completed)} completed, {len(failed)} failed (from {state.get('timestamp', '?')})")
        return completed, failed
    except Exception as e:
        log(f"Error loading state: {e}", "WARN")
        return set(), set()

def clear_state(collection):
    path = state_file_path(collection)
    if os.path.exists(path):
        os.remove(path)


# === INGESTION (adaptive) ===

@dataclass
class IngestResult:
    success: bool
    duration: float
    new_entities: int = 0
    deadlocked: bool = False
    error: str = ""

def ingest_single(filepath, collection, tier, milvus):
    """Ingest a single PDF with adaptive polling and deadlock detection.
    Returns IngestResult."""
    name = os.path.basename(filepath)
    pages = get_page_count(filepath)
    timeout = calc_timeout(pages) if pages > 0 else 120

    if not os.path.exists(filepath):
        log(f"  File not found: {filepath}", "ERROR")
        return IngestResult(success=False, duration=0, error="file not found")

    entities_before = milvus.get_entity_count(collection)

    log(f"  Submitting: {name} (timeout={timeout}s, poll={tier.poll_interval}s)")
    start = time.time()
    try:
        with open(filepath, "rb") as f:
            r = requests.post(
                f"{INGESTOR_URL}/v1/documents",
                files={"documents": (name, f, "application/pdf")},
                data={"data": json.dumps({"collection_name": collection})},
                timeout=120
            )
        resp = r.json()
        task_id = resp.get("task_id")
        if not task_id:
            log(f"  No task_id: {resp}", "ERROR")
            return IngestResult(success=False, duration=0, error="no task_id")
        log(f"  task_id: {task_id}")
    except Exception as e:
        log(f"  Submit failed: {e}", "ERROR")
        return IngestResult(success=False, duration=0, error=str(e))

    # Adaptive polling with deadlock detection
    last_extracted = -1
    last_progress_time = time.time()
    last_log_info = ""

    while (time.time() - start) < timeout:
        time.sleep(tier.poll_interval)
        elapsed = int(time.time() - start)
        try:
            r = requests.get(
                f"{INGESTOR_URL}/v1/status",
                params={"task_id": task_id},
                timeout=10
            )
            status = r.json()
            state = status.get("state", "UNKNOWN")

            if state == "FINISHED":
                result = status.get("result", {})
                docs_completed = result.get("documents_completed", 0)
                total_docs = result.get("total_documents", 0)
                duration = time.time() - start
                entities_after = milvus.get_entity_count(collection)
                new_entities = max(0, entities_after - entities_before)
                log(f"  FINISHED in {int(duration)}s ({docs_completed}/{total_docs} docs, +{new_entities} entities)")
                return IngestResult(success=True, duration=duration, new_entities=new_entities)

            elif state == "FAILED":
                msg = status.get("result", {}).get("message", "unknown")
                log(f"  FAILED: {msg}", "ERROR")
                return IngestResult(success=False, duration=time.time() - start, error=msg)

            else:
                nv = status.get("nv_ingest_status", {})
                extracted = nv.get("extraction_completed", 0)

                # Deadlock detection
                if extracted != last_extracted:
                    last_extracted = extracted
                    last_progress_time = time.time()
                elif time.time() - last_progress_time > DEADLOCK_THRESHOLD:
                    stall = int(time.time() - last_progress_time)
                    log(f"  DEADLOCK: extraction stuck at {extracted} for {stall}s", "ERROR")
                    return IngestResult(
                        success=False, duration=time.time() - start,
                        deadlocked=True, error=f"deadlock at extraction={extracted}"
                    )

                info = f"{elapsed}s state={state} extracted={extracted}"
                # Log every 30s or on tier-appropriate intervals
                log_interval = max(30, tier.poll_interval * 6)
                if elapsed % int(log_interval) < tier.poll_interval and info != last_log_info:
                    log(f"    ... {info}")
                    last_log_info = info

        except Exception as e:
            log(f"    Status error ({elapsed}s): {e}", "WARN")

    log(f"  TIMEOUT after {timeout}s", "ERROR")
    return IngestResult(success=False, duration=timeout, error="timeout")

def try_ingest_with_retry(filepath, collection, tier, milvus, split_dir):
    """Multi-level retry:
    Level 1: restart NV-Ingest + retry same file
    Level 2: if pages > 100, re-split to half size and retry chunks
    Level 3: skip with clear error
    Returns (success, total_new_entities, total_duration, restart_time, retried, resplit)."""

    # First attempt
    result = ingest_single(filepath, collection, tier, milvus)

    if result.success and result.new_entities > 0:
        return True, result.new_entities, result.duration, 0, False, False

    # Level 1: restart + retry
    reason = "deadlock" if result.deadlocked else ("0 entities" if result.success else "failed/timeout")
    log(f"  L1 retry ({reason}): restarting NV-Ingest...", "WARN")
    _, rd = restart_nvingest()

    result2 = ingest_single(filepath, collection, tier, milvus)
    if result2.success and result2.new_entities > 0:
        return True, result2.new_entities, result2.duration, rd, True, False

    # Level 2: re-split if large enough
    pages = get_page_count(filepath)
    if pages > 100:
        half = pages // 2
        log(f"  L2 retry: re-splitting {os.path.basename(filepath)} at {half} pages...", "WARN")
        os.makedirs(split_dir, exist_ok=True)
        chunks = split_pdf(filepath, half, split_dir)
        total_entities = 0
        total_dur = 0.0
        all_ok = True
        for chunk in chunks:
            chunk_pages = get_page_count(chunk)
            chunk_tier = classify_tier(chunk_pages)
            cr = ingest_single(chunk, collection, chunk_tier, milvus)
            total_dur += cr.duration
            if cr.success and cr.new_entities > 0:
                total_entities += cr.new_entities
            else:
                all_ok = False
                log(f"  L2 chunk failed: {os.path.basename(chunk)}", "ERROR")
        if total_entities > 0:
            return all_ok, total_entities, total_dur, rd, True, True

    # Level 3: skip
    return False, 0, result2.duration, rd, True, False


# === THROUGHPUT TRACKER ===

class ThroughputTracker:
    """Track actual throughput for adaptive ETA."""

    def __init__(self):
        self.samples = []  # (pages, duration)

    def record(self, pages, duration):
        if duration > 0 and pages > 0:
            self.samples.append((pages, duration))

    def avg_pages_per_sec(self):
        if not self.samples:
            return 3.0  # default from measurements
        total_pages = sum(p for p, _ in self.samples)
        total_time = sum(d for _, d in self.samples)
        return total_pages / total_time if total_time > 0 else 3.0

    def eta(self, remaining_pages):
        pps = self.avg_pages_per_sec()
        return remaining_pages / pps if pps > 0 else 0


# === MAIN INGESTION LOOP ===

def run_ingestion(work_items, total_pages, collection, indexed, milvus, resume_state):
    """Adaptive ingestion loop with resume support.
    Returns (results, initial_entities, ingest_time, restart_overhead)."""
    initial_entities = milvus.get_entity_count(collection)
    log(f"Initial entities: {initial_entities}")

    completed_files, failed_state = resume_state
    results = {"success": [], "failed": []}
    tracker = ThroughputTracker()
    pages_done = 0
    ingest_time_total = 0.0
    restart_overhead = 0.0
    chunks_since_report = 0

    # Track files since last restart per-tier
    since_restart = 0

    for idx, (original_name, filepath, pages, file_size, tier) in enumerate(work_items):
        if _shutdown_requested:
            log("Shutdown requested — stopping after current chunk", "WARN")
            break

        item_name = os.path.basename(filepath)

        # Resume: skip already completed
        if item_name in completed_files:
            pages_done += pages
            continue

        log("")
        log(f"[{idx+1}/{len(work_items)}] {item_name} ({pages}p, {fmt_size(file_size)}, tier={tier.name})")

        if item_name in indexed:
            log(f"  SKIP (already indexed)")
            completed_files.add(item_name)
            pages_done += pages
            continue

        # ETA
        if pages_done > 0:
            eta_sec = tracker.eta(total_pages - pages_done)
            pps = tracker.avg_pages_per_sec()
            log(f"  Progress: {pages_done}/{total_pages} pages ({100*pages_done//total_pages}%) | "
                f"ETA: {fmt_duration(eta_sec)} | {pps:.1f} p/s actual")

        # Adaptive batch restart
        if since_restart >= tier.restart_after:
            log(f"  Batch limit ({tier.restart_after} for {tier.name}) -> restarting...")
            _, rd = restart_nvingest()
            restart_overhead += rd
            since_restart = 0

        # Health check before submit
        if not check_nvingest_health():
            log("  NV-Ingest not healthy — restarting...", "WARN")
            _, rd = restart_nvingest()
            restart_overhead += rd
            since_restart = 0

        # Ingest with adaptive retry
        split_dir = os.path.join(os.path.dirname(filepath), "auto-split-retry")
        success, new_entities, duration, rd, retried, resplit = try_ingest_with_retry(
            filepath, collection, tier, milvus, split_dir
        )
        restart_overhead += rd

        if success:
            pages_per_sec = pages / duration if duration > 0 else 0
            entities_per_page = new_entities / pages if pages > 0 else 0
            mb_per_min = (file_size / 1024 / 1024) / (duration / 60) if duration > 0 else 0
            prefix = "RESPLIT OK" if resplit else ("RETRY OK" if retried else "OK")
            log(f"  {prefix}: +{new_entities} entities in {fmt_duration(duration)}")
            log(f"     {pages_per_sec:.2f} p/s | {entities_per_page:.1f} e/p | {mb_per_min:.1f} MB/min")
            results["success"].append({
                "name": item_name, "original": original_name,
                "entities": new_entities, "duration": round(duration, 1), "pages": pages,
                "file_size": file_size,
                "pages_per_sec": round(pages_per_sec, 2),
                "entities_per_page": round(entities_per_page, 1),
                "mb_per_min": round(mb_per_min, 1),
                "retried": retried, "resplit": resplit,
                "tier": tier.name,
            })
            tracker.record(pages, duration)
            since_restart = 0 if retried else since_restart + 1
            pages_done += pages
            ingest_time_total += duration
            completed_files.add(item_name)
        else:
            log(f"  ALL RETRIES FAILED: {item_name} — skipping", "ERROR")
            results["failed"].append({
                "name": item_name, "original": original_name,
                "reason": "failed after all retry levels", "pages": pages,
                "tier": tier.name,
            })
            since_restart = 0
            pages_done += pages
            failed_state.add(item_name)

        # Save state after every file
        save_state(collection, completed_files, failed_state)

        # Periodic partial report
        chunks_since_report += 1
        if chunks_since_report >= REPORT_EVERY:
            save_partial_report(results, collection, initial_entities,
                                ingest_time_total, restart_overhead, milvus)
            chunks_since_report = 0

    return results, initial_entities, ingest_time_total, restart_overhead


# === REPORTS ===

def save_partial_report(results, collection, initial_entities,
                        ingest_time_total, restart_overhead, milvus):
    current_entities = milvus.get_entity_count(collection)
    report = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "status": "partial",
        "collection": collection,
        "summary": {
            "initial_entities": initial_entities,
            "current_entities": current_entities,
            "new_entities": current_entities - initial_entities,
            "chunks_ok": len(results["success"]),
            "chunks_failed": len(results["failed"]),
            "ingestion_time_s": round(ingest_time_total),
            "restart_overhead_s": round(restart_overhead),
        },
        "chunks": results["success"],
        "failed": results["failed"],
    }
    with open("/tmp/ingestion_report_partial.json", "w") as fp:
        json.dump(report, fp, indent=2, default=str)
    log(f"  Partial report saved ({len(results['success'])} ok, {len(results['failed'])} failed)")

def generate_report(results, collection, initial_entities, ingest_time_total,
                    restart_overhead, wall_total, config, milvus):
    final_entities = milvus.get_entity_count(collection)
    total_new = final_entities - initial_entities

    log("")
    log("=" * 60)
    log("FINAL REPORT")
    log("=" * 60)

    # Group by original PDF
    by_original = {}
    for s in results["success"]:
        orig = s["original"]
        if orig not in by_original:
            by_original[orig] = {
                "entities": 0, "duration": 0, "parts": 0,
                "pages": 0, "file_size": 0, "retried": False, "resplit": False,
                "tiers": set(),
            }
        by_original[orig]["entities"] += s["entities"]
        by_original[orig]["duration"] += s["duration"]
        by_original[orig]["parts"] += 1
        by_original[orig]["pages"] += s["pages"]
        by_original[orig]["file_size"] += s["file_size"]
        by_original[orig]["tiers"].add(s.get("tier", "?"))
        if s.get("retried"):
            by_original[orig]["retried"] = True
        if s.get("resplit"):
            by_original[orig]["resplit"] = True

    for info in by_original.values():
        d, p, e, sz = info["duration"], info["pages"], info["entities"], info["file_size"]
        info["pages_per_sec"] = round(p / d, 2) if d > 0 else 0
        info["entities_per_page"] = round(e / p, 1) if p > 0 else 0
        info["mb_per_min"] = round((sz / 1024 / 1024) / (d / 60), 1) if d > 0 else 0
        info["tiers"] = list(info["tiers"])

    total_success_pages = sum(i["pages"] for i in by_original.values())
    total_success_entities = sum(i["entities"] for i in by_original.values())
    total_success_bytes = sum(i["file_size"] for i in by_original.values())

    # Tier distribution
    tier_counts = {"tiny": 0, "small": 0, "medium": 0, "large": 0}
    for s in results["success"]:
        t = s.get("tier", "medium")
        tier_counts[t] = tier_counts.get(t, 0) + 1

    log(f"\n{'METRIC':<30} {'VALUE':>15}")
    log(f"{'-'*46}")
    log(f"{'Collection':<30} {collection:>15}")
    log(f"{'Entities (before -> after)':<30} {str(initial_entities) + ' -> ' + str(final_entities):>15}")
    log(f"{'New entities':<30} {'+' + str(total_new):>15}")
    log(f"{'PDFs processed':<30} {len(by_original):>15}")
    log(f"{'PDFs failed':<30} {len(results['failed']):>15}")
    log(f"{'Total pages':<30} {total_success_pages:>15}")
    log(f"{'Total input size':<30} {fmt_size(total_success_bytes):>15}")
    log(f"{'Ingestion time':<30} {fmt_duration(ingest_time_total):>15}")
    log(f"{'Restart overhead':<30} {fmt_duration(restart_overhead):>15}")
    log(f"{'Wall clock time':<30} {fmt_duration(wall_total):>15}")
    log(f"{'-'*46}")
    log(f"{'Tier distribution':<30} tiny={tier_counts['tiny']} small={tier_counts['small']} "
        f"med={tier_counts['medium']} large={tier_counts['large']}")

    if ingest_time_total > 0 and total_success_pages > 0:
        avg_pps = total_success_pages / ingest_time_total
        avg_epp = total_success_entities / total_success_pages
        avg_mpm = (total_success_bytes / 1024 / 1024) / (ingest_time_total / 60)
        avg_spp = ingest_time_total / total_success_pages
        log(f"{'Avg pages/sec':<30} {avg_pps:>15.2f}")
        log(f"{'Avg sec/page':<30} {avg_spp:>15.2f}")
        log(f"{'Avg entities/page':<30} {avg_epp:>15.1f}")
        log(f"{'Avg MB/min':<30} {avg_mpm:>15.1f}")
        log(f"{'-'*46}")

    log(f"\n{'PDF':<50} {'Tier':>6} {'Pages':>6} {'Ents':>7} {'Time':>8} {'p/s':>5} {'e/p':>5} {'Pts':>4}")
    log(f"{'-'*93}")
    for name in sorted(by_original.keys()):
        info = by_original[name]
        short = name[:47] + "..." if len(name) > 50 else name
        flags = ""
        if info["retried"]:
            flags += "R"
        if info["resplit"]:
            flags += "S"
        tier_str = ",".join(info["tiers"])
        log(f"{short:<50} {tier_str:>6} {info['pages']:>6} {info['entities']:>7} "
            f"{fmt_duration(info['duration']):>8} {info['pages_per_sec']:>5.1f} "
            f"{info['entities_per_page']:>5.1f} {str(info['parts']) + flags:>4}")

    if by_original:
        fastest = min(by_original.items(), key=lambda x: x[1]["duration"] / max(x[1]["pages"], 1))
        slowest = max(by_original.items(), key=lambda x: x[1]["duration"] / max(x[1]["pages"], 1))
        log(f"\nFastest: {fastest[0]} ({fastest[1]['pages_per_sec']} p/s)")
        log(f"Slowest: {slowest[0]} ({slowest[1]['pages_per_sec']} p/s)")

    if results["failed"]:
        log(f"\nFAILED ({len(results['failed'])}):", "ERROR")
        for f in results["failed"]:
            log(f"  {f['name']}: {f['reason']} ({f.get('pages', '?')} pages, tier={f.get('tier', '?')})", "ERROR")
    else:
        log("\nNo failures!")

    report = {
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
        "version": "v5-adaptive",
        "status": "complete",
        "collection": collection,
        "config": config,
        "summary": {
            "initial_entities": initial_entities,
            "final_entities": final_entities,
            "new_entities": total_new,
            "pdfs_processed": len(by_original),
            "pdfs_failed": len(results["failed"]),
            "total_pages": total_success_pages,
            "total_input_bytes": total_success_bytes,
            "total_input_size": fmt_size(total_success_bytes),
            "ingestion_time_s": round(ingest_time_total),
            "ingestion_time": fmt_duration(ingest_time_total),
            "restart_overhead_s": round(restart_overhead),
            "restart_overhead": fmt_duration(restart_overhead),
            "wall_clock_s": round(wall_total),
            "wall_clock": fmt_duration(wall_total),
            "tier_distribution": tier_counts,
        },
        "averages": {},
        "by_pdf": {k: {kk: vv for kk, vv in v.items() if kk != "tiers"} | {"tiers": v["tiers"]}
                   for k, v in by_original.items()},
        "chunks": results["success"],
        "failed": results["failed"],
    }

    if ingest_time_total > 0 and total_success_pages > 0:
        report["averages"] = {
            "pages_per_sec": round(total_success_pages / ingest_time_total, 2),
            "sec_per_page": round(ingest_time_total / total_success_pages, 2),
            "entities_per_page": round(total_success_entities / total_success_pages, 1),
            "mb_per_min": round((total_success_bytes / 1024 / 1024) / (ingest_time_total / 60), 1),
        }

    report_path = "/tmp/ingestion_report.json"
    with open(report_path, "w") as fp:
        json.dump(report, fp, indent=2, default=str)
    log(f"\nReport saved: {report_path}")

    return report


# === SIGNAL HANDLING ===

_shutdown_requested = False

def _handle_sigint(signum, frame):
    global _shutdown_requested
    if _shutdown_requested:
        log("Second interrupt — forcing exit", "FATAL")
        sys.exit(1)
    _shutdown_requested = True
    log("")
    log("SIGINT received — will stop after current chunk finishes", "WARN")
    log("Press Ctrl+C again to force exit")


# === MAIN ===

def handle_fresh_mode(collection, milvus):
    log("")
    log("FRESH START: deleting collection...")
    milvus.delete_collection(collection)
    create_collection(collection)
    time.sleep(3)
    _, rd = restart_nvingest()
    time.sleep(5)
    if not milvus.verify_sparse_field(collection):
        log("Collection missing sparse field! Check ingestor SEARCHTYPE config.", "FATAL")
        sys.exit(1)
    clear_state(collection)
    return rd

def ensure_collection(collection, milvus):
    if not milvus.collection_exists(collection):
        log(f"Collection '{collection}' not found, creating...")
        create_collection(collection)
        time.sleep(5)
        milvus.verify_sparse_field(collection)

def preflight_checks(dry_run=False):
    """Run pre-flight checks before ingestion. Returns (ok, errors)."""
    errors = []

    # Check pymilvus
    try:
        import pymilvus
    except ImportError:
        errors.append("pymilvus not installed: pip install pymilvus")

    # Check fitz (pymupdf)
    try:
        import fitz
    except ImportError:
        errors.append("pymupdf not installed: pip install pymupdf")

    if dry_run:
        return len(errors) == 0, errors

    # Check Docker
    ok, msg = check_docker_running()
    if not ok:
        errors.append(f"Docker: {msg}")
    else:
        # Check containers
        ok, missing = check_containers_exist()
        if not ok:
            errors.append(f"Missing containers: {missing}")

    return len(errors) == 0, errors


def main():
    parser = argparse.ArgumentParser(description="Smart PDF ingestion v5 — Adaptive")
    parser.add_argument("collection", help="Milvus collection name")
    parser.add_argument("paths", nargs="+", help="PDF files or directories to ingest")
    parser.add_argument("--fresh", action="store_true",
                        help="Delete collection and start fresh (requires --confirm-delete)")
    parser.add_argument("--confirm-delete", action="store_true",
                        help="Confirm destructive deletion when using --fresh")
    parser.add_argument("--skip", nargs="*", default=[], help="PDF filenames to skip")
    parser.add_argument("--dry-run", action="store_true",
                        help="Scan and classify PDFs but don't ingest")
    parser.add_argument("--resume", action="store_true",
                        help="Resume from last saved state")
    parser.add_argument("--skip-preflight", action="store_true",
                        help="Skip pre-flight checks (use if you know what you're doing)")
    args = parser.parse_args()

    signal.signal(signal.SIGINT, _handle_sigint)

    # Pre-flight checks
    if not args.skip_preflight:
        ok, errors = preflight_checks(dry_run=args.dry_run)
        if not ok:
            log("PRE-FLIGHT CHECK FAILED:", "FATAL")
            for e in errors:
                log(f"  - {e}", "ERROR")
            return 1

    collection = args.collection
    milvus = MilvusClient()
    wall_start = time.time()
    restart_overhead = 0.0

    log("=" * 60)
    log("SMART INGESTION v5 — Adaptive")
    log(f"Collection: {collection}")
    log(f"Fresh: {args.fresh}")
    log(f"Resume: {args.resume}")
    if args.dry_run:
        log("Mode: DRY RUN (no ingestion)")
    log("=" * 60)
    log("")
    log("Tier system:")
    for t in TIERS.values():
        log(f"  {t.name:>6}: poll={t.poll_interval}s, restart_after={t.restart_after}")
    log(f"  Deadlock threshold: {DEADLOCK_THRESHOLD}s")

    # 1. Collect PDFs
    pdf_files, total_input_bytes = collect_pdfs(args.paths, set(args.skip))
    if not pdf_files:
        log("No PDF files found!", "ERROR")
        milvus.close()
        return 1
    log(f"\nFound {len(pdf_files)} PDFs ({fmt_size(total_input_bytes)} total)")

    # 2. Fresh or incremental
    if not args.dry_run:
        if args.fresh:
            if not args.confirm_delete:
                log("--fresh requires --confirm-delete to prevent accidental data loss.", "FATAL")
                milvus.close()
                return 1
            restart_overhead += handle_fresh_mode(collection, milvus)
        else:
            ensure_collection(collection, milvus)

    # 3. Scan & classify
    indexed = set() if args.dry_run else milvus.get_indexed_documents(collection)
    if indexed:
        log(f"Already indexed: {len(indexed)} documents")

    # Resume state
    resume_state = (set(), set())
    if args.resume and not args.dry_run:
        resume_state = load_state(collection)

    log("")
    log("Scanning & classifying PDFs...")
    work_items, total_pages, split_dir = scan_pdfs(pdf_files, indexed)

    if not work_items:
        log("Nothing to ingest — all PDFs already indexed or skipped.")
        milvus.close()
        return 0

    # Filter out resumed files for count
    pending_items = [w for w in work_items if os.path.basename(w[1]) not in resume_state[0]]
    pending_pages = sum(w[2] for w in pending_items)

    log(f"\nWork items: {len(work_items)} total, {len(pending_items)} pending, {total_pages} total pages")

    # Tier breakdown
    tier_summary = {}
    for _, _, pages, _, tier in pending_items:
        tier_summary[tier.name] = tier_summary.get(tier.name, 0) + 1
    log(f"Tier breakdown: {tier_summary}")

    # Dry run
    if args.dry_run:
        log("")
        log("=" * 60)
        log("DRY RUN — Adaptive Plan")
        log("=" * 60)

        # Estimate time per tier
        est_time = 0
        est_restarts = 0
        tier_file_counts = {"tiny": 0, "small": 0, "medium": 0, "large": 0}
        for _, _, pages, _, tier in work_items:
            est_time += calc_timeout(pages) * 0.5  # expect ~50% of timeout on average
            tier_file_counts[tier.name] += 1

        for tname, tobj in TIERS.items():
            count = tier_file_counts[tname]
            if count > 0:
                est_restarts += count // tobj.restart_after

        est_restart_time = est_restarts * 15  # ~15s avg with smart wait
        est_total = est_time + est_restart_time

        log(f"  Chunks to process:     {len(work_items)}")
        log(f"  Total pages:           {total_pages}")
        log(f"  Estimated restarts:    {est_restarts} (~{fmt_duration(est_restart_time)})")
        log(f"  Estimated ingest time: {fmt_duration(est_time)}")
        log(f"  Estimated total:       {fmt_duration(est_total)}")
        log("")
        log(f"  {'#':>4}  {'File':<55} {'Pages':>5} {'Tier':>6} {'Timeout':>7}")
        log(f"  {'-'*82}")
        for idx, (orig, fpath, pages, fsize, tier) in enumerate(work_items):
            name = os.path.basename(fpath)
            short = name[:52] + "..." if len(name) > 55 else name
            timeout = calc_timeout(pages)
            log(f"  [{idx+1:>3}] {short:<55} {pages:>5} {tier.name:>6} {timeout:>6}s")

        if os.path.exists(split_dir):
            shutil.rmtree(split_dir)
            log(f"\nCleaned up: {split_dir}")
        milvus.close()
        return 0

    # 4. Ingest
    log("")
    log("=" * 60)
    log("INGESTING (adaptive)")
    log("=" * 60)

    config = {
        "tiers": {k: {"poll": v.poll_interval, "restart_after": v.restart_after}
                  for k, v in TIERS.items()},
        "deadlock_threshold": DEADLOCK_THRESHOLD,
        "resumed": args.resume,
    }

    results, initial_entities, ingest_time_total, ingest_restart = run_ingestion(
        work_items, total_pages, collection, indexed, milvus, resume_state
    )
    restart_overhead += ingest_restart

    # 5. Report
    wall_total = time.time() - wall_start
    generate_report(results, collection, initial_entities, ingest_time_total,
                    restart_overhead, wall_total, config, milvus)

    # 6. Cleanup
    if os.path.exists(split_dir):
        if not results["failed"]:
            shutil.rmtree(split_dir)
            log(f"Cleaned up: {split_dir}")
        else:
            log(f"Splits preserved (failures detected): {split_dir}")

    # Clear state on full success
    if not results["failed"] and not _shutdown_requested:
        clear_state(collection)
        log("State cleared (all files completed)")

    milvus.close()
    return 0 if not results["failed"] else 1

if __name__ == "__main__":
    sys.exit(main())
