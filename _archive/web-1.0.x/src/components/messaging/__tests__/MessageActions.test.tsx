import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { MessageActions } from "../MessageActions"

afterEach(cleanup)

const MSG = { id: "m-1", userId: 1, content: "Hello world", channelId: "ch-1" }

describe("MessageActions", () => {
  test("renders copy button", () => {
    const { getByText } = render(
      <MessageActions message={MSG} currentUserId={1} isAdmin={false} onClose={() => {}} />
    )
    expect(getByText("Copiar texto")).toBeTruthy()
  })

  test("shows edit button for own messages", () => {
    const { getByText } = render(
      <MessageActions message={MSG} currentUserId={1} isAdmin={false} onClose={() => {}} />
    )
    expect(getByText("Editar")).toBeTruthy()
  })

  test("hides edit button for other users messages", () => {
    const { queryByText } = render(
      <MessageActions message={MSG} currentUserId={2} isAdmin={false} onClose={() => {}} />
    )
    expect(queryByText("Editar")).toBeNull()
  })

  test("shows delete for own messages", () => {
    const { getByText } = render(
      <MessageActions message={MSG} currentUserId={1} isAdmin={false} onClose={() => {}} />
    )
    expect(getByText("Eliminar")).toBeTruthy()
  })

  test("shows delete for admin even on other users messages", () => {
    const { getByText } = render(
      <MessageActions message={MSG} currentUserId={2} isAdmin={true} onClose={() => {}} />
    )
    expect(getByText("Eliminar")).toBeTruthy()
  })

  test("hides delete for non-owner non-admin", () => {
    const { queryByText } = render(
      <MessageActions message={MSG} currentUserId={2} isAdmin={false} onClose={() => {}} />
    )
    expect(queryByText("Eliminar")).toBeNull()
  })

  test("renders reply button when onReply provided", () => {
    const { getByText } = render(
      <MessageActions message={MSG} currentUserId={1} isAdmin={false} onClose={() => {}} onReply={() => {}} />
    )
    expect(getByText("Responder en hilo")).toBeTruthy()
  })
})
