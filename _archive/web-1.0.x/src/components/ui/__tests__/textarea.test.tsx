import { test, expect, describe, mock, afterEach } from "bun:test"
import { render, fireEvent, cleanup } from "@testing-library/react"
import { Textarea } from "@/components/ui/textarea"

afterEach(cleanup)

describe("<Textarea />", () => {
  test("renderiza con placeholder", () => {
    const { getByPlaceholderText } = render(<Textarea placeholder="Escribí aquí..." />)
    expect(getByPlaceholderText("Escribí aquí...")).toBeInTheDocument()
  })

  test("onChange se llama al cambiar el valor", () => {
    const onChange = mock(() => {})
    const { getByPlaceholderText } = render(<Textarea onChange={onChange} placeholder="test" />)
    fireEvent.change(getByPlaceholderText("test"), { target: { value: "texto" } })
    expect(onChange).toHaveBeenCalled()
  })

  test("disabled desactiva el textarea", () => {
    const { getByPlaceholderText } = render(<Textarea disabled placeholder="bloqueado" />)
    expect(getByPlaceholderText("bloqueado")).toBeDisabled()
  })

  test("tiene clase border-border", () => {
    const { getByPlaceholderText } = render(<Textarea placeholder="x" />)
    expect(getByPlaceholderText("x").className).toContain("border-border")
  })

  test("tiene clase resize-y", () => {
    const { getByPlaceholderText } = render(<Textarea placeholder="y" />)
    expect(getByPlaceholderText("y").className).toContain("resize-y")
  })

  test("acepta value controlado", () => {
    const { getByDisplayValue } = render(<Textarea value="contenido" readOnly />)
    expect(getByDisplayValue("contenido")).toBeInTheDocument()
  })
})
