import { type Page } from "@playwright/test";

/**
 * Logs in via the UI and waits for redirect to an authenticated page.
 * Reusable across all test files that need an authenticated session.
 */
export async function login(
  page: Page,
  email = "admin@sda.local",
  password = "admin123",
) {
  await page.goto("/login");

  await page.locator("#email").fill(email);
  await page.locator("#password").fill(password);
  await page.locator('button[type="submit"]').click();

  // After login, the app redirects to /dashboard
  await page.waitForURL(/\/dashboard/, { timeout: 15_000 });
}
