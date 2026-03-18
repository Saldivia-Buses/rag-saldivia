# saldivia/ingestion_worker.py
"""Ingestion worker - processes jobs from queue."""
import os
import time
import logging
import httpx
from pathlib import Path
from saldivia.ingestion_queue import IngestionQueue

logger = logging.getLogger(__name__)

INGESTOR_URL = os.getenv("INGESTOR_URL", "http://localhost:8082")


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


def run_worker(redis_url: str = "redis://localhost:6379"):
    """Run the ingestion worker loop."""
    queue = IngestionQueue(redis_url)
    logger.info("Ingestion worker started")

    while True:
        job = queue.dequeue()

        if job is None:
            time.sleep(5)
            continue

        queue.update_status(job.id, "processing")

        success = process_job(job)

        if success:
            queue.update_status(job.id, "completed")
        else:
            queue.update_status(job.id, "failed", error="See logs for details")


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    run_worker()
