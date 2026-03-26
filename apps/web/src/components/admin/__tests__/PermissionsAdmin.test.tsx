import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { PermissionsAdmin } from "@/components/admin/PermissionsAdmin"

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

const mockCollections = ["contratos", "politicas", "tecpia"]

describe("<PermissionsAdmin />", () => {
  test("renderiza la lista de áreas en el sidebar", () => {
    const { getAllByText } = render(<PermissionsAdmin areas={mockAreas} collections={mockCollections} />)
    expect(getAllByText("Legales").length).toBeGreaterThan(0)
    expect(getAllByText("RRHH").length).toBeGreaterThan(0)
  })

  test("renderiza las colecciones en la matriz", () => {
    const { getByText } = render(<PermissionsAdmin areas={mockAreas} collections={mockCollections} />)
    expect(getByText("contratos")).toBeInTheDocument()
    expect(getByText("politicas")).toBeInTheDocument()
  })

  test("botón Guardar cambios presente", () => {
    const { getByRole } = render(<PermissionsAdmin areas={mockAreas} collections={mockCollections} />)
    expect(getByRole("button", { name: /Guardar cambios/ })).toBeInTheDocument()
  })

  test("sin colecciones muestra mensaje", () => {
    const { getByText } = render(<PermissionsAdmin areas={mockAreas} collections={[]} />)
    expect(getByText("No hay colecciones disponibles")).toBeInTheDocument()
  })

  test("el área seleccionada tiene clase accent en el sidebar", () => {
    const { container } = render(<PermissionsAdmin areas={mockAreas} collections={mockCollections} />)
    // El primer botón de área en el sidebar debe tener clase accent (es el seleccionado)
    const sidebarBtn = container.querySelector("button[class*='accent']")
    expect(sidebarBtn).toBeInTheDocument()
  })
})
