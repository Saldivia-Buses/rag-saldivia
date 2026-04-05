import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { MessageItem, DateSeparator } from "../MessageItem"

afterEach(cleanup)

const MEMBERS = [
  { id: 1, name: "Admin User", email: "admin@test.com" },
  { id: 2, name: "María López", email: "maria@test.com" },
]

const mkMsg = (overrides: Partial<Parameters<typeof MessageItem>[0]["message"]> = {}) => ({
  id: "m-1", userId: 1, content: "Hello world", type: "text",
  createdAt: Date.now(), editedAt: null, deletedAt: null, replyCount: 0, parentId: null,
  ...overrides,
})

describe("MessageItem", () => {
  test("renders message content", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg()} members={MEMBERS} showHeader currentUserId={1} />
    )
    expect(getByText(/Hello world/)).toBeTruthy()
  })

  test("renders author name when showHeader is true", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg()} members={MEMBERS} showHeader currentUserId={1} />
    )
    expect(getByText("Admin User")).toBeTruthy()
  })

  test("hides author name when showHeader is false", () => {
    const { queryByText } = render(
      <MessageItem message={mkMsg()} members={MEMBERS} showHeader={false} currentUserId={1} />
    )
    expect(queryByText("Admin User")).toBeNull()
  })

  test("shows edited badge when editedAt is set", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg({ editedAt: Date.now() })} members={MEMBERS} showHeader currentUserId={1} />
    )
    expect(getByText("(editado)")).toBeTruthy()
  })

  test("shows deleted placeholder", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg({ deletedAt: Date.now() })} members={MEMBERS} showHeader currentUserId={1} />
    )
    expect(getByText("Mensaje eliminado")).toBeTruthy()
  })

  test("shows system message centered", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg({ type: "system", content: "Se unió al canal" })} members={MEMBERS} showHeader currentUserId={1} />
    )
    expect(getByText("Se unió al canal")).toBeTruthy()
  })

  test("shows reply count as button", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg({ replyCount: 3 })} members={MEMBERS} showHeader currentUserId={1} onOpenThread={() => {}} />
    )
    expect(getByText("3 respuestas")).toBeTruthy()
  })

  test("shows singular reply count", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg({ replyCount: 1 })} members={MEMBERS} showHeader currentUserId={1} onOpenThread={() => {}} />
    )
    expect(getByText("1 respuesta")).toBeTruthy()
  })

  test("shows fallback name for unknown user", () => {
    const { getByText } = render(
      <MessageItem message={mkMsg({ userId: 999 })} members={MEMBERS} showHeader currentUserId={1} />
    )
    expect(getByText("Usuario")).toBeTruthy()
  })
})

describe("DateSeparator", () => {
  test("renders Hoy for today", () => {
    const { getByText } = render(<DateSeparator timestamp={Date.now()} />)
    expect(getByText("Hoy")).toBeTruthy()
  })

  test("renders Ayer for yesterday", () => {
    const yesterday = Date.now() - 86400_000
    const { getByText } = render(<DateSeparator timestamp={yesterday} />)
    expect(getByText("Ayer")).toBeTruthy()
  })
})
