import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { MessageList } from "../MessageList"

afterEach(cleanup)

const MEMBERS = [{ id: 1, name: "Admin", email: "admin@test.com" }]
const mkMsg = (id: string, content: string, userId = 1, createdAt = Date.now()) => ({
  id, userId, content, type: "text" as const, createdAt, editedAt: null, deletedAt: null, replyCount: 0, parentId: null,
})

describe("MessageList", () => {
  test("renders empty state when no messages", () => {
    const { getByText } = render(
      <MessageList channelId="ch-1" initialMessages={[]} currentUserId={1} members={MEMBERS} />
    )
    expect(getByText(/Sé el primero/)).toBeTruthy()
  })

  test("renders message content", () => {
    const { getByText } = render(
      <MessageList channelId="ch-1" initialMessages={[mkMsg("1", "Hola mundo")]} currentUserId={1} members={MEMBERS} />
    )
    expect(getByText(/Hola mundo/)).toBeTruthy()
  })

  test("renders multiple messages", () => {
    const msgs = [mkMsg("1", "Primero"), mkMsg("2", "Segundo")]
    const { getByText } = render(
      <MessageList channelId="ch-1" initialMessages={msgs} currentUserId={1} members={MEMBERS} />
    )
    expect(getByText(/Primero/)).toBeTruthy()
    expect(getByText(/Segundo/)).toBeTruthy()
  })

  test("renders author name in header", () => {
    const { getByText } = render(
      <MessageList channelId="ch-1" initialMessages={[mkMsg("1", "Test")]} currentUserId={1} members={MEMBERS} />
    )
    expect(getByText("Admin")).toBeTruthy()
  })
})
