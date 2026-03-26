import { test, expect, describe, mock, afterEach } from "bun:test"
import { render, fireEvent, cleanup } from "@testing-library/react"
import { Input } from "@/components/ui/input"

afterEach(cleanup)

describe("<Input />", () => {
  test("renderiza el placeholder", () => {
    const { getByPlaceholderText } = render(<Input placeholder="email" />)
    expect(getByPlaceholderText("email")).toBeInTheDocument()
  })

  test("onChange se llama al cambiar el valor", () => {
    const onChange = mock(() => {})
    const { getByPlaceholderText } = render(<Input onChange={onChange} placeholder="test" />)
    fireEvent.change(getByPlaceholderText("test"), { target: { value: "hola" } })
    expect(onChange).toHaveBeenCalled()
  })

  test("disabled desactiva el input", () => {
    const { getByPlaceholderText } = render(<Input disabled placeholder="bloqueado" />)
    expect(getByPlaceholderText("bloqueado")).toBeDisabled()
  })

  test("type password configura el tipo correctamente", () => {
    const { getByPlaceholderText } = render(<Input type="password" placeholder="pass" />)
    expect((getByPlaceholderText("pass") as HTMLInputElement).type).toBe("password")
  })

  test("acepta value controlado", () => {
    const { getByDisplayValue } = render(<Input value="valor" readOnly placeholder="x" />)
    expect(getByDisplayValue("valor")).toBeInTheDocument()
  })

  test("tiene clase border-border", () => {
    const { getByPlaceholderText } = render(<Input placeholder="x" />)
    expect(getByPlaceholderText("x").className).toContain("border-border")
  })

  test("se asocia con label via htmlFor", () => {
    const { getByLabelText } = render(
      <><label htmlFor="inp">Email</label><Input id="inp" /></>
    )
    expect(getByLabelText("Email")).toBeInTheDocument()
  })
})
