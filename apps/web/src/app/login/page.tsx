"use client";

import { Login5 } from "@/components/login5";
import { useAuthStore } from "@/lib/auth/store";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function LoginPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      router.push("/inicio");
    }
  }, [isAuthenticated, router]);

  return <Login5 />;
}
