import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { Badge } from "@/components/ui/badge"

afterEach(cleanup)

describe("<Badge />", () => {
  test("renderiza texto", () => {
    const { getByText } = render(<Badge>Admin</Badge>)
    expect(getByText("Admin")).toBeInTheDocument()
  })

  test("variant default tiene bg-primary", () => {
    const { getByText } = render(<Badge variant="default">Test</Badge>)
    expect(getByText("Test").className).toContain("bg-primary")
  })

  test("variant destructive aplica estilos de error", () => {
    const { getByText } = render(<Badge variant="destructive">Error</Badge>)
    expect(getByText("Error").className).toContain("destructive")
  })

  test("variant success aplica estilos de éxito", () => {
    const { getByText } = render(<Badge variant="success">Activo</Badge>)
    expect(getByText("Activo").className).toContain("success")
  })

  test("variant warning aplica estilos de advertencia", () => {
    const { getByText } = render(<Badge variant="warning">Pendiente</Badge>)
    expect(getByText("Pendiente").className).toContain("warning")
  })

  test("variant outline tiene borde sin bg sólido", () => {
    const { getByText } = render(<Badge variant="outline">área</Badge>)
    const el = getByText("área")
    expect(el.className).toContain("border")
    expect(el.className).not.toContain("bg-primary")
  })

  test("variant secondary tiene bg-secondary", () => {
    const { getByText } = render(<Badge variant="secondary">Tag</Badge>)
    expect(getByText("Tag").className).toContain("bg-secondary")
  })
})
