# Server

Server-only modules (BFF layer). These files run exclusively on the SvelteKit server and handle authentication, gateway communication, and session management.

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `auth.ts` | JWT session management. Exports `verifySession(cookies)` (decode and validate JWT from cookie) and `setSessionCookie(cookies, token)` (set secure httpOnly cookie). | jose, @sveltejs/kit |
| `gateway.ts` | Typed HTTP client for the Auth Gateway (port 9000). Exports `GatewayError` class and `gw<T>(path, init)` function for making authenticated requests (Bearer token via `SYSTEM_API_KEY`). Includes timeout support. | None (uses fetch) |

## Design notes

### CRITICAL: Server-only imports

These files import from `$env/static/private`, which means they **MUST NEVER be imported from client-side code** (`.svelte` files or client-side `.ts` files). SvelteKit will throw a build error if you try to import them on the client.

**Correct usage:**
- Import in `+page.server.ts`, `+layout.server.ts`, `+server.ts` (API routes)
- Import in other server-only files (files that are never imported by client code)

**Incorrect usage:**
- Import in `+page.svelte`, `+layout.svelte` (will cause build error)
- Import in client-side stores, components, or utils (will cause build error)

### Authentication flow

1. User logs in via `/login` (form POST)
2. BFF calls gateway `/auth/session` endpoint
3. Gateway returns JWT token
4. BFF sets secure httpOnly cookie via `setSessionCookie()`
5. All subsequent requests to `(app)/*` routes are validated via `verifySession()` in `(app)/+layout.server.ts`
6. If JWT is invalid/expired, user is redirected to `/login`

### BFF → Gateway communication

The `gateway.ts` client uses **`SYSTEM_API_KEY`** (not user JWT) to authenticate BFF → Gateway requests. This is a system-level API key that allows the BFF to act on behalf of the user without exposing the gateway to the public internet.

The gateway validates both:
- BFF requests (Bearer `SYSTEM_API_KEY`)
- User context (user ID, role, area extracted from JWT, passed in request body or headers as needed)
