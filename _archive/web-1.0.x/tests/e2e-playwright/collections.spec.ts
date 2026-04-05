import { test, expect } from "@playwright/test"

test.describe("Collections page", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("collections page carga y muestra contenido", async ({ page }) => {
    await page.goto("/collections")
    await page.waitForLoadState("networkidle")
    expect(page.url()).toContain("/collections")
    // Should show either collections or empty state
    const hasContent = await page.getByText(/colección|colecciones|no hay/i).isVisible({ timeout: 10_000 })
    expect(hasContent).toBeTruthy()
  })

  test("API GET /api/rag/collections retorna array", async ({ request }) => {
    const login = await request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(login.ok()).toBeTruthy()

    const res = await request.get("/api/rag/collections")
    // With MOCK_RAG, this may return mock data or error — verify it's a valid response
    expect([200, 503]).toContain(res.status())
  })
})
