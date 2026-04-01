# Plan 4 — Product Roadmap: UI/UX & Features

> **Estado:** COMPLETADO — 2026-03-25/26
> **Branch:** `experimental/ultra-optimize`
> **Líneas originales:** ~874 → comprimido a resumen post-ejecución

---

## Qué se hizo

Construcción de 50 features de producto sobre el stack técnico ya validado. Se instaló shadcn/ui, se creó el layout dual-sidebar, se agregaron todas las features de UX que transformaron el stack técnico en un producto usable.

### Stack agregado

| Librería | Uso |
|---|---|
| `next-themes` | Dark mode toggle |
| `shadcn/ui` + `radix-ui` | 13 componentes base |
| `clsx` + `tailwind-merge` | Utilidad `cn()` |
| `sonner` | Toasts |
| `react-hotkeys-hook` | Atajos de teclado |
| `cmdk` | Command palette |
| `recharts` | Gráficos en analytics |

### Fases ejecutadas

| Fase | Qué | Features |
|------|-----|----------|
| 0 | Fundación | Design tokens crema-índigo, shadcn setup, dark mode, dual-sidebar layout (NavRail 44px + SecondaryPanel + main) |
| 1 | Chat enhanced | Zen mode, export session, saved responses, templates, keyboard shortcuts, feedback mejorado, onboarding tour |
| 2 | Admin + Power features | Command palette, analytics dashboard con recharts, notifications system, rate limiting UI, preferences granulares |
| 3 | Advanced features | Annotations/highlights, share sessions, session forking, multi-collection support, webhooks UI, scheduled reports |

### Componentes creados

- `NavRail.tsx` — barra lateral de íconos (siempre oscura)
- `AppShellChrome.tsx` — Client Component wrapper (estado de UI)
- `SecondaryPanel.tsx` — panel contextual por ruta
- 13 componentes shadcn/ui instalados
- Command palette, analytics charts, notification center

### Resultado

- 50 features de producto implementadas
- Layout transformado: sidebar único → NavRail + panel contextual
- Dark mode funcional con persistencia
- 79 tests seguían pasando tras cada fase

### Commits

**Fase 0 — Fundación:**

| Commit | Descripción |
|--------|-------------|
| `ffd3d40` | feat(web): design tokens crema-indigo light/dark — fase 0a |
| `9163d6d` | feat(web): shadcn/ui setup + 13 componentes base — fase 0b |
| `220c775` | feat(web): dark mode toggle com next-themes + script anti-fouc — fase 0c |
| `1671f43` | feat(web): dual sidebar layout — navrail + secondary panel contextual — fase 0d |
| `c57b0f8` | fix(web): hydration mismatch en theme-toggle + deps radix faltantes |

**Docs y housekeeping:**

| Commit | Descripción |
|--------|-------------|
| `3f6c820` | docs(plans): agregar roadmap de producto — 50 features en 4 fases |
| `216f33a` | docs(plans): marcar fase 0 completa — plan4 product roadmap |
| `c4c4615` | docs(plans): marcar todas las tareas de fase 1 completadas con fechas y notas |
| `784b66c` | docs(plans): marcar fase 1 completa — plan4 product roadmap |
| `53ff415` | docs(plans): expandir fase 2 con 20 features detalladas |
| `52353e2` | docs(plans): marcar fase 2 completa — plan4 product roadmap |
| `1d0a094` | docs(plans): marcar todas las tareas de fase 2 completadas |
| `a9793e8` | docs(plans): expandir fase 3 con 12 features detalladas |
| `8c6391f` | docs(plans): marcar fase 3 completa — plan4 product roadmap |
| `9a2ea87` | chore: gitignore .cursor + .playwright-mcp + next-env.d.ts |
| `f8ccc4c` | chore: agregar skills de cursor al repo |
| `d7c6b1d` | chore: agregar template de configuracion de mcps para nueva maquina |
| `6096b3e` | chore: keys reales en mcp template — repo privado |

**Fixes transversales (type errors encontrados durante ejecución):**

| Commit | Descripción |
|--------|-------------|
| `9ccb584` | fix(web): tipos de eventos de log invalidos en webhook dispatcher |
| `8a35af7` | fix(web): userid no estaba en scope en processjob del worker |
| `2a25fc8` | fix(web): type errors en analytics dashboard y ingestion route |
| `9ed1a19` | fix(web): type errors en ingestion worker, docpreviewpanel, sessionlist, next-auth |
| `e36b025` | fix(web): tipos locales para web speech api — type-check pasa en pre-push |
| `61e6356` | fix(web): importar artifactdata desde detect-artifact para type-check |

**Fase 1 — Chat Enhanced (f1.5–f1.18):**

| Commit | Descripción |
|--------|-------------|
| `8f3078c` | feat(chat): thinking steps visibles durante streaming — f1.5 |
| `37ee162` | feat(chat): feedback botones shadcn con color activo — f1.6 |
| `f195897` | feat(chat): modos de foco con system prompt — f1.7 |
| `a9d63ff` | feat(chat): voice input con web speech api + fallback graceful — f1.8 |
| `d363d2f` | feat(chat): export de sesion pdf y markdown con fuentes — f1.9 |
| `0de065a` | feat(chat): respuestas guardadas + pagina /saved + bookmark — f1.10 |
| `de7caad` | feat(web): modo zen cmd+shift+z — f1.11 |
| `175d099` | feat(web): notificaciones in-app con sonner + badge en navrail — f1.12 |
| `d75970c` | feat(rag): multilenguaje automatico por deteccion de query — f1.13 |
| `2d6ecf7` | feat(web): atajos de teclado globales cmd+n — f1.14 |
| `d1d6dca` | feat(chat): regenerar respuesta + copy + stats de query — f1.15/16/17 |
| `bb54b1d` | feat(web): panel que hay de nuevo + badge de version — f1.18 |
| `6e4e446` | test(fase1): 47 tests nuevos — saved, focus-modes, export, changelog, detect-language |

**Fase 2 — Admin + Power Features (f2.19–f2.38):**

| Commit | Descripción |
|--------|-------------|
| `abcb804` | feat(chat): panel de fuentes y citas bajo respuestas — f2.19 |
| `b851d0b` | feat(chat): preguntas relacionadas despues de cada respuesta — f2.20 |
| `cdf30b6` | feat(chat): selector multi-coleccion en query — f2.21 |
| `5d0cad6` | feat(chat): anotar fragmentos de respuesta con popover — f2.22 |
| `6ac2d52` | feat(web): command palette cmd+k con navegacion y sesiones recientes — f2.23 |
| `a8c7479` | feat(chat): etiquetas en sesiones + bulk actions + filtro por tag — f2.24 |
| `0886fb2` | feat(chat): compartir sesion con token publico + pagina read-only — f2.25 |
| `801d3ac` | feat(collections): pagina con crud desde ui — f2.26 |
| `5ce8cc3` | feat(collections): chat con documento especifico — f2.27 |
| `5c7e0ea` | feat(chat): templates de query con admin crud — f2.28 |
| `12ea7dc` | feat(admin): ingestion kanban con progreso sse en tiempo real — f2.29 |
| `7e3713f` | feat(admin): analytics dashboard con recharts — f2.30 |
| `92ace26` | feat(admin): brechas de conocimiento detectadas por heuristica — f2.31 |
| `03d0ce0` | feat(collections): historial de ingestas como commits de coleccion — f2.32 |
| `e894bc2` | feat(admin): informes programados con cron worker y destino saved/email — f2.33 |
| `283ef4f` | feat(chat): vista dividida con dos sesiones paralelas — f2.34 |
| `cb839d8` | feat(chat): drag and drop de archivos al chat — f2.35 |
| `890f4b7` | feat(admin): rate limiting por area/usuario con check en generate — f2.36 |
| `3f3f669` | feat(web): onboarding interactivo con driver.js — f2.37 |
| `37f014f` | feat(admin): webhooks salientes con hmac-sha256 — f2.38 |

**Fase 3 — Advanced Features (f3.39–f3.50):**

| Commit | Descripción |
|--------|-------------|
| `04a505f` | feat(web): busqueda universal fts5 en command palette — f3.39 |
| `fe7b4d2` | feat(chat): preview de doc inline con react-pdf — f3.40 |
| `24a1599` | feat(web): proyectos con contexto — schema, paginas, panel sidebar, inject — f3.41 |
| `ae8b86b` | feat(chat): artifacts panel — deteccion en stream y sheet lateral — f3.42 |
| `425a0f2` | feat(chat): bifurcacion de conversaciones con forked_from — f3.43 |
| `a51350c` | feat(web): memoria de usuario inyectada en queries + settings ui — f3.44 |
| `7e56997` | feat(web): superficie proactiva — notificaciones de docs nuevos relevantes — f3.45 |
| `8212035` | feat(collections): grafo de documentos con simulacion force-directed — f3.46 |
| `cc87b2b` | feat(auth): sso google e azure ad con next-auth v5 modo mixto — f3.47 |
| `ae14597` | feat(admin): auto-ingesta externa google drive sharepoint confluence — f3.48 |
| `3fe1792` | feat(web): bot slack y teams respetando rbac — f3.49 |
| `b2cc212` | feat(web): extraccion estructurada a tabla exportable como csv — f3.50 |
