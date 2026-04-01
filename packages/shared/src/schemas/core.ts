/**
 * Core Zod schemas — enums, entities, auth, API wrappers, pagination
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

export const MessageRoleSchema = z.enum(["user", "assistant", "system"])
export type MessageRole = z.infer<typeof MessageRoleSchema>

export const FeedbackRatingSchema = z.enum(["up", "down"])
export type FeedbackRating = z.infer<typeof FeedbackRatingSchema>

// ── Core entities ──────────────────────────────────────────────────────────

export const AreaSchema = z.object({
  id: z.number().int().positive(),
  name: z.string().min(1).max(100),
  description: z.string().default(""),
  createdAt: z.number().int(),
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
  preferences: z.record(z.string(), z.unknown()).default({}),
  createdAt: z.number().int(),
  lastLogin: z.number().int().nullable().default(null),
  areas: z.array(AreaSchema).optional(),
})
export type User = z.infer<typeof UserSchema>

export const UserPublicSchema = UserSchema.omit({}).extend({
  areas: z.array(AreaSchema).optional(),
})
export type UserPublic = z.infer<typeof UserPublicSchema>

// ── Auth ───────────────────────────────────────────────────────────────────

export const LoginRequestSchema = z.object({
  email: z.string().min(1).toLowerCase(),
  password: z.string().min(1),
})
export type LoginRequest = z.infer<typeof LoginRequestSchema>

export const LoginResponseSchema = z.object({
  user: UserPublicSchema,
  token: z.string(),
})
export type LoginResponse = z.infer<typeof LoginResponseSchema>

export const JwtClaimsSchema = z.object({
  sub: z.string(),
  email: z.string().email(),
  name: z.string(),
  role: RoleSchema,
  iat: z.number().int(),
  exp: z.number().int(),
  jti: z.string().optional(),
})
export type JwtClaims = z.infer<typeof JwtClaimsSchema>

// ── User Preferences ───────────────────────────────────────────────────────

export const UserPreferencesSchema = z.object({
  defaultCollection: z.string().optional(),
  vdbTopK: z.number().int().min(1).max(100).default(10),
  rerankerTopK: z.number().int().min(1).max(50).default(5),
  temperature: z.number().min(0).max(2).default(0.2),
  useReranker: z.boolean().default(true),
  crossdocEnabled: z.boolean().default(false),
  crossdocMaxSubQueries: z.number().int().min(1).max(20).default(5),
  crossdocSynthesisModel: z.string().optional(),
  theme: z.enum(["light", "dark", "system"]).default("system"),
  language: z.string().default("es"),
  notifyIngestionComplete: z.boolean().default(true),
  notifyIngestionError: z.boolean().default(true),
})
export type UserPreferences = z.infer<typeof UserPreferencesSchema>

// ── SSO ───────────────────────────────────────────────────────────────────

export const SsoProviderTypeSchema = z.enum(["google", "microsoft", "github", "oidc_generic"])
export type SsoProviderType = z.infer<typeof SsoProviderTypeSchema>

/** Full SSO provider (admin view — no secret) */
export const SsoProviderSchema = z.object({
  id: z.number().int().positive(),
  name: z.string().min(1),
  type: SsoProviderTypeSchema,
  clientId: z.string().min(1),
  tenantId: z.string().nullable(),
  issuerUrl: z.string().nullable(),
  scopes: z.string(),
  autoProvision: z.boolean(),
  defaultRole: RoleSchema,
  active: z.boolean(),
  createdAt: z.number().int(),
  updatedAt: z.number().int(),
})
export type SsoProvider = z.infer<typeof SsoProviderSchema>

/** Public SSO provider (login page — minimal fields) */
export const SsoProviderPublicSchema = z.object({
  id: z.number().int().positive(),
  name: z.string(),
  type: SsoProviderTypeSchema,
})
export type SsoProviderPublic = z.infer<typeof SsoProviderPublicSchema>

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
