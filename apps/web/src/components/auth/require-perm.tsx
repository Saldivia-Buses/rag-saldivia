"use client";

import { useHasPermission } from "@/lib/auth/use-has-permission";

interface RequirePermProps {
  perm: string;
  mode?: "hide" | "disable";
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export function RequirePerm({
  perm,
  mode = "hide",
  children,
  fallback = null,
}: RequirePermProps) {
  const allowed = useHasPermission(perm);

  if (allowed) return <>{children}</>;
  if (mode === "disable") {
    return (
      <div
        aria-disabled="true"
        title={`Requiere permiso: ${perm}`}
        className="pointer-events-none opacity-50"
      >
        {children}
      </div>
    );
  }
  return <>{fallback}</>;
}
