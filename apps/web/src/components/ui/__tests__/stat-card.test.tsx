import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { StatCard } from "@/components/ui/stat-card"
import { Users } from "lucide-react"

afterEach(cleanup)

describe("<StatCard />", () => {
  test("muestra etiqueta y valor", () => {
    const { getByText } = render(<StatCard label="Usuarios" value={42} />)
    expect(getByText("Usuarios")).toBeInTheDocument()
    expect(getByText("42")).toBeInTheDocument()
  })

  test("delta positivo muestra +N% con clase success", () => {
    const { getByText } = render(<StatCard label="Queries" value={100} delta={15} />)
    const badge = getByText("+15%")
    expect(badge).toBeInTheDocument()
    expect(badge.className).toContain("success")
  })

  test("delta negativo muestra -N% con clase destructive", () => {
    const { getByText } = render(<StatCard label="Errores" value={5} delta={-3} />)
    const badge = getByText("-3%")
    expect(badge).toBeInTheDocument()
    expect(badge.className).toContain("destructive")
  })

  test("sin delta no renderiza porcentaje", () => {
    const { container } = render(<StatCard label="Total" value={99} />)
    expect(container.textContent).not.toContain("%")
  })

  test("deltaLabel se muestra junto al delta", () => {
    const { getByText } = render(<StatCard label="X" value={1} delta={5} deltaLabel="vs mes" />)
    expect(getByText("vs mes")).toBeInTheDocument()
  })

  test("renderiza el ícono cuando se proporciona", () => {
    const { container } = render(<StatCard label="Usuarios" value={10} icon={Users} />)
    expect(container.querySelector("svg")).toBeInTheDocument()
  })

  test("el valor string se renderiza correctamente", () => {
    const { getByText } = render(<StatCard label="Tasa" value="92%" />)
    expect(getByText("92%")).toBeInTheDocument()
  })
})
