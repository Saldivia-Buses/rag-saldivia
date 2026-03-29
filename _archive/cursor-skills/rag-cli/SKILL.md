---
name: rag-cli
description: Use the RAG Saldivia CLI to manage users, collections, ingestion, audit logs, and database. Use when asked about CLI commands, running rag users/collections/ingest/audit/config/db/status, or when the user says "cómo uso la CLI", "listar usuarios", "ver logs", "rag status", or wants to perform admin operations from the terminal.
---

# RAG Saldivia — CLI `rag`

Reference: `docs/cli.md` para la referencia completa con flags y output de ejemplo.

## Instalación

```bash
cd apps/cli && bun link   # instala como comando global "rag"
```

La CLI necesita en el entorno:
- `RAG_WEB_URL` — URL del servidor Next.js (default: `http://localhost:3000`)
- `SYSTEM_API_KEY` — clave de auth service-to-service

## Comandos principales

| Comando | Descripción |
|---------|-------------|
| `rag status` | Health check de todos los servicios con latencias |
| `rag users list` | Tabla de usuarios con roles y áreas |
| `rag users create` | Wizard interactivo de creación |
| `rag users delete <id>` | Con confirmación |
| `rag collections list` | Lista colecciones del RAG Server con stats |
| `rag collections create` | Crear colección |
| `rag collections delete` | Con confirmación si tiene documentos |
| `rag ingest start` | Iniciar ingesta (interactivo o con flags) |
| `rag ingest status` | Jobs con barra de progreso |
| `rag ingest cancel <jobId>` | Cancelar job |
| `rag config get` | Parámetros RAG actuales |
| `rag config set <key> <val>` | Cambiar parámetro |
| `rag config reset` | Resetear a defaults |
| `rag audit log` | Últimos 50 eventos del sistema |
| `rag audit log -l ERROR` | Filtrar por nivel |
| `rag audit log -t rag.query` | Filtrar por tipo |
| `rag audit replay <fecha>` | Black box replay desde fecha |
| `rag audit export` | Exportar todos los eventos a JSON |
| `rag db migrate` | Aplicar migraciones pendientes |
| `rag db seed` | Seed de datos de desarrollo |
| `rag db reset` | BORRAR DB + rehacer seed (destructivo) |
| `rag setup` | Onboarding completo |

## Modo interactivo

Sin argumentos, `rag` abre un menú interactivo con las operaciones más comunes.
