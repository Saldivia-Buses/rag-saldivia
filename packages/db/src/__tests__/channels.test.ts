/**
 * Tests for channel queries — CRUD, members, unread counts.
 *
 * Covers: createChannel, getChannel, getChannelsByUser, updateChannel,
 * archiveChannel, addChannelMember, removeChannelMember, getChannelMembers,
 * updateLastRead, getUnreadCounts.
 *
 * Runs with: bun test packages/db/src/__tests__/channels.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertChannel } from "./setup"
import {
  createChannel,
  getChannel,
  getChannelsByUser,
  updateChannel,
  archiveChannel,
  addChannelMember,
  removeChannelMember,
  getChannelMembers,
  updateLastRead,
  getUnreadCounts,
} from "../queries/channels"
import { channelMembers, msgMessages } from "../schema"

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
    "DELETE FROM msg_messages; DELETE FROM channel_members; DELETE FROM channels; DELETE FROM users;"
  )
})

// ── Channel CRUD ─────────────────────────────────────────────────────────────

describe("createChannel", () => {
  test("creates channel and adds creator as owner member", async () => {
    const user = await insertUser(db, "creator@test.com")

    const channel = await createChannel({
      type: "public",
      name: "general",
      description: "Main channel",
      createdBy: user.id,
    })

    expect(channel).toBeDefined()
    expect(channel.id).toBeString()
    expect(channel.type).toBe("public")
    expect(channel.name).toBe("general")
    expect(channel.description).toBe("Main channel")
    expect(channel.createdBy).toBe(user.id)
    expect(channel.createdAt).toBeNumber()
    expect(channel.updatedAt).toBeNumber()
    expect(channel.archivedAt).toBeNull()

    // Creator should be added as owner
    const full = await getChannel(channel.id)
    expect(full).toBeDefined()
    expect(full!.channelMembers).toHaveLength(1)
    expect(full!.channelMembers[0]!.userId).toBe(user.id)
    expect(full!.channelMembers[0]!.role).toBe("owner")
  })

  test("adds memberIds without duplicating creator", async () => {
    const creator = await insertUser(db, "creator@test.com")
    const member1 = await insertUser(db, "member1@test.com")
    const member2 = await insertUser(db, "member2@test.com")

    const channel = await createChannel({
      type: "private",
      name: "team",
      createdBy: creator.id,
      // Include creator in memberIds — should not duplicate
      memberIds: [creator.id, member1.id, member2.id],
    })

    const full = await getChannel(channel.id)
    expect(full).toBeDefined()
    // creator (owner) + member1 + member2 = 3 (creator not duplicated)
    expect(full!.channelMembers).toHaveLength(3)

    const roles = full!.channelMembers.reduce(
      (acc, m) => {
        acc[m.userId] = m.role
        return acc
      },
      {} as Record<number, string>
    )
    expect(roles[creator.id]).toBe("owner")
    expect(roles[member1.id]).toBe("member")
    expect(roles[member2.id]).toBe("member")
  })

  test("creates DM channel with no name", async () => {
    const user1 = await insertUser(db, "user1@test.com")
    const user2 = await insertUser(db, "user2@test.com")

    const channel = await createChannel({
      type: "dm",
      createdBy: user1.id,
      memberIds: [user2.id],
    })

    expect(channel.type).toBe("dm")
    expect(channel.name).toBeNull()
    expect(channel.description).toBeNull()

    const full = await getChannel(channel.id)
    expect(full!.channelMembers).toHaveLength(2)
  })
})

describe("getChannel", () => {
  test("returns channel with channelMembers populated", async () => {
    const user = await insertUser(db, "user@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user.id,
    })

    const result = await getChannel(channel.id)

    expect(result).toBeDefined()
    expect(result!.id).toBe(channel.id)
    expect(result!.name).toBe("test")
    expect(result!.channelMembers).toBeArray()
    expect(result!.channelMembers).toHaveLength(1)
    expect(result!.channelMembers[0]!.userId).toBe(user.id)
  })

  test("returns undefined for non-existent channel", async () => {
    const result = await getChannel("non-existent-uuid")
    expect(result).toBeUndefined()
  })
})

describe("getChannelsByUser", () => {
  test("returns only channels where user is member", async () => {
    const user1 = await insertUser(db, "user1@test.com")
    const user2 = await insertUser(db, "user2@test.com")

    // user1 creates channel A
    const chA = await createChannel({
      type: "public",
      name: "channel-a",
      createdBy: user1.id,
    })

    // user2 creates channel B
    await createChannel({
      type: "public",
      name: "channel-b",
      createdBy: user2.id,
    })

    // user1 should only see channel A
    const result = await getChannelsByUser(user1.id)
    expect(result).toHaveLength(1)
    expect(result[0]!.id).toBe(chA.id)
  })

  test("returns empty array for user with no channels", async () => {
    const user = await insertUser(db, "lonely@test.com")

    const result = await getChannelsByUser(user.id)
    expect(result).toEqual([])
  })

  test("returns channels ordered by updatedAt descending", async () => {
    const user = await insertUser(db, "user@test.com")

    const older = await createChannel({
      type: "public",
      name: "older",
      createdBy: user.id,
    })

    // Small delay to guarantee different timestamp
    await new Promise((r) => setTimeout(r, 10))

    const newer = await createChannel({
      type: "public",
      name: "newer",
      createdBy: user.id,
    })

    const result = await getChannelsByUser(user.id)
    expect(result).toHaveLength(2)
    expect(result[0]!.id).toBe(newer.id)
    expect(result[1]!.id).toBe(older.id)
  })
})

describe("updateChannel", () => {
  test("updates fields and sets updatedAt", async () => {
    const user = await insertUser(db, "user@test.com")
    const channel = await createChannel({
      type: "public",
      name: "original",
      createdBy: user.id,
    })

    // Small delay to guarantee different updatedAt
    await new Promise((r) => setTimeout(r, 10))

    const updated = await updateChannel(channel.id, {
      name: "renamed",
      description: "new description",
      topic: "new topic",
    })

    expect(updated).toBeDefined()
    expect(updated!.name).toBe("renamed")
    expect(updated!.description).toBe("new description")
    expect(updated!.topic).toBe("new topic")
    expect(updated!.updatedAt).toBeGreaterThan(channel.updatedAt)
  })

  test("returns undefined for non-existent channel", async () => {
    const updated = await updateChannel("does-not-exist", { name: "nope" })
    expect(updated).toBeUndefined()
  })
})

describe("archiveChannel", () => {
  test("sets archivedAt timestamp", async () => {
    const user = await insertUser(db, "user@test.com")
    const channel = await createChannel({
      type: "public",
      name: "to-archive",
      createdBy: user.id,
    })

    expect(channel.archivedAt).toBeNull()

    const archived = await archiveChannel(channel.id)

    expect(archived).toBeDefined()
    expect(archived!.archivedAt).toBeNumber()
    expect(archived!.archivedAt).toBeGreaterThan(0)
    expect(archived!.updatedAt).toBeGreaterThanOrEqual(archived!.archivedAt!)
  })
})

// ── Members ──────────────────────────────────────────────────────────────────

describe("addChannelMember", () => {
  test("adds member correctly", async () => {
    const creator = await insertUser(db, "creator@test.com")
    const newMember = await insertUser(db, "new@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: creator.id,
    })

    await addChannelMember(channel.id, newMember.id, "member")

    const full = await getChannel(channel.id)
    expect(full!.channelMembers).toHaveLength(2)

    const added = full!.channelMembers.find((m) => m.userId === newMember.id)
    expect(added).toBeDefined()
    expect(added!.role).toBe("member")
    expect(added!.muted).toBe(false)
  })

  test("adds member with admin role", async () => {
    const creator = await insertUser(db, "creator@test.com")
    const admin = await insertUser(db, "admin@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: creator.id,
    })

    await addChannelMember(channel.id, admin.id, "admin")

    const full = await getChannel(channel.id)
    const added = full!.channelMembers.find((m) => m.userId === admin.id)
    expect(added!.role).toBe("admin")
  })

  test("is idempotent — does not throw on duplicate", async () => {
    const creator = await insertUser(db, "creator@test.com")
    const member = await insertUser(db, "member@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: creator.id,
    })

    await addChannelMember(channel.id, member.id, "member")
    // Second call should not throw
    await addChannelMember(channel.id, member.id, "admin")

    // Original role should be preserved (onConflictDoNothing)
    const full = await getChannel(channel.id)
    const m = full!.channelMembers.find((cm) => cm.userId === member.id)
    expect(m!.role).toBe("member") // original, not overwritten
  })
})

describe("removeChannelMember", () => {
  test("removes member correctly", async () => {
    const creator = await insertUser(db, "creator@test.com")
    const member = await insertUser(db, "member@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: creator.id,
      memberIds: [member.id],
    })

    // Verify member exists
    let full = await getChannel(channel.id)
    expect(full!.channelMembers).toHaveLength(2)

    await removeChannelMember(channel.id, member.id)

    full = await getChannel(channel.id)
    expect(full!.channelMembers).toHaveLength(1)
    expect(full!.channelMembers[0]!.userId).toBe(creator.id)
  })

  test("does not throw when removing non-existent member", async () => {
    const user = await insertUser(db, "user@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user.id,
    })

    // Should not throw
    await removeChannelMember(channel.id, 99999)

    const full = await getChannel(channel.id)
    expect(full!.channelMembers).toHaveLength(1)
  })
})

describe("getChannelMembers", () => {
  test("returns members with user info", async () => {
    const creator = await insertUser(db, "creator@test.com")
    const member = await insertUser(db, "member@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: creator.id,
      memberIds: [member.id],
    })

    const members = await getChannelMembers(channel.id)

    expect(members).toHaveLength(2)
    // Each member should have the nested user object from the `with: { user: true }` relation
    for (const m of members) {
      expect(m.user).toBeDefined()
      expect(m.user.email).toBeString()
      expect(m.user.name).toBe("Test User")
    }

    const emails = members.map((m) => m.user.email).sort()
    expect(emails).toEqual(["creator@test.com", "member@test.com"])
  })

  test("returns empty array for non-existent channel", async () => {
    const members = await getChannelMembers("no-such-channel")
    expect(members).toEqual([])
  })
})

// ── Unread counts ────────────────────────────────────────────────────────────

describe("updateLastRead", () => {
  test("updates the lastReadAt timestamp", async () => {
    const user = await insertUser(db, "user@test.com")
    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user.id,
    })

    // Get initial lastReadAt
    const before = await getChannel(channel.id)
    const initialLastRead = before!.channelMembers[0]!.lastReadAt

    // Small delay to guarantee different timestamp
    await new Promise((r) => setTimeout(r, 10))

    await updateLastRead(channel.id, user.id)

    const after = await getChannel(channel.id)
    const updatedLastRead = after!.channelMembers[0]!.lastReadAt

    expect(updatedLastRead).toBeGreaterThan(initialLastRead)
  })
})

describe("getUnreadCounts", () => {
  test("returns correct unread counts per channel", async () => {
    const user1 = await insertUser(db, "user1@test.com")
    const user2 = await insertUser(db, "user2@test.com")

    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user1.id,
      memberIds: [user2.id],
    })

    // user2 sends 3 messages after user1's lastReadAt
    const futureTs = Date.now() + 10000
    for (let i = 0; i < 3; i++) {
      await db.insert(msgMessages).values({
        id: crypto.randomUUID(),
        channelId: channel.id,
        userId: user2.id,
        content: `Message ${i}`,
        type: "text",
        replyCount: 0,
        createdAt: futureTs + i,
      })
    }

    const counts = await getUnreadCounts(user1.id)
    expect(counts[channel.id]).toBe(3)
  })

  test("excludes own messages from unread count", async () => {
    const user1 = await insertUser(db, "user1@test.com")

    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user1.id,
    })

    // user1 sends messages to own channel — should not count as unread
    const futureTs = Date.now() + 10000
    await db.insert(msgMessages).values({
      id: crypto.randomUUID(),
      channelId: channel.id,
      userId: user1.id,
      content: "My own message",
      type: "text",
      replyCount: 0,
      createdAt: futureTs,
    })

    const counts = await getUnreadCounts(user1.id)
    expect(counts[channel.id]).toBe(0)
  })

  test("excludes deleted messages from unread count", async () => {
    const user1 = await insertUser(db, "user1@test.com")
    const user2 = await insertUser(db, "user2@test.com")

    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user1.id,
      memberIds: [user2.id],
    })

    const futureTs = Date.now() + 10000

    // One normal message
    await db.insert(msgMessages).values({
      id: crypto.randomUUID(),
      channelId: channel.id,
      userId: user2.id,
      content: "Normal message",
      type: "text",
      replyCount: 0,
      createdAt: futureTs,
    })

    // One deleted message
    await db.insert(msgMessages).values({
      id: crypto.randomUUID(),
      channelId: channel.id,
      userId: user2.id,
      content: "Deleted message",
      type: "text",
      replyCount: 0,
      deletedAt: futureTs + 1,
      createdAt: futureTs,
    })

    const counts = await getUnreadCounts(user1.id)
    expect(counts[channel.id]).toBe(1)
  })

  test("returns empty object for user with no channels", async () => {
    const user = await insertUser(db, "lonely@test.com")

    const counts = await getUnreadCounts(user.id)
    expect(counts).toEqual({})
  })

  test("returns zero for channels with no new messages", async () => {
    const user1 = await insertUser(db, "user1@test.com")
    const user2 = await insertUser(db, "user2@test.com")

    const channel = await createChannel({
      type: "public",
      name: "test",
      createdBy: user1.id,
      memberIds: [user2.id],
    })

    // Add a message BEFORE user1's lastReadAt (should not be counted)
    const pastTs = 1000 // way before the channel creation timestamp
    await db.insert(msgMessages).values({
      id: crypto.randomUUID(),
      channelId: channel.id,
      userId: user2.id,
      content: "Old message",
      type: "text",
      replyCount: 0,
      createdAt: pastTs,
    })

    const counts = await getUnreadCounts(user1.id)
    expect(counts[channel.id]).toBe(0)
  })

  test("counts across multiple channels correctly", async () => {
    const user1 = await insertUser(db, "user1@test.com")
    const user2 = await insertUser(db, "user2@test.com")

    const chA = await createChannel({
      type: "public",
      name: "channel-a",
      createdBy: user1.id,
      memberIds: [user2.id],
    })

    const chB = await createChannel({
      type: "public",
      name: "channel-b",
      createdBy: user1.id,
      memberIds: [user2.id],
    })

    const futureTs = Date.now() + 10000

    // 2 messages in channel A
    for (let i = 0; i < 2; i++) {
      await db.insert(msgMessages).values({
        id: crypto.randomUUID(),
        channelId: chA.id,
        userId: user2.id,
        content: `A-${i}`,
        type: "text",
        replyCount: 0,
        createdAt: futureTs + i,
      })
    }

    // 5 messages in channel B
    for (let i = 0; i < 5; i++) {
      await db.insert(msgMessages).values({
        id: crypto.randomUUID(),
        channelId: chB.id,
        userId: user2.id,
        content: `B-${i}`,
        type: "text",
        replyCount: 0,
        createdAt: futureTs + i,
      })
    }

    const counts = await getUnreadCounts(user1.id)
    expect(counts[chA.id]).toBe(2)
    expect(counts[chB.id]).toBe(5)
  })
})
