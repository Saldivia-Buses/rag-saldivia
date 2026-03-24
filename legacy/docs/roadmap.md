# RAG Saldivia — Roadmap

> Fuente de verdad del estado del proyecto. Se actualiza solo cuando Enzo lo pide.
> Última actualización: 2026-03-24 (Fase 10 completada)

---

## ✅ Completadas

| Fase | Descripción | Fecha |
|------|-------------|-------|
| 1 | Fundación SvelteKit 5 BFF — auth, rutas base, JWT cookie | 2026-03-18 |
| 2 | Chat Pro — SSE streaming, historial, assistant bubbles | 2026-03-19 |
| 3 | CI/CD — GitHub Actions + SSH deploy + health check | 2026-03-19 |
| 4 | Colecciones Pro + Upload básico — CRUD UI, drag & drop | 2026-03-19 |
| 5 | Crossdoc Pro — pipeline 4 fases, subqueries paralelas | 2026-03-19 |
| 5.1 | Documentación — architecture.md, deployment.md, READMEs | 2026-03-19 |
| 5.2 | Tests — cobertura >80% gateway, auth, config, providers | 2026-03-19 |
| 5.3 | Bugfix — fixtures conftest, type errors, test isolation | 2026-03-19 |
| 5.5 | Workstation — perfil 1-GPU, workstation física Ubuntu 24.04 | 2026-03-23 |
| 5.6 | Hardening — rate limiting, upload limit, CORS, sanitización | 2026-03-23 |
| 6 | Upload inteligente — tier system, deadlock detection, resume | 2026-03-23 |
| 7 | Chat Sesiones Pro — rename, feedback, follow-ups, export | 2026-03-23 |
| 8 | Settings Pro — preferencias persistidas, perfil, contraseña, notificaciones | 2026-03-23 |
| 9 | Admin Pro — CRUD areas, permissions multi-área, system stats | 2026-03-24 |
| 10 | Admin RAG Config — configuración runtime de parámetros RAG, persistencia vía admin-overrides.yaml | 2026-03-24 |

---

## 🔄 En progreso

_Nada en progreso actualmente._

---

## 📋 Próximas (priorizadas)

| Prioridad | Fase | Descripción | Depende de |
|-----------|------|-------------|-----------|
| 1 | 11 | Audit Log Pro — filtros, export, visualización | 9 |

---

## 💡 Backlog (sin priorizar)

- Multi-tenant — múltiples organizaciones con aislamiento
- Offline mode — cache local para queries frecuentes
- Analytics — métricas de uso por usuario y colección
- Mobile — PWA responsive para acceso desde móvil
- API pública — endpoints REST documentados para integraciones externas
- Webhooks — notificaciones cuando ingesta completa

---

## 🚫 Descartadas

| Qué | Por qué |
|-----|---------|
| Brev cloud como deploy principal | Workstation física tiene más VRAM (96GB) y sin costo por hora |
| Fork del NVIDIA Blueprint | Mantener como submodule es más limpio para recibir updates |

---

## Dependencias entre fases

```
Fase 6 (Upload inteligente)
  └── Fase 7 (Chat Sesiones Pro)
        └── Fase 8 (Settings Pro)
        └── Fase 9 (Admin Pro)
              └── Fase 10 (Admin RAG Config) ✅
              └── Fase 11 (Audit Log Pro)
```
