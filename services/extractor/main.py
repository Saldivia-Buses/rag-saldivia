"""SDA Extractor Service — document destruction pipeline.

Subscribes to NATS for extraction jobs. For each document:
1. Downloads PDF from MinIO
2. OCR via PaddleOCR-VL (SGLang)
3. Image extraction via pymupdf + analysis via Qwen3.5-9B (SGLang)
4. Publishes ExtractionResult back via NATS

No GPU loaded — all model inference goes through SGLang HTTP API.
"""

import asyncio
import json
import logging
import os
import signal
import sys

import nats

from extractor.ocr import OCRClient
from extractor.pipeline import ExtractionPipeline
from extractor.schema import ExtractionJob
from extractor.storage import StorageClient
from extractor.vision import VisionClient

logging.basicConfig(
    level=logging.INFO,
    format='{"time":"%(asctime)s","level":"%(levelname)s","msg":"%(message)s"}',
    stream=sys.stdout,
)
logger = logging.getLogger("extractor")


def env(key: str, default: str = "") -> str:
    return os.environ.get(key, default)


async def main():
    nats_url = env("NATS_URL", "nats://localhost:4222")
    sglang_ocr_url = env("SGLANG_OCR_URL", "http://localhost:8100")
    sglang_vision_url = env("SGLANG_VISION_URL", "http://localhost:8101")
    storage_endpoint = env("STORAGE_ENDPOINT", "http://localhost:9000")
    storage_bucket = env("STORAGE_BUCKET", "sda-documents")
    storage_access_key = env("STORAGE_ACCESS_KEY", "sda-admin")
    storage_secret_key = env("STORAGE_SECRET_KEY", "sda-dev-secret")

    ocr = OCRClient(sglang_ocr_url)
    vision = VisionClient(sglang_vision_url)
    storage = StorageClient(storage_endpoint, storage_bucket, storage_access_key, storage_secret_key)
    pipeline = ExtractionPipeline(ocr, vision, storage)

    nc = await nats.connect(nats_url)
    js = nc.jetstream()
    logger.info("connected to nats=%s", nats_url)

    # Ensure stream exists
    try:
        await js.add_stream(name="EXTRACTOR", subjects=["extractor.>"])
    except nats.errors.Error:
        pass  # stream already exists

    async def handle_job(msg):
        try:
            job = ExtractionJob.model_validate_json(msg.data)
            logger.info("received job document=%s tenant=%s", job.document_id, job.tenant_slug)

            # Run extraction (sync — pipeline uses sync HTTP clients)
            result = await asyncio.to_thread(pipeline.extract, job)

            # Publish result
            result_subject = f"extractor.result.{job.tenant_slug}.{job.document_id}"
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

    # Subscribe to extraction jobs for all tenants
    await js.subscribe(
        "extractor.job.*",
        cb=handle_job,
        durable="extractor-worker",
        manual_ack=True,
    )
    logger.info("subscribed to extractor.job.* — waiting for jobs")

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
