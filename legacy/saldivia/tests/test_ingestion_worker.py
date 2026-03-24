# saldivia/tests/test_ingestion_worker.py
"""Tests for ingestion_worker: process_job, process_job_with_retry, run_worker."""
import pytest
import fakeredis
from pathlib import Path
from unittest.mock import MagicMock, patch, call
from redis.exceptions import ConnectionError as RedisConnectionError

from saldivia.ingestion_queue import IngestionJob, IngestionQueue
import saldivia.ingestion_worker as worker_module
from saldivia.ingestion_worker import (
    process_job,
    process_job_with_retry,
    run_worker,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def make_job(file_path: str = "/tmp/doc.pdf", collection: str = "test-col") -> IngestionJob:
    """Create a minimal IngestionJob for testing."""
    return IngestionJob(
        id="test-job-id",
        file_path=file_path,
        collection=collection,
        status="pending",
        created_at="2026-03-23T00:00:00",
    )


# ---------------------------------------------------------------------------
# process_job — file not found
# ---------------------------------------------------------------------------

def test_process_job_returns_false_if_file_not_found():
    job = make_job(file_path="/nonexistent/path/doc.pdf")
    result = process_job(job)
    assert result is False


# ---------------------------------------------------------------------------
# process_job — ingestor success
# ---------------------------------------------------------------------------

def test_process_job_returns_true_on_success(tmp_path):
    doc = tmp_path / "test.pdf"
    doc.write_bytes(b"%PDF fake content")
    job = make_job(file_path=str(doc))

    mock_response = MagicMock()
    mock_response.raise_for_status.return_value = None

    mock_client_instance = MagicMock()
    mock_client_instance.post.return_value = mock_response
    mock_client_instance.__enter__ = MagicMock(return_value=mock_client_instance)
    mock_client_instance.__exit__ = MagicMock(return_value=False)

    with patch("saldivia.ingestion_worker.httpx.Client", return_value=mock_client_instance):
        result = process_job(job)

    assert result is True
    mock_client_instance.post.assert_called_once()


# ---------------------------------------------------------------------------
# process_job — ingestor raises exception
# ---------------------------------------------------------------------------

def test_process_job_returns_false_on_http_exception(tmp_path):
    doc = tmp_path / "test.pdf"
    doc.write_bytes(b"%PDF fake content")
    job = make_job(file_path=str(doc))

    mock_client_instance = MagicMock()
    mock_client_instance.post.side_effect = Exception("Connection refused")
    mock_client_instance.__enter__ = MagicMock(return_value=mock_client_instance)
    mock_client_instance.__exit__ = MagicMock(return_value=False)

    with patch("saldivia.ingestion_worker.httpx.Client", return_value=mock_client_instance):
        result = process_job(job)

    assert result is False


def test_process_job_returns_false_on_http_status_error(tmp_path):
    import httpx
    doc = tmp_path / "test.pdf"
    doc.write_bytes(b"%PDF fake content")
    job = make_job(file_path=str(doc))

    mock_response = MagicMock()
    mock_response.raise_for_status.side_effect = httpx.HTTPStatusError(
        "500", request=MagicMock(), response=MagicMock()
    )

    mock_client_instance = MagicMock()
    mock_client_instance.post.return_value = mock_response
    mock_client_instance.__enter__ = MagicMock(return_value=mock_client_instance)
    mock_client_instance.__exit__ = MagicMock(return_value=False)

    with patch("saldivia.ingestion_worker.httpx.Client", return_value=mock_client_instance):
        result = process_job(job)

    assert result is False


# ---------------------------------------------------------------------------
# process_job_with_retry — success on first attempt
# ---------------------------------------------------------------------------

def test_process_job_with_retry_success_first_attempt():
    job = make_job()
    with patch("saldivia.ingestion_worker.process_job", return_value=True) as mock_pj:
        result = process_job_with_retry(job)
    assert result is True
    mock_pj.assert_called_once_with(job)


# ---------------------------------------------------------------------------
# process_job_with_retry — exhausts all retries
# ---------------------------------------------------------------------------

def test_process_job_with_retry_exhausts_all_retries():
    job = make_job()
    with patch("saldivia.ingestion_worker.process_job", return_value=False) as mock_pj, \
         patch("saldivia.ingestion_worker.time.sleep") as mock_sleep:
        result = process_job_with_retry(job)

    assert result is False
    # process_job should be called MAX_RETRIES (3) times
    assert mock_pj.call_count == worker_module.MAX_RETRIES
    # sleep should be called MAX_RETRIES - 1 times (between attempts)
    assert mock_sleep.call_count == worker_module.MAX_RETRIES - 1


def test_process_job_with_retry_sleep_uses_retry_delays():
    job = make_job()
    with patch("saldivia.ingestion_worker.process_job", return_value=False), \
         patch("saldivia.ingestion_worker.time.sleep") as mock_sleep:
        process_job_with_retry(job)

    sleep_calls = [c.args[0] for c in mock_sleep.call_args_list]
    # Should use RETRY_DELAYS values for first MAX_RETRIES-1 calls
    assert sleep_calls == worker_module.RETRY_DELAYS[:worker_module.MAX_RETRIES - 1]


# ---------------------------------------------------------------------------
# process_job_with_retry — success on second attempt
# ---------------------------------------------------------------------------

def test_process_job_with_retry_success_on_second_attempt():
    job = make_job()
    # First call fails, second succeeds
    side_effects = [False, True]
    with patch("saldivia.ingestion_worker.process_job", side_effect=side_effects) as mock_pj, \
         patch("saldivia.ingestion_worker.time.sleep"):
        result = process_job_with_retry(job)

    assert result is True
    assert mock_pj.call_count == 2


# ---------------------------------------------------------------------------
# run_worker — clean shutdown when _shutdown is set
# ---------------------------------------------------------------------------

def test_run_worker_clean_shutdown():
    """Worker should exit the loop when _shutdown becomes True after one iteration."""
    fake_redis = fakeredis.FakeRedis(decode_responses=False)
    fake_queue = IngestionQueue.__new__(IngestionQueue)
    fake_queue.redis = fake_redis

    call_count = 0

    def fake_dequeue():
        nonlocal call_count
        call_count += 1
        # Set shutdown after first dequeue call so the loop exits
        worker_module._shutdown = True
        return None  # no job, just wake up and check shutdown

    fake_queue.dequeue = fake_dequeue
    fake_queue.update_status = MagicMock()
    fake_queue.close = MagicMock()

    with patch("saldivia.ingestion_worker.IngestionQueue", return_value=fake_queue), \
         patch("saldivia.ingestion_worker.time.sleep"), \
         patch("saldivia.ingestion_worker.signal.signal"):
        run_worker(redis_url="redis://localhost:6379")

    # close() must always be called on shutdown
    fake_queue.close.assert_called_once()


# ---------------------------------------------------------------------------
# run_worker — Redis backoff on ConnectionError
# ---------------------------------------------------------------------------

def test_run_worker_redis_backoff():
    """Worker should sleep and double the delay on repeated Redis ConnectionErrors."""
    fake_redis = fakeredis.FakeRedis(decode_responses=False)
    fake_queue = IngestionQueue.__new__(IngestionQueue)
    fake_queue.redis = fake_redis

    error_count = 0
    MAX_ERRORS = 3

    def fake_dequeue():
        nonlocal error_count
        error_count += 1
        if error_count <= MAX_ERRORS:
            raise RedisConnectionError("Redis down")
        # After enough errors, trigger shutdown so test ends
        worker_module._shutdown = True
        return None

    fake_queue.dequeue = fake_dequeue
    fake_queue.update_status = MagicMock()
    fake_queue.close = MagicMock()

    sleep_calls = []

    def fake_sleep(seconds):
        sleep_calls.append(seconds)

    with patch("saldivia.ingestion_worker.IngestionQueue", return_value=fake_queue), \
         patch("saldivia.ingestion_worker.time.sleep", side_effect=fake_sleep), \
         patch("saldivia.ingestion_worker.signal.signal"):
        run_worker(redis_url="redis://localhost:6379")

    # Should have slept at least MAX_ERRORS times (once per ConnectionError).
    # The 4th dequeue sets _shutdown=True and returns None, which also triggers
    # the idle sleep(5) before the loop condition re-checks _shutdown.
    assert len(sleep_calls) >= MAX_ERRORS

    # Backoff: 5, 10, 20 (doubles each time, capped at REDIS_MAX_DELAY=60)
    expected_base = worker_module.REDIS_RETRY_DELAY
    assert sleep_calls[0] == expected_base
    assert sleep_calls[1] == expected_base * 2
    assert sleep_calls[2] == expected_base * 4

    fake_queue.close.assert_called_once()


def test_run_worker_redis_delay_resets_on_success():
    """Backoff delay should reset to REDIS_RETRY_DELAY after a successful dequeue."""
    fake_redis = fakeredis.FakeRedis(decode_responses=False)
    fake_queue = IngestionQueue.__new__(IngestionQueue)
    fake_queue.redis = fake_redis

    sequence = iter([
        RedisConnectionError("down"),  # error → sleep(5), delay becomes 10
        RedisConnectionError("down"),  # error → sleep(10), delay becomes 20
        None,                          # success (no job) → delay resets to 5
        RedisConnectionError("down"),  # error → sleep(5) again (reset worked)
        "SHUTDOWN",
    ])

    def fake_dequeue():
        val = next(sequence)
        if val == "SHUTDOWN":
            worker_module._shutdown = True
            return None
        if isinstance(val, Exception):
            raise val
        return val  # None = no job

    fake_queue.dequeue = fake_dequeue
    fake_queue.update_status = MagicMock()
    fake_queue.close = MagicMock()

    sleep_calls = []

    def fake_sleep(seconds):
        sleep_calls.append(seconds)

    with patch("saldivia.ingestion_worker.IngestionQueue", return_value=fake_queue), \
         patch("saldivia.ingestion_worker.time.sleep", side_effect=fake_sleep), \
         patch("saldivia.ingestion_worker.signal.signal"):
        run_worker(redis_url="redis://localhost:6379")

    # sleep calls: error1(5), error2(10), no-job(5), error3(5)
    assert sleep_calls[0] == worker_module.REDIS_RETRY_DELAY      # first error
    assert sleep_calls[1] == worker_module.REDIS_RETRY_DELAY * 2  # second error
    assert sleep_calls[2] == 5                                     # no job idle sleep
    assert sleep_calls[3] == worker_module.REDIS_RETRY_DELAY      # third error (reset)
