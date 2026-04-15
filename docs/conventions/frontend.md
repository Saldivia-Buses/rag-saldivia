---
title: Convention: Frontend (Next.js/React)
audience: ai
last_reviewed: 2026-04-15
related:
  - ./testing.md
  - ./git.md
  - ../architecture/overview.md
---

Rules for `apps/web/` and `apps/login/`. Stack: Next.js App Router, React, shadcn/ui, Tailwind v4, TanStack Query, next-themes.

## File naming

| Element | Rule | Example |
|---|---|---|
| Components | PascalCase, `.tsx` | `VehicleTable.tsx`, `UsersAdmin.tsx` |
| Hooks | camelCase with `use` prefix, `.ts` | `useEnabledModules.ts`, `useChatStream.ts` |
| Lib / utils | kebab-case, `.ts` | `module-guard.ts`, `format-date.ts` |
| Pages | `page.tsx` (App Router convention) | `app/(modules)/fleet/page.tsx` |
| Server actions | `actions/<domain>.ts` | `app/actions/users.ts` |
| API routes | `route.ts` | `app/api/health/route.ts` |

DO place core (always-available) routes under `app/(core)/` and module routes (gated by tenant module enablement) under `app/(modules)/`. Module routes are code-split and lazy loaded.

## Design philosophy — Warm Intelligence

The system has a named visual identity: warm cream backgrounds, deep navy accent, Instrument Sans typography, no decoration without justification.

DO use the cream base (`--bg: #faf8f4`) — never plain white. The navy accent (`--accent: #1a5276`) is reserved for primary affordances (buttons, active states, badges).

DO use warm dark in dark mode — `--bg: #1a1812` (warm brown), never pure black.

DO let the design system carry visual weight. Avoid one-off colours, custom shadows, ad-hoc spacing.

## Tokens

CSS variables defined in `apps/web/src/app/globals.css` and exposed as Tailwind utilities via `@theme inline`. The variables themselves are the spec — read them in the source.

DO use token utilities (`bg-bg`, `bg-surface`, `text-fg`, `text-fg-muted`, `text-fg-subtle`, `bg-accent`, `text-accent-fg`, `bg-destructive-subtle`, `text-success`, `text-warning`).

DO use semantic state utilities (`text-destructive`, `text-success`, `text-warning`) for status.

DON'T hardcode hex colours, RGB values, or arbitrary Tailwind colour classes (`bg-blue-500`). The token map must remain the single source of truth.

DON'T use Tailwind v4 `space-y-*` — it does not behave as expected in this stack. Use `flex gap-*` or explicit margin.

## Density

DO set `data-density="compact"` on admin layouts (tables, dense forms) and `data-density="spacious"` on chat / content pages. Components consume the matching CSS vars automatically — no per-component lookup.

## Dark mode

DO use `next-themes` with `attribute="class"`. The `.dark` class is applied to `<html>`. Read with `useTheme()`; toggle with `setTheme("dark" | "light" | "system")`.

DON'T rely on the `prefers-color-scheme` media query for application styles — the class-based approach is what tokens hook into.

DO activate dark mode in Playwright tests via JavaScript injection (`document.documentElement.classList.add("dark")`). Playwright's `colorScheme: 'dark'` does not trigger class-based dark mode.

## Server vs Client components

DO default to Server Components. Use `"use client"` only when the file needs browser APIs, state, effects, refs, or context that lives in the client tree.

DO put data fetching in Server Components (or route loaders); pass plain serialisable props down to Client Components.

DO use Server Actions for mutations from Client Components. Define them in `app/actions/<domain>.ts` with `"use server"`.

## Component checklist (new component)

1. Create in `src/components/ui/<name>.tsx` (primitive) or `src/components/<feature>/<name>.tsx` (feature).
2. Use tokens — never hardcode colours.
3. Story in `apps/web/stories/primitivos/<name>.stories.tsx` (primitive) or `features/`.
4. Test in `__tests__/<name>.test.tsx` with `afterEach(cleanup)` at the top.
5. Run `bun run test:components` — must pass.
6. Open Storybook → Accessibility panel → 0 violations.
7. If visual change to existing primitive: regenerate baseline (`bun run visual:update`) and commit the snapshots.

## Testing

See [testing](./testing.md) for the cross-language testing strategy. Key frontend rules summarised here:

- `afterEach(cleanup)` is mandatory in every component test file
- Use scoped queries from the render result (`const { getByRole } = render(...)`) — never the global `screen`
- `fireEvent`, not `userEvent` (happy-dom incompatibility)
- Mock server actions with `mock.module("@/app/actions/...", () => ({...}))`

## Realtime data

DO subscribe to backend events via the WebSocket Hub. The frontend never polls.

DO use TanStack Query for HTTP cache + invalidation. Invalidation is triggered by WebSocket messages, not by `setInterval`.

DON'T add `setInterval` polling to refresh data — it violates the realtime invariant.

## Forms and validation

DO validate inputs at the form layer (React Hook Form + Zod schema from `packages/shared` if shared with the backend, else inline schema).

DO display server errors next to the field that produced them; use `httperr` JSON error codes from the backend response (`error.code`).
