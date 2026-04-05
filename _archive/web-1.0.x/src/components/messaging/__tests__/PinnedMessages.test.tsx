import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { PinnedMessages } from "../PinnedMessages"

afterEach(cleanup)

const MEMBERS = [{ id: 1, name: "Admin" }]
const PINNED = [{
  channelId: "ch-1", messageId: "m-1", pinnedBy: 1, pinnedAt: Date.now(),
  message: { id: "m-1", content: "Important message", userId: 1, createdAt: Date.now() },
}]

describe("PinnedMessages", () => {
  test("renders pinned message content", () => {
    const { getByText } = render(
      <PinnedMessages pinnedMessages={PINNED} members={MEMBERS} onClose={() => {}} />
    )
    expect(getByText("Important message")).toBeTruthy()
  })

  test("renders empty state when no pins", () => {
    const { getByText } = render(
      <PinnedMessages pinnedMessages={[]} members={MEMBERS} onClose={() => {}} />
    )
    expect(getByText(/No hay mensajes fijados/)).toBeTruthy()
  })

  test("renders author name", () => {
    const { getByText } = render(
      <PinnedMessages pinnedMessages={PINNED} members={MEMBERS} onClose={() => {}} />
    )
    expect(getByText("Admin")).toBeTruthy()
  })
})
