import { test, expect } from "@playwright/test"

test.describe("Auth flow", () => {
  test("login exitoso redirige a /chat", async ({ page }) => {
    await page.goto("/login")
    await page.getByLabel("Email").fill("admin@localhost")
    await page.getByLabel("Contraseña", { exact: true }).fill("changeme")
    await page.getByRole("button", { name: /iniciar sesión/i }).click()
    await page.waitForURL(/\/chat/)
    expect(page.url()).toContain("/chat")
  })

  test("login con credenciales incorrectas muestra error", async ({ page }) => {
    await page.goto("/login")
    await page.getByLabel("Email").fill("admin@localhost")
    await page.getByLabel("Contraseña", { exact: true }).fill("wrong-password-xyz")
    await page.getByRole("button", { name: /iniciar sesión/i }).click()
    await expect(page.getByText(/incorrectos|inválid/i)).toBeVisible({ timeout: 10_000 })
    expect(page.url()).toContain("/login")
  })

  test("logout elimina la sesión y redirige a /login", async ({ page }) => {
    const login = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(login.ok()).toBeTruthy()

    await page.goto("/chat")
    await page.waitForLoadState("networkidle")

    await page.getByRole("button", { name: /cerrar sesión/i }).click()
    await page.waitForURL(/\/login/)

    await page.goto("/chat")
    await page.waitForURL(/\/login/)
    expect(page.url()).toContain("/login")
  })
})
