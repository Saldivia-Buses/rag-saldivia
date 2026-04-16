"use client";

/**
 * Route guard for module-gated pages.
 *
 * Currently a no-op: every authenticated user can reach every module page.
 * Per-tenant gating will return when there is a real product reason to hide
 * pages from a tenant, paired with an admin UI to toggle modules. Until then,
 * Inicio lists every module from the static registry and the guard mirrors
 * that — anything in the registry is reachable.
 *
 * Reactivation path: re-introduce useEnabledModules() and redirect on miss.
 */
export function ModuleGuard({
  children,
}: {
  moduleId: string;
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
