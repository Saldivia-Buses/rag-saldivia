import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { CommandPalette } from "../CommandPalette"

afterEach(cleanup)

const CHANNELS = [
  { id: "ch-1", name: "general", type: "public" },
  { id: "ch-2", name: "privado", type: "private" },
]

describe("CommandPalette", () => {
  test("renders nothing when closed", () => {
    const { container } = render(
      <CommandPalette open={false} onClose={() => {}} channels={CHANNELS} isAdmin={false} />
    )
    // cmdk CommandDialog should not render content when closed
    expect(container.querySelectorAll("[role='dialog']").length).toBe(0)
  })

  test("renders dialog when open", () => {
    const { getByPlaceholderText } = render(
      <CommandPalette open={true} onClose={() => {}} channels={CHANNELS} isAdmin={false} />
    )
    expect(getByPlaceholderText("Buscar canales, mensajes...")).toBeTruthy()
  })

  test("renders channel names", () => {
    const { getByText } = render(
      <CommandPalette open={true} onClose={() => {}} channels={CHANNELS} isAdmin={false} />
    )
    expect(getByText("general")).toBeTruthy()
    expect(getByText("privado")).toBeTruthy()
  })

  test("renders admin link when isAdmin", () => {
    const { getByText } = render(
      <CommandPalette open={true} onClose={() => {}} channels={CHANNELS} isAdmin={true} />
    )
    expect(getByText("Admin")).toBeTruthy()
  })
})
