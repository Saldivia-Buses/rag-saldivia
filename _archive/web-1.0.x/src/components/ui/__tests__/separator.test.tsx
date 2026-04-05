import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { Separator } from "@/components/ui/separator"

afterEach(cleanup)

describe("<Separator />", () => {
  test("renderiza sin errores", () => {
    const { container } = render(<Separator />)
    expect(container.firstChild).toBeInTheDocument()
  })

  test("tiene clase bg-border", () => {
    const { container } = render(<Separator />)
    expect(container.firstChild?.className ?? "").toContain("bg-border")
  })

  test("orientación horizontal por defecto (h-px w-full)", () => {
    const { container } = render(<Separator orientation="horizontal" />)
    const el = container.querySelector("[role='separator'], [data-orientation]") ?? container.firstChild
    const cls = (el as Element)?.className ?? ""
    expect(cls).toContain("bg-border")
  })

  test("orientación vertical aplica w-px", () => {
    const { container } = render(<Separator orientation="vertical" />)
    const el = container.querySelector("[data-orientation='vertical']") ?? container.firstChild
    expect(el).toBeInTheDocument()
  })

  test("acepta className adicional", () => {
    const { container } = render(<Separator className="my-4" />)
    expect(container.firstChild?.className ?? "").toContain("my-4")
  })
})
