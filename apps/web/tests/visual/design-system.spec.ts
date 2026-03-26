import { test, expect } from "@playwright/test"
import { enableDarkMode, enableLightMode, SNAPSHOT_OPTIONS } from "./helpers"

/**
 * Visual regression tests del design system.
 *
 * IDs de stories: Storybook los genera a partir del title + export name.
 * "Primitivos/Button" + export AllVariants → "primitivos-button--all-variants"
 *
 * Primer run: bun run visual:update (genera baseline)
 * CI: bun run test:visual (compara contra baseline)
 */

const STORIES = [
  // Primitivos
  { id: "primitivos-button--all-variants",          name: "button-all-variants" },
  { id: "primitivos-button--default",               name: "button-default" },
  { id: "primitivos-badge--all-variants",           name: "badge-all-variants" },
  { id: "primitivos-input--all-states",             name: "input-all-states" },
  { id: "primitivos-avatar--with-fallback",         name: "avatar-fallback" },
  { id: "primitivos-table--default",                name: "table-default" },
  { id: "primitivos-skeleton--all-variants",        name: "skeleton-all-variants" },
  // Features
  { id: "features-stat-card--default",              name: "stat-card-default" },
  { id: "features-empty-placeholder--chat",         name: "empty-placeholder-chat" },
  { id: "features-empty-placeholder--all-variants", name: "empty-placeholder-all" },
  // Design System
  { id: "design-system-tokens--palette",            name: "tokens-palette" },
]

for (const story of STORIES) {
  test(`${story.name} — light`, async ({ page }) => {
    await page.goto(`/?path=/story/${story.id}`)
    await page.waitForLoadState("networkidle")
    await enableLightMode(page)
    await expect(page).toHaveScreenshot(`${story.name}-light.png`, SNAPSHOT_OPTIONS)
  })

  test(`${story.name} — dark`, async ({ page }) => {
    await page.goto(`/?path=/story/${story.id}`)
    await page.waitForLoadState("networkidle")
    await enableDarkMode(page)
    await expect(page).toHaveScreenshot(`${story.name}-dark.png`, SNAPSHOT_OPTIONS)
  })
}
