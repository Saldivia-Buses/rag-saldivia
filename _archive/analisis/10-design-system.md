# 10 — Design System "Warm Intelligence"

## Filosofia

Inspirado en la UI de Claude (claude.ai) con acento azure blue. Tonos calidos, nunca negro frio. La idea es que el sistema se sienta "inteligente pero humano".

**Nombre:** Warm Intelligence
**Inspiracion:** claude.ai tokens + azure accent
**Idioma UI:** Espanol

---

## Tokens CSS

### Light mode (default)

```css
--bg:            #faf9f5    /* crema calido — fondo base */
--surface:       #f0eee8    /* elevacion 1 — cards, paneles */
--surface-2:     #e5e3dc    /* elevacion 2 — hover, active */
--border:        #e0ddd6    /* bordes sutiles */
--fg:            #141413    /* texto principal — casi negro */
--fg-muted:      #4a4a47    /* texto secundario */
--fg-subtle:     #6e6c69    /* texto terciario, hints */
--accent:        #2563eb    /* azure blue — CTA, links, focus */
--accent-subtle: #dbeafe    /* background de badges accent */
--success:       #16a34a    /* operaciones exitosas */
--destructive:   #dc2626    /* errores, delete */
--warning:       #d97706    /* advertencias */
```

### Dark mode (`.dark` class)

```css
/* Warm dark — nunca negro frio */
--bg:            #1a1812
--surface:       #24221a
--accent:        #4a9fd4    /* navy mas claro para contraste */
/* ... resto de tokens invertidos */
```

### Uso en Tailwind v4

```html
<!-- CORRECTO: usar utility classes con tokens -->
<div class="bg-bg text-fg border-border">
<div class="bg-surface text-fg-muted">
<button class="bg-accent text-white">

<!-- INCORRECTO: nunca hardcodear colores -->
<div class="bg-[#faf9f5]">  <!-- NO -->
<div style="color: #1a1a1a"> <!-- NO -->
```

**Regla:** `bg-surface` para cards/paneles elevados, `bg-bg` para fondo base.

---

## Tipografia

- **Font:** Instrument Sans via `next/font/google`
- **Variable CSS:** `--font-instrument-sans`
- **Letter-spacing:** `-0.01em` en body
- **Self-hosting automatico** por Next.js (no depende de Google Fonts CDN)

---

## Tailwind v4 especifico

- **`@theme inline`** en globals.css — critico para dark mode class-based en runtime
- **`postcss.config.js`** con `@tailwindcss/postcss` — sin el, utility classes custom no se generan
- Las clases custom (`bg-surface`, `text-fg-muted`) solo se generan si aparecen en archivos escaneados

---

## Componentes UI (shadcn/ui + custom)

### Button — 6 variantes

| Variante | Uso |
|----------|-----|
| `default` | CTA primario (bg-accent) |
| `destructive` | Acciones peligrosas (bg-destructive) |
| `outline` | Secundario con borde |
| `secondary` | Terciario |
| `ghost` | Sin fondo, solo hover |
| `link` | Parece un link |

Sizes: `default`, `sm`, `lg`, `icon`

### Badge — 6 variantes

| Variante | Uso |
|----------|-----|
| `default` | General |
| `secondary` | Informativo |
| `outline` | Sutil |
| `destructive` | Error/peligro |
| `success` | Exito |
| `warning` | Advertencia |

### DataTable

Tabla avanzada con @tanstack/react-table:
- Sorting por columna (click en header)
- Filtrado por texto
- Paginacion configurable
- Seleccion de filas (checkbox)

### StatCard

Tarjeta de estadistica:
- Valor numerico grande
- Delta con flecha (positivo verde, negativo rojo)
- Icono y titulo descriptivo
- Usado en AdminDashboard

### EmptyPlaceholder

Estado vacio con:
- Icono (Lucide)
- Titulo descriptivo
- Descripcion/instruccion
- Accion opcional (boton)

### Skeleton

Loading states con shimmer:
- `Skeleton` — bloque generico
- `SkeletonText` — lineas de texto
- `SkeletonTable` — tabla completa con headers

---

## Densidad adaptiva

```html
<!-- Admin/tablas: mas informacion visible -->
<div data-density="compact">...</div>

<!-- Chat/collections: mas espacio visual -->
<div data-density="spacious">...</div>
```

---

## Storybook

**Version:** 8
**Framework:** react-vite
**Puerto dev:** `:6006`

### Estructura de stories

```
apps/web/stories/
  design-system/     → tokens, tipografia, colores
  primitivos/        → button, badge, input, skeleton, etc.
  features/          → chat, collections, admin
  layout/            → AppShell, NavRail
```

### Addons activos
- **addon-a11y** — WCAG por componente en tiempo real
- **addon-themes** — toggle light/dark en la toolbar
- **addon-essentials** — controls, actions, viewport, docs

### Comandos
```bash
bun run storybook         # dev en :6006
bun run build:storybook   # build estatico
```

---

## Iconos

**Libreria:** Lucide React
**Uso:** import con tree-shaking

```tsx
import { MessageSquare, Settings, Shield } from "lucide-react"
```

---

## Dark mode

**Implementacion:** next-themes con class strategy
**Toggle:** `ThemeToggle` component
**Estado:** Pendiente refinamiento (Plan 20)

El dark mode funciona via clase `.dark` en `<html>`. Los tokens CSS se redefinen dentro de `.dark { ... }` en globals.css.

---

## Paginas y sus densidades

| Pagina | Densidad | Layout |
|--------|---------|--------|
| `/login` | spacious | Sin NavRail, centrado |
| `/chat` | spacious | AppShell + sidebar sesiones |
| `/chat/[id]` | spacious | AppShell + chat + sources panel |
| `/collections` | default | AppShell + tabla |
| `/settings` | spacious | AppShell + formularios |
| `/admin/*` | compact | AppShell + AdminLayout + tablas |
| `/messaging` | default | AppShell + channel list + messages |
