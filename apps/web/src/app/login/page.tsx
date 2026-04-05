"use client";

import { Login5 } from "@/components/login5";
import { useAuthStore } from "@/lib/auth/store";
import { useEffect } from "react";

export default function LoginPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  // If already authenticated, go to dashboard
  useEffect(() => {
    if (isAuthenticated) {
      window.location.href = "/inicio";
    }
  }, [isAuthenticated]);

  return <Login5 />;
}
