import { describe, test, expect, afterEach, mock } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { ThreadPanel } from "../ThreadPanel"

afterEach(cleanup)

const MEMBERS = [{ id: 1, name: "Admin", email: "admin@test.com" }]
const PARENT = {
  id: "m-1", userId: 1, content: "Original message", type: "text",
  createdAt: Date.now(), editedAt: null, deletedAt: null, replyCount: 2, parentId: null,
}

describe("ThreadPanel", () => {
  test("renders panel with close button", () => {
    const { getByText } = render(
      <ThreadPanel channelId="ch-1" parentMessage={PARENT} replies={[]} members={MEMBERS} currentUserId={1} onClose={() => {}} />
    )
    expect(getByText("Hilo")).toBeTruthy()
  })

  test("renders parent message content", () => {
    const { getByText } = render(
      <ThreadPanel channelId="ch-1" parentMessage={PARENT} replies={[]} members={MEMBERS} currentUserId={1} onClose={() => {}} />
    )
    expect(getByText(/Original message/)).toBeTruthy()
  })

  test("renders reply composer", () => {
    const { getByPlaceholderText } = render(
      <ThreadPanel channelId="ch-1" parentMessage={PARENT} replies={[]} members={MEMBERS} currentUserId={1} onClose={() => {}} />
    )
    expect(getByPlaceholderText("Responder en hilo...")).toBeTruthy()
  })

  test("calls onClose when close button clicked", () => {
    const onClose = mock(() => {})
    const { getByText } = render(
      <ThreadPanel channelId="ch-1" parentMessage={PARENT} replies={[]} members={MEMBERS} currentUserId={1} onClose={onClose} />
    )
    const heading = getByText("Hilo")
    const closeBtn = heading.parentElement!.querySelector("button")
    expect(closeBtn).not.toBeNull()
    fireEvent.click(closeBtn!)
    expect(onClose).toHaveBeenCalled()
  })
})
