# Plan N — [Título]

> **Branch:** 1.0.x
> **Prerequisito:** [plan anterior o decisión necesaria]
> **Sprint:** think → plan → execute → review → test → ship
> **Intensity:** quick | standard | thorough

## Contexto

[Qué problema resuelve, por qué ahora, qué decidió Enzo]

## Scope

**Archivos planeados:** [lista EXACTA]
**Tests planeados:** [qué tests se agregan o modifican]
**Fuera de scope:** [qué NO se toca — crítico para scope drift]

## Fases

### Fase N: [Título]

**Archivos:** [exactos]
**Cambios:**
- [cambio concreto con código si es necesario]

**Verificación:**
- [ ] `bunx tsc --noEmit` → 0 errors
- [ ] `bun run test` → green
- [ ] [verificación específica]

**Commit:** `tipo(scope): descripción — planN fN`

---

## Checklist de scope drift

Después de cada fase:
- [ ] Solo se tocaron archivos planeados
- [ ] No se introdujeron dependencias no planeadas
- [ ] Los tests cubren los cambios
- [ ] No se agregaron features fuera del scope

## Artifact

Guardar en `docs/artifacts/planN-fN-tipo.md`
