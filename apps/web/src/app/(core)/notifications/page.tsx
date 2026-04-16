"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Bell,
  CheckCheck,
  MessageSquare,
  ShieldCheck,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface Notification {
  id: string;
  type: string;
  title: string;
  body: string;
  channel: string;
  read_at: string | null;
  created_at: string;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "Justo ahora";
  if (diffMin < 60) return `Hace ${diffMin} min`;
  const diffH = Math.floor(diffMin / 60);
  if (diffH < 24) return `Hace ${diffH}h`;
  const diffD = Math.floor(diffH / 24);
  if (diffD < 7) return `Hace ${diffD}d`;
  return date.toLocaleDateString("es-AR", { day: "numeric", month: "short" });
}

function groupByDate(notifications: Notification[]): { label: string; items: Notification[] }[] {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today.getTime() - 86400000);
  const weekAgo = new Date(today.getTime() - 7 * 86400000);

  const groups: Record<string, Notification[]> = {
    Hoy: [],
    Ayer: [],
    "Esta semana": [],
    Anteriores: [],
  };

  for (const n of notifications) {
    const date = new Date(n.created_at);
    if (date >= today) groups["Hoy"].push(n);
    else if (date >= yesterday) groups["Ayer"].push(n);
    else if (date >= weekAgo) groups["Esta semana"].push(n);
    else groups["Anteriores"].push(n);
  }

  return Object.entries(groups)
    .filter(([, items]) => items.length > 0)
    .map(([label, items]) => ({ label, items }));
}

function getNotificationIcon(type: string) {
  switch (type) {
    case "login":
    case "security":
      return ShieldCheck;
    case "chat":
    case "message":
      return MessageSquare;
    default:
      return Bell;
  }
}

export default function NotificationsPage() {
  const queryClient = useQueryClient();

  const { data: notifications = [], isLoading } = useQuery({
    queryKey: ["notifications"],
    queryFn: () => api.get<Notification[]>("/v1/notifications?limit=50"),
  });

  const markReadMutation = useMutation({
    mutationFn: (id: string) =>
      api.patch(`/v1/notifications/${id}/read`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: () => api.post("/v1/notifications/read-all"),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const unreadCount = notifications.filter((n) => !n.read_at).length;

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-sm text-muted-foreground">Cargando...</p>
      </div>
    );
  }

  if (notifications.length === 0) {
    return (
      <div className="flex flex-1 flex-col">
        <div className="flex items-center justify-between px-6 py-4">
          <h1 className="text-sm font-semibold">Notificaciones</h1>
        </div>
        <div className="flex flex-1 flex-col items-center justify-center gap-4">
          <div className="flex size-12 items-center justify-center rounded-2xl bg-card border border-border/40">
            <Bell className="size-6 text-muted-foreground" />
          </div>
          <div className="text-center">
            <h2 className="text-base font-medium">Sin notificaciones</h2>
            <p className="text-sm text-muted-foreground mt-1">
              Cuando haya novedades, van a aparecer acá.
            </p>
          </div>
        </div>
      </div>
    );
  }

  const grouped = groupByDate(notifications);

  return (
    <div className="flex flex-1 flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-4">
        <div className="flex items-center gap-3">
          <h1 className="text-sm font-semibold">Notificaciones</h1>
          {unreadCount > 0 && (
            <span className="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded-full">
              {unreadCount} sin leer
            </span>
          )}
        </div>
        {unreadCount > 0 && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => markAllReadMutation.mutate()}
            disabled={markAllReadMutation.isPending}
            className="gap-1.5 text-muted-foreground"
          >
            <CheckCheck className="size-3.5" />
            Marcar todas como leídas
          </Button>
        )}
      </div>

      {/* Grouped list */}
      <ScrollArea className="flex-1">
        <div className="flex flex-col pb-4">
          {grouped.map((group) => (
            <div key={group.label}>
              <p className="px-6 pt-4 pb-2 text-[11px] text-muted-foreground uppercase tracking-wider font-medium">
                {group.label}
              </p>
              {group.items.map((notification) => {
                const Icon = getNotificationIcon(notification.type);
                const isUnread = !notification.read_at;

                return (
                  <div
                    key={notification.id}
                    className={cn(
                      "flex items-start gap-3 px-6 py-3 transition-colors cursor-pointer hover:bg-accent/30",
                      isUnread && "hover:bg-accent/40",
                    )}
                    onClick={() => {
                      if (isUnread) markReadMutation.mutate(notification.id);
                    }}
                  >
                    {/* Icon */}
                    <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-card border border-border/40 mt-0.5">
                      <Icon className="size-4 text-muted-foreground" />
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        {/* Unread dot */}
                        {isUnread && (
                          <span className="size-2 shrink-0 rounded-full bg-primary" />
                        )}
                        <p className={cn("text-sm truncate", isUnread && "font-medium")}>
                          {notification.title}
                        </p>
                      </div>
                      {notification.body && (
                        <p className="text-sm text-muted-foreground mt-0.5 truncate">
                          {notification.body}
                        </p>
                      )}
                      <p className="text-[11px] text-muted-foreground/70 mt-1">
                        {formatDate(notification.created_at)}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
