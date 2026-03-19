# Fase 2 — Chat Pro: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mejorar la experiencia de chat con Markdown rendering completo, historial con búsqueda, panel de fuentes toggleable, stop streaming y smart auto-scroll.

**Architecture:** Se extraen 5 componentes Svelte dedicados (`HistoryPanel`, `MessageList`, `MarkdownRenderer`, `SourcesPanel`, `ChatInput`) y 3 utilidades TypeScript testeables. El `[id]/+page.svelte` pasa a ser un coordinador de ~80 líneas. `ChatStore` se extiende con `AbortController` para el stop streaming.

**Tech Stack:** SvelteKit 5 (runes), Svelte 5, Tailwind CSS 4, TypeScript, `marked` + `marked-highlight` + `highlight.js` (Markdown), `dompurify` (sanitización XSS), Vitest (tests).

---

## Spec de referencia
`docs/superpowers/specs/2026-03-19-fase2-chat-pro-design.md`

## Contexto clave del proyecto
- Tests corren con: `cd services/sda-frontend && npm run test` (Vitest, environment: node)
- Archivos existentes relevantes:
  - `src/lib/stores/chat.svelte.ts` — ChatStore (212 líneas de page.svelte usa este store)
  - `src/routes/(app)/chat/[id]/+page.svelte` — página actual a refactorizar
  - `src/lib/stores/toast.svelte.test.ts` — ejemplo de patrón de tests
- CSS usa custom properties: `var(--text)`, `var(--accent)`, `var(--border)`, `var(--bg-surface)`, `var(--bg-base)`, `var(--bg-hover)`, `var(--radius-sm)`, `var(--radius-md)`
- Tailwind 4 con `@custom-variant dark (&:where(.dark, .dark *))`

## Mapa de archivos

| Acción | Archivo |
|--------|---------|
| Crear | `src/lib/utils/markdown.ts` |
| Crear | `src/lib/utils/markdown.test.ts` |
| Crear | `src/lib/utils/chat-utils.ts` |
| Crear | `src/lib/utils/chat-utils.test.ts` |
| Crear | `src/lib/utils/scroll.ts` |
| Crear | `src/lib/utils/scroll.test.ts` |
| Crear | `src/lib/stores/chat.svelte.test.ts` |
| Crear | `src/lib/components/chat/MarkdownRenderer.svelte` |
| Crear | `src/lib/components/chat/HistoryPanel.svelte` |
| Crear | `src/lib/components/chat/SourcesPanel.svelte` |
| Crear | `src/lib/components/chat/ChatInput.svelte` |
| Crear | `src/lib/components/chat/MessageList.svelte` |
| Modificar | `src/lib/stores/chat.svelte.ts` |
| Modificar | `src/routes/(app)/chat/[id]/+page.svelte` |
| Modificar | `package.json` (en `services/sda-frontend/`) |

---

## Task 1: Instalar dependencias y extender ChatStore con AbortController

**Files:**
- Modify: `services/sda-frontend/package.json`
- Modify: `services/sda-frontend/src/lib/stores/chat.svelte.ts`
- Create: `services/sda-frontend/src/lib/stores/chat.svelte.test.ts`

- [ ] **Step 1: Escribir tests que fallan para ChatStore**

Crear `services/sda-frontend/src/lib/stores/chat.svelte.test.ts`:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { ChatStore } from './chat.svelte.js';

describe('ChatStore', () => {
    it('startStream crea abortController y activa streaming', () => {
        const chat = new ChatStore();
        chat.startStream();
        expect(chat.abortController).not.toBeNull();
        expect(chat.abortController).toBeInstanceOf(AbortController);
        expect(chat.streaming).toBe(true);
        expect(chat.streamingContent).toBe('');
        expect(chat.sources).toHaveLength(0);
    });

    it('stopStream llama abort y deja streaming en false', () => {
        const chat = new ChatStore();
        chat.startStream();
        const controller = chat.abortController!;
        const abortSpy = vi.spyOn(controller, 'abort');
        chat.stopStream();
        expect(abortSpy).toHaveBeenCalledOnce();
        expect(chat.streaming).toBe(false);
        expect(chat.streamingContent).toBe('');
    });

    it('stopStream con contenido parcial guarda lo que llegó', () => {
        const chat = new ChatStore();
        chat.startStream();
        chat.appendToken('texto parcial');
        chat.stopStream();
        expect(chat.messages).toHaveLength(1);
        expect(chat.messages[0].content).toBe('texto parcial');
        expect(chat.messages[0].role).toBe('assistant');
    });

    it('stopStream sin startStream previo no lanza error', () => {
        const chat = new ChatStore();
        expect(() => chat.stopStream()).not.toThrow();
    });
});
```

- [ ] **Step 2: Correr tests — deben fallar**

```bash
cd /ruta/del/proyecto/services/sda-frontend && npm run test -- chat.svelte
```

Esperado: FAIL — `stopStream is not a function` o similar.

- [ ] **Step 3: Instalar dependencias**

```bash
cd services/sda-frontend && npm install marked marked-highlight highlight.js dompurify
npm install --save-dev @types/dompurify
```

- [ ] **Step 4: Modificar `src/lib/stores/chat.svelte.ts`**

Reemplazar el archivo completo con:

```typescript
// Svelte 5 runes-based reactive store for chat state

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

export class ChatStore {
    messages = $state<Message[]>([]);
    sources = $state<Source[]>([]);
    streaming = $state(false);
    streamingContent = $state('');
    collection = $state('');
    crossdoc = $state(false);
    abortController = $state<AbortController | null>(null);

    addUserMessage(content: string) {
        this.messages.push({ role: 'user', content, timestamp: new Date().toISOString() });
    }

    startStream() {
        this.abortController = new AbortController();
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

    stopStream() {
        this.abortController?.abort();
        this.finalizeStream();
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
```

- [ ] **Step 5: Correr tests — deben pasar**

```bash
cd services/sda-frontend && npm run test -- chat.svelte
```

Esperado: PASS — 4 tests en `ChatStore`.

- [ ] **Step 6: Commit**

```bash
git add services/sda-frontend/package.json services/sda-frontend/package-lock.json \
        services/sda-frontend/src/lib/stores/chat.svelte.ts \
        services/sda-frontend/src/lib/stores/chat.svelte.test.ts
git commit -m "feat(chat): add AbortController to ChatStore for stop streaming"
```

---

## Task 2: Utilidad parseMarkdown + tests

**Files:**
- Create: `services/sda-frontend/src/lib/utils/markdown.ts`
- Create: `services/sda-frontend/src/lib/utils/markdown.test.ts`

- [ ] **Step 1: Escribir tests que fallan**

Crear `services/sda-frontend/src/lib/utils/markdown.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { parseMarkdown } from './markdown.js';

describe('parseMarkdown', () => {
    it('convierte texto en negrita', () => {
        const result = parseMarkdown('**negrita**');
        expect(result).toContain('<strong>negrita</strong>');
    });

    it('convierte cursiva', () => {
        const result = parseMarkdown('*cursiva*');
        expect(result).toContain('<em>cursiva</em>');
    });

    it('convierte header h1', () => {
        const result = parseMarkdown('# Título');
        expect(result).toContain('<h1>');
        expect(result).toContain('Título');
    });

    it('convierte lista desordenada', () => {
        const result = parseMarkdown('- item uno\n- item dos');
        expect(result).toContain('<li>item uno</li>');
        expect(result).toContain('<li>item dos</li>');
    });

    it('convierte bloque de código con tag pre y code', () => {
        const result = parseMarkdown('```python\nprint("hola")\n```');
        expect(result).toContain('<pre>');
        expect(result).toContain('<code');
    });

    it('retorna string, no promesa', () => {
        const result = parseMarkdown('**texto**');
        expect(typeof result).toBe('string');
    });

    it('maneja string vacío sin error', () => {
        expect(() => parseMarkdown('')).not.toThrow();
    });
});
```

- [ ] **Step 2: Correr tests — deben fallar**

```bash
cd services/sda-frontend && npm run test -- markdown
```

Esperado: FAIL — `Cannot find module './markdown.js'`.

- [ ] **Step 3: Crear `src/lib/utils/markdown.ts`**

```typescript
import { marked } from 'marked';
import { markedHighlight } from 'marked-highlight';
import hljs from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';
import python from 'highlight.js/lib/languages/python';
import javascript from 'highlight.js/lib/languages/javascript';
import typescript from 'highlight.js/lib/languages/typescript';
import json from 'highlight.js/lib/languages/json';
import yaml from 'highlight.js/lib/languages/yaml';
import sql from 'highlight.js/lib/languages/sql';

// Registrar lenguajes para syntax highlighting
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('python', python);
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('json', json);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('sql', sql);

// Configurar marked con highlight.js (una sola vez al cargar el módulo)
marked.use(markedHighlight({
    langPrefix: 'hljs language-',
    highlight(code, lang) {
        const language = hljs.getLanguage(lang) ? lang : 'plaintext';
        return hljs.highlight(code, { language }).value;
    }
}));

/**
 * Convierte Markdown a HTML. Sanitización XSS se aplica
 * en el componente MarkdownRenderer (solo en browser).
 */
export function parseMarkdown(content: string): string {
    return marked.parse(content) as string;
}
```

- [ ] **Step 4: Correr tests — deben pasar**

```bash
cd services/sda-frontend && npm run test -- markdown
```

Esperado: PASS — 7 tests en `parseMarkdown`.

- [ ] **Step 5: Commit**

```bash
git add services/sda-frontend/src/lib/utils/
git commit -m "feat(chat): add parseMarkdown utility with hljs syntax highlighting"
```

---

## Task 3: Componente MarkdownRenderer

**Files:**
- Create: `services/sda-frontend/src/lib/components/chat/MarkdownRenderer.svelte`

No hay lógica pura testeable (usa DOMPurify con APIs del browser). El componente se verifica visualmente en Task 8.

- [ ] **Step 1: Crear directorio `src/lib/components/chat/`**

```bash
mkdir -p services/sda-frontend/src/lib/components/chat
```

- [ ] **Step 2: Crear `src/lib/components/chat/MarkdownRenderer.svelte`**

```svelte
<script lang="ts">
    import { parseMarkdown } from '$lib/utils/markdown';

    let { content }: { content: string } = $props();

    let sanitizedHtml = $state('');

    $effect(async () => {
        const raw = parseMarkdown(content);
        // DOMPurify requiere browser — import dinámico evita crash en SSR
        try {
            const { default: DOMPurify } = await import('dompurify');
            sanitizedHtml = DOMPurify.sanitize(raw, {
                ALLOWED_TAGS: [
                    'p', 'br', 'strong', 'em', 'del', 'code', 'pre',
                    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
                    'ul', 'ol', 'li', 'blockquote', 'hr',
                    'table', 'thead', 'tbody', 'tr', 'th', 'td',
                    'a', 'span', 'div'
                ],
                ALLOWED_ATTR: ['href', 'class', 'target', 'rel'],
            });
        } catch {
            // Fallback sin sanitización (solo en SSR o si DOMPurify no carga)
            sanitizedHtml = raw;
        }
    });
</script>

<div class="markdown-body text-xs text-[var(--text)] leading-relaxed">
    {@html sanitizedHtml}
</div>

<style>
    .markdown-body :global(h1),
    .markdown-body :global(h2),
    .markdown-body :global(h3) {
        color: var(--text);
        font-weight: 600;
        margin: 0.6em 0 0.2em;
        line-height: 1.3;
    }
    .markdown-body :global(h1) { font-size: 1.1em; }
    .markdown-body :global(h2) { font-size: 1.0em; }
    .markdown-body :global(h3) { font-size: 0.95em; }

    .markdown-body :global(p) { margin: 0.4em 0; line-height: 1.7; }

    .markdown-body :global(ul),
    .markdown-body :global(ol) { padding-left: 1.25em; margin: 0.3em 0; }

    .markdown-body :global(li) { margin: 0.15em 0; }

    .markdown-body :global(strong) { color: var(--text); font-weight: 600; }

    .markdown-body :global(code) {
        background: var(--bg-base);
        border: 1px solid var(--border);
        border-radius: 3px;
        padding: 0.1em 0.35em;
        font-family: 'Cascadia Code', 'Fira Code', monospace;
        font-size: 0.85em;
    }

    .markdown-body :global(pre) {
        background: var(--bg-base);
        border: 1px solid var(--border);
        border-radius: var(--radius-md);
        padding: 0.75em 1em;
        overflow-x: auto;
        margin: 0.5em 0;
    }

    .markdown-body :global(pre code) {
        background: none;
        border: none;
        padding: 0;
        font-size: 0.8em;
        line-height: 1.6;
    }

    .markdown-body :global(table) {
        border-collapse: collapse;
        width: 100%;
        margin: 0.5em 0;
        font-size: 0.85em;
    }

    .markdown-body :global(th),
    .markdown-body :global(td) {
        border: 1px solid var(--border);
        padding: 0.3em 0.6em;
        text-align: left;
    }

    .markdown-body :global(th) {
        background: var(--bg-surface);
        font-weight: 600;
        color: var(--text-muted);
    }

    .markdown-body :global(a) {
        color: var(--accent);
        text-decoration: none;
    }

    .markdown-body :global(a:hover) { text-decoration: underline; }

    .markdown-body :global(blockquote) {
        border-left: 3px solid var(--accent);
        padding-left: 0.75em;
        margin: 0.4em 0;
        color: var(--text-muted);
        font-style: italic;
    }

    .markdown-body :global(hr) {
        border: none;
        border-top: 1px solid var(--border);
        margin: 0.5em 0;
    }

    /* highlight.js colors — tema oscuro custom */
    .markdown-body :global(.hljs) { color: #abb2bf; background: transparent; }
    .markdown-body :global(.hljs-keyword),
    .markdown-body :global(.hljs-selector-tag) { color: #c678dd; }
    .markdown-body :global(.hljs-string) { color: #98c379; }
    .markdown-body :global(.hljs-number) { color: #d19a66; }
    .markdown-body :global(.hljs-comment) { color: #5c6370; font-style: italic; }
    .markdown-body :global(.hljs-title),
    .markdown-body :global(.hljs-function) { color: #61afef; }
    .markdown-body :global(.hljs-attr) { color: #e06c75; }
    .markdown-body :global(.hljs-literal) { color: #56b6c2; }
    .markdown-body :global(.hljs-built_in) { color: #e6c07b; }
    .markdown-body :global(.hljs-variable) { color: #e06c75; }
</style>
```

- [ ] **Step 3: Correr todos los tests existentes — deben seguir pasando**

```bash
cd services/sda-frontend && npm run test
```

Esperado: todos los tests previos en PASS (no hay tests nuevos en esta tarea).

- [ ] **Step 4: Commit**

```bash
git add services/sda-frontend/src/lib/components/chat/MarkdownRenderer.svelte
git commit -m "feat(chat): add MarkdownRenderer component with DOMPurify + hljs"
```

---

## Task 4: Utilidad filterSessions + componente HistoryPanel

**Files:**
- Create: `services/sda-frontend/src/lib/utils/chat-utils.ts`
- Create: `services/sda-frontend/src/lib/utils/chat-utils.test.ts`
- Create: `services/sda-frontend/src/lib/components/chat/HistoryPanel.svelte`

- [ ] **Step 1: Escribir tests que fallan**

Crear `services/sda-frontend/src/lib/utils/chat-utils.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { filterSessions } from './chat-utils.js';

describe('filterSessions', () => {
    const sessions = [
        { id: '1', title: 'Manual Aries 365', updated_at: '2026-03-18T10:00:00' },
        { id: '2', title: 'Normativas homologación', updated_at: '2026-03-17T10:00:00' },
        { id: '3', title: 'Motor ZF especificaciones', updated_at: '2026-03-16T10:00:00' },
    ];

    it('retorna todas las sesiones cuando query es vacío', () => {
        expect(filterSessions(sessions, '')).toHaveLength(3);
    });

    it('retorna todas las sesiones cuando query es solo espacios', () => {
        expect(filterSessions(sessions, '   ')).toHaveLength(3);
    });

    it('filtra por título case-insensitive', () => {
        const result = filterSessions(sessions, 'aries');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('1');
    });

    it('filtra con mayúsculas en el query', () => {
        const result = filterSessions(sessions, 'MOTOR');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('3');
    });

    it('retorna array vacío si no hay coincidencias', () => {
        expect(filterSessions(sessions, 'inexistente xyz')).toHaveLength(0);
    });

    it('encuentra coincidencias parciales', () => {
        const result = filterSessions(sessions, 'homol');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('2');
    });
});
```

- [ ] **Step 2: Correr tests — deben fallar**

```bash
cd services/sda-frontend && npm run test -- chat-utils
```

Esperado: FAIL — `Cannot find module './chat-utils.js'`.

- [ ] **Step 3: Crear `src/lib/utils/chat-utils.ts`**

```typescript
export interface SessionSummary {
    id: string;
    title: string;
    updated_at: string;
}

/**
 * Filtra sesiones por título (case-insensitive, substring match).
 * Retorna todas si query es vacío o solo espacios.
 */
export function filterSessions(sessions: SessionSummary[], query: string): SessionSummary[] {
    const q = query.trim().toLowerCase();
    if (!q) return sessions;
    return sessions.filter(s => s.title.toLowerCase().includes(q));
}
```

- [ ] **Step 4: Correr tests — deben pasar**

```bash
cd services/sda-frontend && npm run test -- chat-utils
```

Esperado: PASS — 6 tests en `filterSessions`.

- [ ] **Step 5: Crear `src/lib/components/chat/HistoryPanel.svelte`**

```svelte
<script lang="ts">
    import { filterSessions, type SessionSummary } from '$lib/utils/chat-utils';

    interface Props {
        sessions: SessionSummary[];
        currentId: string;
    }

    let { sessions, currentId }: Props = $props();

    let searchQuery = $state('');
    let filtered = $derived(filterSessions(sessions, searchQuery));
</script>

<div class="w-[200px] flex-shrink-0 bg-[var(--bg-surface)] border-r border-[var(--border)] flex flex-col overflow-hidden">
    <!-- Header con búsqueda -->
    <div class="px-2 pt-3 pb-2 flex-shrink-0">
        <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider mb-2">
            Historial
        </div>
        <input
            bind:value={searchQuery}
            type="search"
            placeholder="Buscar..."
            class="w-full bg-[var(--bg-base)] border border-[var(--border)] rounded-[var(--radius-sm)]
                   px-2 py-1 text-xs text-[var(--text)] placeholder-[var(--text-faint)]
                   outline-none focus:border-[var(--accent)] transition-colors"
        />
    </div>

    <!-- Nueva consulta -->
    <div class="px-2 pb-1 flex-shrink-0">
        <a
            href="/chat"
            data-sveltekit-preload-data="false"
            class="flex items-center gap-1 text-xs text-[var(--accent)] hover:underline"
        >
            + Nueva consulta
        </a>
    </div>

    <!-- Lista de sesiones -->
    <div class="flex-1 overflow-y-auto px-2 pb-2 flex flex-col gap-0.5">
        {#each filtered as session (session.id)}
            <a
                href="/chat/{session.id}"
                data-sveltekit-preload-data="false"
                class="block rounded-[var(--radius-sm)] px-2 py-1.5 transition-colors
                       {session.id === currentId
                           ? 'bg-[var(--bg-hover)] border-l-2 border-[var(--accent)] pl-[7px]'
                           : 'hover:bg-[var(--bg-hover)]'}"
            >
                <div class="text-xs text-[var(--text-muted)] font-medium truncate leading-tight">
                    {session.title}
                </div>
                <div class="text-[10px] text-[var(--text-faint)] mt-0.5">
                    {session.updated_at.slice(0, 10)}
                </div>
            </a>
        {/each}

        {#if filtered.length === 0 && searchQuery.trim()}
            <p class="text-xs text-[var(--text-faint)] px-2 py-4 text-center">
                Sin resultados
            </p>
        {/if}
    </div>
</div>
```

- [ ] **Step 6: Correr todos los tests**

```bash
cd services/sda-frontend && npm run test
```

Esperado: PASS — todos los tests incluyendo los 6 de `filterSessions`.

- [ ] **Step 7: Commit**

```bash
git add services/sda-frontend/src/lib/utils/chat-utils.ts \
        services/sda-frontend/src/lib/utils/chat-utils.test.ts \
        services/sda-frontend/src/lib/components/chat/HistoryPanel.svelte
git commit -m "feat(chat): add filterSessions utility and HistoryPanel component"
```

---

## Task 5: Componente SourcesPanel

**Files:**
- Create: `services/sda-frontend/src/lib/components/chat/SourcesPanel.svelte`

No hay lógica pura testeable (solo renderizado reactivo de props).

- [ ] **Step 1: Crear `src/lib/components/chat/SourcesPanel.svelte`**

```svelte
<script lang="ts">
    import type { Source } from '$lib/stores/chat.svelte';

    interface Props {
        sources: Source[];
        open: boolean;
    }

    let { sources, open }: Props = $props();
</script>

<div
    class="flex-shrink-0 bg-[var(--bg-surface)] border-l border-[var(--border)] overflow-hidden
           transition-[width] duration-200 ease-in-out
           {open && sources.length > 0 ? 'w-64' : 'w-0'}"
>
    <!-- Contenedor con ancho fijo para que no se comprima durante la transición -->
    <div class="w-64 h-full flex flex-col p-3 overflow-y-auto">
        <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider mb-3 flex-shrink-0">
            Fuentes ({sources.length})
        </div>

        {#each sources as source, i (source.document + i)}
            <div
                class="mb-2 rounded-[var(--radius-sm)] p-2 border-l-2 bg-[var(--bg-base)]
                       {i === 0
                           ? 'border-[var(--accent)]'
                           : i === 1
                           ? 'border-[var(--accent-hover)]'
                           : 'border-[var(--border)]'}"
            >
                <div class="text-[11px] text-[var(--accent)] font-semibold truncate">
                    {source.document}
                </div>
                {#if source.page}
                    <div class="text-[10px] text-[var(--text-faint)] mt-0.5">
                        p. {source.page}
                    </div>
                {/if}
                <div class="text-[10px] text-[var(--text-muted)] mt-1 line-clamp-4 leading-relaxed">
                    {source.excerpt}
                </div>
            </div>
        {/each}
    </div>
</div>
```

- [ ] **Step 2: Correr todos los tests**

```bash
cd services/sda-frontend && npm run test
```

Esperado: PASS — sin cambios en tests.

- [ ] **Step 3: Commit**

```bash
git add services/sda-frontend/src/lib/components/chat/SourcesPanel.svelte
git commit -m "feat(chat): add SourcesPanel component with animated toggle"
```

---

## Task 6: Componente ChatInput

**Files:**
- Create: `services/sda-frontend/src/lib/components/chat/ChatInput.svelte`

No hay lógica pura testeable (manejo de eventos DOM).

- [ ] **Step 1: Crear `src/lib/components/chat/ChatInput.svelte`**

```svelte
<script lang="ts">
    import { Send, Square } from 'lucide-svelte';

    interface Props {
        streaming: boolean;
        onsubmit: (query: string) => void;
        onstop: () => void;
    }

    let { streaming, onsubmit, onstop }: Props = $props();

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

        {#if streaming}
            <button
                onclick={onstop}
                title="Detener generación"
                class="flex-shrink-0 text-[var(--danger)] hover:opacity-80 transition-opacity"
            >
                <Square size={14} fill="currentColor" />
            </button>
        {:else}
            <button
                onclick={handleSubmit}
                disabled={!input.trim()}
                title="Enviar (Enter)"
                class="flex-shrink-0 text-[var(--accent)] hover:text-[var(--accent-hover)]
                       disabled:opacity-40 transition-colors"
            >
                <Send size={16} />
            </button>
        {/if}
    </div>
</div>
```

- [ ] **Step 2: Correr todos los tests**

```bash
cd services/sda-frontend && npm run test
```

Esperado: PASS.

- [ ] **Step 3: Commit**

```bash
git add services/sda-frontend/src/lib/components/chat/ChatInput.svelte
git commit -m "feat(chat): add ChatInput component with stop streaming button"
```

---

## Task 7: Utilidad isNearBottom + componente MessageList

**Files:**
- Create: `services/sda-frontend/src/lib/utils/scroll.ts`
- Create: `services/sda-frontend/src/lib/utils/scroll.test.ts`
- Create: `services/sda-frontend/src/lib/components/chat/MessageList.svelte`

- [ ] **Step 1: Escribir tests que fallan**

Crear `services/sda-frontend/src/lib/utils/scroll.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { isNearBottom } from './scroll.js';

describe('isNearBottom', () => {
    it('retorna true cuando está en el fondo exacto', () => {
        // scrollHeight - scrollTop - clientHeight = 0
        const el = { scrollHeight: 1000, scrollTop: 900, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(true);
    });

    it('retorna true cuando está dentro del threshold (por defecto 100px)', () => {
        // 1000 - 850 - 100 = 50 < 100
        const el = { scrollHeight: 1000, scrollTop: 850, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(true);
    });

    it('retorna false cuando está más arriba del threshold', () => {
        // 1000 - 700 - 100 = 200 > 100
        const el = { scrollHeight: 1000, scrollTop: 700, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(false);
    });

    it('retorna false cuando está al tope', () => {
        const el = { scrollHeight: 1000, scrollTop: 0, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(false);
    });

    it('respeta threshold customizado', () => {
        const el = { scrollHeight: 1000, scrollTop: 850, clientHeight: 100 };
        // distancia = 50
        expect(isNearBottom(el, 60)).toBe(true);   // 50 < 60
        expect(isNearBottom(el, 40)).toBe(false);  // 50 > 40
    });

    it('retorna true cuando el contenido es más corto que el viewport', () => {
        // scrollHeight <= clientHeight → siempre en el fondo
        const el = { scrollHeight: 80, scrollTop: 0, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(true);
    });
});
```

- [ ] **Step 2: Correr tests — deben fallar**

```bash
cd services/sda-frontend && npm run test -- scroll
```

Esperado: FAIL — `Cannot find module './scroll.js'`.

- [ ] **Step 3: Crear `src/lib/utils/scroll.ts`**

```typescript
interface ScrollMetrics {
    scrollHeight: number;
    scrollTop: number;
    clientHeight: number;
}

/**
 * Retorna true si el elemento está dentro de `threshold` píxeles del fondo.
 * Útil para decidir si auto-scrollear o mostrar el botón "↓ Ir al fondo".
 */
export function isNearBottom(el: ScrollMetrics, threshold = 100): boolean {
    return el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
}
```

- [ ] **Step 4: Correr tests — deben pasar**

```bash
cd services/sda-frontend && npm run test -- scroll
```

Esperado: PASS — 6 tests en `isNearBottom`.

- [ ] **Step 5: Crear `src/lib/components/chat/MessageList.svelte`**

```svelte
<script lang="ts">
    import { ChevronDown } from 'lucide-svelte';
    import { isNearBottom } from '$lib/utils/scroll';
    import MarkdownRenderer from './MarkdownRenderer.svelte';

    interface Message {
        role: 'user' | 'assistant';
        content: string;
        timestamp: string;
    }

    interface Props {
        messages: Message[];
        streaming: boolean;
        streamingContent: string;
    }

    let { messages, streaming, streamingContent }: Props = $props();

    let scrollEl = $state<HTMLDivElement | null>(null);
    let showScrollButton = $state(false);

    function handleScroll() {
        if (!scrollEl) return;
        showScrollButton = !isNearBottom(scrollEl);
    }

    function scrollToBottom() {
        if (!scrollEl) return;
        scrollEl.scrollTo({ top: scrollEl.scrollHeight, behavior: 'smooth' });
        showScrollButton = false;
    }

    // Auto-scroll: solo si el usuario ya está cerca del fondo
    $effect(() => {
        // Dependencias que disparan el efecto:
        void messages.length;
        void streamingContent.length;

        if (!scrollEl) return;
        if (isNearBottom(scrollEl)) {
            scrollEl.scrollTo({ top: scrollEl.scrollHeight, behavior: 'smooth' });
        }
    });
</script>

<div class="relative flex-1 overflow-hidden">
    <div
        bind:this={scrollEl}
        onscroll={handleScroll}
        class="h-full overflow-y-auto p-3 flex flex-col gap-3"
    >
        {#each messages as msg (msg.timestamp)}
            {#if msg.role === 'user'}
                <div class="flex justify-end">
                    <div class="bg-[var(--accent)] rounded-lg rounded-tr-sm px-3 py-2 max-w-[70%]">
                        <p class="text-xs text-white whitespace-pre-wrap">{msg.content}</p>
                    </div>
                </div>
            {:else}
                <div class="flex gap-2">
                    <div class="w-5 h-5 bg-[var(--accent)] rounded-full flex-shrink-0 mt-0.5"></div>
                    <div class="bg-[var(--bg-surface)] border border-[var(--border)]
                                rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%] min-w-0">
                        <MarkdownRenderer content={msg.content} />
                    </div>
                </div>
            {/if}
        {/each}

        <!-- Mensaje en streaming -->
        {#if streaming}
            <div class="flex gap-2">
                <div class="w-5 h-5 bg-[var(--accent)] rounded-full animate-pulse flex-shrink-0 mt-0.5"></div>
                <div class="bg-[var(--bg-surface)] border border-[var(--border)]
                            rounded-lg rounded-tl-sm px-3 py-2 max-w-[80%] min-w-0">
                    <MarkdownRenderer content={streamingContent} />
                    <span class="text-[var(--text-faint)] animate-pulse">▋</span>
                </div>
            </div>
        {/if}
    </div>

    <!-- Botón "Ir al fondo" — aparece cuando el usuario scrolleó arriba -->
    {#if showScrollButton}
        <button
            onclick={scrollToBottom}
            title="Ir al fondo"
            class="absolute bottom-4 right-4 bg-[var(--bg-surface)] border border-[var(--border)]
                   rounded-full p-2 shadow-lg text-[var(--text-muted)] hover:text-[var(--text)]
                   hover:bg-[var(--bg-hover)] transition-colors z-10"
        >
            <ChevronDown size={16} />
        </button>
    {/if}
</div>
```

- [ ] **Step 6: Correr todos los tests**

```bash
cd services/sda-frontend && npm run test
```

Esperado: PASS — todos los tests incluyendo los 6 de `isNearBottom`.

- [ ] **Step 7: Commit**

```bash
git add services/sda-frontend/src/lib/utils/scroll.ts \
        services/sda-frontend/src/lib/utils/scroll.test.ts \
        services/sda-frontend/src/lib/components/chat/MessageList.svelte
git commit -m "feat(chat): add isNearBottom utility and MessageList with smart auto-scroll"
```

---

## Task 8: Refactorizar `[id]/+page.svelte` en coordinador

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/chat/[id]/+page.svelte`

Esta tarea reemplaza el page.svelte actual (212 líneas) con un coordinador que usa los 5 componentes. El `+page.server.ts` no cambia.

- [ ] **Step 1: Correr tests antes de modificar**

```bash
cd services/sda-frontend && npm run test
```

Esperado: PASS — confirmar baseline antes del refactor.

- [ ] **Step 2: Reemplazar `src/routes/(app)/chat/[id]/+page.svelte`**

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { ChatStore } from '$lib/stores/chat.svelte';
    import HistoryPanel from '$lib/components/chat/HistoryPanel.svelte';
    import MessageList from '$lib/components/chat/MessageList.svelte';
    import SourcesPanel from '$lib/components/chat/SourcesPanel.svelte';
    import ChatInput from '$lib/components/chat/ChatInput.svelte';

    let { data } = $props();
    const chat = new ChatStore();

    let selectedCollection = $state(data.session.collection ?? '');
    let sourcesOpen = $state(false);

    onMount(() => {
        chat.collection = data.session.collection ?? '';
        chat.crossdoc = data.session.crossdoc ?? false;
        if (data.session.messages?.length) {
            chat.loadMessages(data.session.messages);
        }
    });

    async function sendMessage(query: string) {
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
                signal: chat.abortController!.signal,
            });

            if (!resp.ok) {
                chat.appendToken('\n[Error: no se pudo conectar con el servidor]');
                return;
            }

            const reader = resp.body!.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                buffer += decoder.decode(value, { stream: true });
                const lines = buffer.split('\n');
                buffer = lines.pop() ?? '';
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const dataStr = line.slice(6).trim();
                        if (dataStr === '[DONE]') continue;
                        try {
                            const event = JSON.parse(dataStr);
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
                        } catch { /* ignorar errores de parse en chunks parciales */ }
                    }
                }
            }
        } catch (err: any) {
            if (err?.name !== 'AbortError') {
                chat.appendToken('\n[Error: conexión interrumpida]');
            }
        } finally {
            chat.finalizeStream();
        }
    }
</script>

<div class="flex h-screen overflow-hidden">
    <!-- Panel izquierdo: historial con búsqueda -->
    <HistoryPanel
        sessions={data.history}
        currentId={data.session.id}
    />

    <!-- Centro: header + mensajes + input -->
    <div class="flex-1 flex flex-col border-r border-[var(--border)] min-w-0">

        <!-- Header: colección + crossdoc + toggle fuentes -->
        <div class="flex items-center gap-2 px-3 py-2 border-b border-[var(--border)] text-xs flex-shrink-0">
            <select
                bind:value={selectedCollection}
                class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-sm)]
                       px-2 py-0.5 text-[var(--text)] outline-none focus:border-[var(--accent)]"
            >
                {#each data.collections as col}
                    <option value={col}>{col}</option>
                {/each}
            </select>

            <label class="flex items-center gap-1 text-[var(--text-muted)] cursor-pointer select-none">
                <input type="checkbox" bind:checked={chat.crossdoc} class="accent-[var(--accent)]" />
                Crossdoc
            </label>

            <button
                onclick={() => sourcesOpen = !sourcesOpen}
                disabled={chat.sources.length === 0}
                class="ml-auto px-2 py-1 rounded-[var(--radius-sm)] text-xs transition-colors
                       disabled:opacity-40
                       {sourcesOpen && chat.sources.length > 0
                           ? 'bg-[var(--accent)] text-white'
                           : 'bg-[var(--bg-surface)] border border-[var(--border)] text-[var(--text-muted)] hover:text-[var(--text)]'}"
            >
                Fuentes ({chat.sources.length})
            </button>
        </div>

        <!-- Lista de mensajes con auto-scroll -->
        <MessageList
            messages={chat.messages}
            streaming={chat.streaming}
            streamingContent={chat.streamingContent}
        />

        <!-- Input con stop button -->
        <ChatInput
            streaming={chat.streaming}
            onsubmit={sendMessage}
            onstop={() => chat.stopStream()}
        />
    </div>

    <!-- Panel derecho: fuentes (toggleable) -->
    <SourcesPanel
        sources={chat.sources}
        open={sourcesOpen}
    />
</div>
```

- [ ] **Step 3: Correr todos los tests**

```bash
cd services/sda-frontend && npm run test
```

Esperado: PASS — todos los tests anteriores siguen en PASS.

Output esperado al final:
```
Test Files  X passed
Tests       X passed
```

- [ ] **Step 4: Verificar que el dev server arranca sin errores de TypeScript**

```bash
cd services/sda-frontend && npm run build 2>&1 | head -50
```

Esperado: build exitoso o solo warnings, no errores de tipo.

- [ ] **Step 5: Commit final**

```bash
git add services/sda-frontend/src/routes/(app)/chat/[id]/+page.svelte
git commit -m "feat(chat): refactor chat page into component coordinator (Fase 2)"
```

---

## Verificación final

```bash
# Todos los tests
cd services/sda-frontend && npm run test

# Build limpio
npm run build
```

Tests esperados al completar:
- `ChatStore` — 4 tests (abortController, stopStream, contenido parcial, stopStream sin startStream)
- `parseMarkdown` — 7 tests (negrita, cursiva, h1, lista, código, tipo string, vacío)
- `filterSessions` — 6 tests (vacío, espacios, case-insensitive, mayúsculas, sin match, parcial)
- `isNearBottom` — 6 tests (fondo, threshold, arriba, tope, custom threshold, contenido corto)
- Tests previos — todos en PASS

Total esperado: ~23 tests nuevos + todos los tests previos.

## Fases siguientes
- `2026-03-19-fase3-upload-pro.md` — DropZone, queue, colecciones CRUD
- `2026-03-19-fase4-admin-pro.md` — Usuarios, áreas, permisos, RAG config
- `2026-03-19-fase5-polish.md` — Command palette, animaciones, responsive
