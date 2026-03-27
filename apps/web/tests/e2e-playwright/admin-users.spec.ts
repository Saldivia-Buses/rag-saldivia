import { test, expect } from "@playwright/test"

const TEST_EMAIL = `e2e-test-${Date.now()}@test.local`

test.describe("Admin users CRUD", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("crear usuario, verificar en tabla, desactivar", async ({ page }) => {
    await page.goto("/admin/users")
    await page.waitForLoadState("networkidle")

    await page.getByRole("button", { name: /nuevo usuario/i }).click()

    await page.getByPlaceholder(/nombre completo/i).fill("E2E Usuario")
    await page.getByPlaceholder(/^email$/i).fill(TEST_EMAIL)
    await page.getByPlaceholder(/contraseña/i).first().fill("Test1234!")
    await page.getByRole("button", { name: /crear usuario/i }).click()

    await expect(page.getByText(TEST_EMAIL)).toBeVisible({ timeout: 15_000 })

    const row = page.locator("tbody tr").filter({ hasText: TEST_EMAIL })
    await row.getByRole("button", { name: /desactivar/i }).click()

    await expect(row.getByText(/inactivo/i)).toBeVisible({ timeout: 10_000 })
  })
})
