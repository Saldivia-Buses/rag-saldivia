import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { AdminAreas } from "@/components/admin/AdminAreas"

afterEach(cleanup)

mock.module("@/app/actions/areas", () => ({
  actionCreateArea: mock(() => Promise.resolve()),
  actionUpdateArea: mock(() => Promise.resolve()),
  actionDeleteArea: mock(() => Promise.resolve()),
  actionAddUserToArea: mock(() => Promise.resolve()),
  actionRemoveUserFromArea: mock(() => Promise.resolve()),
}))

const allUsers = [
  { id: 1, name: "Admin", email: "admin@test.com" },
  { id: 2, name: "User", email: "user@test.com" },
]

function makeArea(overrides: Partial<{
  id: number
  name: string
  description: string | null
  areaCollections: Array<{ areaId: number; collectionName: string; permission: string }>
  userAreas: Array<{ userId: number; areaId: number; user: { id: number; name: string; email: string } }>
}> = {}) {
  return {
    id: overrides.id ?? 1,
    name: overrides.name ?? "General",
    description: overrides.description ?? "Area general",
    createdAt: Date.now(),
    areaCollections: overrides.areaCollections ?? [],
    userAreas: overrides.userAreas ?? [],
  }
}

describe("<AdminAreas />", () => {
  test("renders area names", () => {
    const areas = [
      makeArea({ id: 1, name: "General" }),
      makeArea({ id: 2, name: "Soporte" }),
    ]
    const { getByText } = render(<AdminAreas areas={areas} allUsers={allUsers} />)
    expect(getByText("General")).toBeInTheDocument()
    expect(getByText("Soporte")).toBeInTheDocument()
  })

  test("shows empty placeholder when no areas", () => {
    const { getByText } = render(<AdminAreas areas={[]} allUsers={allUsers} />)
    expect(getByText("Sin áreas")).toBeInTheDocument()
    expect(getByText(/Creá un área para agrupar/)).toBeInTheDocument()
  })

  test("shows create button", () => {
    const { getByRole } = render(<AdminAreas areas={[]} allUsers={allUsers} />)
    expect(getByRole("button", { name: /Nueva área/ })).toBeInTheDocument()
  })

  test("shows collection badges for areas with collections", () => {
    const areas = [
      makeArea({
        id: 1,
        name: "General",
        areaCollections: [
          { areaId: 1, collectionName: "docs", permission: "read" },
          { areaId: 1, collectionName: "contratos", permission: "write" },
        ],
      }),
    ]
    const { getByText } = render(<AdminAreas areas={areas} allUsers={allUsers} />)
    expect(getByText("docs")).toBeInTheDocument()
    expect(getByText("contratos")).toBeInTheDocument()
  })
})
