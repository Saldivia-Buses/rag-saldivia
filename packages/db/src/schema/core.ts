/**
 * Core domain tables: areas, users, RBAC, rate limits, external integrations.
 */

import {
  sqliteTable,
  text,
  integer,
  primaryKey,
  uniqueIndex,
  index,
} from "drizzle-orm/sqlite-core"

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
    onboardingCompleted: integer("onboarding_completed", { mode: "boolean" }).notNull().default(false),
    ssoProvider: text("sso_provider"),  // "google" | "azure" | null
    ssoSubject: text("sso_subject"),    // external user ID from provider
    createdAt: integer("created_at").notNull(), // epoch ms
    lastLogin: integer("last_login"), // epoch ms, nullable
    lastSeen: integer("last_seen"), // epoch ms, updated on each request for presence
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

// ── User Memory ────────────────────────────────────────────────────────────

export const userMemory = sqliteTable(
  "user_memory",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    key: text("key").notNull(),
    value: text("value").notNull(),
    source: text("source", { enum: ["explicit", "inferred"] }).notNull().default("explicit"),
    createdAt: integer("created_at").notNull(),
    updatedAt: integer("updated_at").notNull(),
  },
  (t) => ({
    uniqueKey: uniqueIndex("idx_user_memory_unique").on(t.userId, t.key),
  })
)

// ── RBAC: Roles ──────────────────────────────────────────────────────────

export const roles = sqliteTable("roles", {
  id: integer("id").primaryKey({ autoIncrement: true }),
  name: text("name").notNull().unique(),
  description: text("description").notNull().default(""),
  level: integer("level").notNull().default(0), // higher = more powerful
  color: text("color").notNull().default("#6e6c69"), // hex for UI badges
  icon: text("icon").notNull().default("user"), // lucide icon name
  isSystem: integer("is_system", { mode: "boolean" }).notNull().default(false),
  createdAt: integer("created_at").notNull(),
})

// ── RBAC: Permission catalog ─────────────────────────────────────────────

export const permissions = sqliteTable("permissions", {
  id: integer("id").primaryKey({ autoIncrement: true }),
  key: text("key").notNull().unique(), // e.g. "users.manage"
  label: text("label").notNull(), // e.g. "Gestionar usuarios"
  category: text("category").notNull(), // e.g. "Usuarios"
  description: text("description").notNull().default(""),
})

// ── RBAC: Role ↔ Permission (many-to-many) ───────────────────────────────

export const rolePermissions = sqliteTable(
  "role_permissions",
  {
    roleId: integer("role_id")
      .notNull()
      .references(() => roles.id, { onDelete: "cascade" }),
    permissionId: integer("permission_id")
      .notNull()
      .references(() => permissions.id, { onDelete: "cascade" }),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.roleId, t.permissionId] }),
  })
)

// ── RBAC: User ↔ Role (many-to-many) ─────────────────────────────────────

export const userRoleAssignments = sqliteTable(
  "user_role_assignments",
  {
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    roleId: integer("role_id")
      .notNull()
      .references(() => roles.id, { onDelete: "cascade" }),
    assignedAt: integer("assigned_at").notNull(),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.userId, t.roleId] }),
  })
)

// ── Rate Limits ────────────────────────────────────────────────────────────

export const rateLimits = sqliteTable(
  "rate_limits",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    targetType: text("target_type", { enum: ["user", "area"] }).notNull(),
    targetId: integer("target_id").notNull(),
    maxQueriesPerHour: integer("max_queries_per_hour").notNull(),
    active: integer("active", { mode: "boolean" }).notNull().default(true),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    targetIdx: index("idx_rate_limits_target").on(t.targetType, t.targetId),
  })
)

// ── Bot User Mappings ──────────────────────────────────────────────────────

export const botUserMappings = sqliteTable(
  "bot_user_mappings",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    platform: text("platform", { enum: ["slack", "teams"] }).notNull(),
    externalUserId: text("external_user_id").notNull(), // Slack/Teams user ID
    systemUserId: integer("system_user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    uniqueMapping: uniqueIndex("idx_bot_user_mapping_unique").on(t.platform, t.externalUserId),
  })
)

// ── External Sources ───────────────────────────────────────────────────────

export const externalSources = sqliteTable(
  "external_sources",
  {
    id: text("id").primaryKey(), // UUID
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    provider: text("provider", { enum: ["google_drive", "sharepoint", "confluence"] }).notNull(),
    name: text("name").notNull(),
    credentials: text("credentials").notNull().default("{}"), // JSON cifrado (en prod: cifrar con SYSTEM_API_KEY)
    collectionDest: text("collection_dest").notNull(),
    schedule: text("schedule", { enum: ["hourly", "daily", "weekly"] }).notNull().default("daily"),
    active: integer("active", { mode: "boolean" }).notNull().default(true),
    lastSync: integer("last_sync"),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    userIdx: index("idx_external_sources_user").on(t.userId),
  })
)

// ── Type exports (Drizzle inferred) ───────────────────────────────────────

export type DbArea = typeof areas.$inferSelect
export type NewArea = typeof areas.$inferInsert
export type DbUser = typeof users.$inferSelect
export type NewUser = typeof users.$inferInsert
export type DbUserArea = typeof userAreas.$inferSelect
export type DbAreaCollection = typeof areaCollections.$inferSelect
export type DbUserMemory = typeof userMemory.$inferSelect
export type DbRole = typeof roles.$inferSelect
export type NewRole = typeof roles.$inferInsert
export type DbPermission = typeof permissions.$inferSelect
export type DbRolePermission = typeof rolePermissions.$inferSelect
export type DbUserRoleAssignment = typeof userRoleAssignments.$inferSelect
export type DbRateLimit = typeof rateLimits.$inferSelect
export type NewRateLimit = typeof rateLimits.$inferInsert
export type DbBotUserMapping = typeof botUserMappings.$inferSelect
export type DbExternalSource = typeof externalSources.$inferSelect
export type NewExternalSource = typeof externalSources.$inferInsert
