import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { SessionList } from "@/components/chat/SessionList"

afterEach(cleanup)

mock.module("@/app/actions/chat", () => ({
  actionDeleteSession: mock(() => Promise.resolve()),
  actionCreateSession: mock(() => Promise.resolve("new-id")),
}))

const sessions = [
  { id: "s1", userId: 1, title: "Primera conversación", collection: "contratos", crossdoc: 0, forkedFrom: null, createdAt: Date.now(), updatedAt: Date.now(), messages: [] },
  { id: "s2", userId: 1, title: "Segunda sesión", collection: "politicas", crossdoc: 0, forkedFrom: null, createdAt: Date.now(), updatedAt: Date.now(), messages: [] },
]

describe("<SessionList />", () => {
  test("renderiza lista de sesiones", () => {
    const { getByText } = render(<SessionList sessions={sessions} />)
    expect(getByText("Primera conversación")).toBeInTheDocument()
    expect(getByText("Segunda sesión")).toBeInTheDocument()
  })

  test("muestra encabezado Chats", () => {
    const { getByText } = render(<SessionList sessions={sessions} />)
    expect(getByText("Chats")).toBeInTheDocument()
  })

  test("sin sesiones muestra mensaje vacío", () => {
    const { getByText } = render(<SessionList sessions={[]} />)
    expect(getByText("Sin chats todavía")).toBeInTheDocument()
  })

  test("botón Nuevo chat presente", () => {
    const { getByTitle } = render(<SessionList sessions={sessions} />)
    expect(getByTitle("Nuevo chat")).toBeInTheDocument()
  })

  test("tiene clase bg-surface en el contenedor", () => {
    const { container } = render(<SessionList sessions={sessions} />)
    expect(container.firstChild?.className ?? "").toContain("bg-surface")
  })
})
