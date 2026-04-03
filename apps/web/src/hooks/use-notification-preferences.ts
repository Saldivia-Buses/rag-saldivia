import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";

export interface NotificationPreferences {
  email_enabled: boolean;
  in_app_enabled: boolean;
  quiet_start: string | null;
  quiet_end: string | null;
  muted_types: string[];
}

export function useNotificationPreferences() {
  return useQuery({
    queryKey: ["notifications", "preferences"],
    queryFn: () =>
      api.get<NotificationPreferences>("/v1/notifications/preferences"),
  });
}

export function useUpdateNotificationPreferences() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<NotificationPreferences>) =>
      api.put("/v1/notifications/preferences", data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["notifications", "preferences"],
      });
    },
  });
}
