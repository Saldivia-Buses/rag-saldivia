# 03 — Base de Datos

## Tecnologia

- **Motor:** SQLite via libsql (ADR-001)
- **ORM:** Drizzle 0.45
- **Cliente:** @libsql/client
- **Migraciones:** Drizzle Kit
- **Timestamps:** epoch ms con `Date.now()` (ADR-004, nunca `_ts()` de SQLite)
- **Formato de paquete:** CJS (ADR-002, ESM rompe webpack)

**Justificacion de SQLite:** suficiente para single-tenant. Drizzle facilita migrar a Postgres en 1 dia si escala.

---

## Schema completo

### Modulo: Core (`schema/core.ts`)

#### `areas`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| name | text | NOT NULL, UNIQUE |
| description | text | NOT NULL, default "" |
| createdAt | integer | NOT NULL (epoch ms) |

#### `users`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| email | text | NOT NULL, UNIQUE |
| name | text | NOT NULL |
| role | text | NOT NULL, enum: admin/area_manager/user, default "user" |
| apiKeyHash | text | NOT NULL (indice idx_users_api_key) |
| passwordHash | text | nullable |
| preferences | text (JSON) | NOT NULL, default {} |
| active | integer (boolean) | NOT NULL, default true |
| onboardingCompleted | integer (boolean) | NOT NULL, default false |
| ssoProvider | text | nullable ("google" / "azure") |
| ssoSubject | text | nullable (external user ID) |
| createdAt | integer | NOT NULL (epoch ms) |
| lastLogin | integer | nullable (epoch ms) |
| lastSeen | integer | nullable (epoch ms, presencia) |

#### `user_areas` (many-to-many)
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| userId | integer | FK → users.id (cascade) |
| areaId | integer | FK → areas.id (cascade) |
| | | PK: (userId, areaId) |

#### `area_collections`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| areaId | integer | FK → areas.id (cascade) |
| collectionName | text | NOT NULL |
| permission | text | enum: read/write/admin, default "read" |
| | | PK: (areaId, collectionName) |

#### `audit_log`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| userId | integer | FK → users.id |
| action | text | NOT NULL |
| collection | text | nullable |
| queryPreview | text | nullable |
| ipAddress | text | NOT NULL, default "" |
| timestamp | integer | NOT NULL (epoch ms) |
| | | Indices: idx_audit_user, idx_audit_timestamp |

#### `user_memory`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| userId | integer | FK → users.id (cascade) |
| key | text | NOT NULL |
| value | text | NOT NULL |
| source | text | enum: explicit/inferred, default "explicit" |
| createdAt | integer | NOT NULL |
| updatedAt | integer | NOT NULL |
| | | Unique index: (userId, key) |

#### `roles` (RBAC)
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| name | text | NOT NULL, UNIQUE |
| description | text | NOT NULL, default "" |
| level | integer | NOT NULL, default 0 (mayor = mas poderoso) |
| color | text | NOT NULL, default "#6e6c69" |
| icon | text | NOT NULL, default "user" (lucide) |
| isSystem | integer (boolean) | NOT NULL, default false |
| createdAt | integer | NOT NULL |

#### `permissions` (RBAC)
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| key | text | NOT NULL, UNIQUE (ej: "users.manage") |
| label | text | NOT NULL (ej: "Gestionar usuarios") |
| category | text | NOT NULL (ej: "Usuarios") |
| description | text | NOT NULL, default "" |

#### `role_permissions` (many-to-many)
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| roleId | integer | FK → roles.id (cascade) |
| permissionId | integer | FK → permissions.id (cascade) |
| | | PK: (roleId, permissionId) |

#### `user_role_assignments` (many-to-many)
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| userId | integer | FK → users.id (cascade) |
| roleId | integer | FK → roles.id (cascade) |
| assignedAt | integer | NOT NULL |
| | | PK: (userId, roleId) |

#### `rate_limits`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| targetType | text | enum: user/area |
| targetId | integer | NOT NULL |
| maxQueriesPerHour | integer | NOT NULL |
| active | integer (boolean) | NOT NULL, default true |
| createdAt | integer | NOT NULL |
| | | Indice: (targetType, targetId) |

#### `bot_user_mappings`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| platform | text | enum: slack/teams |
| externalUserId | text | NOT NULL |
| systemUserId | integer | FK → users.id (cascade) |
| createdAt | integer | NOT NULL |
| | | Unique: (platform, externalUserId) |

#### `external_sources`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK (UUID) |
| userId | integer | FK → users.id (cascade) |
| provider | text | enum: google_drive/sharepoint/confluence |
| name | text | NOT NULL |
| credentials | text | NOT NULL, default "{}" (JSON cifrado) |
| collectionDest | text | NOT NULL |
| schedule | text | enum: hourly/daily/weekly, default "daily" |
| active | integer (boolean) | NOT NULL, default true |
| lastSync | integer | nullable |
| createdAt | integer | NOT NULL |

---

### Modulo: Chat (`schema/chat.ts`)

#### `chat_sessions`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK (UUID) |
| userId | integer | FK → users.id (cascade) |
| title | text | NOT NULL |
| collection | text | NOT NULL |
| crossdoc | integer (boolean) | NOT NULL, default false |
| forkedFrom | text | nullable (sin FK constraint, self-reference) |
| createdAt | integer | NOT NULL |
| updatedAt | integer | NOT NULL |
| | | Indices: idx_user, idx_user_updated |

#### `chat_messages`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| sessionId | text | FK → chat_sessions.id (cascade) |
| role | text | enum: user/assistant/system |
| content | text | NOT NULL |
| sources | text (JSON) | nullable, array de citations |
| timestamp | integer | NOT NULL |

#### `message_feedback`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| messageId | integer | FK → chat_messages.id (cascade) |
| userId | integer | FK → users.id |
| rating | text | enum: up/down |
| createdAt | integer | NOT NULL |
| | | Unique: (messageId, userId) |

#### `session_shares`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK (UUID) |
| sessionId | text | FK → chat_sessions.id (cascade) |
| userId | integer | FK → users.id (cascade) |
| token | text | NOT NULL, UNIQUE (64-char hex) |
| expiresAt | integer | NOT NULL |
| createdAt | integer | NOT NULL |

#### `session_tags`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| sessionId | text | FK → chat_sessions.id (cascade) |
| tag | text | NOT NULL |
| | | PK: (sessionId, tag) |

#### `annotations`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| userId | integer | FK → users.id (cascade) |
| sessionId | text | FK → chat_sessions.id (cascade) |
| messageId | integer | FK → chat_messages.id (set null) |
| selectedText | text | NOT NULL |
| note | text | nullable |
| createdAt | integer | NOT NULL |

#### `saved_responses`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| userId | integer | FK → users.id (cascade) |
| messageId | integer | FK → chat_messages.id (set null) |
| content | text | NOT NULL |
| sessionTitle | text | nullable |
| createdAt | integer | NOT NULL |

#### `prompt_templates`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | integer | PK, autoincrement |
| title | text | NOT NULL |
| prompt | text | NOT NULL |
| focusMode | text | NOT NULL, default "detallado" |
| createdBy | integer | FK → users.id (cascade) |
| active | integer (boolean) | NOT NULL, default true |
| createdAt | integer | NOT NULL |

#### `projects`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK (UUID) |
| userId | integer | FK → users.id (cascade) |
| name | text | NOT NULL |
| description | text | NOT NULL, default "" |
| instructions | text | NOT NULL, default "" |
| createdAt | integer | NOT NULL |
| updatedAt | integer | NOT NULL |

#### `project_sessions` y `project_collections`
Tablas de relacion many-to-many para proyectos.

---

### Modulo: Messaging (`schema/messaging.ts`)

#### `channels`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK |
| type | text | enum: public/private/dm/group_dm |
| name | text | nullable |
| description | text | nullable |
| topic | text | nullable |
| createdBy | integer | FK → users.id |
| createdAt | integer | NOT NULL |
| updatedAt | integer | NOT NULL |
| archivedAt | integer | nullable |

#### `channel_members`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| channelId | text | FK → channels.id (cascade) |
| userId | integer | FK → users.id (cascade) |
| role | text | enum: owner/admin/member, default "member" |
| lastReadAt | integer | NOT NULL |
| muted | integer (boolean) | NOT NULL, default false |
| joinedAt | integer | NOT NULL |
| | | PK: (channelId, userId) |

#### `msg_messages`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK |
| channelId | text | FK → channels.id (cascade) |
| userId | integer | FK → users.id (cascade) |
| parentId | text | nullable (thread parent) |
| content | text | NOT NULL |
| type | text | enum: text/system/file, default "text" |
| replyCount | integer | NOT NULL, default 0 |
| lastReplyAt | integer | nullable |
| editedAt | integer | nullable |
| deletedAt | integer | nullable (soft delete) |
| metadata | text (JSON) | nullable |
| createdAt | integer | NOT NULL |
| | | Indices: (channelId, createdAt), (parentId) |

#### `msg_reactions`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| messageId | text | FK → msg_messages.id (cascade) |
| userId | integer | FK → users.id (cascade) |
| emoji | text | NOT NULL |
| createdAt | integer | NOT NULL |
| | | PK: (messageId, userId, emoji) |

#### `msg_mentions`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| id | text | PK |
| messageId | text | FK → msg_messages.id (cascade) |
| userId | integer | nullable |
| type | text | enum: user/channel/everyone |

#### `pinned_messages`
| Columna | Tipo | Restricciones |
|---------|------|--------------|
| channelId | text | FK → channels.id (cascade) |
| messageId | text | FK → msg_messages.id (cascade) |
| pinnedBy | integer | FK → users.id |
| pinnedAt | integer | NOT NULL |
| | | PK: (channelId, messageId) |

---

### Modulo: Events (`schema/events.ts`)

Tabla unica `events` para audit trail completo del sistema.

---

## Modulos de queries (21)

| Archivo | Funciones principales |
|---------|----------------------|
| `users.ts` | getUserById, createUser, updateUser, verifyPassword, getUsersByRole |
| `areas.ts` | getAreaById, createArea, updateArea, deleteArea, areaCollections |
| `sessions.ts` | createSession, getSessionWithMessages, listSessions |
| `messages.ts` | addMessage, listMessages |
| `events.ts` | logEvent, listEvents, searchEvents, getEventsByUser |
| `saved.ts` | toggleSaved, getSavedResponses |
| `annotations.ts` | saveAnnotation, getAnnotations |
| `tags.ts` | createTag, addTagToSession, listTags |
| `shares.ts` | createShareToken, getShareToken, incrementShareViewCount |
| `templates.ts` | createTemplate, listTemplates, deleteTemplate |
| `collection-history.ts` | logCollectionUpdate, getCollectionHistory |
| `reports.ts` | createReport, listReports, getReportData |
| `rate-limits.ts` | checkRateLimit, incrementCounter, resetCounter |
| `webhooks.ts` | createWebhook, triggerWebhook, listWebhooks |
| `search.ts` | searchSessions, fullTextSearch, searchAcrossCollections |
| `projects.ts` | createProject, listProjects, getProject, addSessionToProject |
| `memory.ts` | addMemory, listMemory, deleteMemory |
| `external-sources.ts` | addExternalSource, syncExternalSource |
| `rbac.ts` | checkPermission, listPermissions, grantPermission |
| `channels.ts` | createChannel, listChannels, getChannel, addMember |
| `messaging.ts` | addChannelMessage, listChannelMessages, addReaction |

---

## Tipos exportados (Drizzle inferred)

```typescript
// Core
DbArea, NewArea, DbUser, NewUser, DbUserArea, DbAreaCollection,
DbUserMemory, DbRole, NewRole, DbPermission, DbRolePermission,
DbUserRoleAssignment, DbRateLimit, NewRateLimit, DbBotUserMapping,
DbExternalSource, NewExternalSource

// Chat
DbChatSession, NewChatSession, DbChatMessage, NewChatMessage,
DbSessionShare, DbSessionTag, DbAnnotation, NewAnnotation,
DbSavedResponse, NewSavedResponse, DbPromptTemplate, NewPromptTemplate,
DbProject, NewProject, DbProjectSession, DbProjectCollection

// Messaging
DbChannel, NewChannel, DbChannelMember, NewChannelMember,
DbMsgMessage, NewMsgMessage, DbMsgReaction, DbMsgMention,
NewMsgMention, DbPinnedMessage
```

---

## Conexion y Redis

```typescript
// packages/db/src/connection.ts
getDb()          // Retorna instancia Drizzle con SQLite
// packages/db/src/redis.ts
getRedisClient() // NUNCA retorna null — lanza error (ADR-010)
getBullMQConnection() // Redis con maxRetriesPerRequest: null para BullMQ
```

**Regla critica:** No importar logger en redis.ts — dependencia circular (ADR-005).

---

## Credenciales de desarrollo (seed)

| Email | Password | Rol |
|-------|----------|-----|
| admin@localhost | changeme | admin |
| user@localhost | test1234 | user |
