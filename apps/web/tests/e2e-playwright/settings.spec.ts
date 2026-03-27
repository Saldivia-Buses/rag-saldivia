import { test, expect } from "@playwright/test"

const NEW_PASSWORD = "NewPass123!"

test.describe("Settings — cambio de contraseña", () => {
  test("cambiar contraseña y verificar login con la nueva", async ({ page }) => {
    const login = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(login.ok()).toBeTruthy()

    await page.goto("/settings")
    await page.waitForLoadState("networkidle")

    await page.getByRole("button", { name: "Contraseña" }).click()

    const pwds = page.locator('input[type="password"]')
    await pwds.nth(0).fill("changeme")
    await pwds.nth(1).fill(NEW_PASSWORD)

    await page.getByRole("button", { name: /actualizar contraseña/i }).click()

    await expect(page.getByText(/contraseña actualizada/i)).toBeVisible({ timeout: 15_000 })

    await page.request.delete("/api/auth/logout")

    await page.goto("/login")
    await page.getByLabel("Email").fill("admin@localhost")
    await page.getByLabel("Contraseña", { exact: true }).fill(NEW_PASSWORD)
    await page.getByRole("button", { name: /iniciar sesión/i }).click()
    await page.waitForURL(/\/chat/)

    await page.goto("/settings")
    await page.getByRole("button", { name: "Contraseña" }).click()
    const pwds2 = page.locator('input[type="password"]')
    await pwds2.nth(0).fill(NEW_PASSWORD)
    await pwds2.nth(1).fill("changeme")
    await page.getByRole("button", { name: /actualizar contraseña/i }).click()
    await expect(page.getByText(/contraseña actualizada/i)).toBeVisible({ timeout: 15_000 })
  })
})
