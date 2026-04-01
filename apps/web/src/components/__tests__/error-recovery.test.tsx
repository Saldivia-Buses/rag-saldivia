import { describe, test, expect, afterEach, mock } from "bun:test"
import { cleanup, render, fireEvent } from "@testing-library/react"
import { ErrorRecovery, ErrorRecoveryFromError } from "../ui/error-recovery"
import type { UserErrorRecovery } from "@/lib/error-recovery"

afterEach(cleanup)

const mockPush = mock(() => {})
mock.module("next/navigation", () => ({
  useRouter: () => ({ push: mockPush, replace: mock(), back: mock(), forward: mock(), refresh: mock(), prefetch: mock() }),
  usePathname: () => "/chat",
  useSearchParams: () => new URLSearchParams(),
}))

function makeRecovery(overrides: Partial<UserErrorRecovery> = {}): UserErrorRecovery {
  return {
    title: "Test Error",
    description: "Something went wrong.",
    suggestion: "Try again later.",
    actions: [{ label: "Reintentar", type: "retry" }],
    icon: "generic",
    ...overrides,
  }
}

describe("ErrorRecovery", () => {
  test("renders title and description", () => {
    const { getByText } = render(
      <ErrorRecovery recovery={makeRecovery({ title: "Servidor caído", description: "No responde." })} />
    )
    expect(getByText("Servidor caído")).toBeTruthy()
    expect(getByText("No responde.")).toBeTruthy()
  })

  test("renders suggestion text", () => {
    const { getByText } = render(
      <ErrorRecovery recovery={makeRecovery({ suggestion: "Reintentá en 5 min." })} />
    )
    expect(getByText("Reintentá en 5 min.")).toBeTruthy()
  })

  test("retry button calls onRetry", () => {
    const onRetry = mock(() => {})
    const { getByText } = render(
      <ErrorRecovery recovery={makeRecovery()} onRetry={onRetry} />
    )
    fireEvent.click(getByText("Reintentar"))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  test("navigate button calls router.push", () => {
    mockPush.mockClear()
    const recovery = makeRecovery({
      actions: [{ label: "Ir a colecciones", type: "navigate", href: "/collections" }],
    })
    const { getByText } = render(<ErrorRecovery recovery={recovery} />)
    fireEvent.click(getByText("Ir a colecciones"))
    expect(mockPush).toHaveBeenCalledWith("/collections")
  })

  test("dismiss button calls onDismiss", () => {
    const onDismiss = mock(() => {})
    const recovery = makeRecovery({
      actions: [{ label: "Cerrar", type: "dismiss" }],
    })
    const { getByText } = render(<ErrorRecovery recovery={recovery} onDismiss={onDismiss} />)
    fireEvent.click(getByText("Cerrar"))
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  test("inline variant does not center text", () => {
    const { container } = render(
      <ErrorRecovery recovery={makeRecovery()} variant="inline" />
    )
    const root = container.firstElementChild as HTMLElement
    expect(root.className).not.toContain("text-center")
  })

  test("page variant centers content", () => {
    const { container } = render(
      <ErrorRecovery recovery={makeRecovery()} variant="page" />
    )
    const root = container.firstElementChild as HTMLElement
    expect(root.className).toContain("text-center")
  })

  test("renders multiple actions", () => {
    const recovery = makeRecovery({
      actions: [
        { label: "Reintentar", type: "retry" },
        { label: "Reportar error", type: "report" },
      ],
    })
    const { getByText } = render(<ErrorRecovery recovery={recovery} />)
    expect(getByText("Reintentar")).toBeTruthy()
    expect(getByText("Reportar error")).toBeTruthy()
  })
})

describe("ErrorRecoveryFromError", () => {
  test("renders recovery from a plain Error", () => {
    const error = new Error("something went wrong")
    const { getByText } = render(<ErrorRecoveryFromError error={error} />)
    expect(getByText("Error inesperado")).toBeTruthy()
  })

  test("reset prop is used as onRetry", () => {
    const reset = mock(() => {})
    const error = new Error("failed")
    const { getByText } = render(<ErrorRecoveryFromError error={error} reset={reset} />)
    fireEvent.click(getByText("Reintentar"))
    expect(reset).toHaveBeenCalledTimes(1)
  })

  test("detects ECONNREFUSED as unavailable", () => {
    const error = new Error("fetch failed: ECONNREFUSED")
    const { getByText } = render(<ErrorRecoveryFromError error={error} />)
    expect(getByText("Servidor no disponible")).toBeTruthy()
  })
})
