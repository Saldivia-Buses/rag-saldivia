/**
 * Drizzle ORM relations — required for `.query` with `with` option.
 *
 * All relation definitions live here to avoid circular imports
 * between domain files.
 */

import { relations } from "drizzle-orm"

import {
  areas,
  users,
  userAreas,
  areaCollections,
  roles,
  permissions,
  rolePermissions,
  userRoleAssignments,
} from "./core"

import {
  chatSessions,
  chatMessages,
  messageFeedback,
} from "./chat"

import {
  channels,
  channelMembers,
  msgMessages,
  msgReactions,
  msgMentions,
  pinnedMessages,
} from "./messaging"

// ── Core relations ─────────────────────────────────────────────────────────

export const usersRelations = relations(users, ({ many }) => ({
  userAreas: many(userAreas),
  chatSessions: many(chatSessions),
  userRoleAssignments: many(userRoleAssignments),
  channelMembers: many(channelMembers),
  msgMessages: many(msgMessages),
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

// ── RBAC relations ─────────────────────────────────────────────────────────

export const rolesRelations = relations(roles, ({ many }) => ({
  rolePermissions: many(rolePermissions),
  userRoleAssignments: many(userRoleAssignments),
}))

export const permissionsRelations = relations(permissions, ({ many }) => ({
  rolePermissions: many(rolePermissions),
}))

export const rolePermissionsRelations = relations(rolePermissions, ({ one }) => ({
  role: one(roles, { fields: [rolePermissions.roleId], references: [roles.id] }),
  permission: one(permissions, { fields: [rolePermissions.permissionId], references: [permissions.id] }),
}))

export const userRoleAssignmentsRelations = relations(userRoleAssignments, ({ one }) => ({
  user: one(users, { fields: [userRoleAssignments.userId], references: [users.id] }),
  role: one(roles, { fields: [userRoleAssignments.roleId], references: [roles.id] }),
}))

// ── Chat relations ─────────────────────────────────────────────────────────

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

// ── Messaging relations ────────────────────────────────────────────────────

export const channelsRelations = relations(channels, ({ many }) => ({
  channelMembers: many(channelMembers),
  msgMessages: many(msgMessages),
  pinnedMessages: many(pinnedMessages),
}))

export const channelMembersRelations = relations(channelMembers, ({ one }) => ({
  channel: one(channels, { fields: [channelMembers.channelId], references: [channels.id] }),
  user: one(users, { fields: [channelMembers.userId], references: [users.id] }),
}))

export const msgMessagesRelations = relations(msgMessages, ({ one, many }) => ({
  channel: one(channels, { fields: [msgMessages.channelId], references: [channels.id] }),
  user: one(users, { fields: [msgMessages.userId], references: [users.id] }),
  reactions: many(msgReactions),
  mentions: many(msgMentions),
}))

export const msgReactionsRelations = relations(msgReactions, ({ one }) => ({
  message: one(msgMessages, { fields: [msgReactions.messageId], references: [msgMessages.id] }),
  user: one(users, { fields: [msgReactions.userId], references: [users.id] }),
}))

export const msgMentionsRelations = relations(msgMentions, ({ one }) => ({
  message: one(msgMessages, { fields: [msgMentions.messageId], references: [msgMessages.id] }),
}))

export const pinnedMessagesRelations = relations(pinnedMessages, ({ one }) => ({
  channel: one(channels, { fields: [pinnedMessages.channelId], references: [channels.id] }),
  message: one(msgMessages, { fields: [pinnedMessages.messageId], references: [msgMessages.id] }),
  pinnedByUser: one(users, { fields: [pinnedMessages.pinnedBy], references: [users.id] }),
}))
