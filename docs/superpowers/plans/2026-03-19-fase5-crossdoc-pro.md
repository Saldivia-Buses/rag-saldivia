# Fase 5 — Crossdoc Pro Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Portar el pipeline crossdoc de React a SvelteKit 5, permitiendo al usuario activar un modo "Crossdoc" que descompone su pregunta en N sub-queries paralelas, las ejecuta contra la KB, y sintetiza una respuesta final.

**Architecture:** `CrossdocStore` (Svelte 5 singleton) orquesta 4 fases llamando 3 BFF endpoints thin que internamente usan 2 nuevas funciones en `gateway.ts`. El progreso se muestra inline en el assistant bubble. El toggle on/off vive en el ChatInput como chip `⚡ Crossdoc`.

**Tech Stack:** SvelteKit 5, Svelte 5 runes, TypeScript, Vitest, `$lib/server/gateway.ts` pattern.

**Spec:** `docs/superpowers/specs/2026-03-19-fase5-crossdoc-pro.md`
**Reference:** `patches/frontend/new/useCrossdocDecompose.ts`, `patches/frontend/new/useCrossdocStream.ts`

---

## File Map

| Archivo | Acción | Responsabilidad |
|---------|--------|-----------------|
| `src/lib/crossdoc/types.ts` | CREAR | Interfaces: CrossdocOptions, CrossdocProgress, SubResult |
| `src/lib/crossdoc/pipeline.ts` | CREAR | Funciones puras: jaccard, dedup, parseSubQueries, hasUsefulData |
| `src/lib/crossdoc/pipeline.test.ts` | CREAR | Tests de funciones puras |
| `src/lib/server/gateway.ts` | MODIFICAR | +gatewayGenerateText, +gatewayGenerateStream |
| `src/routes/api/crossdoc/decompose/+server.ts` | CREAR | BFF decompose → string[] |
| `src/routes/api/crossdoc/decompose/decompose.test.ts` | CREAR | Tests del endpoint |
| `src/routes/api/crossdoc/subquery/+server.ts` | CREAR | BFF sub-query RAG → { content, success } |
| `src/routes/api/crossdoc/subquery/subquery.test.ts` | CREAR | Tests del endpoint |
| `src/routes/api/crossdoc/synthesize/+server.ts` | CREAR | BFF synthesis → SSE proxy |
| `src/routes/api/crossdoc/synthesize/synthesize.test.ts` | CREAR | Tests del endpoint |
| `src/lib/stores/chat.svelte.ts` | MODIFICAR | Message interface +crossdocResults; finalizeStream acepta opts? |
| `src/lib/stores/chat.svelte.test.ts` | MODIFICAR | +tests para crossdocResults |
| `src/lib/stores/crossdoc.svelte.ts` | CREAR | CrossdocStore singleton: run(), stop(), reset(), progress, options |
| `src/lib/stores/crossdoc.svelte.test.ts` | CREAR | Tests del store |
| `src/lib/components/chat/CrossdocProgress.svelte` | CREAR | Pills 4 fases + barra progreso numérica |
| `src/lib/components/chat/DecompositionView.svelte` | CREAR | Accordion sub-queries debajo de la respuesta |
| `src/lib/components/chat/CrossdocSettingsPopover.svelte` | CREAR | Popover compacto con settings de crossdoc |
| `src/lib/actions/clickOutside.ts` | CREAR | Svelte action para cerrar popover con click fuera |
| `src/lib/components/chat/ChatInput.svelte` | MODIFICAR | +prop crossdoc, +oncrossdoctoggle, +CrossdocSettingsPopover |
| `src/routes/(app)/chat/[id]/+page.svelte` | MODIFICAR | +crossdoc branch en sendMessage, +CrossdocProgress en bubble |
| `src/lib/components/chat/MessageList.svelte` | MODIFICAR | +DecompositionView por mensaje assistant |

---

## Task 1: Tipos + funciones puras del pipeline

**Files:**
- Create: `services/sda-frontend/src/lib/crossdoc/types.ts`
- Create: `services/sda-frontend/src/lib/crossdoc/pipeline.ts`
- Create: `services/sda-frontend/src/lib/crossdoc/pipeline.test.ts`

- [ ] **Step 1: Crear types.ts**

```typescript
// services/sda-frontend/src/lib/crossdoc/types.ts

export interface CrossdocOptions {
    maxSubQueries: number;       // 0 = ilimitado
    synthesisModel: string;      // '' = usar LLM por defecto
    followUpRetries: boolean;
    showDecomposition: boolean;
    vdbTopK: number;
    rerankerTopK: number;
}

export interface SubResult {
    query: string;
    content: string;
    success: boolean;
}

export interface CrossdocProgress {
    phase: 'decomposing' | 'querying' | 'retrying' | 'synthesizing' | 'done' | 'error';
    subQueries: string[];
    completed: number;
    total: number;
    results: SubResult[];
    error?: string;
}

export const DEFAULT_CROSSDOC_OPTIONS: CrossdocOptions = {
    maxSubQueries: 4,
    synthesisModel: '',
    followUpRetries: true,
    showDecomposition: false,
    vdbTopK: 10,
    rerankerTopK: 5,
};
```

- [ ] **Step 2: Escribir los tests ANTES de implementar pipeline.ts**

```typescript
// services/sda-frontend/src/lib/crossdoc/pipeline.test.ts
import { describe, it, expect } from 'vitest';
import { jaccard, dedup, parseSubQueries, hasUsefulData } from './pipeline.js';

describe('jaccard', () => {
    it('identical strings → 1.0', () => {
        expect(jaccard('presión bomba', 'presión bomba')).toBe(1);
    });
    it('disjoint strings → 0', () => {
        expect(jaccard('presión bomba', 'temperatura motor')).toBe(0);
    });
    it('partial overlap → between 0 and 1', () => {
        const score = jaccard('presión máxima bomba', 'presión mínima bomba');
        expect(score).toBeGreaterThan(0);
        expect(score).toBeLessThan(1);
    });
});

describe('dedup', () => {
    it('elimina queries con jaccard >= 0.65', () => {
        const queries = [
            'presión máxima bomba centrífuga',
            'presión máxima bomba centrifuga',  // casi idéntica
            'temperatura motor eléctrico',
        ];
        const result = dedup(queries);
        expect(result).toHaveLength(2);
        expect(result[0]).toBe('presión máxima bomba centrífuga');
        expect(result[1]).toBe('temperatura motor eléctrico');
    });

    it('no elimina queries distintas', () => {
        const queries = ['presión bomba', 'temperatura motor', 'voltaje inversor'];
        expect(dedup(queries)).toHaveLength(3);
    });
});

describe('parseSubQueries', () => {
    it('parsea líneas simples', () => {
        const text = 'presión máxima\ntemperatura motor\nvoltaje nominal';
        expect(parseSubQueries(text)).toEqual([
            'presión máxima',
            'temperatura motor',
            'voltaje nominal',
        ]);
    });
    it('elimina líneas con numeración', () => {
        const text = '1. presión máxima\n2) temperatura motor';
        const result = parseSubQueries(text);
        expect(result[0]).toBe('presión máxima');
        expect(result[1]).toBe('temperatura motor');
    });
    it('filtra líneas vacías o muy cortas', () => {
        const text = '\npresión máxima\n   \nok\n';
        const result = parseSubQueries(text);
        expect(result).toHaveLength(1);
        expect(result[0]).toBe('presión máxima');
    });
    it('aplica cap cuando se pasa maxSubQueries', () => {
        const text = 'a uno\nb dos\nc tres\nd cuatro';
        expect(parseSubQueries(text, 2)).toHaveLength(2);
    });
});

describe('hasUsefulData', () => {
    it('texto con contenido → true', () => {
        expect(hasUsefulData('La presión máxima es 12 bar.')).toBe(true);
    });
    it('texto vacío → false', () => {
        expect(hasUsefulData('')).toBe(false);
        expect(hasUsefulData('   ')).toBe(false);
    });
    it('patrones de "sin datos" → false', () => {
        expect(hasUsefulData('No information found')).toBe(false);
        expect(hasUsefulData("I cannot answer this question")).toBe(false);
        expect(hasUsefulData('out of context')).toBe(false);
    });
});
```

- [ ] **Step 3: Correr tests para verificar que FALLAN**

```bash
cd services/sda-frontend && npx vitest run src/lib/crossdoc/pipeline.test.ts
```
Esperado: FAIL — `Cannot find module './pipeline.js'`

- [ ] **Step 4: Implementar pipeline.ts**

```typescript
// services/sda-frontend/src/lib/crossdoc/pipeline.ts

const JACCARD_THRESHOLD = 0.65;
const MAX_RESPONSE_CHARS = 15_000;

export function jaccard(a: string, b: string): number {
    const setA = new Set(a.toLowerCase().split(/\s+/).filter(Boolean));
    const setB = new Set(b.toLowerCase().split(/\s+/).filter(Boolean));
    const intersection = [...setA].filter(x => setB.has(x)).length;
    const union = new Set([...setA, ...setB]).size;
    return union === 0 ? 0 : intersection / union;
}

export function dedup(queries: string[]): string[] {
    const result: string[] = [];
    for (const q of queries) {
        if (!result.some(existing => jaccard(existing, q) >= JACCARD_THRESHOLD)) {
            result.push(q);
        }
    }
    return result;
}

export function parseSubQueries(text: string, maxSubQueries = 0): string[] {
    let queries = text
        .split('\n')
        .map(line => line.replace(/^\d+[\.\)]\s*/, '').trim())
        .filter(line => line.length > 5 && line.length < 200);
    queries = dedup(queries);
    if (maxSubQueries > 0) queries = queries.slice(0, maxSubQueries);
    return queries;
}

export function hasUsefulData(text: string): boolean {
    const trimmed = text.trim();
    if (trimmed.length < 3) return false;
    const emptyPatterns = [
        /^(no|sin)\s+(information|data|results|context)/i,
        /^out of context$/i,
        /^i (cannot|can't|don't)/i,
    ];
    return !emptyPatterns.some(p => p.test(trimmed));
}

/** Trunca texto si hay repetición detectada (portado de useCrossdocStream). */
export function truncateIfRepetitive(text: string): string {
    const WINDOW = 60;
    const THRESHOLD = 3;
    if (text.length <= WINDOW * THRESHOLD) return text;
    const tail = text.slice(-WINDOW);
    const preceding = text.slice(-(WINDOW * (THRESHOLD + 1)), -WINDOW);
    if (preceding.split(tail).length - 1 >= THRESHOLD - 1) {
        const firstIdx = text.indexOf(tail);
        if (firstIdx > 0 && firstIdx < text.length - WINDOW) {
            return text.slice(0, firstIdx + WINDOW);
        }
    }
    if (text.length > MAX_RESPONSE_CHARS) return text.slice(0, MAX_RESPONSE_CHARS);
    return text;
}

export const DECOMPOSE_PROMPT = (question: string) =>
`You are a search query decomposer for a technical document retrieval system.

Given the user's question, generate multiple retrieval-focused sub-queries. Each sub-query should:
- Target a SPECIFIC product, component, or technical specification
- Use generic catalog/manual terminology (not user-specific context)
- Be at most 15 words
- Be independent — each should retrieve different documents

Return ONLY the sub-queries, one per line. No numbering, no explanations.

User question: ${question}`;

export const FOLLOWUP_PROMPT = (failedQueries: string[]) =>
`These search queries returned no useful results:
${failedQueries.map(q => `- ${q}`).join('\n')}

Generate alternative queries using synonyms, broader terms, or different technical vocabulary.
One query per line, no numbering.`;

export const SYNTHESIS_PROMPT = (question: string, results: { query: string; content: string }[]) => {
    const context = results
        .map((r, i) => `[Sub-query ${i + 1}: "${r.query}"]\n${r.content}`)
        .join('\n\n---\n\n');
    return `You are a senior engineer writing a comprehensive technical answer.

Based on the following retrieval results from multiple sub-queries, write a single unified answer to the user's original question.

Rules:
- Cite sources when possible (mention which sub-query or document the info came from)
- Include specific numbers, measurements, and technical specifications
- Be thorough but concise — cover all relevant information
- Use professional technical language

Original question: ${question}

Retrieval results:
${context}`;
};
```

- [ ] **Step 5: Correr tests y verificar que PASAN**

```bash
cd services/sda-frontend && npx vitest run src/lib/crossdoc/pipeline.test.ts
```
Esperado: PASS (8 tests)

- [ ] **Step 6: Commit**

```bash
git add services/sda-frontend/src/lib/crossdoc/
git commit -m "feat(crossdoc): add types and pure pipeline utilities"
```

---

## Task 2: Extender gateway.ts con gatewayGenerateText y gatewayGenerateStream

**Files:**
- Modify: `services/sda-frontend/src/lib/server/gateway.ts`

> Nota: estas funciones se testean indirectamente via los tests de endpoints BFF. No se agregan tests unitarios separados (la función `gw()` interna tampoco los tiene).

- [ ] **Step 1: Agregar las dos funciones al final de gateway.ts, antes de los tipos**

Agregar después de la línea `export async function gatewayGetAudit(...)` (línea ~157) y antes de `// Types`:

```typescript
// Generate — text (para decompose y sub-queries, sin SSE al browser)
export async function gatewayGenerateText(
    opts: {
        messages: { role: string; content: string }[];
        use_knowledge_base?: boolean;
        collection_names?: string[];
        vdb_top_k?: number;
        reranker_top_k?: number;
        enable_reranker?: boolean;
        max_tokens?: number;
    },
    signal?: AbortSignal
): Promise<string> {
    const resp = await fetch(`${GATEWAY_URL}/v1/generate`, {
        method: 'POST',
        headers: headers(),
        body: JSON.stringify(opts),
        signal,
    });
    if (!resp.ok) throw new GatewayError(resp.status, await resp.text());
    if (!resp.body) return '';

    let text = '';
    const reader = resp.body.getReader();
    const decoder = new TextDecoder();
    try {
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            const chunk = decoder.decode(value, { stream: true });
            for (const line of chunk.split('\n')) {
                if (!line.startsWith('data: ')) continue;
                const data = line.slice(6).trim();
                if (data === '[DONE]') continue;
                try {
                    const obj = JSON.parse(data);
                    text += obj?.choices?.[0]?.delta?.content ?? '';
                } catch { /* skip malformed */ }
            }
        }
    } finally {
        reader.releaseLock();
    }
    return text;
}

// Generate — stream (para synthesis, proxy SSE directo al browser)
export async function gatewayGenerateStream(
    opts: {
        messages: { role: string; content: string }[];
        use_knowledge_base?: boolean;
        max_tokens?: number;
    },
    signal?: AbortSignal
): Promise<Response> {
    const resp = await fetch(`${GATEWAY_URL}/v1/generate`, {
        method: 'POST',
        headers: headers(),
        body: JSON.stringify(opts),
        signal,
    });
    if (!resp.ok) throw new GatewayError(resp.status, await resp.text());
    return resp;
}
```

- [ ] **Step 2: Verificar que TypeScript compila sin errores**

```bash
cd services/sda-frontend && npx tsc --noEmit
```
Esperado: sin errores

- [ ] **Step 3: Commit**

```bash
git add services/sda-frontend/src/lib/server/gateway.ts
git commit -m "feat(crossdoc): add gatewayGenerateText and gatewayGenerateStream"
```

---

## Task 3: BFF endpoint /api/crossdoc/decompose

**Files:**
- Create: `services/sda-frontend/src/routes/api/crossdoc/decompose/+server.ts`
- Create: `services/sda-frontend/src/routes/api/crossdoc/decompose/decompose.test.ts`

- [ ] **Step 1: Escribir el test ANTES del endpoint**

```typescript
// services/sda-frontend/src/routes/api/crossdoc/decompose/decompose.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGateway = {
    gatewayGenerateText: vi.fn(),
    GatewayError: class GatewayError extends Error {
        status: number; detail: string;
        constructor(status: number, detail: string) { super(detail); this.status = status; this.detail = detail; }
    },
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('POST /api/crossdoc/decompose', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 401 without auth', async () => {
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question: 'test?' }),
            }),
            locals: { user: null },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 401 });
    });

    it('parsea sub-queries del LLM y aplica dedup', async () => {
        mockGateway.gatewayGenerateText.mockResolvedValue(
            'presión máxima bomba centrífuga\ntemperatura motor eléctrico\nvoltaje nominal inversor'
        );
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question: '¿Cuál es la presión máxima?' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        expect(res.status).toBe(200);
        const body = await res.json();
        expect(Array.isArray(body.subQueries)).toBe(true);
        expect(body.subQueries.length).toBeGreaterThan(0);
    });

    it('respeta el cap de maxSubQueries', async () => {
        mockGateway.gatewayGenerateText.mockResolvedValue(
            'query uno aqui\nquery dos aqui\nquery tres aqui\nquery cuatro aqui\nquery cinco aqui'
        );
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question: 'test', maxSubQueries: 2 }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        const body = await res.json();
        expect(body.subQueries.length).toBeLessThanOrEqual(2);
    });
});
```

- [ ] **Step 2: Correr tests para verificar que FALLAN**

```bash
cd services/sda-frontend && npx vitest run src/routes/api/crossdoc/decompose/decompose.test.ts
```
Esperado: FAIL — module not found

- [ ] **Step 3: Implementar el endpoint**

```typescript
// services/sda-frontend/src/routes/api/crossdoc/decompose/+server.ts
import type { RequestHandler } from './$types';
import { error, json } from '@sveltejs/kit';
import { gatewayGenerateText, GatewayError } from '$lib/server/gateway';
import { parseSubQueries, DECOMPOSE_PROMPT } from '$lib/crossdoc/pipeline';

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401, 'Unauthorized');

    const { question, maxSubQueries = 0 } = await request.json();
    if (!question?.trim()) throw error(400, 'question is required');

    try {
        const text = await gatewayGenerateText({
            messages: [{ role: 'user', content: DECOMPOSE_PROMPT(question) }],
            use_knowledge_base: false,
            max_tokens: 2048,
        });
        const subQueries = parseSubQueries(text, maxSubQueries);
        return json({ subQueries });
    } catch (err) {
        if (err instanceof GatewayError) throw error(err.status, err.detail);
        throw error(502, 'Decompose failed');
    }
};
```

- [ ] **Step 4: Correr tests y verificar que PASAN**

```bash
cd services/sda-frontend && npx vitest run src/routes/api/crossdoc/decompose/decompose.test.ts
```
Esperado: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add services/sda-frontend/src/routes/api/crossdoc/decompose/
git commit -m "feat(crossdoc): add /api/crossdoc/decompose BFF endpoint"
```

---

## Task 4: BFF endpoint /api/crossdoc/subquery

**Files:**
- Create: `services/sda-frontend/src/routes/api/crossdoc/subquery/+server.ts`
- Create: `services/sda-frontend/src/routes/api/crossdoc/subquery/subquery.test.ts`

- [ ] **Step 1: Escribir el test ANTES del endpoint**

```typescript
// services/sda-frontend/src/routes/api/crossdoc/subquery/subquery.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGateway = {
    gatewayGenerateText: vi.fn(),
    GatewayError: class GatewayError extends Error {
        status: number; detail: string;
        constructor(status: number, detail: string) { super(detail); this.status = status; this.detail = detail; }
    },
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('POST /api/crossdoc/subquery', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 401 without auth', async () => {
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: 'presión bomba' }),
            }),
            locals: { user: null },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 401 });
    });

    it('success: true cuando gateway devuelve contenido útil', async () => {
        mockGateway.gatewayGenerateText.mockResolvedValue('La presión máxima es 12 bar según la ficha técnica.');
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: 'presión máxima bomba' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        const body = await res.json();
        expect(body.success).toBe(true);
        expect(body.content).toBe('La presión máxima es 12 bar según la ficha técnica.');
    });

    it('success: false cuando gateway devuelve respuesta vacía', async () => {
        mockGateway.gatewayGenerateText.mockResolvedValue('No information found');
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: 'query sin resultados' }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        const body = await res.json();
        expect(body.success).toBe(false);
    });
});
```

- [ ] **Step 2: Correr tests para verificar que FALLAN**

```bash
cd services/sda-frontend && npx vitest run src/routes/api/crossdoc/subquery/subquery.test.ts
```

- [ ] **Step 3: Implementar el endpoint**

```typescript
// services/sda-frontend/src/routes/api/crossdoc/subquery/+server.ts
import type { RequestHandler } from './$types';
import { error, json } from '@sveltejs/kit';
import { gatewayGenerateText, GatewayError } from '$lib/server/gateway';
import { hasUsefulData, truncateIfRepetitive } from '$lib/crossdoc/pipeline';

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401, 'Unauthorized');

    const { query, collection_names, vdbTopK = 10, rerankerTopK = 5 } = await request.json();
    if (!query?.trim()) throw error(400, 'query is required');

    try {
        const raw = await gatewayGenerateText({
            messages: [{ role: 'user', content: query }],
            use_knowledge_base: true,
            collection_names,
            vdb_top_k: vdbTopK,
            reranker_top_k: rerankerTopK,
            enable_reranker: true,
            max_tokens: 2048,
        });
        const content = truncateIfRepetitive(raw);
        return json({ content, success: hasUsefulData(content) });
    } catch (err) {
        if (err instanceof GatewayError) throw error(err.status, err.detail);
        throw error(502, 'Subquery failed');
    }
};
```

- [ ] **Step 4: Correr tests y verificar que PASAN**

```bash
cd services/sda-frontend && npx vitest run src/routes/api/crossdoc/subquery/subquery.test.ts
```
Esperado: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add services/sda-frontend/src/routes/api/crossdoc/subquery/
git commit -m "feat(crossdoc): add /api/crossdoc/subquery BFF endpoint"
```

---

## Task 5: BFF endpoint /api/crossdoc/synthesize

**Files:**
- Create: `services/sda-frontend/src/routes/api/crossdoc/synthesize/+server.ts`
- Create: `services/sda-frontend/src/routes/api/crossdoc/synthesize/synthesize.test.ts`

- [ ] **Step 1: Escribir el test ANTES del endpoint**

```typescript
// services/sda-frontend/src/routes/api/crossdoc/synthesize/synthesize.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockStream = new ReadableStream({
    start(c) { c.enqueue(new TextEncoder().encode('data: {"choices":[{"delta":{"content":"respuesta"}}]}\n')); c.close(); }
});

const mockGateway = {
    gatewayGenerateStream: vi.fn(),
    GatewayError: class GatewayError extends Error {
        status: number; detail: string;
        constructor(status: number, detail: string) { super(detail); this.status = status; this.detail = detail; }
    },
};
vi.mock('$lib/server/gateway', () => mockGateway);

describe('POST /api/crossdoc/synthesize', () => {
    beforeEach(() => vi.resetAllMocks());

    it('returns 401 without auth', async () => {
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question: 'test', results: [] }),
            }),
            locals: { user: null },
        } as any;
        await expect(POST(event)).rejects.toMatchObject({ status: 401 });
    });

    it('retorna SSE stream con Content-Type text/event-stream', async () => {
        mockGateway.gatewayGenerateStream.mockResolvedValue(
            new Response(mockStream, { headers: { 'Content-Type': 'text/event-stream' } })
        );
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    question: '¿Cuál es la presión máxima?',
                    results: [{ query: 'presión bomba', content: 'La presión es 12 bar', success: true }],
                }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        const res = await POST(event);
        expect(res.status).toBe(200);
        expect(res.headers.get('Content-Type')).toContain('text/event-stream');
    });

    it('llama gatewayGenerateStream con el synthesis prompt correcto', async () => {
        mockGateway.gatewayGenerateStream.mockResolvedValue(
            new Response(mockStream, { headers: { 'Content-Type': 'text/event-stream' } })
        );
        const { POST } = await import('./+server.js');
        const event = {
            request: new Request('http://localhost', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    question: '¿Cuál es la presión?',
                    results: [{ query: 'presión bomba', content: 'La presión es 12 bar', success: true }],
                }),
            }),
            locals: { user: { id: 1 } },
        } as any;
        await POST(event);
        expect(mockGateway.gatewayGenerateStream).toHaveBeenCalledOnce();
        const callOpts = mockGateway.gatewayGenerateStream.mock.calls[0][0];
        expect(callOpts.messages[0].content).toContain('¿Cuál es la presión?');
        expect(callOpts.messages[0].content).toContain('La presión es 12 bar');
    });
});
```

- [ ] **Step 2: Correr tests para verificar que FALLAN**

```bash
cd services/sda-frontend && npx vitest run src/routes/api/crossdoc/synthesize/synthesize.test.ts
```

- [ ] **Step 3: Implementar el endpoint**

```typescript
// services/sda-frontend/src/routes/api/crossdoc/synthesize/+server.ts
import type { RequestHandler } from './$types';
import { error } from '@sveltejs/kit';
import { gatewayGenerateStream, GatewayError } from '$lib/server/gateway';
import { SYNTHESIS_PROMPT } from '$lib/crossdoc/pipeline';
import type { SubResult } from '$lib/crossdoc/types';

export const POST: RequestHandler = async ({ request, locals }) => {
    if (!locals.user) throw error(401, 'Unauthorized');

    const { question, results }: { question: string; results: SubResult[] } = await request.json();
    if (!question?.trim()) throw error(400, 'question is required');

    const successResults = (results ?? []).filter(r => r.success);

    try {
        const resp = await gatewayGenerateStream({
            messages: [{ role: 'user', content: SYNTHESIS_PROMPT(question, successResults) }],
            use_knowledge_base: false,
            max_tokens: 4096,
        });

        return new Response(resp.body, {
            headers: {
                'Content-Type': 'text/event-stream',
                'Cache-Control': 'no-cache',
                'Connection': 'keep-alive',
            },
        });
    } catch (err) {
        if (err instanceof GatewayError) throw error(err.status, err.detail);
        throw error(502, 'Synthesis failed');
    }
};
```

- [ ] **Step 4: Correr tests y verificar que PASAN**

```bash
cd services/sda-frontend && npx vitest run src/routes/api/crossdoc/synthesize/synthesize.test.ts
```
Esperado: PASS (3 tests)

- [ ] **Step 5: Correr todos los tests del proyecto para verificar nada se rompió**

```bash
cd services/sda-frontend && npx vitest run
```
Esperado: todos PASS

- [ ] **Step 6: Commit**

```bash
git add services/sda-frontend/src/routes/api/crossdoc/
git commit -m "feat(crossdoc): add /api/crossdoc/synthesize BFF endpoint"
```

---

## Task 6: Extender ChatStore con crossdocResults

**Files:**
- Modify: `services/sda-frontend/src/lib/stores/chat.svelte.ts`
- Modify: `services/sda-frontend/src/lib/stores/chat.svelte.test.ts`

- [ ] **Step 1: Agregar tests nuevos a chat.svelte.test.ts**

Agregar al final del archivo existente:

```typescript
// Agregar al final de chat.svelte.test.ts
import type { SubResult } from '$lib/crossdoc/types';

describe('ChatStore — crossdocResults', () => {
    it('finalizeStream guarda crossdocResults en el mensaje', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('respuesta de síntesis');
        const results: SubResult[] = [
            { query: 'presión bomba', content: 'La presión es 12 bar', success: true },
        ];
        chat.finalizeStream({ crossdocResults: results });
        expect(chat.messages[0].crossdocResults).toEqual(results);
    });

    it('finalizeStream sin opts funciona igual que antes (backwards compat)', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('respuesta normal');
        chat.finalizeStream();
        expect(chat.messages[0].crossdocResults).toBeUndefined();
        expect(chat.messages[0].content).toBe('respuesta normal');
    });
});
```

- [ ] **Step 2: Correr los tests nuevos para verificar que FALLAN**

```bash
cd services/sda-frontend && npx vitest run src/lib/stores/chat.svelte.test.ts
```
Esperado: los 2 tests nuevos FAIL

- [ ] **Step 3: Modificar chat.svelte.ts**

Cambiar la interfaz `Message` para agregar `crossdocResults?`:

```typescript
// Cambiar la interfaz Message (línea ~3):
export interface Message {
    role: 'user' | 'assistant';
    content: string;
    sources?: Source[];
    timestamp: string;
    crossdocResults?: import('$lib/crossdoc/types').SubResult[];
}
```

Cambiar la firma de `finalizeStream`:

```typescript
// Reemplazar el método finalizeStream:
finalizeStream(opts?: { crossdocResults?: import('$lib/crossdoc/types').SubResult[] }) {
    if (this.streamingContent) {
        this.messages.push({
            role: 'assistant',
            content: this.streamingContent,
            sources: [...this.sources],
            timestamp: new Date().toISOString(),
            crossdocResults: opts?.crossdocResults,
        });
    }
    this.streaming = false;
    this.streamingContent = '';
}
```

- [ ] **Step 4: Correr todos los tests del ChatStore y verificar que PASAN**

```bash
cd services/sda-frontend && npx vitest run src/lib/stores/chat.svelte.test.ts
```
Esperado: PASS (6 tests — 4 originales + 2 nuevos)

- [ ] **Step 5: Commit**

```bash
git add services/sda-frontend/src/lib/stores/chat.svelte.ts services/sda-frontend/src/lib/stores/chat.svelte.test.ts
git commit -m "feat(crossdoc): extend ChatStore.finalizeStream with crossdocResults"
```

---

## Task 7: CrossdocStore

**Files:**
- Create: `services/sda-frontend/src/lib/stores/crossdoc.svelte.ts`
- Create: `services/sda-frontend/src/lib/stores/crossdoc.svelte.test.ts`

- [ ] **Step 1: Escribir los tests ANTES del store**

```typescript
// services/sda-frontend/src/lib/stores/crossdoc.svelte.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CrossdocStore } from './crossdoc.svelte.js';
import { ChatStore } from './chat.svelte.js';
import { DEFAULT_CROSSDOC_OPTIONS } from '$lib/crossdoc/types';

// Mock fetch global
const mockFetch = vi.fn();
global.fetch = mockFetch;

function makeSSEResponse(text: string) {
    const stream = new ReadableStream({
        start(c) {
            c.enqueue(new TextEncoder().encode(`data: {"choices":[{"delta":{"content":"${text}"}}]}\n`));
            c.enqueue(new TextEncoder().encode('data: [DONE]\n'));
            c.close();
        },
    });
    return new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } });
}

describe('CrossdocStore', () => {
    beforeEach(() => {
        mockFetch.mockReset();
    });

    it('options tienen defaults correctos', () => {
        const store = new CrossdocStore();
        expect(store.options).toEqual(DEFAULT_CROSSDOC_OPTIONS);
    });

    it('progress empieza en null', () => {
        const store = new CrossdocStore();
        expect(store.progress).toBeNull();
    });

    it('reset() limpia el progress', () => {
        const store = new CrossdocStore();
        // @ts-ignore — acceso directo para test
        store.progress = { phase: 'done', subQueries: [], completed: 0, total: 0, results: [] };
        store.reset();
        expect(store.progress).toBeNull();
    });

    it('run() llama decompose → subquery → synthesize en orden', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            // decompose call
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query uno test', 'query dos test'] }), { status: 200 }))
            // subquery call × 2
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado uno', success: true }), { status: 200 }))
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado dos', success: true }), { status: 200 }))
            // synthesize call → SSE
            .mockResolvedValueOnce(makeSSEResponse('respuesta final'));

        await store.run('¿Cuál es la presión?', chat);

        expect(mockFetch).toHaveBeenCalledTimes(4);
        expect(mockFetch.mock.calls[0][0]).toContain('/api/crossdoc/decompose');
        expect(mockFetch.mock.calls[1][0]).toContain('/api/crossdoc/subquery');
        expect(mockFetch.mock.calls[3][0]).toContain('/api/crossdoc/synthesize');
    });

    it('run() → chat.messages tiene la respuesta final', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        mockFetch
            .mockResolvedValueOnce(new Response(JSON.stringify({ subQueries: ['query uno test'] }), { status: 200 }))
            .mockResolvedValueOnce(new Response(JSON.stringify({ content: 'resultado', success: true }), { status: 200 }))
            .mockResolvedValueOnce(makeSSEResponse('respuesta final'));

        await store.run('¿test?', chat);

        expect(chat.messages).toHaveLength(1);
        expect(chat.messages[0].role).toBe('assistant');
        expect(chat.messages[0].content).toContain('respuesta final');
    });

    it('stop() aborta el pipeline limpiamente', async () => {
        const store = new CrossdocStore();
        const chat = new ChatStore();

        // decompose devuelve lento (nunca resuelve) → stop() aborta
        mockFetch.mockImplementation(() => new Promise(() => {}));

        const runPromise = store.run('¿test?', chat);
        store.stop();
        await runPromise; // debe resolver sin throw

        expect(chat.streaming).toBe(false);
    });
});
```

- [ ] **Step 2: Correr tests para verificar que FALLAN**

```bash
cd services/sda-frontend && npx vitest run src/lib/stores/crossdoc.svelte.test.ts
```

- [ ] **Step 3: Implementar CrossdocStore**

```typescript
// services/sda-frontend/src/lib/stores/crossdoc.svelte.ts
import type { CrossdocOptions, CrossdocProgress, SubResult } from '$lib/crossdoc/types';
import { DEFAULT_CROSSDOC_OPTIONS } from '$lib/crossdoc/types';
import type { ChatStore } from './chat.svelte';

const MAX_PARALLEL = 6;

export class CrossdocStore {
    progress = $state<CrossdocProgress | null>(null);
    options  = $state<CrossdocOptions>({ ...DEFAULT_CROSSDOC_OPTIONS });

    private abortCtrl: AbortController | null = null;

    async run(question: string, chat: ChatStore): Promise<void> {
        this.abortCtrl = new AbortController();
        const signal = this.abortCtrl.signal;
        chat.startStream();

        try {
            // Phase 1: Decompose
            this.progress = { phase: 'decomposing', subQueries: [], completed: 0, total: 0, results: [] };
            const decompResp = await fetch('/api/crossdoc/decompose', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ question, maxSubQueries: this.options.maxSubQueries }),
                signal,
            });
            if (!decompResp.ok) throw new Error(`Decompose failed: ${decompResp.status}`);
            const { subQueries } = await decompResp.json();

            // Phase 2: Parallel sub-queries
            const results: SubResult[] = [];
            this.progress = { phase: 'querying', subQueries, completed: 0, total: subQueries.length, results };

            const batches: string[][] = [];
            for (let i = 0; i < subQueries.length; i += MAX_PARALLEL) {
                batches.push(subQueries.slice(i, i + MAX_PARALLEL));
            }

            for (const batch of batches) {
                if (signal.aborted) break;
                const batchResults = await Promise.allSettled(
                    batch.map(async (query) => {
                        const resp = await fetch('/api/crossdoc/subquery', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({
                                query,
                                vdbTopK: this.options.vdbTopK,
                                rerankerTopK: this.options.rerankerTopK,
                            }),
                            signal,
                        });
                        if (!resp.ok) return { query, content: '', success: false };
                        const data = await resp.json();
                        return { query, content: data.content, success: data.success } as SubResult;
                    })
                );
                for (const r of batchResults) {
                    results.push(r.status === 'fulfilled' ? r.value : { query: '', content: '', success: false });
                }
                this.progress = { ...this.progress, completed: results.length, results: [...results] };
            }

            // Phase 3: Follow-up retries
            if (this.options.followUpRetries && !signal.aborted) {
                const failed = results.filter(r => !r.success && r.query).map(r => r.query);
                if (failed.length > 0) {
                    this.progress = { ...this.progress, phase: 'retrying' };
                    // Re-usar el endpoint de decompose con prompt de follow-up
                    const altResp = await fetch('/api/crossdoc/decompose', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ question: `Failed queries:\n${failed.join('\n')}\nGenerate alternatives.`, maxSubQueries: failed.length }),
                        signal,
                    });
                    if (altResp.ok) {
                        const { subQueries: alternatives } = await altResp.json();
                        for (const alt of alternatives.slice(0, 10)) {
                            if (signal.aborted) break;
                            const resp = await fetch('/api/crossdoc/subquery', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({ query: alt }),
                                signal,
                            });
                            if (resp.ok) {
                                const data = await resp.json();
                                if (data.success) results.push({ query: alt, content: data.content, success: true });
                            }
                        }
                    }
                }
            }

            // Phase 4: Synthesis → SSE stream
            if (!signal.aborted) {
                this.progress = { ...this.progress, phase: 'synthesizing', results: [...results] };
                const synthResp = await fetch('/api/crossdoc/synthesize', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ question, results }),
                    signal,
                });
                if (synthResp.ok && synthResp.body) {
                    const reader = synthResp.body.getReader();
                    const decoder = new TextDecoder();
                    let buffer = '';
                    try {
                        while (true) {
                            const { done, value } = await reader.read();
                            if (done) break;
                            buffer += decoder.decode(value, { stream: true });
                            const lines = buffer.split('\n');
                            buffer = lines.pop() ?? '';
                            for (const line of lines) {
                                if (!line.startsWith('data: ')) continue;
                                const data = line.slice(6).trim();
                                if (data === '[DONE]') continue;
                                try {
                                    const obj = JSON.parse(data);
                                    const token = obj?.choices?.[0]?.delta?.content;
                                    if (token) chat.appendToken(token);
                                } catch { /* skip */ }
                            }
                        }
                    } finally {
                        reader.releaseLock();
                    }
                }
            }

            this.progress = { phase: 'done', subQueries: this.progress?.subQueries ?? [], completed: results.length, total: results.length, results };
            chat.finalizeStream({ crossdocResults: results });

        } catch (err) {
            if ((err as Error).name === 'AbortError') {
                chat.finalizeStream();
                return;
            }
            this.progress = { phase: 'error', subQueries: [], completed: 0, total: 0, results: [], error: (err as Error).message };
            chat.finalizeStream();
        }
    }

    stop() { this.abortCtrl?.abort(); }
    reset() { this.progress = null; }
}

export const crossdoc = new CrossdocStore();
```

- [ ] **Step 4: Correr tests y verificar que PASAN**

```bash
cd services/sda-frontend && npx vitest run src/lib/stores/crossdoc.svelte.test.ts
```
Esperado: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add services/sda-frontend/src/lib/stores/crossdoc.svelte.ts services/sda-frontend/src/lib/stores/crossdoc.svelte.test.ts
git commit -m "feat(crossdoc): add CrossdocStore with 4-phase pipeline"
```

---

## Task 8: Componentes UI (CrossdocProgress, DecompositionView, clickOutside, CrossdocSettingsPopover)

**Files:**
- Create: `services/sda-frontend/src/lib/actions/clickOutside.ts`
- Create: `services/sda-frontend/src/lib/components/chat/CrossdocProgress.svelte`
- Create: `services/sda-frontend/src/lib/components/chat/DecompositionView.svelte`
- Create: `services/sda-frontend/src/lib/components/chat/CrossdocSettingsPopover.svelte`

> Los componentes UI se verifican visualmente al integrarse en el Task 9. No hay tests unitarios para estos componentes ya que el proyecto no tiene setup de @testing-library/svelte.

- [ ] **Step 1: Crear clickOutside action**

```typescript
// services/sda-frontend/src/lib/actions/clickOutside.ts
import type { Action } from 'svelte/action';

export const clickOutside: Action<HTMLElement, () => void> = (node, callback) => {
    function handle(event: MouseEvent) {
        if (!node.contains(event.target as Node)) {
            callback?.();
        }
    }
    document.addEventListener('click', handle, true);
    return {
        destroy() { document.removeEventListener('click', handle, true); },
        update(newCallback) { callback = newCallback; },
    };
};
```

- [ ] **Step 2: Crear CrossdocProgress.svelte**

```svelte
<!-- services/sda-frontend/src/lib/components/chat/CrossdocProgress.svelte -->
<script lang="ts">
    import { crossdoc } from '$lib/stores/crossdoc.svelte';

    const STEPS = ['decomposing', 'querying', 'retrying', 'synthesizing'] as const;
    const LABELS: Record<string, string> = {
        decomposing: 'Analizando pregunta',
        querying:    'Consultando documentos',
        retrying:    'Reintentando fallidos',
        synthesizing:'Sintetizando respuesta',
    };

    const p = $derived(crossdoc.progress);

    function stepState(step: string): 'done' | 'active' | 'pending' {
        if (!p) return 'pending';
        const currentIdx = STEPS.indexOf(p.phase as any);
        const stepIdx = STEPS.indexOf(step as any);
        if (stepIdx < currentIdx) return 'done';
        if (stepIdx === currentIdx) return 'active';
        return 'pending';
    }
</script>

{#if p && p.phase !== 'done' && p.phase !== 'error'}
    <div class="space-y-2 py-2">
        <!-- Pills de fases -->
        <div class="flex flex-wrap gap-2">
            {#each STEPS as step}
                {@const state = stepState(step)}
                <span class="text-xs px-2 py-0.5 rounded-full border transition-colors
                    {state === 'done'    ? 'bg-green-500/10 text-green-400 border-green-500/30' : ''}
                    {state === 'active'  ? 'bg-[var(--accent)]/10 text-[var(--accent)] border-[var(--accent)]/30 animate-pulse' : ''}
                    {state === 'pending' ? 'bg-transparent text-[var(--text-faint)] border-[var(--border)]' : ''}
                ">
                    {state === 'done' ? '✓' : state === 'active' ? '⟳' : '○'}
                    {LABELS[step]}
                </span>
            {/each}
        </div>

        <!-- Barra de progreso numérica -->
        {#if (p.phase === 'querying' || p.phase === 'retrying') && p.total > 0}
            <div class="space-y-1">
                <div class="h-1 bg-[var(--bg-surface)] rounded-full overflow-hidden">
                    <div class="h-full bg-[var(--accent)] rounded-full transition-all duration-300"
                         style="width: {(p.completed / p.total) * 100}%"></div>
                </div>
                <p class="text-xs text-[var(--text-faint)]">{p.completed} / {p.total} sub-queries</p>
            </div>
        {/if}
    </div>
{/if}
```

- [ ] **Step 3: Crear DecompositionView.svelte**

```svelte
<!-- services/sda-frontend/src/lib/components/chat/DecompositionView.svelte -->
<script lang="ts">
    import type { SubResult } from '$lib/crossdoc/types';

    let { results }: { results: SubResult[] } = $props();
    let open = $state(false);
</script>

{#if results?.length}
    <div class="border-t border-[var(--border)] mt-3 pt-2">
        <button
            onclick={() => open = !open}
            class="flex items-center gap-1 text-xs text-[var(--text-faint)] hover:text-[var(--text)] transition-colors"
        >
            ⚡ Sub-queries usadas ({results.length}) <span>{open ? '▴' : '▾'}</span>
        </button>

        {#if open}
            <ul class="mt-2 space-y-1">
                {#each results as r}
                    <li class="flex items-start gap-2 text-xs">
                        <span class="flex-shrink-0 mt-0.5
                            {r.success ? 'text-green-400' : 'text-red-400'}">
                            {r.success ? '✓' : '✗'}
                        </span>
                        <span class="text-[var(--text-faint)] leading-relaxed">{r.query}</span>
                    </li>
                {/each}
            </ul>
        {/if}
    </div>
{/if}
```

- [ ] **Step 4: Crear CrossdocSettingsPopover.svelte**

```svelte
<!-- services/sda-frontend/src/lib/components/chat/CrossdocSettingsPopover.svelte -->
<script lang="ts">
    import { crossdoc } from '$lib/stores/crossdoc.svelte';
    import { clickOutside } from '$lib/actions/clickOutside';

    let { active, ontoggle }: { active: boolean; ontoggle: () => void } = $props();
    let open = $state(false);
</script>

<div class="relative" use:clickOutside={() => open = false}>
    <!-- Chip toggle -->
    <button
        onclick={() => { ontoggle(); }}
        oncontextmenu|preventDefault={() => open = !open}
        class="flex items-center gap-1 text-xs px-2 py-1 rounded-full border transition-colors
               {active
                   ? 'bg-[var(--accent)]/10 text-[var(--accent)] border-[var(--accent)]/40'
                   : 'bg-transparent text-[var(--text-faint)] border-[var(--border)] hover:border-[var(--accent)]/40'}"
        title="Click: toggle Crossdoc | Click derecho: settings"
    >
        ⚡ Crossdoc
        <span class="opacity-50 text-[10px]" onclick|stopPropagation={() => open = !open}>▾</span>
    </button>

    <!-- Popover de settings -->
    {#if open}
        <div class="absolute bottom-full mb-2 right-0 w-56
                    bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-md)]
                    shadow-lg p-3 space-y-3 z-50">
            <p class="text-xs font-semibold text-[var(--text)]">⚡ Crossdoc Settings</p>

            <label class="flex items-center justify-between gap-2">
                <span class="text-xs text-[var(--text-faint)]">Max sub-queries</span>
                <input
                    type="number" min="0" max="20"
                    bind:value={crossdoc.options.maxSubQueries}
                    class="w-14 text-xs text-center bg-[var(--bg)] border border-[var(--border)]
                           rounded px-1 py-0.5 text-[var(--text)]"
                />
            </label>

            <label class="flex items-center justify-between gap-2">
                <span class="text-xs text-[var(--text-faint)]">Synthesis model</span>
                <input
                    type="text"
                    bind:value={crossdoc.options.synthesisModel}
                    placeholder="(default)"
                    class="w-24 text-xs bg-[var(--bg)] border border-[var(--border)]
                           rounded px-1 py-0.5 text-[var(--text)] placeholder-[var(--text-faint)]"
                />
            </label>

            <label class="flex items-center justify-between gap-2">
                <span class="text-xs text-[var(--text-faint)]">Follow-up retries</span>
                <input type="checkbox" bind:checked={crossdoc.options.followUpRetries}
                       class="accent-[var(--accent)]" />
            </label>

            <label class="flex items-center justify-between gap-2">
                <span class="text-xs text-[var(--text-faint)]">Show decomposition</span>
                <input type="checkbox" bind:checked={crossdoc.options.showDecomposition}
                       class="accent-[var(--accent)]" />
            </label>
        </div>
    {/if}
</div>
```

- [ ] **Step 5: Verificar que TypeScript compila sin errores**

```bash
cd services/sda-frontend && npx tsc --noEmit
```

- [ ] **Step 6: Commit**

```bash
git add services/sda-frontend/src/lib/actions/ \
        services/sda-frontend/src/lib/components/chat/CrossdocProgress.svelte \
        services/sda-frontend/src/lib/components/chat/DecompositionView.svelte \
        services/sda-frontend/src/lib/components/chat/CrossdocSettingsPopover.svelte
git commit -m "feat(crossdoc): add CrossdocProgress, DecompositionView, CrossdocSettingsPopover components"
```

---

## Task 9: Integrar en ChatInput, chat page y MessageList

**Files:**
- Modify: `services/sda-frontend/src/lib/components/chat/ChatInput.svelte`
- Modify: `services/sda-frontend/src/routes/(app)/chat/[id]/+page.svelte`
- Modify: `services/sda-frontend/src/lib/components/chat/MessageList.svelte`

- [ ] **Step 1: Modificar ChatInput.svelte**

Reemplazar el archivo completo:

```svelte
<!-- services/sda-frontend/src/lib/components/chat/ChatInput.svelte -->
<script lang="ts">
    import { Send, Square } from 'lucide-svelte';
    import CrossdocSettingsPopover from './CrossdocSettingsPopover.svelte';

    interface Props {
        streaming: boolean;
        crossdoc: boolean;
        onsubmit: (query: string) => void;
        onstop: () => void;
        oncrossdoctoggle: () => void;
    }

    let { streaming, crossdoc, onsubmit, onstop, oncrossdoctoggle }: Props = $props();

    let input = $state('');

    function handleSubmit() {
        const query = input.trim();
        if (!query || streaming) return;
        input = '';
        onsubmit(query);
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
        }
    }
</script>

<div class="p-3 border-t border-[var(--border)] flex-shrink-0">
    <div
        class="flex gap-2 bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-md)]
               px-3 py-2 focus-within:border-[var(--accent)] transition-colors"
    >
        <textarea
            bind:value={input}
            onkeydown={handleKeydown}
            rows={1}
            placeholder="Escribí tu consulta..."
            disabled={streaming}
            class="flex-1 bg-transparent text-xs text-[var(--text)] placeholder-[var(--text-faint)]
                   resize-none outline-none disabled:opacity-60"
            style="max-height: 120px; overflow-y: auto;"
        ></textarea>

        <div class="flex items-center gap-2 flex-shrink-0">
            <!-- Crossdoc toggle -->
            <CrossdocSettingsPopover active={crossdoc} ontoggle={oncrossdoctoggle} />

            {#if streaming}
                <button
                    onclick={onstop}
                    title="Detener generación"
                    class="text-[var(--danger)] hover:opacity-80 transition-opacity"
                >
                    <Square size={14} fill="currentColor" />
                </button>
            {:else}
                <button
                    onclick={handleSubmit}
                    disabled={!input.trim()}
                    title="Enviar (Enter)"
                    class="text-[var(--accent)] hover:text-[var(--accent-hover)]
                           disabled:opacity-40 transition-colors"
                >
                    <Send size={16} />
                </button>
            {/if}
        </div>
    </div>
</div>
```

- [ ] **Step 2: Modificar chat/[id]/+page.svelte**

Leer el archivo actual y aplicar estos cambios:

**a) Agregar imports al bloque `<script>`:**
```typescript
import { crossdoc } from '$lib/stores/crossdoc.svelte';
import CrossdocProgress from '$lib/components/chat/CrossdocProgress.svelte';
```

**b) Extraer la lógica de streaming normal a una función `streamNormal()`:**
```typescript
async function streamNormal(query: string) {
    // Mover aquí el cuerpo actual de sendMessage (el fetch a /api/chat/stream/[id])
}
```

**c) Reemplazar `sendMessage` para branching crossdoc:**
```typescript
async function sendMessage(query: string) {
    chat.addUserMessage(query);
    if (chat.crossdoc) {
        await crossdoc.run(query, chat);
    } else {
        await streamNormal(query);
    }
}
```

**d) Agregar prop `crossdoc` y `oncrossdoctoggle` al `<ChatInput>`:**
```svelte
<ChatInput
    streaming={chat.streaming}
    crossdoc={chat.crossdoc}
    onsubmit={sendMessage}
    onstop={() => { chat.stopStream(); crossdoc.stop(); }}
    oncrossdoctoggle={() => { chat.crossdoc = !chat.crossdoc; }}
/>
```

**e) En el template, mostrar CrossdocProgress dentro del assistant bubble cuando streaming + crossdoc:**

Buscar donde se renderiza el `streamingContent` (el bubble del asistente en progreso) y agregar arriba del texto:
```svelte
{#if chat.streaming && chat.crossdoc && !chat.streamingContent}
    <CrossdocProgress />
{/if}
```

- [ ] **Step 3: Leer MessageList.svelte y agregar DecompositionView**

```bash
# Leer el archivo para entender su estructura antes de modificar
```

Agregar en el template de MessageList, luego de renderizar el contenido de cada mensaje assistant:
```svelte
import DecompositionView from './DecompositionView.svelte';
import { crossdoc } from '$lib/stores/crossdoc.svelte';

<!-- Dentro del loop de mensajes, luego del contenido assistant: -->
{#if msg.role === 'assistant' && msg.crossdocResults && crossdoc.options.showDecomposition}
    <DecompositionView results={msg.crossdocResults} />
{/if}
```

- [ ] **Step 4: Verificar TypeScript sin errores**

```bash
cd services/sda-frontend && npx tsc --noEmit
```
Esperado: 0 errores

- [ ] **Step 5: Correr todos los tests del proyecto**

```bash
cd services/sda-frontend && npx vitest run
```
Esperado: todos PASS

- [ ] **Step 6: Smoke test manual**
1. `cd services/sda-frontend && npm run dev`
2. Navegar a `/chat/[cualquier-sesión]`
3. Verificar que aparece el chip `⚡ Crossdoc` en el input
4. Click derecho (o flecha ▾) → settings popover abre
5. Activar Crossdoc con click en el chip → chip se pone azul
6. Enviar un mensaje → verificar que aparece el CrossdocProgress con las 4 pills
7. Activar "Show decomposition" en settings → verificar accordion debajo de la respuesta

- [ ] **Step 7: Commit final**

```bash
git add services/sda-frontend/src/lib/components/chat/ChatInput.svelte \
        services/sda-frontend/src/routes/\(app\)/chat/\[id\]/+page.svelte \
        services/sda-frontend/src/lib/components/chat/MessageList.svelte
git commit -m "feat(crossdoc): wire up Crossdoc UI in ChatInput, chat page and MessageList"
```

---

## Checklist final

- [ ] `npx vitest run` — todos los tests pasan
- [ ] `npx tsc --noEmit` — TypeScript sin errores
- [ ] Smoke test manual: pipeline completo funciona de punta a punta
- [ ] El path normal (non-crossdoc) sigue funcionando sin regresiones
