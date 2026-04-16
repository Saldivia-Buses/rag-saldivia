/**
 * Smoke suite — visit every public route in the app and assert:
 *   - Page navigates without redirect-bouncing (URL contains the requested path).
 *   - No 5xx responses for any /v1/* request fired during the visit.
 *   - No uncaught JS errors / page errors.
 *
 * Console errors are filtered to ignore the dev-mode HMR WebSocket noise
 * and React DevTools nag — both are environmental, not regressions.
 */
import { test, expect, type Page, type Request, type Response } from "@playwright/test";

interface Route {
  path: string;
  /** Optional locator that must be visible — extra confidence the page rendered. */
  expectVisible?: string;
}

const CORE: Route[] = [
  { path: "/inicio" },
  { path: "/chat" },
  { path: "/collections" },
  { path: "/notifications" },
];

const MODULE_ROOTS: Route[] = [
  { path: "/manufactura" },
  { path: "/produccion" },
  { path: "/calidad" },
  { path: "/ingenieria" },
  { path: "/mantenimiento" },
  { path: "/compras" },
  { path: "/administracion" },
  { path: "/rrhh" },
  { path: "/seguridad" },
  { path: "/astro" },
  { path: "/feedback" },
];

const SUB_PAGES: Route[] = [
  // Manufactura
  { path: "/manufactura/unidades" },
  { path: "/manufactura/controles" },
  { path: "/manufactura/certificaciones" },
  // Producción
  { path: "/produccion/ordenes" },
  { path: "/produccion/seguimiento" },
  { path: "/produccion/pcp" },
  { path: "/produccion/preentrega" },
  // Calidad
  { path: "/calidad/inspecciones" },
  { path: "/calidad/no-conformidades" },
  { path: "/calidad/trazabilidad" },
  // Compras
  { path: "/compras/proveedores" },
  { path: "/compras/ordenes" },
  // RRHH
  { path: "/rrhh/legajos" },
  { path: "/rrhh/licencias" },
  // Seguridad
  { path: "/seguridad/incidentes" },
  { path: "/seguridad/inspecciones" },
  // Mantenimiento
  { path: "/mantenimiento/equipos" },
  { path: "/mantenimiento/preventivo" },
];

const ALL_ROUTES: Route[] = [...CORE, ...MODULE_ROOTS, ...SUB_PAGES];

function isIgnorableConsoleError(text: string): boolean {
  return (
    text.includes("webpack-hmr") ||
    text.includes("react-devtools") ||
    text.includes("Download the React DevTools")
  );
}

async function visitAndAssert(page: Page, route: Route) {
  const apiCalls: { url: string; status: number }[] = [];
  const consoleErrors: string[] = [];
  const pageErrors: string[] = [];

  page.on("console", (msg) => {
    if (msg.type() === "error" && !isIgnorableConsoleError(msg.text())) {
      consoleErrors.push(msg.text());
    }
  });
  page.on("pageerror", (err) => pageErrors.push(err.message));
  page.on("response", (res: Response) => {
    const url = res.url();
    if (url.includes("/v1/")) {
      apiCalls.push({ url, status: res.status() });
    }
  });

  await page.goto(route.path);

  // Some pages load data after mount — give them a moment.
  await page.waitForLoadState("networkidle", { timeout: 8_000 }).catch(() => {});

  // Asserts
  expect(page.url(), `URL mismatch on ${route.path}`).toContain(route.path);

  const fivexx = apiCalls.filter((c) => c.status >= 500);
  if (fivexx.length > 0) {
    throw new Error(
      `5xx on ${route.path}:\n` + fivexx.map((c) => `  ${c.status} ${c.url}`).join("\n"),
    );
  }

  if (consoleErrors.length > 0) {
    throw new Error(
      `console errors on ${route.path}:\n  - ` + consoleErrors.join("\n  - "),
    );
  }

  if (pageErrors.length > 0) {
    throw new Error(
      `pageerror on ${route.path}:\n  - ` + pageErrors.join("\n  - "),
    );
  }

  if (route.expectVisible) {
    await expect(page.locator(route.expectVisible)).toBeVisible({ timeout: 5_000 });
  }
}

for (const route of ALL_ROUTES) {
  test(`visit ${route.path}`, async ({ page }) => {
    await visitAndAssert(page, route);
  });
}
