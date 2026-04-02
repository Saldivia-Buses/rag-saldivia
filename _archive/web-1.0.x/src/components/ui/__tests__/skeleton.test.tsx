import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { Skeleton, SkeletonText, SkeletonAvatar, SkeletonCard, SkeletonTable } from "@/components/ui/skeleton"

afterEach(cleanup)

describe("<Skeleton />", () => {
  test("renderiza sin errores", () => {
    const { container } = render(<Skeleton />)
    expect(container.firstChild).toBeInTheDocument()
  })

  test("tiene clase animate-pulse", () => {
    const { container } = render(<Skeleton />)
    expect(container.firstChild?.className ?? "").toContain("animate-pulse")
  })

  test("acepta className adicional", () => {
    const { container } = render(<Skeleton className="h-4 w-32" />)
    expect(container.firstChild?.className ?? "").toContain("h-4")
  })
})

describe("<SkeletonText />", () => {
  test("renderiza N líneas animadas", () => {
    const { container } = render(<SkeletonText lines={4} />)
    expect(container.querySelectorAll(".animate-pulse").length).toBe(4)
  })

  test("valor por defecto es 3 líneas", () => {
    const { container } = render(<SkeletonText />)
    expect(container.querySelectorAll(".animate-pulse").length).toBe(3)
  })
})

describe("<SkeletonAvatar />", () => {
  test("renderiza redondeado", () => {
    const { container } = render(<SkeletonAvatar />)
    expect(container.firstChild?.className ?? "").toContain("rounded-full")
  })

  test("size sm aplica h-6 w-6", () => {
    const { container } = render(<SkeletonAvatar size="sm" />)
    expect(container.firstChild?.className ?? "").toContain("h-6")
  })
})

describe("<SkeletonCard />", () => {
  test("renderiza sin errores", () => {
    const { container } = render(<SkeletonCard />)
    expect(container.firstChild).toBeInTheDocument()
  })
})

describe("<SkeletonTable />", () => {
  test("renderiza filas de datos", () => {
    const { container } = render(<SkeletonTable rows={3} cols={2} />)
    const rows = container.querySelectorAll(".flex.gap-4")
    expect(rows.length).toBeGreaterThanOrEqual(4) // header + 3 filas
  })
})
