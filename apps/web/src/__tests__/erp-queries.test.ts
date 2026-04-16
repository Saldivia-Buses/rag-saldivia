/**
 * ERP query key factory unit tests.
 *
 * Tests erpKeys from lib/erp/queries.ts.
 *
 * Key invariants:
 * 1. All keys start with ["erp"] so queryClient.invalidateQueries({ queryKey: erpKeys.all })
 *    invalidates the entire ERP cache.
 * 2. Different parameters produce different keys (cache isolation).
 * 3. Same parameters produce structurally equal keys (cache hits).
 */

import { describe, it, expect } from "bun:test";
import { erpKeys } from "@/lib/erp/queries";

describe("erpKeys.all", () => {
  it("is ['erp']", () => {
    expect(erpKeys.all).toEqual(["erp"]);
  });
});

describe("erpKeys root invariant", () => {
  it("every factory key starts with 'erp'", () => {
    const keys = [
      erpKeys.accounts(),
      erpKeys.entries(),
      erpKeys.balance(),
      erpKeys.fiscalYears(),
      erpKeys.stockLevels(),
      erpKeys.warehouses(),
      erpKeys.withholdings(),
      erpKeys.treasuryMovements(),
      erpKeys.treasuryBalance(),
      erpKeys.bankAccounts(),
      erpKeys.checks(),
      erpKeys.accountBalances(),
      erpKeys.accountOverdue(),
      erpKeys.employees(),
      erpKeys.dashboardKPIs(),
    ];
    for (const key of keys) {
      expect(key[0]).toBe("erp");
    }
  });
});

describe("erpKeys.accounts", () => {
  it("starts with erp", () => {
    expect(erpKeys.accounts()[0]).toBe("erp");
  });

  it("includes 'accounts' segment", () => {
    expect(erpKeys.accounts()).toContain("accounts");
  });
});

describe("erpKeys.entries", () => {
  it("no params key differs from key with params", () => {
    const noParams = erpKeys.entries();
    const withParams = erpKeys.entries({ page: "1" });
    expect(JSON.stringify(noParams)).not.toBe(JSON.stringify(withParams));
  });

  it("different page params produce different keys", () => {
    const page1 = erpKeys.entries({ page: "1" });
    const page2 = erpKeys.entries({ page: "2" });
    expect(JSON.stringify(page1)).not.toBe(JSON.stringify(page2));
  });

  it("same params produce equal keys", () => {
    const a = erpKeys.entries({ page: "1", status: "posted" });
    const b = erpKeys.entries({ page: "1", status: "posted" });
    expect(JSON.stringify(a)).toBe(JSON.stringify(b));
  });
});

describe("erpKeys.entry", () => {
  it("different ids produce different keys", () => {
    const k1 = erpKeys.entry("id-1");
    const k2 = erpKeys.entry("id-2");
    expect(JSON.stringify(k1)).not.toBe(JSON.stringify(k2));
  });

  it("includes the id in the key", () => {
    const k = erpKeys.entry("abc-123");
    expect(k).toContain("abc-123");
  });
});

describe("erpKeys.entities", () => {
  it("different types produce different keys", () => {
    const customer = erpKeys.entities("customer");
    const supplier = erpKeys.entities("supplier");
    expect(JSON.stringify(customer)).not.toBe(JSON.stringify(supplier));
  });

  it("same type + search produce equal keys", () => {
    const a = erpKeys.entities("customer", "saldivia");
    const b = erpKeys.entities("customer", "saldivia");
    expect(JSON.stringify(a)).toBe(JSON.stringify(b));
  });

  it("different search terms produce different keys", () => {
    const a = erpKeys.entities("customer", "foo");
    const b = erpKeys.entities("customer", "bar");
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });
});

describe("erpKeys.ledger", () => {
  it("different accountIds produce different keys", () => {
    const a = erpKeys.ledger("acc-1");
    const b = erpKeys.ledger("acc-2");
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });

  it("undefined accountId is a valid key", () => {
    expect(() => erpKeys.ledger()).not.toThrow();
  });
});

describe("erpKeys.analytics", () => {
  it("returns correct structure for domain + report", () => {
    const k = erpKeys.analytics("erp", "kpis");
    expect(k[0]).toBe("erp");
    expect(k).toContain("analytics");
    expect(k).toContain("erp");
    expect(k).toContain("kpis");
  });

  it("different domains produce different keys", () => {
    const a = erpKeys.analytics("erp", "kpis");
    const b = erpKeys.analytics("hr", "kpis");
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });

  it("different reports produce different keys", () => {
    const a = erpKeys.analytics("erp", "kpis");
    const b = erpKeys.analytics("erp", "yoy");
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });

  it("params are included in key", () => {
    const withParams = erpKeys.analytics("erp", "kpis", { year: "2024" });
    const noParams = erpKeys.analytics("erp", "kpis");
    expect(JSON.stringify(withParams)).not.toBe(JSON.stringify(noParams));
  });
});

describe("erpKeys.stockArticles", () => {
  it("no params differs from params", () => {
    const a = erpKeys.stockArticles();
    const b = erpKeys.stockArticles({ category: "fuel" });
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });
});

describe("erpKeys.purchaseOrders", () => {
  it("different statuses produce different keys", () => {
    const pending = erpKeys.purchaseOrders("pending");
    const approved = erpKeys.purchaseOrders("approved");
    expect(JSON.stringify(pending)).not.toBe(JSON.stringify(approved));
  });
});

describe("erpKeys.invoices", () => {
  it("different params produce different keys", () => {
    const a = erpKeys.invoices({ status: "pending" });
    const b = erpKeys.invoices({ status: "paid" });
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });
});

describe("erpKeys.invoice", () => {
  it("different ids produce different keys", () => {
    const a = erpKeys.invoice("inv-1");
    const b = erpKeys.invoice("inv-2");
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });

  it("includes id in key", () => {
    expect(erpKeys.invoice("inv-42")).toContain("inv-42");
  });
});

describe("erpKeys.accountStatement", () => {
  it("different entityIds produce different keys", () => {
    const a = erpKeys.accountStatement("ent-1");
    const b = erpKeys.accountStatement("ent-2");
    expect(JSON.stringify(a)).not.toBe(JSON.stringify(b));
  });
});
