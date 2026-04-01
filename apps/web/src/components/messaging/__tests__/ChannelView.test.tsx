import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { ChannelView } from "../ChannelView"

afterEach(cleanup)

const MEMBERS = [{ id: 1, name: "Admin", email: "admin@test.com" }]
const MESSAGES = [
  { id: "m-1", userId: 1, content: "Hello world", type: "text", createdAt: Date.now(), editedAt: null, deletedAt: null, replyCount: 0, parentId: null },
]

describe("ChannelView", () => {
  test("renders without crash", () => {
    const { container } = render(
      <ChannelView channelId="ch-1" initialMessages={MESSAGES} currentUserId={1} members={MEMBERS} />
    )
    expect(container.firstChild).toBeTruthy()
  })

  test("renders message content", () => {
    const { getByText } = render(
      <ChannelView channelId="ch-1" initialMessages={MESSAGES} currentUserId={1} members={MEMBERS} />
    )
    expect(getByText("Hello world")).toBeTruthy()
  })

  test("renders with empty messages", () => {
    const { container } = render(
      <ChannelView channelId="ch-1" initialMessages={[]} currentUserId={1} members={MEMBERS} />
    )
    expect(container.firstChild).toBeTruthy()
  })
})
