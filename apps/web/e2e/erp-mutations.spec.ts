/**
 * ERP mutation tests — create actions.
 *
 * Run against the local app with TENANT_SLUG=test (saldivia bench
 * mirror): all create tests execute every run, no skip flag. Each
 * test verifies the row exists after creation (toast alone is not
 * evidence).
 */

import { test, expect } from "@playwright/test";
import { login } from "./helpers/auth";

test.describe("ERP Mutations", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  // ── Catálogos ────────────────────────────────────────────────────────────

  test("catalogos — create + readback", async ({ page }) => {
    await page.goto("/administracion/catalogos");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Catálogos" }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByRole("button", { name: /Nueva entrada/i }).click();
    await expect(
      page.getByRole("heading", { name: "Nueva entrada de catálogo" }),
    ).toBeVisible({ timeout: 5_000 });

    const unique = Date.now().toString().slice(-6);
    await page.getByLabel("Tipo").fill("test_e2e");
    await page.getByLabel("Código").fill(`E2E-${unique}`);
    await page.getByLabel("Nombre").fill(`Test entrada ${unique}`);

    const postReq = page.waitForRequest((req) =>
      req.method() === "POST" && /\/v1\/erp\/.*catalog/i.test(req.url()),
    );
    await page.getByRole("button", { name: /^Crear$/ }).click();
    await postReq;

    await expect(page.locator("text=Entrada creada exitosamente")).toBeVisible({
      timeout: 10_000,
    });
    await expect(
      page.getByRole("heading", { name: "Nueva entrada de catálogo" }),
    ).not.toBeVisible({ timeout: 5_000 });

    await expect(page.locator(`text=E2E-${unique}`)).toBeVisible({
      timeout: 10_000,
    });
  });

  // ── Almacén ──────────────────────────────────────────────────────────────

  test("almacen — create article + readback", async ({ page }) => {
    await page.goto("/administracion/almacen");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Almacén" }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByRole("button", { name: /Nuevo artículo/i }).click();
    await expect(
      page.getByRole("heading", { name: /Nuevo artículo/i }),
    ).toBeVisible({ timeout: 5_000 });

    const unique = Date.now().toString().slice(-6);
    const code = `ART-${unique}`;
    await page.getByLabel("Código").fill(code);
    await page.getByLabel("Nombre").fill(`Artículo E2E ${unique}`);

    const postReq = page.waitForRequest((req) =>
      req.method() === "POST" && /\/v1\/erp\/.*articles?/i.test(req.url()),
    );
    await page.getByRole("button", { name: /^Crear$/ }).click();
    await postReq;

    await expect(
      page.locator("text=Artículo creado exitosamente"),
    ).toBeVisible({ timeout: 10_000 });
    await expect(
      page.getByRole("heading", { name: /Nuevo artículo/i }),
    ).not.toBeVisible({ timeout: 5_000 });

    await expect(page.locator(`text=${code}`)).toBeVisible({ timeout: 10_000 });
  });

  // ── Tesorería — movimiento ────────────────────────────────────────────────

  test("tesoreria — register movement + readback", async ({ page }) => {
    await page.goto("/tesoreria");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Tesorería" }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByRole("button", { name: /Nuevo movimiento/i }).click();
    await expect(
      page.getByRole("heading", { name: "Nuevo movimiento" }),
    ).toBeVisible({ timeout: 5_000 });

    const unique = Date.now().toString().slice(-6);
    const number = `MOV-${unique}`;
    await page.getByLabel("Número").fill(number);
    await page.getByLabel("Monto").fill("1000");

    const postReq = page.waitForRequest((req) =>
      req.method() === "POST" && /\/v1\/erp\/treasury\/movements?/i.test(req.url()),
    );
    await page.getByRole("button", { name: /Registrar/ }).click();
    await postReq;

    await expect(page.locator("text=Movimiento registrado")).toBeVisible({
      timeout: 10_000,
    });
    await expect(page.locator(`text=${number}`)).toBeVisible({ timeout: 10_000 });
  });

  // ── Contabilidad — asiento ────────────────────────────────────────────────

  test("contabilidad — Nuevo asiento dialog opens and has required fields", async ({
    page,
  }) => {
    await page.goto("/administracion/contable");
    await page.waitForLoadState("networkidle");

    await page.getByRole("button", { name: /Nuevo asiento/i }).click();

    await expect(
      page.getByRole("heading", { name: "Nuevo asiento" }),
    ).toBeVisible({ timeout: 5_000 });

    // Required fields are visible
    await expect(page.getByLabel("Concepto")).toBeVisible();
    await expect(
      page.getByRole("button", { name: /Agregar línea/i }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Cancelar" }),
    ).toBeVisible();

    // Close
    await page.getByRole("button", { name: "Cancelar" }).click();
    await expect(
      page.getByRole("heading", { name: "Nuevo asiento" }),
    ).not.toBeVisible({ timeout: 5_000 });
  });

  // ── Calidad — No Conformidad ─────────────────────────────────────────────

  test("calidad — can open NC creation dialog and cancel", async ({ page }) => {
    await page.goto("/calidad/no-conformidades");
    await page.waitForLoadState("networkidle");

    await page.getByRole("button", { name: /Nueva NC/i }).click();

    await expect(
      page.getByRole("heading", { name: "Nueva No Conformidad" }),
    ).toBeVisible({ timeout: 5_000 });

    await expect(page.getByLabel("Número")).toBeVisible();
    await expect(page.getByLabel("Severidad")).toBeVisible();

    await page.getByRole("button", { name: "Cancelar" }).click();
    await expect(
      page.getByRole("heading", { name: "Nueva No Conformidad" }),
    ).not.toBeVisible({ timeout: 5_000 });
  });

  // ── Manufactura — unidad ──────────────────────────────────────────────────

  test("manufactura — Nueva unidad dialog opens with OT number field", async ({
    page,
  }) => {
    await page.goto("/manufactura/unidades");
    await page.waitForLoadState("networkidle");

    await page.getByRole("button", { name: /Nueva unidad/i }).click();

    await expect(
      page.getByRole("heading", { name: "Nueva unidad" }),
    ).toBeVisible({ timeout: 5_000 });

    await expect(page.getByLabel(/Nro OT/i)).toBeVisible();
    await expect(page.getByLabel(/Nro chasis/i)).toBeVisible();
    await expect(page.getByLabel(/Nro motor/i)).toBeVisible();

    await page.getByRole("button", { name: "Cancelar" }).click();
    await expect(
      page.getByRole("heading", { name: "Nueva unidad" }),
    ).not.toBeVisible({ timeout: 5_000 });
  });

  // ── Seguridad — accidente ────────────────────────────────────────────────

  test("seguridad — Registrar accidente dialog opens with incident date field", async ({
    page,
  }) => {
    await page.goto("/seguridad/incidentes");
    await page.waitForLoadState("networkidle");

    await page
      .getByRole("button", { name: /Registrar accidente/i })
      .click();

    await expect(
      page.getByRole("heading", { name: "Registrar Accidente" }),
    ).toBeVisible({ timeout: 5_000 });

    await expect(page.getByLabel("Fecha del accidente")).toBeVisible();

    await page.getByRole("button", { name: "Cancelar" }).click();
    await expect(
      page.getByRole("heading", { name: "Registrar Accidente" }),
    ).not.toBeVisible({ timeout: 5_000 });
  });
});
