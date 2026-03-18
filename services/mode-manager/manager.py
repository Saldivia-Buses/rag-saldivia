# services/mode-manager/manager.py
"""Mode Manager Service - monitors queue and switches modes automatically."""
import os
import time
import redis
import logging
from saldivia.mode_manager import ModeManager, Mode

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")
QUEUE_NAME = "ingestion_queue"
IDLE_TIMEOUT = int(os.getenv("IDLE_TIMEOUT", "300"))  # 5 min default
CHECK_INTERVAL = int(os.getenv("CHECK_INTERVAL", "10"))  # 10 sec

def main():
    manager = ModeManager(gpu_memory_gb=float(os.getenv("GPU_MEMORY_GB", "98")))
    r = redis.from_url(REDIS_URL)

    last_job_time = time.time()
    logger.info(f"Mode Manager started. IDLE_TIMEOUT={IDLE_TIMEOUT}s")

    while True:
        try:
            queue_length = r.llen(QUEUE_NAME)

            if queue_length > 0:
                last_job_time = time.time()
                if manager.current_mode != Mode.INGEST:
                    logger.info(f"Jobs pending ({queue_length}), switching to INGEST mode")
                    manager.switch_to_ingest_mode()

            else:
                idle_time = time.time() - last_job_time
                if manager.current_mode == Mode.INGEST and idle_time > IDLE_TIMEOUT:
                    logger.info(f"Idle for {idle_time:.0f}s, switching to QUERY mode")
                    manager.switch_to_query_mode()

            # Publish status
            status = manager.get_status()
            r.set("mode_manager:status", f"{status.mode.value}")
            r.set("mode_manager:vlm_loaded", str(status.vlm_loaded).lower())

        except Exception as e:
            logger.error(f"Error in main loop: {e}")

        time.sleep(CHECK_INTERVAL)

if __name__ == "__main__":
    main()
