import { test, expect } from "@playwright/test"
import { writeFileSync, mkdtempSync } from "fs"
import { join } from "path"
import { tmpdir } from "os"

test.describe("Upload flow (MOCK_RAG=true)", () => {
  test.beforeEach(async ({ page }) => {
    const res = await page.request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(res.ok()).toBeTruthy()
  })

  test("subir archivo encola ingesta y muestra éxito en la UI", async ({ page }) => {
    const dir = mkdtempSync(join(tmpdir(), "rag-e2e-upload-"))
    const name = `e2e-upload-${Date.now()}.txt`
    const filePath = join(dir, name)
    writeFileSync(filePath, "test content e2e")

    await page.goto("/upload")
    await page.waitForLoadState("networkidle")

    await page.setInputFiles('input[type="file"]', filePath)

    await expect(page.getByText(name)).toBeVisible({ timeout: 30_000 })
    await expect(page.locator(".text-success").first()).toBeVisible({ timeout: 20_000 })
  })
})
