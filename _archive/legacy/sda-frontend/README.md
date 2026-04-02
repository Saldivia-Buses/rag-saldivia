# SDA Frontend

Backend for Frontend (BFF) for RAG Saldivia. This SvelteKit 5 application serves the user interface and acts as a secure proxy to the Auth Gateway, handling JWT authentication and API calls.

## Tech stack

- **SvelteKit 5** with Svelte 5 runes (`$state`, `$derived`, `$effect`)
- **TypeScript** strict mode
- **Tailwind CSS v4** for styling
- **Vitest 3.x** for unit/integration tests
- **Playwright** (planned) for E2E tests
- **marked + highlight.js** for markdown rendering
- **DOMPurify** for XSS protection

## Running the application

```bash
# Development server (hot reload, port 5173)
npm run dev

# Production build
npm run build

# Preview production build (port 4173)
npm run preview
```

## Testing

```bash
# Run all tests (Vitest)
npm run test
```

## Environment variables

Required environment variables (set in `.env` or via SvelteKit private env):

- `GATEWAY_URL` — URL of the Auth Gateway (default: `http://localhost:9000`)
- `JWT_SECRET` — JWT signing secret (must match gateway)
- `SYSTEM_API_KEY` — API key for BFF → Gateway communication (system-level auth)
- `ORIGIN` — SvelteKit CSRF origin (e.g. `http://localhost:3000`)

## Directory structure

```
src/
├── lib/                # Library code (components, stores, utils, server-side modules)
│   ├── actions/        # Svelte actions (clickOutside, etc.)
│   ├── components/     # Reusable UI components
│   ├── crossdoc/       # Crossdoc query pipeline
│   ├── server/         # Server-only modules (auth, gateway client)
│   ├── stores/         # Svelte 5 runes-based reactive stores
│   └── utils/          # Utility functions
├── routes/             # SvelteKit pages and API routes
│   ├── (app)/          # Auth-guarded routes (chat, collections, admin, etc.)
│   ├── (auth)/         # Public routes (login)
│   └── api/            # BFF API endpoints
└── app.html            # HTML shell
tests/                  # E2E tests (Playwright)
```

## Architecture notes

This is a **BFF** (Backend for Frontend). Client-side code calls `/api/*` routes, which are implemented as SvelteKit server routes. These routes authenticate the user via JWT cookie, then proxy the request to the Auth Gateway using a system-level API key.

The Auth Gateway (port 9000) handles RBAC and proxies to the RAG Server (port 8081).

For more details, see the main project's `docs/architecture.md`.
