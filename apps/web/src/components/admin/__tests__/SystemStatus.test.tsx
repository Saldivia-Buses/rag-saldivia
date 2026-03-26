import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { SystemStatus } from "@/components/admin/SystemStatus"

afterEach(cleanup)

const stats = { activeUsers: 9, areas: 3, collections: 5, activeJobs: 2, recentErrors: 0 }

describe("<SystemStatus />", () => {
  test("renderiza las stat cards con valores", () => {
    const { getByText } = render(<SystemStatus stats={stats} activeJobs={[]} />)
    expect(getByText("9")).toBeInTheDocument() // usuarios
    expect(getByText("3")).toBeInTheDocument() // áreas
    expect(getByText("5")).toBeInTheDocument() // colecciones
  })

  test("muestra botón Actualizar", () => {
    const { getByRole } = render(<SystemStatus stats={stats} activeJobs={[]} />)
    expect(getByRole("button", { name: /Actualizar/ })).toBeInTheDocument()
  })

  test("sin jobs activos muestra mensaje", () => {
    const { getByText } = render(<SystemStatus stats={stats} activeJobs={[]} />)
    expect(getByText("Sin jobs activos")).toBeInTheDocument()
  })

  test("con jobs muestra la tabla", () => {
    const jobs = [{
      id: "abc-123", filename: "doc.pdf", collection: "contratos",
      status: "locked" as const, progress: 50, tier: "standard",
      retryCount: 0, createdAt: Date.now(), error: null,
      lockedAt: null, nextRun: null, payload: null,
    }]
    const { getByText } = render(<SystemStatus stats={stats} activeJobs={jobs} />)
    expect(getByText("contratos")).toBeInTheDocument()
  })
})
