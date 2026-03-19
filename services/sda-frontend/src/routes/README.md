# Routes

SvelteKit routes for pages and API endpoints. The route structure uses SvelteKit's file-based routing with layout groups.

## Structure

```
routes/
├── (app)/              # Auth-guarded routes (requires valid JWT cookie)
│   ├── +layout.server.ts   # Validates JWT, redirects to /login if invalid
│   ├── chat/           # Chat interface
│   ├── collections/    # Collection management
│   ├── admin/          # Admin-only routes (users, areas, permissions, system, rag-config)
│   ├── audit/          # Audit log viewer
│   ├── settings/       # User settings
│   └── upload/         # File upload for ingestion
├── (auth)/             # Public routes (no auth required)
│   └── login/          # Login form
├── api/                # BFF API endpoints (called by client-side code via fetch)
│   ├── auth/           # Session management
│   ├── chat/           # Chat sessions and streaming
│   ├── crossdoc/       # Crossdoc pipeline endpoints
│   ├── collections/    # Collection CRUD
│   ├── upload/         # File upload
│   └── dev-login/      # Development-only login bypass (disabled in production)
└── +page.server.ts     # Root redirect (/ → /chat if authenticated)
```

## Route groups

### `(app)` — Auth-guarded routes

All routes under `(app)/` are protected by JWT authentication. The layout at `(app)/+layout.server.ts` calls `verifySession(cookies)` on every request. If the JWT is invalid or expired, the user is redirected to `/login`.

### `(auth)` — Public routes

Routes under `(auth)/` are accessible without authentication. Currently only contains `/login`.

### `api/` — BFF API endpoints

These are SvelteKit server routes (`.ts` files with HTTP method exports like `GET`, `POST`, `DELETE`). They are called by client-side code via `fetch()`. Each endpoint:

1. Validates the user's JWT cookie via `verifySession()`
2. Proxies the request to the Auth Gateway using `SYSTEM_API_KEY`
3. Returns the response to the client

## Root redirect

The root `+page.server.ts` checks if the user has a valid JWT cookie. If yes, redirects to `/chat`. If no, redirects to `/login`.
