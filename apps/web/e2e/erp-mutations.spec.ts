/**
 * ERP mutation tests — create actions.
 *
 * These tests require the app to be running against a test tenant
 * with a seeded database. They are skipped by default and can be
 * run when the full stack is available.
 *
 * To run: set environment variable E2E_MUTATIONS=1
 *   E2E_MUTATIONS=1 npx playwright test erp-mutations
 */

import { test, expect } from "@playwright/test";
import { login } from "./helpers/auth";

const MUTATIONS_ENABLED = !!process.env.E2E_MUTATIONS;

test.describe("ERP Mutations", () => {
  test.beforeEach(async ({ page }) => {
    if (!MUTATIONS_ENABLED) {
      test.skip(
        true,
        "Skipped: set E2E_MUTATIONS=1 to run mutations against a seeded test tenant",
      );
    }
    await login(page);
  });

  // ── Catálogos ────────────────────────────────────────────────────────────

  test("catalogos — can create a new catalog entry", async ({ page }) => {
    await page.goto("/administracion/catalogos");
    await page.waitForLoadState("networkidle");

    // Heading confirms we are on the right page
    await expect(
      page.getByRole("heading", { name: "Catálogos" }),
    ).toBeVisible({ timeout: 15_000 });

    // Open the create dialog
    await page.getByRole("button", { name: /Nueva entrada/i }).click();

    // Dialog title
    await expect(
      page.getByRole("heading", { name: "Nueva entrada de catálogo" }),
    ).toBeVisible({ timeout: 5_000 });

    // Fill in type, code, name
    const unique = Date.now().toString().slice(-6);
    await page.getByLabel("Tipo").fill("test_e2e");
    await page.getByLabel("Código").fill(`E2E-${unique}`);
    await page.getByLabel("Nombre").fill(`Test entrada ${unique}`);

    // Submit
    await page.getByRole("button", { name: /^Crear$/ }).click();

    // Toast confirm
    await expect(page.locator("text=Entrada creada exitosamente")).toBeVisible({
      timeout: 10_000,
    });

    // Dialog should close
    await expect(
      page.getByRole("heading", { name: "Nueva entrada de catálogo" }),
    ).not.toBeVisible({ timeout: 5_000 });

    // The new entry appears in the list (search for the unique code)
    await expect(
      page.locator(`text=E2E-${unique}`),
    ).toBeVisible({ timeout: 10_000 });
  });

  // ── Almacén ──────────────────────────────────────────────────────────────

  test("almacen — can create a new article", async ({ page }) => {
    await page.goto("/administracion/almacen");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Almacén" }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByRole("button", { name: /Nuevo artículo/i }).click();

    // Dialog
    await expect(
      page.getByRole("heading", { name: /Nuevo artículo/i }),
    ).toBeVisible({ timeout: 5_000 });

    const unique = Date.now().toString().slice(-6);
    await page.getByLabel("Código").fill(`ART-${unique}`);
    await page.getByLabel("Nombre").fill(`Artículo E2E ${unique}`);

    // Submit form (button text is "Crear")
    await page.getByRole("button", { name: /^Crear$/ }).click();

    // Success toast
    await expect(
      page.locator("text=Artículo creado exitosamente"),
    ).toBeVisible({ timeout: 10_000 });

    // Dialog closes
    await expect(
      page.getByRole("heading", { name: /Nuevo artículo/i }),
    ).not.toBeVisible({ timeout: 5_000 });
  });

  // ── Tesorería — movimiento ────────────────────────────────────────────────

  test("tesoreria — can register a new movement", async ({ page }) => {
    await page.goto("/administracion/tesoreria");
    await page.waitForLoadState("networkidle");

    await expect(
      page.getByRole("heading", { name: "Tesorería" }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByRole("button", { name: /Nuevo movimiento/i }).click();

    await expect(
      page.getByRole("heading", { name: "Nuevo movimiento" }),
    ).toBeVisible({ timeout: 5_000 });

    const unique = Date.now().toString().slice(-6);
    await page.getByLabel("Número").fill(`MOV-${unique}`);
    await page.getByLabel("Monto").fill("1000");

    await page.getByRole("button", { name: /Registrar/ }).click();

    await expect(
      page.locator("text=Movimiento registrado"),
    ).toBeVisible({ timeout: 10_000 });
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
    await page.goto("/administracion/calidad");
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
