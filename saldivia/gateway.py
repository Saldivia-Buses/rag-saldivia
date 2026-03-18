# saldivia/gateway.py
"""Auth Gateway - FastAPI middleware for RAG API."""
import os
import json
import hashlib
import logging
from datetime import datetime, timedelta, timezone
from typing import Optional
from dataclasses import asdict

import jwt as pyjwt
from fastapi import FastAPI, Request, HTTPException, Depends
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel
from starlette.responses import StreamingResponse
import httpx

from saldivia.auth import AuthDB, User, Role, Permission
from saldivia.auth.models import generate_api_key, hash_password, verify_password
from saldivia.collections import CollectionManager

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="RAG Saldivia Gateway")

# Configuration
RAG_SERVER_URL = os.getenv("RAG_SERVER_URL", "http://localhost:8081")
INGESTOR_URL = os.getenv("INGESTOR_URL", "http://localhost:8082")
BYPASS_AUTH = os.getenv("BYPASS_AUTH", "false").lower() == "true"

# JWT Configuration
JWT_SECRET = os.getenv("JWT_SECRET", "")
JWT_ALGORITHM = "HS256"
JWT_EXPIRE_HOURS = 8

if not JWT_SECRET and os.getenv("BYPASS_AUTH", "").lower() != "true":
    raise RuntimeError("JWT_SECRET environment variable must be set and non-empty")


class LoginRequest(BaseModel):
    email: str
    password: str


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

    # M9: Validate critical env vars so misconfiguration is caught at startup,
    # not on the first request. The deploy.sh sets these from YAML config.
    missing = [var for var in ("RAG_SERVER_URL", "INGESTOR_URL") if not os.getenv(var)]
    if missing:
        logger.warning(f"⚠️  Missing env vars: {missing} — using defaults")

    logger.info(
        f"Gateway starting: RAG={RAG_SERVER_URL}, Ingestor={INGESTOR_URL}, "
        f"auth={'bypassed' if BYPASS_AUTH else 'enabled'}"
    )


@app.exception_handler(HTTPException)
async def auth_failure_handler(request: Request, exc: HTTPException):
    """M10: Log authentication/authorization failures for security monitoring.
    Returns a JSONResponse so FastAPI properly converts the exception to HTTP response.
    """
    if exc.status_code in (401, 403):
        ip = request.client.host if request.client else "unknown"
        logger.warning(
            f"Auth failure {exc.status_code} [{request.method} {request.url.path}] "
            f"from {ip}: {exc.detail}"
        )
    # Return JSONResponse with the exception's status code and detail
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail}
    )


def create_jwt(user: User) -> str:
    """Create JWT token for a user."""
    payload = {
        "user_id": user.id,
        "email": user.email,
        "role": user.role.value,
        "area_id": user.area_id,
        "exp": datetime.now(timezone.utc) + timedelta(hours=JWT_EXPIRE_HOURS),
    }
    return pyjwt.encode(payload, JWT_SECRET, algorithm=JWT_ALGORITHM)


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




def admin_required(user: User = Depends(get_user_from_token)) -> User:
    """Require ADMIN role."""
    if user is None and not BYPASS_AUTH:
        raise HTTPException(status_code=401, detail="Auth required")
    if user and user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin only")
    return user


def admin_or_manager_required(user: User = Depends(get_user_from_token)) -> User:
    """Require ADMIN or AREA_MANAGER role."""
    if user is None and not BYPASS_AUTH:
        raise HTTPException(status_code=401, detail="Auth required")
    if user and user.role == Role.USER:
        raise HTTPException(status_code=403, detail="Insufficient permissions")
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


@app.get("/v1/collections/{collection_name}/stats")
async def collection_stats(collection_name: str, user: User = Depends(get_user_from_token)):
    """Stats for a specific collection."""
    if user and user.role != Role.ADMIN:
        if not db.can_access(user, collection_name, Permission.READ):
            raise HTTPException(status_code=403, detail="No access to collection")
    try:
        stats = CollectionManager().stats(collection_name)
        if stats is None:
            raise HTTPException(status_code=404, detail="Collection not found")
        return asdict(stats)
    except HTTPException:
        raise  # Re-raise HTTP exceptions as-is
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error getting stats: {str(e)}")


@app.get("/v1/collections")
async def list_collections(user: User = Depends(get_user_from_token)):
    """List collections user can access."""
    if user is None:
        return {"collections": CollectionManager().list()}

    return {"collections": db.get_user_collections(user)}


# Auth endpoints
@app.post("/auth/session")
async def login(body: LoginRequest, user: User = Depends(get_user_from_token)):
    """Issue JWT for a valid email+password. Caller must be authenticated (BFF uses SYSTEM_API_KEY)."""
    from saldivia.auth.models import verify_password
    target = db.get_user_by_email(body.email)
    if not target or not target.password_hash:
        raise HTTPException(status_code=401, detail="Invalid credentials")
    if not verify_password(body.password, target.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    if not target.active:
        raise HTTPException(status_code=403, detail="Account disabled")
    db.update_last_login(target.id)
    token = create_jwt(target)
    return {"token": token, "user": {"id": target.id, "email": target.email,
                                      "name": target.name, "role": target.role.value,
                                      "area_id": target.area_id}}


@app.delete("/auth/session")
async def logout(user: User = Depends(get_user_from_token)):
    """Logout endpoint (stateless — BFF clears the cookie)."""
    return {"ok": True}


@app.get("/auth/me")
async def me(user_id: int, user: User = Depends(get_user_from_token)):
    """Get profile for a user_id (BFF passes user_id from JWT).
    Note: user_id is supplied by the BFF from the JWT payload — the gateway trusts it
    because the BFF is the only caller (SYSTEM_API_KEY gating)."""
    if user is None or user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin role required")
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    return {"id": target.id, "email": target.email, "name": target.name,
            "role": target.role.value, "area_id": target.area_id,
            "last_login": target.last_login.isoformat() if target.last_login else None}


@app.post("/auth/refresh-key")
async def refresh_my_key(user_id: int, user: User = Depends(get_user_from_token)):
    """Regenerate API key for a user (admin only).
    Note: user_id from JWT payload supplied by BFF."""
    from saldivia.auth.models import generate_api_key
    if user is None or user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin role required")
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    new_key, new_hash = generate_api_key()
    db.update_api_key(user_id, new_hash)
    return {"api_key": new_key}




class CreateUserRequest(BaseModel):
    email: str
    name: str
    area_id: int
    role: str = "user"
    password: Optional[str] = None


class UpdateUserRequest(BaseModel):
    name: Optional[str] = None
    area_id: Optional[int] = None
    role: Optional[str] = None
    active: Optional[bool] = None



class CreateAreaRequest(BaseModel):
    name: str
    description: str = ""


class UpdateAreaRequest(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None


class GrantCollectionRequest(BaseModel):
    collection_name: str
    permission: str = "read"


class CreateSessionRequest(BaseModel):
    collection: str
    crossdoc: bool = False


@app.get("/admin/users")
async def list_users_endpoint(user: User = Depends(admin_required)):
    users = db.list_users()
    return {"users": [{"id": u.id, "email": u.email, "name": u.name,
                        "area_id": u.area_id, "role": u.role.value,
                        "active": u.active,
                        "last_login": u.last_login.isoformat() if u.last_login else None}
                       for u in users]}


@app.post("/admin/users", status_code=201)
async def create_user_endpoint(body: CreateUserRequest, user: User = Depends(admin_required)):
    new_key, new_hash = generate_api_key()
    pw_hash = hash_password(body.password) if body.password else None
    try:
        new_user = db.create_user(
            email=body.email, name=body.name, area_id=body.area_id,
            role=Role(body.role), api_key_hash=new_hash, password_hash=pw_hash
        )
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
    return {"id": new_user.id, "email": new_user.email, "api_key": new_key}


@app.put("/admin/users/{user_id}")
async def update_user_endpoint(user_id: int, body: UpdateUserRequest,
                                user: User = Depends(admin_required)):
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    updates = {k: v for k, v in body.model_dump().items() if v is not None}
    if "role" in updates:
        updates["role"] = Role(updates["role"])
    db.update_user(user_id, **updates)
    return {"ok": True}


@app.delete("/admin/users/{user_id}")
async def delete_user_endpoint(user_id: int, user: User = Depends(admin_required)):
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    db.update_user(user_id, active=False)
    return {"ok": True}


@app.post("/admin/users/{user_id}/reset-key")
async def reset_user_key(user_id: int, user: User = Depends(admin_required)):
    target = db.get_user_by_id(user_id)
    if not target:
        raise HTTPException(status_code=404, detail="User not found")
    new_key, new_hash = generate_api_key()
    db.update_api_key(user_id, new_hash)
    return {"api_key": new_key}


# Area management endpoints
@app.get("/admin/areas")
async def list_areas_endpoint(user: User = Depends(admin_or_manager_required)):
    areas = db.list_areas()
    return {"areas": [{"id": a.id, "name": a.name, "description": a.description} for a in areas]}


@app.post("/admin/areas", status_code=201)
async def create_area_endpoint(body: CreateAreaRequest, user: User = Depends(admin_required)):
    area = db.create_area(body.name, body.description)
    return {"id": area.id, "name": area.name}


@app.put("/admin/areas/{area_id}")
async def update_area_endpoint(area_id: int, body: UpdateAreaRequest,
                                user: User = Depends(admin_required)):
    area = db.get_area(area_id)
    if area is None:
        raise HTTPException(status_code=404, detail="Area not found")
    db.update_area(area_id, name=body.name, description=body.description)
    return {"ok": True}


@app.delete("/admin/areas/{area_id}")
async def delete_area_endpoint(area_id: int, user: User = Depends(admin_required)):
    area = db.get_area(area_id)
    if area is None:
        raise HTTPException(status_code=404, detail="Area not found")
    try:
        db.delete_area(area_id)
    except ValueError as e:
        raise HTTPException(status_code=409, detail=str(e))
    return {"ok": True}


@app.get("/admin/areas/{area_id}/collections")
async def get_area_collections_endpoint(area_id: int, user: User = Depends(admin_or_manager_required)):
    if user and user.role == Role.AREA_MANAGER and user.area_id != area_id:
        raise HTTPException(status_code=403, detail="Can only view your own area")
    collections = db.get_area_collections(area_id)
    return {"collections": [{"name": c.collection_name, "permission": c.permission.value}
                              for c in collections]}


@app.post("/admin/areas/{area_id}/collections")
async def grant_collection_endpoint(area_id: int, body: GrantCollectionRequest,
                                     user: User = Depends(admin_or_manager_required)):
    if user and user.role == Role.AREA_MANAGER and user.area_id != area_id:
        raise HTTPException(status_code=403, detail="Can only modify your own area")
    db.grant_collection_access(area_id, body.collection_name, Permission(body.permission))
    return {"ok": True}


@app.delete("/admin/areas/{area_id}/collections/{collection_name}")
async def revoke_collection_endpoint(area_id: int, collection_name: str,
                                      user: User = Depends(admin_or_manager_required)):
    if user and user.role == Role.AREA_MANAGER and user.area_id != area_id:
        raise HTTPException(status_code=403, detail="Can only modify your own area")
    db.revoke_collection_access(area_id, collection_name)
    return {"ok": True}

# Admin endpoints
@app.get("/admin/audit")
async def get_audit(
    user_id: Optional[int] = None,
    action: Optional[str] = None,
    collection: Optional[str] = None,
    from_ts: Optional[str] = None,
    to_ts: Optional[str] = None,
    limit: int = 100,
    user: User = Depends(get_user_from_token)
):
    """Get audit log with optional filters (admin only)."""
    if user is None or user.role != Role.ADMIN:
        raise HTTPException(status_code=403, detail="Admin only")
    entries = db.get_audit_log_filtered(
        user_id=user_id, action=action, collection=collection,
        from_ts=from_ts, to_ts=to_ts, limit=limit
    )
    return {"entries": [
        {
            "id": e.id,
            "user_id": e.user_id,
            "action": e.action,
            "collection": e.collection,
            "query_preview": e.query_preview,
            "ip_address": e.ip_address,
            "timestamp": e.timestamp.isoformat() if hasattr(e.timestamp, 'isoformat') else str(e.timestamp)
        }
        for e in entries
    ]}




@app.get("/chat/sessions")
async def list_sessions(user_id: int, limit: int = 50,
                         user: User = Depends(get_user_from_token)):
    """List chat sessions for a user (BFF passes user_id from JWT)."""
    sessions = db.list_chat_sessions(user_id=user_id, limit=limit)
    return {"sessions": [{"id": s.id, "title": s.title, "collection": s.collection,
                           "crossdoc": s.crossdoc,
                           "updated_at": s.updated_at.isoformat() if hasattr(s.updated_at, 'isoformat') else str(s.updated_at)}
                          for s in sessions]}


@app.post("/chat/sessions", status_code=201)
async def create_session(body: CreateSessionRequest, user_id: int,
                          user: User = Depends(get_user_from_token)):
    """Create a new chat session."""
    session = db.create_chat_session(user_id=user_id, collection=body.collection,
                                     crossdoc=body.crossdoc)
    return {"id": session.id, "title": session.title, "collection": session.collection}


@app.get("/chat/sessions/{session_id}")
async def get_session(session_id: str, user_id: int,
                       user: User = Depends(get_user_from_token)):
    """Get a specific chat session with messages."""
    session = db.get_chat_session(session_id=session_id, user_id=user_id)
    if not session:
        raise HTTPException(status_code=404, detail="Session not found")
    return {"id": session.id, "title": session.title, "collection": session.collection,
            "crossdoc": session.crossdoc,
            "messages": [{"role": m.role, "content": m.content, "sources": m.sources,
                           "timestamp": m.timestamp.isoformat() if hasattr(m.timestamp, 'isoformat') else str(m.timestamp)}
                          for m in session.messages]}


@app.delete("/chat/sessions/{session_id}")
async def delete_session(session_id: str, user_id: int,
                          user: User = Depends(get_user_from_token)):
    """Delete a chat session and its messages."""
    db.delete_chat_session(session_id=session_id, user_id=user_id)
    return {"ok": True}



def main():
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8090)


if __name__ == "__main__":
    main()
