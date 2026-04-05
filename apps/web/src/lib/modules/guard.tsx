"use client";

import { redirect } from "next/navigation";
import { useHasModule } from "./hooks";

/**
 * Route guard for module-gated pages.
 * If the tenant doesn't have the module enabled, silently redirects to /chat.
 */
export function ModuleGuard({
  moduleId,
  children,
}: {
  moduleId: string;
  children: React.ReactNode;
}) {
  const hasModule = useHasModule(moduleId);

  if (!hasModule) {
    redirect("/chat");
  }

  return <>{children}</>;
}
