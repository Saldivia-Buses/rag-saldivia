"use client";

import { redirect } from "next/navigation";
import { useEnabledModules } from "./hooks";
import { MODULE_REGISTRY } from "./registry";

/**
 * Route guard for module-gated pages.
 * If the tenant doesn't have the module enabled, silently redirects to /chat.
 * While loading or if the API is unavailable, allows access (fail-open).
 */
export function ModuleGuard({
  moduleId,
  children,
}: {
  moduleId: string;
  children: React.ReactNode;
}) {
  const { data: modules, isSuccess } = useEnabledModules();

  // Only enforce guard when API successfully returned a non-empty module list.
  // Otherwise fail-open (dev mode, backend offline, or no modules configured yet).
  if (isSuccess && modules && modules.length > 0) {
    const allowed = modules.some((m) => m.id === moduleId);
    if (!allowed) {
      redirect("/chat");
    }
  }

  return <>{children}</>;
}
