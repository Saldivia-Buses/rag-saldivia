/**
 * @rag-saldivia/shared — Zod schemas
 *
 * Fuente de verdad para todos los tipos del sistema.
 * Reemplaza la combinación de modelos Pydantic (Python) + interfaces TypeScript
 * que estaban duplicados entre gateway.py y gateway.ts.
 */

import { z } from "zod"

// ── Enums ──────────────────────────────────────────────────────────────────

export const RoleSchema = z.enum(["admin", "area_manager", "user"])
export type Role = z.infer<typeof RoleSchema>

export const PermissionSchema = z.enum(["read", "write", "admin"])
export type Permission = z.infer<typeof PermissionSchema>

export const LogLevelSchema = z.enum(["TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"])
export type LogLevel = z.infer<typeof LogLevelSchema>

export const EventSourceSchema = z.enum(["frontend", "backend"])
export type EventSource = z.infer<typeof EventSourceSchema>

export const IngestionStateSchema = z.enum([
  "pending",
  "running",
  "stalled",
  "done",
  "error",
  "cancelled",
])
export type IngestionState = z.infer<typeof IngestionStateSchema>

export const IngestionTierSchema = z.enum(["tiny", "small", "medium", "large"])
export type IngestionTier = z.infer<typeof IngestionTierSchema>

export const MessageRoleSchema = z.enum(["user", "assistant", "system"])
export type MessageRole = z.infer<typeof MessageRoleSchema>

export const FeedbackRatingSchema = z.enum(["up", "down"])
export type FeedbackRating = z.infer<typeof FeedbackRatingSchema>

// ── Core entities ──────────────────────────────────────────────────────────

export const AreaSchema = z.object({
  id: z.number().int().positive(),
  name: z.string().min(1).max(100),
  description: z.string().default(""),
  createdAt: z.number().int(), // epoch ms (Temporal.Now.instant().epochMilliseconds)
})
export type Area = z.infer<typeof AreaSchema>

export const AreaCollectionSchema = z.object({
  areaId: z.number().int().positive(),
  collectionName: z.string().min(1),
  permission: PermissionSchema,
})
export type AreaCollection = z.infer<typeof AreaCollectionSchema>

export const UserSchema = z.object({
  id: z.number().int().positive(),
  email: z.string().email(),
  name: z.string().min(1).max(200),
  role: RoleSchema,
  active: z.boolean().default(true),
  preferences: z.record(z.unknown()).default({}),
  createdAt: z.number().int(),
  lastLogin: z.number().int().nullable().default(null),
  areas: z.array(AreaSchema).optional(), // populated on request
})
export type User = z.infer<typeof UserSchema>

// UserPublic: lo que se envía al cliente (sin hashes)
export const UserPublicSchema = UserSchema.omit({}).extend({
  areas: z.array(AreaSchema).optional(),
})
export type UserPublic = z.infer<typeof UserPublicSchema>

// ── Auth ───────────────────────────────────────────────────────────────────

export const LoginRequestSchema = z.object({
  // Acepta emails con y sin TLD (admin@localhost es válido en desarrollo)
  email: z.string().min(1).toLowerCase(),
  password: z.string().min(1),
})
export type LoginRequest = z.infer<typeof LoginRequestSchema>

export const LoginResponseSchema = z.object({
  user: UserPublicSchema,
  token: z.string(), // JWT (también seteado como cookie HttpOnly)
})
export type LoginResponse = z.infer<typeof LoginResponseSchema>

export const JwtClaimsSchema = z.object({
  sub: z.string(), // user id as string
  email: z.string().email(),
  name: z.string(),
  role: RoleSchema,
  iat: z.number().int(),
  exp: z.number().int(),
})
export type JwtClaims = z.infer<typeof JwtClaimsSchema>

// ── Chat ───────────────────────────────────────────────────────────────────

export const CitationSchema = z.object({
  id: z.string().optional(),
  document: z.string().optional(),
  content: z.string().optional(),
  score: z.number().optional(),
  metadata: z.record(z.unknown()).optional(),
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

export const CollectionSchema = z.object({
  name: z.string().min(1),
  documentCount: z.number().int().nonnegative().default(0),
  createdAt: z.number().int().optional(),
})
export type Collection = z.infer<typeof CollectionSchema>

// ── Ingestion ──────────────────────────────────────────────────────────────

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

// ── Ingestion Queue (reemplaza Redis) ─────────────────────────────────────

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
  lockedBy: z.string().nullable().default(null), // worker instance id
  createdAt: z.number().int(),
  startedAt: z.number().int().nullable().default(null),
  completedAt: z.number().int().nullable().default(null),
  error: z.string().nullable().default(null),
  retryCount: z.number().int().default(0),
})
export type IngestionQueueItem = z.infer<typeof IngestionQueueItemSchema>

// ── Black Box Events ───────────────────────────────────────────────────────

export const EventTypeSchema = z.enum([
  // Auth
  "auth.login",
  "auth.logout",
  "auth.refresh",
  "auth.failed",
  "auth.password_changed",
  // Users
  "user.created",
  "user.updated",
  "user.deleted",
  "user.area_assigned",
  "user.area_removed",
  // RAG
  "rag.query",
  "rag.query_crossdoc",
  "rag.error",
  "rag.stream_started",
  "rag.stream_completed",
  // Ingestion
  "ingestion.started",
  "ingestion.completed",
  "ingestion.failed",
  "ingestion.cancelled",
  "ingestion.stalled",
  // Areas
  "area.created",
  "area.updated",
  "area.deleted",
  // Collections
  "collection.created",
  "collection.deleted",
  // Admin
  "admin.config_changed",
  "admin.profile_switched",
  // Frontend (eventos del browser)
  "client.action",
  "client.navigation",
  "client.error",
  // System
  "system.start",
  "system.error",
  "system.warning",
])
export type EventType = z.infer<typeof EventTypeSchema>

export const LogEventSchema = z.object({
  id: z.string().uuid(),
  ts: z.number().int(), // Temporal.Now.instant().epochMilliseconds
  source: EventSourceSchema,
  level: LogLevelSchema,
  type: EventTypeSchema,
  userId: z.number().int().nullable().default(null),
  sessionId: z.string().nullable().default(null),
  payload: z.record(z.unknown()).default({}),
  sequence: z.number().int(), // monotónico, para replay ordenado
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

// ── User Preferences ───────────────────────────────────────────────────────

export const UserPreferencesSchema = z.object({
  // RAG
  defaultCollection: z.string().optional(),
  vdbTopK: z.number().int().min(1).max(100).default(10),
  rerankerTopK: z.number().int().min(1).max(50).default(5),
  temperature: z.number().min(0).max(2).default(0.2),
  useReranker: z.boolean().default(true),
  // Crossdoc
  crossdocEnabled: z.boolean().default(false),
  crossdocMaxSubQueries: z.number().int().min(1).max(20).default(5),
  crossdocSynthesisModel: z.string().optional(),
  // UI
  theme: z.enum(["light", "dark", "system"]).default("system"),
  language: z.string().default("es"),
  // Notifications
  notifyIngestionComplete: z.boolean().default(true),
  notifyIngestionError: z.boolean().default(true),
})
export type UserPreferences = z.infer<typeof UserPreferencesSchema>

// ── API Response wrappers ──────────────────────────────────────────────────

export const ApiSuccessSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    ok: z.literal(true),
    data: dataSchema,
  })

export const ApiErrorSchema = z.object({
  ok: z.literal(false),
  error: z.string(),
  code: z.string().optional(),
  suggestion: z.string().optional(),
})
export type ApiError = z.infer<typeof ApiErrorSchema>

export type ApiResponse<T> = { ok: true; data: T } | ApiError

// ── Pagination ─────────────────────────────────────────────────────────────

export const PaginationSchema = z.object({
  page: z.number().int().min(1).default(1),
  limit: z.number().int().min(1).max(200).default(50),
  total: z.number().int().nonnegative(),
})
export type Pagination = z.infer<typeof PaginationSchema>

// ── Audit log ──────────────────────────────────────────────────────────────

export const AuditEntrySchema = z.object({
  id: z.number().int(),
  userId: z.number().int().positive(),
  action: z.string(),
  collection: z.string().nullable(),
  queryPreview: z.string().nullable(),
  ipAddress: z.string(),
  timestamp: z.number().int(),
})
export type AuditEntry = z.infer<typeof AuditEntrySchema>
