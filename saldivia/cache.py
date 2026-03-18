# saldivia/cache.py
"""Query result caching."""
import hashlib
import json
import redis
from typing import Optional
from dataclasses import dataclass


@dataclass
class CacheConfig:
    enabled: bool = True
    ttl_seconds: int = 3600
    max_entries: int = 1000


class QueryCache:
    """Redis-backed query cache."""

    PREFIX = "rag_cache:"

    def __init__(self, redis_url: str = "redis://localhost:6379", config: CacheConfig = None):
        self.redis = redis.from_url(redis_url)
        self.config = config or CacheConfig()

    def _key(self, query: str, collection: str) -> str:
        """Generate cache key."""
        content = f"{query}:{collection}"
        hash_val = hashlib.md5(content.encode()).hexdigest()[:16]
        return f"{self.PREFIX}{collection}:{hash_val}"

    def get(self, query: str, collection: str) -> Optional[str]:
        """Get cached result."""
        if not self.config.enabled:
            return None
        key = self._key(query, collection)
        result = self.redis.get(key)
        return result.decode() if result else None

    def set(self, query: str, collection: str, result: str):
        """Cache a result."""
        if not self.config.enabled:
            return
        key = self._key(query, collection)
        self.redis.setex(key, self.config.ttl_seconds, result)

    def invalidate(self, collection: str = None):
        """Invalidate cache entries, optionally scoped to a collection."""
        if collection:
            pattern = f"{self.PREFIX}{collection}:*"
        else:
            pattern = f"{self.PREFIX}*"
        for key in self.redis.scan_iter(pattern):
            self.redis.delete(key)

    def stats(self) -> dict:
        """Get cache statistics."""
        pattern = f"{self.PREFIX}*"
        count = sum(1 for _ in self.redis.scan_iter(pattern))
        return {"entries": count, "enabled": self.config.enabled}
