"""SDA Extractor Service — document destruction pipeline.

Subscribes to NATS for extraction jobs. For each document:
1. Downloads PDF from MinIO
2. OCR via PaddleOCR-VL (SGLang)
3. Image extraction via pymupdf + analysis via Qwen3.5-9B (SGLang)
4. Publishes ExtractionResult back via NATS

No GPU loaded — all model inference goes through SGLang HTTP API.
"""

import asyncio
import logging
import os
import re
import signal
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
from threading import Thread

import nats
from nats.js.api import ConsumerConfig

from extractor.ocr import OCRClient
from extractor.pipeline import ExtractionPipeline
from extractor.schema import ExtractionJob
from extractor.storage import StorageClient
from extractor.vision import VisionClient

from pythonjsonlogger.json import JsonFormatter

_handler = logging.StreamHandler(sys.stdout)
_handler.setFormatter(JsonFormatter(
    fmt="%(asctime)s %(levelname)s %(name)s %(message)s",
    rename_fields={"asctime": "time", "levelname": "level", "name": "logger"},
))
logging.basicConfig(level=logging.INFO, handlers=[_handler])
logger = logging.getLogger("extractor")

_SAFE_SUBJECT_RE = re.compile(r"^[a-zA-Z0-9_-]+$")


def _require_env(key: str) -> str:
    val = os.environ.get(key, "")
    if not val:
        logger.error("required env var %s is not set", key)
        sys.exit(1)
    return val


def _env(key: str, default: str = "") -> str:
    return os.environ.get(key, default)


class _HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b'{"status":"ok"}')
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        pass  # suppress access logs


def _start_health_server(port: int) -> Thread:
    server = HTTPServer(("0.0.0.0", port), _HealthHandler)
    t = Thread(target=server.serve_forever, daemon=True)
    t.start()
    return t


def _validate_subject_token(value: str, field: str) -> None:
    """Reject NATS subject injection (dots, wildcards, whitespace)."""
    if not _SAFE_SUBJECT_RE.match(value):
        raise ValueError(f"invalid NATS subject token in {field}: {value!r}")


async def main():
    nats_url = _env("NATS_URL", "nats://localhost:4222")
    sglang_ocr_url = _env("SGLANG_OCR_URL", "http://localhost:8100")
    sglang_vision_url = _env("SGLANG_VISION_URL", "http://localhost:8101")

    # Storage: fail fast if not configured (no hardcoded secrets)
    storage_endpoint = _require_env("STORAGE_ENDPOINT")
    storage_bucket = _require_env("STORAGE_BUCKET")
    storage_access_key = _require_env("STORAGE_ACCESS_KEY")
    storage_secret_key = _require_env("STORAGE_SECRET_KEY")

    # D4: Health check endpoint for Docker/k8s
    health_port = int(_env("HEALTH_PORT", "8090"))
    health_thread = _start_health_server(health_port)
    logger.info("health endpoint on :%d/health", health_port)

    ocr = OCRClient(sglang_ocr_url)
    vision = VisionClient(sglang_vision_url)
    storage = StorageClient(storage_endpoint, storage_bucket, storage_access_key, storage_secret_key)
    pipeline = ExtractionPipeline(ocr, vision, storage)

    nc = await nats.connect(nats_url)
    js = nc.jetstream()
    logger.info("connected to nats=%s", nats_url)

    # Ensure stream exists — tenant.*.extractor.> convention
    try:
        await js.add_stream(name="EXTRACTOR", subjects=["tenant.*.extractor.>"])
    except nats.errors.Error:
        pass  # stream already exists

    async def handle_job(msg):
        try:
            job = ExtractionJob.model_validate_json(msg.data)

            # B1: validate subject tokens to prevent injection
            _validate_subject_token(job.tenant_slug, "tenant_slug")
            _validate_subject_token(job.document_id, "document_id")

            logger.info("received job document=%s tenant=%s", job.document_id, job.tenant_slug)

            # Run extraction (sync — pipeline uses sync HTTP clients)
            result = await asyncio.to_thread(pipeline.extract, job)

            # B2: use tenant-prefixed subject convention
            result_subject = f"tenant.{job.tenant_slug}.extractor.result.{job.document_id}"
            await nc.publish(result_subject, result.model_dump_json().encode())

            await msg.ack()
            logger.info(
                "extraction complete document=%s pages=%d images=%d time_ms=%d",
                job.document_id,
                result.total_pages,
                sum(len(p.images) for p in result.pages),
                result.metadata.extraction_time_ms,
            )
        except Exception:
            logger.exception("extraction failed for message")
            await msg.nak()

    # B3: cap retries at 3 to match Go services
    config = ConsumerConfig(
        durable_name="extractor-worker",
        max_deliver=3,
        ack_wait=300,  # 5 min for large PDFs
    )
    await js.subscribe(
        "tenant.*.extractor.job",
        cb=handle_job,
        config=config,
        manual_ack=True,
    )
    logger.info("subscribed to tenant.*.extractor.job — waiting for jobs")

    # Graceful shutdown
    stop = asyncio.Event()
    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, stop.set)

    await stop.wait()
    logger.info("shutting down")
    await nc.drain()
    ocr.close()
    vision.close()


if __name__ == "__main__":
    asyncio.run(main())
