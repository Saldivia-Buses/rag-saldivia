/**
 * One-shot login that persists session state for the smoke project to reuse.
 * Saves the auth-store + cookies to .auth/state.json so every page visit
 * starts already-logged-in (no /login round-trip per test).
 */
import { test as setup } from "@playwright/test";

const TEST_EMAIL = process.env.TEST_EMAIL ?? "e2e-test@saldivia.local";
const TEST_PASSWORD = process.env.TEST_PASSWORD ?? "testpassword123";
const STATE_PATH = "apps/web/e2e/workstation/.auth/state.json";

setup("authenticate", async ({ page }) => {
  await page.goto("/login");

  // Wait for hydration so the React handler is attached before clicking.
  await page.waitForFunction(
    () => {
      const form = document.querySelector("form");
      if (!form) return false;
      // React 18+ exposes props on the DOM via a __reactProps key.
      const propsKey = Object.keys(form).find((k) => k.startsWith("__reactProps"));
      return !!propsKey;
    },
    { timeout: 10_000 },
  );

  await page.locator("#email").fill(TEST_EMAIL);
  await page.locator("#password").fill(TEST_PASSWORD);
  await page.locator('button[type="submit"]').click();

  // Login redirects to /inicio on success (login5.tsx → router.push).
  await page.waitForURL(/\/inicio/, { timeout: 15_000 });
  await page.context().storageState({ path: STATE_PATH });
});
