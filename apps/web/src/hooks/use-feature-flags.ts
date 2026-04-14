import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";

interface FeatureFlagsResponse {
  flags: Record<string, boolean>;
}

export function useFeatureFlags() {
  return useQuery({
    queryKey: ["feature-flags"],
    queryFn: () => api.get<FeatureFlagsResponse>("/v1/flags/evaluate"),
    staleTime: 30_000,
  });
}

export function useFeatureFlag(flag: string): boolean {
  const { data } = useFeatureFlags();
  return data?.flags[flag] ?? false;
}
