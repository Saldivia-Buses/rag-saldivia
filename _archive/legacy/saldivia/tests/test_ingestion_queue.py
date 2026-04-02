# saldivia/tests/test_ingestion_queue.py
"""Tests for IngestionQueue and IngestionJob using fakeredis."""
import pytest
import fakeredis
from unittest.mock import MagicMock
from redis.exceptions import ConnectionError as RedisConnectionError

from saldivia.ingestion_queue import IngestionJob, IngestionQueue


# ---------------------------------------------------------------------------
# Fixture
# ---------------------------------------------------------------------------

@pytest.fixture
def queue():
    """IngestionQueue backed by an in-process fakeredis server."""
    fake = fakeredis.FakeRedis(decode_responses=False)
    q = IngestionQueue.__new__(IngestionQueue)
    q.redis = fake
    return q


# ---------------------------------------------------------------------------
# enqueue
# ---------------------------------------------------------------------------

def test_enqueue_returns_ingestion_job(queue):
    job = queue.enqueue("/tmp/doc.pdf", "my-collection")
    assert isinstance(job, IngestionJob)


def test_enqueue_job_fields(queue):
    job = queue.enqueue("/tmp/doc.pdf", "my-collection")
    assert job.file_path == "/tmp/doc.pdf"
    assert job.collection == "my-collection"
    assert job.status == "pending"
    assert job.id  # non-empty UUID string
    assert job.created_at  # ISO timestamp


def test_enqueue_job_optional_fields_default_none(queue):
    job = queue.enqueue("/tmp/doc.pdf", "col")
    assert job.started_at is None
    assert job.completed_at is None
    assert job.error is None
    assert job.pages is None


# ---------------------------------------------------------------------------
# dequeue
# ---------------------------------------------------------------------------

def test_dequeue_returns_enqueued_job(queue):
    job = queue.enqueue("/tmp/a.pdf", "col")
    result = queue.dequeue()
    assert result is not None
    assert result.id == job.id
    assert result.file_path == "/tmp/a.pdf"


def test_dequeue_empty_queue_returns_none(queue):
    result = queue.dequeue()
    assert result is None


# ---------------------------------------------------------------------------
# FIFO ordering
# ---------------------------------------------------------------------------

def test_fifo_order(queue):
    job1 = queue.enqueue("/tmp/first.pdf", "col")
    job2 = queue.enqueue("/tmp/second.pdf", "col")
    dequeued1 = queue.dequeue()
    dequeued2 = queue.dequeue()
    assert dequeued1.id == job1.id
    assert dequeued2.id == job2.id


# ---------------------------------------------------------------------------
# update_status
# ---------------------------------------------------------------------------

def test_update_status_processing(queue):
    job = queue.enqueue("/tmp/doc.pdf", "col")
    queue.update_status(job.id, "processing")
    # Confirm via list_jobs
    jobs = queue.list_jobs()
    updated = next(j for j in jobs if j.id == job.id)
    assert updated.status == "processing"
    assert updated.started_at is not None
    assert updated.completed_at is None


def test_update_status_completed(queue):
    job = queue.enqueue("/tmp/doc.pdf", "col")
    queue.update_status(job.id, "completed")
    jobs = queue.list_jobs()
    updated = next(j for j in jobs if j.id == job.id)
    assert updated.status == "completed"
    assert updated.completed_at is not None


def test_update_status_failed_with_error(queue):
    job = queue.enqueue("/tmp/doc.pdf", "col")
    queue.update_status(job.id, "failed", error="timeout after 30s")
    jobs = queue.list_jobs()
    updated = next(j for j in jobs if j.id == job.id)
    assert updated.status == "failed"
    assert updated.completed_at is not None
    assert updated.error == "timeout after 30s"


def test_update_status_nonexistent_job_no_crash(queue):
    # Must not raise
    queue.update_status("nonexistent-id", "failed", error="boom")


# ---------------------------------------------------------------------------
# list_jobs
# ---------------------------------------------------------------------------

def test_list_jobs_no_filter_returns_all(queue):
    queue.enqueue("/tmp/a.pdf", "col")
    queue.enqueue("/tmp/b.pdf", "col")
    jobs = queue.list_jobs()
    assert len(jobs) == 2


def test_list_jobs_filter_by_status(queue):
    j1 = queue.enqueue("/tmp/a.pdf", "col")
    j2 = queue.enqueue("/tmp/b.pdf", "col")
    queue.update_status(j1.id, "processing")
    pending = queue.list_jobs(status="pending")
    processing = queue.list_jobs(status="processing")
    assert len(pending) == 1
    assert pending[0].id == j2.id
    assert len(processing) == 1
    assert processing[0].id == j1.id


def test_list_jobs_empty_queue_returns_empty(queue):
    assert queue.list_jobs() == []


# ---------------------------------------------------------------------------
# clear_completed
# ---------------------------------------------------------------------------

def test_clear_completed_removes_completed_and_failed(queue):
    j1 = queue.enqueue("/tmp/a.pdf", "col")
    j2 = queue.enqueue("/tmp/b.pdf", "col")
    j3 = queue.enqueue("/tmp/c.pdf", "col")
    queue.update_status(j1.id, "completed")
    queue.update_status(j2.id, "failed", error="err")
    # j3 stays pending
    queue.clear_completed()
    remaining = queue.list_jobs()
    ids = {j.id for j in remaining}
    assert j1.id not in ids
    assert j2.id not in ids
    assert j3.id in ids


def test_clear_completed_no_completed_does_nothing(queue):
    j = queue.enqueue("/tmp/a.pdf", "col")
    queue.clear_completed()
    assert len(queue.list_jobs()) == 1


# ---------------------------------------------------------------------------
# pending_count
# ---------------------------------------------------------------------------

def test_pending_count_zero_when_empty(queue):
    assert queue.pending_count() == 0


def test_pending_count_after_enqueue(queue):
    queue.enqueue("/tmp/a.pdf", "col")
    queue.enqueue("/tmp/b.pdf", "col")
    assert queue.pending_count() == 2


def test_pending_count_decreases_after_dequeue(queue):
    queue.enqueue("/tmp/a.pdf", "col")
    queue.enqueue("/tmp/b.pdf", "col")
    queue.dequeue()
    assert queue.pending_count() == 1


# ---------------------------------------------------------------------------
# ConnectionError propagation
# ---------------------------------------------------------------------------

def test_enqueue_propagates_connection_error():
    """If Redis is unreachable, enqueue should raise ConnectionError."""
    q = IngestionQueue.__new__(IngestionQueue)
    broken = MagicMock()
    broken.lpush.side_effect = RedisConnectionError("Connection refused")
    q.redis = broken
    with pytest.raises(RedisConnectionError):
        q.enqueue("/tmp/doc.pdf", "col")
