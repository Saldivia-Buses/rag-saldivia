---
name: git-worker
description: Worker para operaciones de git — commit, PR creation, changelog update. Usado para milestones de cierre de branch.
---

# Git Worker

NOTE: Startup and cleanup are handled by `worker-base`. This skill defines the WORK PROCEDURE.

## When to Use This Skill

Features que involucran commit de cambios existentes, creación de PRs, actualización de CHANGELOG, y merge de branches. No escriben código nuevo — solo cierre administrativo de trabajo ya completado.

## Required Skills

None.

## Work Procedure

1. **Verificar estado del repo** — ejecutar `git status` y `git diff --stat` para entender qué hay sin commitear. Confirmar que estás en el branch correcto.

2. **Ejecutar tests** — siempre antes de cualquier commit: `uv run pytest saldivia/tests/ -v 2>&1 | tail -20`. Si fallan, NO proceder al commit — retornar al orquestador.

3. **Revisar cambios para secrets** — `git diff --staged` o `git diff` para verificar que no hay API keys, JWT secrets, passwords, o credenciales en los cambios. Si se detectan, DETENER y retornar al orquestador.

4. **Actualizar CHANGELOG.md** — mover el contenido de `[Unreleased]` a una sección versionada con fecha:
   - Formato: `## [0.X.0] - YYYY-MM-DD`
   - Dejar `[Unreleased]` vacío

5. **Staging y commit** — stagear TODOS los archivos modificados y untracked relevantes. Commit con mensaje descriptivo siguiendo convención del repo (ver `git log --oneline -5` para el estilo).

6. **Crear PR** — usar `gh pr create` con título descriptivo, base `main`, y descripción que incluya qué se cierra.

7. **Verificar merge** — confirmar que el PR está creado correctamente. NO hacer merge automático a menos que la feature lo especifique explícitamente.

8. **Verificar tests post-merge** — si la feature requiere verificar tests en `main`, hacer checkout de main y correr pytest.

## Example Handoff

```json
{
  "salientSummary": "Commiteados 14 archivos de Fase 10 en fase10/rag-config-pro. Tests: 186 passed, 0 failed. PR #18 creado hacia main. CHANGELOG actualizado a [0.10.0] - 2026-03-24.",
  "whatWasImplemented": "Commit de todos los cambios sin staging de Fase 10 (ConfigSlider, ModelSelector, GuardrailsToggle, ProfileSwitcher, BFF routes /api/admin/config y /api/admin/profile, tests). CHANGELOG movido de [Unreleased] a [0.10.0]. PR #18 creado.",
  "whatWasLeftUndone": "",
  "verification": {
    "commandsRun": [
      {
        "command": "git status",
        "exitCode": 0,
        "observation": "14 archivos modificados, 6 untracked antes del commit"
      },
      {
        "command": "uv run pytest saldivia/tests/ -v 2>&1 | tail -5",
        "exitCode": 0,
        "observation": "186 passed, 0 failed"
      },
      {
        "command": "git add -A && git commit -m 'feat(fase10): admin RAG config — sliders, model selector, guardrails, profile switcher'",
        "exitCode": 0,
        "observation": "Commit sha abc1234"
      },
      {
        "command": "gh pr create --title 'feat(fase10): Admin RAG Config Pro' --base main --body '...'",
        "exitCode": 0,
        "observation": "PR #18 creado en https://github.com/Camionerou/rag-saldivia/pull/18"
      }
    ],
    "interactiveChecks": []
  },
  "tests": {
    "added": []
  },
  "discoveredIssues": []
}
```

## When to Return to Orchestrator

- Tests fallan antes del commit
- Se detectan secrets en los diffs
- Conflictos de merge que no son triviales de resolver
- El branch ya fue mergeado o PR ya existe
- `gh` CLI no disponible o sin autenticación
