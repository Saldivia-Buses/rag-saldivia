import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { CollectionsList } from "@/components/collections/CollectionsList"

afterEach(cleanup)

mock.module("@/app/actions/chat", () => ({
  actionCreateSessionForDoc: mock(() => Promise.resolve("session-id")),
}))

const mockUser = { id: 1, email: "admin@test.com", name: "Admin", role: "admin" as const, active: true, preferences: {} }

describe("<CollectionsList />", () => {
  test("renderiza las colecciones", () => {
    const { getByText } = render(<CollectionsList collections={["contratos", "politicas"]} user={mockUser} />)
    expect(getByText("contratos")).toBeInTheDocument()
    expect(getByText("politicas")).toBeInTheDocument()
  })

  test("sin colecciones muestra EmptyPlaceholder para admin", () => {
    const { getByText } = render(<CollectionsList collections={[]} user={mockUser} />)
    expect(getByText("Sin colecciones disponibles")).toBeInTheDocument()
    expect(getByText(/Creá una colección/)).toBeInTheDocument()
  })

  test("sin colecciones muestra mensaje diferente para user no admin", () => {
    const user = { ...mockUser, role: "user" as const }
    const { getByText } = render(<CollectionsList collections={[]} user={user} />)
    expect(getByText(/No tenés acceso/)).toBeInTheDocument()
  })

  test("admin ve botón Nueva colección", () => {
    const { getByRole } = render(<CollectionsList collections={[]} user={mockUser} />)
    expect(getByRole("button", { name: /Nueva colección/ })).toBeInTheDocument()
  })

  test("las colecciones tienen botón Chat", () => {
    const { getAllByRole } = render(<CollectionsList collections={["contratos"]} user={mockUser} />)
    const chatBtns = getAllByRole("button", { name: /Chat/ })
    expect(chatBtns.length).toBeGreaterThan(0)
  })
})
