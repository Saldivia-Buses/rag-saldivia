# saldivia/watch.py
"""Watch folder for automatic ingestion."""
import os
import time
import logging
from pathlib import Path
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler, FileCreatedEvent
from saldivia.ingestion_queue import IngestionQueue

logger = logging.getLogger(__name__)


class IngestionHandler(FileSystemEventHandler):
    """Handler for new files in watch folder."""

    SUPPORTED_EXTENSIONS = {".pdf", ".docx", ".txt", ".md"}

    def __init__(self, queue: IngestionQueue, collection: str):
        self.queue = queue
        self.collection = collection
        self._processed = set()

    def on_created(self, event: FileCreatedEvent):
        if event.is_directory:
            return

        path = Path(event.src_path)
        if path.suffix.lower() not in self.SUPPORTED_EXTENSIONS:
            return

        # Avoid duplicates
        if str(path) in self._processed:
            return
        self._processed.add(str(path))

        # Wait for file to be fully written
        time.sleep(1)

        logger.info(f"New file detected: {path}")
        job = self.queue.enqueue(str(path), self.collection)
        logger.info(f"Queued job {job.id} for {path.name}")


def start_watcher(
    watch_dir: str,
    collection: str,
    redis_url: str = "redis://localhost:6379"
):
    """Start watching a directory for new files."""
    queue = IngestionQueue(redis_url)
    handler = IngestionHandler(queue, collection)

    observer = Observer()
    observer.schedule(handler, watch_dir, recursive=True)
    observer.start()

    logger.info(f"Watching {watch_dir} for new files -> collection '{collection}'")

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()


if __name__ == "__main__":
    import sys
    logging.basicConfig(level=logging.INFO)
    if len(sys.argv) < 3:
        print("Usage: python -m saldivia.watch <directory> <collection>")
        sys.exit(1)
    start_watcher(sys.argv[1], sys.argv[2])
