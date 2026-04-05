import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { AdminUsers } from "../AdminUsers"

afterEach(cleanup)

const USERS = [
  { id: 1, email: "admin@test.com", name: "Admin", role: "admin", active: true, createdAt: Date.now(), lastLogin: Date.now() },
  { id: 2, email: "user@test.com", name: "María", role: "user", active: true, createdAt: Date.now(), lastLogin: null },
  { id: 3, email: "inactive@test.com", name: "Juan", role: "user", active: false, createdAt: Date.now(), lastLogin: null },
]

describe("AdminUsers", () => {
  test("renders user count", () => {
    const { getByText } = render(<AdminUsers initialUsers={USERS} currentUserId={1} />)
    expect(getByText("3 usuarios registrados")).toBeTruthy()
  })

  test("renders user names in table", () => {
    const { getByText } = render(<AdminUsers initialUsers={USERS} currentUserId={1} />)
    expect(getByText("Admin")).toBeTruthy()
    expect(getByText("María")).toBeTruthy()
    expect(getByText("Juan")).toBeTruthy()
  })

  test("renders user emails", () => {
    const { getByText } = render(<AdminUsers initialUsers={USERS} currentUserId={1} />)
    expect(getByText("admin@test.com")).toBeTruthy()
    expect(getByText("user@test.com")).toBeTruthy()
  })

  test("renders nuevo usuario button", () => {
    const { getByText } = render(<AdminUsers initialUsers={USERS} currentUserId={1} />)
    expect(getByText("Nuevo usuario")).toBeTruthy()
  })

  test("renders active/inactive status", () => {
    const { getAllByText, getByText } = render(<AdminUsers initialUsers={USERS} currentUserId={1} />)
    expect(getAllByText("Activo").length).toBeGreaterThanOrEqual(2)
    expect(getByText("Inactivo")).toBeTruthy()
  })

  test("renders Nunca for users without last login", () => {
    const { getAllByText } = render(<AdminUsers initialUsers={USERS} currentUserId={1} />)
    expect(getAllByText("Nunca").length).toBe(2)
  })
})
