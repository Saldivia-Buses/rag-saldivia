/**
 * API client unit tests.
 *
 * Tests pure functions from the API client module:
 * - getTenantSlug: extracts tenant from hostname
 * - getApiBaseUrl: resolves API base URL
 * - ApiError: error class shape
 * - api.get / api.post / api.delete: fetch wrapper behavior via fetch mock
 *
 * These tests run in a non-browser environment (bun), so window is undefined
 * and we test the SSR/server-side code paths. Browser paths (hostname parsing)
 * are documented but cannot be exercised without a real DOM.
 */

import { describe, it, expect, beforeEach, afterEach } from "bun:test";
import { getTenantSlug, getApiBaseUrl, ApiError, api } from "@/lib/api/client";
import { useAuthStore } from "@/lib/auth/store";

// ---------------------------------------------------------------------------
// getTenantSlug
// ---------------------------------------------------------------------------

describe("getTenantSlug", () => {
  const originalEnv = process.env.NEXT_PUBLIC_TENANT_SLUG;

  afterEach(() => {
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_TENANT_SLUG = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_TENANT_SLUG;
    }
  });

  it("returns env variable when window is undefined (SSR)", () => {
    process.env.NEXT_PUBLIC_TENANT_SLUG = "saldivia";
    const slug = getTenantSlug();
    expect(slug).toBe("saldivia");
  });

  it("defaults to 'dev' when env variable is not set (SSR)", () => {
    delete process.env.NEXT_PUBLIC_TENANT_SLUG;
    const slug = getTenantSlug();
    expect(slug).toBe("dev");
  });

  it("returns env variable even when set to an unusual value (SSR)", () => {
    process.env.NEXT_PUBLIC_TENANT_SLUG = "my-tenant-123";
    const slug = getTenantSlug();
    expect(slug).toBe("my-tenant-123");
  });
});

// ---------------------------------------------------------------------------
// getApiBaseUrl
// ---------------------------------------------------------------------------

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

  it("returns URL without trailing slash as-is", () => {
    process.env.NEXT_PUBLIC_API_URL = "http://localhost:8080";
    expect(getApiBaseUrl()).toBe("http://localhost:8080");
  });
});

// ---------------------------------------------------------------------------
// ApiError
// ---------------------------------------------------------------------------

describe("ApiError", () => {
  it("is constructable with status and message", () => {
    const err = new ApiError(404, "not found");
    expect(err.status).toBe(404);
    expect(err.message).toBe("not found");
    expect(err.name).toBe("ApiError");
    expect(err instanceof Error).toBe(true);
    expect(err instanceof ApiError).toBe(true);
  });

  it("status 401 is stored correctly", () => {
    const err = new ApiError(401, "unauthorized");
    expect(err.status).toBe(401);
    expect(err.message).toBe("unauthorized");
  });

  it("status 403 is stored correctly", () => {
    const err = new ApiError(403, "forbidden");
    expect(err.status).toBe(403);
    expect(err.message).toBe("forbidden");
  });

  it("status 500 is stored correctly", () => {
    const err = new ApiError(500, "internal server error");
    expect(err.status).toBe(500);
    expect(err.message).toBe("internal server error");
  });
});

// ---------------------------------------------------------------------------
// api.* — fetch wrapper (fetch is monkey-patched per test)
// ---------------------------------------------------------------------------

/**
 * Build a minimal Response-like object that satisfies our fetch wrapper.
 * status: HTTP status code
 * body: object to return from res.json()
 * ok: derived from status < 400 unless overridden
 */
function mockResponse(
  status: number,
  body: unknown,
  options: { ok?: boolean; statusText?: string } = {},
): Response {
  const ok = options.ok ?? (status >= 200 && status < 300);
  return {
    status,
    ok,
    statusText: options.statusText ?? "",
    json: async () => body,
    headers: new Headers(),
  } as unknown as Response;
}

describe("api.get — fetch wrapper", () => {
  const originalFetch = globalThis.fetch;
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.test";
    // Reset auth store — no token by default
    useAuthStore.setState({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_API_URL = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_API_URL;
    }
  });

  it("constructs URL by concatenating base URL and path", async () => {
    let capturedUrl = "";
    globalThis.fetch = async (input: RequestInfo | URL) => {
      capturedUrl = input.toString();
      return mockResponse(200, { ok: true });
    };

    await api.get("/v1/users");
    expect(capturedUrl).toBe("https://api.test/v1/users");
  });

  it("constructs URL with multiple path segments", async () => {
    let capturedUrl = "";
    globalThis.fetch = async (input: RequestInfo | URL) => {
      capturedUrl = input.toString();
      return mockResponse(200, {});
    };

    await api.get("/v1/users/u1/sessions");
    expect(capturedUrl).toBe("https://api.test/v1/users/u1/sessions");
  });

  it("attaches Authorization header when token is present", async () => {
    useAuthStore.setState({ accessToken: "my-token" });

    let capturedHeaders: Record<string, string> = {};
    globalThis.fetch = async (_input: RequestInfo | URL, init?: RequestInit) => {
      capturedHeaders = Object.fromEntries(
        new Headers(init?.headers).entries(),
      );
      return mockResponse(200, { data: true });
    };

    await api.get("/v1/protected");
    expect(capturedHeaders["authorization"]).toBe("Bearer my-token");
  });

  it("does not attach Authorization header when no token", async () => {
    useAuthStore.setState({ accessToken: null });

    let capturedHeaders: Record<string, string> = {};
    globalThis.fetch = async (_input: RequestInfo | URL, init?: RequestInit) => {
      capturedHeaders = Object.fromEntries(
        new Headers(init?.headers).entries(),
      );
      return mockResponse(200, {});
    };

    await api.get("/v1/public");
    expect(capturedHeaders["authorization"]).toBeUndefined();
  });

  it("throws ApiError with status 403 on forbidden", async () => {
    globalThis.fetch = async () =>
      mockResponse(403, { error: "forbidden" }, { ok: false });

    let thrown: unknown;
    try {
      await api.get("/v1/admin");
    } catch (e) {
      thrown = e;
    }

    expect(thrown instanceof ApiError).toBe(true);
    expect((thrown as ApiError).status).toBe(403);
    expect((thrown as ApiError).message).toBe("forbidden");
  });

  it("throws ApiError with status 500 on server error (after retry)", async () => {
    // Returns 500 both times (initial + single retry)
    let callCount = 0;
    globalThis.fetch = async () => {
      callCount++;
      return mockResponse(500, { error: "internal error" }, { ok: false });
    };

    let thrown: unknown;
    try {
      await api.get("/v1/broken");
    } catch (e) {
      thrown = e;
    }

    expect(thrown instanceof ApiError).toBe(true);
    expect((thrown as ApiError).status).toBe(500);
    // Should have retried once (2 fetch calls total)
    expect(callCount).toBe(2);
  });

  it("returns data on 200 without throwing", async () => {
    globalThis.fetch = async () =>
      mockResponse(200, { id: "u1", name: "Enzo" });

    const result = await api.get<{ id: string; name: string }>("/v1/users/u1");
    expect(result.id).toBe("u1");
    expect(result.name).toBe("Enzo");
  });

  it("returns undefined on 204 No Content", async () => {
    globalThis.fetch = async () => ({
      status: 204,
      ok: true,
      statusText: "",
      json: async () => { throw new Error("no body"); },
      headers: new Headers(),
    } as unknown as Response);

    const result = await api.get<undefined>("/v1/sessions/s1");
    expect(result).toBeUndefined();
  });

  it("allows 206 Partial Content without throwing", async () => {
    globalThis.fetch = async () => ({
      status: 206,
      ok: false, // 206 triggers !res.ok but the client exempts it
      statusText: "Partial Content",
      json: async () => ({ partial: true }),
      headers: new Headers(),
    } as unknown as Response);

    const result = await api.get<{ partial: boolean }>("/v1/dashboard/kpis");
    expect(result).toEqual({ partial: true });
  });

  it("handles non-JSON error body gracefully (falls back to statusText)", async () => {
    globalThis.fetch = async () => ({
      status: 400,
      ok: false,
      statusText: "Bad Request",
      json: async () => { throw new SyntaxError("not json"); },
      headers: new Headers(),
    } as unknown as Response);

    let thrown: unknown;
    try {
      await api.get("/v1/malformed");
    } catch (e) {
      thrown = e;
    }

    expect(thrown instanceof ApiError).toBe(true);
    expect((thrown as ApiError).status).toBe(400);
    // Falls back to statusText when json() throws
    expect((thrown as ApiError).message).toBe("Bad Request");
  });

  it("throws ApiError on 401 and clears auth when refresh also fails", async () => {
    useAuthStore.setState({ accessToken: "expired-token", isAuthenticated: true });

    // All fetches return 401 (initial request + refresh attempt in tryRefresh)
    globalThis.fetch = async () => mockResponse(401, { error: "unauthorized" }, { ok: false });

    let thrown: unknown;
    try {
      await api.get("/v1/protected");
    } catch (e) {
      thrown = e;
    }

    expect(thrown instanceof ApiError).toBe(true);
    expect((thrown as ApiError).status).toBe(401);
    expect((thrown as ApiError).message).toBe("session expired");

    // Auth must be cleared after failed refresh
    const state = useAuthStore.getState();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });
});

describe("api.post — fetch wrapper", () => {
  const originalFetch = globalThis.fetch;
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.test";
    useAuthStore.setState({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_API_URL = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_API_URL;
    }
  });

  it("sets Content-Type: application/json when body is an object", async () => {
    let capturedHeaders: Record<string, string> = {};
    globalThis.fetch = async (_input: RequestInfo | URL, init?: RequestInit) => {
      capturedHeaders = Object.fromEntries(new Headers(init?.headers).entries());
      return mockResponse(200, { created: true });
    };

    await api.post("/v1/sessions", { title: "New session" });
    expect(capturedHeaders["content-type"]).toBe("application/json");
  });

  it("serializes body to JSON string", async () => {
    let capturedBody = "";
    globalThis.fetch = async (_input: RequestInfo | URL, init?: RequestInit) => {
      capturedBody = init?.body as string;
      return mockResponse(200, {});
    };

    await api.post("/v1/sessions", { title: "My Chat" });
    expect(capturedBody).toBe(JSON.stringify({ title: "My Chat" }));
  });

  it("sends no body and no Content-Type when body is undefined", async () => {
    let capturedHeaders: Record<string, string> = {};
    let capturedBody: BodyInit | null | undefined;
    globalThis.fetch = async (_input: RequestInfo | URL, init?: RequestInit) => {
      capturedHeaders = Object.fromEntries(new Headers(init?.headers).entries());
      capturedBody = init?.body;
      return mockResponse(200, {});
    };

    await api.post("/v1/auth/logout");
    expect(capturedHeaders["content-type"]).toBeUndefined();
    expect(capturedBody).toBeUndefined();
  });

  it("does not attach auth header when skipAuth: true", async () => {
    useAuthStore.setState({ accessToken: "should-not-appear" });

    let capturedHeaders: Record<string, string> = {};
    globalThis.fetch = async (_input: RequestInfo | URL, init?: RequestInit) => {
      capturedHeaders = Object.fromEntries(new Headers(init?.headers).entries());
      return mockResponse(200, { access_token: "new-tok" });
    };

    await api.post("/v1/auth/refresh", undefined, { skipAuth: true });
    expect(capturedHeaders["authorization"]).toBeUndefined();
  });

});

describe("api.delete — fetch wrapper", () => {
  const originalFetch = globalThis.fetch;
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.test";
    useAuthStore.setState({ accessToken: null, isAuthenticated: false });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_API_URL = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_API_URL;
    }
  });

  it("returns undefined on 204 for delete", async () => {
    globalThis.fetch = async () => ({
      status: 204,
      ok: true,
      statusText: "",
      json: async () => { throw new Error("no body"); },
      headers: new Headers(),
    } as unknown as Response);

    const result = await api.delete("/v1/sessions/s1");
    expect(result).toBeUndefined();
  });
});

describe("concurrent requests", () => {
  const originalFetch = globalThis.fetch;
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.test";
    useAuthStore.setState({ accessToken: "tok", isAuthenticated: true });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    if (originalEnv !== undefined) {
      process.env.NEXT_PUBLIC_API_URL = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_API_URL;
    }
  });

  it("multiple simultaneous requests all succeed independently", async () => {
    const calls: string[] = [];
    globalThis.fetch = async (input: RequestInfo | URL) => {
      const url = input.toString();
      calls.push(url);
      return mockResponse(200, { url });
    };

    const [a, b, c] = await Promise.all([
      api.get<{ url: string }>("/v1/a"),
      api.get<{ url: string }>("/v1/b"),
      api.get<{ url: string }>("/v1/c"),
    ]);

    expect(a.url).toBe("https://api.test/v1/a");
    expect(b.url).toBe("https://api.test/v1/b");
    expect(c.url).toBe("https://api.test/v1/c");
    expect(calls).toHaveLength(3);
  });
});
