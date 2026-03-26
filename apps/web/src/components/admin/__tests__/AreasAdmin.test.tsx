import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { AreasAdmin } from "@/components/admin/AreasAdmin"

afterEach(cleanup)

mock.module("@/app/actions/areas", () => ({
  actionCreateArea: mock(() => Promise.resolve()),
  actionUpdateArea: mock(() => Promise.resolve()),
  actionDeleteArea: mock(() => Promise.resolve()),
  actionSetAreaCollections: mock(() => Promise.resolve()),
}))

const mockAreas = [
  { id: 1, name: "Legales", description: "Área legal", createdAt: Date.now(), areaCollections: [{ collectionName: "contratos", permission: "read" }] },
  { id: 2, name: "RRHH", description: "", createdAt: Date.now(), areaCollections: [] },
]

describe("<AreasAdmin />", () => {
  test("renderiza lista de áreas", () => {
    const { getByText } = render(<AreasAdmin areas={mockAreas} />)
    expect(getByText("Legales")).toBeInTheDocument()
    expect(getByText("RRHH")).toBeInTheDocument()
  })

  test("muestra las colecciones asignadas", () => {
    const { getByText } = render(<AreasAdmin areas={mockAreas} />)
    expect(getByText("contratos")).toBeInTheDocument()
  })

  test("muestra — cuando no hay colecciones", () => {
    const { getAllByText } = render(<AreasAdmin areas={mockAreas} />)
    expect(getAllByText("—").length).toBeGreaterThan(0)
  })

  test("muestra header con conteo", () => {
    const { getByText } = render(<AreasAdmin areas={mockAreas} />)
    expect(getByText(/2 área/)).toBeInTheDocument()
  })

  test("botón Nueva área presente", () => {
    const { getByRole } = render(<AreasAdmin areas={mockAreas} />)
    expect(getByRole("button", { name: /Nueva área/ })).toBeInTheDocument()
  })

  test("sin áreas muestra EmptyPlaceholder", () => {
    const { getByText } = render(<AreasAdmin areas={[]} />)
    expect(getByText("Sin áreas")).toBeInTheDocument()
  })
})
