# 05 — API Routes y Server Actions

## API Routes (18 endpoints)

### Auth (3 routes)

| Ruta | Metodo | Descripcion | Auth |
|------|--------|-------------|------|
| `/api/auth/login` | POST | Autenticacion JWT, devuelve cookie HttpOnly | Publica |
| `/api/auth/logout` | DELETE | Revoca JWT en Redis, limpia cookie | Requerida |
| `/api/auth/refresh` | POST | Renueva JWT, revoca el viejo | Publica (token en cookie) |

### RAG (6 routes)

| Ruta | Metodo | Descripcion | Auth |
|------|--------|-------------|------|
| `/api/rag/generate` | POST | Proxy SSE al RAG Server (o OpenRouter mock) | Requerida |
| `/api/rag/collections` | GET | Lista colecciones de Milvus | Requerida |
| `/api/rag/collections` | POST | Crea coleccion en Milvus | Requerida |
| `/api/rag/collections/[name]` | DELETE | Elimina coleccion | Requerida |
| `/api/rag/collections/[name]/history` | GET | Historial de cambios de coleccion | Requerida |
| `/api/rag/document/[name]` | GET | Documento por nombre | Requerida |
| `/api/rag/suggest` | POST | Preguntas sugeridas/relacionadas | Requerida |

### Messaging (9 routes — Plan 25)

| Ruta | Metodo | Descripcion | Auth |
|------|--------|-------------|------|
| `/api/messaging/channels` | GET | Lista canales del usuario | Requerida |
| `/api/messaging/channels` | POST | Crea canal | Requerida |
| `/api/messaging/channels/[id]` | GET | Detalle de canal | Requerida |
| `/api/messaging/channels/[id]` | PUT | Actualiza canal | Requerida |
| `/api/messaging/channels/[id]` | DELETE | Elimina canal | Requerida |
| `/api/messaging/channels/[id]/members` | GET | Miembros del canal | Requerida |
| `/api/messaging/channels/[id]/members` | POST | Agrega miembro | Requerida |
| `/api/messaging/messages` | POST | Envia mensaje | Requerida |
| `/api/messaging/messages/[id]/pin` | POST | Pin/unpin mensaje | Requerida |
| `/api/messaging/messages/[id]/reactions` | POST | Agrega/quita reaccion | Requerida |
| `/api/messaging/search` | GET | Busca mensajes | Requerida |
| `/api/messaging/upload` | POST | Sube archivo adjunto | Requerida |

### Sistema (2 routes)

| Ruta | Metodo | Descripcion | Auth |
|------|--------|-------------|------|
| `/api/health` | GET | Health check | Publica |
| `/api/feedback` | POST | Error feedback (guarda en audit_log) — Plan 28 | Requerida |

---

## Server Actions (50+ en 10 archivos)

Todas usan `next-safe-action` con middleware `authAction` o `adminAction`.
Input validado con Zod. Retorno wrapped en `result?.data`.

### `actions/auth.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionLogout` | authAction | Invalida sesion |

### `actions/chat.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionCreateSession` | authAction | Crea nueva sesion de chat |
| `actionDeleteSession` | authAction | Elimina sesion |
| `actionRenameSession` | authAction | Renombra sesion |
| `actionAddMessage` | authAction | Agrega mensaje a sesion |
| `actionAddFeedback` | authAction | Feedback up/down a mensaje |
| `actionToggleSaved` | authAction | Guardar/desguardar respuesta |
| `actionForkSession` | authAction | Duplicar sesion |
| `actionSaveAnnotation` | authAction | Guardar anotacion |
| `actionCreateSessionForDoc` | authAction | Crear sesion desde documento |
| `actionAddTag` | authAction | Agregar tag a sesion |
| `actionRemoveTag` | authAction | Quitar tag de sesion |

### `actions/collections.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionCreateCollection` | adminAction | Crea coleccion en Milvus |
| `actionDeleteCollection` | adminAction | Elimina coleccion |

### `actions/config.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionUpdateRagParams` | adminAction | Actualiza parametros RAG |
| `actionResetRagParams` | adminAction | Reset a defaults |

### `actions/settings.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionUpdateProfile` | authAction | Actualiza nombre/email |
| `actionUpdatePassword` | authAction | Cambia password |
| `actionUpdatePreferences` | authAction | Actualiza preferencias |
| `actionCompleteOnboarding` | authAction | Marca onboarding completado |
| `actionAddMemory` | authAction | Agrega memoria personalizada |
| `actionDeleteMemory` | authAction | Elimina memoria |

### `actions/admin.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionListUsers` | adminAction | Lista usuarios |
| `actionCreateUser` | adminAction | Crea usuario |
| `actionUpdateUser` | adminAction | Actualiza usuario |
| `actionResetPassword` | adminAction | Reset password |
| `actionDeleteUser` | adminAction | Elimina usuario |

### `actions/roles.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionListRoles` | authAction + permission | Lista roles |
| `actionCreateRole` | authAction + permission | Crea rol |
| `actionUpdateRole` | authAction + permission | Actualiza rol |
| `actionDeleteRole` | authAction + permission | Elimina rol |
| `actionSetRolePermissions` | authAction + permission | Asigna permisos a rol |
| `actionSetUserRoles` | authAction + permission | Asigna roles a usuario |
| `actionGetRolePermissions` | authAction + permission | Obtiene permisos de rol |
| `actionListPermissions` | authAction + permission | Lista permisos disponibles |

### `actions/areas.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionCreateArea` | adminAction | Crea area |
| `actionUpdateArea` | adminAction | Actualiza area |
| `actionDeleteArea` | adminAction | Elimina area |
| `actionSetAreaCollections` | adminAction | Asigna colecciones a area |
| `actionAddUserToArea` | adminAction | Agrega usuario a area |
| `actionRemoveUserFromArea` | adminAction | Quita usuario de area |

### `actions/templates.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionListTemplates` | authAction | Lista templates de prompts |
| `actionCreateTemplate` | adminAction | Crea template |
| `actionDeleteTemplate` | adminAction | Elimina template |

### `actions/messaging.ts`
| Action | Middleware | Descripcion |
|--------|-----------|-------------|
| `actionSendMessage` | authAction | Enviar mensaje a canal |
| `actionEditMessage` | authAction | Editar mensaje |
| `actionDeleteMessage` | authAction | Eliminar mensaje |
| `actionCreateChannel` | authAction | Crear canal |
| `actionJoinChannel` | authAction | Unirse a canal |
| `actionLeaveChannel` | authAction | Salir de canal |
| `actionPinMessage` | authAction | Pinear mensaje |
| `actionUnpinMessage` | authAction | Des-pinear mensaje |
| `actionReactToMessage` | authAction | Agregar reaccion |
| `actionRemoveReaction` | authAction | Quitar reaccion |
| `actionMarkAsRead` | authAction | Marcar como leido |
