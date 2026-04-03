import { test, expect } from "@playwright/test";

test.describe("Login", () => {
  test("redirects to /login when not authenticated", async ({ page }) => {
    await page.goto("/chat");
    // AuthInitializer redirects unauthenticated users via window.location.href
    await page.waitForURL(/\/login/, { timeout: 10_000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test("login with valid credentials redirects to dashboard", async ({
    page,
  }) => {
    await page.goto("/login");

    // Fill the login form — inputs have id="email" and id="password"
    await page.locator("#email").fill("admin@sda.local");
    await page.locator("#password").fill("admin123");
    await page.locator('button[type="submit"]').click();

    // After login, the component does window.location.href = "/dashboard"
    await page.waitForURL(/\/dashboard/, { timeout: 15_000 });
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test("login with invalid credentials shows error", async ({ page }) => {
    await page.goto("/login");

    await page.locator("#email").fill("wrong@email.com");
    await page.locator("#password").fill("wrongpass");
    await page.locator('button[type="submit"]').click();

    // Error message: "Email o contrasena incorrectos" (401)
    // or "No se pudo conectar al servidor." (network error)
    // Look for the error container that appears on failure
    const errorBox = page.locator(".bg-destructive\\/10");
    await expect(errorBox).toBeVisible({ timeout: 10_000 });
  });

  test("login page shows form elements", async ({ page }) => {
    await page.goto("/login");

    // Heading
    await expect(
      page.getByRole("heading", { name: "Iniciar sesión" }),
    ).toBeVisible({ timeout: 5_000 });

    // Form fields
    await expect(page.locator("#email")).toBeVisible();
    await expect(page.locator("#password")).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();

    // Social login buttons
    await expect(page.locator("text=Google")).toBeVisible();
  });
});
