import { test } from "@playwright/test"
import { checkA11y, injectAxe } from "axe-playwright"

/**
 * Auditoría WCAG 2.1 AA en páginas críticas.
 *
 * Prerequisitos:
 * - App corriendo: MOCK_RAG=true bun run dev
 * - Usuario admin logueado (se configura via cookies)
 *
 * Correr: bun run test:a11y
 */

// Autenticación: inyectar cookie de sesión válida antes de cada test
test.beforeEach(async ({ page }) => {
  // Login programático para obtener cookie
  const res = await page.request.post("http://localhost:3000/api/auth/login", {
    data: { email: "admin@localhost", password: "changeme" },
  })
  if (res.ok()) {
    const cookies = await page.context().cookies()
    await page.context().addCookies(cookies)
  }
})

const PAGES_TO_AUDIT = [
  { name: "login",        path: "/login",        requiresAuth: false },
  { name: "chat",         path: "/chat",         requiresAuth: true  },
  { name: "collections",  path: "/collections",  requiresAuth: true  },
  { name: "admin-users",  path: "/admin/users",  requiresAuth: true  },
  { name: "settings",     path: "/settings",     requiresAuth: true  },
]

for (const p of PAGES_TO_AUDIT) {
  test(`a11y: ${p.name} — sin violations WCAG AA`, async ({ page }) => {
    await page.goto(p.path)
    await page.waitForLoadState("networkidle")
    await injectAxe(page)
    await checkA11y(page, undefined, {
      detailedReport: true,
      runOnly: {
        type: "tag",
        values: ["wcag2a", "wcag2aa"],
      },
      // Ignorar violaciones conocidas y aceptadas del design system
      axeOptions: {
        rules: {
          // El radio del color en algunos elementos de Storybook — solo en dev
          "color-contrast": { enabled: true },
        },
      },
    })
  })
}
