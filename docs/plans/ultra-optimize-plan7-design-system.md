# Plan 7 — Design System "Warm Intelligence"

> **Estado:** COMPLETADO — 2026-03-26
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~1046 → comprimido a resumen post-ejecución
> **Spec:** `docs/superpowers/specs/2026-03-26-design-system-design.md`

---

## Qué se hizo

Identidad visual completa de Saldivia RAG. Sistema de tokens CSS coherente, tipografía Instrument Sans, paleta crema + navy, dark mode cálido, densidad adaptiva, Storybook como catálogo.

### Decisiones de diseño (permanentes)

- **Estética:** Warm Intelligence — crema cálido, tranquilidad, nitidez
- **Acento:** Navy azul profundo (`#1a5276`)
- **Tipografía:** Instrument Sans (Google Fonts via `next/font`)
- **Dark mode:** warm dark (`#1a1812` — nunca negro frío)
- **Densidad:** adaptiva (`compact` en admin, `spacious` en chat)
- **Componentes:** shadcn base + origin-ui + shadcnblocks

### Fases ejecutadas

| Fase | Qué | Resultado |
|------|-----|-----------|
| 1 | Tokens CSS | `globals.css` reescrito: paleta light crema-navy + dark warm. Aliases shadcn para backward compat |
| 2 | Tipografía | Instrument Sans via `next/font/google`. Escala tipográfica definida |
| 3 | Componentes rediseñados | Todos los componentes actualizados a nuevos tokens |
| 4 | Dark mode refinado | Tokens dark independientes, contraste verificado |
| 5 | Storybook 8 | Catálogo de componentes como entorno de desarrollo y visual regression |
| 6 | Densidad adaptiva | CSS variables `--density-*` para compact/spacious |

### Tokens CSS resultantes (light)

```css
--bg: #faf8f4;  --surface: #f0ebe0;  --surface-2: #e8e1d4;
--fg: #1a1a1a;  --fg-muted: #5a5048;  --fg-subtle: #9a9088;
--accent: #1a5276;  --accent-hover: #154360;  --accent-subtle: #d4e8f7;
```

### Commits

| Fase | Commit | Descripción |
|------|--------|-------------|
| F1/F2 | `b112851` | feat(web): tokens crema-navy + instrument sans |
| F3 | `972a1d1` | feat(web): reestilizar primitivos shadcn con tokens crema-navy |
| F4 | `3594576` | feat(web): layout system con tokens crema-navy y densidad adaptiva |
| F5 | `73d0db5` | feat(web): componentes de comunidad — data-table, stat-card, skeleton, empty-placeholder |
| F5 | `300c388` | fix(web): agregar postcss.config.js con @tailwindcss/postcss |
| F5 | `0e023e8` | fix(web): mover react-scan dynamic import a client component |
| F6 | `c4f66ab` | feat(web): momentos especiales — login redesign, animated background, skeletons |
| F6 | `452cf28` | fix(web): aumentar visibilidad de orbes animados en login |
| F6 | `54d70ea` | refactor(web): sacar animación del login — fondo crema estático, card limpio |
| F7 | `5fd40f6` | feat(web): storybook 8 con react-vite — design system tokens + 9 stories |
| F7 | `4503fe6` | fix(web): storybook — jsx automatic transform, preview.tsx con bg-bg |
| deps | `d7bd068` | chore(deps): actualizar react 19.2.4, tailwindcss 4.2.2, typescript 6.0.2 |
| F8 | `000831d` | feat(web): redesign admin/users y admin/analytics con design system |
| F8.2 | `0c0d3a6` | fix(web): sessionslist sin inline styles — tokens tailwind |
| F8.3 | `13fdad3` | feat(web): chatinterface con tokens crema-navy — mensajes, input, feedback |
| F8.7-11 | `92c6079` | feat(web): settings, upload, system-status y projects con design system |
| F8.9/12 | `6187428` | feat(web): redesign masivo — 10 componentes admin y audit con design system |
| Cierre | `7dd8cc4` | feat(web): completar f8 — extract, saved, settings/memory, share con design system |
| Docs | `f759c89` | docs(plans): actualizar changelog y checklists plan 6 y 7 — f8 en progreso |
| Docs | `06bbb60` | docs: agregar secciones faltantes post-repomix + codegraphcontext |
| Docs | `1e70fd2` | docs: onboarding actualizado + skills rag-testing y rag-nextjs actualizados |

> **Nota:** Los tokens fueron posteriormente ajustados en la serie 1.0.x (Plan 15/16) a azure blue (`#2563eb`) y fondo `#faf9f5`. Este plan estableció la arquitectura de tokens que sigue vigente.
