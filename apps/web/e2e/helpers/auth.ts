import { type Page } from "@playwright/test";

export const TEST_EMAIL = process.env.TEST_EMAIL ?? "e2e@sda.local";
export const TEST_PASSWORD = process.env.TEST_PASSWORD ?? "e2e-saldivia-2026!";

export async function login(
  page: Page,
  email = TEST_EMAIL,
  password = TEST_PASSWORD,
) {
  await page.goto("/login");

  await page.locator("#email").fill(email);
  await page.locator("#password").fill(password);
  await page.locator('button[type="submit"]').click();

  await page.waitForURL(/\/dashboard/, { timeout: 15_000 });
}
