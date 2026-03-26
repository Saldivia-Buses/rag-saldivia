# Plan 7 — Design System "Warm Intelligence"

**Fecha:** 2026-03-26  
**Estado:** Aprobado por el usuario  
**Prereqs:** Ninguno (primer plan visual del proyecto)  
**Sigue a:** ultra-optimize-plan5-testing-foundation.md  
**Seguido por:** Plan 6 (UI Testing) — depende de este plan

> **Orden de ejecución:** Plan 7 se implementa completo antes de Plan 6, con una excepción: react-scan (Plan 6, Fase 1) se corre al inicio de Plan 7 para capturar el baseline de performance previo al rediseño.

---

## Resumen ejecutivo

RAG Saldivia tiene 62 componentes y 24 páginas funcionales sin un sistema de diseño real. Este plan establece la identidad visual completa del producto: tokens de diseño, tipografía, paleta de colores, modo oscuro, sistema de densidad adaptiva, catálogo de componentes en Storybook, y rediseño aplicado a todas las páginas. El objetivo es transformar una interfaz funcional en un producto con personalidad visual consistente y memorable.

---

## Decisiones de diseño

### Estética: Warm Intelligence

Dirección inspirada en Claude.ai pero con identidad propia. Crema cálido como fondo base (no blanco frío), sin ruido visual, prioridad al contenido. La herramienta debe sentirse tranquila, inteligente y confiable.

**Principios:**
- La interfaz no compite con el contenido — es transparente
- Calidez sin decoración innecesaria
- Claridad mental: cada elemento justifica su presencia
- Inteligencia sin arrogancia

### Paleta de tokens

#### Light mode

| Token | Valor | Uso |
|---|---|---|
| `--bg` | `#faf8f4` | Fondo base de la app |
| `--surface` | `#f0ebe0` | Superficies elevadas (cards, paneles) |
| `--surface-2` | `#e8e1d4` | Superficies de segundo nivel (hover, subtablas) |
| `--border` | `#ede9e0` | Bordes y separadores |
| `--border-strong` | `#d5cfc4` | Bordes con énfasis |
| `--fg` | `#1a1a1a` | Texto principal |
| `--fg-muted` | `#5a5048` | Texto secundario |
| `--fg-subtle` | `#9a9088` | Placeholders, metadata |
| `--accent` | `#1a5276` | Navy profundo — acción principal |
| `--accent-hover` | `#154360` | Navy hover |
| `--accent-dark` | `#0d3349` | Navy pressed/active |
| `--accent-subtle` | `#d4e8f7` | Fondos de badges, highlights |
| `--accent-fg` | `#ffffff` | Texto sobre acento |
| `--destructive` | `#c0392b` | Errores, eliminación |
| `--destructive-subtle` | `#fde8e8` | Fondo de estado de error |
| `--success` | `#2d6a4f` | Confirmaciones |
| `--success-subtle` | `#d4f1e4` | Fondo de estado de éxito |
| `--warning` | `#d68910` | Advertencias |
| `--warning-subtle` | `#fef9e7` | Fondo de estado de warning |

#### Dark mode (warm dark)

| Token | Valor | Uso |
|---|---|---|
| `--bg` | `#1a1812` | Fondo base — marrón-negro cálido |
| `--surface` | `#24221a` | Superficies elevadas |
| `--surface-2` | `#2e2b21` | Segundo nivel |
| `--border` | `#35312a` | Bordes |
| `--border-strong` | `#46423a` | Bordes con énfasis |
| `--fg` | `#f0ebe0` | Texto principal — crema claro |
| `--fg-muted` | `#b0a898` | Texto secundario |
| `--fg-subtle` | `#7a7268` | Placeholders |
| `--accent` | `#4a9fd4` | Navy más claro para dark (contraste WCAG) |
| `--accent-hover` | `#5aafde` | |
| `--accent-dark` | `#6bbfe8` | |
| `--accent-subtle` | `#1a3a52` | |
| `--accent-fg` | `#0d1a24` | Texto sobre acento en dark |
| `--destructive` | `#e57373` | Errores en dark (más claro para contraste) |
| `--destructive-subtle` | `#2a1515` | Fondo de error en dark |
| `--success` | `#66bb6a` | Confirmaciones en dark |
| `--success-subtle` | `#0d2a14` | Fondo de éxito en dark |
| `--warning` | `#ffa726` | Advertencias en dark |
| `--warning-subtle` | `#2a1e00` | Fondo de warning en dark |

### Tipografía

**Fuente:** Instrument Sans (Google Fonts, self-hosted para performance)

```
Escala tipográfica:
--text-xs:   0.75rem / line-height: 1rem
--text-sm:   0.875rem / line-height: 1.25rem
--text-base: 1rem     / line-height: 1.5rem
--text-lg:   1.125rem / line-height: 1.75rem
--text-xl:   1.25rem  / line-height: 1.75rem
--text-2xl:  1.5rem   / line-height: 2rem
--text-3xl:  1.875rem / line-height: 2.25rem
--text-4xl:  2.25rem  / line-height: 2.5rem

Pesos:
400 — cuerpo normal
500 — énfasis suave, etiquetas
600 — bold funcional, headings medianos
700 — headings grandes, CTAs
```

**Letter-spacing:** `-0.01em` en textos de cuerpo para Instrument Sans (se lee más ajustado y sofisticado).

### Modo oscuro

Implementado con `next-themes`. El `ThemeProvider` se configura con `attribute="class"` (default de next-themes) para mantener compatibilidad con la clase `.dark` que shadcn/ui ya usa internamente. Las variables CSS viven en `:root` (light) y `.dark` (dark). El toggle vive en la NavRail. El dark mode es "warm dark" — nunca `#000000` sino `#1a1812`. Todos los tokens se invierten preservando la calidez.

### Densidad adaptiva

Dos niveles de densidad como variantes CSS, no como componentes separados:

**`data-density="compact"`** (admin, tablas, listas densas):
```
--spacing-row: 0.5rem      (altura de fila en tablas)
--spacing-cell: 0.5rem 0.75rem
--avatar-size: 24px
--icon-size: 14px
--input-height: 32px
```

**`data-density="spacious"`** (chat, collections, páginas de contenido):
```
--spacing-row: 1rem
--spacing-cell: 0.875rem 1rem
--avatar-size: 36px
--icon-size: 18px
--input-height: 40px
```

Cada layout aplica `data-density` al contenedor raíz. Los componentes leen los tokens y se adaptan automáticamente — no hay lógica duplicada.

---

## Stack de código comunitario

### Componentes copiables (código propio, sin dependencias extra)

| Fuente | Uso |
|---|---|
| [shadcnblocks.com](https://shadcnblocks.com) | Bloques de página: hero del login, tablas completas, sidebars |
| [origin-ui.com](https://originui.com) | Formularios avanzados, inputs complejos, selects |
| [21st.dev](https://21st.dev) | Microinteracciones, componentes modernos con animaciones |
| [Hyper UI](https://hyperui.dev) | Componentes Tailwind para páginas de marketing/auth |
| [Flowbite](https://flowbite.com) | Admin components no cubiertos por shadcn |

**Regla:** todo código de comunidad se copia y adapta — nunca se instala como dependencia. Se estiliza con nuestros tokens CSS.

### Efectos visuales (momentos especiales)

| Fuente | Uso |
|---|---|
| [Radiant shaders](https://github.com/pbakaus/radiant) | Shader GLSL animado para el fondo del login |
| [Aceternity UI](https://ui.aceternity.com) | Onboarding tour, estados vacíos, cards con spotlight |
| [Magic UI](https://magicui.design) | Números animados en analytics, texto reveal, shimmer |

### Inspiración visual (referencia)

Godly, Land-Book, Mobbin (patrones de chat/admin), Bento Grids (dashboard analytics), Refero, Framer Marketplace, Lapa Ninja.

### Herramientas

- **[uicolors.app](https://uicolors.app)** — generar la escala Tailwind completa de navy y cream
- **[HTMLrev](https://htmlrev.com)** — catálogo de templates Next.js/Shadcn para estructura de páginas

---

## Storybook

### Setup

Storybook 8 con Next.js framework. Addons:
- `@storybook/addon-a11y` — auditoría de accesibilidad por componente
- `@storybook/addon-themes` — toggle light/dark dentro de Storybook
- `@storybook/addon-viewport` — preview en distintos breakpoints
- `@chromatic-com/storybook` — preparado para visual regression (Plan 6)

### Organización de stories

```
.storybook/
  main.ts
  preview.ts          ← importa globals.css, configura themes
stories/
  design-system/
    tokens.stories.tsx     ← paleta completa, escala tipográfica
    icons.stories.tsx
  primitives/
    button.stories.tsx
    input.stories.tsx
    badge.stories.tsx
    ...
  layout/
    appshell.stories.tsx
    navrail.stories.tsx
  features/
    chat-interface.stories.tsx
    ingestion-kanban.stories.tsx
    analytics-dashboard.stories.tsx
    ...
```

**Cada story tiene estos states:** default, hover, focus, disabled, loading, error, dark mode.

### Decisiones de Storybook con consultas visuales

Cada decisión de diseño dentro de Storybook (layout de stories, variantes de componentes, naming) se consulta al usuario con preguntas visuales en el browser companion antes de implementar.

---

## Fases de implementación

### Fase 1 — Tokens (1-2 días)
Reescribir `globals.css` completo con la nueva paleta light + dark. Usar uicolors.app para generar escalas Tailwind. 

**Migración de tokens existentes:** El CSS actual usa tokens de shadcn (`--background`, `--foreground`, `--primary`, `--primary-foreground`, `--muted`, `--muted-foreground`, `--ring`, `--radius`, `--sidebar-bg`, `--nav-bg`, etc.). La estrategia es: mantener los nombres shadcn como aliases que apuntan a los nuevos tokens semánticos, para no romper los componentes shadcn existentes.

```css
/* Patrón de migración */
:root {
  /* Nuevos tokens semánticos */
  --bg: #faf8f4;
  --surface: #f0ebe0;
  --accent: #1a5276;
  /* ... */

  /* Aliases shadcn → apuntan a los nuevos tokens */
  --background: var(--bg);
  --foreground: var(--fg);
  --primary: var(--accent);
  --primary-foreground: var(--accent-fg);
  --muted: var(--surface);
  --muted-foreground: var(--fg-muted);
  /* Nota: --border usa el mismo nombre en shadcn y en nuestro sistema.
     Se redefine directamente con el nuevo valor (#ede9e0) — sin alias. */
  --ring: var(--accent);
}
```

**Exposición como utilidades Tailwind v4:** Los tokens se registran en `@theme inline` para que sean accesibles como clases utilitarias y se resuelvan en runtime contra las variables CSS (necesario para que el dark mode funcione):

```css
@theme inline {
  --color-bg: var(--bg);
  --color-surface: var(--surface);
  --color-accent: var(--accent);
  --color-accent-subtle: var(--accent-subtle);
  --color-fg: var(--fg);
  --color-fg-muted: var(--fg-muted);
  /* ... todos los tokens semánticos */
}
```

El modificador `inline` es crítico en Tailwind v4 — sin él, `var()` se resuelve en build time y el dark mode no funciona. Con `inline`, Tailwind genera utilidades que referencian la variable CSS en runtime.

Esto habilita `bg-surface`, `text-fg-muted`, `border-accent-subtle`, etc. como clases Tailwind que respetan el tema activo.

Verificar contraste WCAG en todos los pares fg/bg antes de continuar.

### Fase 2 — Tipografía (medio día)
Usar `next/font/google` para Instrument Sans — hace self-hosting automático, optimiza preload, y gestiona `font-display: swap` sin configuración manual:

```typescript
// apps/web/src/app/layout.tsx
import { Instrument_Sans } from 'next/font/google'

const instrumentSans = Instrument_Sans({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  style: ['normal', 'italic'],
  variable: '--font-sans',
})
```

La fuente se expone como variable CSS `--font-sans` y se referencia desde `globals.css`. Definir la escala tipográfica completa como tokens CSS. Aplicar `font-feature-settings: "cv01", "cv02"` para variantes de caracteres de Instrument Sans.

### Fase 3 — Primitivos shadcn (2-3 días)
Reestilizar todos los componentes shadcn existentes (button, input, textarea, dialog, popover, table, badge, avatar, separator, tooltip, sheet, command) con los nuevos tokens. Sin cambiar la API de los componentes — solo CSS.

### Fase 4 — Layout system (2 días)
- `AppShell`: aplicar nuevos tokens, implementar `data-density`
- `NavRail`: rediseño con Instrument Sans + navy + dark mode toggle
- `SecondaryPanel`, `CommandPalette`: integración al nuevo sistema
- Responsive: mobile-first con los siguientes breakpoints:

```css
/* Breakpoints del sistema */
--breakpoint-sm:  640px   /* Mobile landscape */
--breakpoint-md:  768px   /* Tablet */
--breakpoint-lg:  1024px  /* Desktop pequeño */
--breakpoint-xl:  1280px  /* Desktop estándar */
--breakpoint-2xl: 1536px  /* Desktop grande / workstation */
```

Layout principal a 3 columnas en lg+, 2 columnas en md, 1 columna en sm/mobile. NavRail colapsa a bottom bar en mobile.

### Fase 5 — Componentes comunidad (3-4 días)
Curar y adaptar los mejores componentes de shadcnblocks, origin-ui, 21st.dev, Hyper UI. Copiar como código propio, aplicar tokens. Candidatos prioritarios:
- Tabla avanzada con sorting/filtering (para usuarios, colecciones, audit)
- Form avanzado con validación inline
- Data card / stat card (para analytics)
- Kanban board (ya existe IngestionKanban — refinar)
- Command palette mejorada

### Fase 6 — Momentos especiales (2 días)
- **Login**: shader Radiant de fondo (sutil, animado), formulario crema flotante con sombra suave. **Fallback obligatorio**: si WebGL no está disponible (headless, dispositivos sin GPU), el shader se reemplaza por un gradiente CSS animado equivalente. La detección se hace con `WebGLRenderingContext` en `useEffect` client-side.
- **Onboarding**: spotlight Aceternity sobre elementos del UI
- **Estados vacíos**: ilustraciones mínimas + copy con personalidad
- **Loading states**: shimmer en crema, no spinners genéricos
- **Números en analytics**: contadores animados con Magic UI

### Fase 7 — Storybook (2-3 días)
Setup completo. Stories para los 62 componentes. Organización por categorías. Addons configurados. Baseline lista para visual regression del Plan 6.

### Fase 8 — Páginas (5-7 días)
Aplicar el sistema a las 24 páginas en orden de importancia:

**Prioridad alta:**
1. `/login` — primera impresión
2. `/chat` y `/chat/[id]` — uso más frecuente
3. `/admin/users`, `/admin/analytics` — más usadas en admin
4. `/collections` — core del producto

**Prioridad media:**
5. `/settings` y `/settings/memory`
6. `/upload` y `/extract`
7. `/admin/ingestion`, `/admin/system`
8. `/projects` y `/projects/[id]`

**Prioridad baja:**
9. Resto de páginas admin
10. `/share/[token]`, `/saved`, `/audit`

---

## Criterios de éxito

- [ ] Todos los tokens CSS documentados y aplicados consistentemente
- [ ] Contraste WCAG AA en todos los pares de texto/fondo (ratio ≥ 4.5:1)
- [ ] Dark mode sin ningún color hardcodeado (todo variables)
- [ ] Instrument Sans cargando en < 100ms (self-hosted, `font-display: swap`)
- [ ] Storybook con stories para los 62 componentes
- [ ] Cada componente con estado light y dark en Storybook
- [ ] Las 24 páginas aplicando los nuevos tokens
- [ ] `data-density` funcionando correctamente en layouts compact y spacious
- [ ] 0 regresiones en tests existentes de lógica (bun test pasa)

---

## Relación con otros planes

- **Plan 6 (Testing UI)**: depende de este plan. El Plan 6 toma el Storybook del Plan 7 como base para visual regression.
- **Plan 5 (Testing Foundation)**: este plan no modifica la lógica — los 79 tests existentes deben seguir pasando.
- **Futuros planes**: el design system se convierte en la base de cualquier nueva feature visual.
