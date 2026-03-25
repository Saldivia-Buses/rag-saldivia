/**
 * @rag-saldivia/db — Drizzle schema
 *
 * Fuente de verdad para la estructura de la base de datos SQLite.
 * Reemplaza la combinación de CREATE TABLE dispersos en saldivia/auth/database.py.
 *
 * Decisión de timestamps: se usan INTEGER (epoch ms) en lugar de TEXT/TIMESTAMP.
 * Esto elimina el bug de SQLite que causó el helper _ts() en el código Python,
 * y permite usar Temporal.Now.instant().epochMilliseconds directamente.
 */

import {
  sqliteTable,
  text,
  integer,
  primaryKey,
  uniqueIndex,
  index,
} from "drizzle-orm/sqlite-core"
import { relations } from "drizzle-orm"

// ── Areas ──────────────────────────────────────────────────────────────────

export const areas = sqliteTable("areas", {
  id: integer("id").primaryKey({ autoIncrement: true }),
  name: text("name").notNull().unique(),
  description: text("description").notNull().default(""),
  createdAt: integer("created_at").notNull(), // epoch ms
})

// ── Users ──────────────────────────────────────────────────────────────────

export const users = sqliteTable(
  "users",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    email: text("email").notNull().unique(),
    name: text("name").notNull(),
    role: text("role", { enum: ["admin", "area_manager", "user"] })
      .notNull()
      .default("user"),
    apiKeyHash: text("api_key_hash").notNull(),
    passwordHash: text("password_hash"),
    preferences: text("preferences", { mode: "json" })
      .$type<Record<string, unknown>>()
      .notNull()
      .default({}),
    active: integer("active", { mode: "boolean" }).notNull().default(true),
    createdAt: integer("created_at").notNull(), // epoch ms
    lastLogin: integer("last_login"), // epoch ms, nullable
  },
  (t) => ({
    apiKeyIdx: index("idx_users_api_key").on(t.apiKeyHash),
  })
)

// ── User ↔ Areas (many-to-many) ────────────────────────────────────────────

export const userAreas = sqliteTable(
  "user_areas",
  {
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    areaId: integer("area_id")
      .notNull()
      .references(() => areas.id, { onDelete: "cascade" }),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.userId, t.areaId] }),
  })
)

// ── Area ↔ Collections ─────────────────────────────────────────────────────

export const areaCollections = sqliteTable(
  "area_collections",
  {
    areaId: integer("area_id")
      .notNull()
      .references(() => areas.id, { onDelete: "cascade" }),
    collectionName: text("collection_name").notNull(),
    permission: text("permission", { enum: ["read", "write", "admin"] })
      .notNull()
      .default("read"),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.areaId, t.collectionName] }),
  })
)

// ── Audit Log ──────────────────────────────────────────────────────────────

export const auditLog = sqliteTable(
  "audit_log",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id),
    action: text("action").notNull(),
    collection: text("collection"),
    queryPreview: text("query_preview"),
    ipAddress: text("ip_address").notNull().default(""),
    timestamp: integer("timestamp").notNull(), // epoch ms
  },
  (t) => ({
    userIdx: index("idx_audit_user").on(t.userId),
    timestampIdx: index("idx_audit_timestamp").on(t.timestamp),
  })
)

// ── Chat Sessions ──────────────────────────────────────────────────────────

export const chatSessions = sqliteTable(
  "chat_sessions",
  {
    id: text("id").primaryKey(), // UUID
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    title: text("title").notNull(),
    collection: text("collection").notNull(),
    crossdoc: integer("crossdoc", { mode: "boolean" }).notNull().default(false),
    createdAt: integer("created_at").notNull(), // epoch ms
    updatedAt: integer("updated_at").notNull(), // epoch ms
  },
  (t) => ({
    userIdx: index("idx_chat_sessions_user").on(t.userId),
    userUpdatedIdx: index("idx_chat_sessions_user_updated").on(t.userId, t.updatedAt),
  })
)

// ── Chat Messages ──────────────────────────────────────────────────────────

export const chatMessages = sqliteTable(
  "chat_messages",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    sessionId: text("session_id")
      .notNull()
      .references(() => chatSessions.id, { onDelete: "cascade" }),
    role: text("role", { enum: ["user", "assistant", "system"] }).notNull(),
    content: text("content").notNull(),
    sources: text("sources", { mode: "json" }).$type<unknown[]>(),
    timestamp: integer("timestamp").notNull(), // epoch ms
  },
  (t) => ({
    sessionIdx: index("idx_chat_messages_session").on(t.sessionId),
  })
)

// ── Message Feedback ───────────────────────────────────────────────────────

export const messageFeedback = sqliteTable(
  "message_feedback",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    messageId: integer("message_id")
      .notNull()
      .references(() => chatMessages.id, { onDelete: "cascade" }),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id),
    rating: text("rating", { enum: ["up", "down"] }).notNull(),
    createdAt: integer("created_at").notNull(), // epoch ms
  },
  (t) => ({
    uniqueFeedback: uniqueIndex("idx_feedback_unique").on(t.messageId, t.userId),
    messageIdx: index("idx_message_feedback_message").on(t.messageId),
  })
)

// ── Session Tags ───────────────────────────────────────────────────────────

export const sessionTags = sqliteTable(
  "session_tags",
  {
    sessionId: text("session_id")
      .notNull()
      .references(() => chatSessions.id, { onDelete: "cascade" }),
    tag: text("tag").notNull(),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.sessionId, t.tag] }),
    tagIdx: index("idx_session_tags_tag").on(t.tag),
  })
)

// ── Annotations ────────────────────────────────────────────────────────────

export const annotations = sqliteTable(
  "annotations",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    sessionId: text("session_id")
      .notNull()
      .references(() => chatSessions.id, { onDelete: "cascade" }),
    messageId: integer("message_id")
      .references(() => chatMessages.id, { onDelete: "set null" }),
    selectedText: text("selected_text").notNull(),
    note: text("note"),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    userIdx: index("idx_annotations_user").on(t.userId),
    sessionIdx: index("idx_annotations_session").on(t.sessionId),
  })
)

// ── Saved Responses ────────────────────────────────────────────────────────

export const savedResponses = sqliteTable(
  "saved_responses",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    messageId: integer("message_id")
      .references(() => chatMessages.id, { onDelete: "set null" }),
    content: text("content").notNull(),
    sessionTitle: text("session_title"),
    createdAt: integer("created_at").notNull(), // epoch ms
  },
  (t) => ({
    userIdx: index("idx_saved_responses_user").on(t.userId),
  })
)

// ── Ingestion Jobs ─────────────────────────────────────────────────────────

export const ingestionJobs = sqliteTable(
  "ingestion_jobs",
  {
    id: text("id").primaryKey(), // UUID
    userId: integer("user_id")
      .notNull()
      .references(() => users.id),
    taskId: text("task_id").notNull(),
    filename: text("filename").notNull(),
    collection: text("collection").notNull(),
    tier: text("tier", { enum: ["tiny", "small", "medium", "large"] }).notNull(),
    pageCount: integer("page_count"),
    state: text("state", {
      enum: ["pending", "running", "stalled", "done", "error", "cancelled"],
    })
      .notNull()
      .default("pending"),
    progress: integer("progress").notNull().default(0),
    fileHash: text("file_hash"),
    retryCount: integer("retry_count").notNull().default(0),
    lastChecked: integer("last_checked"), // epoch ms
    createdAt: integer("created_at").notNull(), // epoch ms
    completedAt: integer("completed_at"), // epoch ms
  },
  (t) => ({
    userIdx: index("idx_ingestion_jobs_user").on(t.userId),
    stateIdx: index("idx_ingestion_jobs_state").on(t.state),
  })
)

// ── Ingestion Alerts ───────────────────────────────────────────────────────

export const ingestionAlerts = sqliteTable(
  "ingestion_alerts",
  {
    id: text("id").primaryKey(),
    jobId: text("job_id").notNull(),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id),
    filename: text("filename").notNull(),
    collection: text("collection").notNull(),
    tier: text("tier", { enum: ["tiny", "small", "medium", "large"] }).notNull(),
    pageCount: integer("page_count"),
    fileHash: text("file_hash"),
    error: text("error"),
    retryCount: integer("retry_count"),
    progressAtFailure: integer("progress_at_failure"),
    gatewayVersion: text("gateway_version"),
    createdAt: integer("created_at").notNull(), // epoch ms
    resolvedAt: integer("resolved_at"), // epoch ms
    resolvedBy: text("resolved_by"),
    notes: text("notes"),
  },
  (t) => ({
    resolvedIdx: index("idx_alerts_resolved").on(t.resolvedAt),
  })
)

// ── Ingestion Queue (reemplaza Redis) ─────────────────────────────────────
// Worker hace SELECT + UPDATE locked_at en una transacción.
// SQLite serializa writes → no hay race condition.

export const ingestionQueue = sqliteTable(
  "ingestion_queue",
  {
    id: text("id").primaryKey(), // UUID
    collection: text("collection").notNull(),
    filePath: text("file_path").notNull(),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id),
    priority: integer("priority").notNull().default(0),
    status: text("status", { enum: ["pending", "locked", "done", "error"] })
      .notNull()
      .default("pending"),
    lockedAt: integer("locked_at"), // epoch ms — null si no está bloqueado
    lockedBy: text("locked_by"), // worker instance id
    createdAt: integer("created_at").notNull(), // epoch ms
    startedAt: integer("started_at"), // epoch ms
    completedAt: integer("completed_at"), // epoch ms
    error: text("error"),
    retryCount: integer("retry_count").notNull().default(0),
  },
  (t) => ({
    statusIdx: index("idx_queue_status").on(t.status, t.priority),
    pendingIdx: index("idx_queue_pending").on(t.status, t.lockedAt),
  })
)

// ── Events (Black Box) ─────────────────────────────────────────────────────
// Registro inmutable de todos los eventos del sistema.
// Permite reconstruir el estado con packages/logger/blackbox.ts.

export const events = sqliteTable(
  "events",
  {
    id: text("id").primaryKey(), // UUID
    ts: integer("ts").notNull(), // epoch ms (Temporal.Now.instant().epochMilliseconds)
    source: text("source", { enum: ["frontend", "backend"] }).notNull(),
    level: text("level", {
      enum: ["TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"],
    }).notNull(),
    type: text("type").notNull(), // EventType
    userId: integer("user_id").references(() => users.id),
    sessionId: text("session_id"),
    payload: text("payload", { mode: "json" })
      .$type<Record<string, unknown>>()
      .notNull()
      .default({}),
    sequence: integer("sequence").notNull(), // auto-incremented monotónico para replay
  },
  (t) => ({
    tsIdx: index("idx_events_ts").on(t.ts),
    typeIdx: index("idx_events_type").on(t.type),
    userIdx: index("idx_events_user").on(t.userId),
    levelIdx: index("idx_events_level").on(t.level),
    sequenceIdx: index("idx_events_sequence").on(t.sequence),
  })
)

// ── Relations (necesarias para .query con `with`) ─────────────────────────

export const usersRelations = relations(users, ({ many }) => ({
  userAreas: many(userAreas),
  chatSessions: many(chatSessions),
}))

export const areasRelations = relations(areas, ({ many }) => ({
  userAreas: many(userAreas),
  areaCollections: many(areaCollections),
}))

export const userAreasRelations = relations(userAreas, ({ one }) => ({
  user: one(users, { fields: [userAreas.userId], references: [users.id] }),
  area: one(areas, { fields: [userAreas.areaId], references: [areas.id] }),
}))

export const areaCollectionsRelations = relations(areaCollections, ({ one }) => ({
  area: one(areas, { fields: [areaCollections.areaId], references: [areas.id] }),
}))

export const chatSessionsRelations = relations(chatSessions, ({ one, many }) => ({
  user: one(users, { fields: [chatSessions.userId], references: [users.id] }),
  messages: many(chatMessages),
}))

export const chatMessagesRelations = relations(chatMessages, ({ one, many }) => ({
  session: one(chatSessions, { fields: [chatMessages.sessionId], references: [chatSessions.id] }),
  feedback: many(messageFeedback),
}))

export const messageFeedbackRelations = relations(messageFeedback, ({ one }) => ({
  message: one(chatMessages, { fields: [messageFeedback.messageId], references: [chatMessages.id] }),
  user: one(users, { fields: [messageFeedback.userId], references: [users.id] }),
}))

// ── Type exports (Drizzle inferred) ───────────────────────────────────────

export type DbArea = typeof areas.$inferSelect
export type NewArea = typeof areas.$inferInsert
export type DbUser = typeof users.$inferSelect
export type NewUser = typeof users.$inferInsert
export type DbUserArea = typeof userAreas.$inferSelect
export type DbAreaCollection = typeof areaCollections.$inferSelect
export type DbChatSession = typeof chatSessions.$inferSelect
export type NewChatSession = typeof chatSessions.$inferInsert
export type DbChatMessage = typeof chatMessages.$inferSelect
export type NewChatMessage = typeof chatMessages.$inferInsert
export type DbSessionTag = typeof sessionTags.$inferSelect
export type DbAnnotation = typeof annotations.$inferSelect
export type NewAnnotation = typeof annotations.$inferInsert
export type DbSavedResponse = typeof savedResponses.$inferSelect
export type NewSavedResponse = typeof savedResponses.$inferInsert
export type DbIngestionJob = typeof ingestionJobs.$inferSelect
export type NewIngestionJob = typeof ingestionJobs.$inferInsert
export type DbIngestionQueueItem = typeof ingestionQueue.$inferSelect
export type NewIngestionQueueItem = typeof ingestionQueue.$inferInsert
export type DbEvent = typeof events.$inferSelect
export type NewEvent = typeof events.$inferInsert
