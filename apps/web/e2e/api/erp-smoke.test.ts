/**
 * ERP smoke tests — every /v1/erp/* read endpoint exercised with the
 * test user's token. Asserts each endpoint:
 *   - is routed (no 404 from Traefik / no 502 = service down)
 *   - is authenticated (200 not 401, given our admin token)
 *   - returns valid JSON (no HTML error page)
 *   - is NOT a 5xx (DB tables exist, query parses, etc.)
 *
 * The full list mirrors the routes the frontend actually calls (extracted
 * via grep over apps/web/src). When the frontend adds a new endpoint, add
 * a row here so we catch DB / handler regressions early.
 */

import { test, expect, beforeAll } from "bun:test";
import { login, apiGet } from "./helpers";

const ERP_READ_ENDPOINTS: { path: string; query?: Record<string, string> }[] = [
  // Master data
  { path: "/v1/erp/catalogs/types" },
  { path: "/v1/erp/entities", query: { type: "customer", page_size: "20" } },
  { path: "/v1/erp/entities", query: { type: "supplier", page_size: "20" } },
  { path: "/v1/erp/entities", query: { type: "employee", page_size: "20" } },

  // Stock / inventory
  { path: "/v1/erp/stock/articles", query: { page_size: "20" } },
  { path: "/v1/erp/stock/levels", query: { page_size: "20" } },
  { path: "/v1/erp/stock/movements", query: { page_size: "20" } },
  { path: "/v1/erp/stock/warehouses" },

  // Accounting
  { path: "/v1/erp/accounting/accounts", query: { page_size: "20" } },
  { path: "/v1/erp/accounting/balance" },
  { path: "/v1/erp/accounting/cost-centers" },
  { path: "/v1/erp/accounting/fiscal-years" },

  // Treasury / current accounts / invoicing
  { path: "/v1/erp/accounts/balances", query: { page_size: "20" } },
  { path: "/v1/erp/accounts/overdue" },
  { path: "/v1/erp/invoicing/invoices", query: { page_size: "20" } },
  { path: "/v1/erp/invoicing/withholdings", query: { page_size: "20" } },

  // Purchasing / sales
  { path: "/v1/erp/purchasing/orders", query: { page_size: "20" } },
  { path: "/v1/erp/purchasing/receipts", query: { page_size: "20" } },
  { path: "/v1/erp/purchasing/inspections", query: { page_size: "20" } },
  { path: "/v1/erp/sales/orders", query: { page_size: "20" } },
  { path: "/v1/erp/sales/quotations", query: { page_size: "20" } },

  // Production & manufacturing
  { path: "/v1/erp/production/orders", query: { page_size: "20" } },
  { path: "/v1/erp/production/units", query: { page_size: "20" } },
  { path: "/v1/erp/production/centers" },
  { path: "/v1/erp/manufacturing/units", query: { page_size: "20" } },
  { path: "/v1/erp/manufacturing/carroceria-models" },

  // HR
  { path: "/v1/erp/hr/employees", query: { page_size: "20" } },
  { path: "/v1/erp/hr/events", query: { page_size: "20" } },
  { path: "/v1/erp/hr/training", query: { page_size: "20" } },

  // Quality
  { path: "/v1/erp/quality/nc", query: { page_size: "20" } },
  { path: "/v1/erp/quality/audits", query: { page_size: "20" } },
  { path: "/v1/erp/quality/documents", query: { page_size: "20" } },
  { path: "/v1/erp/quality/action-plans", query: { page_size: "20" } },

  // Maintenance
  { path: "/v1/erp/maintenance/assets", query: { page_size: "20" } },
  { path: "/v1/erp/maintenance/work-orders", query: { page_size: "20" } },
  { path: "/v1/erp/maintenance/fuel-logs", query: { page_size: "20" } },

  // Safety
  { path: "/v1/erp/safety/accidents", query: { page_size: "20" } },
  { path: "/v1/erp/safety/medical-leaves", query: { page_size: "20" } },
  { path: "/v1/erp/safety/medical-log", query: { page_size: "20" } },
  { path: "/v1/erp/safety/accident-types" },
  { path: "/v1/erp/safety/body-parts" },

  // Analytics
  { path: "/v1/erp/analytics/dashboard/kpis" },
];

let token = "";

beforeAll(async () => {
  token = await login();
});

for (const ep of ERP_READ_ENDPOINTS) {
  test(`GET ${ep.path}${ep.query ? "?" + new URLSearchParams(ep.query).toString() : ""}`, async () => {
    const res = await apiGet(ep.path, { token, query: ep.query });

    // Surface useful detail when the assertion fails — easier than chasing
    // error messages in the bun-test output for 40 endpoints.
    const detail = `status=${res.status} ct=${res.contentType} body=${res.bodyPreview}`;

    // Allow 200 (success) or 204 (no content). Anything else is a regression.
    if (res.status >= 500) {
      throw new Error(`5xx from backend (handler/DB error): ${detail}`);
    }
    if (res.status === 401) {
      throw new Error(`unexpected 401 — admin token should grant access: ${detail}`);
    }
    if (res.status === 404) {
      throw new Error(`route not registered (404 from Traefik or handler): ${detail}`);
    }
    if (res.status === 502 || res.status === 503) {
      throw new Error(`upstream not reachable: ${detail}`);
    }

    expect(res.status).toBeLessThan(400);
    expect(res.contentType).toContain("json");
    expect(res.parseError).toBeNull();
  });
}
