/**
 * Events domain tables: black-box events, webhooks, scheduled reports,
 * collection history, ingestion jobs, ingestion alerts.
 */

import {
  sqliteTable,
  text,
  integer,
  index,
} from "drizzle-orm/sqlite-core"
import { users } from "./core"

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
    // Índice compuesto para queries de analytics: type AND userId AND ts (O(log n) vs full scan)
    queryIdx: index("idx_events_query").on(t.type, t.userId, t.ts),
  })
)

// ── Webhooks ───────────────────────────────────────────────────────────────

export const webhooks = sqliteTable(
  "webhooks",
  {
    id: text("id").primaryKey(), // UUID
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    url: text("url").notNull(),
    events: text("events", { mode: "json" })
      .$type<string[]>()
      .notNull()
      .default([]),
    secret: text("secret").notNull(),
    active: integer("active", { mode: "boolean" }).notNull().default(true),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    activeIdx: index("idx_webhooks_active").on(t.active),
  })
)

// ── Scheduled Reports ──────────────────────────────────────────────────────

export const scheduledReports = sqliteTable(
  "scheduled_reports",
  {
    id: text("id").primaryKey(), // UUID
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    query: text("query").notNull(),
    collection: text("collection").notNull(),
    schedule: text("schedule", { enum: ["daily", "weekly", "monthly"] }).notNull(),
    destination: text("destination", { enum: ["saved", "email"] }).notNull(),
    email: text("email"),
    active: integer("active", { mode: "boolean" }).notNull().default(true),
    lastRun: integer("last_run"),
    nextRun: integer("next_run").notNull(),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    activeNextRunIdx: index("idx_reports_active_next_run").on(t.active, t.nextRun),
  })
)

// ── Collection History ─────────────────────────────────────────────────────

export const collectionHistory = sqliteTable(
  "collection_history",
  {
    id: text("id").primaryKey(), // UUID
    collection: text("collection").notNull(),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    action: text("action", { enum: ["added", "removed"] }).notNull(),
    filename: text("filename"),
    docCount: integer("doc_count"),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    collectionIdx: index("idx_collection_history_collection").on(t.collection),
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

// ── Type exports (Drizzle inferred) ───────────────────────────────────────

export type DbEvent = typeof events.$inferSelect
export type NewEvent = typeof events.$inferInsert
export type DbWebhook = typeof webhooks.$inferSelect
export type NewWebhook = typeof webhooks.$inferInsert
export type DbScheduledReport = typeof scheduledReports.$inferSelect
export type NewScheduledReport = typeof scheduledReports.$inferInsert
export type DbCollectionHistory = typeof collectionHistory.$inferSelect
export type NewCollectionHistory = typeof collectionHistory.$inferInsert
export type DbIngestionJob = typeof ingestionJobs.$inferSelect
export type NewIngestionJob = typeof ingestionJobs.$inferInsert
