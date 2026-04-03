"use client";

import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useNotificationPreferences,
  useUpdateNotificationPreferences,
} from "@/hooks/use-notification-preferences";

export function NotificationPreferences() {
  const { data: prefs, isLoading } = useNotificationPreferences();
  const update = useUpdateNotificationPreferences();

  if (isLoading) {
    return (
      <div className="space-y-4 p-6">
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-8 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <div>
        <h3 className="text-sm font-semibold">Preferencias</h3>
        <p className="text-xs text-muted-foreground">
          Configura como recibis las notificaciones.
        </p>
      </div>

      <div className="flex items-center justify-between">
        <div>
          <Label>Notificaciones por email</Label>
          <p className="text-xs text-muted-foreground">
            Recibir notificaciones en tu casilla de correo.
          </p>
        </div>
        <Switch
          checked={prefs?.email_enabled ?? true}
          onCheckedChange={(checked: boolean) => {
            update.mutate({ email_enabled: checked });
          }}
          disabled={update.isPending}
        />
      </div>

      <div className="flex items-center justify-between">
        <div>
          <Label>Notificaciones in-app</Label>
          <p className="text-xs text-muted-foreground">
            Mostrar notificaciones dentro de la plataforma.
          </p>
        </div>
        <Switch
          checked={prefs?.in_app_enabled ?? true}
          onCheckedChange={(checked: boolean) => {
            update.mutate({ in_app_enabled: checked });
          }}
          disabled={update.isPending}
        />
      </div>
    </div>
  );
}
