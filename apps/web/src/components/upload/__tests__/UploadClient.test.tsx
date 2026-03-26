import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { UploadClient } from "@/components/upload/UploadClient"

afterEach(cleanup)

describe("<UploadClient />", () => {
  test("renderiza el selector de colección", () => {
    const { getByText } = render(<UploadClient collections={["contratos", "politicas"]} />)
    expect(getByText("Colección destino")).toBeInTheDocument()
  })

  test("sin colecciones muestra mensaje de error", () => {
    const { getByText } = render(<UploadClient collections={[]} />)
    expect(getByText(/no tenés colecciones/i)).toBeInTheDocument()
  })

  test("drop zone visible", () => {
    const { getByText } = render(<UploadClient collections={["contratos"]} />)
    expect(getByText(/Arrastrá archivos/)).toBeInTheDocument()
  })

  test("drop zone muestra tipos aceptados", () => {
    const { getByText } = render(<UploadClient collections={["contratos"]} />)
    expect(getByText(/PDF, DOCX, TXT/)).toBeInTheDocument()
  })

  test("tiene header con título", () => {
    const { getByText } = render(<UploadClient collections={["contratos"]} />)
    expect(getByText("Subir documentos")).toBeInTheDocument()
  })
})
