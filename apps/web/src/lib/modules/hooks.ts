"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";

export interface EnabledModule {
  id: string;
  name: string;
  category: string;
}

/**
 * Fetches enabled modules for the current tenant.
 * staleTime: Infinity — only invalidated by WebSocket events from the Hub.
 */
export function useEnabledModules() {
  return useQuery({
    queryKey: ["modules", "enabled"],
    queryFn: () => api.get<EnabledModule[]>("/v1/modules/enabled"),
    staleTime: Infinity,
  });
}

/**
 * Returns true if the current tenant has the given module enabled.
 */
export function useHasModule(moduleId: string): boolean {
  const { data: modules } = useEnabledModules();
  if (!modules) return false;
  return modules.some((m) => m.id === moduleId);
}
