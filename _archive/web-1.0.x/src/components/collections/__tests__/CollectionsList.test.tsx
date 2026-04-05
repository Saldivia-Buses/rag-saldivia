import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { CollectionsList } from "@/components/collections/CollectionsList"

afterEach(cleanup)

const mockUser = { id: 1, email: "admin@test.com", name: "Admin", role: "admin" as const, active: true, preferences: {} }

function cols(...names: string[]) {
  return names.map((name) => ({ name, permission: null }))
}

describe("<CollectionsList />", () => {
  test("renderiza las colecciones", () => {
    const { getByText } = render(<CollectionsList collections={cols("contratos", "politicas")} user={mockUser} />)
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
    const { getAllByRole } = render(<CollectionsList collections={cols("contratos")} user={mockUser} />)
    const chatBtns = getAllByRole("button", { name: /Chat/ })
    expect(chatBtns.length).toBeGreaterThan(0)
  })

  test("muestra badge de permiso cuando tiene permiso", () => {
    const user = { ...mockUser, role: "user" as const }
    const { getByText } = render(
      <CollectionsList collections={[{ name: "docs", permission: "read" }]} user={user} />
    )
    expect(getByText("read")).toBeInTheDocument()
  })
})
