/**
 * Auth store — Zustand store for authentication state.
 *
 * Access token lives in memory only (not localStorage) to prevent XSS theft.
 * Refresh token lives in an HttpOnly cookie managed by the backend.
 */

import { create } from "zustand";
import { api, ApiError } from "@/lib/api/client";

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  role: string;
  tenantId: string;
  tenantSlug: string;
  perms: string[];
}

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  setAccessToken: (token: string) => void;
  clearAuth: () => void;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchMe: () => Promise<void>;
  refresh: () => Promise<boolean>;
}

function decodeJwtPerms(token: string): string[] {
  try {
    const payload = token.split(".")[1];
    if (!payload) return [];
    const padded = payload + "=".repeat((4 - (payload.length % 4)) % 4);
    const json = atob(padded.replace(/-/g, "+").replace(/_/g, "/"));
    const claims = JSON.parse(json) as { perms?: unknown };
    return Array.isArray(claims.perms)
      ? claims.perms.filter((p): p is string => typeof p === "string")
      : [];
  } catch {
    return [];
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  isAuthenticated: false,
  isLoading: true,

  setAccessToken: (token: string) => {
    set({ accessToken: token });
  },

  clearAuth: () => {
    set({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
  },

  login: async (email: string, password: string) => {
    const data = await api.post<{
      access_token: string;
      refresh_token: string;
      expires_in: number;
    }>("/v1/auth/login", { email, password }, { skipAuth: true });

    set({ accessToken: data.access_token });
    await get().fetchMe();
  },

  logout: async () => {
    try {
      await api.post("/v1/auth/logout");
    } catch {
      // Best-effort — clear local state regardless
    }
    get().clearAuth();
  },

  fetchMe: async () => {
    try {
      const user = await api.get<{
        id: string;
        email: string;
        name: string;
        role: string;
        tenant_id: string;
        tenant_slug: string;
      }>("/v1/auth/me");

      const token = get().accessToken;
      const perms = token ? decodeJwtPerms(token) : [];

      set({
        user: {
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role,
          tenantId: user.tenant_id,
          tenantSlug: user.tenant_slug,
          perms,
        },
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        get().clearAuth();
      }
      set({ isLoading: false });
    }
  },

  refresh: async () => {
    try {
      const data = await api.post<{ access_token: string; refresh_token: string; expires_in: number }>(
        "/v1/auth/refresh",
        undefined,
        { skipAuth: true },
      );
      set({ accessToken: data.access_token });
      const currentUser = get().user;
      if (currentUser) {
        set({ user: { ...currentUser, perms: decodeJwtPerms(data.access_token) } });
      }
      return true;
    } catch {
      get().clearAuth();
      return false;
    }
  },
}));
