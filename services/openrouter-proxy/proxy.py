# services/openrouter-proxy/proxy.py
"""OpenRouter proxy with header injection."""
import os
from fastapi import FastAPI, Request, Response
import httpx

app = FastAPI(title="OpenRouter Proxy")

OPENROUTER_URL = os.getenv("OPENROUTER_URL", "https://openrouter.ai/api/v1")
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY", "")

@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE"])
async def proxy(path: str, request: Request):
    headers = dict(request.headers)
    headers["HTTP-Referer"] = "https://rag-saldivia.local"
    headers["X-Title"] = "RAG Saldivia"
    if OPENROUTER_API_KEY:
        headers["Authorization"] = f"Bearer {OPENROUTER_API_KEY}"
    headers.pop("host", None)

    async with httpx.AsyncClient(timeout=120) as client:
        resp = await client.request(
            method=request.method,
            url=f"{OPENROUTER_URL}/{path}",
            headers=headers,
            content=await request.body(),
        )
        return Response(content=resp.content, status_code=resp.status_code)

@app.get("/health")
async def health():
    return {"status": "ok"}
