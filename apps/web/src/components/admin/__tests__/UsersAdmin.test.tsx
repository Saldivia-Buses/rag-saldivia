import { test, expect, describe, afterEach, mock, beforeAll } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { UsersAdmin } from "@/components/admin/UsersAdmin"

afterEach(cleanup)

const mockAreas = [
  { id: 1, name: "General", description: "Área general", createdAt: Date.now() },
]

const mockUsers = [
  { id: 1, email: "admin@test.com", name: "Admin", role: "admin" as const, active: true, passwordHash: "", apiKeyHash: "", preferences: {}, createdAt: Date.now(), onboardingCompleted: false, userAreas: [{ area: mockAreas[0]! }] },
  { id: 2, email: "user@test.com",  name: "Usuario", role: "user" as const,  active: true, passwordHash: "", apiKeyHash: "", preferences: {}, createdAt: Date.now(), onboardingCompleted: false, userAreas: [] },
  { id: 3, email: "inactive@test.com", name: "Inactivo", role: "user" as const, active: false, passwordHash: "", apiKeyHash: "", preferences: {}, createdAt: Date.now(), onboardingCompleted: false, userAreas: [] },
]

// Mock de server actions
mock.module("@/app/actions/users", () => ({
  actionCreateUser: mock(() => Promise.resolve()),
  actionDeleteUser: mock(() => Promise.resolve()),
  actionUpdateUser: mock(() => Promise.resolve()),
}))

describe("<UsersAdmin />", () => {
  test("renderiza la tabla con usuarios", () => {
    const { getByText } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    expect(getByText("admin@test.com")).toBeInTheDocument()
    expect(getByText("user@test.com")).toBeInTheDocument()
  })

  test("muestra el header con conteo de usuarios", () => {
    const { getByText } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    expect(getByText(/3 usuarios/)).toBeInTheDocument()
  })

  test("badge Admin tiene variant default (navy)", () => {
    const { container } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    // Buscar el primer badge con bg-primary en la tabla
    const primaryBadge = container.querySelector(".bg-primary")
    expect(primaryBadge).toBeInTheDocument()
  })

  test("badge Activo tiene variant success", () => {
    const { getAllByText } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    const activoBadges = getAllByText("Activo")
    expect(activoBadges[0]?.className).toContain("success")
  })

  test("badge Inactivo tiene variant destructive", () => {
    const { container } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    // Buscar badges con clase destructive
    const destructiveBadge = container.querySelector("[class*='destructive']")
    expect(destructiveBadge).toBeInTheDocument()
  })

  test("botón Nuevo usuario presente", () => {
    const { getByRole } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    expect(getByRole("button", { name: /Nuevo usuario/ })).toBeInTheDocument()
  })

  test("sin usuarios muestra EmptyPlaceholder", () => {
    const { getByText } = render(<UsersAdmin users={[]} areas={mockAreas} />)
    expect(getByText("Sin usuarios")).toBeInTheDocument()
  })

  test("muestra el área asignada del usuario", () => {
    const { getByText } = render(<UsersAdmin users={mockUsers} areas={mockAreas} />)
    expect(getByText("General")).toBeInTheDocument()
  })
})
