# API Routes

BFF (Backend for Frontend) API endpoints. These are SvelteKit server routes called by client-side code via `fetch()`.

## Authentication

All API routes (except `/api/dev-login`) validate the user's JWT cookie via `verifySession(cookies)` before processing the request. If the JWT is invalid or expired, the route returns a 401 Unauthorized response.

## Files

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth/session` | GET | Validate current session. Returns user info if JWT is valid. |
| `/api/auth/session` | DELETE | Destroy session (logout). Clears the JWT cookie. |
| `/api/chat/sessions` | GET | List all chat sessions for the current user. |
| `/api/chat/sessions` | POST | Create a new chat session. |
| `/api/chat/sessions/[id]` | GET | Get a specific chat session (messages, metadata). |
| `/api/chat/sessions/[id]` | DELETE | Delete a chat session. |
| `/api/chat/stream/[id]` | GET | SSE stream endpoint. Proxies to gateway `/rag/stream`, returns server-sent events with LLM response chunks. |
| `/api/collections` | GET | List all collections. |
| `/api/collections` | POST | Create a new collection. |
| `/api/collections/[name]` | GET | Get collection details (document count, stats). |
| `/api/collections/[name]` | DELETE | Delete a collection. |
| `/api/crossdoc/decompose` | POST | Decompose a question into sub-queries. Returns `{ subQueries: string[] }`. |
| `/api/crossdoc/subquery` | POST | Execute a single sub-query. Returns `{ content: string, sources: Source[] }`. |
| `/api/crossdoc/synthesize` | POST | Synthesize results from sub-queries into a final answer. Returns `{ synthesis: string }`. |
| `/api/upload` | POST | Upload a file for ingestion. Proxies to NV-Ingest service. Returns ingestion job ID. |
| `/api/dev-login` | POST | Development-only login bypass. Returns a JWT token for the specified email without password validation. **Disabled in production.** |

## BFF pattern

Each endpoint follows this pattern:

1. Validate user's JWT cookie via `verifySession(cookies)`
2. Proxy request to Auth Gateway (port 9000) using `SYSTEM_API_KEY` Bearer auth
3. Gateway validates RBAC and proxies to RAG Server (port 8081)
4. Gateway returns response
5. BFF returns response to client

This ensures:
- User authentication is handled consistently
- Gateway/RAG Server are not exposed to the public internet
- RBAC is enforced at the gateway level
- Client-side code has a clean, typed API to work with
