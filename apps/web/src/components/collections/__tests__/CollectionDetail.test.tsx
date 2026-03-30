import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { CollectionDetail } from "@/components/collections/CollectionDetail"

afterEach(cleanup)

mock.module("@/app/actions/chat", () => ({
  actionCreateSession: mock(() => Promise.resolve({ data: { id: "new-session" } })),
}))

const baseProps = {
  name: "docs",
  userPermission: "read" as string | null,
  isAdmin: false,
  areas: [] as Array<{ name: string; permission: string }>,
  history: [] as Array<{
    id: string
    collection: string
    action: "added" | "removed"
    filename: string | null
    docCount: number | null
    userId: number
    createdAt: number
  }>,
}

describe("<CollectionDetail />", () => {
  test("renders collection name as heading", () => {
    const { getByRole } = render(<CollectionDetail {...baseProps} />)
    expect(getByRole("heading", { level: 1 })).toHaveTextContent("docs")
  })

  test("shows chat button", () => {
    const { getByRole } = render(<CollectionDetail {...baseProps} />)
    expect(getByRole("button", { name: /Chatear con esta colección/ })).toBeInTheDocument()
  })

  test("shows history table when events provided", () => {
    const history = [
      {
        id: "evt-1",
        collection: "docs",
        action: "added" as const,
        filename: "manual.pdf",
        docCount: 42,
        userId: 1,
        createdAt: Date.now(),
      },
    ]
    const { getByText } = render(<CollectionDetail {...baseProps} history={history} />)
    expect(getByText("manual.pdf")).toBeInTheDocument()
    expect(getByText("42")).toBeInTheDocument()
    expect(getByText("ingesta")).toBeInTheDocument()
  })

  test("shows empty message when no history", () => {
    const { getByText } = render(<CollectionDetail {...baseProps} history={[]} />)
    expect(getByText("Sin eventos de ingesta registrados.")).toBeInTheDocument()
  })
})
