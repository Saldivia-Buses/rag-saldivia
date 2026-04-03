/**
 * Auth store unit tests.
 *
 * Tests the Zustand store for authentication state management.
 * Verifies initial state, setUser behavior, and clearAuth reset.
 */

import { describe, it, expect, beforeEach } from "bun:test";
import { useAuthStore } from "@/lib/auth/store";

describe("useAuthStore", () => {
  beforeEach(() => {
    // Reset store to initial state between tests
    useAuthStore.setState({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: true,
    });
  });

  it("initial state is unauthenticated", () => {
    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(state.isLoading).toBe(true);
  });

  it("setAccessToken stores the token", () => {
    useAuthStore.getState().setAccessToken("test-token-abc");
    const state = useAuthStore.getState();
    expect(state.accessToken).toBe("test-token-abc");
  });

  it("setting user + isAuthenticated marks as authenticated", () => {
    useAuthStore.setState({
      user: {
        id: "u1",
        email: "enzo@saldivia.com",
        name: "Enzo",
        role: "admin",
        tenantId: "t1",
        tenantSlug: "saldivia",
      },
      isAuthenticated: true,
      isLoading: false,
    });

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.user?.email).toBe("enzo@saldivia.com");
    expect(state.user?.tenantSlug).toBe("saldivia");
  });

  it("clearAuth resets to unauthenticated", () => {
    // First set authenticated state
    useAuthStore.setState({
      user: {
        id: "u1",
        email: "enzo@saldivia.com",
        name: "Enzo",
        role: "admin",
        tenantId: "t1",
        tenantSlug: "saldivia",
      },
      accessToken: "some-token",
      isAuthenticated: true,
      isLoading: false,
    });

    // Now clear
    useAuthStore.getState().clearAuth();

    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.accessToken).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(state.isLoading).toBe(false);
  });
});
