import type { Page } from "@playwright/test"

/**
 * Activar dark mode en Storybook.
 * El design system usa class-based dark mode (next-themes attribute="class"),
 * NO emulación de media query — por eso no usamos colorScheme: 'dark' en el config.
 */
export async function enableDarkMode(page: Page) {
  await page.evaluate(() => {
    document.documentElement.classList.add("dark")
    localStorage.setItem("theme", "dark")
  })
  await page.waitForTimeout(200) // esperar transición CSS
}

export async function enableLightMode(page: Page) {
  await page.evaluate(() => {
    document.documentElement.classList.remove("dark")
    localStorage.setItem("theme", "light")
  })
  await page.waitForTimeout(200)
}

/**
 * Navegar a una story de Storybook y esperar que esté lista.
 */
export async function goToStory(page: Page, storyId: string) {
  await page.goto(`/?path=/story/${storyId}`)
  // Esperar que el iframe de Storybook cargue el componente
  await page.waitForSelector("#storybook-preview-iframe", { timeout: 10_000 })
  await page.waitForTimeout(300) // esperar animaciones de entrada
}

export const SNAPSHOT_OPTIONS = {
  // Tolerancia para diferencias de antialiasing entre plataformas
  threshold: 0.01,
  maxDiffPixels: 10,
}
