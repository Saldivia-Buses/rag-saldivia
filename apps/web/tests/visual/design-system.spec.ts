import { test, expect } from "@playwright/test"
import { enableDarkMode, enableLightMode, SNAPSHOT_OPTIONS } from "./helpers"

/**
 * Visual regression tests del design system.
 *
 * Flujo:
 * 1. Storybook corre en localhost:6006 (webServer en playwright.config.ts)
 * 2. Cada test captura la story en light y dark mode
 * 3. Al primer run: `bun run visual:update` genera los snapshots de baseline
 * 4. En CI: comparar contra el baseline — diff > threshold falla el test
 */

const STORIES = [
  { id: "primitivos-button--all-variants",        name: "button-all-variants" },
  { id: "primitivos-badge--all-variants",         name: "badge-all-variants" },
  { id: "primitivos-input--all-states",           name: "input-all-states" },
  { id: "primitivos-avatar--with-fallback",       name: "avatar-fallback" },
  { id: "primitivos-table--default",              name: "table-default" },
  { id: "primitivos-skeleton--all-variants",      name: "skeleton-all-variants" },
  { id: "features-stat-card--default",            name: "stat-card-default" },
  { id: "features-empty-placeholder--all-variants", name: "empty-placeholder-all" },
  { id: "design-system-tokens--palette",          name: "tokens-palette" },
]

for (const story of STORIES) {
  test(`${story.name} — light mode`, async ({ page }) => {
    await page.goto(`/?path=/story/${story.id}`)
    await page.waitForLoadState("networkidle")
    await enableLightMode(page)
    await expect(page).toHaveScreenshot(`${story.name}-light.png`, SNAPSHOT_OPTIONS)
  })

  test(`${story.name} — dark mode`, async ({ page }) => {
    await page.goto(`/?path=/story/${story.id}`)
    await page.waitForLoadState("networkidle")
    await enableDarkMode(page)
    await expect(page).toHaveScreenshot(`${story.name}-dark.png`, SNAPSHOT_OPTIONS)
  })
}

// Layout tests
test("navrail — light mode", async ({ page }) => {
  await page.goto("/?path=/story/layout-navrail--default")
  await page.waitForLoadState("networkidle")
  await enableLightMode(page)
  await expect(page).toHaveScreenshot("navrail-light.png", SNAPSHOT_OPTIONS)
})

test("navrail — dark mode", async ({ page }) => {
  await page.goto("/?path=/story/layout-navrail--default")
  await page.waitForLoadState("networkidle")
  await enableDarkMode(page)
  await expect(page).toHaveScreenshot("navrail-dark.png", SNAPSHOT_OPTIONS)
})
