/**
 * Tests for small messaging components (< 60 lines each).
 * TypingIndicator, UnreadBadge, PresenceIndicator, FileAttachment, VoiceInput
 */
import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { TypingIndicator } from "../TypingIndicator"
import { UnreadBadge } from "../UnreadBadge"

afterEach(cleanup)

// ── TypingIndicator ──

describe("TypingIndicator", () => {
  test("renders nothing when no one is typing", () => {
    const { container } = render(<TypingIndicator typingUsers={[]} />)
    expect(container.innerHTML).toBe("")
  })

  test("renders single user typing", () => {
    const { getByText } = render(
      <TypingIndicator typingUsers={[{ userId: 1, displayName: "Admin" }]} />
    )
    expect(getByText(/Admin está escribiendo/)).toBeTruthy()
  })

  test("renders two users typing", () => {
    const { getByText } = render(
      <TypingIndicator typingUsers={[
        { userId: 1, displayName: "Admin" },
        { userId: 2, displayName: "María" },
      ]} />
    )
    expect(getByText(/Admin y María están escribiendo/)).toBeTruthy()
  })

  test("renders 3+ users typing with count", () => {
    const { getByText } = render(
      <TypingIndicator typingUsers={[
        { userId: 1, displayName: "Admin" },
        { userId: 2, displayName: "María" },
        { userId: 3, displayName: "Juan" },
      ]} />
    )
    expect(getByText(/Admin y 2 más están escribiendo/)).toBeTruthy()
  })
})

// ── UnreadBadge ──

describe("UnreadBadge", () => {
  test("renders nothing when count is 0", () => {
    const { container } = render(<UnreadBadge count={0} />)
    expect(container.innerHTML).toBe("")
  })

  test("renders count", () => {
    const { getByText } = render(<UnreadBadge count={5} />)
    expect(getByText("5")).toBeTruthy()
  })

  test("renders 99+ for large counts", () => {
    const { getByText } = render(<UnreadBadge count={150} />)
    expect(getByText("99+")).toBeTruthy()
  })

  test("renders nothing for negative count", () => {
    const { container } = render(<UnreadBadge count={-1} />)
    expect(container.innerHTML).toBe("")
  })
})
