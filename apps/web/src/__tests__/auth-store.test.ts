/**
 * Auth store unit tests.
 *
 * Tests the Zustand store for authentication state management.
 * Verifies initial state, setAccessToken, clearAuth, login, logout,
 * fetchMe, and refresh — with fetch mocked for async actions.
 *
 * The access token lives in memory only (not localStorage).
 * The refresh token lives in an HttpOnly cookie — not visible here.
 */

import { describe, it, expect, beforeEach, afterEach } from "bun:test";
import { useAuthStore } from "@/lib/auth/store";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function mockResponse(status: number, body: unknown): Response {
  return {
    status,
    ok: status >= 200 && status < 300,
    statusText: "",
    json: async () => body,
    headers: new Headers(),
  } as unknown as Response;
}

const sampleUser = {
  id: "u1",
  email: "enzo@saldivia.com",
  name: "Enzo",
  role: "admin",
  tenantId: "t1",
  tenantSlug: "saldivia",
};

const sampleUserRaw = {
  id: "u1",
  email: "enzo@saldivia.com",
  name: "Enzo",
  role: "admin",
  tenant_id: "t1",
  tenant_slug: "saldivia",
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("useAuthStore", () => {
  const originalFetch = globalThis.fetch;
  const originalEnv = process.env.NEXT_PUBLIC_API_URL;

  beforeEach(() => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.test";
    // Reset store to initial state between tests
    useAuthStore.setState({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: true,
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

  // ── Initial state ──────────────────────────────────────────────────────────

  it("initial state is unauthenticated", () => {
    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(state.isLoading).toBe(true);
  });

  // ── setAccessToken ─────────────────────────────────────────────────────────

  it("setAccessToken stores the token", () => {
    useAuthStore.getState().setAccessToken("test-token-abc");
    const state = useAuthStore.getState();
    expect(state.accessToken).toBe("test-token-abc");
  });

  it("setAccessToken overwrites a previous token", () => {
    useAuthStore.getState().setAccessToken("old-token");
    useAuthStore.getState().setAccessToken("new-token");
    expect(useAuthStore.getState().accessToken).toBe("new-token");
  });

  it("setAccessToken does not change isAuthenticated on its own", () => {
    useAuthStore.setState({ isAuthenticated: false });
    useAuthStore.getState().setAccessToken("tok");
    // isAuthenticated is only set true by fetchMe (after successful /me call)
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  // ── clearAuth ─────────────────────────────────────────────────────────────

  it("clearAuth resets to unauthenticated", () => {
    useAuthStore.setState({
      user: sampleUser,
      accessToken: "some-token",
      isAuthenticated: true,
      isLoading: false,
    });

    useAuthStore.getState().clearAuth();

    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(state.isLoading).toBe(false);
  });

  it("setTokens then clearAuth → isAuthenticated becomes false", () => {
    useAuthStore.setState({
      accessToken: "tok",
      isAuthenticated: true,
      user: sampleUser,
    });

    useAuthStore.getState().clearAuth();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().accessToken).toBeNull();
  });

  // ── isAuthenticated ────────────────────────────────────────────────────────

  it("isAuthenticated is true when user and token are both present", () => {
    useAuthStore.setState({
      user: sampleUser,
      accessToken: "valid-token",
      isAuthenticated: true,
      isLoading: false,
    });

    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().user?.email).toBe("enzo@saldivia.com");
  });

  it("isAuthenticated is false when no user", () => {
    useAuthStore.setState({
      user: null,
      accessToken: "some-token",
      isAuthenticated: false,
    });

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it("setting user + isAuthenticated marks as authenticated", () => {
    useAuthStore.setState({
      user: sampleUser,
      isAuthenticated: true,
      isLoading: false,
    });

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.user?.email).toBe("enzo@saldivia.com");
    expect(state.user?.tenantSlug).toBe("saldivia");
  });

  // ── logout ────────────────────────────────────────────────────────────────

  it("logout clears all stored tokens and user", async () => {
    useAuthStore.setState({
      user: sampleUser,
      accessToken: "tok",
      isAuthenticated: true,
      isLoading: false,
    });

    // Mock: logout POST returns 204
    globalThis.fetch = async () =>
      ({
        status: 204,
        ok: true,
        statusText: "",
        json: async () => { throw new Error("no body"); },
        headers: new Headers(),
      } as unknown as Response);

    await useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });

  it("logout clears state even when the API call fails", async () => {
    useAuthStore.setState({
      user: sampleUser,
      accessToken: "tok",
      isAuthenticated: true,
    });

    // Mock: logout POST returns 500
    globalThis.fetch = async () =>
      mockResponse(500, { error: "server error" });

    await useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });

  it("logout clears state even when fetch throws a network error", async () => {
    useAuthStore.setState({
      user: sampleUser,
      accessToken: "tok",
      isAuthenticated: true,
    });

    globalThis.fetch = async () => { throw new Error("network failure"); };

    await useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.accessToken).toBeNull();
  });

  // ── fetchMe ───────────────────────────────────────────────────────────────

  it("fetchMe populates user and sets isAuthenticated on success", async () => {
    useAuthStore.setState({ accessToken: "valid-tok" });

    globalThis.fetch = async () => mockResponse(200, sampleUserRaw);

    await useAuthStore.getState().fetchMe();

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.isLoading).toBe(false);
    expect(state.user?.id).toBe("u1");
    expect(state.user?.tenantId).toBe("t1");
    expect(state.user?.tenantSlug).toBe("saldivia");
  });

  it("fetchMe with 401 response clears auth", async () => {
    useAuthStore.setState({
      accessToken: "expired-tok",
      isAuthenticated: true,
      user: sampleUser,
    });

    // First call: the /me request returns 401
    // Second call: the refresh attempt (triggered by 401 in request()) also returns 401
    globalThis.fetch = async () =>
      mockResponse(401, { error: "unauthorized" });

    await useAuthStore.getState().fetchMe();

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.accessToken).toBeNull();
    expect(state.isLoading).toBe(false);
  });

  it("fetchMe sets isLoading=false even on non-401 errors", async () => {
    useAuthStore.setState({ isLoading: true });

    globalThis.fetch = async () =>
      mockResponse(503, { error: "service unavailable" });

    await useAuthStore.getState().fetchMe();

    expect(useAuthStore.getState().isLoading).toBe(false);
  });

  // ── refresh ───────────────────────────────────────────────────────────────

  it("refresh returns true and stores new access token on success", async () => {
    globalThis.fetch = async () =>
      mockResponse(200, { access_token: "new-access-tok", expires_in: 900 });

    const ok = await useAuthStore.getState().refresh();

    expect(ok).toBe(true);
    expect(useAuthStore.getState().accessToken).toBe("new-access-tok");
  });

  it("refresh returns false and clears auth on failure", async () => {
    useAuthStore.setState({ accessToken: "old", isAuthenticated: true });

    globalThis.fetch = async () =>
      mockResponse(401, { error: "refresh token expired" });

    const ok = await useAuthStore.getState().refresh();

    expect(ok).toBe(false);
    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it("refresh returns false and clears auth on network error", async () => {
    useAuthStore.setState({ accessToken: "old", isAuthenticated: true });

    globalThis.fetch = async () => { throw new Error("network failure"); };

    const ok = await useAuthStore.getState().refresh();

    expect(ok).toBe(false);
    expect(useAuthStore.getState().accessToken).toBeNull();
  });
});
