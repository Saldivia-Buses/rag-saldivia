import { test, expect } from "@playwright/test"

test.describe("Navigation between sections", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("navegar entre chat, collections, settings via NavRail", async ({ page }) => {
    // Start at chat
    await page.goto("/chat")
    await page.waitForLoadState("networkidle")
    expect(page.url()).toContain("/chat")

    // Go to collections
    await page.getByLabel("Colecciones").click()
    await page.waitForURL(/\/collections/)
    expect(page.url()).toContain("/collections")

    // Go to settings
    await page.getByLabel("Configuración").click()
    await page.waitForURL(/\/settings/)
    expect(page.url()).toContain("/settings")

    // Back to chat
    await page.getByLabel("Chat").click()
    await page.waitForURL(/\/chat/)
    expect(page.url()).toContain("/chat")
  })

  test("admin tabs navegan correctamente", async ({ page }) => {
    await page.goto("/admin")
    await page.waitForLoadState("networkidle")

    await page.getByText("Usuarios").click()
    await page.waitForURL(/\/admin\/users/)

    await page.getByText("Roles").click()
    await page.waitForURL(/\/admin\/roles/)

    await page.getByText("Dashboard").click()
    await page.waitForURL(/\/admin$/)
  })
})
