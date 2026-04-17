---
name: frontend-next
description: Use when editing apps/web/** — the Next.js App Router frontend. Covers component conventions, Server vs Client Components, server actions, data fetching against the Go gateway through Traefik, WebSocket subscription patterns, auth/session handling, styling (Tailwind + tokens), and the local dev / deployed frontend split.
---

# frontend-next

Scope: `apps/web/**`.

## Stack (pinned)

- **Next.js 16.2.3** — App Router, Cache Components, Turbopack.
- **React 19.2** — Server Components by default, `use()` hook, async `useTransition`.
- **Tailwind CSS v4** — CSS-first config (`@theme`), Oxide engine, no `tailwind.config.js`.
- **Vercel AI SDK v6** — `streamText`/`generateText` with `stopWhen` for multi-step tools.
- **TypeScript 5** strict, **bun** for install/scripts.
- UI kits in use: **shadcn**, **@radix-ui**, **@base-ui/react** (they coexist — prefer the
  one already present in the component family you touch).

## Next.js 16 — Cache Components

Next 16 replaces the old implicit route cache with opt-in **Cache Components**.
Enable once in `next.config.ts` with `cacheComponents: true`, then annotate what
should be cached:

```tsx
// Inside a Server Component or data-access function
'use cache'
export async function getTenantPlans(slug: string) {
  cacheLife('minutes')
  cacheTag(`tenant:${slug}:plans`)
  return fetch(`${INTERNAL_GATEWAY_URL}/platform/plans`, {
    headers: { 'X-Tenant-Slug': slug },
  }).then(r => r.json())
}
```

- `cacheLife('seconds' | 'minutes' | 'hours' | 'days' | 'max' | custom)` — TTL profile.
- `cacheTag('some:key')` — for targeted invalidation.
- `updateTag('some:key')` (new in 16) — imperative invalidation from a Server Action.
- `revalidateTag` still works; prefer `updateTag` in new code.

Rule of thumb: put `'use cache'` as close to the data access as possible, not
at page level. Mix static + dynamic in the same component.

## React 19 — async patterns

- **`use(promise)`** is the unified async hook — works in Server and Client Components.
  Replaces the useEffect-for-data pattern.
- **Async `useTransition` / `useActionState`** — don't wrap async ops in useState
  pairs; let the transition track pending/error.
- **No `"use server"` at the top of a Server Component** — that marker is only for
  Server Actions. Server Components are the default.

## Tailwind v4

- `@import "tailwindcss"` replaces the three v3 directives.
- Theme tokens live in CSS:

  ```css
  @import "tailwindcss";

  @theme {
    --color-brand: oklch(0.72 0.15 250);
    --font-display: "Inter", sans-serif;
  }
  ```

- No `tailwind.config.js` for most projects. If one exists, it is being phased out.
- Dark mode: `@variant dark (&:where(.dark, .dark *));` — the v3 `darkMode: 'class'`
  is gone.
- Run the upgrade tool (`npx @tailwindcss/upgrade`) for automated migrations.

## Vercel AI SDK v6

- `streamText({ model, messages, tools, stopWhen: ... })` for the chat loop.
- `stopWhen: stepCountIs(5)` bounds multi-step agent loops. Without it, a tool that
  always triggers another tool call can loop indefinitely.
- UI side: `useChat()` from `@ai-sdk/react` — it handles streaming + optimistic UI.
- The `agent` Go service talks to the LLM provider (OpenRouter). The frontend should
  **not** hold an API key — all LLM calls go through `agent` or `chat` via the gateway.

## Supabase — UI-only, not canonical

`@supabase/ssr` and `@supabase/supabase-js` are installed but are **not** part of
the identity or data plane. They power UI-only concerns (incidental pieces, not
load-bearing). The canonical identity source is the Go `auth` service (ed25519
JWT cookie) — see `auth-security`.

Rules:

- Don't use Supabase for auth, session, tenant data, or anything that should flow
  through the Go gateway.
- Treat existing Supabase call sites as UI plumbing — leave them alone unless the
  task explicitly asks.
- Don't add new Supabase call sites for backend-adjacent needs; route through the
  gateway like everything else.

## Commands

```bash
cd apps/web
bun install
bun run dev       # localhost:3000 (dev mode — does NOT hydrate over remote IP)
bun run build
bun run start
bun test          # bun test runner
bun run typecheck
```

For remote-IP access (workstation/test envs) the frontend runs inside Docker via
`deploy/docker-compose*.yml`. Dev mode on bare metal only serves `localhost`.

## Directory layout

```
apps/web/
  src/
    app/                 # App Router — one folder per route
      (core)/            # authenticated group (has login guard in layout)
      (auth)/            # public group (login, recovery)
      api/               # route handlers (only for BFF concerns)
    components/          # shared UI
      tailgrids/core/    # base atoms (Button, Input, …)
    hooks/               # client-side hooks (ws subscriptions, decompose flow)
    lib/                 # utils, api client, tokens
    styles/
```

## Server vs Client

- **Default to Server Components.** Mark with `"use client"` only when you need: local
  state, effects, browser APIs, a library that accesses `window`.
- Server Components fetch data directly (no `useEffect` for data).
- Pass server-rendered data down; hydrate only what needs interactivity.

## Data fetching

- From a Server Component: call the Go gateway directly via `fetch` on the server.
  Traefik routes `/api/*` → services. Use `process.env.INTERNAL_GATEWAY_URL` on server,
  `NEXT_PUBLIC_API_URL` on client.
- For mutations: Server Actions (`"use server"`) OR client → `/api/...` route handler.
- Revalidate with `updateTag(tag)` (preferred in Next 16) or `revalidatePath` after
  mutation. Tag names should match what the cached reader used (e.g. `tenant:acme:plans`).

## Auth

- JWT comes from `/api/auth/login` as an HTTP-only cookie.
- Layout at `app/(core)/layout.tsx` checks the cookie and redirects to `/login` if missing.
- Never parse the JWT client-side; the Go gateway is the identity authority.

## WebSocket / real-time

- A single hub connection in a React context (`WebSocketProvider`).
- Subscriptions are tenant-namespaced: the hub proxies `tenant.{slug}.*` events.
- Custom hooks: `useCrossdocStream`, `useChatStream`, …  Each maps subject → state.
- Never poll. If you see polling, it is a bug — subscribe to the event.

## Styling

- Tailwind utilities first.
- Shared tokens live in `src/styles/tokens.css` as CSS variables.
- Dark mode via `class="dark"` on `<html>` — toggled by a client component.
- No inline colors that bypass tokens.

## Tests

- Component tests: `bun:test` + testing-library/react.
- E2E: Playwright in `apps/web/e2e/`.
- `bun run test` excludes e2e; `bun run e2e` runs Playwright.

## Don't

- Don't use Pages Router patterns (`getServerSideProps`, `_app`, `_document`).
- Don't use raw `fetch('/api/...')` from a Server Component — call the gateway directly.
- Don't import from `_archive/` — it does not exist; it was purged.
- Don't add a new UI library when an existing token/component covers it.

## Known anti-patterns in this repo (fix on contact)

### `"use client"` on pure presentational components

Audited: **68 instances**. Most of `src/components/ui/*.tsx` declare `"use client"`
without using state, effects, or browser APIs — they just forward props to a
Radix primitive with Tailwind classes. Rule:

- If the file has no `useState`, `useEffect`, `useRef`, `onClick`/`onChange` handlers,
  or browser globals → **drop the directive**. The component becomes a Server
  Component and costs zero client bytes.

### Supabase doing **auth** (not just UI)

Audited: 14 Supabase calls include `supabase.auth.signOut`, `signInWithOAuth`,
`signUp`, `resetPasswordForEmail`, `getSession`. **These are auth flows, not UI
plumbing.** The canonical identity is the Go `auth` service (ed25519 JWT cookie).

Rule:

- `supabase.auth.*` calls are **blocking** — route through the Go `auth` gateway instead.
- `supabase.from(...).select()` / Supabase realtime that drives **purely cosmetic
  UI** (live presence indicators, typing dots that don't persist) is OK.
- Anything that reads or writes product data goes through the gateway. Period.

Offenders to refactor: `components/supabase/{login-form,sign-up-form,forgot-password-form,logout-button,current-user-avatar}.tsx`.

### Bundle bloat — no `dynamic()` / no lazy

Audited: **148 `import * as` statements.** Critical ones:

- **46× `import * as THREE from 'three'`** in `src/components/reactbits/` — synchronous
  import of a ~600 KB library in the main bundle.
- **`import * as Y from 'yjs'`** (CRDT collab) — synchronous.
- **`import * as math from 'mathjs'`** — synchronous.
- **`import * as faceapi from 'face-api.js'`** — full ML model bundle, synchronous.

Rule: any library >100 KB that isn't used on the first paint **must** load via
`next/dynamic({ ssr: false })` inside a Client Component boundary. Same for Monaco
editor, Shiki (unless streamed), Mermaid, Rive, xyflow.

`import * as React from 'react'` is an antipattern too (108 occurrences). Use
named imports: `import { forwardRef, useState } from 'react'`.

### Cache Components (Next 16) not used

Audited: **zero** `'use cache'`, `cacheTag`, `updateTag`, `cacheLife` in the repo.
All data fetching is either client-side or uncached server fetch.

Rule: every server-side data read on an authenticated route gets `'use cache'`
with a `cacheTag` tied to the entity (e.g. `tenant:plans`, `user:{id}:threads`).
Mutations call `updateTag(tag)` from their Server Action.

### Duplicated component families

Three overlapping directories: `src/components/ui/`, `src/components/shadcnblocks/`,
`src/components/supabase/`. Same primitives (Button, Label, Tooltip) exist 2–3 times
with slight differences.

Rule: when editing one of these, check if the other variants can be collapsed
into the canonical one (`ui/` is the canonical family). Consolidation task for
`continuous-improvement`.

### Manual DOM access (`window`, `document`, `localStorage`)

Audited: **624+ occurrences.** Dark mode is toggled by `document.documentElement.classList`
and persisted in `localStorage`. Rule: use `next-themes` (already installed) and
the stock Server/Client split; no manual toggles.

### Hardcoded hex colors in components

Audited: badges use literal `#ef4444`, `#3b82f6` instead of Tailwind tokens.
Rule: no inline hex/rgb in TSX. Use Tailwind classes or CSS variables from
`@theme`. A raw color on sight is a `simplicity`/`bloat` finding in review.
