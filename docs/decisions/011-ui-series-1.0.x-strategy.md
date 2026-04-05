# ADR-011: Estrategia de UI para la serie 1.0.x

**Fecha:** 2026-03-27
**Estado:** Superada por Plan Maestro 1.0.x (`docs/plans/1.0.x-plan-maestro.md`)
**Contexto:** Plan 13 — UI Foundation "Claude-style"

---

## Contexto

En el release 1.0.0 la UI ("Warm Intelligence") funcionaba correctamente pero se sentía genérica:
tonos crema saturados, borders de 1px con tinte amarillo, NavRail icon-only sin labels, y Instrument Sans como font. No reflejaba la calidad del stack técnico subyacente.

La referencia elegida es **Claude.ai** (Anthropic) — investigada directamente con firecrawl + tokens oficiales de `claude.com/docs`. Razones:
- El producto es un chat RAG, mismo dominio
- El design language de Anthropic es editorial y cálido (no tech-frío), lo que genera confianza
- Los tokens oficiales están documentados y son estables
- Tiene un design system aplicable directamente sin invenciones propias

El rediseño completo de 24 páginas en un solo plan sería demasiado ambicioso y frágil. Se eligió una estrategia de **cimiento incremental** que se ejecuta en la serie 1.0.x.

---

## Opciones consideradas

- **Opción A — Redesign big bang:** rediseñar todas las 24 páginas en un único plan.
  Pros: consistencia inmediata. Contras: plan enorme, alto riesgo de regresiones, difícil de revisar.

- **Opción B — Cimiento incremental (elegida):** establecer los tokens y shell en 1.0.1, extender a más páginas en 1.0.2, 1.0.3, etc. Cada release es autónomo y testeable.
  Pros: bajo riesgo, fácil de revisar, cada step tiene valor propio. Contras: la consistencia llega gradualmente.

- **Opción C — Design system library separada:** construir un paquete `@rag-saldivia/design` con todos los tokens y componentes, integrarlo en múltiples pasos.
  Pros: reutilizable. Contras: overhead innecesario para un monorepo con un único frontend.

---

## Decisión

Elegimos **Opción B — Cimiento incremental** porque maximiza el valor entregado por release mientras minimiza el riesgo de regresiones.

### Paleta de colores

**Light mode — tokens oficiales Claude/Anthropic:**

| Token | Valor | Fuente |
|---|---|---|
| `--bg` | `#FAF9F5` | Claude bg-tertiary oficial |
| `--surface` | `#F5F4ED` | Claude bg-secondary oficial |
| `--surface-2` | `#EDEADF` | interpolado |
| `--border` | `rgba(31, 30, 29, 0.12)` | Claude border-tertiary oficial |
| `--border-strong` | `rgba(31, 30, 29, 0.25)` | Claude border-secondary oficial |
| `--fg` | `#141413` | Claude text-primary oficial |
| `--fg-muted` | `#3D3D3A` | Claude text-secondary oficial |
| `--fg-subtle` | `#73726C` | Claude text-tertiary oficial |
| `--accent` | `#0066cc` | Azure — decisión del proyecto |

**Dark mode:** se rehace desde cero en 1.0.4 con los dark tokens oficiales de Claude. Los tokens `.dark {}` actuales quedan intactos hasta entonces.

### Tipografía

- **Font:** DM Sans (Google Fonts, variable font con optical sizing)
- **Razón:** closest open-source a "Anthropic Sans". Diseñada por Colophon Foundry para DeepMind. Combina geometric + humanist corrections, optimizada para 12-16px.
- **Base size:** 16px. **Line-height body:** 1.65. **Letter-spacing:** 0 (sin negativo en body).
- **Nota:** la serif de headings que usa Anthropic en su marketing NO se adopta. En el chat/app, DM Sans para todo.

### Border system

- **Width:** 0.5px (en lugar de 1px). Efecto: UI radicalmente más liviana.
- **Color:** semitransparente `rgba(31, 30, 29, 0.12)` — se adapta naturalmente a distintos fondos.
- **Radius:** sistema de 5 niveles: 4 / 6 / **8** (default) / 10 / 12 px.

### Botón primario

El CTA primario usa `#141413` (negro/oscuro) sobre ivory — igual que Claude. El `--accent` azure se reserva para links, focus rings y estados de selección activa. Esto se implementa poniendo `--primary: #141413` en los tokens, que shadcn/ui usa automáticamente.

### NavRail

Pasa de icon-only (44px) a sidebar con labels (196px). El cambio de layout más visible del plan 13.

### Dark mode

**Fuera del scope de la serie 1.0.x hasta 1.0.4.** Los tokens dark existentes se mantienen intactos para no romper el toggle existente. En 1.0.4 se rehace completo con los dark tokens oficiales de Claude:

| Token | Dark oficial Claude |
|---|---|
| `--bg` | `#141413` |
| `--surface` | `#1F1E1D` |
| `--fg` | `#FAF9F5` |
| `--border` | `rgba(222, 220, 209, 0.15)` |

---

## Estructura de trabajo en 1.0.x

Cada release de la serie es ejecutado por un plan numerado. Los planes viven en `docs/plans/`.

| Release | Plan | Scope | Estado |
|---|---|---|---|
| 1.0.1 | Plan 13 | Tokens + Font + AppShell/NavRail + Login + Chat | En curso |
| 1.0.2 | Plan 14 | /collections, /upload, /extract, /saved, /projects + componentes UI | Pendiente |
| 1.0.3 | Plan 15 | 12 páginas admin + DataTable + /audit + /settings | Pendiente |
| 1.0.4 | Plan 16 | Dark mode rehecho desde cero | Pendiente |

**Reglas de la serie:**
1. Cada plan no rompe el anterior. Los tests de lógica siempre pasan.
2. Cada plan regenera los visual baselines al cierre (el último step).
3. CHANGELOG: cada task agrega a `[Unreleased]`. El último task del plan mueve a `[1.0.x]`.
4. Commits granulares: un commit por fase. Mensaje: `style/feat/test(scope): descripcion — planN fX`.
5. Los cambios son CSS/visual. No se toca lógica de negocio, auth, DB ni API routes.
6. Tests de componentes deben pasar en cada fase (no solo al final).

---

## Consecuencias

**Positivas:**
- Cada release entrega valor visual concreto y verificable
- Los tokens CSS siendo la base, un cambio de accent en el futuro es editar un solo valor en `globals.css`
- La elección de tokens oficiales Claude da estabilidad — si Anthropic actualiza su design, podemos seguirlo
- DM Sans como font unifica todo el producto con un estilo más profesional inmediatamente

**Negativas / trade-offs:**
- La UI tendrá inconsistencias temporales entre páginas hasta que complete 1.0.3 (algunas páginas en nuevo design, otras en old)
- Dark mode queda en "estado de transición" hasta 1.0.4 — el toggle existe pero el dark mode no refleja el nuevo sistema
- El sidebar de 196px reduce el espacio disponible para contenido en pantallas < 1280px

## Referencias

- Tokens oficiales Claude: `claude.com/docs/connectors/building/mcp-apps/design-guidelines`
- Anthropic Design Language: análisis en `seedflip.co/blog/anthropic-design-language`
- Spec completo: `docs/superpowers/specs/2026-03-27-ui-foundation-claude-style-design.md` (gitignored — local only)
- Plan de implementación: `docs/plans/plan13-ui-foundation.md`
