# ADR 024 — Next.js frontend lives inside the tenant container

**Status:** accepted
**Date:** 2026-04-17
**Deciders:** Enzo Saldivia
**Refines:** ADR 023 (one container per tenant) — adds the frontend to the runtime contents.

## Context

ADR 023 put the Go app, Postgres, NATS, Redis, and MinIO inside a single per-tenant
container. It left the Next.js frontend on a CDN (per CLAUDE.md). Reviewing that split
against the actual product shape:

- **The tenant is supposed to be the container.** With the frontend on CDN, a tenant
  is really `container + CDN config + DNS + signed URLs + CORS rules`. The clarity
  win of ADR 023 ("docker run, that's the tenant") is only partial.
- **On-prem delivery.** Current and future customers are B2B-enterprise. "Here is one
  image, run it on your box" is the target delivery mechanism. CDN dependency means
  the customer has to operate CDN infrastructure or accept a proxy into the vendor's.
- **Cookies / WebSockets / CSRF.** Same-origin eliminates the CORS dance that every
  cross-origin deployment carries (pre-flight, `credentials: include`, `SameSite`
  gymnastics, WebSocket origin whitelisting).
- **Branding is per-tenant.** Done at runtime via env vars + SSR, one image can still
  serve every tenant. No per-tenant build, no per-tenant CDN invalidation.
- **The box hosts the workload anyway.** The same workstation that runs the
  container has the spare capacity to serve the frontend — there is no scale reason
  to push rendering to an edge.

Keeping the frontend on CDN buys CDN edge caching for static assets. The workload
is tenant-internal (authenticated users on the enterprise LAN or VPN), not public
web traffic. Edge caching is not load-bearing here.

## Decision

**Bake the Next.js frontend into the per-tenant image. Serve it from the same
container as the backend.**

Runtime contents of the image, extending ADR 023:

- **App** — Go monolith, `:8080`. Also owns the container's public port `:80` and
  reverse-proxies non-API requests to Next.js. One fewer process under `s6-overlay`
  and zero extra config: the Go HTTP mux is already there.
  - `/v1/*`, `/ws/*` → served in-process by chi handlers.
  - everything else → reverse-proxied to `http://127.0.0.1:3000` (Next.js).
- **Frontend** — Next.js standalone build, `node server.js` on `:3000`, managed by
  `s6-overlay` as another service alongside Postgres, NATS, Redis, MinIO.
- Per-tenant branding via runtime env vars (`SDA_TENANT`, `SDA_BRAND_*`), resolved
  server-side during SSR. One image, many tenants, zero rebuilds for branding.

## Consequences

**Positive**

- **Tenant = container, literally.** One artifact, one `docker run`, one backup target.
- **Zero CORS configuration.** Frontend and backend share an origin.
- **On-prem delivery collapses to sending an image.** No DNS, CDN, or cross-origin
  config required on the customer side.
- **WebSockets, cookies, CSRF** all work same-origin out of the box. Less security
  surface (no `Access-Control-Allow-Credentials`).
- **Branding per tenant without rebuilding.** SSR reads env vars at request time.
- **Production parity with dev.** Same routing shape locally and in production.

**Negative**

- **Image size grows by ~150–300 MB** (Node runtime + `.next/standalone`). Irrelevant
  for inhouse; matters only if/when we push to a customer-facing registry.
- **Node runtime in the image** — breaks distroless-compatibility. Acceptable: the
  image already includes Postgres, NATS, Redis, MinIO under `s6-overlay`, so the
  distroless ship has sailed.
- **No CDN edge caching.** Static assets are served from the box, not from edge.
  Fine: the workload is LAN/VPN-scoped, not public web.
- **Cold start adds ~2–3 s** for Next.js to be ready. Total container start stays in
  the ~5–8 s range (dominated by Postgres), which was already the budget.
- **Frontend and backend share the image lifecycle.** A backend-only change forces a
  frontend rebuild in CI. Cache mitigates; the marginal cost is small.

**Neutral**

- `NEXT_PUBLIC_*` build-time variables (bundled into the client) still exist at
  build time; they become "defaults" that runtime SSR values can override for
  any value that truly is per-tenant.

## Alternatives considered

1. **Keep the frontend on CDN (status quo per CLAUDE.md).**
   Rejected. Breaks "tenant = container", complicates on-prem, forces CORS and
   same-site cookie workarounds, and forces per-tenant CDN config for branding.

2. **`next export` static build, served by the Go monolith.**
   Rejected. Next.js App Router with server components, server actions, and
   dynamic routes is not fully static-exportable without rearchitecting. The
   project also already uses SSR-only features (auth-aware shells, tenant branding).

3. **Build-time branding (one image per tenant).**
   Rejected. Defeats the "one image serves all tenants" shape from ADR 023.
   Branding per tenant belongs in env vars, not image variants.

4. **Separate frontend container beside the backend container.**
   Rejected. Reintroduces multi-container-per-tenant coordination that ADR 023
   explicitly ruled out. If everything else collapses into one container, the
   frontend is the smallest additional step, not a reason to undo the collapse.

5. **Embedded Traefik on `:80` as the internal router.**
   Rejected. Adds a third process under `s6-overlay` for routing behavior that
   is 30 lines of `chi` + `httputil.ReverseProxy` in the Go app already running
   on the box. ADR 023 keeps Traefik as the *cross-tenant* edge on the host, not
   inside the tenant container.

6. **Next.js owning `/v1/*` as a proxy** (reverse the direction).
   Rejected. Next.js is not the process we want on the hot path for API traffic,
   WebSocket upgrades, or long-lived SSE streams. Keep Next.js serving what it
   was built for (HTML / RSC / static) and let Go handle API + realtime.

## Open items

- **How `NEXT_PUBLIC_*` is resolved at runtime** (SSR reads env vars directly vs
  a small `/config` endpoint exposed by Go that the client fetches on boot). The
  SSR path is simpler and the default.
- **Hot-reload in dev** stays as-is: `apps/web` runs with `bun dev` on the host
  against a localhost Go. The all-in-one image is a production/test artifact, not
  a development environment.
