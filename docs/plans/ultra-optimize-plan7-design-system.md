# Plan 7: Design System "Warm Intelligence"

> Este documento vive en `docs/plans/ultra-optimize-plan7-design-system.md`.
> Se actualiza a medida que se completan las tareas. Cada fase completada genera entrada en CHANGELOG.md.
> Spec completo: `docs/superpowers/specs/2026-03-26-design-system-design.md`

---

## Contexto

Los Planes 1–6 construyeron stack, tests, bugfixes y UI testing infrastructure. La app tiene 62 componentes y 24 páginas funcionales pero sin sistema de diseño real: tokens inconsistentes, tipografía genérica, cero identidad visual.

**Lo que construye este plan:** identidad visual completa de RAG Saldivia. Un sistema de tokens CSS coherente, tipografía Instrument Sans, paleta crema + navy, dark mode cálido, densidad adaptiva, componentes comunitarios curados, momentos especiales animados, y Storybook como catálogo vivo.

**Decisiones de diseño tomadas (no re-abrir):**
- Estética: Warm Intelligence (crema cálido, tranquilidad, nitidez)
- Acento: Navy/azul profundo (`#1a5276`)
- Tipografía: Instrument Sans (Google Fonts via `next/font`)
- Dark mode: sí, warm dark (`#1a1812` — nunca negro frío)
- Densidad: adaptiva (`compact` en admin, `spacious` en chat/colecciones)
- Código comunidad: shadcn base + origin-ui + shadcnblocks + 21st.dev + Aceternity + MagicUI + Radiant shaders
- Catálogo: Storybook 8 (base del Plan 6 para visual regression)

**Nota de orden de ejecución:** La Fase 1 del Plan 6 (react-scan baseline) se corre ANTES de empezar la Fase 1 de este plan para capturar el estado previo.

---

## Seguimiento

Formato: `- [ ] Descripción — estimación`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`
Cada fase completada → entrada en CHANGELOG.md → commit.

---

## Fase 1 — Tokens CSS *(2-3 hs)*

Objetivo: reescribir `apps/web/src/app/globals.css` con la paleta completa light + dark. Todos los componentes existentes deben seguir funcionando al terminar (aliases shadcn).

**Archivos a modificar:**
- `apps/web/src/app/globals.css` — reescritura completa

### F1.1 — Preparación: generar escala Tailwind

- [ ] Abrir [uicolors.app](https://uicolors.app/create) con el valor `#1a5276`
- [ ] Generar la escala completa de navy (50–950) y copiar los valores
- [ ] Abrir [uicolors.app](https://uicolors.app/create) con el valor `#faf8f4`
- [ ] Generar la escala completa de cream y copiar los valores
- [ ] Documentar ambas escalas en un comentario al inicio de `globals.css`

### F1.2 — Tokens semánticos light mode

Reemplazar el bloque `:root` actual con:

```css
:root {
  /* ── Fondos ── */
  --bg:          #faf8f4;
  --surface:     #f0ebe0;
  --surface-2:   #e8e1d4;

  /* ── Bordes ── */
  --border:      #ede9e0;
  --border-strong: #d5cfc4;

  /* ── Texto ── */
  --fg:          #1a1a1a;
  --fg-muted:    #5a5048;
  --fg-subtle:   #9a9088;

  /* ── Acento navy ── */
  --accent:         #1a5276;
  --accent-hover:   #154360;
  --accent-dark:    #0d3349;
  --accent-subtle:  #d4e8f7;
  --accent-fg:      #ffffff;

  /* ── Estados ── */
  --destructive:        #c0392b;
  --destructive-subtle: #fde8e8;
  --success:            #2d6a4f;
  --success-subtle:     #d4f1e4;
  --warning:            #d68910;
  --warning-subtle:     #fef9e7;

  /* ── Sidebar / nav (legado — apuntan a nuevos tokens) ── */
  --sidebar-bg: var(--surface);
  --nav-bg:     var(--surface);

  /* ── Aliases shadcn (no romper componentes existentes) ── */
  --background:          var(--bg);
  --foreground:          var(--fg);
  --card:                var(--surface);
  --card-foreground:     var(--fg);
  --popover:             var(--surface);
  --popover-foreground:  var(--fg);
  --primary:             var(--accent);
  --primary-foreground:  var(--accent-fg);
  --secondary:           var(--surface-2);
  --secondary-foreground: var(--fg-muted);
  --muted:               var(--surface);
  --muted-foreground:    var(--fg-muted);
  --accent-shadcn:       var(--surface-2);
  --accent-foreground:   var(--fg);
  --destructive-shadcn:  var(--destructive);
  --destructive-foreground: #ffffff;
  --border:              #ede9e0;
  --input:               var(--border);
  --ring:                var(--accent);
  --radius:              0.5rem;
}
```

- [ ] Reemplazar bloque `:root` en `globals.css`
- [ ] `bun run dev` — verificar que la app carga sin errores de CSS

### F1.3 — Tokens dark mode

```css
.dark {
  /* ── Fondos ── */
  --bg:          #1a1812;
  --surface:     #24221a;
  --surface-2:   #2e2b21;

  /* ── Bordes ── */
  --border:      #35312a;
  --border-strong: #46423a;

  /* ── Texto ── */
  --fg:          #f0ebe0;
  --fg-muted:    #b0a898;
  --fg-subtle:   #7a7268;

  /* ── Acento navy (más claro para contraste en dark) ── */
  --accent:         #4a9fd4;
  --accent-hover:   #5aafde;
  --accent-dark:    #6bbfe8;
  --accent-subtle:  #1a3a52;
  --accent-fg:      #0d1a24;

  /* ── Estados dark ── */
  --destructive:        #e57373;
  --destructive-subtle: #2a1515;
  --success:            #66bb6a;
  --success-subtle:     #0d2a14;
  --warning:            #ffa726;
  --warning-subtle:     #2a1e00;
}
```

- [ ] Agregar bloque `.dark` después de `:root`
- [ ] Togglear dark mode en el browser — verificar que los colores cambian correctamente

### F1.4 — @theme inline para Tailwind v4

```css
@theme inline {
  --color-bg:             var(--bg);
  --color-surface:        var(--surface);
  --color-surface-2:      var(--surface-2);
  --color-border:         var(--border);
  --color-border-strong:  var(--border-strong);
  --color-fg:             var(--fg);
  --color-fg-muted:       var(--fg-muted);
  --color-fg-subtle:      var(--fg-subtle);
  --color-accent:         var(--accent);
  --color-accent-hover:   var(--accent-hover);
  --color-accent-subtle:  var(--accent-subtle);
  --color-accent-fg:      var(--accent-fg);
  --color-destructive:    var(--destructive);
  --color-success:        var(--success);
  --color-warning:        var(--warning);
}
```

- [ ] Agregar bloque `@theme inline` al final de las declaraciones de variables
- [ ] Verificar: `<div className="bg-surface text-fg-muted">` funciona como clase Tailwind

### F1.5 — Verificación WCAG

- [ ] Verificar ratio fg/bg con [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/):
  - `--fg` (`#1a1a1a`) sobre `--bg` (`#faf8f4`) → debe ser ≥ 4.5:1 ✓
  - `--fg-muted` (`#5a5048`) sobre `--bg` → debe ser ≥ 4.5:1
  - `--accent-fg` (`#ffffff`) sobre `--accent` (`#1a5276`) → debe ser ≥ 4.5:1
  - `--fg` (`#f0ebe0`) sobre dark `--bg` (`#1a1812`) → debe ser ≥ 4.5:1
- [ ] Ajustar valores que no pasen hasta que cumplan WCAG AA

### Criterio de done
`bun run dev` corre, la app se ve en crema/navy, dark mode funciona, `bun run type-check` sin errores.

### Checklist de cierre
- [ ] `bun run test` — todos pasan (79+ tests)
- [ ] CHANGELOG.md actualizado bajo `### Plan 7 — Design System`
- [ ] `git commit -m "style(tokens): sistema de tokens CSS crema-navy con dark mode cálido"`

---

## Fase 2 — Tipografía *(1 hs)*

Objetivo: Instrument Sans cargando via `next/font/google`, escala tipográfica definida como tokens CSS, aplicada al layout raíz.

**Archivos a modificar:**
- `apps/web/src/app/layout.tsx`
- `apps/web/src/app/globals.css`

### F2.1 — Instalar Instrument Sans via next/font

```typescript
// apps/web/src/app/layout.tsx
import { Instrument_Sans } from 'next/font/google'

const sans = Instrument_Sans({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  style: ['normal', 'italic'],
  variable: '--font-sans',
  display: 'swap',
})
```

- [ ] Importar `Instrument_Sans` en `layout.tsx`
- [ ] Agregar `variable: '--font-sans'` a la config
- [ ] Aplicar `className={sans.variable}` al elemento `<html>` (junto con las clases existentes)

### F2.2 — Escala tipográfica en globals.css

```css
:root {
  /* Tipografía */
  --font-sans: 'Instrument Sans', system-ui, sans-serif;

  /* Escala */
  --text-xs:   0.75rem;
  --text-sm:   0.875rem;
  --text-base: 1rem;
  --text-lg:   1.125rem;
  --text-xl:   1.25rem;
  --text-2xl:  1.5rem;
  --text-3xl:  1.875rem;
  --text-4xl:  2.25rem;

  /* Line heights */
  --leading-tight:  1.25;
  --leading-snug:   1.375;
  --leading-normal: 1.5;
  --leading-relaxed: 1.625;

  /* Letter spacing */
  --tracking-tight:  -0.025em;
  --tracking-normal: -0.01em;   /* Instrument Sans se ve mejor ajustado */
  --tracking-wide:   0.025em;
}
```

- [ ] Agregar variables de tipografía al bloque `:root`
- [ ] Actualizar `body` en globals.css:
  ```css
  body {
    font-family: var(--font-sans);
    letter-spacing: var(--tracking-normal);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  ```

### F2.3 — Verificación

- [ ] `bun run dev` — Instrument Sans visible en DevTools → Network → Fonts
- [ ] Verificar que el font-weight 600 (semibold) se usa en los headings del NavRail
- [ ] `bun run build` — sin errores (next/font optimiza en build time)

### Criterio de done
Instrument Sans carga en el browser, la app se ve con la tipografía nueva.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "style(typography): instrument sans via next/font con escala tipográfica completa"`

---

## Fase 3 — Primitivos shadcn *(3-4 hs)*

Objetivo: reestilizar los 12 componentes `ui/` con los nuevos tokens. Sin cambiar la API ni la lógica — solo CSS y variantes.

**Archivos a modificar:** todos los archivos en `apps/web/src/components/ui/`

> **Regla:** no cambiar props, exports, ni lógica interna. Solo `className`, `cva` variants, y `cn()` calls.

### F3.1 — button.tsx

Variantes a actualizar:
- `default`: `bg-accent text-accent-fg hover:bg-accent-hover`
- `destructive`: `bg-destructive text-white hover:opacity-90`
- `outline`: `border-border bg-bg text-fg hover:bg-surface`
- `secondary`: `bg-surface-2 text-fg hover:bg-surface`
- `ghost`: `text-fg hover:bg-surface`
- `link`: `text-accent underline-offset-4`

Sizes:
- `sm`: `h-8 px-3 text-xs`
- `default`: `h-9 px-4 text-sm`
- `lg`: `h-10 px-6 text-base`
- `icon`: `h-9 w-9`

- [ ] Actualizar `cva` variants en `button.tsx`
- [ ] Verificar: los botones actuales en `/admin/users` se ven con el nuevo estilo

### F3.2 — input.tsx y textarea.tsx

```css
/* Estilo base para input */
border border-border bg-bg px-3 py-2 text-sm text-fg
placeholder:text-fg-subtle
focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-accent
disabled:cursor-not-allowed disabled:opacity-50
rounded-md
```

- [ ] Actualizar `input.tsx`
- [ ] Actualizar `textarea.tsx` (mismo patrón, agrega `resize-y`)
- [ ] Verificar: el form de login usa los nuevos estilos

### F3.3 — badge.tsx

Variantes:
- `default`: `bg-accent text-accent-fg`
- `secondary`: `bg-surface-2 text-fg-muted`
- `destructive`: `bg-destructive-subtle text-destructive`
- `outline`: `border border-border text-fg`
- `success`: `bg-success-subtle text-success` (nueva variante)
- `warning`: `bg-warning-subtle text-warning` (nueva variante)

- [ ] Actualizar `badge.tsx` con las 6 variantes
- [ ] Verificar badges en `/admin/users` (roles de usuario)

### F3.4 — avatar.tsx

```css
/* AvatarFallback */
bg-accent-subtle text-accent text-sm font-medium
```

- [ ] Actualizar `AvatarFallback` con colores navy/crema
- [ ] `AvatarImage`: agregar `object-cover`

### F3.5 — dialog.tsx y sheet.tsx

```css
/* DialogContent / SheetContent */
bg-bg border-border shadow-lg

/* DialogHeader */
border-b border-border pb-4 mb-4

/* DialogTitle */
text-lg font-semibold text-fg
```

- [ ] Actualizar `dialog.tsx`
- [ ] Actualizar `sheet.tsx`

### F3.6 — table.tsx

Estilos compact-first (la tabla es el componente más denso del sistema):
```css
/* Table */
text-sm text-fg w-full

/* TableHeader */
border-b border-border bg-surface

/* TableHead */
text-xs font-semibold text-fg-muted uppercase tracking-wide px-3 py-2 text-left

/* TableRow */
border-b border-border hover:bg-surface transition-colors

/* TableCell */
px-3 py-2 text-sm
```

- [ ] Actualizar `table.tsx` con estilos dense y hover states

### F3.7 — tooltip.tsx

```css
/* TooltipContent */
bg-fg text-bg text-xs px-2 py-1 rounded shadow-sm
```

- [ ] Actualizar `tooltip.tsx`

### F3.8 — separator.tsx

```css
bg-border
```

- [ ] Actualizar `separator.tsx`

### F3.9 — command.tsx (CommandPalette base)

```css
/* CommandInput */
border-b border-border bg-bg text-fg placeholder:text-fg-subtle

/* CommandItem */
text-sm text-fg hover:bg-surface cursor-pointer px-2 py-1.5 rounded

/* CommandItem[aria-selected] */
bg-accent-subtle text-accent

/* CommandGroup heading */
text-xs font-semibold text-fg-subtle uppercase tracking-wide px-2 py-1
```

- [ ] Actualizar `command.tsx`

### F3.10 — sonner.tsx (Toasts)

- [ ] Actualizar el `Toaster` con `theme` prop que usa el tema activo de next-themes:
  ```typescript
  import { useTheme } from 'next-themes'
  const { theme } = useTheme()
  return <Toaster theme={theme as 'light' | 'dark'} richColors />
  ```
- [ ] Verificar toast de éxito, error, y warning con los nuevos colores

### F3.11 — theme-toggle.tsx

- [ ] Verificar que el toggle usa `next-themes` con `attribute="class"` (no `data-theme`)
- [ ] Iconos: sol para light, luna para dark — usando `lucide-react`
- [ ] Transición suave: `transition-colors duration-200`

### Criterio de done
Todos los primitivos se ven con crema + navy + Instrument Sans. Dark mode funciona en todos.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] Navegar `/admin/users`, `/chat`, `/login` y verificar visualmente
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "style(primitives): reestilizar shadcn primitivos con tokens crema-navy"`

---

## Fase 4 — Layout system *(2-3 hs)*

Objetivo: AppShell, NavRail, y layouts de página con los nuevos tokens. Sistema de densidad adaptiva funcionando.

**Archivos a modificar:**
- `apps/web/src/app/globals.css` — tokens de densidad
- `apps/web/src/components/layout/AppShell.tsx`
- `apps/web/src/components/layout/AppShellChrome.tsx`
- `apps/web/src/components/layout/NavRail.tsx`
- `apps/web/src/components/layout/SecondaryPanel.tsx`
- `apps/web/src/components/layout/CommandPalette.tsx`
- `apps/web/src/app/(app)/layout.tsx`

### F4.1 — Tokens de densidad en globals.css

```css
/* Densidad adaptiva — aplicar data-density="compact" o "spacious" al contenedor */
[data-density="compact"] {
  --row-height:    2rem;        /* 32px */
  --cell-px:       0.5rem;
  --cell-py:       0.375rem;
  --avatar-size:   1.5rem;      /* 24px */
  --icon-size:     0.875rem;    /* 14px */
  --input-height:  2rem;        /* 32px */
  --gap-items:     0.25rem;
}

[data-density="spacious"] {
  --row-height:    3rem;        /* 48px */
  --cell-px:       0.75rem;
  --cell-py:       0.75rem;
  --avatar-size:   2.25rem;     /* 36px */
  --icon-size:     1.125rem;    /* 18px */
  --input-height:  2.5rem;      /* 40px */
  --gap-items:     0.5rem;
}
```

- [ ] Agregar los dos bloques de densidad en `globals.css`

### F4.2 — AppShell y AppShellChrome

- [ ] Aplicar `bg-bg` al fondo principal
- [ ] La columna de sidebar (si existe): `bg-surface border-r border-border`
- [ ] Header/topbar: `bg-surface border-b border-border h-12`
- [ ] Área de contenido principal: `flex-1 overflow-auto`

### F4.3 — NavRail

La NavRail es el elemento visual más importante del layout. Rediseño completo:

```
┌──────────────────┐
│  Logo RAG        │  ← 40px logo + nombre
│  ─────────────── │
│  🗨 Chat         │  ← items con hover state
│  📚 Colecciones  │
│  📁 Proyectos    │
│  ────────────── │
│  (admin)         │
│  👥 Usuarios    │
│  ─────────────── │
│  (bottom)        │
│  ⚙ Settings     │
│  ☀/🌙 Theme    │  ← dark mode toggle
│  [Avatar]        │  ← usuario activo
└──────────────────┘
```

- [ ] Fondo: `bg-surface border-r border-border`
- [ ] Logo: `text-fg font-semibold text-sm` + ícono navy
- [ ] Nav items: `flex items-center gap-2 px-3 py-2 rounded-md text-sm`
  - Inactivo: `text-fg-muted hover:bg-surface-2 hover:text-fg`
  - Activo: `bg-accent-subtle text-accent font-medium`
- [ ] Separadores entre secciones: `<Separator className="my-2" />`
- [ ] Dark mode toggle: al fondo, con `Sun`/`Moon` de lucide
- [ ] Avatar del usuario: `AvatarFallback` con iniciales, tooltip con nombre completo

### F4.4 — SecondaryPanel

- [ ] `bg-surface border-r border-border` (o `border-l` según posición)
- [ ] Heading del panel: `text-xs font-semibold text-fg-subtle uppercase tracking-wide px-4 py-3 border-b border-border`

### F4.5 — CommandPalette

- [ ] Container: `bg-bg border border-border rounded-lg shadow-xl`
- [ ] Width: `max-w-xl w-full`
- [ ] Animación de apertura: `data-[state=open]:animate-in data-[state=closed]:animate-out`

### F4.6 — Densidad en layouts

- [ ] `apps/web/src/app/(app)/layout.tsx`: aplicar `data-density="spacious"` al contenedor raíz del app group
- [ ] `apps/web/src/components/admin/`: aplicar `data-density="compact"` al wrapper de cada componente admin
- [ ] `apps/web/src/components/chat/ChatInterface.tsx`: `data-density="spacious"`

### F4.7 — Breakpoints responsive

- [ ] Agregar a `globals.css`:
  ```css
  /* Breakpoints del sistema (documenta los que usa Tailwind) */
  /* sm:640px  md:768px  lg:1024px  xl:1280px  2xl:1536px */
  ```
- [ ] En mobile (<768px): NavRail oculta, bottom bar de 5 tabs
- [ ] En tablet (768-1023px): NavRail icon-only (sin texto)
- [ ] En desktop (1024px+): NavRail completa con texto

### Criterio de done
El layout completo se ve en crema/navy, la NavRail está rediseñada, densidad adaptiva funciona.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] Verificar layout en 3 breakpoints (mobile, tablet, desktop) con DevTools
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "style(layout): appshell y navrail con tokens crema-navy y densidad adaptiva"`

---

## Fase 5 — Componentes comunidad *(3-4 hs)*

Objetivo: incorporar los mejores componentes de la comunidad, copiados y adaptados con nuestros tokens. Sin instalar nuevas dependencias de npm.

**Archivos a crear:**
- `apps/web/src/components/ui/data-table.tsx` — tabla avanzada
- `apps/web/src/components/ui/stat-card.tsx` — tarjeta de estadísticas
- `apps/web/src/components/ui/empty-placeholder.tsx` — estados vacíos (pre-fase 6)

### F5.1 — Tabla avanzada (de shadcnblocks.com)

La tabla avanzada reemplaza el uso directo de `<Table>` en componentes admin. Features necesarios:
- Sorting por columna
- Búsqueda/filtro por texto
- Paginación
- Selección de filas (checkbox)
- Columna de acciones

- [ ] Ir a [shadcnblocks.com](https://shadcnblocks.com) → buscar "data table"
- [ ] Copiar el componente más completo que use `@tanstack/react-table`
- [ ] Verificar si `@tanstack/react-table` ya está instalado en `apps/web/package.json`
  - Si no está: `bun add @tanstack/react-table`
- [ ] Adaptar todos los colores a nuestros tokens (`bg-surface`, `text-fg-muted`, etc.)
- [ ] Guardar como `apps/web/src/components/ui/data-table.tsx`
- [ ] Integrar en `UsersAdmin.tsx` como reemplazo de la tabla actual

### F5.2 — Formularios avanzados (de origin-ui.com)

- [ ] Ir a [originui.com](https://originui.com) → buscar formulario con validación inline
- [ ] Copiar el patrón de error inline (texto rojo bajo el input, no un toast)
- [ ] Adaptar con nuestros tokens: `text-destructive text-xs mt-1`
- [ ] Aplicar al formulario de login y al formulario de creación de usuario

### F5.3 — Stat cards para analytics (de 21st.dev)

- [ ] Ir a [21st.dev](https://21st.dev) → buscar "stat card" o "metric card"
- [ ] Seleccionar uno con: valor principal, etiqueta, delta (+ o -), y ícono
- [ ] Adaptar con tokens: `bg-surface border border-border rounded-lg`
- [ ] Guardar como `apps/web/src/components/ui/stat-card.tsx`
- [ ] Reemplazar las stat cards actuales en `AnalyticsDashboard.tsx`

### F5.4 — Empty placeholder

- [ ] Crear `apps/web/src/components/ui/empty-placeholder.tsx`:
```typescript
// Estructura:
// <EmptyPlaceholder>
//   <EmptyPlaceholder.Icon icon={MessageSquare} />
//   <EmptyPlaceholder.Title>Sin conversaciones</EmptyPlaceholder.Title>
//   <EmptyPlaceholder.Description>...</EmptyPlaceholder.Description>
//   <Button>Empezar</Button>
// </EmptyPlaceholder>
```
- [ ] Estilo: centrado, ícono grande en `text-fg-subtle`, título `text-fg`, descripción `text-fg-muted text-sm`
- [ ] Usar en: `/chat` (sin sesiones), `/collections` (sin colecciones), `/projects` (sin proyectos)

### F5.5 — Kanban de ingesta (refinar IngestionKanban)

- [ ] Columnas: `bg-surface border border-border rounded-lg`
- [ ] Cards: `bg-bg border border-border rounded-md shadow-sm hover:shadow-md transition-shadow`
- [ ] Badge de estado: usar las variantes de `badge.tsx` (success, warning, destructive)
- [ ] Drag handles: visibles en hover con `text-fg-subtle`

### Criterio de done
Tabla avanzada en UsersAdmin, stat cards en Analytics, empty states en Chat/Collections.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] `bun run type-check` — sin errores
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "feat(ui): data-table, stat-card y empty-placeholder de comunidad"`

---

## Fase 6 — Momentos especiales *(2-3 hs)*

Objetivo: los momentos de alta visibilidad tienen tratamiento visual especial. Login, onboarding, estados vacíos, loading, y números animados.

### F6.1 — Login con shader Radiant

El login es la primera impresión. Un shader sutil animado de fondo + formulario crema flotante.

- [ ] Ir a [github.com/pbakaus/radiant](https://github.com/pbakaus/radiant) — leer el README
- [ ] Copiar el shader GLSL relevante (mesh gradient animado, colores: cream `#faf8f4`, navy `#1a5276`, warm beige `#f0ebe0`)
- [ ] Crear `apps/web/src/components/auth/ShaderBackground.tsx`:
  ```typescript
  'use client'
  // Usa WebGL para el shader. Fallback a gradiente CSS si WebGL no disponible.
  // Detectar: typeof WebGLRenderingContext !== 'undefined'
  // Fallback CSS: background: linear-gradient(135deg, #faf8f4 0%, #d4e8f7 50%, #f0ebe0 100%)
  ```
- [ ] Verificar fallback: deshabilitar WebGL en DevTools → ver gradiente CSS
- [ ] Formulario de login: `bg-bg/90 backdrop-blur-sm border border-border rounded-xl shadow-xl p-8 max-w-sm w-full`
- [ ] Aplicar en `apps/web/src/app/(auth)/login/page.tsx`

### F6.2 — Onboarding con spotlight (Aceternity)

- [ ] Ir a [ui.aceternity.com](https://ui.aceternity.com) → buscar "spotlight" o "feature section"
- [ ] Copiar el efecto spotlight (sigue al cursor, resalta el elemento)
- [ ] Crear `apps/web/src/components/onboarding/SpotlightStep.tsx`
- [ ] Adaptar colores: spotlight en navy con 15% opacidad
- [ ] Integrar en `OnboardingTour.tsx` — reemplazar los tooltips planos actuales por spotlight steps
- [ ] Instalar `framer-motion` si no está: `bun add framer-motion` (para las animaciones de Aceternity)

### F6.3 — Loading states (shimmer)

Reemplazar todos los spinners genéricos por un shimmer en crema:

- [ ] Crear `apps/web/src/components/ui/skeleton.tsx` (si no existe):
  ```css
  /* Shimmer animation */
  @keyframes shimmer {
    from { background-position: -200% 0; }
    to   { background-position:  200% 0; }
  }

  .skeleton {
    background: linear-gradient(90deg,
      var(--surface) 25%,
      var(--surface-2) 50%,
      var(--surface) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s ease-in-out infinite;
    border-radius: var(--radius);
  }
  ```
- [ ] Crear skeleton variants: `SkeletonText`, `SkeletonAvatar`, `SkeletonCard`, `SkeletonTable`
- [ ] Reemplazar spinners en: `ChatInterface` (esperando respuesta), `CollectionsList` (cargando), `AnalyticsDashboard` (cargando datos)

### F6.4 — Números animados en analytics (MagicUI)

- [ ] Ir a [magicui.design](https://magicui.design) → buscar "number ticker" o "animated counter"
- [ ] Copiar el componente de contador animado
- [ ] Adaptar con `framer-motion` (ya instalado en F6.2)
- [ ] Crear `apps/web/src/components/ui/number-ticker.tsx`
- [ ] Aplicar en `AnalyticsDashboard.tsx` para los números de: queries totales, usuarios activos, documentos indexados

### F6.5 — Texto reveal en vacíos y confirmaciones (MagicUI)

- [ ] Copiar "text reveal" o "typing animation" de MagicUI
- [ ] Crear `apps/web/src/components/ui/text-reveal.tsx`
- [ ] Aplicar en el empty state del chat: el mensaje "Hacé una pregunta" aparece con animación suave

### Criterio de done
Login con shader (+ fallback), onboarding con spotlight, shimmer reemplaza spinners, números animados en analytics.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] `bun run build` — sin errores (los efectos no rompen el build)
- [ ] Probar login con WebGL desactivado — ver fallback
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "feat(ui): momentos especiales — shader login, spotlight onboarding, shimmer, animaciones"`

---

## Fase 7 — Storybook *(3-4 hs)*

Objetivo: Storybook 8 configurado con todos los addons. Stories para los 62 componentes. Base para visual regression del Plan 6.

**Archivos a crear:**
- `.storybook/main.ts`
- `.storybook/preview.ts`
- `apps/web/stories/**/*.stories.tsx` (62 archivos)

### F7.1 — Instalación Storybook 8

```bash
cd apps/web
bunx storybook@latest init --skip-install
bun install
```

- [ ] Correr el comando de init
- [ ] Verificar que `.storybook/main.ts` y `.storybook/preview.ts` fueron creados
- [ ] Verificar que `bun run storybook` levanta en el puerto 6006

### F7.2 — Configurar main.ts

```typescript
// .storybook/main.ts
import type { StorybookConfig } from '@storybook/nextjs'

const config: StorybookConfig = {
  stories: ['../stories/**/*.stories.@(ts|tsx)'],
  addons: [
    '@storybook/addon-essentials',
    '@storybook/addon-a11y',
    '@storybook/addon-themes',
    '@chromatic-com/storybook',
  ],
  framework: {
    name: '@storybook/nextjs',
    options: {},
  },
  docs: { autodocs: 'tag' },
}

export default config
```

- [ ] Actualizar `main.ts` con la config completa
- [ ] Instalar addons: `bun add -D @storybook/addon-a11y @storybook/addon-themes @chromatic-com/storybook`

### F7.3 — Configurar preview.ts

```typescript
// .storybook/preview.ts
import type { Preview } from '@storybook/react'
import '../src/app/globals.css'
import { withThemeByClassName } from '@storybook/addon-themes'

const preview: Preview = {
  parameters: {
    backgrounds: { disable: true }, // usamos nuestros tokens CSS
    layout: 'centered',
  },
  decorators: [
    withThemeByClassName({
      themes: {
        light: '',      // clase vacía — light es el default de :root
        dark: 'dark',   // aplica .dark al html
      },
      defaultTheme: 'light',
    }),
  ],
}

export default preview
```

- [ ] Actualizar `preview.ts`
- [ ] Verificar: toggle light/dark en Storybook cambia los colores correctamente

### F7.4 — Agregar scripts al package.json de apps/web

```json
"storybook": "storybook dev -p 6006",
"build:storybook": "storybook build",
"test:storybook": "test-storybook"
```

- [ ] Agregar scripts
- [ ] Agregar task en `turbo.json`: `"build:storybook": { "dependsOn": ["^build"], "outputs": ["storybook-static/**"] }`

### F7.5 — Stories de primitivos

Para cada componente en `ui/`, crear un archivo `stories/primitives/<nombre>.stories.tsx` con:
- Story `Default`
- Story por cada variante principal
- Story `Dark` (forzar dark mode)
- Story `AllVariants` (grilla de todas las variantes — para visual regression)

- [ ] `stories/primitives/button.stories.tsx` — variantes: default, destructive, outline, secondary, ghost, sizes
- [ ] `stories/primitives/input.stories.tsx` — default, error state, disabled, with icon
- [ ] `stories/primitives/badge.stories.tsx` — las 6 variantes en una grilla
- [ ] `stories/primitives/avatar.stories.tsx` — con imagen, con fallback iniciales
- [ ] `stories/primitives/dialog.stories.tsx` — abierto (usar `open` prop fixed para Storybook)
- [ ] `stories/primitives/table.stories.tsx` — con datos de ejemplo (5 filas mínimo)
- [ ] `stories/primitives/tooltip.stories.tsx`
- [ ] `stories/primitives/command.stories.tsx` — con items de búsqueda pre-populados
- [ ] `stories/primitives/skeleton.stories.tsx` — todas las variantes en animación

### F7.6 — Stories de layout

- [ ] `stories/layout/navrail.stories.tsx` — light, dark, icon-only, con badge de notificación
- [ ] `stories/layout/appshell.stories.tsx` — layout completo con contenido placeholder
- [ ] `stories/layout/density.stories.tsx` — mismo componente en compact vs spacious side-by-side

### F7.7 — Stories de features (prioritarios)

- [ ] `stories/features/data-table.stories.tsx` — con 20 filas de datos mock de usuarios
- [ ] `stories/features/stat-card.stories.tsx` — 4 tarjetas en grilla
- [ ] `stories/features/empty-placeholder.stories.tsx` — todos los tipos
- [ ] `stories/features/ingestion-kanban.stories.tsx` — con jobs en cada estado
- [ ] `stories/features/analytics-dashboard.stories.tsx` — con datos mock
- [ ] `stories/features/chat-interface.stories.tsx` — con mensajes mock
- [ ] `stories/design-system/tokens.stories.tsx` — showcase de toda la paleta y tipografía

### F7.8 — Verificación de a11y

- [ ] Abrir Storybook y navegar al panel "Accessibility" en cada primitivo
- [ ] Corregir cualquier violación marcada como error (no warnings)
- [ ] Documentar las violaciones conocidas (warnings aceptables) en `.storybook/a11y-known-issues.md`

### Criterio de done
`bun run storybook` levanta, 62 components tienen stories, addon-a11y sin errores en primitivos.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] `bun run build:storybook` — build exitoso
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "feat(storybook): setup completo con 62 stories — base para visual regression"`

---

## Fase 8 — Páginas *(5-7 hs)*

Objetivo: aplicar el nuevo design system a las 24 páginas en orden de prioridad. Cada página migra a los nuevos tokens, densidad correcta, y componentes actualizados.

> **Nota:** la lógica de las páginas NO cambia. Solo styling, tokens, y reemplazo de componentes por los nuevos.

### F8.1 — /login (Alta prioridad)

**Archivo:** `apps/web/src/app/(auth)/login/page.tsx`

- [ ] Aplicar `ShaderBackground` (de Fase 6)
- [ ] Formulario: `max-w-sm mx-auto bg-bg/90 backdrop-blur-sm rounded-xl border border-border shadow-xl p-8`
- [ ] Logo: `text-2xl font-semibold text-fg mb-2` + ícono navy
- [ ] Tagline: `text-sm text-fg-muted mb-6`
- [ ] Inputs: usar componentes actualizados de F3.2
- [ ] Botón: `Button` variante `default` full-width
- [ ] Error state: usar el patrón de error inline de F5.2

### F8.2 — /chat (Alta prioridad)

**Archivos:** `apps/web/src/app/(app)/chat/page.tsx`, `apps/web/src/components/chat/ChatInterface.tsx`

- [ ] Empty state de chat: usar `EmptyPlaceholder` + `TextReveal` (de Fase 6)
- [ ] Input de chat: estilo prominente, `rounded-xl border-border bg-surface shadow-sm`
- [ ] Aplica `data-density="spacious"`
- [ ] Skeleton mientras carga la lista de sesiones

### F8.3 — /chat/[id] (Alta prioridad)

**Archivos:** `apps/web/src/app/(app)/chat/[id]/page.tsx`, `ChatInterface.tsx`

- [ ] Mensajes del usuario: `bg-accent text-accent-fg rounded-2xl rounded-br-sm px-4 py-3 max-w-[80%] ml-auto`
- [ ] Mensajes del asistente: `bg-surface rounded-2xl rounded-bl-sm px-4 py-3 max-w-[80%]`
- [ ] Avatar del asistente: logo de RAG Saldivia en navy, 32px
- [ ] Panel de fuentes (SourcesPanel): `bg-surface border-l border-border`
- [ ] Skeleton de "pensando": 3 líneas shimmer

### F8.4 — /admin/users (Alta prioridad)

**Archivo:** `apps/web/src/components/admin/UsersAdmin.tsx`

- [ ] Reemplazar tabla actual por `DataTable` (de F5.1)
- [ ] Badges de rol: usar variantes de `badge.tsx`
- [ ] Header de sección: `text-lg font-semibold text-fg` + botón "Nuevo usuario"
- [ ] `data-density="compact"`

### F8.5 — /admin/analytics (Alta prioridad)

**Archivo:** `apps/web/src/components/admin/AnalyticsDashboard.tsx`

- [ ] Stat cards: usar `StatCard` (de F5.3) con `NumberTicker` (de F6.4)
- [ ] Gráficos (recharts): actualizar colores a `#1a5276` (accent navy) y `#d4e8f7` (accent-subtle)
- [ ] Layout: grid de 4 stat cards arriba + gráficos debajo

### F8.6 — /collections (Media prioridad)

**Archivo:** `apps/web/src/components/collections/CollectionsList.tsx`

- [ ] Cards de colección: `bg-surface border border-border rounded-lg hover:shadow-md transition-shadow`
- [ ] Empty state: usar `EmptyPlaceholder` con ícono de carpeta

### F8.7 — /settings (Media prioridad)

**Archivo:** `apps/web/src/components/settings/SettingsClient.tsx`

- [ ] Layout: sidebar de secciones + área de contenido
- [ ] Secciones con `border-b border-border pb-6 mb-6`
- [ ] Labels: `text-sm font-medium text-fg`

### F8.8 — /upload (Media prioridad)

**Archivo:** `apps/web/src/components/upload/UploadClient.tsx`

- [ ] Drop zone: `border-2 border-dashed border-border rounded-xl hover:border-accent transition-colors`
- [ ] Estado de upload: progress bar en `bg-accent`
- [ ] Lista de archivos: `bg-surface rounded-lg`

### F8.9 — /admin/ingestion (Media prioridad)

**Archivo:** `apps/web/src/components/admin/IngestionKanban.tsx`

- [ ] Aplicar estilos de F5.5 (Kanban refinado)
- [ ] `data-density="compact"`

### F8.10 — /admin/system (Media prioridad)

**Archivo:** `apps/web/src/components/admin/SystemStatus.tsx`

- [ ] Status indicators: punto verde/amarillo/rojo con `bg-success`/`bg-warning`/`bg-destructive`
- [ ] Cards de servicio: `bg-surface border border-border rounded-lg`

### F8.11 — /projects (Media prioridad)

**Archivo:** `apps/web/src/components/projects/ProjectsClient.tsx`

- [ ] Cards de proyecto con `EmptyPlaceholder` si no hay proyectos

### F8.12-F8.24 — Páginas restantes (Baja prioridad)

Aplicar el mismo patrón (tokens + componentes actualizados) a:

- [ ] `/extract` — `ExtractionWizard.tsx`: steps con navy active
- [ ] `/admin/areas` — `AreasAdmin.tsx`: `DataTable`
- [ ] `/admin/permissions` — `PermissionsAdmin.tsx`
- [ ] `/admin/rag-config` — `RagConfigAdmin.tsx`
- [ ] `/admin/knowledge-gaps` — `KnowledgeGapsClient.tsx`
- [ ] `/admin/reports` — `ReportsAdmin.tsx`
- [ ] `/admin/webhooks` — `WebhooksAdmin.tsx`
- [ ] `/admin/integrations` — `IntegrationsAdmin.tsx`
- [ ] `/admin/external-sources` — `ExternalSourcesAdmin.tsx`
- [ ] `/audit` — `AuditTable.tsx`: `DataTable`
- [ ] `/saved` — skeleton + empty placeholder
- [ ] `/settings/memory` — mismo patrón que /settings
- [ ] `/share/[token]` — página pública, layout simple sin NavRail

### Criterio de done
Las 24 páginas usan los nuevos tokens. Nada se ve con los colores viejos.

### Checklist de cierre
- [ ] `bun run test` — todos pasan
- [ ] `bun run build` — build completo exitoso
- [ ] Navegar cada página y verificar visualmente (light + dark)
- [ ] CHANGELOG.md actualizado
- [ ] `git commit -m "style(pages): aplicar design system crema-navy a las 24 páginas"`

---

## Estado global

| Fase | Estado | Fecha |
|------|--------|-------|
| Fase 1 — Tokens CSS | ✅ completado | 2026-03-26 |
| Fase 2 — Tipografía | ✅ completado | 2026-03-26 |
| Fase 3 — Primitivos shadcn | ✅ completado | 2026-03-26 |
| Fase 4 — Layout system | ✅ completado | 2026-03-26 |
| Fase 5 — Componentes comunidad | ✅ completado | 2026-03-26 |
| Fase 6 — Momentos especiales | ✅ completado | 2026-03-26 |
| Fase 7 — Storybook | ⬜ pendiente | — |
| Fase 8 — Páginas | ⬜ pendiente | — |

## Estimaciones

| Fase | Estimación |
|------|-----------|
| Fase 1 — Tokens | 2-3 hs |
| Fase 2 — Tipografía | 1 hs |
| Fase 3 — Primitivos | 3-4 hs |
| Fase 4 — Layout | 2-3 hs |
| Fase 5 — Comunidad | 3-4 hs |
| Fase 6 — Especiales | 2-3 hs |
| Fase 7 — Storybook | 3-4 hs |
| Fase 8 — Páginas | 5-7 hs |
| **Total** | **21-29 hs** |

## Resultado esperado

| Métrica | Inicio | Meta |
|---------|--------|------|
| Tokens CSS consistentes | 0 (hardcoded) | 100% |
| Componentes con stories | 0 | 62 |
| Páginas con design system | 0 | 24 |
| Dark mode funcional | parcial | completo |
| WCAG AA violations | desconocido | 0 en primitivos |
| Build exitoso post-cambios | — | ✅ |
