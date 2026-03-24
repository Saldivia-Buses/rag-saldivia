---
name: backend-worker
description: Worker para cambios en el backend Python — gateway.py, database.py, models.py, tests. Sigue TDD estricto.
---

# Backend Worker

NOTE: Startup and cleanup are handled by `worker-base`. This skill defines the WORK PROCEDURE.

## When to Use This Skill

Features que modifican `saldivia/gateway.py`, `saldivia/auth/database.py`, `saldivia/auth/models.py`, `saldivia/config.py`, u otros módulos Python. Siempre incluye tests unitarios.

## Required Skills

None.

## Work Procedure

### Preparación

1. Leer los archivos existentes que se van a modificar antes de hacer cualquier cambio.
2. Revisar tests existentes relacionados con el área: `saldivia/tests/test_gateway_extended.py`, `test_config.py`, etc.
3. Identificar dependencias: ¿qué otros módulos usan lo que voy a cambiar?

### TDD — Red-Green

4. **Escribir tests primero (Red)** — en el archivo de test correspondiente, agregar los nuevos test cases ANTES de implementar. Correr `uv run pytest saldivia/tests/ -v -k "nuevo_test" 2>&1 | tail -20` para confirmar que FALLAN (red).

5. **Implementar (Green)** — modificar los archivos de producción para que los tests pasen. Correr `uv run pytest saldivia/tests/ -v 2>&1 | tail -30` para confirmar que pasan y que no rompiste tests existentes.

### Reglas críticas para este codebase

- **SQLite migrations:** Agregar columnas con `ALTER TABLE ... ADD COLUMN ... DEFAULT NULL` dentro de un try/except que silencia exactamente `sqlite3.OperationalError` con `"duplicate column name"` — no `Exception` genérico.
- **_ts() helper:** Usar siempre `_ts()` para timestamps en SQLite — `detect_types=PARSE_DECLTYPES` crashea con date-only strings.
- **LEFT JOIN vs INNER JOIN:** Para JOINs con tablas que pueden tener filas huérfanas (e.g., `audit_log` vs `users` eliminados), usar siempre LEFT JOIN.
- **JSON serialization:** Python `None` debe serializar como JSON `null`, no como `"None"` string. Verificar con `resp.json()["field"] is None` en tests.
- **log_action():** Siempre pasar `success=` explícitamente en todos los call sites.
- **Admin-only endpoints:** Verificar que el check de role usa `Role.ADMIN`, no `is_admin`.

### Verificación final

6. Correr suite completa: `uv run pytest saldivia/tests/ -v 2>&1 | tail -30`. **Todos los tests existentes deben seguir pasando.**
7. Verificar manualmente con curl el endpoint modificado (si aplica):
   ```bash
   # Ejemplo para audit endpoint
   TOKEN=$(curl -s -X POST http://localhost:9000/auth/login -d '...' | jq -r .token)
   curl -s "http://localhost:9000/admin/audit?offset=0&limit=5" -H "Authorization: Bearer $TOKEN" | jq .
   ```
8. Para cambios en database.py: verificar con SQLite in-memory que las migraciones son idempotentes.

## Example Handoff

```json
{
  "salientSummary": "Implementado offset+username en get_audit_log_filtered y GET /admin/audit. Migración DDL success column idempotente. log_action actualizado para aceptar success. 12 tests nuevos, todos passing. Suite completa: 192 passed, 0 failed.",
  "whatWasImplemented": "database.py: get_audit_log_filtered agrega LIMIT ? OFFSET ?, LEFT JOIN con users para username, SELECT incluye success. models.py: AuditEntry agrega username y success. gateway.py: GET /admin/audit acepta offset/limit(50), response incluye username y success. PATCH /admin/config, POST /admin/config/reset, POST /admin/profile llaman log_action con success=True. Migration DDL para columna success. BFF src/routes/api/audit/+server.ts con GET handler.",
  "whatWasLeftUndone": "",
  "verification": {
    "commandsRun": [
      {
        "command": "uv run pytest saldivia/tests/ -v 2>&1 | tail -15",
        "exitCode": 0,
        "observation": "192 passed, 0 failed in 14.2s"
      },
      {
        "command": "uv run pytest saldivia/tests/test_gateway_extended.py -v 2>&1 | tail -20",
        "exitCode": 0,
        "observation": "All 18 tests passing including new test_audit_offset_forwarded, test_audit_default_limit_is_50, test_audit_response_has_username_and_success"
      },
      {
        "command": "uv run pytest saldivia/tests/test_config.py -v 2>&1 | tail -10",
        "exitCode": 0,
        "observation": "8 passed — incluyendo test_audit_log_success_column_migration, test_log_action_with_success"
      }
    ],
    "interactiveChecks": []
  },
  "tests": {
    "added": [
      {
        "file": "saldivia/tests/test_gateway_extended.py",
        "cases": [
          {"name": "test_audit_offset_forwarded", "verifies": "offset param forwarded a DB con offset=15"},
          {"name": "test_audit_default_limit_is_50", "verifies": "sin ?limit=, DB llamado con limit=50"},
          {"name": "test_audit_response_has_username_and_success", "verifies": "JSON incluye username y success con tipos correctos"},
          {"name": "test_patch_admin_config_creates_audit_entry", "verifies": "PATCH /admin/config genera entry con success=True"},
          {"name": "test_post_admin_config_reset_creates_audit_entry", "verifies": "POST reset genera entry con success=True"},
          {"name": "test_post_admin_profile_creates_audit_entry", "verifies": "POST profile genera entry con success=True"},
          {"name": "test_get_admin_audit_non_admin_403", "verifies": "role=USER → 403"}
        ]
      },
      {
        "file": "saldivia/tests/test_auth.py",
        "cases": [
          {"name": "test_get_audit_log_filtered_with_offset", "verifies": "paginación offset correcta con in-memory SQLite"},
          {"name": "test_get_audit_log_filtered_username_join", "verifies": "username del JOIN en resultado"},
          {"name": "test_get_audit_log_filtered_orphaned_user_entry_survives", "verifies": "LEFT JOIN: user eliminado no borra sus audit entries"},
          {"name": "test_get_audit_log_filtered_success_column_retrieved", "verifies": "success=1 en DB → entry.success=True"},
          {"name": "test_audit_log_success_column_migration", "verifies": "migración DDL idempotente"},
          {"name": "test_log_action_with_success", "verifies": "log_action persiste success=True"}
        ]
      }
    ]
  },
  "discoveredIssues": []
}
```

## When to Return to Orchestrator

- Un test existente falla y no está relacionado con el feature actual
- La firma de una función que se necesita cambiar es usada en más de 5 lugares del codebase
- Se descubre que el schema de la DB requiere una migración con pérdida de datos
- El gateway está corriendo en producción y el cambio requiere restart
