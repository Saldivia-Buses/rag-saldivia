/**
 * Chat domain tables: sessions, messages, feedback, shares, tags,
 * annotations, saved responses, prompt templates, projects.
 */

import {
  sqliteTable,
  text,
  integer,
  primaryKey,
  uniqueIndex,
  index,
} from "drizzle-orm/sqlite-core"
import { users } from "./core"

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
    forkedFrom: text("forked_from"),  // FK a chat_sessions(id) — sin constraint Drizzle para evitar self-reference circular
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

// ── Session Shares ─────────────────────────────────────────────────────────

export const sessionShares = sqliteTable(
  "session_shares",
  {
    id: text("id").primaryKey(), // UUID
    sessionId: text("session_id")
      .notNull()
      .references(() => chatSessions.id, { onDelete: "cascade" }),
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    token: text("token").notNull().unique(), // 64-char hex
    expiresAt: integer("expires_at").notNull(), // epoch ms
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    tokenIdx: uniqueIndex("idx_session_shares_token").on(t.token),
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

// ── Prompt Templates ───────────────────────────────────────────────────────

export const promptTemplates = sqliteTable(
  "prompt_templates",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    title: text("title").notNull(),
    prompt: text("prompt").notNull(),
    focusMode: text("focus_mode").notNull().default("detallado"),
    createdBy: integer("created_by")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    active: integer("active", { mode: "boolean" }).notNull().default(true),
    createdAt: integer("created_at").notNull(),
  },
  (t) => ({
    activeIdx: index("idx_prompt_templates_active").on(t.active),
  })
)

// ── Projects ───────────────────────────────────────────────────────────────

export const projects = sqliteTable(
  "projects",
  {
    id: text("id").primaryKey(), // UUID
    userId: integer("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    name: text("name").notNull(),
    description: text("description").notNull().default(""),
    instructions: text("instructions").notNull().default(""), // system prompt adicional
    createdAt: integer("created_at").notNull(),
    updatedAt: integer("updated_at").notNull(),
  },
  (t) => ({
    userIdx: index("idx_projects_user").on(t.userId),
  })
)

export const projectSessions = sqliteTable(
  "project_sessions",
  {
    projectId: text("project_id")
      .notNull()
      .references(() => projects.id, { onDelete: "cascade" }),
    sessionId: text("session_id")
      .notNull()
      .references(() => chatSessions.id, { onDelete: "cascade" }),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.projectId, t.sessionId] }),
  })
)

export const projectCollections = sqliteTable(
  "project_collections",
  {
    projectId: text("project_id")
      .notNull()
      .references(() => projects.id, { onDelete: "cascade" }),
    collectionName: text("collection_name").notNull(),
  },
  (t) => ({
    pk: primaryKey({ columns: [t.projectId, t.collectionName] }),
  })
)

// ── Type exports (Drizzle inferred) ───────────────────────────────────────

export type DbChatSession = typeof chatSessions.$inferSelect
export type NewChatSession = typeof chatSessions.$inferInsert
export type DbChatMessage = typeof chatMessages.$inferSelect
export type NewChatMessage = typeof chatMessages.$inferInsert
export type DbSessionShare = typeof sessionShares.$inferSelect
export type DbSessionTag = typeof sessionTags.$inferSelect
export type DbAnnotation = typeof annotations.$inferSelect
export type NewAnnotation = typeof annotations.$inferInsert
export type DbSavedResponse = typeof savedResponses.$inferSelect
export type NewSavedResponse = typeof savedResponses.$inferInsert
export type DbPromptTemplate = typeof promptTemplates.$inferSelect
export type NewPromptTemplate = typeof promptTemplates.$inferInsert
export type DbProject = typeof projects.$inferSelect
export type NewProject = typeof projects.$inferInsert
export type DbProjectSession = typeof projectSessions.$inferSelect
export type DbProjectCollection = typeof projectCollections.$inferSelect
