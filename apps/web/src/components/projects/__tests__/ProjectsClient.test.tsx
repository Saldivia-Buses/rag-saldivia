import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { ProjectsClient } from "@/components/projects/ProjectsClient"

afterEach(cleanup)

const mockProjects = [
  { id: "p1", userId: 1, name: "Proyecto Legal", description: "Análisis contractual", instructions: "Responde en términos legales", createdAt: Date.now(), updatedAt: Date.now() },
  { id: "p2", userId: 1, name: "RRHH", description: "", instructions: "", createdAt: Date.now(), updatedAt: Date.now() },
]

describe("<ProjectsClient />", () => {
  test("renderiza los proyectos", () => {
    const { getByText } = render(<ProjectsClient initialProjects={mockProjects} />)
    expect(getByText("Proyecto Legal")).toBeInTheDocument()
    expect(getByText("RRHH")).toBeInTheDocument()
  })

  test("muestra la descripción del proyecto", () => {
    const { getByText } = render(<ProjectsClient initialProjects={mockProjects} />)
    expect(getByText("Análisis contractual")).toBeInTheDocument()
  })

  test("muestra las instrucciones del proyecto", () => {
    const { getByText } = render(<ProjectsClient initialProjects={mockProjects} />)
    expect(getByText(/Responde en términos/)).toBeInTheDocument()
  })

  test("sin proyectos muestra EmptyPlaceholder", () => {
    const { getByText } = render(<ProjectsClient initialProjects={[]} />)
    expect(getByText("Sin proyectos")).toBeInTheDocument()
  })

  test("botón Nuevo proyecto presente", () => {
    const { getByRole } = render(<ProjectsClient initialProjects={[]} />)
    expect(getByRole("button", { name: /Nuevo proyecto/ })).toBeInTheDocument()
  })

  test("cada proyecto tiene botón Ver sesiones", () => {
    const { getAllByRole } = render(<ProjectsClient initialProjects={mockProjects} />)
    const btns = getAllByRole("button", { name: /Ver sesiones/ })
    expect(btns.length).toBe(2)
  })
})
