import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { useAuthStore } from "@/lib/auth/store";

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  const fetchMe = useAuthStore((s) => s.fetchMe);
  return useMutation({
    mutationFn: (data: { name: string }) => api.patch("/v1/auth/me", data),
    onSuccess: async () => {
      await fetchMe();
      queryClient.invalidateQueries({ queryKey: ["user"] });
    },
  });
}
