# SDA Frontend — Fase 1: Fundación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reemplazar el design system roto con design tokens, componentes UI base, sidebar rediseñada, error boundaries que no crashean la app, y dark/light mode toggle — dejando el frontend estable y consistente como base para todas las fases siguientes.

**Architecture:** CSS variables duales (dark/light) en `app.css`, componentes UI en `src/lib/components/ui/`, sidebar refactorizada en `src/lib/components/layout/`, toast system con store Svelte 5 runes, `<svelte:boundary>` en el root layout.

**Tech Stack:** SvelteKit 5, Svelte 5 runes ($state/$derived/$props), Tailwind CSS 4, Vitest, mode-watcher (ya instalado), lucide-svelte (ya instalado)

**Working directory:** `services/sda-frontend/`

**Run dev server:** `npm run dev` (desde `services/sda-frontend/`)

**Run tests:** `npm run test` (desde `services/sda-frontend/`)

---

## Mapa de archivos

### Crear

```
src/
├── lib/
│   ├── components/
│   │   ├── ui/                          ← NUEVO directorio
│   │   │   ├── Button.svelte            ← variantes: primary/secondary/danger/ghost
│   │   │   ├── Input.svelte             ← label, error, icon slot
│   │   │   ├── Badge.svelte             ← variantes semánticas de color
│   │   │   ├── Card.svelte              ← surface container
│   │   │   ├── Modal.svelte             ← backdrop, Esc, focus trap
│   │   │   ├── Skeleton.svelte          ← shimmer loading
│   │   │   ├── Toast.svelte             ← toast individual
│   │   │   └── ToastContainer.svelte    ← outlet que muestra todos los toasts
│   │   └── layout/                      ← NUEVO directorio
│   │       └── Sidebar.svelte           ← reemplaza sidebar/Sidebar.svelte
│   └── stores/
│       └── toast.svelte.ts              ← NUEVO: ToastStore con $state
└── routes/
    └── (app)/
        └── chat/
            └── +page.svelte             ← NUEVO: empty state del chat
```

### Modificar

```
src/app.css                              ← reescribir con design tokens dark+light
src/routes/+layout.svelte                ← agregar ModeWatcher
src/routes/+error.svelte                 ← actualizar estilos con CSS vars
src/routes/(app)/+layout.svelte          ← agregar svelte:boundary + ToastContainer + nueva Sidebar
src/routes/(app)/+error.svelte           ← actualizar estilos con CSS vars
src/lib/components/sidebar/Sidebar.svelte     ← eliminar (reemplazado)
src/lib/components/sidebar/SidebarItem.svelte ← eliminar (absorbido por nueva Sidebar)
src/routes/(app)/chat/[id]/+page.svelte  ← fix font sizes
src/routes/(app)/collections/+page.svelte     ← fix font sizes
src/routes/(app)/admin/users/+page.svelte     ← fix font sizes
src/routes/(app)/audit/+page.svelte      ← fix font sizes
src/routes/(app)/settings/+page.svelte   ← fix font sizes
src/routes/(app)/chat/[id]/+page.server.ts    ← agregar try/catch
src/routes/(app)/collections/+page.server.ts  ← agregar try/catch
src/routes/(app)/admin/users/+page.server.ts  ← agregar try/catch
src/routes/(app)/audit/+page.server.ts   ← agregar try/catch
```

---

## Task 1: Design tokens — reescribir app.css

**Files:**
- Modify: `src/app.css`

Esta es la base de todo. Sin los tokens, ningún componente usa los colores correctos.

**Paleta Saldivia Warm Adaptive:**
- Dark default: fondos charcoal cálido (`#181510`, `#1e1a14`)
- Light: fondos crema (`#faf8f4`, `#f2ede3`)
- Acento único: azul Saldivia `#2093d3`
- Tipografía: Roboto, base 14px

- [ ] **Reemplazar `src/app.css` completo:**

```css
@import "tailwindcss";
@import url('https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;600;700&display=swap');

/* ===== DARK MODE (default) ===== */
:root {
    --bg-base:       #181510;
    --bg-surface:    #1e1a14;
    --bg-hover:      #2a2418;
    --bg-accent-dim: #143248;
    --border:        #2e2820;
    --border-focus:  #3a3025;
    --text:          #ede8e0;
    --text-muted:    #8c7b6a;
    --text-faint:    #4a4035;
    --accent:        #2093d3;
    --accent-light:  #4db8f0;
    --accent-hover:  #1a7ab5;
    --success:       #4ade80;
    --success-bg:    #0d2b1e;
    --warning:       #fbbf24;
    --warning-bg:    #2a1800;
    --danger:        #f87171;
    --danger-bg:     #2d1010;
    --info:          #60a5fa;
    --info-bg:       #0a1628;
    --radius-sm:     4px;
    --radius-md:     8px;
    --radius-lg:     12px;
    --sidebar-width: 220px;
    --sidebar-collapsed: 56px;
    --topbar-height: 44px;
}

/* ===== LIGHT MODE ===== */
:root[data-theme="light"],
.light {
    --bg-base:       #faf8f4;
    --bg-surface:    #f2ede3;
    --bg-hover:      #e6dfd4;
    --bg-accent-dim: #ddedf7;
    --border:        #ddd5c7;
    --border-focus:  #c8bfb3;
    --text:          #1a2030;
    --text-muted:    #8c7b6a;
    --text-faint:    #b0a090;
    --accent:        #2093d3;
    --accent-light:  #1a7ab5;
    --accent-hover:  #166fa8;
    --success:       #16a34a;
    --success-bg:    #dcfce7;
    --warning:       #d97706;
    --warning-bg:    #fef3c7;
    --danger:        #dc2626;
    --danger-bg:     #fee2e2;
    --info:          #2563eb;
    --info-bg:       #dbeafe;
}

/* ===== BASE ===== */
*, *::before, *::after { box-sizing: border-box; }

html { font-size: 14px; }

body {
    font-family: 'Roboto', system-ui, sans-serif;
    background-color: var(--bg-base);
    color: var(--text);
    line-height: 1.6;
    -webkit-font-smoothing: antialiased;
}

/* Scrollbar */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: var(--bg-base); }
::-webkit-scrollbar-thumb { background: var(--border-focus); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--text-faint); }

/* Focus ring accesible */
:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
}

/* Transición de tema */
*, *::before, *::after {
    transition: background-color 0.2s ease, border-color 0.2s ease, color 0.1s ease;
}
```

- [ ] **Verificar que el dev server compila sin errores:**

```bash
cd services/sda-frontend && npm run dev
```
Esperado: sin errores en terminal. La app se ve con fondo `#181510` (dark por default).

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/app.css
git commit -m "feat(frontend): design tokens Saldivia Warm Adaptive (dark/light CSS vars)"
```

---

## Task 2: Toast store

**Files:**
- Create: `src/lib/stores/toast.svelte.ts`
- Test: `src/lib/stores/toast.svelte.test.ts`

Lógica pura — fácil de testear con Vitest antes de construir el componente visual.

- [ ] **Crear el test primero** (`src/lib/stores/toast.svelte.test.ts`):

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock Svelte runes environment
vi.mock('svelte', () => ({ untrack: (fn: () => any) => fn() }));

describe('ToastStore', () => {
    it('adds a success toast', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.success('Operación exitosa');
        expect(toastStore.toasts).toHaveLength(1);
        expect(toastStore.toasts[0].type).toBe('success');
        expect(toastStore.toasts[0].message).toBe('Operación exitosa');
    });

    it('adds an error toast', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.error('Error de conexión');
        const last = toastStore.toasts.at(-1)!;
        expect(last.type).toBe('error');
    });

    it('removes a toast by id', async () => {
        const { toastStore } = await import('./toast.svelte.js');
        toastStore.success('test');
        const id = toastStore.toasts.at(-1)!.id;
        toastStore.dismiss(id);
        expect(toastStore.toasts.find(t => t.id === id)).toBeUndefined();
    });
});
```

- [ ] **Ejecutar el test — debe fallar:**

```bash
cd services/sda-frontend && npm run test -- --reporter=verbose toast
```
Esperado: `FAIL — Cannot find module './toast.svelte.js'`

- [ ] **Crear `src/lib/stores/toast.svelte.ts`:**

```typescript
export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
    id: string;
    type: ToastType;
    message: string;
    duration: number;
}

class ToastStore {
    toasts = $state<Toast[]>([]);

    private add(type: ToastType, message: string, duration: number) {
        const id = crypto.randomUUID();
        this.toasts.push({ id, type, message, duration });
        if (duration > 0) {
            setTimeout(() => this.dismiss(id), duration);
        }
    }

    success(message: string, duration = 4000) { this.add('success', message, duration); }
    error(message: string, duration = 6000)   { this.add('error',   message, duration); }
    info(message: string, duration = 4000)    { this.add('info',    message, duration); }
    warning(message: string, duration = 5000) { this.add('warning', message, duration); }

    dismiss(id: string) {
        this.toasts = this.toasts.filter(t => t.id !== id);
    }
}

export const toastStore = new ToastStore();
```

- [ ] **Ejecutar tests — deben pasar:**

```bash
cd services/sda-frontend && npm run test -- --reporter=verbose toast
```
Esperado: `PASS — 3 tests`

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/lib/stores/toast.svelte.ts \
        services/sda-frontend/src/lib/stores/toast.svelte.test.ts
git commit -m "feat(frontend): ToastStore con Svelte 5 runes + tests"
```

---

## Task 3: Componentes base — Button, Badge, Card, Skeleton

**Files:**
- Create: `src/lib/components/ui/Button.svelte`
- Create: `src/lib/components/ui/Badge.svelte`
- Create: `src/lib/components/ui/Card.svelte`
- Create: `src/lib/components/ui/Skeleton.svelte`

Componentes stateless simples. No requieren tests unitarios (sin lógica), verificar visualmente en dev server.

- [ ] **Crear `src/lib/components/ui/Button.svelte`:**

```svelte
<script lang="ts">
    interface Props {
        variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
        size?: 'sm' | 'md' | 'lg';
        disabled?: boolean;
        loading?: boolean;
        type?: 'button' | 'submit' | 'reset';
        onclick?: () => void;
        children: any;
    }

    let {
        variant = 'primary',
        size = 'md',
        disabled = false,
        loading = false,
        type = 'button',
        onclick,
        children,
    }: Props = $props();

    const base = 'inline-flex items-center justify-center gap-2 font-medium rounded-[var(--radius-md)] transition-all duration-150 focus-visible:outline-2 focus-visible:outline-[var(--accent)] disabled:opacity-50 disabled:cursor-not-allowed';

    const variants = {
        primary:   'bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)] active:scale-[0.98]',
        secondary: 'bg-[var(--bg-surface)] border border-[var(--border)] text-[var(--text-muted)] hover:border-[var(--accent)] hover:text-[var(--text)]',
        danger:    'bg-[var(--danger-bg)] border border-[var(--danger)] text-[var(--danger)] hover:bg-[var(--danger)] hover:text-white',
        ghost:     'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]',
    };

    const sizes = {
        sm: 'text-xs px-2.5 py-1.5 h-7',
        md: 'text-sm px-3.5 py-2 h-9',
        lg: 'text-sm px-5 py-2.5 h-11',
    };
</script>

<button
    {type}
    {onclick}
    disabled={disabled || loading}
    class="{base} {variants[variant]} {sizes[size]}"
>
    {#if loading}
        <span class="w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full animate-spin"></span>
    {/if}
    {@render children()}
</button>
```

- [ ] **Crear `src/lib/components/ui/Badge.svelte`:**

```svelte
<script lang="ts">
    interface Props {
        variant?: 'blue' | 'green' | 'red' | 'yellow' | 'gray' | 'orange';
        children: any;
    }
    let { variant = 'gray', children }: Props = $props();

    const variants = {
        blue:   'bg-[var(--info-bg)] text-[var(--info)]',
        green:  'bg-[var(--success-bg)] text-[var(--success)]',
        red:    'bg-[var(--danger-bg)] text-[var(--danger)]',
        yellow: 'bg-[var(--warning-bg)] text-[var(--warning)]',
        gray:   'bg-[var(--bg-hover)] text-[var(--text-muted)]',
        orange: 'bg-orange-950/40 text-orange-400',
    };
</script>

<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold {variants[variant]}">
    {@render children()}
</span>
```

- [ ] **Crear `src/lib/components/ui/Card.svelte`:**

```svelte
<script lang="ts">
    interface Props {
        padding?: boolean;
        class?: string;
        children: any;
    }
    let { padding = true, class: cls = '', children }: Props = $props();
</script>

<div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] {padding ? 'p-4' : ''} {cls}">
    {@render children()}
</div>
```

- [ ] **Crear `src/lib/components/ui/Skeleton.svelte`:**

```svelte
<script lang="ts">
    interface Props {
        width?: string;
        height?: string;
        rounded?: string;
        class?: string;
    }
    let { width = '100%', height = '1rem', rounded = 'var(--radius-sm)', class: cls = '' }: Props = $props();
</script>

<div
    style="width:{width}; height:{height}; border-radius:{rounded}"
    class="bg-[var(--bg-hover)] animate-pulse {cls}"
    style:background="linear-gradient(90deg, var(--bg-hover) 25%, var(--border-focus) 50%, var(--bg-hover) 75%)"
    style:background-size="200% 100%"
></div>

<style>
    div { animation: shimmer 1.5s infinite; }
    @keyframes shimmer {
        0%   { background-position: 200% 0; }
        100% { background-position: -200% 0; }
    }
</style>
```

- [ ] **Verificar compilación en dev server** (sin errores TypeScript):

```bash
cd services/sda-frontend && npm run dev
```

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/lib/components/ui/
git commit -m "feat(frontend): Button, Badge, Card, Skeleton UI components"
```

---

## Task 4: Input component

**Files:**
- Create: `src/lib/components/ui/Input.svelte`

- [ ] **Crear `src/lib/components/ui/Input.svelte`:**

```svelte
<script lang="ts">
    interface Props {
        label?: string;
        error?: string;
        type?: string;
        placeholder?: string;
        value?: string;
        name?: string;
        id?: string;
        required?: boolean;
        disabled?: boolean;
        class?: string;
    }

    let {
        label,
        error,
        type = 'text',
        placeholder,
        value = $bindable(''),
        name,
        id,
        required = false,
        disabled = false,
        class: cls = '',
    }: Props = $props();

    const inputId = id ?? name ?? crypto.randomUUID();
</script>

<div class="flex flex-col gap-1 {cls}">
    {#if label}
        <label
            for={inputId}
            class="text-xs font-medium text-[var(--text-muted)]"
        >
            {label}{#if required}<span class="text-[var(--danger)] ml-0.5">*</span>{/if}
        </label>
    {/if}

    <input
        {type}
        {name}
        {placeholder}
        {required}
        {disabled}
        id={inputId}
        bind:value
        class="
            w-full px-3 py-2 text-sm rounded-[var(--radius-md)]
            bg-[var(--bg-surface)] border text-[var(--text)]
            placeholder:text-[var(--text-faint)]
            focus:outline-none focus:ring-2 focus:ring-[var(--accent)] focus:border-[var(--accent)]
            disabled:opacity-50 disabled:cursor-not-allowed
            transition-colors
            {error ? 'border-[var(--danger)]' : 'border-[var(--border)]'}
        "
    />

    {#if error}
        <p class="text-xs text-[var(--danger)]">{error}</p>
    {/if}
</div>
```

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/lib/components/ui/Input.svelte
git commit -m "feat(frontend): Input component con label, error y estados"
```

---

## Task 5: Modal component

**Files:**
- Create: `src/lib/components/ui/Modal.svelte`

El Modal necesita: backdrop click para cerrar, Esc para cerrar, y focus trap básico.

- [ ] **Crear `src/lib/components/ui/Modal.svelte`:**

```svelte
<script lang="ts">
    import { onMount } from 'svelte';

    interface Props {
        open?: boolean;
        title?: string;
        onclose?: () => void;
        size?: 'sm' | 'md' | 'lg';
        children: any;
        footer?: any;
    }

    let {
        open = $bindable(false),
        title,
        onclose,
        size = 'md',
        children,
        footer,
    }: Props = $props();

    const sizes = { sm: 'max-w-sm', md: 'max-w-md', lg: 'max-w-2xl' };

    function close() {
        open = false;
        onclose?.();
    }

    function onKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape') close();
    }

    function onBackdropClick(e: MouseEvent) {
        if (e.target === e.currentTarget) close();
    }
</script>

<svelte:window onkeydown={open ? onKeydown : undefined} />

{#if open}
    <!-- Backdrop -->
    <div
        class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
        role="dialog"
        aria-modal="true"
        aria-label={title}
        onclick={onBackdropClick}
    >
        <!-- Panel -->
        <div class="
            bg-[var(--bg-surface)] border border-[var(--border)]
            rounded-[var(--radius-lg)] shadow-2xl w-full {sizes[size]}
            animate-in fade-in zoom-in-95 duration-150
        ">
            {#if title}
                <div class="flex items-center justify-between px-5 py-4 border-b border-[var(--border)]">
                    <h2 class="text-sm font-semibold text-[var(--text)]">{title}</h2>
                    <button
                        onclick={close}
                        class="text-[var(--text-faint)] hover:text-[var(--text)] transition-colors p-1 rounded"
                        aria-label="Cerrar"
                    >✕</button>
                </div>
            {/if}

            <div class="px-5 py-4">
                {@render children()}
            </div>

            {#if footer}
                <div class="px-5 py-4 border-t border-[var(--border)] flex justify-end gap-2">
                    {@render footer()}
                </div>
            {/if}
        </div>
    </div>
{/if}
```

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/lib/components/ui/Modal.svelte
git commit -m "feat(frontend): Modal con backdrop, Esc close y accesibilidad"
```

---

## Task 6: Toast component + ToastContainer

**Files:**
- Create: `src/lib/components/ui/Toast.svelte`
- Create: `src/lib/components/ui/ToastContainer.svelte`

- [ ] **Crear `src/lib/components/ui/Toast.svelte`:**

```svelte
<script lang="ts">
    import { toastStore, type Toast } from '$lib/stores/toast.svelte';

    interface Props { toast: Toast; }
    let { toast }: Props = $props();

    const styles = {
        success: 'bg-[var(--success-bg)] border-[var(--success)] text-[var(--success)]',
        error:   'bg-[var(--danger-bg)] border-[var(--danger)] text-[var(--danger)]',
        warning: 'bg-[var(--warning-bg)] border-[var(--warning)] text-[var(--warning)]',
        info:    'bg-[var(--info-bg)] border-[var(--info)] text-[var(--info)]',
    };

    const icons = { success: '✓', error: '✕', warning: '⚠', info: 'ℹ' };
</script>

<div class="
    flex items-start gap-3 px-4 py-3 rounded-[var(--radius-md)]
    border shadow-lg min-w-72 max-w-sm text-sm
    {styles[toast.type]}
">
    <span class="font-bold text-base leading-none mt-0.5">{icons[toast.type]}</span>
    <p class="flex-1 leading-snug">{toast.message}</p>
    <button
        onclick={() => toastStore.dismiss(toast.id)}
        class="opacity-60 hover:opacity-100 transition-opacity leading-none mt-0.5"
        aria-label="Cerrar notificación"
    >✕</button>
</div>
```

- [ ] **Crear `src/lib/components/ui/ToastContainer.svelte`:**

```svelte
<script lang="ts">
    import { toastStore } from '$lib/stores/toast.svelte';
    import Toast from './Toast.svelte';
</script>

<!-- Portal de toasts — esquina inferior derecha -->
<div
    class="fixed bottom-5 right-5 z-[100] flex flex-col gap-2 pointer-events-none"
    aria-live="polite"
    aria-atomic="false"
>
    {#each toastStore.toasts as toast (toast.id)}
        <div class="pointer-events-auto">
            <Toast {toast} />
        </div>
    {/each}
</div>
```

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/lib/components/ui/Toast.svelte \
        services/sda-frontend/src/lib/components/ui/ToastContainer.svelte
git commit -m "feat(frontend): Toast y ToastContainer components"
```

---

## Task 7: Sidebar rediseñada

**Files:**
- Create: `src/lib/components/layout/Sidebar.svelte`

La sidebar actual tiene 40px y labels solo en tooltip. La nueva: 220px expandida, 56px colapsada, labels visibles, secciones agrupadas, avatar + theme toggle en el footer.

- [ ] **Crear `src/lib/components/layout/Sidebar.svelte`:**

```svelte
<script lang="ts">
    import { page } from '$app/stores';
    import { toggleMode } from 'mode-watcher';
    import {
        MessageSquare, BookOpen, Upload, Users,
        Building2, Shield, ClipboardList, Settings,
        LogOut, ChevronLeft, ChevronRight, LayoutDashboard,
        Sun, Moon
    } from 'lucide-svelte';

    interface Props {
        role: string;
        userName: string;
        userEmail: string;
    }

    let { role, userName, userEmail }: Props = $props();

    let collapsed = $state(false);
    let loggingOut = $state(false);

    let isAdmin    = $derived(role === 'admin');
    let isManager  = $derived(role === 'admin' || role === 'area_manager');

    type NavItem = {
        href: string;
        label: string;
        icon: any;
        adminOnly?: boolean;
        managerOnly?: boolean;
    };

    const mainNav: NavItem[] = [
        { href: '/chat',        label: 'Chat',        icon: MessageSquare },
        { href: '/collections', label: 'Colecciones', icon: BookOpen },
        { href: '/collections', label: 'Documentos',  icon: Upload },
    ];

    const adminNav: NavItem[] = [
        { href: '/admin/users',       label: 'Usuarios',     icon: Users,         managerOnly: true },
        { href: '/admin/areas',       label: 'Áreas',        icon: Building2,     managerOnly: true },
        { href: '/admin/permissions', label: 'Permisos',     icon: Shield,        adminOnly: true },
        { href: '/admin/rag-config',  label: 'RAG Config',   icon: Settings,      adminOnly: true },
        { href: '/admin/system',      label: 'Sistema',      icon: LayoutDashboard, adminOnly: true },
        { href: '/audit',             label: 'Auditoría',    icon: ClipboardList, adminOnly: true },
    ];

    function isActive(href: string) {
        return $page.url.pathname.startsWith(href);
    }

    async function handleLogout() {
        loggingOut = true;
        try {
            await fetch('/api/auth/session', { method: 'DELETE' });
        } finally {
            window.location.href = '/login';
        }
    }
</script>

<nav
    style="width: {collapsed ? 'var(--sidebar-collapsed)' : 'var(--sidebar-width)'}"
    class="
        flex-shrink-0 h-screen
        bg-[#120f0c] border-r border-[var(--border)]
        flex flex-col
        transition-[width] duration-200 ease-in-out
        overflow-hidden
    "
>
    <!-- Header -->
    <div class="flex items-center gap-2.5 px-3 py-3.5 border-b border-[var(--border)] min-h-[var(--topbar-height)]">
        <div class="w-7 h-7 bg-[var(--accent)] rounded-lg flex items-center justify-center text-white font-bold text-xs flex-shrink-0">
            S
        </div>
        {#if !collapsed}
            <div class="overflow-hidden">
                <div class="text-sm font-bold text-[var(--text)] whitespace-nowrap">SDA</div>
                <div class="text-[10px] text-[var(--text-faint)] whitespace-nowrap">Saldivia Buses</div>
            </div>
        {/if}
        <button
            onclick={() => collapsed = !collapsed}
            class="ml-auto text-[var(--text-faint)] hover:text-[var(--text)] transition-colors flex-shrink-0 p-0.5"
            title={collapsed ? 'Expandir' : 'Colapsar'}
        >
            {#if collapsed}
                <ChevronRight size={14} />
            {:else}
                <ChevronLeft size={14} />
            {/if}
        </button>
    </div>

    <!-- Nav -->
    <div class="flex-1 overflow-y-auto py-2 px-2 flex flex-col gap-0.5">
        {#if !collapsed}
            <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider px-2 pt-2 pb-1">
                Principal
            </div>
        {/if}

        {#each mainNav as item}
            <a
                href={item.href}
                title={collapsed ? item.label : undefined}
                class="
                    flex items-center gap-2.5 px-2.5 py-2 rounded-[var(--radius-md)]
                    text-sm transition-colors group relative
                    {isActive(item.href)
                        ? 'bg-[var(--bg-surface)] text-[var(--text)] border-l-2 border-[var(--accent)] pl-[9px]'
                        : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]'}
                "
            >
                <item.icon size={16} class="flex-shrink-0" />
                {#if !collapsed}
                    <span class="whitespace-nowrap">{item.label}</span>
                {:else}
                    <span class="
                        absolute left-full ml-2 px-2 py-1 rounded-[var(--radius-sm)]
                        bg-[var(--bg-surface)] border border-[var(--border)]
                        text-xs text-[var(--text)] whitespace-nowrap
                        opacity-0 group-hover:opacity-100 pointer-events-none z-50
                        transition-opacity
                    ">{item.label}</span>
                {/if}
            </a>
        {/each}

        {#if isManager}
            {#if !collapsed}
                <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider px-2 pt-4 pb-1">
                    Administración
                </div>
            {:else}
                <div class="h-px bg-[var(--border)] mx-2 my-2"></div>
            {/if}

            {#each adminNav as item}
                {#if (item.adminOnly && isAdmin) || (item.managerOnly && isManager)}
                    <a
                        href={item.href}
                        title={collapsed ? item.label : undefined}
                        class="
                            flex items-center gap-2.5 px-2.5 py-2 rounded-[var(--radius-md)]
                            text-sm transition-colors group relative
                            {isActive(item.href)
                                ? 'bg-[var(--bg-surface)] text-[var(--text)] border-l-2 border-[var(--accent)] pl-[9px]'
                                : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]'}
                        "
                    >
                        <item.icon size={16} class="flex-shrink-0" />
                        {#if !collapsed}
                            <span class="whitespace-nowrap">{item.label}</span>
                        {:else}
                            <span class="
                                absolute left-full ml-2 px-2 py-1 rounded-[var(--radius-sm)]
                                bg-[var(--bg-surface)] border border-[var(--border)]
                                text-xs text-[var(--text)] whitespace-nowrap
                                opacity-0 group-hover:opacity-100 pointer-events-none z-50
                                transition-opacity
                            ">{item.label}</span>
                        {/if}
                    </a>
                {/if}
            {/each}
        {/if}

        {#if !collapsed}
            <div class="text-[9px] font-bold text-[var(--text-faint)] uppercase tracking-wider px-2 pt-4 pb-1">
                Cuenta
            </div>
        {:else}
            <div class="h-px bg-[var(--border)] mx-2 my-2"></div>
        {/if}

        <a href="/settings"
           title={collapsed ? 'Configuración' : undefined}
           class="
               flex items-center gap-2.5 px-2.5 py-2 rounded-[var(--radius-md)]
               text-sm transition-colors group relative
               {isActive('/settings')
                   ? 'bg-[var(--bg-surface)] text-[var(--text)]'
                   : 'text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]'}
           "
        >
            <Settings size={16} class="flex-shrink-0" />
            {#if !collapsed}
                <span>Configuración</span>
            {:else}
                <span class="absolute left-full ml-2 px-2 py-1 rounded-[var(--radius-sm)] bg-[var(--bg-surface)] border border-[var(--border)] text-xs text-[var(--text)] whitespace-nowrap opacity-0 group-hover:opacity-100 pointer-events-none z-50 transition-opacity">Configuración</span>
            {/if}
        </a>
    </div>

    <!-- Footer: avatar + theme toggle + logout -->
    <div class="border-t border-[var(--border)] p-2">
        <!-- Theme toggle -->
        <button
            onclick={toggleMode}
            title="Cambiar tema"
            class="
                flex items-center gap-2.5 w-full px-2.5 py-2 rounded-[var(--radius-md)]
                text-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text)]
                transition-colors text-sm mb-1
            "
        >
            <Sun size={16} class="flex-shrink-0 dark:hidden" />
            <Moon size={16} class="flex-shrink-0 hidden dark:block" />
            {#if !collapsed}
                <span>Cambiar tema</span>
            {/if}
        </button>

        <!-- User -->
        <div class="flex items-center gap-2.5 px-2.5 py-2">
            <div class="w-6 h-6 bg-[var(--accent)] rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                {userName.charAt(0).toUpperCase()}
            </div>
            {#if !collapsed}
                <div class="flex-1 overflow-hidden">
                    <div class="text-xs font-semibold text-[var(--text)] truncate">{userName}</div>
                    <div class="text-[10px] text-[var(--text-faint)] truncate">{userEmail}</div>
                </div>
                <button
                    onclick={handleLogout}
                    disabled={loggingOut}
                    title="Cerrar sesión"
                    class="text-[var(--text-faint)] hover:text-[var(--danger)] transition-colors disabled:opacity-50 flex-shrink-0"
                >
                    <LogOut size={14} />
                </button>
            {/if}
        </div>
    </div>
</nav>
```

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/lib/components/layout/Sidebar.svelte
git commit -m "feat(frontend): Sidebar rediseñada — 220px, collapsable, secciones, avatar"
```

---

## Task 8: Actualizar layout principal

**Files:**
- Modify: `src/routes/(app)/+layout.svelte`
- Modify: `src/routes/+layout.svelte`
- Modify: `src/routes/(app)/+layout.server.ts`

Conectar la nueva Sidebar, agregar ToastContainer, `<svelte:boundary>`, y ModeWatcher.

- [ ] **Leer `src/routes/(app)/+layout.server.ts`** para ver qué datos se pasan:

El loader ya devuelve `user` con `role` y `area_id`. Hay que agregar `name` y `email` al tipo retornado si no están. Verificar:

```bash
grep -n "user\|name\|email" services/sda-frontend/src/routes/\(app\)/+layout.server.ts
```

- [ ] **Actualizar `src/routes/+layout.svelte`** (root layout, agrega ModeWatcher):

```svelte
<script lang="ts">
    import { ModeWatcher } from 'mode-watcher';
    let { children } = $props();
</script>

<ModeWatcher defaultMode="dark" />
{@render children()}
```

- [ ] **Actualizar `src/routes/(app)/+layout.svelte`:**

```svelte
<script lang="ts">
    import Sidebar from '$lib/components/layout/Sidebar.svelte';
    import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
    import { toastStore } from '$lib/stores/toast.svelte';

    interface Props { data: any; children: any; }
    let { data, children }: Props = $props();

    function onBoundaryError(error: Error, reset: () => void) {
        toastStore.error(`Error inesperado: ${error.message}`);
        console.error('[app boundary]', error);
    }
</script>

<div class="flex h-screen overflow-hidden bg-[var(--bg-base)]">
    <Sidebar
        role={data.user.role}
        userName={data.user.name ?? data.user.email}
        userEmail={data.user.email}
    />

    <main class="flex-1 overflow-auto min-w-0">
        <svelte:boundary onerror={onBoundaryError}>
            {@render children()}
            {#snippet failed(error, reset)}
                <div class="flex items-center justify-center h-64 p-8">
                    <div class="text-center">
                        <div class="text-4xl mb-3">⚠️</div>
                        <p class="text-sm font-semibold text-[var(--text)] mb-1">Algo salió mal</p>
                        <p class="text-xs text-[var(--text-muted)] mb-4">{error.message}</p>
                        <button
                            onclick={reset}
                            class="text-xs text-[var(--accent)] hover:underline"
                        >Reintentar</button>
                    </div>
                </div>
            {/snippet}
        </svelte:boundary>
    </main>
</div>

<ToastContainer />
```

- [ ] **Verificar que no haya errores de TypeScript** (el campo `name` en user):

```bash
cd services/sda-frontend && npx svelte-check --tsconfig ./tsconfig.json 2>&1 | grep "Error\|error" | head -20
```

Si hay error por `data.user.name` undefined, editar `+layout.server.ts` para incluirlo en el return.

- [ ] **Confirmar en el browser** que la nueva sidebar se ve con 220px, el logo SDA, y las secciones.

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/routes/+layout.svelte \
        services/sda-frontend/src/routes/\(app\)/+layout.svelte
git commit -m "feat(frontend): layout con svelte:boundary, ToastContainer y nueva Sidebar"
```

---

## Task 9: Error pages con nuevo diseño

**Files:**
- Modify: `src/routes/+error.svelte`
- Modify: `src/routes/(app)/+error.svelte`

- [ ] **Reemplazar `src/routes/+error.svelte`:**

```svelte
<script lang="ts">
    import { page } from '$app/stores';
</script>

<div class="min-h-screen flex items-center justify-center bg-[var(--bg-base)] p-6">
    <div class="text-center max-w-md">
        <div class="text-7xl font-bold text-[var(--border-focus)] mb-4 tabular-nums">
            {$page.status}
        </div>
        <h1 class="text-lg font-semibold text-[var(--text)] mb-2">
            {$page.status === 404 ? 'Página no encontrada' : 'Error del servidor'}
        </h1>
        <p class="text-sm text-[var(--text-muted)] mb-8">
            {$page.error?.message ?? 'Ocurrió un error inesperado.'}
        </p>
        <div class="flex gap-3 justify-center">
            <a href="/chat" class="
                inline-flex items-center gap-2 px-4 py-2 rounded-[var(--radius-md)]
                bg-[var(--accent)] text-white text-sm font-medium
                hover:bg-[var(--accent-hover)] transition-colors
            ">
                Ir al Chat
            </a>
            <button onclick={() => window.location.reload()} class="
                inline-flex items-center gap-2 px-4 py-2 rounded-[var(--radius-md)]
                bg-[var(--bg-surface)] border border-[var(--border)]
                text-[var(--text-muted)] text-sm
                hover:text-[var(--text)] hover:border-[var(--accent)] transition-colors
            ">
                Reintentar
            </button>
        </div>
    </div>
</div>
```

- [ ] **Reemplazar `src/routes/(app)/+error.svelte`** con el mismo componente pero sin `min-h-screen` (ya está dentro del layout):

```svelte
<script lang="ts">
    import { page } from '$app/stores';
</script>

<div class="flex items-center justify-center flex-1 p-8 min-h-[400px]">
    <div class="text-center max-w-md">
        <div class="text-6xl font-bold text-[var(--border-focus)] mb-4">{$page.status}</div>
        <h1 class="text-base font-semibold text-[var(--text)] mb-2">
            {$page.status === 404 ? 'Página no encontrada' : 'Error del servidor'}
        </h1>
        <p class="text-sm text-[var(--text-muted)] mb-6">
            {$page.error?.message ?? 'Ocurrió un error inesperado.'}
        </p>
        <div class="flex gap-3 justify-center">
            <a href="/chat" class="
                px-4 py-2 rounded-[var(--radius-md)] bg-[var(--accent)] text-white text-sm font-medium
                hover:bg-[var(--accent-hover)] transition-colors
            ">Ir al Chat</a>
            <button onclick={() => window.location.reload()} class="
                px-4 py-2 rounded-[var(--radius-md)] bg-[var(--bg-surface)] border border-[var(--border)]
                text-[var(--text-muted)] text-sm hover:border-[var(--accent)] hover:text-[var(--text)] transition-colors
            ">Reintentar</button>
        </div>
    </div>
</div>
```

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/routes/+error.svelte \
        services/sda-frontend/src/routes/\(app\)/+error.svelte
git commit -m "feat(frontend): error pages con design tokens y estilos actualizados"
```

---

## Task 10: Chat empty state (`/chat` sin sesión)

**Files:**
- Create: `src/routes/(app)/chat/+page.svelte`

- [ ] **Verificar si existe `src/routes/(app)/chat/+page.svelte`:**

```bash
ls services/sda-frontend/src/routes/\(app\)/chat/
```

Si no existe (actualmente el `/chat` sin ID redirige via server), crearlo:

- [ ] **Crear `src/routes/(app)/chat/+page.svelte`:**

```svelte
<script lang="ts">
    let { data } = $props();

    const suggestions = [
        '¿Cuáles son las especificaciones del Aries 365?',
        'Normativas de homologación vigentes',
        '¿Qué documentos hay disponibles?',
    ];
</script>

<div class="flex-1 flex items-center justify-center h-full p-8 min-h-[500px]">
    <div class="text-center max-w-lg">
        <div class="w-14 h-14 bg-[var(--accent)] rounded-2xl flex items-center justify-center mx-auto mb-5">
            <span class="text-white font-bold text-xl">S</span>
        </div>

        <h1 class="text-xl font-semibold text-[var(--text)] mb-2">
            ¿En qué puedo ayudarte?
        </h1>
        <p class="text-sm text-[var(--text-muted)] mb-8">
            Consultá sobre los documentos de Saldivia Buses
        </p>

        <div class="flex flex-wrap gap-2 justify-center">
            {#each suggestions as suggestion}
                <a
                    href="/chat/new"
                    class="
                        px-3.5 py-2 rounded-full text-sm
                        bg-[var(--bg-surface)] border border-[var(--border)]
                        text-[var(--text-muted)] hover:text-[var(--text)] hover:border-[var(--accent)]
                        transition-colors
                    "
                >
                    {suggestion}
                </a>
            {/each}
        </div>
    </div>
</div>
```

> **Nota:** El href `/chat/new` necesita que el server loader de `[id]/+page.server.ts` maneje `id === 'new'` creando una sesión nueva. Si actualmente `/chat` redirige a `/chat/{id}` via el server, este componente puede quedarse como empty state visual mientras la redirección funciona.

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/routes/\(app\)/chat/+page.svelte
git commit -m "feat(frontend): empty state del chat con sugerencias"
```

---

## Task 11: Fix font sizes — sweep de todas las páginas

**Files:**
- Modify: `src/routes/(app)/chat/[id]/+page.svelte`
- Modify: `src/routes/(app)/collections/+page.svelte`
- Modify: `src/routes/(app)/collections/[name]/+page.svelte`
- Modify: `src/routes/(app)/admin/users/+page.svelte`
- Modify: `src/routes/(app)/audit/+page.svelte`
- Modify: `src/routes/(app)/settings/+page.svelte`
- Modify: `src/routes/(auth)/login/+page.svelte`

Reemplazar todos los colores hardcodeados y fuentes microscópicas por CSS variables. Las páginas existentes seguirán funcionando — este es un **visual sweep**, no una refactorización lógica.

- [ ] **Buscar todos los colores hardcodeados para tener dimensión del trabajo:**

```bash
grep -rn "text-\[#\|bg-\[#\|border-\[#" services/sda-frontend/src/routes/ | wc -l
```

- [ ] **Estrategia de reemplazo**: Para cada página, hacer un find-and-replace de la tabla:

| Viejo (hardcoded) | Nuevo (CSS var) |
|---|---|
| `text-[#e2e8f0]` / `text-[#cbd5e1]` | `text-[var(--text)]` |
| `text-[#94a3b8]` | `text-[var(--text-muted)]` |
| `text-[#475569]` / `text-[#64748b]` | `text-[var(--text-faint)]` |
| `bg-[#0f172a]` / `bg-[#0c1220]` / `bg-[#070d1a]` | `bg-[var(--bg-base)]` |
| `bg-[#1e293b]` | `bg-[var(--bg-surface)]` |
| `border-[#1e293b]` / `border-[#334155]` | `border-[var(--border)]` |
| `text-[#6366f1]` / `text-[#a5b4fc]` | `text-[var(--accent)]` |
| `bg-[#6366f1]` | `bg-[var(--accent)]` |
| `text-[7px]`, `text-[8px]`, `text-[9px]` | `text-xs` (12px) |
| `text-[11px]` | `text-sm` (14px) |

- [ ] **Aplicar en `chat/[id]/+page.svelte`** — la más compleja, hacerla primera:

El panel de historial izquierdo (w-40) pasa a ser un componente separado en Fase 2. Por ahora solo fix fonts/colors.

```bash
# Verificar antes:
grep -c "text-\[#\|bg-\[#\|text-\[7\|text-\[8\|text-\[9\|text-\[11" services/sda-frontend/src/routes/\(app\)/chat/\[id\]/+page.svelte
```

Aplicar los reemplazos, luego verificar en el dev server.

- [ ] **Aplicar en las demás páginas** (`collections`, `admin/users`, `audit`, `settings`, `login`).

- [ ] **Verificar en dev server** que todas las páginas se ven correctamente con los nuevos colores.

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/routes/
git commit -m "fix(frontend): reemplazar colores hardcodeados y fuentes microscópicas por CSS vars"
```

---

## Task 12: Try/catch en server loaders

**Files:**
- Modify: `src/routes/(app)/chat/[id]/+page.server.ts`
- Modify: `src/routes/(app)/collections/+page.server.ts`
- Modify: `src/routes/(app)/collections/[name]/+page.server.ts`
- Modify: `src/routes/(app)/admin/users/+page.server.ts`
- Modify: `src/routes/(app)/audit/+page.server.ts`
- Modify: `src/routes/(app)/settings/+page.server.ts`

- [ ] **Ver los loaders actuales para entender el patrón:**

```bash
grep -n "export.*load\|async function load" services/sda-frontend/src/routes/\(app\)/collections/+page.server.ts
```

- [ ] **Patrón a aplicar en cada loader** — wrappear la llamada al gateway con try/catch que devuelve fallback en lugar de crashear:

```typescript
// ANTES (crashea si el gateway está caído):
export const load = async ({ locals }) => {
    const collections = await gateway.getCollections(locals.user!);
    return { collections };
};

// DESPUÉS (maneja errores gracefully):
export const load = async ({ locals }) => {
    try {
        const collections = await gateway.getCollections(locals.user!);
        return { collections };
    } catch (err) {
        console.error('[collections loader]', err);
        // Devolver estado vacío + flag de error para mostrar en UI
        return { collections: [], error: 'No se pudo cargar las colecciones' };
    }
};
```

- [ ] **Aplicar el patrón en los 6 loaders.** El hook `hooks.server.ts` ya maneja el 401 redirigiendo al login — no es necesario manejarlo aquí.

- [ ] **Test manual**: Detener el gateway y verificar que las páginas muestran estado vacío en lugar de página de error 500.

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/routes/
git commit -m "fix(frontend): try/catch en todos los page.server.ts loaders"
```

---

## Task 13: Login page con nuevo diseño

**Files:**
- Modify: `src/routes/(auth)/login/+page.svelte`

Actualizar la página de login para usar los nuevos componentes UI y CSS variables.

- [ ] **Reemplazar `src/routes/(auth)/login/+page.svelte`:**

```svelte
<script lang="ts">
    import { enhance } from '$app/forms';
    import Input from '$lib/components/ui/Input.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    let { form } = $props();
    let submitting = $state(false);
</script>

<div class="min-h-screen flex items-center justify-center bg-[var(--bg-base)] p-6">
    <div class="w-full max-w-sm">
        <!-- Logo -->
        <div class="text-center mb-8">
            <div class="inline-flex items-center justify-center w-12 h-12 bg-[var(--accent)] rounded-2xl mb-3">
                <span class="text-white font-bold text-lg">S</span>
            </div>
            <h1 class="text-lg font-bold text-[var(--text)]">SDA</h1>
            <p class="text-sm text-[var(--text-muted)] mt-1">Sistema de Documentación Asistida</p>
            <p class="text-xs text-[var(--text-faint)] mt-0.5">Saldivia Buses</p>
        </div>

        <!-- Form -->
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-6">
            {#if form?.error}
                <div class="bg-[var(--danger-bg)] border border-[var(--danger)] rounded-[var(--radius-md)] px-3 py-2 text-sm text-[var(--danger)] mb-4">
                    {form.error}
                </div>
            {/if}

            <form method="POST" use:enhance={() => {
                submitting = true;
                return async ({ update }) => { submitting = false; await update(); };
            }} class="flex flex-col gap-4">
                <Input
                    name="email"
                    type="email"
                    label="Email"
                    placeholder="usuario@saldivia.com.ar"
                    required
                    disabled={submitting}
                />
                <Input
                    name="password"
                    type="password"
                    label="Contraseña"
                    required
                    disabled={submitting}
                />
                <Button type="submit" loading={submitting} class="w-full mt-2">
                    Ingresar
                </Button>
            </form>
        </div>
    </div>
</div>
```

- [ ] **Verificar en el dev server** que el login se ve correcto y funciona.

- [ ] **Commit:**

```bash
git add services/sda-frontend/src/routes/\(auth\)/login/+page.svelte
git commit -m "feat(frontend): login page rediseñada con componentes UI base"
```

---

## Verificación final de Fase 1

- [ ] **Correr todos los tests:**

```bash
cd services/sda-frontend && npm run test
```
Esperado: `PASS` — mínimo los tests del toast store.

- [ ] **Verificar TypeScript sin errores:**

```bash
cd services/sda-frontend && npx svelte-check --tsconfig ./tsconfig.json
```
Esperado: 0 errores (puede haber warnings menores).

- [ ] **Build de producción:**

```bash
cd services/sda-frontend && npm run build
```
Esperado: build exitoso sin errores.

- [ ] **Checklist manual en el browser:**
  - [ ] Fondo `#181510` (dark) en todas las páginas
  - [ ] Sidebar de 220px con labels visibles y secciones
  - [ ] Toggle colapsa sidebar a 56px
  - [ ] Dark/light toggle funciona sin recargar
  - [ ] Fuentes legibles (mínimo 12px en toda la app)
  - [ ] Login page con nuevo diseño
  - [ ] Desconectar el gateway: páginas muestran estado vacío, no crash
  - [ ] Error 404: página de error con diseño nuevo

- [ ] **Commit final:**

```bash
git add -A
git commit -m "feat(frontend): Fase 1 completa — design system, sidebar, error boundaries, dark/light"
```

---

## Fases siguientes

Una vez deployada y verificada la Fase 1, los planes de las fases siguientes se escriben en orden:

- `2026-03-18-fase2-chat-pro.md` — Markdown, auto-scroll, historial mejorado
- `2026-03-18-fase3-upload-pro.md` — DropZone, queue, colecciones CRUD
- `2026-03-18-fase4-admin-pro-max.md` — Usuarios edit, áreas, permisos, RAG config
- `2026-03-18-fase5-polish.md` — Command palette, animaciones, responsive
