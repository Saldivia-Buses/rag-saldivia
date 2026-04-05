/**
 * Tests for chat server actions (next-safe-action).
 *
 * Mocks: next/headers, next/cache, @rag-saldivia/db, @rag-saldivia/logger
 * Runs with: bun test apps/web/src/__tests__/actions/chat.test.ts
 */

import { describe, test, expect, mock, beforeEach } from "bun:test"

// ── Mocks (must be declared before importing the actions) ────────────────

const mockCreateSession = mock(() =>
  Promise.resolve({ id: "sess-1", title: "Test", collection: "default", crossdoc: false })
)
const mockUpdateSessionTitle = mock(() =>
  Promise.resolve({ id: "sess-1", title: "Renamed" })
)
const mockDeleteSession = mock(() => Promise.resolve())
const mockAddMessage = mock(() =>
  Promise.resolve({ id: 1, sessionId: "sess-1", role: "user", content: "hello" })
)
const mockAddFeedback = mock(() => Promise.resolve())
const mockSaveResponse = mock(() => Promise.resolve())
const mockUnsaveByMessageId = mock(() => Promise.resolve())
const mockSaveAnnotation = mock(() => Promise.resolve())
const mockAddTag = mock(() => Promise.resolve())
const mockRemoveTag = mock(() => Promise.resolve())
const mockGetSessionById = mock(() =>
  Promise.resolve({ id: "sess-1", title: "Original", collection: "default", crossdoc: false })
)

// getDb mock — returns a chainable Drizzle-like object for actionForkSession
const mockInsertValues = mock(() => Promise.resolve())
const mockSelectWhereResult: Array<Record<string, unknown>> = []
const mockGetDb = mock(() => ({
  select: () => ({
    from: () => ({
      where: () => Promise.resolve(mockSelectWhereResult),
    }),
  }),
  insert: () => ({
    values: mockInsertValues,
  }),
}))

mock.module("@rag-saldivia/db", () => ({
  createSession: mockCreateSession,
  updateSessionTitle: mockUpdateSessionTitle,
  deleteSession: mockDeleteSession,
  addMessage: mockAddMessage,
  addFeedback: mockAddFeedback,
  saveResponse: mockSaveResponse,
  unsaveByMessageId: mockUnsaveByMessageId,
  saveAnnotation: mockSaveAnnotation,
  addTag: mockAddTag,
  removeTag: mockRemoveTag,
  getSessionById: mockGetSessionById,
  getDb: mockGetDb,
  chatMessages: { sessionId: "sessionId" },
  chatSessions: {},
  touchUserPresence: mock(() => Promise.resolve()),
}))

mock.module("@rag-saldivia/logger/backend", () => ({
  log: { info: () => {}, warn: () => {}, error: () => {} },
}))

// Mock next/headers — simulate authenticated user via x-user-* headers
mock.module("next/headers", () => ({
  headers: () =>
    new Headers({
      "x-user-id": "1",
      "x-user-email": "test@test.com",
      "x-user-name": "Test User",
      "x-user-role": "admin",
    }),
  cookies: () => ({ get: () => null, set: () => {}, delete: () => {} }),
}))

// Mock next/cache
const mockRevalidatePath = mock(() => {})
mock.module("next/cache", () => ({
  revalidatePath: mockRevalidatePath,
}))

// Mock drizzle-orm eq
mock.module("drizzle-orm", () => ({
  eq: (col: unknown, val: unknown) => ({ col, val }),
}))

// ── Import actions AFTER mocks ───────────────────────────────────────────

import {
  actionCreateSession,
  actionRenameSession,
  actionDeleteSession,
  actionAddMessage,
  actionAddFeedback,
  actionToggleSaved,
  actionSaveAnnotation,
  actionAddTag,
  actionRemoveTag,
  actionForkSession,
  actionCreateSessionForDoc,
} from "@/app/actions/chat"

// ── Reset mocks between tests ────────────────────────────────────────────

beforeEach(() => {
  mockCreateSession.mockClear()
  mockUpdateSessionTitle.mockClear()
  mockDeleteSession.mockClear()
  mockAddMessage.mockClear()
  mockAddFeedback.mockClear()
  mockSaveResponse.mockClear()
  mockUnsaveByMessageId.mockClear()
  mockSaveAnnotation.mockClear()
  mockAddTag.mockClear()
  mockRemoveTag.mockClear()
  mockRevalidatePath.mockClear()
  mockGetSessionById.mockClear()
  mockInsertValues.mockClear()
})

// ── Tests ────────────────────────────────────────────────────────────────

describe("actionCreateSession", () => {
  test("valid input returns session data via result.data", async () => {
    const result = await actionCreateSession({
      collection: "default",
    })

    expect(result?.data).toBeDefined()
    expect(result?.data?.id).toBe("sess-1")
    expect(result?.data?.collection).toBe("default")
    expect(mockCreateSession).toHaveBeenCalledTimes(1)
  })

  test("calls revalidatePath after creating session", async () => {
    await actionCreateSession({ collection: "default" })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat")
  })

  test("passes userId from context to createSession", async () => {
    await actionCreateSession({ collection: "my-collection" })

    const callArgs = mockCreateSession.mock.calls[0]?.[0] as Record<string, unknown>
    expect(callArgs?.["userId"]).toBe(1)
    expect(callArgs?.["collection"]).toBe("my-collection")
  })

  test("validation error when collection is missing", async () => {
    // @ts-expect-error -- intentionally passing invalid input
    const result = await actionCreateSession({})

    expect(result?.validationErrors).toBeDefined()
    expect(mockCreateSession).not.toHaveBeenCalled()
  })

  test("passes optional crossdoc flag", async () => {
    await actionCreateSession({ collection: "default", crossdoc: true })

    const callArgs = mockCreateSession.mock.calls[0]?.[0] as Record<string, unknown>
    expect(callArgs?.["crossdoc"]).toBe(true)
  })
})

describe("actionRenameSession", () => {
  test("valid input returns updated session", async () => {
    const result = await actionRenameSession({
      id: "sess-1",
      title: "Renamed",
    })

    expect(result?.data).toBeDefined()
    expect(result?.data?.title).toBe("Renamed")
    expect(mockUpdateSessionTitle).toHaveBeenCalledWith("sess-1", 1, "Renamed")
  })

  test("revalidates both /chat and /chat/[id]", async () => {
    await actionRenameSession({ id: "sess-1", title: "New title" })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat")
    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat/sess-1")
  })

  test("validation error when title is missing", async () => {
    // @ts-expect-error -- intentionally passing invalid input
    const result = await actionRenameSession({ id: "sess-1" })

    expect(result?.validationErrors).toBeDefined()
    expect(mockUpdateSessionTitle).not.toHaveBeenCalled()
  })
})

describe("actionDeleteSession", () => {
  test("calls deleteSession with id and userId", async () => {
    await actionDeleteSession({ id: "sess-1" })

    expect(mockDeleteSession).toHaveBeenCalledWith("sess-1", 1)
  })

  test("revalidates /chat path", async () => {
    await actionDeleteSession({ id: "sess-1" })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat")
  })
})

describe("actionAddMessage", () => {
  test("valid input calls addMessage and returns result", async () => {
    const result = await actionAddMessage({
      sessionId: "sess-1",
      role: "user",
      content: "Hello world",
    })

    expect(result?.data).toBeDefined()
    expect(mockAddMessage).toHaveBeenCalledTimes(1)
  })

  test("revalidates the session path", async () => {
    await actionAddMessage({
      sessionId: "sess-1",
      role: "user",
      content: "Hello",
    })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat/sess-1")
  })

  test("validation error for invalid role", async () => {
    const result = await actionAddMessage({
      sessionId: "sess-1",
      // @ts-expect-error -- intentionally passing invalid role
      role: "invalid",
      content: "test",
    })

    expect(result?.validationErrors).toBeDefined()
    expect(mockAddMessage).not.toHaveBeenCalled()
  })

  test("accepts assistant role", async () => {
    await actionAddMessage({
      sessionId: "sess-1",
      role: "assistant",
      content: "I can help with that",
    })

    expect(mockAddMessage).toHaveBeenCalledTimes(1)
  })
})

describe("actionAddFeedback", () => {
  test("calls addFeedback with messageId, userId, and rating", async () => {
    await actionAddFeedback({ messageId: 42, rating: "up" })

    expect(mockAddFeedback).toHaveBeenCalledWith(42, 1, "up")
  })

  test("accepts 'down' rating", async () => {
    await actionAddFeedback({ messageId: 42, rating: "down" })

    expect(mockAddFeedback).toHaveBeenCalledWith(42, 1, "down")
  })
})

describe("actionToggleSaved", () => {
  test("saves response when currentlySaved is false", async () => {
    await actionToggleSaved({
      messageId: 10,
      content: "Useful answer",
      sessionTitle: "My Session",
      currentlySaved: false,
    })

    expect(mockSaveResponse).toHaveBeenCalledTimes(1)
    expect(mockUnsaveByMessageId).not.toHaveBeenCalled()
  })

  test("unsaves when currentlySaved is true", async () => {
    await actionToggleSaved({
      messageId: 10,
      content: "Useful answer",
      sessionTitle: "My Session",
      currentlySaved: true,
    })

    expect(mockUnsaveByMessageId).toHaveBeenCalledWith(10, 1)
    expect(mockSaveResponse).not.toHaveBeenCalled()
  })
})

describe("actionSaveAnnotation", () => {
  test("calls saveAnnotation with user context", async () => {
    await actionSaveAnnotation({
      sessionId: "sess-1",
      selectedText: "important text",
      note: "Remember this",
    })

    expect(mockSaveAnnotation).toHaveBeenCalledTimes(1)
    const args = mockSaveAnnotation.mock.calls[0]?.[0] as Record<string, unknown>
    expect(args?.["userId"]).toBe(1)
    expect(args?.["selectedText"]).toBe("important text")
    expect(args?.["note"]).toBe("Remember this")
  })
})

describe("actionAddTag / actionRemoveTag", () => {
  test("addTag calls DB and revalidates", async () => {
    await actionAddTag({ sessionId: "sess-1", tag: "important" })

    expect(mockAddTag).toHaveBeenCalledWith("sess-1", "important")
    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat")
  })

  test("removeTag calls DB and revalidates", async () => {
    await actionRemoveTag({ sessionId: "sess-1", tag: "old" })

    expect(mockRemoveTag).toHaveBeenCalledWith("sess-1", "old")
    expect(mockRevalidatePath).toHaveBeenCalledWith("/chat")
  })
})

describe("actionCreateSessionForDoc", () => {
  test("creates session and adds system message with doc name", async () => {
    mockCreateSession.mockResolvedValueOnce({
      id: "doc-sess-1",
      title: "Chat: report.pdf",
      collection: "docs",
      crossdoc: false,
    })

    const result = await actionCreateSessionForDoc({
      collection: "docs",
      docName: "report.pdf",
    })

    expect(result?.data?.id).toBe("doc-sess-1")
    expect(mockCreateSession).toHaveBeenCalledTimes(1)
    expect(mockAddMessage).toHaveBeenCalledTimes(1)

    const msgArgs = mockAddMessage.mock.calls[0]?.[0] as Record<string, unknown>
    expect(msgArgs?.["role"]).toBe("system")
    expect((msgArgs?.["content"] as string)).toContain("report.pdf")
  })
})

describe("actionForkSession", () => {
  test("returns null when original session not found", async () => {
    mockGetSessionById.mockResolvedValueOnce(null)

    const result = await actionForkSession({
      sessionId: "nonexistent",
      upToMessageId: 1,
    })

    expect(result?.data).toBeNull()
  })
})
