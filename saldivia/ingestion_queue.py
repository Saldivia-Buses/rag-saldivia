# saldivia/ingestion_queue.py
"""Redis-backed ingestion job queue."""
import json
import logging
import redis
import uuid
from dataclasses import dataclass, asdict
from typing import Optional
from datetime import datetime

logger = logging.getLogger(__name__)


@dataclass
class IngestionJob:
    id: str
    file_path: str
    collection: str
    status: str  # pending, processing, completed, failed
    created_at: str
    started_at: Optional[str] = None
    completed_at: Optional[str] = None
    error: Optional[str] = None
    pages: Optional[int] = None


class IngestionQueue:
    """Manages ingestion jobs via Redis."""

    QUEUE_KEY = "ingestion_queue"
    JOBS_KEY = "ingestion_jobs"

    def __init__(self, redis_url: str = "redis://localhost:6379"):
        self.redis = redis.from_url(redis_url)

    def enqueue(self, file_path: str, collection: str) -> IngestionJob:
        """Add a file to the ingestion queue."""
        job = IngestionJob(
            id=str(uuid.uuid4())[:8],
            file_path=file_path,
            collection=collection,
            status="pending",
            created_at=datetime.now().isoformat(),
        )
        self.redis.lpush(self.QUEUE_KEY, job.id)
        self.redis.hset(self.JOBS_KEY, job.id, json.dumps(asdict(job)))
        return job

    def dequeue(self) -> Optional[IngestionJob]:
        """Get next job from queue."""
        job_id = self.redis.rpop(self.QUEUE_KEY)
        if not job_id:
            return None
        job_data = self.redis.hget(self.JOBS_KEY, job_id)
        if not job_data:
            logger.warning(f"Job {job_id!r} found in queue but no data in jobs hash — discarding")
            return None
        return IngestionJob(**json.loads(job_data))

    def update_status(self, job_id: str, status: str, error: str = None):
        """Update job status."""
        job_data = self.redis.hget(self.JOBS_KEY, job_id)
        if not job_data:
            return
        job = json.loads(job_data)
        job["status"] = status
        if status == "processing":
            job["started_at"] = datetime.now().isoformat()
        elif status in ("completed", "failed"):
            job["completed_at"] = datetime.now().isoformat()
        if error:
            job["error"] = error
        self.redis.hset(self.JOBS_KEY, job_id, json.dumps(job))

    def pending_count(self) -> int:
        """Get number of pending jobs."""
        return self.redis.llen(self.QUEUE_KEY)

    def list_jobs(self, status: str = None) -> list[IngestionJob]:
        """List all jobs, optionally filtered by status."""
        jobs = []
        for job_id in self.redis.hkeys(self.JOBS_KEY):
            job_data = self.redis.hget(self.JOBS_KEY, job_id)
            if job_data:
                job = IngestionJob(**json.loads(job_data))
                if status is None or job.status == status:
                    jobs.append(job)
        return sorted(jobs, key=lambda j: j.created_at, reverse=True)

    def clear_completed(self):
        """Remove completed/failed jobs from history."""
        for job_id in self.redis.hkeys(self.JOBS_KEY):
            job_data = self.redis.hget(self.JOBS_KEY, job_id)
            if job_data:
                job = json.loads(job_data)
                if job["status"] in ("completed", "failed"):
                    self.redis.hdel(self.JOBS_KEY, job_id)
