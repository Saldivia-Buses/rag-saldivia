"use client";

import { useEffect } from "react";
import { useAuthStore } from "./store";

/**
 * Initializes auth state on app mount.
 * Tries to refresh the session using the HttpOnly cookie.
 * If refresh succeeds, fetches user profile.
 * If refresh fails, user remains unauthenticated.
 */
export function AuthInitializer({ children }: { children: React.ReactNode }) {
  const { refresh, fetchMe, isLoading, accessToken } = useAuthStore();

  useEffect(() => {
    // If we already have a token (e.g., from login), just fetch user info
    if (accessToken) {
      fetchMe();
      return;
    }

    // Try silent refresh via HttpOnly cookie
    refresh().then((ok) => {
      if (ok) fetchMe();
    });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // While checking auth, show nothing (prevents flash of wrong UI)
  if (isLoading) {
    return null;
  }

  return <>{children}</>;
}
