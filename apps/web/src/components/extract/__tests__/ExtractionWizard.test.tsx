import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { ExtractionWizard } from "@/components/extract/ExtractionWizard"

afterEach(cleanup)

mock.module("@/lib/export", () => ({
  downloadFile: mock(() => {}),
  exportToMarkdown: mock(() => ""),
}))

describe("<ExtractionWizard />", () => {
  test("renderiza step 1 por defecto", () => {
    const { getByText } = render(<ExtractionWizard />)
    expect(getByText(/Seleccioná la colección/)).toBeInTheDocument()
  })

  test("muestra los 3 pasos en el indicador", () => {
    const { getByText } = render(<ExtractionWizard />)
    expect(getByText("Colección")).toBeInTheDocument()
    expect(getByText("Campos")).toBeInTheDocument()
    expect(getByText("Resultados")).toBeInTheDocument()
  })

  test("botón Siguiente deshabilitado sin colección", () => {
    const { getByRole } = render(<ExtractionWizard />)
    expect(getByRole("button", { name: /Siguiente/ })).toBeDisabled()
  })

  test("botón Siguiente se habilita con colección", () => {
    const { getByRole, getByPlaceholderText } = render(<ExtractionWizard />)
    fireEvent.change(getByPlaceholderText("nombre-de-coleccion"), { target: { value: "contratos" } })
    expect(getByRole("button", { name: /Siguiente/ })).not.toBeDisabled()
  })

  test("avanza al step 2 al hacer click en Siguiente", () => {
    const { getByRole, getByText, getByPlaceholderText } = render(<ExtractionWizard />)
    fireEvent.change(getByPlaceholderText("nombre-de-coleccion"), { target: { value: "contratos" } })
    fireEvent.click(getByRole("button", { name: /Siguiente/ }))
    expect(getByText(/Definí los campos/)).toBeInTheDocument()
  })

  test("en step 2 muestra campos para completar", () => {
    const { getByRole, getByText, getByPlaceholderText } = render(<ExtractionWizard />)
    fireEvent.change(getByPlaceholderText("nombre-de-coleccion"), { target: { value: "contratos" } })
    fireEvent.click(getByRole("button", { name: /Siguiente/ }))
    expect(getByPlaceholderText(/Nombre del campo/)).toBeInTheDocument()
  })
})
