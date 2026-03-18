# saldivia/collections.py
"""Collection management for RAG Saldivia."""
import os
import httpx
from dataclasses import dataclass
from typing import Optional
from pymilvus import connections, utility, Collection


@dataclass
class CollectionStats:
    name: str
    entity_count: int
    index_type: str
    has_sparse: bool


class CollectionManager:
    """Manages Milvus collections via ingestor API and direct connection."""

    def __init__(
        self,
        ingestor_url: str = None,
        milvus_host: str = None,
        milvus_port: int = 19530,
    ):
        self.ingestor_url = ingestor_url or os.getenv("INGESTOR_URL", "http://localhost:8082")
        self.milvus_host = milvus_host or os.getenv("MILVUS_HOST", "localhost")
        self.milvus_port = milvus_port
        self._connected = False

    def _connect_milvus(self):
        """Connect to Milvus if not connected."""
        if not self._connected:
            connections.connect(host=self.milvus_host, port=self.milvus_port)
            self._connected = True

    def list(self) -> list[str]:
        """List all collections."""
        self._connect_milvus()
        return utility.list_collections()

    def create(self, name: str, schema: str = "hybrid") -> bool:
        """Create a new collection via ingestor API."""
        with httpx.Client(timeout=30) as client:
            resp = client.post(
                f"{self.ingestor_url}/v1/collections",
                json={"collection_name": name, "schema_type": schema}
            )
            return resp.status_code == 200

    def delete(self, name: str) -> bool:
        """Delete a collection."""
        self._connect_milvus()
        if name in self.list():
            utility.drop_collection(name)
            return True
        return False

    def stats(self, name: str) -> Optional[CollectionStats]:
        """Get collection statistics."""
        self._connect_milvus()
        if name not in self.list():
            return None

        col = Collection(name)
        col.load()

        # Check for sparse field
        has_sparse = any(f.name == "sparse" for f in col.schema.fields)

        return CollectionStats(
            name=name,
            entity_count=col.num_entities,
            index_type=col.indexes[0].params.get("index_type", "unknown") if col.indexes else "none",
            has_sparse=has_sparse,
        )

    def health(self) -> dict:
        """Check Milvus health."""
        try:
            self._connect_milvus()
            return {
                "status": "healthy",
                "collections": len(self.list()),
            }
        except Exception as e:
            return {"status": "unhealthy", "error": str(e)}
