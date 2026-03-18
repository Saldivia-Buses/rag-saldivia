# saldivia/gateway.py
"""Auth Gateway - FastAPI middleware for RAG API."""
import os
import json
import hashlib
import logging
from typing import Optional
from fastapi import FastAPI, Request, HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from starlette.responses import StreamingResponse
import httpx

from saldivia.auth import AuthDB, User, Role, Permission

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="RAG Saldivia Gateway")

# Configuration
RAG_SERVER_URL = os.getenv("RAG_SERVER_URL", "http://localhost:8081")
INGESTOR_URL = os.getenv("INGESTOR_URL", "http://localhost:8082")
BYPASS_AUTH = os.getenv("BYPASS_AUTH", "false").lower() == "true"


@app.on_event("startup")
async def on_startup():
    """Validate configuration at startup."""
    env = os.getenv("ENVIRONMENT", "production")
    if BYPASS_AUTH and env == "production":
        raise RuntimeError(
            "BYPASS_AUTH cannot be true in production. "
            "Set ENVIRONMENT=development to allow bypass (dev/test only)."
        )
    if BYPASS_AUTH:
        logger.warning("⚠️  BYPASS_AUTH is enabled — authentication bypassed (dev mode only)")

security = HTTPBearer(auto_error=False)
db = AuthDB()


def get_user_from_token(credentials: HTTPAuthorizationCredentials = Depends(security)) -> Optional[User]:
    """Extract and validate user from Bearer token."""
    if BYPASS_AUTH:
        return None  # Allow all requests in dev mode

    if not credentials:
        raise HTTPException(status_code=401, detail="Missing API key")

    api_key = credentials.credentials
    api_key_hash = hashlib.sha256(api_key.encode()).hexdigest()

    user = db.get_user_by_api_key_hash(api_key_hash)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid API key")

    return user


def filter_collections(user: User, requested: list[str]) -> list[str]:
    """Filter collections to only those the user can access."""
    if user is None or user.role == Role.ADMIN:
        return requested

    allowed = set(db.get_user_collections(user))
    filtered = [c for c in requested if c in allowed]

    if not filtered:
        raise HTTPException(
            status_code=403,
            detail=f"No access to requested collections. You have access to: {list(allowed)}"
        )

    return filtered


@app.post("/v1/generate")
async def generate(request: Request, user: User = Depends(get_user_from_token)):
    """Proxy to RAG generate endpoint with auth filtering. Streams SSE response."""
    body = await request.json()

    # Filter collections
    if "collection_names" in body:
        body["collection_names"] = filter_collections(user, body["collection_names"])

    # Log query
    if user:
        query_preview = ""
        if "messages" in body and body["messages"]:
            last_msg = body["messages"][-1].get("content", "")
            query_preview = last_msg[:100] if isinstance(last_msg, str) else str(last_msg)[:100]

        db.log_action(
            user_id=user.id,
            action="query",
            collection=",".join(body.get("collection_names", [])),
            query_preview=query_preview,
            ip_address=request.client.host if request.client else ""
        )

    # Proxy request, streaming SSE response.
    # Use send(stream=True) to check the upstream status code before committing
    # to StreamingResponse (which always sends HTTP 200 to the client once started).
    client = httpx.AsyncClient(timeout=120)
    req = client.build_request(
        "POST", f"{RAG_SERVER_URL}/v1/generate",
        json=body,
        headers={"Content-Type": "application/json"}
    )
    resp = await client.send(req, stream=True)

    if resp.status_code >= 400:
        error_body = await resp.aread()
        await resp.aclose()
        await client.aclose()
        raise HTTPException(status_code=resp.status_code, detail=error_body.decode())

    async def _stream():
        try:
            async for chunk in resp.aiter_bytes():
                yield chunk
        finally:
            await resp.aclose()
            await client.aclose()

    return StreamingResponse(_stream(), media_type="text/event-stream")


@app.post("/v1/search")
async def search(request: Request, user: User = Depends(get_user_from_token)):
    """Proxy to RAG search endpoint with auth filtering."""
    body = await request.json()

    if "collection_names" in body:
        body["collection_names"] = filter_collections(user, body["collection_names"])

    if user:
        db.log_action(
            user_id=user.id,
            action="search",
            collection=",".join(body.get("collection_names", [])),
            query_preview=body.get("query", "")[:100],
            ip_address=request.client.host if request.client else ""
        )

    async with httpx.AsyncClient(timeout=60) as client:
        resp = await client.post(
            f"{RAG_SERVER_URL}/v1/search",
            json=body,
            headers={"Content-Type": "application/json"}
        )
        return resp.json()


@app.post("/v1/documents")
async def ingest(request: Request, user: User = Depends(get_user_from_token)):
    """Proxy to ingestor with write permission check."""
    if user and user.role == Role.USER:
        raise HTTPException(status_code=403, detail="Users cannot ingest documents directly")

    if user and user.role == Role.AREA_MANAGER:
        # Parse the target collection from the multipart 'data' JSON field.
        # Check that the user has write access to THAT specific collection,
        # not just any collection in their area.
        form = await request.form()
        data_str = form.get("data", "{}")
        try:
            data = json.loads(data_str)
            collection_name = data.get("collection_name", "")
            if collection_name and not db.can_access(user, collection_name, Permission.WRITE):
                raise HTTPException(
                    status_code=403,
                    detail=f"No write access to collection: {collection_name}"
                )
        except (json.JSONDecodeError, KeyError):
            pass  # If we can't parse, let the ingestor handle it

    # Forward multipart request as-is.
    # Note: request.form() above caches the body; request.body() returns the same bytes.
    body = await request.body()
    headers = dict(request.headers)
    headers.pop("host", None)

    async with httpx.AsyncClient(timeout=600) as client:
        resp = await client.post(
            f"{INGESTOR_URL}/v1/documents",
            content=body,
            headers=headers
        )

        if user:
            db.log_action(
                user_id=user.id,
                action="ingest",
                ip_address=request.client.host if request.client else ""
            )

        return resp.json()


@app.get("/health")
async def health():
    return {"status": "ok", "auth_enabled": not BYPASS_AUTH}


@app.get("/v1/collections")
async def list_collections(user: User = Depends(get_user_from_token)):
    """List collections user can access."""
    if user is None:
        from saldivia.collections import CollectionManager
        return {"collections": CollectionManager().list()}

    return {"collections": db.get_user_collections(user)}


# Admin endpoints
@app.get("/admin/audit")
async def get_audit(limit: int = 100, user: User = Depends(get_user_from_token)):
    """Get audit log (admin only)."""
    if user and user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin only")

    entries = db.get_audit_log(limit=limit)
    return {"entries": [
        {
            "id": e.id,
            "user_id": e.user_id,
            "action": e.action,
            "collection": e.collection,
            "timestamp": e.timestamp.isoformat() if e.timestamp else None
        }
        for e in entries
    ]}


def main():
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8090)


if __name__ == "__main__":
    main()
