# (auth) Routes

Public routes accessible without authentication.

## Files

| Route | Description |
|-------|-------------|
| `login/+page.svelte` | Login form with email and password fields. Includes error display for invalid credentials. |
| `login/+page.server.ts` | Handles form POST. Calls `/api/auth/session` with credentials, receives JWT token on success, sets secure httpOnly cookie via `setSessionCookie()`, and redirects to `/chat`. |

## Authentication flow

1. User visits `/login`
2. User submits email + password
3. `+page.server.ts` form action:
   - Validates input (non-empty email and password)
   - Calls `/api/auth/session` POST endpoint (BFF API)
   - BFF proxies request to Auth Gateway `/auth/session`
   - Gateway validates credentials against AuthDB (SQLite)
   - Gateway returns JWT token on success
4. BFF sets JWT as secure httpOnly cookie
5. User is redirected to `/chat`
6. Subsequent requests to `(app)/*` routes include the cookie
7. `(app)/+layout.server.ts` validates the JWT on every request

## Design notes

The login form does NOT expose the JWT token to client-side JavaScript. The token is set as an httpOnly cookie, making it inaccessible to client code (XSS protection).
