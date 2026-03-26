# Design System "Warm Intelligence"

> Plan 7 completado 2026-03-26

---

## Filosofía

RAG Saldivia usa un design system propio llamado **"Warm Intelligence"**:

- **Crema cálido** como fondo base — nunca blanco frío
- **Navy profundo** como acento — confianza, rigor, claridad
- **Instrument Sans** — tipografía limpia y moderna
- **Sin decoración innecesaria** — cada elemento justifica su presencia
- **Dark mode cálido** — marrón oscuro (`#1a1812`), nunca negro

---

## Tokens CSS

Definidos en `apps/web/src/app/globals.css`.

### Light mode (`:root`)

```css
/* Fondos */
--bg:            #faf8f4   /* fondo base — crema cálido */
--surface:       #f0ebe0   /* cards, paneles elevados */
--surface-2:     #e8e1d4   /* hover, segundo nivel */

/* Bordes */
--border:        #ede9e0
--border-strong: #d5cfc4

/* Texto */
--fg:            #1a1a1a   /* texto principal */
--fg-muted:      #5a5048   /* texto secundario */
--fg-subtle:     #9a9088   /* placeholders, metadata */

/* Acento navy */
--accent:        #1a5276
--accent-hover:  #154360
--accent-dark:   #0d3349
--accent-subtle: #d4e8f7   /* fondos de badges, highlights */
--accent-fg:     #ffffff   /* texto sobre acento */

/* Estados semánticos */
--destructive:        #c0392b
--destructive-subtle: #fde8e8
--success:            #2d6a4f
--success-subtle:     #d4f1e4
--warning:            #d68910
--warning-subtle:     #fef9e7
```

### Dark mode (`.dark`)

next-themes aplica la clase `.dark` al elemento `<html>` cuando el usuario activa dark mode.

```css
--bg:          #1a1812   /* warm dark — nunca negro frío */
--surface:     #24221a
--accent:      #4a9fd4   /* navy más claro para contraste en dark */
--fg:          #f0ebe0   /* crema claro */
```

### Utility classes de Tailwind v4

Los tokens se exponen como utilities via `@theme inline` en globals.css:

```html
<!-- Fondos -->
<div class="bg-bg">...</div>
<div class="bg-surface">...</div>
<div class="bg-surface-2">...</div>

<!-- Texto -->
<p class="text-fg">Principal</p>
<p class="text-fg-muted">Secundario</p>
<p class="text-fg-subtle">Metadata</p>

<!-- Acento -->
<div class="bg-accent text-accent-fg">Botón primario</div>
<div class="bg-accent-subtle text-accent">Badge activo</div>

<!-- Estados -->
<p class="text-destructive">Error</p>
<div class="bg-destructive-subtle text-destructive">Panel de error</div>
<p class="text-success">Éxito</p>
<p class="text-warning">Advertencia</p>
```

---

## Tipografía

**Instrument Sans** via `next/font/google`.

```typescript
// apps/web/src/app/layout.tsx
const instrumentSans = Instrument_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  style: ["normal", "italic"],
  variable: "--font-instrument-sans",
  display: "swap",
})
```

Escala tipográfica (CSS vars):

```css
--text-xs:   0.75rem
--text-sm:   0.875rem
--text-base: 1rem
--text-lg:   1.125rem
--text-xl:   1.25rem
--text-2xl:  1.5rem
--text-3xl:  1.875rem
--text-4xl:  2.25rem

/* Letter-spacing ajustado para Instrument Sans */
--tracking-normal: -0.01em
```

---

## Densidad adaptiva

El sistema tiene dos niveles de densidad controlados por `data-density`:

```html
<!-- Admin / tablas densas -->
<div data-density="compact">
  <!-- row-height: 2rem, input-height: 2rem, avatar-size: 1.5rem -->
</div>

<!-- Chat / páginas de contenido -->
<div data-density="spacious">
  <!-- row-height: 3rem, input-height: 2.5rem, avatar-size: 2.25rem -->
</div>
```

**Dónde se aplica:**
- `(app)/layout.tsx` → `data-density="spacious"` por defecto
- `(app)/admin/layout.tsx` → `data-density="compact"` para todas las rutas admin

Los componentes usan estas CSS vars automáticamente — no hay lógica duplicada.

---

## Componentes UI

Todos en `apps/web/src/components/ui/`.

### Button

```tsx
<Button>Default</Button>
<Button variant="destructive">Eliminar</Button>
<Button variant="outline">Cancelar</Button>
<Button variant="secondary">Secundario</Button>
<Button variant="ghost">Ghost</Button>
<Button variant="link">Enlace</Button>

<Button size="sm">Pequeño (h-8)</Button>
<Button size="default">Normal (h-9)</Button>
<Button size="lg">Grande (h-10)</Button>
<Button size="icon"><Trash2 /></Button>

<Button disabled>Deshabilitado</Button>
<Button asChild><a href="/link">Como link</a></Button>
```

### Badge

```tsx
<Badge>Default (navy)</Badge>
<Badge variant="secondary">Secundario</Badge>
<Badge variant="outline">Outline</Badge>
<Badge variant="destructive">Error</Badge>
<Badge variant="success">Activo ✓</Badge>
<Badge variant="warning">Pendiente</Badge>
```

### DataTable

```tsx
import { DataTable } from "@/components/ui/data-table"
import type { ColumnDef } from "@tanstack/react-table"

const columns: ColumnDef<User>[] = [
  { accessorKey: "email", header: "Email" },
  { accessorKey: "role",  header: "Rol" },
]

<DataTable
  columns={columns}
  data={users}
  searchKey="email"
  searchPlaceholder="Buscar por email..."
  pageSize={10}
/>
```

### StatCard

```tsx
<StatCard
  label="Queries (30d)"
  value={1284}
  delta={18}
  deltaLabel="vs mes anterior"
  icon={MessageSquare}
/>
```

### EmptyPlaceholder

```tsx
<EmptyPlaceholder>
  <EmptyPlaceholder.Icon icon={MessageSquare} />
  <EmptyPlaceholder.Title>Sin conversaciones</EmptyPlaceholder.Title>
  <EmptyPlaceholder.Description>
    Hacé una pregunta para empezar.
  </EmptyPlaceholder.Description>
  <Button>Nueva sesión</Button>
</EmptyPlaceholder>
```

### Skeleton (estados de carga)

```tsx
<Skeleton className="h-4 w-32" />        // genérico
<SkeletonText lines={3} />                // párrafo
<SkeletonAvatar size="sm|md|lg" />        // avatar
<SkeletonCard />                          // tarjeta
<SkeletonTable rows={5} cols={4} />       // tabla
```

---

## Dark mode

next-themes con `attribute="class"` (clase `.dark` en `<html>`):

```tsx
// Para activar/desactivar desde código:
const { setTheme } = useTheme()
setTheme("dark")   // activa .dark
setTheme("light")  // quita .dark
setTheme("system") // sigue preferencia del sistema
```

**En tests de Playwright** (visual regression), el dark mode se activa vía JavaScript porque el `colorScheme: 'dark'` de Playwright NO activa el class-based dark mode:

```typescript
// helpers.ts
await page.evaluate(() => {
  document.documentElement.classList.add("dark")
  localStorage.setItem("theme", "dark")
})
```

---

## Storybook

Catálogo visual de todos los componentes.

```bash
bun run storybook        # dev en :6006
bun run build:storybook  # build estático
```

**Estructura de stories** (`apps/web/stories/`):

```
design-system/
  tokens.stories.tsx    — paleta completa + escala tipográfica
primitivos/
  button.stories.tsx    — 6 variantes + 4 tamaños
  badge.stories.tsx     — 6 variantes
  input.stories.tsx     — estados
  avatar.stories.tsx    — fallback + imagen
  table.stories.tsx     — con datos mock
  skeleton.stories.tsx  — todas las variantes
features/
  stat-card.stories.tsx
  empty-placeholder.stories.tsx
```

**Addons activos:**
- `addon-a11y` — auditoría WCAG por componente (ver panel "Accessibility")
- `addon-themes` — toggle light/dark en el canvas
- `addon-essentials` — controls, docs, actions, etc.

---

## Agregar un nuevo componente al design system

1. Crear el componente en `apps/web/src/components/ui/<nombre>.tsx`
2. Usar tokens CSS (`bg-surface`, `text-fg-muted`, etc.) — nunca hardcodear colores
3. Crear `apps/web/stories/primitivos/<nombre>.stories.tsx` o `features/` según corresponda
4. Crear `apps/web/src/components/ui/__tests__/<nombre>.test.tsx` con:
   - `afterEach(cleanup)` al inicio
   - Tests de render, variantes, interacción, accesibilidad
5. Correr `bun run test:components` — debe pasar
6. Abrir Storybook y verificar que el panel a11y no tiene violaciones

---

## Fuentes de código comunitario usadas

| Fuente | Uso |
|---|---|
| shadcn/ui | Base de primitivos (button, badge, input, etc.) |
| @tanstack/react-table | Motor de DataTable |
| origin-ui.com | Referencia para formularios avanzados |
| shadcnblocks.com | Referencia para bloques de página |
| 21st.dev | Referencia para microinteracciones |
| Recharts | Gráficos en AnalyticsDashboard |

**Regla:** el código de comunidad siempre se copia y adapta — nunca se instala como dependencia externa sin evaluar.
