import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { NavRail } from "../NavRail"
import { SidebarProvider } from "@/components/chat/ChatLayout"

afterEach(cleanup)

const USER = { id: 1, email: "admin@test.com", name: "Admin", role: "admin" as const }
const CHANGELOG = { version: "1.0.0", entries: [] }

function renderNavRail(props: Partial<Parameters<typeof NavRail>[0]> = {}) {
  return render(
    <SidebarProvider>
      <NavRail user={USER} changelog={CHANGELOG} {...props} />
    </SidebarProvider>
  )
}

describe("NavRail", () => {
  test("renders nav landmark", () => {
    const { getByLabelText } = renderNavRail()
    expect(getByLabelText("Navegación principal")).toBeTruthy()
  })

  test("renders nav links with aria-labels", () => {
    const { getByLabelText } = renderNavRail()
    expect(getByLabelText("Chat")).toBeTruthy()
    expect(getByLabelText("Colecciones")).toBeTruthy()
    expect(getByLabelText("Configuración")).toBeTruthy()
  })

  test("renders admin link for admin users", () => {
    const { getByLabelText } = renderNavRail()
    expect(getByLabelText("Admin")).toBeTruthy()
  })

  test("hides admin link for regular users", () => {
    const { queryByLabelText } = renderNavRail({ user: { ...USER, role: "user" } })
    expect(queryByLabelText("Admin")).toBeNull()
  })

  test("renders logout button", () => {
    const { getByLabelText } = renderNavRail()
    expect(getByLabelText("Cerrar sesión")).toBeTruthy()
  })

  test("renders messaging link", () => {
    const { getByLabelText } = renderNavRail()
    expect(getByLabelText("Mensajería")).toBeTruthy()
  })

  test("renders brand logo", () => {
    const { getByText } = renderNavRail()
    expect(getByText("S")).toBeTruthy()
  })
})
