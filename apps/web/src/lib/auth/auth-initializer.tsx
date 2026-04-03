"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { useAuthStore } from "./store";

// Routes that don't require authentication
const publicRoutes = ["/login", "/forgot-password"];

/**
 * Initializes auth state on app mount.
 * Tries to refresh the session using the HttpOnly cookie.
 * If refresh succeeds, fetches user profile.
 * If refresh fails, redirects to login (unless already on a public route).
 */
export function AuthInitializer({ children }: { children: React.ReactNode }) {
  const { refresh, fetchMe, isLoading, isAuthenticated, accessToken } =
    useAuthStore();
  const pathname = usePathname();
  const isPublicRoute = publicRoutes.some((r) => pathname.startsWith(r));

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

  // Public routes render immediately (login page, etc.)
  if (isPublicRoute) {
    return <>{children}</>;
  }

  // While checking auth, show nothing (prevents flash of wrong UI)
  if (isLoading) {
    return null;
  }

  // Not authenticated — redirect to login
  if (!isAuthenticated) {
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
    return null;
  }

  return <>{children}</>;
}
