# saldivia/ingestion_worker.py
"""Ingestion worker - processes jobs from queue."""
import os
import time
import signal
import logging
import httpx
import redis.exceptions
from pathlib import Path
from saldivia.ingestion_queue import IngestionQueue

logger = logging.getLogger(__name__)

INGESTOR_URL = os.getenv("INGESTOR_URL", "http://localhost:8082")

MAX_RETRIES = 3
RETRY_DELAYS = [10, 30, 60]  # seconds between retries

REDIS_RETRY_DELAY = 5   # seconds base delay on Redis unavailable
REDIS_MAX_DELAY = 60    # max backoff cap

_shutdown = False


def handle_sigterm(sig, frame):
    """Graceful shutdown on SIGTERM (docker stop)."""
    global _shutdown
    logger.info("SIGTERM received, finishing current job then stopping...")
    _shutdown = True


def process_job(job) -> bool:
    """Process a single ingestion job."""
    logger.info(f"Processing job {job.id}: {job.file_path}")

    file_path = Path(job.file_path)
    if not file_path.exists():
        logger.error(f"File not found: {file_path}")
        return False

    try:
        with open(file_path, "rb") as f:
            files = {"documents": (file_path.name, f, "application/pdf")}
            data = {"data": f'{{"collection_name": "{job.collection}"}}'}

            with httpx.Client(timeout=600) as client:
                resp = client.post(
                    f"{INGESTOR_URL}/v1/documents",
                    files=files,
                    data=data,
                )
                resp.raise_for_status()

        logger.info(f"Job {job.id} completed successfully")
        return True

    except Exception as e:
        logger.error(f"Job {job.id} failed: {e}")
        return False


def process_job_with_retry(job) -> bool:
    """Process a job with exponential backoff retries."""
    for attempt in range(MAX_RETRIES):
        if process_job(job):
            return True
        if attempt < MAX_RETRIES - 1:
            delay = RETRY_DELAYS[attempt]
            logger.info(f"Job {job.id} failed, retry {attempt + 1}/{MAX_RETRIES - 1} in {delay}s")
            time.sleep(delay)
    logger.error(f"Job {job.id} failed after {MAX_RETRIES} attempts")
    return False


def run_worker(redis_url: str = None):
    """Run the ingestion worker loop."""
    global _shutdown
    _shutdown = False
    signal.signal(signal.SIGTERM, handle_sigterm)

    redis_url = redis_url or os.getenv("REDIS_URL", "redis://localhost:6379")
    queue = IngestionQueue(redis_url)
    logger.info(f"Ingestion worker started (redis: {redis_url})")

    redis_delay = REDIS_RETRY_DELAY
    try:
        while not _shutdown:
            try:
                job = queue.dequeue()
                redis_delay = REDIS_RETRY_DELAY  # reset backoff on successful contact

                if job is None:
                    time.sleep(5)
                    continue

                queue.update_status(job.id, "processing")
                success = process_job_with_retry(job)
                queue.update_status(
                    job.id,
                    "completed" if success else "failed",
                    error=None if success else "Max retries exceeded"
                )
            except redis.exceptions.ConnectionError as e:
                logger.warning(f"Redis unavailable: {e}. Retrying in {redis_delay}s...")
                time.sleep(redis_delay)
                redis_delay = min(redis_delay * 2, REDIS_MAX_DELAY)
    finally:
        queue.close()
        logger.info("Ingestion worker stopped")


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    run_worker()
