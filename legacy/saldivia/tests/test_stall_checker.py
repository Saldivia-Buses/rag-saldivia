# saldivia/tests/test_stall_checker.py
"""Tests for the background stall checker."""
import pytest
from unittest.mock import patch, MagicMock, AsyncMock
from datetime import datetime, timedelta


@pytest.mark.asyncio
async def test_stall_checker_marks_stalled_job():
    """Job sin progreso por más de deadlock_threshold es marcado stalled."""
    from saldivia.gateway import _run_stall_check

    stale_time = (datetime.now() - timedelta(seconds=200)).isoformat()
    job = {
        "id": "job-1", "task_id": "t1", "filename": "doc.pdf",
        "collection": "col", "tier": "tiny", "page_count": 5,
        "user_id": 1, "file_hash": "h", "state": "running",
        "progress": 30, "retry_count": 0, "last_checked": stale_time,
    }

    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_http_cls:
        mock_db.get_all_active_ingestion_jobs.return_value = [job]
        mock_http = AsyncMock()
        mock_http.__aenter__ = AsyncMock(return_value=mock_http)
        mock_http.__aexit__ = AsyncMock(return_value=None)
        mock_http.get.return_value = MagicMock(
            status_code=200, json=lambda: {"state": "PENDING"}
        )
        mock_http_cls.return_value = mock_http

        await _run_stall_check(ingestion_cfg={
            "server_max_retries": 3,
            "tiers": {"tiny": {"deadlock_threshold": 30, "timeout": 300}},
        })

    mock_db.increment_ingestion_retry.assert_called_once_with("job-1")


@pytest.mark.asyncio
async def test_stall_checker_creates_alert_on_max_retries():
    """Job con retry_count >= server_max_retries crea alerta y marca failed."""
    from saldivia.gateway import _run_stall_check

    stale_time = (datetime.now() - timedelta(seconds=200)).isoformat()
    job = {
        "id": "job-1", "task_id": "t1", "filename": "doc.pdf",
        "collection": "col", "tier": "tiny", "page_count": 5,
        "user_id": 1, "file_hash": "h", "state": "running",
        "progress": 30, "retry_count": 3, "last_checked": stale_time,
    }

    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_http_cls, \
         patch("saldivia.gateway._cleanup_ingest_cache") as mock_cleanup:
        mock_db.get_all_active_ingestion_jobs.return_value = [job]
        mock_http = AsyncMock()
        mock_http.__aenter__ = AsyncMock(return_value=mock_http)
        mock_http.__aexit__ = AsyncMock(return_value=None)
        mock_http.get.return_value = MagicMock(
            status_code=200, json=lambda: {"state": "PENDING"}
        )
        mock_http_cls.return_value = mock_http

        await _run_stall_check(ingestion_cfg={
            "server_max_retries": 3,
            "tiers": {"tiny": {"deadlock_threshold": 30, "timeout": 300}},
        })

    mock_db.create_ingestion_alert.assert_called_once()
    mock_db.update_ingestion_job.assert_called()
    mock_cleanup.assert_called_once_with("job-1")


@pytest.mark.asyncio
async def test_stall_checker_skips_fresh_jobs():
    """Job reciente (no stalled) no genera retry."""
    from saldivia.gateway import _run_stall_check

    fresh_time = datetime.now().isoformat()
    job = {
        "id": "job-2", "task_id": "t2", "filename": "fresh.pdf",
        "collection": "col", "tier": "small", "page_count": 10,
        "user_id": 1, "file_hash": None, "state": "running",
        "progress": 50, "retry_count": 0, "last_checked": fresh_time,
    }

    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_http_cls:
        mock_db.get_all_active_ingestion_jobs.return_value = [job]
        mock_http_cls.return_value = AsyncMock()

        await _run_stall_check(ingestion_cfg={
            "server_max_retries": 3,
            "tiers": {"small": {"deadlock_threshold": 60, "timeout": 900}},
        })

    mock_db.increment_ingestion_retry.assert_not_called()
    mock_db.create_ingestion_alert.assert_not_called()


@pytest.mark.asyncio
async def test_stall_checker_completes_finished_job():
    """Job marcado FINISHED por el ingestor se marca completed y limpia cache."""
    from saldivia.gateway import _run_stall_check

    stale_time = (datetime.now() - timedelta(seconds=200)).isoformat()
    job = {
        "id": "job-3", "task_id": "t3", "filename": "done.pdf",
        "collection": "col", "tier": "tiny", "page_count": 5,
        "user_id": 1, "file_hash": None, "state": "running",
        "progress": 90, "retry_count": 0, "last_checked": stale_time,
    }

    with patch("saldivia.gateway.db") as mock_db, \
         patch("saldivia.gateway.httpx.AsyncClient") as mock_http_cls, \
         patch("saldivia.gateway._cleanup_ingest_cache") as mock_cleanup:
        mock_db.get_all_active_ingestion_jobs.return_value = [job]
        mock_http = AsyncMock()
        mock_http.__aenter__ = AsyncMock(return_value=mock_http)
        mock_http.__aexit__ = AsyncMock(return_value=None)
        mock_http.get.return_value = MagicMock(
            status_code=200, json=lambda: {"state": "FINISHED"}
        )
        mock_http_cls.return_value = mock_http

        await _run_stall_check(ingestion_cfg={
            "server_max_retries": 3,
            "tiers": {"tiny": {"deadlock_threshold": 30, "timeout": 300}},
        })

    mock_db.update_ingestion_job.assert_called_once()
    call_args = mock_db.update_ingestion_job.call_args
    assert call_args[0][0] == "job-3"
    assert call_args[0][1] == "completed"
    mock_cleanup.assert_called_once_with("job-3")
