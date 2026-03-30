import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { AdminCollections } from "@/components/admin/AdminCollections"

afterEach(cleanup)

mock.module("@/app/actions/collections", () => ({
  actionCreateCollection: mock(() => Promise.resolve()),
  actionDeleteCollection: mock(() => Promise.resolve()),
}))

function makeCol(name: string, areas: Array<{ areaName: string; permission: string }> = []) {
  return { name, areas }
}

describe("<AdminCollections />", () => {
  test("renders collection names", () => {
    const collections = [makeCol("docs"), makeCol("contratos")]
    const { getByText } = render(<AdminCollections collections={collections} />)
    expect(getByText("docs")).toBeInTheDocument()
    expect(getByText("contratos")).toBeInTheDocument()
  })

  test("shows empty placeholder when no collections", () => {
    const { getByText } = render(<AdminCollections collections={[]} />)
    expect(getByText("Sin colecciones")).toBeInTheDocument()
    expect(getByText(/Creá una colección para empezar/)).toBeInTheDocument()
  })

  test("shows area badges", () => {
    const collections = [
      makeCol("docs", [
        { areaName: "General", permission: "read" },
        { areaName: "Soporte", permission: "write" },
      ]),
    ]
    const { getByText } = render(<AdminCollections collections={collections} />)
    expect(getByText("General")).toBeInTheDocument()
    expect(getByText("Soporte")).toBeInTheDocument()
  })
})
