import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { ThemeToggle } from "@/components/ui/theme-toggle"

afterEach(cleanup)

describe("<ThemeToggle />", () => {
  test("renderiza un botón sin errores", () => {
    const { container } = render(<ThemeToggle />)
    expect(container.querySelector("button")).toBeInTheDocument()
  })

  test("el botón tiene aria-label o title descriptivo", () => {
    const { container } = render(<ThemeToggle />)
    const btn = container.querySelector("button")
    // Antes del mount (SSR): aria-label="Tema"
    // Después del mount: title con "Cambiar a X mode"
    const hasLabel = btn?.getAttribute("aria-label") || btn?.getAttribute("title")
    expect(hasLabel).toBeTruthy()
  })

  test("tiene tamaño 44px", () => {
    const { container } = render(<ThemeToggle />)
    const btn = container.querySelector("button")
    expect(btn?.style.width).toBe("44px")
  })
})
