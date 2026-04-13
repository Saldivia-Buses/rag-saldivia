/**
 * Module accessibility smoke tests.
 *
 * For every major module route, verifies:
 *   1. Page loads (no 500, no crash, no redirect to error page)
 *   2. Expected h1 heading is visible
 *   3. ErrorState component ("Error cargando …") is NOT shown
 *
 * Tests are read-only — no mutations.
 * Auth is handled with the shared `login` helper.
 */

import { test, expect } from "@playwright/test";
import { login } from "./helpers/auth";

// ─────────────────────────────────────────────────────────────────────────────
// Fixtures: [route, expectedHeading, optional errorText]
// headings are sourced from the actual h1 in each page.tsx
// ─────────────────────────────────────────────────────────────────────────────

const moduleRoutes: Array<[string, string, string?]> = [
  // Administración ERP
  ["/administracion/contable", "Contabilidad", "Error cargando contabilidad"],
  ["/administracion/tesoreria", "Tesorería", "Error cargando tesorería"],
  ["/administracion/facturacion", "Facturación"],
  ["/administracion/compras", "Compras"],
  ["/administracion/ventas", "Ventas"],
  ["/administracion/almacen", "Almacén", "Error cargando almacén"],
  ["/administracion/calidad", "Calidad", "Error cargando calidad"],
  ["/administracion/mantenimiento", "Mantenimiento"],
  ["/administracion/produccion", "Producción"],
  ["/administracion/rrhh", "Recursos Humanos"],
  ["/administracion/catalogos", "Catálogos", "Error cargando catálogos"],
  ["/administracion/clientes", "Clientes"],
  ["/administracion/proveedores", "Proveedores"],
  // Manufactura
  ["/manufactura/unidades", "Unidades", "Error cargando unidades"],
  ["/manufactura/controles", "Controles"],
  ["/manufactura/certificaciones", "Certificaciones"],
  // Seguridad
  ["/seguridad/incidentes", "Incidentes", "Error cargando incidentes"],
  ["/seguridad/medicina", "Medicina"],
  // Calidad module
  ["/calidad/no-conformidades", "No Conformidades"],
  ["/calidad/inspecciones", "Inspecciones de Calidad"],
  // Compras module
  ["/compras/ordenes", "Órdenes de Compra"],
  ["/compras/proveedores", "Proveedores"],
  // Producción module
  ["/produccion/ordenes", "Órdenes de Producción"],
  // Mantenimiento module
  ["/mantenimiento/equipos", "Equipos"],
  ["/mantenimiento/preventivo", "Mantenimiento Preventivo"],
  ["/mantenimiento/correctivo", "Mantenimiento Correctivo"],
  // RRHH module
  ["/rrhh/legajos", "Legajos"],
];

test.describe("Module accessibility — smoke tests", () => {
  // Login once before all tests — reuse the page object across the loop
  // (workers: 1 in playwright.config, so serial execution is safe)
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  for (const [route, heading, knownErrorText] of moduleRoutes) {
    test(`${route} — loads and shows "${heading}"`, async ({ page }) => {
      await page.goto(route);

      // Wait for any data-fetch to settle before checking
      await page.waitForLoadState("networkidle");

      // Heading must be visible
      await expect(
        page.getByRole("heading", { name: heading }),
      ).toBeVisible({ timeout: 15_000 });

      // Generic crash indicators must NOT be present
      await expect(page.locator("text=500")).not.toBeVisible();

      // If the page has a known ErrorState text, it must not be shown
      if (knownErrorText) {
        await expect(page.locator(`text=${knownErrorText}`)).not.toBeVisible();
      }
    });
  }
});

// ─────────────────────────────────────────────────────────────────────────────
// Module index pages — render the subnav card grid
// ─────────────────────────────────────────────────────────────────────────────

test.describe("Module index pages — subnav cards render", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  const indexPages: Array<[string, string]> = [
    ["/administracion", "Administración"],
    ["/manufactura", "Manufactura"],
    ["/seguridad", "Higiene y Seguridad"],
    ["/calidad", "Calidad"],
    ["/compras", "Compras"],
    ["/produccion", "Producción"],
    ["/mantenimiento", "Mantenimiento"],
    ["/rrhh", "RRHH"],
  ];

  for (const [route, heading] of indexPages) {
    test(`${route} — index shows module heading`, async ({ page }) => {
      await page.goto(route);
      await page.waitForLoadState("networkidle");

      await expect(
        page.getByRole("heading", { name: heading }),
      ).toBeVisible({ timeout: 10_000 });

      // Should not redirect to login or crash page
      await expect(page).not.toHaveURL(/\/login/);
      await expect(page.locator("text=500")).not.toBeVisible();
    });
  }
});

// ─────────────────────────────────────────────────────────────────────────────
// Authentication boundary — unauthenticated users are redirected
// (belt-and-suspenders — also covered in login.spec.ts)
// ─────────────────────────────────────────────────────────────────────────────

test.describe("Authentication boundary", () => {
  test("protected ERP route redirects to /login when not authenticated", async ({
    page,
  }) => {
    await page.goto("/administracion/contable");
    await page.waitForURL(/\/login/, { timeout: 10_000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test("protected manufactura route redirects to /login when not authenticated", async ({
    page,
  }) => {
    await page.goto("/manufactura/unidades");
    await page.waitForURL(/\/login/, { timeout: 10_000 });
    await expect(page).toHaveURL(/\/login/);
  });
});
