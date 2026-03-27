import { test, expect } from "@playwright/test"

test.describe("Chat flow (MOCK_RAG=true)", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("crear sesión nueva y enviar mensaje", async ({ page }) => {
    await page.goto("/chat")
    await page.waitForLoadState("networkidle")

    await page.getByRole("button", { name: "Nueva sesión" }).click()
    await page.waitForURL(/\/chat\/[a-f0-9-]+/)

    const query = "Hola test e2e"
    const input = page.getByPlaceholder(/preguntá sobre/i)
    await input.fill(query)
    await input.press("Enter")

    await expect(page.getByText(/respuesta simulada del RAG/i)).toBeVisible({ timeout: 30_000 })
    await expect(page.getByText(query)).toBeVisible()

    await page.reload()
    await page.waitForLoadState("networkidle")
    await expect(page.getByText(query)).toBeVisible()
    await expect(page.getByText("Nueva sesión").first()).toBeVisible()
  })
})
