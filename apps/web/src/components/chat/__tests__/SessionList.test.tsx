import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { SessionList } from "@/components/chat/SessionList"

afterEach(cleanup)

mock.module("@/app/actions/chat", () => ({
  actionDeleteSession: mock(() => Promise.resolve()),
  actionCreateSession: mock(() => Promise.resolve("new-id")),
  actionRenameSession: mock(() => Promise.resolve()),
  actionAddTag: mock(() => Promise.resolve()),
  actionRemoveTag: mock(() => Promise.resolve()),
  actionForkSession: mock(() => Promise.resolve("fork-id")),
  actionToggleSaved: mock(() => Promise.resolve()),
}))

mock.module("@/lib/export", () => ({
  exportToMarkdown: mock(() => "# Sesión"),
  downloadFile: mock(() => {}),
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

  test("muestra encabezado Sesiones", () => {
    const { getAllByText } = render(<SessionList sessions={sessions} />)
    expect(getAllByText("Sesiones").length).toBeGreaterThan(0)
  })

  test("sin sesiones muestra lista vacía (sin mensaje de error)", () => {
    const { container } = render(<SessionList sessions={[]} />)
    expect(container.querySelector(".flex-1")).toBeInTheDocument()
  })

  test("botón Nueva sesión presente", () => {
    const { getByTitle } = render(<SessionList sessions={sessions} />)
    expect(getByTitle("Nueva sesión")).toBeInTheDocument()
  })

  test("tiene clase bg-surface en el contenedor", () => {
    const { container } = render(<SessionList sessions={sessions} />)
    expect(container.firstChild?.className ?? "").toContain("bg-surface")
  })
})
