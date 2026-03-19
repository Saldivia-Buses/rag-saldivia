# Fase 4 — Colecciones Pro + Upload Básico — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** UI completa para gestión de colecciones (CRUD con modales, stats, badges) y primera versión funcional de upload de documentos con drag & drop.

**Architecture:** Se expanden las rutas de collections existentes con componentes reutilizables (`CollectionCard`, `CreateModal`, `DeleteModal`), se añaden BFF endpoints (`/api/collections` GET/POST, `/api/collections/[name]` DELETE, `/api/upload` POST), una `CollectionsStore` Svelte 5 runes para estado cliente (requerida por Fase 17), y se reemplaza el stub de upload con una página funcional. El upload multipart se maneja inline en el BFF sin pasar por el helper `gw()` (que fuerza JSON).

**Tech Stack:** SvelteKit 5, Svelte 5 runes, TypeScript, Vitest, Tailwind CSS v4, CSS variables de diseño, lucide-svelte

---

## Contexto del codebase

- **Repo local:** `~/rag-saldivia/`
- **Frontend:** `services/sda-frontend/`
- **Correr tests:** `cd ~/rag-saldivia/services/sda-frontend && npm test`
- **Env de tests:** `node` (vitest.config.ts), NO hay jsdom — los tests son pure TS
- **Patrón de API routes:** `export const POST: RequestHandler = async ({ request, locals }) => {}`; `locals.user` viene del `+layout.server.ts` que hace el guard de auth
- **Patrón de gateway:** `gw<T>(path, init?)` helper con timeout y error handling; las funciones se exportan desde `$lib/server/gateway`
- **CSS vars disponibles:** `--bg-surface`, `--bg-base`, `--bg-hover`, `--border`, `--accent`, `--text`, `--text-muted`, `--text-faint`, `--radius-lg`, `--success`, `--success-bg`, `--danger`, `--danger-bg`, `--info`, `--info-bg`
- **Componentes UI disponibles:** `Badge`, `Button`, `Card`, `Input`, `Modal`, `Skeleton`, `Toast`, `ToastContainer`
- **Test pre-existente que falla:** `markdown.test.ts` falla por resolver `marked` — ignorar, es pre-existente

## Archivos a crear/modificar

| Archivo | Acción |
|---------|--------|
| `services/sda-frontend/src/lib/server/gateway.ts` | Modificar: add `gatewayCreateCollection`, `gatewayDeleteCollection`; extend `CollectionStats` |
| `services/sda-frontend/src/lib/stores/collections.svelte.ts` | Crear: `CollectionsStore` singleton |
| `services/sda-frontend/src/lib/stores/collections.svelte.test.ts` | Crear: tests del store |
| `services/sda-frontend/src/routes/api/collections/+server.ts` | Crear: GET (list) + POST (create) |
| `services/sda-frontend/src/routes/api/collections/[name]/+server.ts` | Crear: DELETE |
| `services/sda-frontend/src/routes/api/collections/collections.test.ts` | Crear: tests BFF collections |
| `services/sda-frontend/src/routes/api/upload/+server.ts` | Crear: POST multipart |
| `services/sda-frontend/src/routes/api/upload/upload.test.ts` | Crear: test BFF upload |
| `services/sda-frontend/src/routes/(app)/collections/_components/CollectionCard.svelte` | Crear |
| `services/sda-frontend/src/routes/(app)/collections/_components/CreateModal.svelte` | Crear |
| `services/sda-frontend/src/routes/(app)/collections/_components/DeleteModal.svelte` | Crear |
| `services/sda-frontend/src/routes/(app)/collections/+page.svelte` | Modificar: usar CollectionCard + CreateModal |
| `services/sda-frontend/src/routes/(app)/collections/[name]/+page.svelte` | Modificar: CSS vars + DeleteModal |
| `services/sda-frontend/src/routes/(app)/collections/[name]/+page.server.ts` | Modificar: agregar action delete |
| `services/sda-frontend/src/routes/(app)/upload/+page.svelte` | Modificar: reemplazar stub con drag & drop |
| `services/sda-frontend/src/routes/(app)/upload/+page.server.ts` | Crear: load collections |

---

## Task 1: Extender gateway.ts

**Files:**
- Modify: `services/sda-frontend/src/lib/server/gateway.ts`

Agregar `index_type` y `has_sparse` a `CollectionStats`, y dos nuevas funciones.

- [ ] **Step 1: Agregar campos a `CollectionStats` interface**

Encontrar la interfaz `CollectionStats` en el archivo (línea ~152) y reemplazarla:

```typescript
export interface CollectionStats {
    collection: string; entity_count: number; document_count?: number;
    index_type?: string; has_sparse?: boolean;
}
```

- [ ] **Step 2: Agregar `gatewayCreateCollection` y `gatewayDeleteCollection`**

Después de la función `gatewayCollectionStats` (línea ~81), agregar:

```typescript
export async function gatewayCreateCollection(name: string, schema = 'default') {
    return gw<{ name: string }>(
        '/v1/collections',
        { method: 'POST', body: JSON.stringify({ name, schema }) }
    );
}

export async function gatewayDeleteCollection(name: string) {
    return gw<{ ok: boolean }>(`/v1/collections/${name}`, { method: 'DELETE' });
}
```

- [ ] **Step 3: Verificar que TypeScript no tiene errores**

```bash
cd ~/rag-saldivia/services/sda-frontend && npx svelte-check --tsconfig tsconfig.json 2>&1 | grep -E "Error|error" | head -20
```

Expected: 0 errores en gateway.ts

- [ ] **Step 4: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/lib/server/gateway.ts
git commit -m "feat(collections): extend gateway with create/delete collection functions"
```

---

## Task 2: CollectionsStore

**Files:**
- Create: `services/sda-frontend/src/lib/stores/collections.svelte.ts`
- Test: `services/sda-frontend/src/lib/stores/collections.svelte.test.ts`

El store hace fetch al BFF (no directamente al gateway). Será usado por Fase 17 (CommandPalette) para buscar colecciones client-side.

- [ ] **Step 1: Escribir el test (failing)**

Crear `services/sda-frontend/src/lib/stores/collections.svelte.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CollectionsStore } from './collections.svelte.js';

describe('CollectionsStore', () => {
    let store: CollectionsStore;

    beforeEach(() => {
        store = new CollectionsStore();
        vi.resetAllMocks();
    });

    it('load() popula collections desde /api/collections', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ collections: ['col-a', 'col-b'] }),
        }));
        await store.load();
        expect(store.collections).toEqual(['col-a', 'col-b']);
        expect(store.loading).toBe(false);
    });

    it('load() con error fetch no rompe — deja collections vacías', async () => {
        vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('network error')));
        await store.load();
        expect(store.collections).toEqual([]);
        expect(store.loading).toBe(false);
    });

    it('create() llama POST /api/collections y agrega a la lista', async () => {
        store.collections = ['existing'];
        const mockFetch = vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ name: 'new-col' }),
        });
        vi.stubGlobal('fetch', mockFetch);
        await store.create('new-col', 'default');
        expect(mockFetch).toHaveBeenCalledWith('/api/collections', expect.objectContaining({
            method: 'POST',
        }));
        expect(store.collections).toContain('new-col');
    });

    it('delete() llama DELETE /api/collections/[name] y quita de la lista', async () => {
        store.collections = ['col-a', 'col-b'];
        const mockFetch = vi.fn().mockResolvedValue({
            ok: true,
            json: () => Promise.resolve({ ok: true }),
        });
        vi.stubGlobal('fetch', mockFetch);
        await store.delete('col-a');
        expect(mockFetch).toHaveBeenCalledWith('/api/collections/col-a', expect.objectContaining({
            method: 'DELETE',
        }));
        expect(store.collections).toEqual(['col-b']);
    });

    it('init() hidrata colecciones desde datos del servidor', () => {
        store.init(['a', 'b', 'c']);
        expect(store.collections).toEqual(['a', 'b', 'c']);
    });
});
```

- [ ] **Step 2: Correr test — verificar que falla**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test -- collections.svelte.test 2>&1 | tail -15
```

Expected: FAIL — "Cannot find module './collections.svelte.js'"

- [ ] **Step 3: Implementar el store**

Crear `services/sda-frontend/src/lib/stores/collections.svelte.ts`:

```typescript
// Svelte 5 runes-based store for collections state.
// Used by CommandPalette (Fase 17) for client-side collection search.

export class CollectionsStore {
    collections = $state<string[]>([]);
    loading = $state(false);

    /** Hydrate from server-loaded data (call from +page.svelte) */
    init(collections: string[]) {
        this.collections = collections;
    }

    /** Client-side refresh from BFF */
    async load(): Promise<void> {
        this.loading = true;
        try {
            const res = await fetch('/api/collections');
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const data = await res.json();
            this.collections = data.collections ?? [];
        } catch {
            this.collections = [];
        } finally {
            this.loading = false;
        }
    }

    async create(name: string, schema = 'default'): Promise<void> {
        const res = await fetch('/api/collections', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, schema }),
        });
        if (!res.ok) {
            const err = await res.json().catch(() => ({}));
            throw new Error(err.message ?? `Error ${res.status}`);
        }
        this.collections = [...this.collections, name];
    }

    async delete(name: string): Promise<void> {
        const res = await fetch(`/api/collections/${name}`, { method: 'DELETE' });
        if (!res.ok) {
            const err = await res.json().catch(() => ({}));
            throw new Error(err.message ?? `Error ${res.status}`);
        }
        this.collections = this.collections.filter(c => c !== name);
    }
}

export const collectionsStore = new CollectionsStore();
```

- [ ] **Step 4: Correr test — verificar que pasa**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test -- collections.svelte.test 2>&1 | tail -15
```

Expected: 5 tests PASS

- [ ] **Step 5: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/lib/stores/collections.svelte.ts src/lib/stores/collections.svelte.test.ts
git commit -m "feat(collections): add CollectionsStore with load/create/delete"
```

---

## Task 3: API BFF collections routes + tests

**Files:**
- Create: `services/sda-frontend/src/routes/api/collections/+server.ts`
- Create: `services/sda-frontend/src/routes/api/collections/[name]/+server.ts`
- Create: `services/sda-frontend/src/routes/api/collections/collections.test.ts`

- [ ] **Step 1: Escribir los tests (failing)**

Crear `services/sda-frontend/src/routes/api/collections/collections.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGateway = {
    gatewayListCollections: vi.fn(),
    gatewayCreateCollection: vi.fn(),
    gatewayDeleteCollection: vi.fn(),
    GatewayError: class GatewayError extends Error {
        status: number; detail: string;
        constructor(status: number, detail: string) { super(detail); this.status = status; this.detail = detail; }
    },
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('GET /api/collections', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 200 with collections list', async () => {
        mockGateway.gatewayListCollections.mockResolvedValue({ collections: ['col-a', 'col-b'] });
        const { GET } = await import('./+server.js');
        const event = { locals: { user: { id: 1 } } } as any;
        const res = await GET(event);
        expect(res.status).toBe(200);
        const body = await res.json();
        expect(body.collections).toEqual(['col-a', 'col-b']);
    });

    it('returns 401 when not authenticated', async () => {
        const { GET } = await import('./+server.js');
        const event = { locals: { user: null } } as any;
        await expect(GET(event)).rejects.toMatchObject({ status: 401 });
    });
});

describe('POST /api/collections', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 201 with created collection name', async () => {
        mockGateway.gatewayCreateCollection.mockResolvedValue({ name: 'nueva-col' });
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost/api/collections', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: 'nueva-col', schema: 'default' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        expect(res.status).toBe(201);
        const body = await res.json();
        expect(body.name).toBe('nueva-col');
    });

    it('returns 400 when name is missing', async () => {
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost/api/collections', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ schema: 'default' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 400 });
    });
});

describe('DELETE /api/collections/[name]', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 204 on successful delete', async () => {
        mockGateway.gatewayDeleteCollection.mockResolvedValue({ ok: true });
        const { DELETE } = await import('./[name]/+server.js');
        const event = {
            params: { name: 'col-a' },
            locals: { user: { id: 1 } },
        } as any;
        const res = await DELETE(event);
        expect(res.status).toBe(204);
    });

    it('returns 401 when not authenticated', async () => {
        const { DELETE } = await import('./[name]/+server.js');
        const event = { params: { name: 'col-a' }, locals: { user: null } } as any;
        await expect(DELETE(event)).rejects.toMatchObject({ status: 401 });
    });
});
```

- [ ] **Step 2: Correr test — verificar que falla**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test -- collections.test 2>&1 | tail -15
```

Expected: FAIL — "Cannot find module './+server.js'"

- [ ] **Step 3: Implementar GET + POST handler**

Crear `services/sda-frontend/src/routes/api/collections/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { gatewayListCollections, gatewayCreateCollection, GatewayError } from '$lib/server/gateway';
import { json, error } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ locals }) => {
    if (!locals.user) throw error(401);
    try {
        const data = await gatewayListCollections();
        return json(data);
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudieron cargar las colecciones.');
    }
};

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);
    const body = await request.json().catch(() => ({}));
    const { name, schema = 'default' } = body as { name?: string; schema?: string };
    if (!name?.trim()) throw error(400, 'El nombre de la colección es requerido.');
    try {
        const result = await gatewayCreateCollection(name.trim(), schema);
        return json(result, { status: 201 });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, 'No se pudo crear la colección.');
    }
};
```

- [ ] **Step 4: Implementar DELETE handler**

Crear `services/sda-frontend/src/routes/api/collections/[name]/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { gatewayDeleteCollection, GatewayError } from '$lib/server/gateway';
import { error } from '@sveltejs/kit';

export const DELETE: RequestHandler = async ({ params, locals }) => {
    if (!locals.user) throw error(401);
    try {
        await gatewayDeleteCollection(params.name);
        return new Response(null, { status: 204 });
    } catch (err) {
        const status = err instanceof GatewayError ? err.status : 503;
        throw error(status, `No se pudo eliminar la colección "${params.name}".`);
    }
};
```

- [ ] **Step 5: Correr tests — verificar que pasan**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test -- collections.test 2>&1 | tail -15
```

Expected: 5 tests PASS

- [ ] **Step 6: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/routes/api/collections/
git commit -m "feat(collections): add BFF GET/POST /api/collections + DELETE /api/collections/[name]"
```

---

## Task 4: API BFF upload route + test

**Files:**
- Create: `services/sda-frontend/src/routes/api/upload/+server.ts`
- Create: `services/sda-frontend/src/routes/api/upload/upload.test.ts`

El upload forward multipart al gateway usando `fetch` crudo (no el helper `gw()` que fuerza JSON). Necesita `X-User-Id` header además del Bearer token.

- [ ] **Step 1: Escribir el test (failing)**

Crear `services/sda-frontend/src/routes/api/upload/upload.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('POST /api/upload', () => {
    beforeEach(() => vi.resetAllMocks());

    it('forwards multipart al gateway y retorna su respuesta', async () => {
        process.env.GATEWAY_URL = 'http://gateway:9000';
        process.env.SYSTEM_API_KEY = 'test-key';

        const mockFetch = vi.fn().mockResolvedValue({
            ok: true,
            status: 202,
            json: () => Promise.resolve({ job_id: 'abc123', status: 'queued' }),
        });
        vi.stubGlobal('fetch', mockFetch);

        const { POST } = await import('./+server.js');

        const formData = new FormData();
        formData.append('file', new File(['contenido pdf'], 'doc.pdf', { type: 'application/pdf' }));
        formData.append('collection', 'mi-coleccion');

        const event = {
            request: new Request('http://localhost/api/upload', {
                method: 'POST',
                body: formData,
            }),
            locals: { user: { id: 42, email: 'test@test.com', role: 'user', area_id: 1, name: 'Test' } },
        } as any;

        const res = await POST(event);
        expect(res.status).toBe(202);

        // Verificar que llamó al gateway con los headers correctos
        expect(mockFetch).toHaveBeenCalledWith(
            'http://gateway:9000/v1/documents',
            expect.objectContaining({
                method: 'POST',
                headers: expect.objectContaining({
                    'Authorization': 'Bearer test-key',
                    'X-User-Id': '42',
                }),
            })
        );
    });

    it('retorna 401 cuando no hay sesión', async () => {
        const { POST } = await import('./+server.js');
        const formData = new FormData();
        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: null },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 401 });
    });

    it('retorna 400 cuando falta archivo o collection', async () => {
        const { POST } = await import('./+server.js');
        const formData = new FormData();
        formData.append('collection', 'mi-col');
        // No file
        const event = {
            request: new Request('http://localhost/api/upload', { method: 'POST', body: formData }),
            locals: { user: { id: 1 } },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 400 });
    });
});
```

- [ ] **Step 2: Correr test — verificar que falla**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test -- upload.test 2>&1 | tail -15
```

Expected: FAIL — "Cannot find module './+server.js'"

- [ ] **Step 3: Implementar el upload handler**

Crear `services/sda-frontend/src/routes/api/upload/+server.ts`:

```typescript
import type { RequestHandler } from './$types';
import { json, error } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401);

    const formData = await request.formData();
    const file = formData.get('file');
    const collection = formData.get('collection');

    if (!file || !(file instanceof File)) throw error(400, 'Se requiere un archivo.');
    if (!collection || typeof collection !== 'string' || !collection.trim()) {
        throw error(400, 'Se requiere seleccionar una colección.');
    }

    const gatewayUrl = process.env.GATEWAY_URL ?? 'http://localhost:9000';
    const apiKey = process.env.SYSTEM_API_KEY;
    if (!apiKey) throw error(503, 'SYSTEM_API_KEY no configurado.');

    const gw = new FormData();
    gw.append('file', file);
    gw.append('collection_name', collection.trim());

    const resp = await fetch(`${gatewayUrl}/v1/documents`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${apiKey}`,
            'X-User-Id': String(locals.user.id),
        },
        body: gw,
    });

    const body = await resp.json().catch(() => ({}));
    return json(body, { status: resp.status });
};
```

- [ ] **Step 4: Correr tests — verificar que pasan**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test -- upload.test 2>&1 | tail -15
```

Expected: 3 tests PASS

- [ ] **Step 5: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/routes/api/upload/
git commit -m "feat(upload): add BFF POST /api/upload multipart → gateway"
```

---

## Task 5: Componentes de colecciones

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/collections/_components/CollectionCard.svelte`
- Create: `services/sda-frontend/src/routes/(app)/collections/_components/CreateModal.svelte`
- Create: `services/sda-frontend/src/routes/(app)/collections/_components/DeleteModal.svelte`

- [ ] **Step 1: Crear CollectionCard.svelte**

Crear `services/sda-frontend/src/routes/(app)/collections/_components/CollectionCard.svelte`:

```svelte
<script lang="ts">
    import Badge from '$lib/components/ui/Badge.svelte';
    import Skeleton from '$lib/components/ui/Skeleton.svelte';
    import type { CollectionStats } from '$lib/server/gateway';

    interface Props {
        name: string;
        stats: CollectionStats | null;
        href: string;
    }
    let { name, stats, href }: Props = $props();
</script>

<a {href} class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-4
                  hover:border-[var(--accent)] transition-colors block group">
    <div class="flex items-start justify-between mb-3">
        <span class="text-sm font-semibold text-[var(--text)] truncate">{name}</span>
        {#if stats?.has_sparse}
            <Badge variant="blue">Sparse</Badge>
        {/if}
    </div>
    {#if stats}
        <div class="text-2xl font-bold text-[var(--text)] mb-1">
            {stats.entity_count.toLocaleString()}
        </div>
        <div class="text-xs text-[var(--text-faint)]">
            entidades{stats.index_type ? ` · ${stats.index_type}` : ''}
        </div>
    {:else}
        <Skeleton class="h-8 w-24 mb-1" />
        <Skeleton class="h-3 w-32" />
    {/if}
</a>
```

- [ ] **Step 2: Crear CreateModal.svelte**

Crear `services/sda-frontend/src/routes/(app)/collections/_components/CreateModal.svelte`:

```svelte
<script lang="ts">
    import Modal from '$lib/components/ui/Modal.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    interface Props {
        open?: boolean;
        oncreate?: (name: string, schema: string) => void;
        onclose?: () => void;
    }
    let { open = $bindable(false), oncreate, onclose }: Props = $props();

    let name = $state('');
    let schema = $state('default');
    let error = $state('');
    let loading = $state(false);

    const SCHEMAS = [
        { value: 'default', label: 'Default (dense vectors)' },
        { value: 'sparse', label: 'Sparse (hybrid search)' },
    ];

    function validate(): string {
        if (!name.trim()) return 'El nombre es requerido.';
        if (!/^[a-z0-9_-]+$/i.test(name.trim())) return 'Solo letras, números, guiones y guiones bajos.';
        if (name.trim().length > 64) return 'Máximo 64 caracteres.';
        return '';
    }

    async function handleSubmit() {
        error = validate();
        if (error) return;
        loading = true;
        try {
            await oncreate?.(name.trim(), schema);
            open = false;
            name = '';
            schema = 'default';
            error = '';
        } catch (e: any) {
            error = e.message ?? 'Error al crear la colección.';
        } finally {
            loading = false;
        }
    }

    function handleClose() {
        name = '';
        schema = 'default';
        error = '';
        open = false;
        onclose?.();
    }
</script>

<Modal bind:open title="Nueva colección" onclose={handleClose} size="sm">
    <div class="space-y-4">
        <div>
            <label class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
                Nombre
            </label>
            <Input
                bind:value={name}
                placeholder="ej: documentos-legales"
                onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && handleSubmit()}
            />
            {#if error}
                <p class="text-xs text-[var(--danger)] mt-1.5">{error}</p>
            {/if}
        </div>
        <div>
            <label class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
                Schema
            </label>
            <select
                bind:value={schema}
                class="w-full bg-[var(--bg-base)] border border-[var(--border)] rounded-lg
                       px-3 py-2 text-sm text-[var(--text)] focus:outline-none
                       focus:border-[var(--accent)] transition-colors"
            >
                {#each SCHEMAS as s}
                    <option value={s.value}>{s.label}</option>
                {/each}
            </select>
        </div>
    </div>

    {#snippet footer()}
        <Button variant="ghost" onclick={handleClose} disabled={loading}>Cancelar</Button>
        <Button onclick={handleSubmit} disabled={loading}>
            {loading ? 'Creando...' : 'Crear colección'}
        </Button>
    {/snippet}
</Modal>
```

- [ ] **Step 3: Crear DeleteModal.svelte**

Crear `services/sda-frontend/src/routes/(app)/collections/_components/DeleteModal.svelte`:

```svelte
<script lang="ts">
    import Modal from '$lib/components/ui/Modal.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    interface Props {
        open?: boolean;
        name?: string;
        onconfirm?: () => void;
        onclose?: () => void;
    }
    let { open = $bindable(false), name = '', onconfirm, onclose }: Props = $props();

    let loading = $state(false);

    async function handleConfirm() {
        loading = true;
        try {
            await onconfirm?.();
            open = false;
        } finally {
            loading = false;
        }
    }
</script>

<Modal bind:open title="Eliminar colección" onclose={onclose} size="sm">
    <p class="text-sm text-[var(--text-muted)]">
        ¿Estás seguro de que querés eliminar
        <span class="font-semibold text-[var(--text)]">"{name}"</span>?
        Esta acción no se puede deshacer.
    </p>

    {#snippet footer()}
        <Button variant="ghost" onclick={() => { open = false; onclose?.(); }} disabled={loading}>
            Cancelar
        </Button>
        <Button variant="danger" onclick={handleConfirm} disabled={loading}>
            {loading ? 'Eliminando...' : 'Eliminar'}
        </Button>
    {/snippet}
</Modal>
```

**Nota:** Si `Button` no tiene variante `danger`, usar clase directa:
```svelte
<button
    onclick={handleConfirm}
    disabled={loading}
    class="px-3 py-1.5 text-sm bg-[var(--danger)] text-white rounded-lg
           hover:opacity-90 transition-opacity disabled:opacity-50"
>
    {loading ? 'Eliminando...' : 'Eliminar'}
</button>
```

- [ ] **Step 4: Verificar que los componentes no tienen errores TS**

```bash
cd ~/rag-saldivia/services/sda-frontend && npx svelte-check --tsconfig tsconfig.json 2>&1 | grep -E "Error|error" | grep -v "markdown" | head -20
```

Expected: 0 errores en los nuevos archivos

- [ ] **Step 5: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/routes/\(app\)/collections/_components/
git commit -m "feat(collections): add CollectionCard, CreateModal, DeleteModal components"
```

---

## Task 6: Expandir collections index page

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/collections/+page.svelte`

El `+page.server.ts` ya carga correctamente la lista + stats — no necesita cambios.

- [ ] **Step 1: Reemplazar +page.svelte con grid mejorado**

Contenido actual (4 líneas) → reemplazar completamente:

```svelte
<script lang="ts">
    import CollectionCard from './_components/CollectionCard.svelte';
    import CreateModal from './_components/CreateModal.svelte';
    import { collectionsStore } from '$lib/stores/collections.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import { goto, invalidateAll } from '$app/navigation';
    import { Plus } from 'lucide-svelte';

    let { data } = $props();

    // Hidratar el store con datos del servidor (para Fase 17)
    $effect(() => {
        collectionsStore.init(data.collections);
    });

    let showCreate = $state(false);

    async function handleCreate(name: string, schema: string) {
        await collectionsStore.create(name, schema);
        toastStore.success(`Colección "${name}" creada.`);
        await invalidateAll();
    }
</script>

<div class="p-6">
    <div class="flex items-center justify-between mb-5">
        <h1 class="text-lg font-semibold text-[var(--text)]">Colecciones</h1>
        <button
            onclick={() => showCreate = true}
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm
                   bg-[var(--accent)] text-white rounded-lg hover:opacity-90 transition-opacity"
        >
            <Plus size={14} />
            Nueva colección
        </button>
    </div>

    {#if data.error}
        <p class="text-sm text-[var(--danger)]">{data.error}</p>
    {:else if data.collections.length === 0}
        <p class="text-sm text-[var(--text-muted)]">No hay colecciones. Creá una para empezar.</p>
    {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each data.collections as name (name)}
                <CollectionCard
                    {name}
                    stats={data.stats[name]}
                    href="/collections/{name}"
                />
            {/each}
        </div>
    {/if}
</div>

<CreateModal bind:open={showCreate} oncreate={handleCreate} />
```

- [ ] **Step 2: Verificar svelte-check**

```bash
cd ~/rag-saldivia/services/sda-frontend && npx svelte-check --tsconfig tsconfig.json 2>&1 | grep -E "Error|error" | grep -v "markdown" | head -20
```

Expected: 0 errores en collections/+page.svelte

- [ ] **Step 3: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/routes/\(app\)/collections/+page.svelte
git commit -m "feat(collections): upgrade index page with CollectionCard grid + create modal"
```

---

## Task 7: Expandir collection detail page + delete action

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/collections/[name]/+page.svelte`
- Modify: `services/sda-frontend/src/routes/(app)/collections/[name]/+page.server.ts`

- [ ] **Step 1: Agregar action `delete` al page.server.ts**

El archivo actual solo tiene `load`. Agregar `actions`:

```typescript
import type { PageServerLoad, Actions } from './$types';
import { error, redirect } from '@sveltejs/kit';
import { gatewayCollectionStats, gatewayDeleteCollection, GatewayError } from '$lib/server/gateway';

export const load: PageServerLoad = async ({ params }) => {
    try {
        const stats = await gatewayCollectionStats(params.name);
        return { name: params.name, stats };
    } catch (err) {
        if (err instanceof GatewayError && err.status === 404) {
            throw error(404, `Colección "${params.name}" no encontrada.`);
        }
        console.error('[collection detail loader]', err);
        return { name: params.name, stats: null, error: 'No se pudo cargar las estadísticas de la colección' };
    }
};

export const actions: Actions = {
    delete: async ({ params, locals }) => {
        if (!locals.user) throw error(401);
        try {
            await gatewayDeleteCollection(params.name);
        } catch (err) {
            const status = err instanceof GatewayError ? err.status : 503;
            throw error(status, `No se pudo eliminar la colección "${params.name}".`);
        }
        throw redirect(303, '/collections');
    },
};
```

- [ ] **Step 2: Reemplazar +page.svelte con versión mejorada**

```svelte
<script lang="ts">
    import DeleteModal from '../_components/DeleteModal.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';
    import { goto } from '$app/navigation';
    import { Trash2, FileText, Database } from 'lucide-svelte';

    let { data } = $props();

    let showDelete = $state(false);

    async function handleDelete() {
        const res = await fetch(`/api/collections/${data.name}`, { method: 'DELETE' });
        if (!res.ok) {
            const body = await res.json().catch(() => ({}));
            throw new Error(body.message ?? `Error ${res.status}`);
        }
        toastStore.success(`Colección "${data.name}" eliminada.`);
        goto('/collections');
    }
</script>

<div class="p-6 max-w-2xl">
    <div class="flex items-center gap-3 mb-6">
        <a href="/collections" class="text-[var(--text-faint)] hover:text-[var(--text-muted)] text-sm transition-colors">
            ← Colecciones
        </a>
        <h1 class="text-lg font-semibold text-[var(--text)]">{data.name}</h1>
    </div>

    {#if data.error}
        <p class="text-sm text-[var(--danger)] mb-4">{data.error}</p>
    {/if}

    <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-5 mb-4">
        <div class="grid grid-cols-2 gap-6 text-sm">
            <div>
                <div class="flex items-center gap-1.5 text-[var(--text-faint)] text-xs mb-1.5">
                    <Database size={12} />
                    Entidades
                </div>
                <div class="text-[var(--text)] font-semibold text-xl">
                    {data.stats?.entity_count?.toLocaleString() ?? '—'}
                </div>
            </div>
            <div>
                <div class="flex items-center gap-1.5 text-[var(--text-faint)] text-xs mb-1.5">
                    <FileText size={12} />
                    Documentos
                </div>
                <div class="text-[var(--text)] font-semibold text-xl">
                    {data.stats?.document_count?.toLocaleString() ?? '—'}
                </div>
            </div>
        </div>
    </div>

    <div class="flex items-center gap-3">
        <a
            href="/chat"
            class="inline-flex items-center px-4 py-2 text-sm bg-[var(--accent)] text-white
                   rounded-lg hover:opacity-90 transition-opacity"
        >
            Consultar esta colección
        </a>
        <button
            onclick={() => showDelete = true}
            class="inline-flex items-center gap-1.5 px-3 py-2 text-sm text-[var(--danger)]
                   border border-[var(--border)] rounded-lg hover:border-[var(--danger)]
                   hover:bg-[var(--danger)]/10 transition-colors"
        >
            <Trash2 size={14} />
            Eliminar
        </button>
    </div>
</div>

<DeleteModal bind:open={showDelete} name={data.name} onconfirm={handleDelete} />
```

- [ ] **Step 3: Verificar svelte-check**

```bash
cd ~/rag-saldivia/services/sda-frontend && npx svelte-check --tsconfig tsconfig.json 2>&1 | grep -E "Error|error" | grep -v "markdown" | head -20
```

Expected: 0 errores

- [ ] **Step 4: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/routes/\(app\)/collections/\[name\]/
git commit -m "feat(collections): add delete action + improve detail page UI"
```

---

## Task 8: Upload page

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/upload/+page.server.ts`
- Modify: `services/sda-frontend/src/routes/(app)/upload/+page.svelte`

- [ ] **Step 1: Crear upload/+page.server.ts**

Carga la lista de colecciones para el selector:

```typescript
import type { PageServerLoad } from './$types';
import { gatewayListCollections } from '$lib/server/gateway';

export const load: PageServerLoad = async () => {
    try {
        const { collections } = await gatewayListCollections();
        return { collections };
    } catch {
        return { collections: [] };
    }
};
```

- [ ] **Step 2: Reemplazar upload/+page.svelte con drag & drop funcional**

```svelte
<script lang="ts">
    import { toastStore } from '$lib/stores/toast.svelte';
    import { Upload, File as FileIcon, X } from 'lucide-svelte';

    let { data } = $props();

    const ACCEPTED_EXTENSIONS = ['.pdf', '.txt', '.md', '.docx'];
    const ACCEPTED_MIME = ['application/pdf', 'text/plain', 'text/markdown', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];

    let selectedFile = $state<File | null>(null);
    let selectedCollection = $state(data.collections[0] ?? '');
    let dragging = $state(false);
    let uploading = $state(false);
    let progress = $state(0);
    let fileError = $state('');

    function validateFile(file: File): string {
        const ext = '.' + file.name.split('.').pop()?.toLowerCase();
        if (!ACCEPTED_EXTENSIONS.includes(ext)) {
            return `Tipo no soportado. Acepta: ${ACCEPTED_EXTENSIONS.join(', ')}`;
        }
        if (file.size > 50 * 1024 * 1024) return 'El archivo no puede superar 50 MB.';
        return '';
    }

    function handleFileSelect(files: FileList | null) {
        if (!files || files.length === 0) return;
        const file = files[0];
        fileError = validateFile(file);
        selectedFile = fileError ? null : file;
    }

    function onDragEnter(e: DragEvent) { e.preventDefault(); dragging = true; }
    function onDragLeave(e: DragEvent) { e.preventDefault(); dragging = false; }
    function onDrop(e: DragEvent) {
        e.preventDefault();
        dragging = false;
        handleFileSelect(e.dataTransfer?.files ?? null);
    }

    async function handleUpload() {
        if (!selectedFile || !selectedCollection) return;
        uploading = true;
        progress = 0;

        // Simular progress bar (la ingesta es async)
        const interval = setInterval(() => {
            progress = Math.min(progress + 10, 85);
        }, 200);

        try {
            const formData = new FormData();
            formData.append('file', selectedFile);
            formData.append('collection', selectedCollection);

            const res = await fetch('/api/upload', { method: 'POST', body: formData });
            clearInterval(interval);
            progress = 100;

            if (!res.ok) {
                const body = await res.json().catch(() => ({}));
                throw new Error(body.message ?? `Error ${res.status}`);
            }

            toastStore.success(`"${selectedFile.name}" enviado a ingesta en "${selectedCollection}".`);
            selectedFile = null;
            progress = 0;
        } catch (e: any) {
            clearInterval(interval);
            progress = 0;
            toastStore.error(e.message ?? 'Error al subir el archivo.');
        } finally {
            uploading = false;
        }
    }
</script>

<div class="p-6 max-w-xl">
    <h1 class="text-lg font-semibold text-[var(--text)] mb-6">Subir documentos</h1>

    <!-- Drop zone -->
    <div
        role="button"
        tabindex="0"
        ondragenter={onDragEnter}
        ondragleave={onDragLeave}
        ondragover={(e) => e.preventDefault()}
        ondrop={onDrop}
        onclick={() => document.getElementById('file-input')?.click()}
        onkeydown={(e) => e.key === 'Enter' && document.getElementById('file-input')?.click()}
        class="border-2 border-dashed rounded-[var(--radius-lg)] p-8 text-center cursor-pointer
               transition-colors {dragging
                   ? 'border-[var(--accent)] bg-[var(--accent)]/5'
                   : 'border-[var(--border)] hover:border-[var(--text-faint)]'}"
    >
        <input
            id="file-input"
            type="file"
            class="hidden"
            accept={ACCEPTED_MIME.join(',')}
            onchange={(e) => handleFileSelect((e.target as HTMLInputElement).files)}
        />

        {#if selectedFile}
            <div class="flex items-center justify-center gap-3">
                <FileIcon size={20} class="text-[var(--accent)]" />
                <span class="text-sm text-[var(--text)] font-medium">{selectedFile.name}</span>
                <button
                    onclick={(e) => { e.stopPropagation(); selectedFile = null; fileError = ''; }}
                    class="text-[var(--text-faint)] hover:text-[var(--text)]"
                    aria-label="Quitar archivo"
                >
                    <X size={16} />
                </button>
            </div>
            <p class="text-xs text-[var(--text-faint)] mt-1.5">
                {(selectedFile.size / 1024).toFixed(1)} KB
            </p>
        {:else}
            <Upload size={24} class="text-[var(--text-faint)] mx-auto mb-3" />
            <p class="text-sm text-[var(--text-muted)]">
                Arrastrá un archivo o <span class="text-[var(--accent)]">hacé click para elegir</span>
            </p>
            <p class="text-xs text-[var(--text-faint)] mt-1">
                {ACCEPTED_EXTENSIONS.join(', ')} · máx 50 MB
            </p>
        {/if}
    </div>

    {#if fileError}
        <p class="text-xs text-[var(--danger)] mt-2">{fileError}</p>
    {/if}

    <!-- Collection selector -->
    <div class="mt-5">
        <label for="collection-select" class="block text-xs font-medium text-[var(--text-muted)] mb-1.5">
            Colección destino
        </label>
        {#if data.collections.length === 0}
            <p class="text-sm text-[var(--text-faint)]">
                No hay colecciones disponibles.
                <a href="/collections" class="text-[var(--accent)] hover:underline">Creá una primero.</a>
            </p>
        {:else}
            <select
                id="collection-select"
                bind:value={selectedCollection}
                class="w-full bg-[var(--bg-base)] border border-[var(--border)] rounded-lg
                       px-3 py-2 text-sm text-[var(--text)] focus:outline-none
                       focus:border-[var(--accent)] transition-colors"
            >
                {#each data.collections as col}
                    <option value={col}>{col}</option>
                {/each}
            </select>
        {/if}
    </div>

    <!-- Progress bar -->
    {#if uploading && progress > 0}
        <div class="mt-4">
            <div class="h-1.5 bg-[var(--bg-hover)] rounded-full overflow-hidden">
                <div
                    class="h-full bg-[var(--accent)] transition-all duration-200 ease-out rounded-full"
                    style="width: {progress}%"
                ></div>
            </div>
            <p class="text-xs text-[var(--text-faint)] mt-1">{progress < 100 ? 'Subiendo...' : 'Procesando...'}</p>
        </div>
    {/if}

    <!-- Upload button -->
    <button
        onclick={handleUpload}
        disabled={!selectedFile || !selectedCollection || uploading}
        class="mt-5 w-full py-2.5 px-4 text-sm font-medium text-white bg-[var(--accent)]
               rounded-lg hover:opacity-90 transition-opacity
               disabled:opacity-40 disabled:cursor-not-allowed"
    >
        {uploading ? 'Subiendo...' : 'Subir documento'}
    </button>
</div>
```

- [ ] **Step 3: Verificar svelte-check**

```bash
cd ~/rag-saldivia/services/sda-frontend && npx svelte-check --tsconfig tsconfig.json 2>&1 | grep -E "Error|error" | grep -v "markdown" | head -20
```

Expected: 0 errores

- [ ] **Step 4: Commit**

```bash
cd ~/rag-saldivia/services/sda-frontend
git add src/routes/\(app\)/upload/
git commit -m "feat(upload): replace stub with functional drag-and-drop upload page"
```

---

## Task 9: Verificación final

**Files:** ninguno nuevo

- [ ] **Step 1: Correr todos los tests**

```bash
cd ~/rag-saldivia/services/sda-frontend && npm test 2>&1 | tail -20
```

Expected:
- Tests Files: 1 failed (markdown.test.ts, pre-existente) | ≥7 passed
- Tests: ≥28 passed (20 pre-existentes + 5 store + 5 BFF collections + 3 BFF upload)

- [ ] **Step 2: Verificar svelte-check completo**

```bash
cd ~/rag-saldivia/services/sda-frontend && npx svelte-check --tsconfig tsconfig.json 2>&1 | grep -c "Error" || echo "0 errors"
```

Expected: 0 errors (el warning de markdown es un error pre-existente, verificar que no agregamos nuevos)

- [ ] **Step 3: Commit final si hubiera cambios pendientes**

```bash
cd ~/rag-saldivia/services/sda-frontend
git status
# Solo commitear si hay algo sin commitear
```

---

## Troubleshooting

### `vi.mock('$lib/server/gateway', ...)` no funciona en tests BFF
Vitest resuelve `$lib` como alias. Si falla, usar la ruta relativa real en el mock:
```typescript
vi.mock('../../../lib/server/gateway.js', () => mockGateway);
```

### `Button` no tiene variante `danger`
Ver el componente en `src/lib/components/ui/Button.svelte`. Si no existe `danger`, en DeleteModal usar `<button class="...danger styles...">` directamente (está documentado en el plan).

### `svelte-check` reporta errores en `_components`
El path `(app)` necesita estar escapado en la terminal. Usar siempre:
```bash
git add src/routes/\(app\)/collections/
```

### Los imports de `$app/navigation` no resuelven en tests
`goto` y `invalidateAll` son runtime de SvelteKit. Si un test necesita importar un componente que los usa, mockear `$app/navigation`:
```typescript
vi.mock('$app/navigation', () => ({ goto: vi.fn(), invalidateAll: vi.fn() }));
```
