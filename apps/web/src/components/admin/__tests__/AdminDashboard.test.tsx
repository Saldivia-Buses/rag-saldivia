import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { AdminDashboard } from "../AdminDashboard"

// cleanup() unmounts the component, which triggers useEffect cleanup
// and clears the 30s setInterval in AdminDashboard
afterEach(cleanup)

const STATS = {
  users: { total: 10, active: 8, inactive: 2 },
  sessions: 45,
  messages: 230,
  roles: [
    { id: 1, name: "Admin", color: "#2563eb", level: 100, userCount: 2 },
    { id: 2, name: "User", color: "#6e6c69", level: 10, userCount: 8 },
  ],
  usersPresence: [
    { id: 1, name: "AdminUser", email: "admin@test.com", lastSeen: Date.now(), active: true },
    { id: 2, name: "María", email: "maria@test.com", lastSeen: Date.now() - 300_000, active: true },
  ],
}

describe("AdminDashboard", () => {
  test("renders without crash", () => {
    const { container } = render(<AdminDashboard stats={STATS} />)
    expect(container.firstChild).toBeTruthy()
  })

  test("renders stat values", () => {
    const { getByText } = render(<AdminDashboard stats={STATS} />)
    expect(getByText("10")).toBeTruthy()
    expect(getByText("45")).toBeTruthy()
    expect(getByText("230")).toBeTruthy()
  })

  test("renders user names in presence list", () => {
    const { getByText } = render(<AdminDashboard stats={STATS} />)
    expect(getByText("AdminUser")).toBeTruthy()
    expect(getByText("María")).toBeTruthy()
  })
})
