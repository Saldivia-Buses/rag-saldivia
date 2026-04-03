/**
 * API client unit tests.
 *
 * Tests pure functions from the API client module:
 * - getTenantSlug: extracts tenant from hostname
 * - getApiBaseUrl: resolves API base URL
 *
 * These tests run in a non-browser environment (bun), so window is undefined
 * and we test the SSR/server-side code paths.
 */

import { describe, it, expect, beforeEach, afterEach } from "bun:test";
import { getTenantSlug, getApiBaseUrl } from "@/lib/api/client";

describe("getTenantSlug", () => {
  const originalEnv = process.env.NEXT_PUBLIC_TENANT_SLUG;

  afterEach(() => {
    // Restore env
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_TENANT_SLUG = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_TENANT_SLUG;
    }
  });

  it("returns env variable when window is undefined (SSR)", () => {
    process.env.NEXT_PUBLIC_TENANT_SLUG = "saldivia";
    // In bun test, window is undefined — this exercises the SSR path
    const slug = getTenantSlug();
    expect(slug).toBe("saldivia");
  });

  it("defaults to 'dev' when env variable is not set (SSR)", () => {
    delete process.env.NEXT_PUBLIC_TENANT_SLUG;
    const slug = getTenantSlug();
    expect(slug).toBe("dev");
  });
});

describe("getApiBaseUrl", () => {
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  afterEach(() => {
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_API_URL = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_API_URL;
    }
  });

  it("returns env variable when set", () => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.sda.app";
    const url = getApiBaseUrl();
    expect(url).toBe("https://api.sda.app");
  });

  it("returns empty string when env is not set", () => {
    delete process.env.NEXT_PUBLIC_API_URL;
    const url = getApiBaseUrl();
    expect(url).toBe("");
  });
});

describe("ApiError", () => {
  it("is constructable with status and message", async () => {
    const { ApiError } = await import("@/lib/api/client");
    const err = new ApiError(404, "not found");
    expect(err.status).toBe(404);
    expect(err.message).toBe("not found");
    expect(err.name).toBe("ApiError");
    expect(err instanceof Error).toBe(true);
  });
});
