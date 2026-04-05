import { test, expect } from "@playwright/test";
import { login } from "./helpers/auth";

test.describe("Settings", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test("settings page loads and shows user email", async ({ page }) => {
    await page.goto("/settings");

    // The settings page shows "Mi cuenta" heading
    await expect(page.locator("text=Mi cuenta").first()).toBeVisible({
      timeout: 10_000,
    });

    // The email input shows the logged-in user's email (readonly)
    const emailInput = page.locator("#email");
    await expect(emailInput).toBeVisible({ timeout: 10_000 });
    await expect(emailInput).toHaveValue("admin@sda.local");
  });

  test("settings page shows personal info section", async ({ page }) => {
    await page.goto("/settings");

    // Section heading — "Información personal" (with accent)
    await expect(
      page.locator("text=Información personal").first(),
    ).toBeVisible({ timeout: 10_000 });

    // Description text
    await expect(
      page.locator("text=Tu nombre y datos de acceso"),
    ).toBeVisible();
  });

  test("settings page has editable name field", async ({ page }) => {
    await page.goto("/settings");

    // The name input should be visible and editable
    const nameInput = page.locator("#name");
    await expect(nameInput).toBeVisible({ timeout: 10_000 });

    // It should have a value (the current user's name)
    const currentName = await nameInput.inputValue();
    expect(currentName.length).toBeGreaterThan(0);
  });

  test("settings page shows notification preferences link", async ({
    page,
  }) => {
    await page.goto("/settings");

    await expect(
      page.locator("text=Preferencias de notificaciones"),
    ).toBeVisible({ timeout: 10_000 });
  });
});
