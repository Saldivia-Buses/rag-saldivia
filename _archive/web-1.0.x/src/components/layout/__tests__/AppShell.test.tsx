import { describe, test, expect, afterEach } from "bun:test"
import { cleanup, render } from "@testing-library/react"
import { AppShell } from "../AppShell"

afterEach(cleanup)

const USER = { id: 1, email: "admin@test.com", name: "Admin", role: "admin" as const }
const CHANGELOG = { version: "1.0.0", entries: [] }

describe("AppShell", () => {
  test("renders children", () => {
    const { getByText } = render(
      <AppShell user={USER} changelog={CHANGELOG}>
        <p>Page content</p>
      </AppShell>
    )
    expect(getByText("Page content")).toBeTruthy()
  })

  test("renders nav rail", () => {
    const { getByLabelText } = render(
      <AppShell user={USER} changelog={CHANGELOG}>
        <div />
      </AppShell>
    )
    expect(getByLabelText("Navegación principal")).toBeTruthy()
  })
})
