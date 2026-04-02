/**
 * Tests for small admin components.
 * AdminLayout, RoleBadge, UserRoleSelector, PermissionMatrix
 */
import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { AdminLayout } from "../AdminLayout"
import { RoleBadge } from "../RoleBadge"
import { PermissionMatrix } from "../PermissionMatrix"

afterEach(cleanup)

// ── AdminLayout ──

describe("AdminLayout", () => {
  test("renders heading", () => {
    const { getByText } = render(<AdminLayout><div>Content</div></AdminLayout>)
    expect(getByText("Administración")).toBeTruthy()
  })

  test("renders tab labels", () => {
    const { getByText } = render(<AdminLayout><div /></AdminLayout>)
    expect(getByText("Dashboard")).toBeTruthy()
    expect(getByText("Usuarios")).toBeTruthy()
    expect(getByText("Roles")).toBeTruthy()
    expect(getByText("Áreas")).toBeTruthy()
    expect(getByText("Permisos")).toBeTruthy()
    expect(getByText("Colecciones")).toBeTruthy()
    expect(getByText("Config RAG")).toBeTruthy()
  })

  test("renders children", () => {
    const { getByText } = render(<AdminLayout><p>Hello admin</p></AdminLayout>)
    expect(getByText("Hello admin")).toBeTruthy()
  })
})

// ── RoleBadge ──

describe("RoleBadge", () => {
  test("renders role name", () => {
    const { getByText } = render(<RoleBadge name="Admin" color="#2563eb" />)
    expect(getByText("Admin")).toBeTruthy()
  })

  test("renders with xs size", () => {
    const { getByText } = render(<RoleBadge name="User" color="#6e6c69" size="xs" />)
    expect(getByText("User")).toBeTruthy()
  })

  test("applies color styling", () => {
    const { getByText } = render(<RoleBadge name="Test" color="#dc2626" />)
    const el = getByText("Test")
    expect(el.style.color).toBe("#dc2626")
  })
})

// ── PermissionMatrix ──

describe("PermissionMatrix", () => {
  const PERMISSIONS = [
    { id: 1, key: "users.manage", label: "Gestionar usuarios", category: "Usuarios", description: "CRUD users" },
    { id: 2, key: "users.view", label: "Ver usuarios", category: "Usuarios", description: "List users" },
    { id: 3, key: "roles.manage", label: "Gestionar roles", category: "Roles", description: "CRUD roles" },
  ]

  test("renders category headings", () => {
    const { getByText } = render(
      <PermissionMatrix roleId={1} roleName="Admin" isSystem={false} allPermissions={PERMISSIONS} activeKeys={[]} onUpdate={() => {}} />
    )
    expect(getByText("Usuarios")).toBeTruthy()
    expect(getByText("Roles")).toBeTruthy()
  })

  test("renders permission labels", () => {
    const { getByText } = render(
      <PermissionMatrix roleId={1} roleName="Admin" isSystem={false} allPermissions={PERMISSIONS} activeKeys={[]} onUpdate={() => {}} />
    )
    expect(getByText("Gestionar usuarios")).toBeTruthy()
    expect(getByText("Ver usuarios")).toBeTruthy()
    expect(getByText("Gestionar roles")).toBeTruthy()
  })

  test("renders role name in header", () => {
    const { getByText } = render(
      <PermissionMatrix roleId={1} roleName="Soporte" isSystem={false} allPermissions={PERMISSIONS} activeKeys={[]} onUpdate={() => {}} />
    )
    expect(getByText(/Soporte/)).toBeTruthy()
  })
})
