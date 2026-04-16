/**
 * Permission messages unit tests.
 *
 * Tests lib/erp/permission-messages.ts:
 * - permissionMessages: lookup record for known permission codes
 * - permissionErrorToast: routes 403 ApiErrors to toast.error,
 *   other errors to a generic fallback
 *
 * sonner's toast is mocked so no DOM is needed.
 * The module under test is loaded via dynamic import AFTER mock.module()
 * is registered — this ensures the top-level `toast` binding in
 * permission-messages.ts resolves to the mock, not the real sonner.
 */

import { describe, it, expect, mock, beforeEach } from "bun:test";
import { ApiError } from "@/lib/api/client";

// ---------------------------------------------------------------------------
// Register sonner mock BEFORE any import of the module under test.
// The mock intercepts toast.error calls and records their arguments.
// ---------------------------------------------------------------------------

const toastErrorCalls: Array<{ msg: string; opts: unknown }> = [];

mock.module("sonner", () => ({
  toast: Object.assign((_msg: string) => {}, {
    error: (msg: string, opts?: unknown) => {
      toastErrorCalls.push({ msg, opts });
    },
    success: () => {},
    info: () => {},
    warning: () => {},
  }),
}));

// Load the module AFTER mock registration so its static `toast` import
// is resolved against the mock.
const { permissionMessages, permissionErrorToast } = await import(
  "@/lib/erp/permission-messages"
);

// ---------------------------------------------------------------------------
// permissionMessages record (pure data)
// ---------------------------------------------------------------------------

describe("permissionMessages record", () => {
  it("contains all expected ERP permission codes", () => {
    const expected = [
      "erp.accounting.close",
      "erp.treasury.reconcile",
      "erp.invoicing.void",
      "erp.purchasing.inspect",
      "erp.treasury.receipt",
      "erp.stock.write",
      "erp.entities.write",
      "erp.catalogs.write",
    ];
    for (const code of expected) {
      expect(permissionMessages[code]).toBeDefined();
    }
  });

  it("all messages are non-empty strings", () => {
    for (const [key, msg] of Object.entries(permissionMessages)) {
      expect(typeof msg).toBe("string");
      expect((msg as string).length).toBeGreaterThan(0);
    }
  });

  it("unknown permission code returns undefined", () => {
    expect(permissionMessages["erp.unknown.action"]).toBeUndefined();
    expect(permissionMessages["admin.superpower"]).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// permissionErrorToast
// ---------------------------------------------------------------------------

describe("permissionErrorToast", () => {
  beforeEach(() => {
    toastErrorCalls.length = 0;
  });

  it("shows permission-denied message for a 403 ApiError", () => {
    permissionErrorToast(new ApiError(403, "Forbidden"));
    expect(toastErrorCalls).toHaveLength(1);
    expect(toastErrorCalls[0].msg).toBe("No tenés permiso para esta acción");
  });

  it("shows generic message for a 500 ApiError", () => {
    permissionErrorToast(new ApiError(500, "Internal Server Error"));
    expect(toastErrorCalls).toHaveLength(1);
    expect(toastErrorCalls[0].msg).toBe("Error inesperado");
  });

  it("includes error message as description for non-403 ApiError", () => {
    permissionErrorToast(new ApiError(422, "Validation failed"));
    const opts = toastErrorCalls[0].opts as { description?: string } | undefined;
    expect(opts?.description).toBe("Validation failed");
  });

  it("shows generic message with description for a plain Error", () => {
    permissionErrorToast(new Error("Something broke"));
    expect(toastErrorCalls[0].msg).toBe("Error inesperado");
    const opts = toastErrorCalls[0].opts as { description?: string } | undefined;
    expect(opts?.description).toBe("Something broke");
  });

  it("handles a thrown string without crashing", () => {
    expect(() => permissionErrorToast("string error")).not.toThrow();
    expect(toastErrorCalls).toHaveLength(1);
    expect(toastErrorCalls[0].msg).toBe("Error inesperado");
  });

  it("non-Error object gets undefined description", () => {
    permissionErrorToast({ code: 42 });
    const opts = toastErrorCalls[0].opts as { description?: string } | undefined;
    expect(opts?.description).toBeUndefined();
  });

  it("handles null without throwing", () => {
    expect(() => permissionErrorToast(null)).not.toThrow();
    expect(toastErrorCalls).toHaveLength(1);
    expect(toastErrorCalls[0].msg).toBe("Error inesperado");
  });

  it("handles undefined without throwing", () => {
    expect(() => permissionErrorToast(undefined)).not.toThrow();
    expect(toastErrorCalls).toHaveLength(1);
  });

  it("401 is treated as unexpected, not a permission error", () => {
    permissionErrorToast(new ApiError(401, "Unauthorized"));
    expect(toastErrorCalls[0].msg).toBe("Error inesperado");
  });

  it("calls toast.error exactly once per invocation", () => {
    permissionErrorToast(new ApiError(403, "Forbidden"));
    expect(toastErrorCalls).toHaveLength(1);
  });
});
