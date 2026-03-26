import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { SettingsClient } from "@/components/settings/SettingsClient"

afterEach(cleanup)

mock.module("@/app/actions/settings", () => ({
  actionUpdateProfile: mock(() => Promise.resolve()),
  actionUpdatePassword: mock(() => Promise.resolve()),
  actionUpdatePreferences: mock(() => Promise.resolve()),
}))

const mockUser = {
  id: 1, email: "admin@test.com", name: "Admin (dev)", role: "admin" as const,
  active: true, passwordHash: "", apiKeyHash: "",
  preferences: { theme: "system" }, createdAt: Date.now(), onboardingCompleted: false,
}

describe("<SettingsClient />", () => {
  test("renderiza las tabs de navegación", () => {
    const { getByRole } = render(<SettingsClient user={mockUser} />)
    expect(getByRole("button", { name: "Perfil" })).toBeInTheDocument()
    expect(getByRole("button", { name: "Contraseña" })).toBeInTheDocument()
    expect(getByRole("button", { name: "Preferencias" })).toBeInTheDocument()
  })

  test("tab Perfil muestra el nombre del usuario", () => {
    const { getByDisplayValue } = render(<SettingsClient user={mockUser} />)
    expect(getByDisplayValue("Admin (dev)")).toBeInTheDocument()
  })

  test("el email aparece como campo deshabilitado", () => {
    const { getByDisplayValue } = render(<SettingsClient user={mockUser} />)
    const emailInput = getByDisplayValue("admin@test.com")
    expect(emailInput).toBeDisabled()
  })

  test("tab Contraseña muestra campos de password", () => {
    const { getByRole, getByText } = render(<SettingsClient user={mockUser} />)
    fireEvent.click(getByRole("button", { name: "Contraseña" }))
    expect(getByText("Contraseña actual")).toBeInTheDocument()
    expect(getByText("Nueva contraseña")).toBeInTheDocument()
  })

  test("tab Preferencias muestra toggles", () => {
    const { getByRole, getByText } = render(<SettingsClient user={mockUser} />)
    fireEvent.click(getByRole("button", { name: "Preferencias" }))
    expect(getByText("Tema")).toBeInTheDocument()
  })

  test("botón Guardar cambios presente", () => {
    const { getByRole } = render(<SettingsClient user={mockUser} />)
    expect(getByRole("button", { name: /Guardar cambios/ })).toBeInTheDocument()
  })
})
