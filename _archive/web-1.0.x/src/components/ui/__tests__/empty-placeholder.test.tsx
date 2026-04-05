import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { MessageSquare } from "lucide-react"
import { Button } from "@/components/ui/button"

afterEach(cleanup)

describe("<EmptyPlaceholder />", () => {
  test("renderiza título y descripción", () => {
    const { getByText } = render(
      <EmptyPlaceholder>
        <EmptyPlaceholder.Title>Sin sesiones</EmptyPlaceholder.Title>
        <EmptyPlaceholder.Description>Creá una nueva.</EmptyPlaceholder.Description>
      </EmptyPlaceholder>
    )
    expect(getByText("Sin sesiones")).toBeInTheDocument()
    expect(getByText("Creá una nueva.")).toBeInTheDocument()
  })

  test("renderiza ícono cuando se proporciona", () => {
    const { container } = render(
      <EmptyPlaceholder>
        <EmptyPlaceholder.Icon icon={MessageSquare} />
        <EmptyPlaceholder.Title>Vacío</EmptyPlaceholder.Title>
      </EmptyPlaceholder>
    )
    expect(container.querySelector("svg")).toBeInTheDocument()
  })

  test("renderiza children (botón de acción)", () => {
    const { getByRole } = render(
      <EmptyPlaceholder>
        <EmptyPlaceholder.Title>Sin proyectos</EmptyPlaceholder.Title>
        <Button>Crear proyecto</Button>
      </EmptyPlaceholder>
    )
    expect(getByRole("button", { name: "Crear proyecto" })).toBeInTheDocument()
  })

  test("acepta className adicional", () => {
    const { container } = render(
      <EmptyPlaceholder className="max-w-sm">
        <EmptyPlaceholder.Title>Test</EmptyPlaceholder.Title>
      </EmptyPlaceholder>
    )
    expect(container.firstChild?.className ?? "").toContain("max-w-sm")
  })

  test("tiene border-dashed en el contenedor", () => {
    const { container } = render(
      <EmptyPlaceholder>
        <EmptyPlaceholder.Title>Test</EmptyPlaceholder.Title>
      </EmptyPlaceholder>
    )
    expect(container.firstChild?.className ?? "").toContain("border-dashed")
  })
})
