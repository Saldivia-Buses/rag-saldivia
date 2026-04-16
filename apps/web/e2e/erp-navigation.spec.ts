import { test, expect } from "@playwright/test";
import { login } from "./helpers/auth";

test.describe("ERP Navigation — read-only smoke tests", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // ── Contabilidad ─────────────────────────────────────────────────────────

  test("contabilidad — page loads and shows Libro Diario tab", async ({
    page,
  }) => {
    await page.goto("/administracion/contable");
    await page.waitForLoadState("networkidle");

    // Page heading
    await expect(
      page.getByRole("heading", { name: "Contabilidad" }),
    ).toBeVisible({ timeout: 15_000 });

    // The default active tab is Libro Diario
    await expect(page.getByRole("tab", { name: /Libro Diario/i })).toBeVisible();

    // No crash: error component is NOT shown
    await expect(page.locator("text=Error cargando contabilidad")).not.toBeVisible();
  });

  test("contabilidad — balance tab shows table headers", async ({ page }) => {
    await page.goto("/administracion/contable");
    await page.waitForLoadState("networkidle");

    await page.getByRole("tab", { name: /Balance/i }).click();

    // Table headers confirm the right tab is active
    await expect(page.getByRole("columnheader", { name: "Debe" })).toBeVisible({
      timeout: 10_000,
    });
    await expect(
      page.getByRole("columnheader", { name: "Haber" }),
    ).toBeVisible();
    await expect(
      page.getByRole("columnheader", { name: "Saldo" }),
    ).toBeVisible();
  });

  test("contabilidad — Nuevo asiento button is visible", async ({ page }) => {
    await page.goto("/administracion/contable");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("button", { name: /Nuevo asiento/i }),
    ).toBeVisible({ timeout: 10_000 });
  });

  // ── Tesorería ────────────────────────────────────────────────────────────

  test("tesoreria — page loads and shows Movimientos tab", async ({ page }) => {
    await page.goto("/administracion/tesoreria");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Tesorería" }),
    ).toBeVisible({ timeout: 15_000 });

    await expect(
      page.getByRole("tab", { name: /Movimientos/i }),
    ).toBeVisible();

    await expect(
      page.locator("text=Error cargando tesorería"),
    ).not.toBeVisible();
  });

  test("tesoreria — movements table renders (empty state or data)", async ({
    page,
  }) => {
    await page.goto("/administracion/tesoreria");
    await page.waitForLoadState("networkidle");

    // The Movimientos tab is active by default — table must be present
    const movementsTable = page.getByRole("table").first();
    await expect(movementsTable).toBeVisible({ timeout: 10_000 });

    // Either rows are present OR empty state text
    const isEmpty = await page
      .locator("text=Sin movimientos.")
      .isVisible()
      .catch(() => false);
    const hasRows = await page
      .locator("tbody tr")
      .first()
      .isVisible()
      .catch(() => false);

    expect(isEmpty || hasRows).toBe(true);
  });

  // ── Manufactura / Unidades ────────────────────────────────────────────────

  test("manufactura — Unidades page loads with correct heading", async ({
    page,
  }) => {
    await page.goto("/manufactura/unidades");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Unidades" }),
    ).toBeVisible({ timeout: 15_000 });

    await expect(
      page.locator("text=Error cargando unidades"),
    ).not.toBeVisible();
  });

  test("manufactura — Unidades shows table with OT column header", async ({
    page,
  }) => {
    await page.goto("/manufactura/unidades");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("columnheader", { name: "OT" }),
    ).toBeVisible({ timeout: 10_000 });

    await expect(
      page.getByRole("button", { name: /Nueva unidad/i }),
    ).toBeVisible();
  });

  // ── Seguridad / Incidentes ────────────────────────────────────────────────

  test("seguridad — incidentes page loads without crash", async ({ page }) => {
    await page.goto("/seguridad/incidentes");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Incidentes" }),
    ).toBeVisible({ timeout: 15_000 });

    await expect(
      page.locator("text=Error cargando incidentes"),
    ).not.toBeVisible();
  });

  test("seguridad — incidentes shows employee and type columns", async ({
    page,
  }) => {
    await page.goto("/seguridad/incidentes");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("columnheader", { name: "Empleado" }),
    ).toBeVisible({ timeout: 10_000 });

    await expect(
      page.getByRole("columnheader", { name: "Tipo de accidente" }),
    ).toBeVisible();
  });

  // ── Calidad — Planes de acción ────────────────────────────────────────────

  test("calidad — page loads and Planes de accion tab exists", async ({
    page,
  }) => {
    await page.goto("/administracion/calidad");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Calidad" }),
    ).toBeVisible({ timeout: 15_000 });

    // The Planes de acción tab must exist in the tab list
    await expect(
      page.getByRole("tab", { name: /Planes de acción/i }),
    ).toBeVisible({ timeout: 10_000 });

    await expect(
      page.locator("text=Error cargando calidad"),
    ).not.toBeVisible();
  });

  test("calidad — clicking Planes de accion tab shows plan table headers", async ({
    page,
  }) => {
    await page.goto("/administracion/calidad");
    await page.waitForLoadState("networkidle");

    await page.getByRole("tab", { name: /Planes de acción/i }).click();

    await expect(
      page.getByRole("columnheader", { name: "Descripción" }),
    ).toBeVisible({ timeout: 10_000 });

    await expect(
      page.getByRole("button", { name: /Nuevo plan/i }),
    ).toBeVisible();
  });
});
