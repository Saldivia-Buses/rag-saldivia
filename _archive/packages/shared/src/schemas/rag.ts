/**
 * RAG Zod schemas — chat, collections, ingestion, events, config, focus modes
 */

import { z } from "zod"
import {
  MessageRoleSchema,
  FeedbackRatingSchema,
  EventSourceSchema,
  LogLevelSchema,
} from "./core"

// ── Chat ───────────────────────────────────────────────────────────────────

export const CitationSchema = z.object({
  id: z.string().optional(),
  document: z.string().optional(),
  content: z.string().optional(),
  score: z.number().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
})
export type Citation = z.infer<typeof CitationSchema>

export const ChatMessageSchema = z.object({
  id: z.number().int().optional(),
  sessionId: z.string().uuid(),
  role: MessageRoleSchema,
  content: z.string(),
  sources: z.array(CitationSchema).nullable().default(null),
  feedback: FeedbackRatingSchema.nullable().default(null),
  timestamp: z.number().int(),
})
export type ChatMessage = z.infer<typeof ChatMessageSchema>

export const ChatSessionSchema = z.object({
  id: z.string().uuid(),
  userId: z.number().int().positive(),
  title: z.string(),
  collection: z.string(),
  crossdoc: z.boolean().default(false),
  messageCount: z.number().int().default(0),
  createdAt: z.number().int(),
  updatedAt: z.number().int(),
})
export type ChatSession = z.infer<typeof ChatSessionSchema>

export const CreateSessionSchema = z.object({
  collection: z.string().min(1),
  crossdoc: z.boolean().default(false),
  title: z.string().optional(),
})
export type CreateSession = z.infer<typeof CreateSessionSchema>

// ── Collections ────────────────────────────────────────────────────────────

export const CollectionNameSchema = z.string()
  .min(1).max(64)
  .regex(/^[a-z0-9_-]+$/, "Solo minúsculas, números, guiones y guiones bajos")

export const CollectionSchema = z.object({
  name: CollectionNameSchema,
  documentCount: z.number().int().nonnegative().default(0),
  createdAt: z.number().int().optional(),
})
export type Collection = z.infer<typeof CollectionSchema>

// ── Ingestion ──────────────────────────────────────────────────────────────

export const IngestionStateSchema = z.enum([
  "pending", "running", "stalled", "done", "error", "cancelled",
])
export type IngestionState = z.infer<typeof IngestionStateSchema>

export const IngestionTierSchema = z.enum(["tiny", "small", "medium", "large"])
export type IngestionTier = z.infer<typeof IngestionTierSchema>

export const IngestionJobSchema = z.object({
  id: z.string(),
  userId: z.number().int().positive(),
  taskId: z.string(),
  filename: z.string(),
  collection: z.string(),
  tier: IngestionTierSchema,
  pageCount: z.number().int().nullable().default(null),
  state: IngestionStateSchema.default("pending"),
  progress: z.number().int().min(0).max(100).default(0),
  fileHash: z.string().nullable().default(null),
  retryCount: z.number().int().default(0),
  lastChecked: z.number().int().nullable().default(null),
  createdAt: z.number().int(),
  completedAt: z.number().int().nullable().default(null),
})
export type IngestionJob = z.infer<typeof IngestionJobSchema>

export const IngestionAlertSchema = z.object({
  id: z.string(),
  jobId: z.string(),
  userId: z.number().int().positive(),
  filename: z.string(),
  collection: z.string(),
  tier: IngestionTierSchema,
  pageCount: z.number().int().nullable(),
  fileHash: z.string().nullable(),
  error: z.string().nullable(),
  retryCount: z.number().int().nullable(),
  progressAtFailure: z.number().int().nullable(),
  gatewayVersion: z.string().nullable(),
  createdAt: z.number().int(),
  resolvedAt: z.number().int().nullable(),
  resolvedBy: z.string().nullable(),
  notes: z.string().nullable(),
})
export type IngestionAlert = z.infer<typeof IngestionAlertSchema>

export const QueueItemStatusSchema = z.enum(["pending", "locked", "done", "error"])
export type QueueItemStatus = z.infer<typeof QueueItemStatusSchema>

export const IngestionQueueItemSchema = z.object({
  id: z.string().uuid(),
  collection: z.string(),
  filePath: z.string(),
  userId: z.number().int().positive(),
  priority: z.number().int().default(0),
  status: QueueItemStatusSchema.default("pending"),
  lockedAt: z.number().int().nullable().default(null),
  lockedBy: z.string().nullable().default(null),
  createdAt: z.number().int(),
  startedAt: z.number().int().nullable().default(null),
  completedAt: z.number().int().nullable().default(null),
  error: z.string().nullable().default(null),
  retryCount: z.number().int().default(0),
})
export type IngestionQueueItem = z.infer<typeof IngestionQueueItemSchema>

// ── Events ─────────────────────────────────────────────────────────────────

export const EventTypeSchema = z.enum([
  "auth.login", "auth.logout", "auth.refresh", "auth.failed", "auth.password_changed",
  "user.created", "user.updated", "user.deleted", "user.area_assigned", "user.area_removed",
  "rag.query", "rag.query_crossdoc", "rag.error", "rag.stream_started", "rag.stream_completed",
  "ingestion.started", "ingestion.completed", "ingestion.failed", "ingestion.cancelled", "ingestion.stalled",
  "area.created", "area.updated", "area.deleted",
  "collection.created", "collection.deleted",
  "admin.config_changed", "admin.profile_switched",
  "client.action", "client.navigation", "client.error",
  "system.start", "system.error", "system.warning", "system.request",
])
export type EventType = z.infer<typeof EventTypeSchema>

export const LogEventSchema = z.object({
  id: z.string().uuid(),
  ts: z.number().int(),
  source: EventSourceSchema,
  level: LogLevelSchema,
  type: EventTypeSchema,
  userId: z.number().int().nullable().default(null),
  sessionId: z.string().nullable().default(null),
  payload: z.record(z.string(), z.unknown()).default({}),
  sequence: z.number().int(),
})
export type LogEvent = z.infer<typeof LogEventSchema>

// ── RAG Config ─────────────────────────────────────────────────────────────

export const RagParamsSchema = z.object({
  temperature: z.number().min(0).max(2).default(0.2),
  top_p: z.number().min(0).max(1).default(0.7),
  max_tokens: z.number().int().min(1).max(8192).default(1024),
  vdb_top_k: z.number().int().min(1).max(100).default(10),
  reranker_top_k: z.number().int().min(1).max(50).default(5),
  use_guardrails: z.boolean().default(false),
  use_reranker: z.boolean().default(true),
  chunk_size: z.number().int().min(128).max(2048).default(512),
  chunk_overlap: z.number().int().min(0).max(512).default(50),
  embedding_model: z.string().default("nvidia/nv-embedqa-e5-v5"),
})
export type RagParams = z.infer<typeof RagParamsSchema>

// ── Focus modes ─────────────────────────────────────────────────────────────

export const FOCUS_MODE_IDS = ["detallado", "ejecutivo", "tecnico", "comparativo"] as const
export type FocusModeId = typeof FOCUS_MODE_IDS[number]

export type FocusMode = {
  id: FocusModeId
  label: string
  systemPrompt: string
}

export const FOCUS_MODES: FocusMode[] = [
  { id: "detallado", label: "Detallado", systemPrompt: "Respond with a thorough, comprehensive answer. Include context, examples, and detailed explanations." },
  { id: "ejecutivo", label: "Ejecutivo", systemPrompt: "Respond concisely with an executive summary. Lead with the key insight in 1-2 sentences, then bullet points if needed. Avoid jargon." },
  { id: "tecnico", label: "Técnico", systemPrompt: "Respond with precise technical detail. Include specifications, data types, code examples, and exact values where relevant." },
  { id: "comparativo", label: "Comparativo", systemPrompt: "Respond by comparing and contrasting the relevant options or perspectives. Use a structured comparison format." },
]
