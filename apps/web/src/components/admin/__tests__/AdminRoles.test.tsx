import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { AdminRoles } from "../AdminRoles"

afterEach(cleanup)

const ROLES = [
  { id: 1, name: "Admin", description: "Administrador", level: 100, color: "#2563eb", icon: "shield", isSystem: true, userCount: 1 },
  { id: 2, name: "User", description: "Usuario estándar", level: 10, color: "#6e6c69", icon: "user", isSystem: true, userCount: 5 },
  { id: 3, name: "Soporte", description: "Equipo de soporte", level: 30, color: "#16a34a", icon: "headphones", isSystem: false, userCount: 2 },
]

const PERMISSIONS = [
  { id: 1, key: "users.manage", label: "Gestionar usuarios", category: "Usuarios", description: "" },
  { id: 2, key: "roles.manage", label: "Gestionar roles", category: "Roles", description: "" },
]

describe("AdminRoles", () => {
  test("renders heading", () => {
    const { getByText } = render(<AdminRoles initialRoles={ROLES} initialPermissions={PERMISSIONS} />)
    expect(getByText("Roles del sistema")).toBeTruthy()
  })

  test("renders role names", () => {
    const { getByText } = render(<AdminRoles initialRoles={ROLES} initialPermissions={PERMISSIONS} />)
    expect(getByText("Admin")).toBeTruthy()
    expect(getByText("User")).toBeTruthy()
    expect(getByText("Soporte")).toBeTruthy()
  })

  test("renders user counts", () => {
    const { getByText } = render(<AdminRoles initialRoles={ROLES} initialPermissions={PERMISSIONS} />)
    expect(getByText("1 usuario")).toBeTruthy()
    expect(getByText("5 usuarios")).toBeTruthy()
    expect(getByText("2 usuarios")).toBeTruthy()
  })

  test("renders nuevo rol button", () => {
    const { getByText } = render(<AdminRoles initialRoles={ROLES} initialPermissions={PERMISSIONS} />)
    expect(getByText("Nuevo rol")).toBeTruthy()
  })

  test("renders system badge on system roles", () => {
    const { getAllByText } = render(<AdminRoles initialRoles={ROLES} initialPermissions={PERMISSIONS} />)
    expect(getAllByText("sistema").length).toBe(2) // Admin + User
  })
})
