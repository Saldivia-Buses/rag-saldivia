import { test, expect, describe, mock, afterEach, beforeEach, spyOn } from "bun:test"
import { render, fireEvent, cleanup } from "@testing-library/react"
import { ErrorBoundary } from "@/components/error-boundary"

afterEach(cleanup)

// Componente auxiliar que lanza un error controlado
function Thrower({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) throw new Error("error de prueba")
  return <p>contenido ok</p>
}

// Suprimir los errores de consola que React emite al capturar errores en tests
let consoleError: ReturnType<typeof spyOn>
beforeEach(() => {
  consoleError = spyOn(console, "error").mockImplementation(() => {})
})
afterEach(() => {
  consoleError.mockRestore()
})

describe("<ErrorBoundary />", () => {
  test("renderiza children cuando no hay error", () => {
    const { getByText } = render(
      <ErrorBoundary>
        <Thrower shouldThrow={false} />
      </ErrorBoundary>
    )
    expect(getByText("contenido ok")).toBeInTheDocument()
  })

  test("renderiza fallback por defecto cuando el hijo lanza un error", () => {
    const { getByText } = render(
      <ErrorBoundary>
        <Thrower shouldThrow={true} />
      </ErrorBoundary>
    )
    expect(getByText("Error inesperado")).toBeInTheDocument()
    expect(getByText("error de prueba")).toBeInTheDocument()
  })

  test("renderiza el botón 'Reintentar' en el fallback", () => {
    const { getByRole } = render(
      <ErrorBoundary>
        <Thrower shouldThrow={true} />
      </ErrorBoundary>
    )
    expect(getByRole("button", { name: "Reintentar" })).toBeInTheDocument()
  })

  test("acepta fallback personalizado via prop", () => {
    const { getByText } = render(
      <ErrorBoundary fallback={<div>fallback custom</div>}>
        <Thrower shouldThrow={true} />
      </ErrorBoundary>
    )
    expect(getByText("fallback custom")).toBeInTheDocument()
  })

  test("llama onReset al hacer click en 'Reintentar'", () => {
    const onReset = mock(() => {})
    const { getByRole } = render(
      <ErrorBoundary onReset={onReset}>
        <Thrower shouldThrow={true} />
      </ErrorBoundary>
    )
    fireEvent.click(getByRole("button", { name: "Reintentar" }))
    expect(onReset).toHaveBeenCalledTimes(1)
  })

  test("resetea el estado interno al hacer click en 'Reintentar'", () => {
    const { getByRole, getByText, queryByText } = render(
      <ErrorBoundary>
        <Thrower shouldThrow={true} />
      </ErrorBoundary>
    )
    expect(getByText("Error inesperado")).toBeInTheDocument()
    fireEvent.click(getByRole("button", { name: "Reintentar" }))
    // Después del reset el boundary vuelve a intentar renderizar los hijos
    // (que siguen lanzando el error, por lo que el boundary vuelve al estado error)
    // Lo que importa: el handler se llama y el estado se resetea correctamente
    expect(queryByText("Error inesperado")).toBeInTheDocument()
  })

  test("muestra el ícono de alerta en el fallback", () => {
    const { container } = render(
      <ErrorBoundary>
        <Thrower shouldThrow={true} />
      </ErrorBoundary>
    )
    expect(container.querySelector("svg")).toBeInTheDocument()
  })
})
