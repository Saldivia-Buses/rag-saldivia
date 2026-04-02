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
    // Empty state input uses "¿Cómo puedo ayudarte hoy?" or "Responder..."
    const input = page.getByRole("textbox").first()
    await input.fill(query)
    await input.press("Enter")

    // Wait for mock RAG response
    await expect(page.getByText(query)).toBeVisible({ timeout: 30_000 })

    // Verify message persists after reload
    await page.reload()
    await page.waitForLoadState("networkidle")
    await expect(page.getByText(query)).toBeVisible({ timeout: 10_000 })
  })

  test("sesión aparece en sidebar después de enviar mensaje", async ({ page }) => {
    await page.goto("/chat")
    await page.waitForLoadState("networkidle")

    await page.getByRole("button", { name: "Nueva sesión" }).click()
    await page.waitForURL(/\/chat\/[a-f0-9-]+/)

    const input = page.getByRole("textbox").first()
    await input.fill("Pregunta para sidebar test")
    await input.press("Enter")

    // Wait for response and auto-rename
    await page.waitForTimeout(3000)
    await page.goto("/chat")
    await page.waitForLoadState("networkidle")

    // The session should appear in the sidebar with the auto-generated title
    await expect(page.getByText(/Pregunta para sidebar/i)).toBeVisible({ timeout: 10_000 })
  })

  test("empty state muestra sugerencias", async ({ page }) => {
    await page.goto("/chat")
    await page.waitForLoadState("networkidle")

    // Create a new session to get empty state
    await page.getByRole("button", { name: "Nueva sesión" }).click()
    await page.waitForURL(/\/chat\/[a-f0-9-]+/)

    await expect(page.getByText("¿En qué pensamos?")).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText("Buscar documentos")).toBeVisible()
  })

  test("disclaimer visible en chat", async ({ page }) => {
    await page.goto("/chat")
    await page.waitForLoadState("networkidle")

    await page.getByRole("button", { name: "Nueva sesión" }).click()
    await page.waitForURL(/\/chat\/[a-f0-9-]+/)

    await expect(page.getByText(/puede cometer errores/)).toBeVisible({ timeout: 10_000 })
  })
})
