# Plan 12: Release 1.0.0 — Version Bump, CHANGELOG, Tag y GitHub Release

> Este documento vive en `docs/plans/ultra-optimize-plan12-release-1.0.0.md`.
> Es el plan final antes del primer release público del stack TypeScript.

---

## Contexto

Los Planes 1–11 construyeron, optimizaron, testearon y documentaron el stack completo. El resultado:

- **~2.516 líneas de dead code eliminadas** a lo largo de los planes
- **413+ tests en verde** (259 lógica + 154 componentes + E2E + visual regression + a11y)
- **TypeScript a cero errores** en todos los packages
- **ESLint + commitlint + lint-staged** activos
- **Redis + BullMQ** en producción (sin workarounds)
- **Next.js 16, Zod 4, Drizzle 0.45, Lucide 1.7**
- **10 ADRs documentados**
- **Documentación completa** (README, CONTRIBUTING, SECURITY, LICENSE, API docs, ER diagram)
- **Repo remoto limpio** (sin archivos trackeados incorrectamente)

Es el momento del release 1.0.0.

**Prerequisitos:** Planes 9, 10 y 11 completados y mergeados. Todos los tests pasan. El repo remoto está limpio.

**Lo que este plan NO hace:**
- No agrega features
- No cambia el código de producción
- No implementa features RAG que requieren la workstation (esas van en versiones futuras)

---

## Orden de ejecución

```
R1 → R2 → R3 → R4 → R5 → R6 → R7
```

El orden es estricto: el tag se crea último, después de que todo está en el repo remoto y verificado.

---

## Seguimiento

Formato: `- [ ] Descripción`
Al completar: `- [x] Descripción — completado YYYY-MM-DD`

---

## R1 — Version bump a 1.0.0 *(15 min)*

Objetivo: todos los `package.json` del monorepo dicen `"version": "1.0.0"`.

**Los 7 archivos a modificar** — en cada uno, cambiar `"version": "0.1.0"` por `"version": "1.0.0"`:
1. `package.json` (root del monorepo)
2. `apps/web/package.json`
3. `apps/cli/package.json`
4. `packages/db/package.json`
5. `packages/logger/package.json`
6. `packages/shared/package.json`
7. `packages/config/package.json`

**Comando para verificar el resultado:**
```bash
grep '"version"' package.json apps/*/package.json packages/*/package.json | grep -v "workspace\|node_modules"
```
Salida esperada: 7 líneas, todas con `"version": "1.0.0"`.

**No cambiar** las referencias internas `"@rag-saldivia/db": "workspace:*"` — son workspace refs, no versiones.

- [ ] Editar los 7 archivos `package.json`: `"version": "0.1.0"` → `"version": "1.0.0"`
- [ ] `grep '"version"' package.json apps/*/package.json packages/*/package.json | grep -v workspace` → 7 líneas con `"1.0.0"`
- [ ] `export PATH="$HOME/.bun/bin:$PATH" && bun run test` → todos los tests pasan

**Estado: pendiente**

---

## R2 — CHANGELOG.md — Entrada 1.0.0 *(1-2 hs)*

Objetivo: el CHANGELOG tiene una entrada `## [1.0.0] — 2026-MM-DD` que documenta todo lo construido. La sección `[Unreleased]` queda vacía.

**Operación:** mover todo el contenido de `## [Unreleased]` a una nueva sección `## [1.0.0] — 2026-MM-DD`, y agregar el bloque de highlights al inicio de esa sección.

**Estructura exacta de la entrada `[1.0.0]`:**

```markdown
## [1.0.0] — 2026-03-XX

Primer release del stack TypeScript de RAG Saldivia. Reescritura completa del overlay
sobre NVIDIA RAG Blueprint v2.5.0 — reemplaza el stack Python + SvelteKit con un proceso
único Next.js 16 que incluye UI, autenticación, proxy RAG, admin y CLI TypeScript.

### Highlights

- **Next.js 16** App Router como proceso único — UI + auth + proxy + admin
- **Autenticación JWT** con Redis blacklist para revocación inmediata + RBAC por roles y áreas
- **BullMQ** para cola de ingesta — reemplaza worker manual + tabla SQLite
- **Design system "Warm Intelligence"** — 24 páginas, dark mode, WCAG AA
- **CLI TypeScript** — `rag users/collections/ingest/audit/config/db/status`
- **413+ tests** — lógica, componentes, visual regression, a11y, E2E Playwright
- **Código production-grade** — TypeScript strict, ESLint, commitlint, lint-staged, knip
- **10 ADRs** documentando las decisiones de arquitectura

### Plans completados (Plan 1 → Plan 11)

**Plan 1 — Monorepo TypeScript**
Birth del stack. Turborepo + Bun workspaces + Next.js 15 + Drizzle + JWT auth + CLI base.

**Plan 2 — Testing sistemático**
Primera suite de tests. 270 tests de lógica en verde. Estrategia de testing documentada.

**Plan 3 — Bugfix CodeGraphContext**
Estabilización post-birth. Fixes de imports, build, y MCP.

**Plan 4 — Product Roadmap (Fases 0–2)**
50 features en 3 fases. Design system base, dark mode, 24 páginas, shadcn/ui, design system.

**Plan 5 — Testing Foundation**
Coverage al 95% en lógica pura. 270 tests.

**Plan 6 — UI Testing Suite**
Visual regression con Playwright (22 snapshots). A11y con axe-playwright (WCAG AA).

**Plan 7 — Design System "Warm Intelligence"**
Paleta crema-navy, tokens CSS, 147 tests de componentes, Storybook 8.

**Plan 8 — Optimización + Redis + BullMQ**
~2.516 líneas eliminadas. Next.js 16, Zod 4, Drizzle 0.45, Lucide 1.7.
Redis obligatorio: JWT revocación, cache, Pub/Sub, BullMQ.
10 ADRs. CI paralelo con turbo --affected.

**Plan 9 — Repo Limpio**
46 archivos purgados del remoto. TypeScript a 0 errores.
Dead code eliminado (crossdoc, SSO stub, wrappers). ESLint + husky + commitlint.

**Plan 10 — Testing Completo**
E2E Playwright (5 flujos críticos + smoke Redis).
Visual regression verificada post-upgrades. A11y WCAG AA. Coverage ≥80%.

**Plan 11 — Documentación Perfecta**
README, CONTRIBUTING, SECURITY, LICENSE, CODEOWNERS, issue templates.
ER diagram, API reference (30+ endpoints), JSDoc en funciones críticas.
READMEs de packages. CLAUDE.md actualizado.

[aquí van todas las entradas del [Unreleased] que se movieron]
```

Después del bloque de highlights y resumen por plan, pegar todo el contenido que estaba en `[Unreleased]`.

**La sección `[Unreleased]` queda así después del cambio:**
```markdown
## [Unreleased]

<!-- Cambios pendientes de release van aquí -->
```

**Agregar al final del archivo (antes del último salto de línea):**
```markdown
[1.0.0]: https://github.com/Camionerou/rag-saldivia/releases/tag/v1.0.0
```

- [ ] Copiar todo el contenido de `## [Unreleased]` (desde línea 12 hasta la próxima sección `##`)
- [ ] Reemplazar `## [Unreleased]` con el bloque vacío documentado arriba
- [ ] Crear sección `## [1.0.0] — 2026-03-XX` con el bloque de highlights + resumen + el contenido copiado
- [ ] Completar la fecha actual (2026-03-XX → fecha real del día del release)
- [ ] Agregar el link `[1.0.0]: https://...` al final del archivo
- [ ] Verificar que el formato es válido Keep a Changelog: `## [version] — YYYY-MM-DD`

**Estado: pendiente**

---

## R3 — .editorconfig *(10 min)*

Objetivo: cualquier editor (VS Code, Neovim, WebStorm, Cursor) formatea automáticamente de la misma manera sin configuración adicional.

**Archivo:** `.editorconfig` (nuevo en el root)

**Configuración:**
- `indent_style = space`
- `indent_size = 2`
- `end_of_line = lf`
- `charset = utf-8`
- `trim_trailing_whitespace = true`
- `insert_final_newline = true`
- Override para Markdown: `trim_trailing_whitespace = false` (los trailing spaces son saltos de línea en MD)

**Criterio de done:**
- `.editorconfig` existe en el root
- Cursor/VS Code lo detecta automáticamente

- [ ] Crear `.editorconfig`
- [ ] Verificar que Cursor lo detecta (debería aparecer en la barra de estado)

**Estado: pendiente**

---

## R4 — Dependabot *(15 min)*

Objetivo: el repo recibe PRs automáticos semanales cuando hay actualizaciones de dependencias disponibles.

**Archivo:** `.github/dependabot.yml` (nuevo)

**Configuración:**
- `package-ecosystem: npm` — para Bun workspaces (dependabot soporta npm lockfiles que Bun genera)
- `directory: "/"` — monorepo root
- `schedule: weekly` — los viernes, para revisar el lunes
- `open-pull-requests-limit: 5` — no abrumar con PRs
- `package-ecosystem: github-actions` — para mantener las actions actualizadas

**Criterio de done:**
- `.github/dependabot.yml` existe
- GitHub muestra "Dependabot enabled" en el tab Security

- [ ] Crear `.github/dependabot.yml`
- [ ] Verificar en GitHub que Dependabot está habilitado

**Estado: pendiente**

---

## R5 — GitHub Release Action mejorada *(30 min)*

Objetivo: cuando se hace un push del tag `v1.0.0`, GitHub Actions crea automáticamente una GitHub Release con el contenido del CHANGELOG.

**Archivo:** `.github/workflows/release.yml` (ya existe, mejorar)

**Mejoras al `release.yml` actual:**
- Usar `actions/create-release` o `softprops/action-gh-release` para crear la release
- Extraer automáticamente el contenido de la sección `[1.0.0]` del CHANGELOG.md como body de la release
- Adjuntar como asset el bundle de la CLI (`apps/cli/` compilado) si aplica
- Trigger: `on: push: tags: ['v*.*.*']`

**Criterio de done:**
- El workflow actualizado está en `release.yml`
- Hacer un push del tag crea la release automáticamente con el contenido del CHANGELOG

- [ ] Actualizar `.github/workflows/release.yml`
- [ ] Verificar con un dry-run o con el tag de test que funciona

**Estado: pendiente**

---

## R6 — Verificación final antes del tag *(30-45 min)*

Objetivo: confirmar que absolutamente todo está en orden antes de crear el tag `v1.0.0`.

**Checklist de verificación:**

- [ ] `bun run test` → todos pasan (exit 0)
- [ ] `bun run test:components` → 154 pass
- [ ] `bun run test:visual` → 22 pass
- [ ] `bun run test:a11y` → 0 violaciones
- [ ] `bun run test:e2e` → todos pass
- [ ] `cd apps/web && bunx tsc --noEmit` → exit 0
- [ ] `cd packages/db && bunx tsc --noEmit` → exit 0
- [ ] `bunx knip` → exit 0
- [ ] `cd apps/web && bunx eslint src --max-warnings 0` → exit 0
- [ ] `git status --short` → sin archivos modificados o no trackeados que no deban estar
- [ ] `git log --oneline -5` → el commit más reciente es el del Plan 12
- [ ] `grep -r '"version"' apps/web/package.json` → `"1.0.0"`
- [ ] `cat CHANGELOG.md | head -10` → sección `[Unreleased]` vacía, `[1.0.0]` presente
- [ ] `cat LICENSE` → MIT License existe
- [ ] `cat README.md | wc -l` → ≥ 300 líneas
- [ ] `ls .github/ISSUE_TEMPLATE/` → 2 templates
- [ ] `ls packages/db/README.md packages/logger/README.md` → existen
- [ ] `ls docs/api.md` → existe
- [ ] El CI del último push pasó completamente en GitHub

**Estado: pendiente**

---

## R7 — Git tag + push + GitHub Release *(15 min)*

Objetivo: el tag `v1.0.0` existe en el remoto y la GitHub Release está publicada.

**Proceso:**

1. Commit final del Plan 12:
   `chore(release): bump version a 1.0.0 — editorconfig, dependabot, changelog — plan12`

2. Push del branch:
   `git push origin experimental/ultra-optimize`

3. Verificar que el CI pasa en el remoto

4. Crear el tag anotado:
   `git tag -a v1.0.0 -m "Release 1.0.0 — primer release del stack TypeScript"`

5. Push del tag:
   `git push origin v1.0.0`

6. Verificar en GitHub que la Release Action corrió y creó la GitHub Release

7. **Decisión sobre merge a main:** el stack Python en `main` sigue siendo el deploy activo en producción. Esta branch (`experimental/ultra-optimize`) puede mergear a main cuando la workstation esté disponible para testear el stack completo con RAG real. Por ahora, `experimental/ultra-optimize` es el branch oficial del stack TypeScript — la GitHub Release lo refleja.

**Criterio de done:**
- `git tag | grep v1.0.0` → existe
- `git push origin v1.0.0` → exit 0
- GitHub muestra la Release v1.0.0 con el contenido del CHANGELOG
- El badge de versión en el README muestra `1.0.0`

- [ ] Commit final del Plan 12
- [ ] `git push origin experimental/ultra-optimize`
- [ ] Verificar CI en GitHub → pasa
- [ ] `git tag -a v1.0.0 -m "Release 1.0.0 — primer release del stack TypeScript"`
- [ ] `git push origin v1.0.0`
- [ ] Verificar GitHub Release creada automáticamente
- [ ] Verificar badge de versión en README

**Estado: pendiente**

---

## Criterio de done global del Plan 12

- Todos los `package.json` dicen `"version": "1.0.0"`
- CHANGELOG tiene sección `[1.0.0]` completa, `[Unreleased]` vacía
- `.editorconfig` existe
- `.github/dependabot.yml` existe
- `release.yml` mejorado
- Tag `v1.0.0` en el remoto
- GitHub Release publicada con contenido del CHANGELOG
- CI completamente verde en el tag

### Checklist de cierre

- [ ] 7 package.json en `1.0.0`
- [ ] CHANGELOG `[1.0.0]` completo
- [ ] `.editorconfig` creado
- [ ] `dependabot.yml` creado
- [ ] `release.yml` mejorado
- [ ] Verificación final R6 completada (todos los checkboxes)
- [ ] Tag `v1.0.0` pusheado
- [ ] GitHub Release publicada

**Estado: pendiente**

---

## Resumen del journey — Plan 1 a 1.0.0

| Plan | Tema | Hito |
|---|---|---|
| Plan 1 | Birth del monorepo TS | Stack técnico base |
| Plan 2 | Testing sistemático | Primera suite de tests |
| Plan 3 | Bugfix CodeGraphContext | Estabilización |
| Plan 4 | Product roadmap (Fases 0–2) | 24 páginas + design system base |
| Plan 5 | Testing foundation 95% | 270 tests en verde |
| Plan 6 | UI Testing Suite | Visual regression + a11y |
| Plan 7 | Design System "Warm Intelligence" | 24 páginas con paleta crema-navy |
| Plan 8 | Optimización + Redis + BullMQ | Código production-grade |
| Plan 9 | Repo limpio | Cero dead code, TypeScript perfecto, linting |
| Plan 10 | Testing completo | E2E, cobertura, smoke Redis |
| Plan 11 | Documentación perfecta | README, CONTRIBUTING, API docs, ER diagram |
| **Plan 12** | **Release 1.0.0** | **El primer release público** |
