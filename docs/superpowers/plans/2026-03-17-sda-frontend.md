# SDA Frontend — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the SDA SvelteKit 5 enterprise frontend — auth, chat (split view + SSE streaming), collections, admin, audit, and settings — packaged as a Docker service in the existing stack.

**Architecture:** SvelteKit 5 with adapter-node as a BFF (Backend for Frontend). The browser sends JWT in httpOnly cookie; the server-side BFF decodes the JWT locally (using `JWT_SECRET`) and forwards requests to the gateway using `SYSTEM_API_KEY` as Bearer token. All gateway calls are server-side — the API key never reaches the browser.

**Tech Stack:** SvelteKit 5, Svelte 5, Shadcn-Svelte, Tailwind CSS 4, Lucide-Svelte, mode-watcher, Node.js 22, Docker adapter-node

**Prereq:** Plan A (Gateway Extensions) must be deployed. The gateway must have `JWT_SECRET` set and be responding on `:9000` (external) or `:8090` (internal Docker).

---

## File Map

```
services/sda-frontend/
  src/
    routes/
      (auth)/
        login/
          +page.svelte            ← Login form (email + password)
          +page.server.ts         ← POST /auth/session → set cookie
      (app)/
        +layout.svelte            ← Sidebar + top bar + auth guard wrapper
        +layout.server.ts         ← Validate JWT cookie; redirect to /login if invalid
        chat/
          +page.svelte            ← Redirect to new session or show empty state
          +page.server.ts         ← Create session, redirect to /chat/[id]
          [id]/
            +page.svelte          ← 3-panel split: history | conversation | sources
            +page.server.ts       ← Load session data
        collections/
          +page.svelte            ← Grid of collections with stats
          +page.server.ts         ← Load collections list
          [name]/
            +page.svelte          ← Collection detail + ingest form
            +page.server.ts       ← Load collection stats
        admin/
          +layout.svelte          ← Guard: ADMIN or AREA_MANAGER
          users/
            +page.svelte          ← Users table with CRUD actions
            +page.server.ts       ← Load users + handle form actions
          areas/
            [id]/
              +page.svelte        ← Area detail + collection permissions
              +page.server.ts     ← Load area data + handle actions
        audit/
          +page.svelte            ← Audit log with filters
          +page.server.ts         ← Load audit entries (ADMIN only)
        settings/
          +page.svelte            ← Profile + API key + preferences
          +page.server.ts         ← Load user profile, handle actions
    api/
      auth/
        session/+server.ts        ← POST login → cookie; DELETE logout → clear cookie
      chat/
        sessions/+server.ts       ← GET list, POST create
        sessions/[id]/+server.ts  ← GET session + messages, DELETE
        stream/[id]/+server.ts    ← POST → SSE stream from /v1/generate + save messages
      admin/
        users/+server.ts          ← Proxy admin user CRUD
        areas/+server.ts          ← Proxy admin areas + collection perms
    lib/
      server/
        gateway.ts                ← Typed fetch wrapper for all gateway calls
        auth.ts                   ← JWT decode (local, no network) + session helpers
      stores/
        session.svelte.ts         ← Svelte 5 rune-based store: current user info
        chat.svelte.ts            ← Active session, messages, sources, streaming state
      components/
        sidebar/
          Sidebar.svelte           ← Icon-only sidebar with role-based items
          SidebarItem.svelte       ← Individual nav item with tooltip
        chat/
          MessageBubble.svelte     ← Single message (user or assistant)
          SourceCard.svelte        ← Source reference card in right panel
          HistoryItem.svelte       ← Session item in left panel
          ChatInput.svelte         ← Input bar with send button
        ui/                        ← Shadcn-svelte components (auto-generated)
  static/
    favicon.png
  svelte.config.js                ← adapter-node config
  tailwind.config.ts              ← Tailwind 4 config with SDA color tokens
  app.html                        ← HTML shell with mode-watcher
  package.json
  Dockerfile
```

---

### Task 1: Project scaffold

**Files:**
- Create: `services/sda-frontend/` (whole directory)
- Create: `services/sda-frontend/package.json`
- Create: `services/sda-frontend/svelte.config.js`
- Create: `services/sda-frontend/tailwind.config.ts`
- Create: `services/sda-frontend/app.html`
- Create: `services/sda-frontend/src/app.css`
- Create: `services/sda-frontend/Dockerfile`

- [ ] **Create SvelteKit project**

```bash
cd ~/rag-saldivia/services
npm create svelte@latest sda-frontend
# Select: Skeleton project, TypeScript, No additional options
cd sda-frontend
```

- [ ] **Install dependencies**

```bash
npm install
# SvelteKit adapter
npm install @sveltejs/adapter-node

# Tailwind 4
npm install -D tailwindcss @tailwindcss/vite

# Shadcn-Svelte (init after Tailwind)
npx shadcn-svelte@latest init

# Lucide icons
npm install lucide-svelte

# mode-watcher (dark mode)
npm install mode-watcher

# JWT decode (client-side decode only, no verify — just read claims)
npm install jose
```

- [ ] **Configure adapter-node**

Edit `svelte.config.js`:

```javascript
import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
    preprocess: vitePreprocess(),
    kit: {
        adapter: adapter({
            out: 'build'
        })
    }
};

export default config;
```

- [ ] **Configure Tailwind 4**

Edit `vite.config.ts` to add the Tailwind plugin:

```typescript
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
    plugins: [tailwindcss(), sveltekit()]
});
```

Create `src/app.css` with SDA tokens:

```css
@import "tailwindcss";

:root {
    --color-bg: #0f172a;
    --color-bg-secondary: #0c1220;
    --color-border: #1e293b;
    --color-accent: #6366f1;
    --color-accent-dark: #4338ca;
    --color-text: #e2e8f0;
    --color-text-muted: #94a3b8;
    --color-text-faint: #475569;
}
```

- [ ] **Create app.html**

```html
<!doctype html>
<html lang="es" class="dark">
    <head>
        <meta charset="utf-8" />
        <link rel="icon" href="%sveltekit.assets%/favicon.png" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        %sveltekit.head%
    </head>
    <body class="bg-[#0f172a] text-[#e2e8f0] antialiased" data-sveltekit-preload-data="hover">
        <div style="display: contents">%sveltekit.body%</div>
    </body>
</html>
```

- [ ] **Create Dockerfile**

```dockerfile
# Stage 1: build
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: runtime (~100 MB)
FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json .
ENV NODE_ENV=production
EXPOSE 3000
CMD ["node", "build"]
```

- [ ] **Verify dev server starts**

```bash
cd ~/rag-saldivia/services/sda-frontend
npm run dev
```

Expected: `Local: http://localhost:5173` with no errors

- [ ] **Verify build works**

```bash
npm run build
```

Expected: `build/` directory created with `index.js` inside

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/
git commit -m "feat: scaffold SvelteKit 5 + Tailwind 4 + adapter-node for SDA frontend"
```

---

### Task 2: Gateway client + auth utilities

**Files:**
- Create: `services/sda-frontend/src/lib/server/gateway.ts`
- Create: `services/sda-frontend/src/lib/server/auth.ts`
- Create: `services/sda-frontend/src/lib/server/gateway.test.ts`

All gateway calls are server-side only (in `src/lib/server/`). The browser never calls the gateway directly.

- [ ] **Create gateway.ts**

```typescript
// src/lib/server/gateway.ts
// Typed wrapper for all gateway API calls. Uses SYSTEM_API_KEY Bearer auth.
const GATEWAY_URL = process.env.GATEWAY_URL ?? 'http://localhost:9000';
const SYSTEM_API_KEY = process.env.SYSTEM_API_KEY ?? '';

const headers = () => ({
    'Authorization': `Bearer ${SYSTEM_API_KEY}`,
    'Content-Type': 'application/json',
});

async function gw<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${GATEWAY_URL}${path}`, {
        ...init,
        headers: { ...headers(), ...(init?.headers ?? {}) },
    });
    if (!res.ok) {
        const detail = await res.text();
        throw { status: res.status, detail };
    }
    return res.json() as Promise<T>;
}

// Auth
export async function gatewayLogin(email: string, password: string) {
    return gw<{ token: string; user: SessionUser }>(
        '/auth/session',
        { method: 'POST', body: JSON.stringify({ email, password }) }
    );
}

export async function gatewayGetMe(userId: number) {
    return gw<SessionUser>(`/auth/me?user_id=${userId}`);
}

export async function gatewayRefreshKey(userId: number) {
    return gw<{ api_key: string }>(`/auth/refresh-key?user_id=${userId}`, { method: 'POST' });
}

// Collections
export async function gatewayListCollections() {
    return gw<{ collections: string[] }>('/v1/collections');
}

export async function gatewayCollectionStats(name: string) {
    return gw<CollectionStats>(`/v1/collections/${name}/stats`);
}

// Chat sessions
export async function gatewayListSessions(userId: number) {
    return gw<{ sessions: ChatSessionSummary[] }>(`/chat/sessions?user_id=${userId}`);
}

export async function gatewayCreateSession(userId: number, collection: string, crossdoc = false) {
    return gw<{ id: string; title: string; collection: string }>(
        `/chat/sessions?user_id=${userId}`,
        { method: 'POST', body: JSON.stringify({ collection, crossdoc }) }
    );
}

export async function gatewayGetSession(sessionId: string, userId: number) {
    return gw<ChatSessionDetail>(`/chat/sessions/${sessionId}?user_id=${userId}`);
}

export async function gatewayDeleteSession(sessionId: string, userId: number) {
    return gw<{ ok: boolean }>(`/chat/sessions/${sessionId}?user_id=${userId}`, { method: 'DELETE' });
}

// Admin
export async function gatewayListUsers() {
    return gw<{ users: AdminUser[] }>('/admin/users');
}

export async function gatewayCreateUser(data: CreateUserData) {
    return gw<{ id: number; email: string; api_key: string }>(
        '/admin/users', { method: 'POST', body: JSON.stringify(data) }
    );
}

export async function gatewayUpdateUser(id: number, data: Partial<AdminUser>) {
    return gw<{ ok: boolean }>(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export async function gatewayDeleteUser(id: number) {
    return gw<{ ok: boolean }>(`/admin/users/${id}`, { method: 'DELETE' });
}

export async function gatewayListAreas() {
    return gw<{ areas: AreaSummary[] }>('/admin/areas');
}

export async function gatewayGetAreaCollections(areaId: number) {
    return gw<{ collections: AreaCollection[] }>(`/admin/areas/${areaId}/collections`);
}

export async function gatewayGrantCollection(areaId: number, collectionName: string, permission = 'read') {
    return gw<{ ok: boolean }>(
        `/admin/areas/${areaId}/collections`,
        { method: 'POST', body: JSON.stringify({ collection_name: collectionName, permission }) }
    );
}

export async function gatewayRevokeCollection(areaId: number, collectionName: string) {
    return gw<{ ok: boolean }>(`/admin/areas/${areaId}/collections/${collectionName}`, { method: 'DELETE' });
}

export async function gatewayGetAudit(params: AuditParams) {
    const qs = new URLSearchParams(
        Object.entries(params).filter(([, v]) => v != null).map(([k, v]) => [k, String(v)])
    ).toString();
    return gw<{ entries: AuditEntry[] }>(`/admin/audit${qs ? '?' + qs : ''}`);
}

// Types
export interface SessionUser {
    id: number; email: string; name: string; role: string; area_id: number;
}
export interface CollectionStats {
    collection: string; entity_count: number; document_count?: number;
}
export interface ChatSessionSummary {
    id: string; title: string; collection: string; crossdoc: boolean; updated_at: string;
}
export interface ChatSessionDetail extends ChatSessionSummary {
    messages: ChatMessage[];
}
export interface ChatMessage {
    role: 'user' | 'assistant'; content: string; sources?: Source[]; timestamp: string;
}
export interface Source {
    document: string; page?: number; excerpt: string;
}
export interface AdminUser {
    id: number; email: string; name: string; area_id: number; role: string;
    active: boolean; last_login: string | null;
}
export interface CreateUserData {
    email: string; name: string; area_id: number; role: string; password?: string;
}
export interface AreaSummary {
    id: number; name: string; description: string;
}
export interface AreaCollection {
    name: string; permission: string;
}
export interface AuditEntry {
    id: number; user_id: number; action: string; collection: string | null;
    query_preview: string | null; ip_address: string; timestamp: string;
}
export interface AuditParams {
    user_id?: number; action?: string; collection?: string;
    from?: string; to?: string; limit?: number;
}
```

- [ ] **Create auth.ts**

```typescript
// src/lib/server/auth.ts
// JWT decode (server-side, reads JWT_SECRET from env).
import { jwtVerify, SignJWT } from 'jose';
import type { SessionUser } from './gateway.ts';
import type { Cookies } from '@sveltejs/kit';

const JWT_SECRET_RAW = process.env.JWT_SECRET ?? '';
const secret = new TextEncoder().encode(JWT_SECRET_RAW);
const COOKIE_NAME = 'sda_session';

export async function verifySession(cookies: Cookies): Promise<SessionUser | null> {
    const token = cookies.get(COOKIE_NAME);
    if (!token) return null;
    try {
        const { payload } = await jwtVerify(token, secret);
        return {
            id: payload['user_id'] as number,
            email: payload['email'] as string,
            name: payload['name'] as string ?? '',
            role: payload['role'] as string,
            area_id: payload['area_id'] as number,
        };
    } catch {
        return null;
    }
}

export function setSessionCookie(cookies: Cookies, token: string) {
    cookies.set(COOKIE_NAME, token, {
        path: '/',
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 60 * 60 * 8   // 8 hours
    });
}

export function clearSessionCookie(cookies: Cookies) {
    cookies.delete(COOKIE_NAME, { path: '/' });
}
```

- [ ] **Write test for auth.ts**

Create `src/lib/server/auth.test.ts`:

```typescript
import { describe, it, expect, vi } from 'vitest';
// Test that verifySession returns null for missing token
describe('auth', () => {
    it('returns null for missing cookie', async () => {
        const { verifySession } = await import('./auth.ts');
        const mockCookies = { get: () => undefined } as any;
        const result = await verifySession(mockCookies);
        expect(result).toBeNull();
    });
});
```

Add to `package.json` scripts: `"test": "vitest run"`

Add `vitest` and `@vitest/ui` to devDependencies:
```bash
npm install -D vitest
```

- [ ] **Run test**

```bash
npm test
```

Expected: PASS — `returns null for missing cookie`

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/server/
git commit -m "feat: add gateway client + JWT auth utilities (server-side)"
```

---

### Task 3: Auth flow (login + hooks + cookie)

**Files:**
- Create: `services/sda-frontend/src/hooks.server.ts`
- Create: `services/sda-frontend/src/routes/(auth)/login/+page.svelte`
- Create: `services/sda-frontend/src/routes/(auth)/login/+page.server.ts`
- Create: `services/sda-frontend/src/routes/(app)/+layout.server.ts`

- [ ] **Create hooks.server.ts**

```typescript
// src/hooks.server.ts
import type { Handle } from '@sveltejs/kit';
import { verifySession } from '$lib/server/auth';

export const handle: Handle = async ({ event, resolve }) => {
    event.locals.user = await verifySession(event.cookies) ?? null;
    return resolve(event);
};
```

Create `src/app.d.ts` to type the locals:

```typescript
// src/app.d.ts
import type { SessionUser } from '$lib/server/gateway';
declare global {
    namespace App {
        interface Locals {
            user: SessionUser | null;
        }
    }
}
export {};
```

- [ ] **Create login page**

`src/routes/(auth)/login/+page.server.ts`:

```typescript
import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { gatewayLogin } from '$lib/server/gateway';
import { setSessionCookie } from '$lib/server/auth';

export const load: PageServerLoad = async ({ locals }) => {
    if (locals.user) throw redirect(302, '/chat');
    return {};
};

export const actions: Actions = {
    default: async ({ request, cookies }) => {
        const data = await request.formData();
        const email = data.get('email') as string;
        const password = data.get('password') as string;

        if (!email || !password) {
            return fail(400, { error: 'Email y contraseña requeridos' });
        }

        try {
            const result = await gatewayLogin(email, password);
            setSessionCookie(cookies, result.token);
        } catch (e: any) {
            return fail(401, { error: 'Email o contraseña incorrectos' });
        }

        throw redirect(302, '/chat');
    }
};
```

`src/routes/(auth)/login/+page.svelte`:

```svelte
<script lang="ts">
    import { enhance } from '$app/forms';
    let { form } = $props();
</script>

<div class="min-h-screen flex items-center justify-center bg-[#070d1a]">
    <div class="w-full max-w-sm">
        <!-- Logo -->
        <div class="text-center mb-8">
            <div class="inline-block bg-[#6366f1] text-white font-bold text-lg px-4 py-2 rounded">
                SDA
            </div>
            <p class="text-[#475569] mt-2 text-sm">Sistema de Documentación Asistida</p>
        </div>

        <!-- Form -->
        <form method="POST" use:enhance class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-8">
            {#if form?.error}
                <p class="text-red-400 text-sm mb-4">{form.error}</p>
            {/if}

            <div class="mb-4">
                <label for="email" class="block text-sm text-[#94a3b8] mb-1">Email</label>
                <input
                    id="email" name="email" type="email" required
                    class="w-full bg-[#1e293b] border border-[#334155] rounded px-3 py-2
                           text-[#e2e8f0] text-sm focus:outline-none focus:border-[#6366f1]"
                />
            </div>

            <div class="mb-6">
                <label for="password" class="block text-sm text-[#94a3b8] mb-1">Contraseña</label>
                <input
                    id="password" name="password" type="password" required
                    class="w-full bg-[#1e293b] border border-[#334155] rounded px-3 py-2
                           text-[#e2e8f0] text-sm focus:outline-none focus:border-[#6366f1]"
                />
            </div>

            <button
                type="submit"
                class="w-full bg-[#6366f1] hover:bg-[#4f46e5] text-white rounded py-2 text-sm
                       font-medium transition-colors"
            >
                Ingresar
            </button>
        </form>
    </div>
</div>
```

- [ ] **Create app layout guard**

`src/routes/(app)/+layout.server.ts`:

```typescript
import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ locals, url }) => {
    if (!locals.user) {
        throw redirect(302, `/login?next=${url.pathname}`);
    }
    return { user: locals.user };
};
```

- [ ] **Create logout API route**

`src/routes/api/auth/session/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { gatewayLogin } from '$lib/server/gateway';
import { setSessionCookie, clearSessionCookie } from '$lib/server/auth';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request, cookies }) => {
    const { email, password } = await request.json();
    try {
        const result = await gatewayLogin(email, password);
        setSessionCookie(cookies, result.token);
        return json({ user: result.user });
    } catch (e: any) {
        throw error(401, e.detail ?? 'Invalid credentials');
    }
};

export const DELETE: RequestHandler = async ({ cookies }) => {
    clearSessionCookie(cookies);
    return json({ ok: true });
};
```

- [ ] **Test login flow manually**

```bash
npm run dev
# Open http://localhost:5173/login
# Try logging in with bad credentials → should show error
# (Full test requires gateway running)
```

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/hooks.server.ts \
        services/sda-frontend/src/app.d.ts \
        services/sda-frontend/src/routes/
git commit -m "feat: add auth flow (login page, hooks, JWT cookie)"
```

---

### Task 4: App layout + sidebar

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/+layout.svelte`
- Create: `services/sda-frontend/src/lib/components/sidebar/Sidebar.svelte`
- Create: `services/sda-frontend/src/lib/components/sidebar/SidebarItem.svelte`

- [ ] **Create SidebarItem component**

`src/lib/components/sidebar/SidebarItem.svelte`:

```svelte
<script lang="ts">
    import { page } from '$app/stores';
    let { href, label, icon: Icon } = $props<{
        href: string;
        label: string;
        icon: any;
    }>();

    let active = $derived($page.url.pathname.startsWith(href));
</script>

<a
    {href}
    class="relative group flex items-center justify-center w-7 h-6 rounded
           transition-colors {active
               ? 'bg-[#1e293b] border-l-2 border-[#6366f1]'
               : 'hover:bg-[#1e293b]'}"
    title={label}
>
    <Icon size={14} class="text-{active ? '[#a5b4fc]' : '[#64748b]'}" />
    <!-- Tooltip -->
    <span class="absolute left-full ml-2 px-2 py-1 bg-[#1e293b] text-[#e2e8f0] text-xs
                 rounded opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap z-50">
        {label}
    </span>
</a>
```

- [ ] **Create Sidebar component**

`src/lib/components/sidebar/Sidebar.svelte`:

```svelte
<script lang="ts">
    import { MessageSquare, BookOpen, Upload, Users, ClipboardList, Settings } from 'lucide-svelte';
    import SidebarItem from './SidebarItem.svelte';

    let { role, areaId } = $props<{ role: string; areaId: number }>();

    const isAdmin = role === 'admin';
    const isManager = role === 'admin' || role === 'area_manager';
</script>

<nav class="w-10 min-h-screen bg-[#0a0f1e] border-r border-[#1e293b]
            flex flex-col items-center py-2 gap-1">

    <!-- Logo badge -->
    <div class="bg-[#6366f1] text-white font-bold text-[9px] w-[26px] h-5
                flex items-center justify-center rounded mb-1.5">
        SDA
    </div>

    <SidebarItem href="/chat" label="Chat" icon={MessageSquare} />
    <SidebarItem href="/collections" label="Colecciones" icon={BookOpen} />

    {#if isManager}
        <SidebarItem href="/collections" label="Ingestión" icon={Upload} />
    {/if}

    {#if isManager}
        <SidebarItem href="/admin/users" label={isAdmin ? 'Admin global' : 'Admin área'} icon={Users} />
    {/if}

    {#if isAdmin}
        <SidebarItem href="/audit" label="Auditoría" icon={ClipboardList} />
    {/if}

    <!-- Settings at bottom -->
    <div class="mt-auto">
        <SidebarItem href="/settings" label="Configuración" icon={Settings} />
    </div>
</nav>
```

- [ ] **Create app layout**

`src/routes/(app)/+layout.svelte`:

```svelte
<script lang="ts">
    import Sidebar from '$lib/components/sidebar/Sidebar.svelte';
    let { data, children } = $props();
</script>

<div class="flex min-h-screen bg-[#0f172a]">
    <Sidebar role={data.user.role} areaId={data.user.area_id} />
    <main class="flex-1 overflow-auto">
        {@render children()}
    </main>
</div>
```

- [ ] **Verify sidebar renders**

```bash
npm run dev
# Login → should see sidebar with icons on the left
```

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/lib/components/sidebar/ \
        services/sda-frontend/src/routes/(app)/+layout.svelte
git commit -m "feat: add app layout with icon-only sidebar (role-based)"
```

---

### Task 5: Chat module — session + SSE streaming

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/chat/[id]/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/chat/[id]/+page.server.ts`
- Create: `services/sda-frontend/src/routes/(app)/chat/+page.server.ts`
- Create: `services/sda-frontend/src/routes/api/chat/stream/[id]/+server.ts`
- Create: `services/sda-frontend/src/routes/api/chat/sessions/+server.ts`
- Create: `services/sda-frontend/src/routes/api/chat/sessions/[id]/+server.ts`
- Create: `services/sda-frontend/src/lib/stores/chat.svelte.ts`

- [ ] **Create chat store**

`src/lib/stores/chat.svelte.ts`:

```typescript
// Svelte 5 runes-based reactive store for chat state
export class ChatStore {
    messages = $state<Message[]>([]);
    sources = $state<Source[]>([]);
    streaming = $state(false);
    streamingContent = $state('');
    collection = $state('');
    crossdoc = $state(false);

    addUserMessage(content: string) {
        this.messages.push({ role: 'user', content, timestamp: new Date().toISOString() });
    }

    startStream() {
        this.streaming = true;
        this.streamingContent = '';
        this.sources = [];
    }

    appendToken(token: string) {
        this.streamingContent += token;
    }

    setSources(sources: Source[]) {
        this.sources = sources;
    }

    finalizeStream() {
        if (this.streamingContent) {
            this.messages.push({
                role: 'assistant',
                content: this.streamingContent,
                sources: [...this.sources],
                timestamp: new Date().toISOString()
            });
        }
        this.streaming = false;
        this.streamingContent = '';
    }

    loadMessages(messages: Message[]) {
        this.messages = messages;
    }
}

export interface Message {
    role: 'user' | 'assistant';
    content: string;
    sources?: Source[];
    timestamp: string;
}

export interface Source {
    document: string;
    page?: number;
    excerpt: string;
}
```

- [ ] **Create SSE stream API route**

`src/routes/api/chat/stream/[id]/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';

const GATEWAY_URL = process.env.GATEWAY_URL ?? 'http://localhost:9000';
const SYSTEM_API_KEY = process.env.SYSTEM_API_KEY ?? '';

export const POST: RequestHandler = async ({ params, request, locals }) => {
    if (!locals.user) throw error(401, 'Unauthorized');

    const { query, collection_names, crossdoc } = await request.json();
    const sessionId = params.id;

    // Forward to gateway /v1/generate as SSE
    const gatewayResp = await fetch(`${GATEWAY_URL}/v1/generate`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${SYSTEM_API_KEY}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            messages: [{ role: 'user', content: query }],
            collection_names,
            use_knowledge_base: true,
        }),
    });

    if (!gatewayResp.ok) {
        throw error(gatewayResp.status, await gatewayResp.text());
    }

    // Pipe the SSE stream back to the browser
    return new Response(gatewayResp.body, {
        headers: {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
        },
    });
};
```

- [ ] **Create sessions API route**

`src/routes/api/chat/sessions/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { gatewayListSessions, gatewayCreateSession, gatewayDeleteSession } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ locals }) => {
    if (!locals.user) throw error(401);
    const data = await gatewayListSessions(locals.user.id);
    return json(data);
};

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);
    const { collection, crossdoc } = await request.json();
    const session = await gatewayCreateSession(locals.user.id, collection, crossdoc);
    return json(session, { status: 201 });
};
```

- [ ] **Create per-session API route (GET + DELETE)**

`src/routes/api/chat/sessions/[id]/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { gatewayGetSession, gatewayDeleteSession } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    const session = await gatewayGetSession(params.id, locals.user.id);
    if (!session) throw error(404, 'Session not found');
    return json(session);
};

export const DELETE: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    await gatewayDeleteSession(params.id, locals.user.id);
    return json({ ok: true });
};
```

- [ ] **Create chat page server**

`src/routes/(app)/chat/+page.server.ts` (redirects to new session):

```typescript
import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayCreateSession, gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    const { collections } = await gatewayListCollections();
    const defaultCollection = collections[0] ?? '';
    const session = await gatewayCreateSession(locals.user!.id, defaultCollection);
    throw redirect(302, `/chat/${session.id}`);
};
```

`src/routes/(app)/chat/[id]/+page.server.ts`:

```typescript
import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { gatewayGetSession, gatewayListSessions, gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params, locals }) => {
    const [sessionData, historyData, collectionsData] = await Promise.all([
        gatewayGetSession(params.id, locals.user!.id),
        gatewayListSessions(locals.user!.id),
        gatewayListCollections(),
    ]);

    if (!sessionData) throw error(404, 'Session not found');

    return {
        session: sessionData,
        history: historyData.sessions,
        collections: collectionsData.collections,
    };
};
```

- [ ] **Create 3-panel chat page**

`src/routes/(app)/chat/[id]/+page.svelte`:

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { ChatStore } from '$lib/stores/chat.svelte';
    import { MessageSquare, Send, RefreshCw } from 'lucide-svelte';

    let { data } = $props();
    const chat = new ChatStore();

    let input = $state('');
    let selectedCollection = $state(data.session.collection);

    onMount(() => {
        chat.collection = data.session.collection;
        chat.crossdoc = data.session.crossdoc;
        if (data.session.messages?.length) {
            chat.loadMessages(data.session.messages);
        }
    });

    async function sendMessage() {
        if (!input.trim() || chat.streaming) return;
        const query = input;
        input = '';

        chat.addUserMessage(query);
        chat.startStream();

        try {
            const resp = await fetch(`/api/chat/stream/${data.session.id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    query,
                    collection_names: [selectedCollection],
                    crossdoc: chat.crossdoc,
                }),
            });

            const reader = resp.body!.getReader();
            const decoder = new TextDecoder();

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                const text = decoder.decode(value);
                // Parse SSE events
                for (const line of text.split('\n')) {
                    if (line.startsWith('data: ')) {
                        const data_str = line.slice(6);
                        if (data_str === '[DONE]') continue;
                        try {
                            const event = JSON.parse(data_str);
                            if (event.choices?.[0]?.delta?.content) {
                                chat.appendToken(event.choices[0].delta.content);
                            }
                            if (event.citations) {
                                chat.setSources(event.citations.map((c: any) => ({
                                    document: c.source_name ?? c.document ?? '',
                                    page: c.page,
                                    excerpt: c.content ?? c.excerpt ?? '',
                                })));
                            }
                        } catch { /* ignore parse errors */ }
                    }
                }
            }
        } finally {
            chat.finalizeStream();
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    }
</script>

<div class="flex h-screen overflow-hidden">

    <!-- Panel izquierdo: historial -->
    <div class="w-40 bg-[#0c1220] border-r border-[#1e293b] flex flex-col p-2 gap-1 overflow-y-auto">
        <div class="text-[9px] text-[#475569] font-semibold uppercase tracking-wide mb-1">
            Historial
        </div>
        <a href="/chat" class="flex items-center gap-1.5 text-[#6366f1] text-[9px] mb-2 hover:underline">
            <MessageSquare size={10} /> Nueva consulta
        </a>
        {#each data.history as session}
            <a
                href="/chat/{session.id}"
                class="bg-[#1e293b] rounded p-1.5 block hover:bg-[#334155] transition-colors
                       {session.id === data.session.id ? 'border-l-2 border-[#6366f1]' : ''}"
            >
                <div class="text-[8px] text-[#94a3b8] font-medium truncate">{session.title}</div>
                <div class="text-[7px] text-[#475569] mt-0.5">{session.updated_at.slice(0,10)}</div>
            </a>
        {/each}
    </div>

    <!-- Panel central: conversación -->
    <div class="flex-1 flex flex-col border-r border-[#1e293b] min-w-0">
        <!-- Header -->
        <div class="flex items-center gap-2 px-3 py-2 border-b border-[#1e293b] text-[9px]">
            <select
                bind:value={selectedCollection}
                class="bg-[#1e293b] border border-[#334155] rounded px-2 py-0.5 text-[#e2e8f0]"
            >
                {#each data.collections as col}
                    <option value={col}>{col}</option>
                {/each}
            </select>
            <label class="flex items-center gap-1 text-[#64748b] cursor-pointer">
                <input type="checkbox" bind:checked={chat.crossdoc} class="accent-[#6366f1]" />
                Crossdoc
            </label>
        </div>

        <!-- Messages -->
        <div class="flex-1 overflow-y-auto p-3 flex flex-col gap-3">
            {#each chat.messages as msg}
                {#if msg.role === 'user'}
                    <div class="flex justify-end">
                        <div class="bg-[#4338ca] rounded-lg rounded-tr-sm px-3 py-2 max-w-[70%]">
                            <p class="text-[11px] text-[#e0e7ff]">{msg.content}</p>
                        </div>
                    </div>
                {:else}
                    <div class="flex gap-2">
                        <div class="w-5 h-5 bg-[#6366f1] rounded-full flex-shrink-0 mt-0.5"></div>
                        <div class="bg-[#1e293b] border border-[#334155] rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%]">
                            <p class="text-[11px] text-[#cbd5e1] leading-relaxed whitespace-pre-wrap">
                                {msg.content}
                            </p>
                        </div>
                    </div>
                {/if}
            {/each}

            <!-- Streaming -->
            {#if chat.streaming}
                <div class="flex gap-2">
                    <div class="w-5 h-5 bg-[#6366f1] rounded-full flex-shrink-0 mt-0.5 animate-pulse"></div>
                    <div class="bg-[#1e293b] border border-[#334155] rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%]">
                        <p class="text-[11px] text-[#cbd5e1] leading-relaxed whitespace-pre-wrap">
                            {chat.streamingContent}<span class="animate-pulse">▋</span>
                        </p>
                    </div>
                </div>
            {/if}
        </div>

        <!-- Input -->
        <div class="p-3 border-t border-[#1e293b]">
            <div class="flex gap-2 bg-[#1e293b] border border-[#334155] rounded-lg px-3 py-2">
                <textarea
                    bind:value={input}
                    onkeydown={handleKeydown}
                    rows={1}
                    placeholder="Escribí tu consulta..."
                    class="flex-1 bg-transparent text-[11px] text-[#e2e8f0] placeholder-[#475569]
                           resize-none outline-none"
                ></textarea>
                <button
                    onclick={sendMessage}
                    disabled={chat.streaming || !input.trim()}
                    class="text-[#6366f1] hover:text-[#a5b4fc] disabled:opacity-40 transition-colors"
                >
                    {#if chat.streaming}
                        <RefreshCw size={16} class="animate-spin" />
                    {:else}
                        <Send size={16} />
                    {/if}
                </button>
            </div>
        </div>
    </div>

    <!-- Panel derecho: fuentes -->
    <div class="w-48 bg-[#0c1220] p-3 overflow-y-auto">
        <div class="text-[9px] text-[#475569] font-semibold uppercase tracking-wide mb-2">
            Fuentes ({chat.sources.length})
        </div>
        {#each chat.sources as source, i}
            <div class="bg-[#1e293b] rounded p-2 mb-2 border-l-2
                        border-{i === 0 ? '[#6366f1]' : i === 1 ? '[#4338ca]' : '[#334155]'}">
                <div class="text-[8px] text-[#a5b4fc] font-semibold truncate">{source.document}</div>
                {#if source.page}
                    <div class="text-[7px] text-[#475569]">p. {source.page}</div>
                {/if}
                <div class="text-[7px] text-[#64748b] mt-1 line-clamp-3">{source.excerpt}</div>
            </div>
        {/each}
    </div>

</div>
```

- [ ] **Verify chat renders**

```bash
npm run dev
# Navigate to /chat → should redirect to /chat/<new-id>
# Should see 3-panel layout (history | conversation | sources)
# (Actual streaming requires gateway running)
```

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/routes/(app)/chat/ \
        services/sda-frontend/src/routes/api/chat/ \
        services/sda-frontend/src/lib/stores/
git commit -m "feat: add chat module (SSE streaming, 3-panel split view)"
```

---

### Task 6: Collections module

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/collections/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/collections/+page.server.ts`
- Create: `services/sda-frontend/src/routes/(app)/collections/[name]/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/collections/[name]/+page.server.ts`

- [ ] **Collections list page**

`+page.server.ts`:

```typescript
import type { PageServerLoad } from './$types';
import { gatewayListCollections, gatewayCollectionStats } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
    const { collections } = await gatewayListCollections();
    const statsResults = await Promise.allSettled(
        collections.map(name => gatewayCollectionStats(name))
    );
    const stats = Object.fromEntries(
        collections.map((name, i) => [
            name,
            statsResults[i].status === 'fulfilled' ? statsResults[i].value : null
        ])
    );
    return { collections, stats };
};
```

`+page.svelte`:

```svelte
<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6">
    <h1 class="text-lg font-semibold text-[#e2e8f0] mb-4">Colecciones</h1>
    <div class="grid grid-cols-3 gap-4">
        {#each data.collections as name}
            <a href="/collections/{name}"
               class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4
                      hover:border-[#6366f1] transition-colors block">
                <div class="text-sm font-medium text-[#e2e8f0] mb-2">{name}</div>
                {#if data.stats[name]}
                    <div class="text-xs text-[#475569]">
                        {data.stats[name].entity_count?.toLocaleString() ?? '—'} entidades
                    </div>
                {/if}
            </a>
        {/each}
    </div>
</div>
```

- [ ] **Collection detail page**

`[name]/+page.server.ts`:

```typescript
import type { PageServerLoad } from './$types';
import { gatewayCollectionStats } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params }) => {
    const stats = await gatewayCollectionStats(params.name);
    return { name: params.name, stats };
};
```

`[name]/+page.svelte`:

```svelte
<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6 max-w-2xl">
    <div class="flex items-center gap-3 mb-6">
        <a href="/collections" class="text-[#475569] hover:text-[#94a3b8] text-sm">← Colecciones</a>
        <h1 class="text-lg font-semibold text-[#e2e8f0]">{data.name}</h1>
    </div>

    <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4 mb-4">
        <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
                <div class="text-[#475569] text-xs mb-1">Entidades</div>
                <div class="text-[#e2e8f0] font-medium">
                    {data.stats?.entity_count?.toLocaleString() ?? '—'}
                </div>
            </div>
            <div>
                <div class="text-[#475569] text-xs mb-1">Documentos</div>
                <div class="text-[#e2e8f0] font-medium">
                    {data.stats?.document_count?.toLocaleString() ?? '—'}
                </div>
            </div>
        </div>
    </div>

    <a href="/chat?collection={data.name}"
       class="inline-block bg-[#6366f1] hover:bg-[#4f46e5] text-white text-sm
              px-4 py-2 rounded transition-colors">
        Consultar esta colección
    </a>
</div>
```

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/routes/(app)/collections/
git commit -m "feat: add collections module (list + detail with stats)"
```

---

### Task 7: Admin + Audit + Settings modules

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/admin/+layout.svelte`
- Create: `services/sda-frontend/src/routes/(app)/admin/users/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/admin/users/+page.server.ts`
- Create: `services/sda-frontend/src/routes/(app)/admin/areas/[id]/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/admin/areas/[id]/+page.server.ts`
- Create: `services/sda-frontend/src/routes/(app)/audit/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/audit/+page.server.ts`
- Create: `services/sda-frontend/src/routes/(app)/settings/+page.svelte`
- Create: `services/sda-frontend/src/routes/(app)/settings/+page.server.ts`

- [ ] **Admin layout guard**

`src/routes/(app)/admin/+layout.svelte`:

```svelte
<script lang="ts">
    import { page } from '$app/stores';
    let { children, data } = $props();
    const user = $derived(data.user);
    const allowed = $derived(user.role === 'admin' || user.role === 'area_manager');
</script>

{#if allowed}
    {@render children()}
{:else}
    <div class="p-6 text-[#475569]">Acceso denegado.</div>
{/if}
```

- [ ] **Admin users page**

`+page.server.ts` (simplified — full CRUD via form actions):

```typescript
import type { PageServerLoad, Actions } from './$types';
import { fail } from '@sveltejs/kit';
import { gatewayListUsers, gatewayCreateUser, gatewayDeleteUser,
         gatewayListAreas } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
    const [usersData, areasData] = await Promise.all([
        gatewayListUsers(),
        gatewayListAreas(),
    ]);
    return { users: usersData.users, areas: areasData.areas };
};

export const actions: Actions = {
    create: async ({ request }) => {
        const data = await request.formData();
        try {
            const result = await gatewayCreateUser({
                email: data.get('email') as string,
                name: data.get('name') as string,
                area_id: Number(data.get('area_id')),
                role: data.get('role') as string,
                password: data.get('password') as string,
            });
            return { success: true, api_key: result.api_key };
        } catch (e: any) {
            return fail(400, { error: e.detail ?? 'Error creating user' });
        }
    },
    delete: async ({ request }) => {
        const data = await request.formData();
        await gatewayDeleteUser(Number(data.get('id')));
        return { success: true };
    }
};
```

`+page.svelte` (table + modals):

```svelte
<script lang="ts">
    import { enhance } from '$app/forms';
    let { data, form } = $props();
    let showCreate = $state(false);
</script>

<div class="p-6">
    <div class="flex items-center justify-between mb-4">
        <h1 class="text-lg font-semibold text-[#e2e8f0]">Usuarios</h1>
        <button
            onclick={() => showCreate = true}
            class="bg-[#6366f1] hover:bg-[#4f46e5] text-white text-sm px-3 py-1.5 rounded"
        >
            + Nuevo usuario
        </button>
    </div>

    {#if form?.success && form?.api_key}
        <div class="bg-[#065f46] text-[#6ee7b7] p-3 rounded mb-4 text-sm">
            Usuario creado. API key: <code class="font-mono">{form.api_key}</code>
        </div>
    {/if}

    <table class="w-full text-sm">
        <thead>
            <tr class="text-[#475569] text-xs border-b border-[#1e293b]">
                <th class="text-left pb-2">Email</th>
                <th class="text-left pb-2">Nombre</th>
                <th class="text-left pb-2">Área</th>
                <th class="text-left pb-2">Rol</th>
                <th class="text-left pb-2">Estado</th>
                <th class="pb-2"></th>
            </tr>
        </thead>
        <tbody>
            {#each data.users as user}
                <tr class="border-b border-[#1e293b] text-[#94a3b8]">
                    <td class="py-2">{user.email}</td>
                    <td class="py-2">{user.name}</td>
                    <td class="py-2">{data.areas.find(a => a.id === user.area_id)?.name ?? user.area_id}</td>
                    <td class="py-2">{user.role}</td>
                    <td class="py-2">
                        <span class="text-xs {user.active ? 'text-[#4ade80]' : 'text-[#f87171]'}">
                            {user.active ? 'Activo' : 'Inactivo'}
                        </span>
                    </td>
                    <td class="py-2 text-right">
                        <form method="POST" action="?/delete" use:enhance class="inline">
                            <input type="hidden" name="id" value={user.id} />
                            <button type="submit" class="text-xs text-[#f87171] hover:underline">
                                Desactivar
                            </button>
                        </form>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>

    {#if showCreate}
        <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
            <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-6 w-96">
                <h2 class="text-sm font-semibold text-[#e2e8f0] mb-4">Nuevo usuario</h2>
                <form method="POST" action="?/create" use:enhance class="flex flex-col gap-3"
                      onsubmit={() => showCreate = false}>
                    {#each [['email','Email','email'],['name','Nombre','text'],['password','Contraseña','password']] as [name,label,type]}
                        <div>
                            <label class="text-xs text-[#475569]">{label}</label>
                            <input {name} {type} required
                                   class="w-full mt-0.5 bg-[#1e293b] border border-[#334155] rounded
                                          px-2 py-1 text-sm text-[#e2e8f0] focus:outline-none focus:border-[#6366f1]" />
                        </div>
                    {/each}
                    <div>
                        <label class="text-xs text-[#475569]">Área</label>
                        <select name="area_id" class="w-full mt-0.5 bg-[#1e293b] border border-[#334155]
                                                       rounded px-2 py-1 text-sm text-[#e2e8f0]">
                            {#each data.areas as area}
                                <option value={area.id}>{area.name}</option>
                            {/each}
                        </select>
                    </div>
                    <div>
                        <label class="text-xs text-[#475569]">Rol</label>
                        <select name="role" class="w-full mt-0.5 bg-[#1e293b] border border-[#334155]
                                                    rounded px-2 py-1 text-sm text-[#e2e8f0]">
                            <option value="user">Usuario</option>
                            <option value="area_manager">Gestor de Área</option>
                            <option value="admin">Admin</option>
                        </select>
                    </div>
                    <div class="flex gap-2 mt-2">
                        <button type="submit"
                                class="flex-1 bg-[#6366f1] text-white text-sm py-1.5 rounded">
                            Crear
                        </button>
                        <button type="button" onclick={() => showCreate = false}
                                class="flex-1 bg-[#1e293b] text-[#94a3b8] text-sm py-1.5 rounded">
                            Cancelar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    {/if}
</div>
```

- [ ] **Audit page**

`audit/+page.server.ts`:

```typescript
import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { gatewayGetAudit } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals, url }) => {
    if (locals.user?.role !== 'admin') throw redirect(302, '/chat');
    const params = {
        user_id: url.searchParams.get('user_id') ? Number(url.searchParams.get('user_id')) : undefined,
        action: url.searchParams.get('action') ?? undefined,
        collection: url.searchParams.get('collection') ?? undefined,
        limit: 100,
    };
    const data = await gatewayGetAudit(params);
    return { entries: data.entries };
};
```

`audit/+page.svelte` (simple table with filter bar):

```svelte
<script lang="ts">
    let { data } = $props();
</script>

<div class="p-6">
    <h1 class="text-lg font-semibold text-[#e2e8f0] mb-4">Auditoría</h1>
    <table class="w-full text-xs">
        <thead>
            <tr class="text-[#475569] border-b border-[#1e293b]">
                <th class="text-left pb-2">Timestamp</th>
                <th class="text-left pb-2">User ID</th>
                <th class="text-left pb-2">Acción</th>
                <th class="text-left pb-2">Colección</th>
                <th class="text-left pb-2">Preview</th>
            </tr>
        </thead>
        <tbody>
            {#each data.entries as entry}
                <tr class="border-b border-[#1e293b] text-[#64748b]">
                    <td class="py-1.5">{entry.timestamp}</td>
                    <td class="py-1.5">{entry.user_id}</td>
                    <td class="py-1.5">{entry.action}</td>
                    <td class="py-1.5">{entry.collection ?? '—'}</td>
                    <td class="py-1.5 max-w-xs truncate">{entry.query_preview ?? '—'}</td>
                </tr>
            {/each}
        </tbody>
    </table>
</div>
```

- [ ] **Settings page**

`settings/+page.server.ts`:

```typescript
import type { PageServerLoad, Actions } from './$types';
import { gatewayRefreshKey } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ locals }) => {
    return { user: locals.user };
};

export const actions: Actions = {
    refresh_key: async ({ locals }) => {
        const result = await gatewayRefreshKey(locals.user!.id);
        return { api_key: result.api_key };
    }
};
```

`settings/+page.svelte`:

```svelte
<script lang="ts">
    import { enhance } from '$app/forms';
    let { data, form } = $props();
</script>

<div class="p-6 max-w-lg">
    <h1 class="text-lg font-semibold text-[#e2e8f0] mb-4">Configuración</h1>

    <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4 mb-4">
        <div class="text-xs text-[#475569] mb-3">Perfil</div>
        <div class="text-sm text-[#e2e8f0]">{data.user?.name}</div>
        <div class="text-xs text-[#64748b]">{data.user?.email}</div>
        <div class="text-xs text-[#6366f1] mt-1">{data.user?.role}</div>
    </div>

    <div class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-4">
        <div class="text-xs text-[#475569] mb-3">API Key personal</div>
        {#if form?.api_key}
            <div class="bg-[#065f46] text-[#6ee7b7] p-2 rounded text-xs font-mono mb-3 break-all">
                {form.api_key}
            </div>
        {/if}
        <form method="POST" action="?/refresh_key" use:enhance>
            <button type="submit"
                    class="bg-[#1e293b] hover:bg-[#334155] text-[#94a3b8] text-xs px-3 py-1.5 rounded">
                Regenerar API key
            </button>
        </form>
    </div>
</div>
```

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add services/sda-frontend/src/routes/(app)/admin/ \
        services/sda-frontend/src/routes/(app)/audit/ \
        services/sda-frontend/src/routes/(app)/settings/
git commit -m "feat: add admin, audit, and settings modules"
```

---

### Task 8: Docker + compose integration + deploy

**Files:**
- Modify: `config/compose-platform-services.yaml`
- Modify: `config/profiles/brev-2gpu.yaml`
- Modify: `config/profiles/workstation-1gpu.yaml`
- Modify: `config/.env.saldivia`

- [ ] **Add sda-frontend to compose**

In `config/compose-platform-services.yaml`, add service:

```yaml
  sda-frontend:
    build:
      context: ${SALDIVIA_ROOT}
      dockerfile: services/sda-frontend/Dockerfile
    environment:
      - GATEWAY_URL=http://auth-gateway:8090
      - JWT_SECRET=${JWT_SECRET}
      - SYSTEM_API_KEY=${SYSTEM_API_KEY}
      - ORIGIN=${SDA_ORIGIN:-http://localhost:3000}
      - PUBLIC_APP_NAME=SDA
      - NODE_ENV=production
    ports:
      - "3000:3000"
    depends_on:
      - auth-gateway
    networks:
      - default
    restart: unless-stopped
```

- [ ] **Add SDA_ORIGIN to .env.saldivia**

```bash
SDA_ORIGIN=https://sda.tecpia.local
```

- [ ] **Test Docker build locally**

```bash
cd ~/rag-saldivia/services/sda-frontend
docker build -t sda-frontend-test .
docker run --rm -e GATEWAY_URL=http://host.docker.internal:9000 \
           -e JWT_SECRET=test \
           -e SYSTEM_API_KEY=test \
           -p 3000:3000 sda-frontend-test
```

Expected: `http://localhost:3000` serves the login page

- [ ] **Deploy to Brev**

```bash
ssh nvidia-enterprise-rag-deb106
cd ~/rag-saldivia && git pull && make deploy PROFILE=brev-2gpu
```

- [ ] **Smoke test on Brev**

```bash
# Check frontend container is running
docker ps | grep sda-frontend

# Check logs
docker logs sda-frontend 2>&1 | tail -20

# Curl login page
curl -s http://localhost:3000/login | grep -o "<title>.*</title>"
```

Expected: container running, login page returns HTML

- [ ] **Set first admin password (if not done in Plan A)**

```bash
ssh nvidia-enterprise-rag-deb106
docker exec auth-gateway python3 -c "
from saldivia.auth.database import AuthDB
from saldivia.auth.models import hash_password
db = AuthDB()
db.set_password(1, hash_password('changeme'))
print('Admin password set')
"
```

- [ ] **End-to-end login test**

Open browser → `http://<brev-host>:3000` → Login with admin credentials → Should land on `/chat`.

- [ ] **Commit**

```bash
cd ~/rag-saldivia
git add config/compose-platform-services.yaml config/.env.saldivia
git commit -m "feat: add sda-frontend to Docker Compose stack"
```

---

**SDA Frontend complete.** Full platform is live on Brev at port 3000.

**Next steps (not in scope):**
- Add areas page (`admin/areas/[id]`) — same pattern as users page
- Password change flow for users
- Ingest upload UI in collections detail page
- Cloudflare Tunnel for external access
