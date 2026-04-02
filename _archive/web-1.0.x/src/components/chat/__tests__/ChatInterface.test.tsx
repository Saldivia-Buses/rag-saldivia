import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { ChatInterface } from "../ChatInterface"
import { SidebarProvider } from "../ChatLayout"

afterEach(cleanup)

const SESSION = {
  id: "session-1",
  userId: 1,
  title: "Mi sesión de prueba",
  collection: "docs",
  createdAt: Date.now(),
  updatedAt: Date.now(),
  messages: [],
}

function renderWithSidebar(ui: React.ReactElement) {
  return render(<SidebarProvider>{ui}</SidebarProvider>)
}

describe("ChatInterface", () => {
  test("renders empty state when no messages", () => {
    const { getByText } = renderWithSidebar(
      <ChatInterface session={SESSION} userId={1} />
    )
    expect(getByText("¿En qué pensamos?")).toBeTruthy()
  })

  test("renders input placeholder in empty state", () => {
    const { getByPlaceholderText } = renderWithSidebar(
      <ChatInterface session={SESSION} userId={1} />
    )
    expect(getByPlaceholderText("¿Cómo puedo ayudarte hoy?")).toBeTruthy()
  })

  test("renders suggestion chips when no templates", () => {
    const { getByText } = renderWithSidebar(
      <ChatInterface session={SESSION} userId={1} />
    )
    expect(getByText("Buscar documentos")).toBeTruthy()
    expect(getByText("Hacer preguntas")).toBeTruthy()
  })

  test("renders sr-only heading with collection", () => {
    const { container } = renderWithSidebar(
      <ChatInterface session={SESSION} userId={1} />
    )
    const srHeading = container.querySelector(".sr-only")
    expect(srHeading?.textContent).toContain("docs")
  })

  test("renders disclaimer text", () => {
    const { getByText } = renderWithSidebar(
      <ChatInterface session={SESSION} userId={1} />
    )
    expect(getByText(/puede cometer errores/)).toBeTruthy()
  })
})
