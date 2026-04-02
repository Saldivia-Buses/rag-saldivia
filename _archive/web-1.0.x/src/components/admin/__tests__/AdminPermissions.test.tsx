import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { AdminPermissions } from "../AdminPermissions"

afterEach(cleanup)

const AREAS = [
  { id: 1, name: "Operaciones", description: null, areaCollections: [{ collectionName: "docs", permission: "read" }] },
  { id: 2, name: "Finanzas", description: null, areaCollections: [] },
]
const COLLECTIONS = ["docs", "legal"]

describe("AdminPermissions", () => {
  test("renders without crash", () => {
    const { container } = render(<AdminPermissions areas={AREAS} collections={COLLECTIONS} />)
    expect(container.firstChild).toBeTruthy()
  })

  test("renders area sidebar with heading", () => {
    const { getByText, getAllByText } = render(<AdminPermissions areas={AREAS} collections={COLLECTIONS} />)
    expect(getByText("Áreas")).toBeTruthy()
    // Area names appear in buttons with count suffix
    expect(getAllByText(/Operaciones/).length).toBeGreaterThanOrEqual(1)
    expect(getAllByText(/Finanzas/).length).toBeGreaterThanOrEqual(1)
  })

  test("renders empty state when no areas", () => {
    const { getByText } = render(<AdminPermissions areas={[]} collections={COLLECTIONS} />)
    expect(getByText(/Sin áreas/)).toBeTruthy()
  })
})
