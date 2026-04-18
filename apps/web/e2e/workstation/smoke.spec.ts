/**
 * Smoke suite — visit every public route in the app and assert:
 *   - URL stays on the requested route (no redirect-bounce to /login).
 *   - No 5xx responses for any /v1/* request fired during the visit.
 *   - No uncaught JS errors / page errors.
 *   - No unexpected console errors (HMR + WS subprotocol noise filtered).
 *
 * Each test does a fresh API login in beforeEach to dodge the single-use
 * refresh-token rotation problem (a saved storageState would die after one
 * use as the first test's refresh consumes the token).
 */
import { test, expect, type Page, type Response } from "@playwright/test";

const TARGET = process.env.TARGET ?? "http://172.22.100.23";
const TEST_EMAIL = process.env.TEST_EMAIL ?? "e2e-test@saldivia.local";
const TEST_PASSWORD = process.env.TEST_PASSWORD ?? "testpassword123";

interface Route {
  path: string;
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
  { path: "/feedback" },
];

const SUB_PAGES: Route[] = [
  { path: "/manufactura/unidades" },
  { path: "/manufactura/controles" },
  { path: "/manufactura/certificaciones" },
  { path: "/produccion/ordenes" },
  { path: "/produccion/seguimiento" },
  { path: "/produccion/pcp" },
  { path: "/produccion/preentrega" },
  { path: "/calidad/inspecciones" },
  { path: "/calidad/no-conformidades" },
  { path: "/calidad/trazabilidad" },
  { path: "/compras/proveedores" },
  { path: "/compras/ordenes" },
  { path: "/rrhh/legajos" },
  { path: "/rrhh/licencias" },
  { path: "/seguridad/incidentes" },
  { path: "/seguridad/inspecciones" },
  { path: "/mantenimiento/equipos" },
  { path: "/mantenimiento/preventivo" },
];

const ALL_ROUTES: Route[] = [...CORE, ...MODULE_ROOTS, ...SUB_PAGES];

function isIgnorableConsoleError(text: string): boolean {
  return (
    // Dev-mode HMR client tries webpack-hmr endpoint that bun/turbopack
    // doesn't expose — environmental, not a regression.
    text.includes("webpack-hmr") ||
    text.includes("react-devtools") ||
    text.includes("Download the React DevTools")
  );
}

// Single login for the whole suite. Each test reuses the latest refresh
// token (rotated by the previous test's silent refresh) via cookie injection.
// Avoids /v1/auth/login rate-limit (429) while still testing real auth.
let cachedRefreshToken: string | null = null;

async function loginOnce(): Promise<string> {
  if (cachedRefreshToken) return cachedRefreshToken;
  const res = await fetch(`${TARGET}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: TEST_EMAIL, password: TEST_PASSWORD }),
  });
  if (!res.ok) throw new Error(`login failed: ${res.status} ${await res.text()}`);
  const cookieHeader = res.headers.get("set-cookie") ?? "";
  const match = cookieHeader.match(/sda_refresh=([^;]+)/);
  if (!match) throw new Error(`no sda_refresh in Set-Cookie: ${cookieHeader}`);
  cachedRefreshToken = match[1];
  return cachedRefreshToken;
}

test.beforeEach(async ({ context }) => {
  const refreshToken = await loginOnce();
  const url = new URL(TARGET);
  await context.addCookies([
    {
      name: "sda_refresh",
      value: refreshToken,
      domain: url.hostname,
      path: "/v1/auth",
      httpOnly: true,
      secure: false,
      sameSite: "Lax",
    },
  ]);
});

// After each test the page's silent refresh has rotated the cookie. Capture
// the latest one so the next test's beforeEach starts with a valid token.
test.afterEach(async ({ context }) => {
  const cookies = await context.cookies();
  const fresh = cookies.find((c) => c.name === "sda_refresh");
  if (fresh?.value) cachedRefreshToken = fresh.value;
});

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
    if (res.url().includes("/v1/")) {
      apiCalls.push({ url: res.url(), status: res.status() });
    }
  });

  await page.goto(route.path);
  await page.waitForLoadState("networkidle", { timeout: 8_000 }).catch(() => {});

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
