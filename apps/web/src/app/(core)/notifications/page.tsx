import { Bell } from "lucide-react";

export default function NotificationsPage() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4">
      <div className="flex size-16 items-center justify-center rounded-full bg-muted">
        <Bell className="size-8 text-muted-foreground" />
      </div>
      <div className="text-center">
        <h2 className="text-lg font-semibold">Sin notificaciones</h2>
        <p className="text-sm text-muted-foreground mt-1">
          Cuando haya novedades, van a aparecer acá.
        </p>
      </div>
    </div>
  );
}
