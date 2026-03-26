import { test, expect, describe, mock, afterEach } from "bun:test"
import { render, fireEvent, cleanup } from "@testing-library/react"
import { Button } from "@/components/ui/button"

afterEach(cleanup)

describe("<Button />", () => {
  test("renderiza con texto", () => {
    const { getByRole } = render(<Button>Enviar</Button>)
    expect(getByRole("button", { name: "Enviar" })).toBeInTheDocument()
  })

  test("tiene role button", () => {
    const { getByRole } = render(<Button>Btn</Button>)
    expect(getByRole("button")).toBeInTheDocument()
  })

  test("variant default tiene bg-primary", () => {
    const { getByRole } = render(<Button variant="default">Test</Button>)
    expect(getByRole("button").className).toContain("bg-primary")
  })

  test("variant destructive aplica bg-destructive", () => {
    const { getByRole } = render(<Button variant="destructive">Eliminar</Button>)
    expect(getByRole("button").className).toContain("bg-destructive")
  })

  test("variant outline tiene borde", () => {
    const { getByRole } = render(<Button variant="outline">Cancelar</Button>)
    expect(getByRole("button").className).toContain("border")
  })

  test("variant ghost usa hover:bg-secondary", () => {
    const { getByRole } = render(<Button variant="ghost">Ghost</Button>)
    expect(getByRole("button").className).toContain("hover:bg-secondary")
  })

  test("variant link tiene underline-offset", () => {
    const { getByRole } = render(<Button variant="link">Link</Button>)
    expect(getByRole("button").className).toContain("underline-offset")
  })

  test("disabled desactiva el botón", () => {
    const { getByRole } = render(<Button disabled>Bloqueado</Button>)
    expect(getByRole("button")).toBeDisabled()
  })

  test("llama onClick al hacer click", () => {
    const onClick = mock(() => {})
    const { getByRole } = render(<Button onClick={onClick}>Click</Button>)
    fireEvent.click(getByRole("button"))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  test("disabled no llama onClick", () => {
    const onClick = mock(() => {})
    const { getByRole } = render(<Button disabled onClick={onClick}>Bloqueado</Button>)
    fireEvent.click(getByRole("button"))
    expect(onClick).not.toHaveBeenCalled()
  })

  test("asChild renderiza el elemento hijo como raíz", () => {
    const { getByRole } = render(
      <Button asChild><a href="/link">Enlace</a></Button>
    )
    expect(getByRole("link", { name: "Enlace" })).toBeInTheDocument()
  })
})
