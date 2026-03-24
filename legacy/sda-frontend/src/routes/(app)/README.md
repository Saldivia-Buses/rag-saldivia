# (app) Routes

Auth-guarded application routes. All routes under this group require a valid JWT session cookie.

## Authentication

The layout file `+layout.server.ts` validates the JWT cookie on every request via `verifySession(cookies)`. If the JWT is invalid or expired, the user is redirected to `/login`.

## Files

| Route | Description |
|-------|-------------|
| `chat/` | New chat page. Shows empty state with "Start a new conversation" prompt. |
| `chat/[id]/` | Specific chat session. Displays message history, streaming responses, crossdoc progress, and sources panel. |
| `collections/` | List all collections. Provides search/filter, create, and delete actions. |
| `collections/[name]/` | Single collection detail page. Shows collection stats, document count, and ingestion status. |
| `admin/users/` | User management (admin only). CRUD operations on users, role assignment, area assignment. |
| `admin/areas/` | Area management (admin only). Create, edit, delete areas. |
| `admin/permissions/` | Permissions management (admin only). Assign role-based permissions, manage access rules. |
| `admin/rag-config/` | RAG configuration (admin only). Adjust embedding model, LLM parameters, top-K settings, reranker config. |
| `admin/system/` | System status dashboard (admin only). Shows service health, uptime, resource usage (GPU, memory, disk). |
| `audit/` | Audit log viewer. Shows timestamped log of user actions, API calls, and system events. Filterable by user, action type, date range. |
| `settings/` | User settings. Profile info, password change, API key management. |
| `upload/` | File upload interface. Drag-and-drop or file picker for document ingestion. Shows upload progress, OCR status, and ingestion queue. |

## Admin-only routes

Routes under `admin/*` are only accessible to users with `role = 'admin'`. The UI conditionally renders admin navigation items based on the user's role.

Non-admin users attempting to access admin routes will see a 403 Forbidden error or be redirected (depending on implementation).
