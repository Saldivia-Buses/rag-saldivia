# saldivia/tier.py
"""Document tier classification — shared between gateway and ingestion scripts."""
from dataclasses import dataclass


@dataclass(frozen=True)
class TierConfig:
    """Configuration for a document processing tier."""
    poll_interval: int   # seconds between status polls
    restart_after: int   # seconds before restarting a stalled job
    timeout: int         # max seconds for the whole job


TIERS: dict[str, TierConfig] = {
    "tiny":   TierConfig(poll_interval=5,  restart_after=120,  timeout=300),
    "small":  TierConfig(poll_interval=10, restart_after=300,  timeout=900),
    "medium": TierConfig(poll_interval=20, restart_after=600,  timeout=2700),
    "large":  TierConfig(poll_interval=30, restart_after=1200, timeout=7200),
}


def classify_tier(
    page_count: int | None = None,
    file_size: int = 0,
) -> str:
    """Classify a document into a processing tier.

    Uses page_count when available (more accurate for PDFs).
    Falls back to file_size for non-PDF or when page count is unknown.

    Returns:
        One of: "tiny", "small", "medium", "large"
    """
    if page_count is not None:
        if page_count <= 20:
            return "tiny"
        elif page_count <= 80:
            return "small"
        elif page_count <= 300:
            return "medium"
        else:
            return "large"

    # Fallback: classify by file size (bytes)
    if file_size < 100_000:          # < 100 KB
        return "tiny"
    elif file_size < 1_000_000:      # < 1 MB
        return "small"
    elif file_size < 10_000_000:     # < 10 MB
        return "medium"
    else:
        return "large"
