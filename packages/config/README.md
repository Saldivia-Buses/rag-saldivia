# `@rag-saldivia/config`

Carga configuración desde YAML bajo `config/` (perfiles por `RAG_PROFILE`) y **sobrescribe con variables de entorno**. Valida con **Zod** (`AppConfigSchema`).

## Variables de entorno relevantes

| Variable | Requerida | Uso |
|----------|-----------|-----|
| `JWT_SECRET` | Sí en producción | Firma de JWT (mín. 16 caracteres en schema) |
| `SYSTEM_API_KEY` | Sí en producción | API de servicio (mín. 8 caracteres) |
| `DATABASE_PATH` | No | Default `./data/app.db` |
| `RAG_SERVER_URL` | No | Default `http://localhost:8081` |
| `MOCK_RAG` | No | `true`/`1` desactiva llamadas reales al RAG en capas que lo respetan |
| `RAG_PROFILE` | No | Nombre del YAML en `config/profiles/` |
| `PORT` | No | Default 3000 |
| `LOG_LEVEL` | No | `TRACE` … `FATAL` |
| `CORS_ORIGINS` | No | Lista separada por comas |

En desarrollo, el loader puede aplicar defaults inseguros si faltan secretos — **no** usar esos defaults en producción.

## Añadir una variable nueva

1. Extender `AppConfigSchema` en `src/loader.ts`.
2. Mapear `process.env["NOMBRE"]` en el objeto `raw` dentro de `loadConfig()`.
3. Documentar en `.env.example`.
4. Si afecta seguridad, mencionar en `SECURITY.md` o ADR.

## Parámetros RAG (admin)

`loadRagParams()` / `saveRagParams()` leen y escriben overrides en `config/admin-overrides.yaml` (tipado con `RagParamsSchema`).

## API pública

```typescript
import { loadConfig, loadRagParams, saveRagParams, AppConfigSchema } from "@rag-saldivia/config"
```
