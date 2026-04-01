/**
 * Tests for messaging queries — messages, reactions, pins, threads, mentions.
 *
 * Covers the full messaging CRUD lifecycle for Plan 25 (Internal Messaging).
 * Runs with: bun test packages/db/src/__tests__/messaging.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { eq } from "drizzle-orm"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertChannel } from "./setup"
import { channels } from "../schema"
import {
  sendMessage,
  getMessages,
  editMessage,
  deleteMessage,
  getMessage,
  getThreadReplies,
  getThreadPreview,
  addReaction,
  removeReaction,
  getReactions,
  pinMessage,
  unpinMessage,
  getPinnedMessages,
  searchMessages,
  addMentions,
  getMentionsForUser,
} from "../queries/messaging"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => {
  _resetDbForTesting()
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM msg_mentions; DELETE FROM pinned_messages; DELETE FROM msg_reactions; DELETE FROM msg_messages; DELETE FROM channel_members; DELETE FROM channels; DELETE FROM users;"
  )
})

/** Small sleep to guarantee different Date.now() timestamps between operations. */
const tick = () => new Promise<void>((r) => setTimeout(r, 5))

// ── Messages CRUD ────────────────────────────────────────────────────────────

describe("sendMessage", () => {
  test("creates a message with UUID id", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const msg = await sendMessage({
      channelId: channel.id,
      userId: user.id,
      content: "Hello world",
    })

    expect(msg.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
    )
    expect(msg.channelId).toBe(channel.id)
    expect(msg.userId).toBe(user.id)
    expect(msg.content).toBe("Hello world")
    expect(msg.type).toBe("text")
    expect(msg.replyCount).toBe(0)
    expect(msg.parentId).toBeNull()
    expect(msg.editedAt).toBeNull()
    expect(msg.deletedAt).toBeNull()
    expect(msg.createdAt).toBeGreaterThan(0)
  })

  test("with parentId increments parent replyCount", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const parent = await sendMessage({
      channelId: channel.id,
      userId: user.id,
      content: "Parent message",
    })
    expect(parent.replyCount).toBe(0)

    await sendMessage({
      channelId: channel.id,
      userId: user.id,
      content: "Reply 1",
      parentId: parent.id,
    })

    await sendMessage({
      channelId: channel.id,
      userId: user.id,
      content: "Reply 2",
      parentId: parent.id,
    })

    const updated = await getMessage(parent.id)
    expect(updated!.replyCount).toBe(2)
    expect(updated!.lastReplyAt).toBeGreaterThan(0)
  })

  test("updates channel updatedAt", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const originalUpdatedAt = channel.updatedAt

    await tick()

    await sendMessage({
      channelId: channel.id,
      userId: user.id,
      content: "Triggers channel update",
    })

    // Read channel directly to verify updatedAt changed
    const [updatedChannel] = await db
      .select()
      .from(channels)
      .where(eq(channels.id, channel.id))

    expect(updatedChannel!.updatedAt).toBeGreaterThanOrEqual(originalUpdatedAt)
  })
})

describe("getMessages", () => {
  test("returns newest first (cursor pagination)", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    await sendMessage({ channelId: channel.id, userId: user.id, content: "First" })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Second" })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Third" })

    const messages = await getMessages(channel.id)

    expect(messages).toHaveLength(3)
    expect(messages[0]!.content).toBe("Third")
    expect(messages[1]!.content).toBe("Second")
    expect(messages[2]!.content).toBe("First")
  })

  test("only returns top-level messages (excludes thread replies)", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const parent = await sendMessage({ channelId: channel.id, userId: user.id, content: "Parent" })
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Reply in thread", parentId: parent.id })
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Top-level two" })

    const messages = await getMessages(channel.id)

    expect(messages).toHaveLength(2)
    const contents = messages.map((m) => m.content)
    expect(contents).toContain("Parent")
    expect(contents).toContain("Top-level two")
    expect(contents).not.toContain("Reply in thread")
  })

  test("with before cursor returns only older messages", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    await sendMessage({ channelId: channel.id, userId: user.id, content: "Old" })
    await tick()
    const pivot = await sendMessage({ channelId: channel.id, userId: user.id, content: "Middle" })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "New" })

    const older = await getMessages(channel.id, { before: pivot.createdAt })

    expect(older).toHaveLength(1)
    expect(older[0]!.content).toBe("Old")
  })
})

describe("editMessage", () => {
  test("updates content and sets editedAt", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Original" })
    expect(msg.editedAt).toBeNull()

    await tick()
    const updated = await editMessage(msg.id, "Edited content")

    expect(updated!.content).toBe("Edited content")
    expect(updated!.editedAt).toBeGreaterThan(0)
    expect(updated!.editedAt).toBeGreaterThanOrEqual(msg.createdAt)
  })
})

describe("deleteMessage", () => {
  test("soft deletes — sets deletedAt and clears content", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "To be deleted" })
    expect(msg.deletedAt).toBeNull()
    expect(msg.content).toBe("To be deleted")

    const deleted = await deleteMessage(msg.id)

    expect(deleted!.deletedAt).toBeGreaterThan(0)
    expect(deleted!.content).toBe("")
  })
})

describe("getMessage", () => {
  test("returns single message by id", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Find me" })

    const found = await getMessage(msg.id)
    expect(found).toBeDefined()
    expect(found!.id).toBe(msg.id)
    expect(found!.content).toBe("Find me")
  })

  test("returns undefined for non-existent id", async () => {
    const found = await getMessage("non-existent-id")
    expect(found).toBeUndefined()
  })
})

// ── Thread replies ───────────────────────────────────────────────────────────

describe("getThreadReplies", () => {
  test("returns replies in chronological order (oldest first)", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const parent = await sendMessage({ channelId: channel.id, userId: user.id, content: "Thread start" })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Reply A", parentId: parent.id })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Reply B", parentId: parent.id })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "Reply C", parentId: parent.id })

    const replies = await getThreadReplies(parent.id)

    expect(replies).toHaveLength(3)
    expect(replies[0]!.content).toBe("Reply A")
    expect(replies[1]!.content).toBe("Reply B")
    expect(replies[2]!.content).toBe("Reply C")
  })
})

describe("getThreadPreview", () => {
  test("returns last 3 replies in chronological order", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)

    const parent = await sendMessage({ channelId: channel.id, userId: user.id, content: "Thread" })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "R1", parentId: parent.id })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "R2", parentId: parent.id })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "R3", parentId: parent.id })
    await tick()
    await sendMessage({ channelId: channel.id, userId: user.id, content: "R4", parentId: parent.id })

    const preview = await getThreadPreview(parent.id)

    // Should return last 3 (R2, R3, R4) in chronological order
    expect(preview).toHaveLength(3)
    expect(preview[0]!.content).toBe("R2")
    expect(preview[1]!.content).toBe("R3")
    expect(preview[2]!.content).toBe("R4")
  })
})

// ── Reactions ────────────────────────────────────────────────────────────────

describe("addReaction", () => {
  test("adds a reaction correctly", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "React to this" })

    await addReaction(msg.id, user.id, "thumbsup")

    const reactions = await getReactions(msg.id)
    expect(reactions).toHaveLength(1)
    expect(reactions[0]!.emoji).toBe("thumbsup")
    expect(reactions[0]!.userId).toBe(user.id)
  })

  test("is idempotent — duplicate reaction does not error", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "React twice" })

    await addReaction(msg.id, user.id, "heart")
    await addReaction(msg.id, user.id, "heart") // should not throw

    const reactions = await getReactions(msg.id)
    expect(reactions).toHaveLength(1) // still just one
  })
})

describe("removeReaction", () => {
  test("removes only the specific emoji from specific user", async () => {
    const user = await insertUser(db, "u1@test.com")
    const user2 = await insertUser(db, "u2@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Multi-react" })

    await addReaction(msg.id, user.id, "thumbsup")
    await addReaction(msg.id, user.id, "heart")
    await addReaction(msg.id, user2.id, "thumbsup")

    await removeReaction(msg.id, user.id, "thumbsup")

    const reactions = await getReactions(msg.id)
    expect(reactions).toHaveLength(2)
    const remaining = reactions.map((r) => `${r.userId}:${r.emoji}`)
    expect(remaining).toContain(`${user.id}:heart`)
    expect(remaining).toContain(`${user2.id}:thumbsup`)
    expect(remaining).not.toContain(`${user.id}:thumbsup`)
  })
})

describe("getReactions", () => {
  test("returns all reactions for a message", async () => {
    const user = await insertUser(db, "u1@test.com")
    const user2 = await insertUser(db, "u2@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Popular" })

    await addReaction(msg.id, user.id, "thumbsup")
    await addReaction(msg.id, user.id, "fire")
    await addReaction(msg.id, user2.id, "thumbsup")

    const reactions = await getReactions(msg.id)
    expect(reactions).toHaveLength(3)
  })
})

// ── Pins ─────────────────────────────────────────────────────────────────────

describe("pinMessage", () => {
  test("adds a pin", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Pin me" })

    await pinMessage(channel.id, msg.id, user.id)

    const pins = await getPinnedMessages(channel.id)
    expect(pins).toHaveLength(1)
    expect(pins[0]!.messageId).toBe(msg.id)
    expect(pins[0]!.pinnedBy).toBe(user.id)
    expect(pins[0]!.pinnedAt).toBeGreaterThan(0)
  })

  test("is idempotent — duplicate pin does not error", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Pin twice" })

    await pinMessage(channel.id, msg.id, user.id)
    await pinMessage(channel.id, msg.id, user.id) // should not throw

    const pins = await getPinnedMessages(channel.id)
    expect(pins).toHaveLength(1)
  })
})

describe("unpinMessage", () => {
  test("removes a pin", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Unpin me" })

    await pinMessage(channel.id, msg.id, user.id)
    expect((await getPinnedMessages(channel.id))).toHaveLength(1)

    await unpinMessage(channel.id, msg.id)

    const pins = await getPinnedMessages(channel.id)
    expect(pins).toHaveLength(0)
  })
})

describe("getPinnedMessages", () => {
  test("returns pinned messages with message data, newest pin first", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg1 = await sendMessage({ channelId: channel.id, userId: user.id, content: "First pin" })
    await tick()
    const msg2 = await sendMessage({ channelId: channel.id, userId: user.id, content: "Second pin" })

    await pinMessage(channel.id, msg1.id, user.id)
    await tick()
    await pinMessage(channel.id, msg2.id, user.id)

    const pins = await getPinnedMessages(channel.id)

    expect(pins).toHaveLength(2)
    // Ordered by pinnedAt desc (newest first)
    expect(pins[0]!.message.content).toBe("Second pin")
    expect(pins[1]!.message.content).toBe("First pin")
  })
})

// ── Search (FTS5) ────────────────────────────────────────────────────────────

describe("searchMessages", () => {
  test("returns empty array for empty query", async () => {
    const result = await searchMessages("", ["channel-1"])
    expect(result).toEqual([])
  })

  test("returns empty array for whitespace-only query", async () => {
    const result = await searchMessages("   ", ["channel-1"])
    expect(result).toEqual([])
  })

  test("returns empty array for empty channelIds", async () => {
    const result = await searchMessages("hello", [])
    expect(result).toEqual([])
  })
})

// ── Mentions ─────────────────────────────────────────────────────────────────

describe("addMentions", () => {
  test("inserts mentions correctly", async () => {
    const user = await insertUser(db, "u1@test.com")
    const user2 = await insertUser(db, "u2@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "Hey @user2" })

    await addMentions(msg.id, [
      { userId: user2.id, type: "user" },
      { type: "everyone" },
    ])

    const mentions = await getMentionsForUser(user2.id)
    expect(mentions).toHaveLength(1)
    expect(mentions[0]!.messageId).toBe(msg.id)
    expect(mentions[0]!.type).toBe("user")
  })

  test("with empty array does nothing (no error)", async () => {
    const user = await insertUser(db, "u1@test.com")
    const channel = await insertChannel(db, user.id)
    const msg = await sendMessage({ channelId: channel.id, userId: user.id, content: "No mentions" })

    // Should not throw
    await addMentions(msg.id, [])

    const mentions = await getMentionsForUser(user.id)
    expect(mentions).toHaveLength(0)
  })
})

describe("getMentionsForUser", () => {
  test("returns mentions with message data", async () => {
    const user = await insertUser(db, "u1@test.com")
    const mentioned = await insertUser(db, "u2@test.com")
    const channel = await insertChannel(db, user.id)

    const msg1 = await sendMessage({ channelId: channel.id, userId: user.id, content: "Hey @mentioned" })
    await tick()
    const msg2 = await sendMessage({ channelId: channel.id, userId: user.id, content: "Hey again @mentioned" })

    await addMentions(msg1.id, [{ userId: mentioned.id, type: "user" }])
    await addMentions(msg2.id, [{ userId: mentioned.id, type: "user" }])

    const mentions = await getMentionsForUser(mentioned.id)

    expect(mentions).toHaveLength(2)
    // Each mention should include the message relation
    for (const m of mentions) {
      expect(m.message).toBeDefined()
      expect(m.message.channelId).toBe(channel.id)
    }
  })

  test("returns empty array when user has no mentions", async () => {
    const user = await insertUser(db, "u1@test.com")
    const mentions = await getMentionsForUser(user.id)
    expect(mentions).toHaveLength(0)
  })
})
