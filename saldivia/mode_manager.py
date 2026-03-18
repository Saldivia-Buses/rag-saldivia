# saldivia/mode_manager.py
"""1-GPU Mode Manager for dynamic model loading."""
import enum
import subprocess
import time
import logging
from dataclasses import dataclass
from typing import Optional

logger = logging.getLogger(__name__)


class Mode(enum.Enum):
    QUERY = "query"      # NIMs loaded, VLM unloaded
    INGEST = "ingest"    # NIMs + VLM loaded
    TRANSITION = "transition"


# VRAM requirements in GB (from Brev measurements)
MEMORY_REQUIREMENTS = {
    Mode.QUERY: 46,    # Triton NIMs only
    Mode.INGEST: 90,   # NIMs (46) + VLM (44)
}

# Container names
VLM_CONTAINER = "qwen3-vl-8b"
NIMS_CONTAINERS = [
    "nemotron-embedding-ms",
    "nemotron-ranking-ms",
    "compose-nv-ingest-ms-runtime-1",
]


@dataclass
class ModeStatus:
    mode: Mode
    vlm_loaded: bool
    nims_loaded: bool
    gpu_memory_used_gb: float
    pending_ingestion_jobs: int


class ModeManager:
    """Manages GPU memory by loading/unloading models based on workload."""

    def __init__(self, gpu_memory_gb: float = 98):
        self.gpu_memory_gb = gpu_memory_gb
        self.current_mode = Mode.QUERY
        self._vlm_loaded = False

    def can_switch_to(self, target: Mode) -> bool:
        """Check if we have enough VRAM for target mode."""
        required = MEMORY_REQUIREMENTS.get(target, 0)
        return required <= self.gpu_memory_gb

    def get_status(self) -> ModeStatus:
        """Get current mode status."""
        return ModeStatus(
            mode=self.current_mode,
            vlm_loaded=self._vlm_loaded,
            nims_loaded=self._check_nims_running(),
            gpu_memory_used_gb=self._get_gpu_memory_used(),
            pending_ingestion_jobs=self._get_pending_jobs(),
        )

    def switch_to_ingest_mode(self) -> bool:
        """Load VLM for ingestion. Returns True if successful."""
        if self.current_mode == Mode.INGEST:
            return True

        if not self.can_switch_to(Mode.INGEST):
            logger.error(f"Not enough VRAM for ingest mode")
            return False

        self.current_mode = Mode.TRANSITION
        logger.info("Switching to INGEST mode - loading VLM...")

        try:
            self._start_vlm()
            self._wait_for_vlm_healthy()
            self._vlm_loaded = True
            self.current_mode = Mode.INGEST
            logger.info("INGEST mode active")
            return True
        except Exception as e:
            logger.error(f"Failed to switch to ingest mode: {e}")
            self.current_mode = Mode.QUERY
            return False

    def switch_to_query_mode(self) -> bool:
        """Unload VLM for query-only mode. Returns True if successful."""
        if self.current_mode == Mode.QUERY:
            return True

        self.current_mode = Mode.TRANSITION
        logger.info("Switching to QUERY mode - unloading VLM...")

        try:
            self._stop_vlm()
            self._vlm_loaded = False
            self.current_mode = Mode.QUERY
            logger.info("QUERY mode active")
            return True
        except Exception as e:
            logger.error(f"Failed to switch to query mode: {e}")
            self.current_mode = Mode.INGEST
            return False

    def _start_vlm(self):
        """Start VLM container."""
        subprocess.run(
            ["docker", "start", VLM_CONTAINER],
            check=True,
            capture_output=True
        )

    def _stop_vlm(self):
        """Stop VLM container."""
        subprocess.run(
            ["docker", "stop", VLM_CONTAINER],
            check=True,
            capture_output=True
        )

    def _wait_for_vlm_healthy(self, timeout: int = 120):
        """Wait for VLM to be healthy."""
        import httpx
        start = time.time()
        while time.time() - start < timeout:
            try:
                resp = httpx.get("http://localhost:8000/health", timeout=5)
                if resp.status_code == 200:
                    return
            except:
                pass
            time.sleep(3)
        raise TimeoutError("VLM failed to become healthy")

    def _check_nims_running(self) -> bool:
        """Check if NIM containers are running."""
        for container in NIMS_CONTAINERS:
            result = subprocess.run(
                ["docker", "inspect", "-f", "{{.State.Running}}", container],
                capture_output=True, text=True
            )
            if result.returncode != 0 or result.stdout.strip() != "true":
                return False
        return True

    def _get_gpu_memory_used(self) -> float:
        """Get GPU memory usage in GB."""
        try:
            result = subprocess.run(
                ["nvidia-smi", "--query-gpu=memory.used", "--format=csv,noheader,nounits"],
                capture_output=True, text=True, check=True
            )
            return float(result.stdout.strip()) / 1024
        except:
            return 0.0

    def _get_pending_jobs(self) -> int:
        """Get number of pending ingestion jobs from queue."""
        # Will be implemented with Redis queue
        return 0
