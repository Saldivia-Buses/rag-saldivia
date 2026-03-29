# Plan N — [Título]

> **Branch:** 1.0.x
> **Prerequisito:** [plan anterior o decisión necesaria]
> **Sprint:** think → plan → execute → review → test → ship
> **Intensity:** quick | standard | thorough

## Contexto

[Qué problema resuelve, por qué ahora, qué decidió Enzo]

## Scope

**Archivos planeados:** [lista EXACTA — contar con Grep antes de escribir]
**Tests planeados:** [qué tests se agregan o modifican]
**Fuera de scope:** [qué NO se toca — crítico para scope drift]

## Fases

### Fase 1: [Título]

**Archivos:** [paths exactos]
**Cambios:**
- [cambio concreto con código si es necesario]

**Verificación:**
- [ ] `bunx tsc --noEmit` → 0 errors
- [ ] `bun run test` → green
- [ ] [verificación específica de la fase]

**Commit:** `tipo(scope): descripción — planN f1`

### Fase 2: [Título]

[misma estructura]

---

## Checklist de scope drift

Después de cada fase:
- [ ] Solo se tocaron archivos planeados
- [ ] No se introdujeron dependencias no planeadas
- [ ] Los tests cubren los cambios
- [ ] No se agregaron features fuera del scope

## Artifact

Guardar en `docs/artifacts/planN-fN-tipo.md`

---

## Ejemplo: cómo se ve un plan real

```markdown
# Plan 15 — UI token extraction + branding

> **Branch:** 1.0.x
> **Prerequisito:** Plan 13
> **Intensity:** standard

## Contexto
Extraer tokens de claude.ai y aplicarlos con acento azure blue.
Enzo decidió: NO clonar literal, solo extraer tokens.

## Scope
**Archivos planeados:**
- apps/web/src/app/globals.css
- apps/web/src/components/ui/Button.tsx (ajustar variantes)
- apps/web/src/components/ui/Badge.tsx (ajustar variantes)
**Tests:** visual regression update
**Fuera de scope:** dark mode (Plan 20), layout changes, nuevos componentes

## Fases

### Fase 1: extraer tokens de claude.ai
**Archivos:** docs/artifacts/plan15-f1-tokens.md
**Cambios:** correr reconnaissance, documentar tokens extraídos
**Commit:** `docs(ui): extract claude.ai design tokens — plan15 f1`

### Fase 2: aplicar tokens a globals.css
**Archivos:** apps/web/src/app/globals.css
**Cambios:** reemplazar tokens actuales con los extraídos, acento azure
**Verificación:**
- [ ] `bunx tsc --noEmit` → 0 errors
- [ ] `bun run dev` → UI refleja nuevos tokens
**Commit:** `style(ui): apply claude azure tokens — plan15 f2`
```
