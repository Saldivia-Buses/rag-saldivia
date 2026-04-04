"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { useAuthStore } from "@/lib/auth/store";
import { useEnabledModules } from "@/lib/modules/hooks";
import {
  MessageSquare,
  Bell,
  Database,
  Upload,
  ArrowRight,
} from "lucide-react";
import Link from "next/link";

interface ApiSession {
  id: string;
  title: string;
  created_at: string;
}

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);

  const { data: sessions = [] } = useQuery({
    queryKey: ["chat", "sessions"],
    queryFn: () => api.get<ApiSession[]>("/v1/chat/sessions"),
  });

  const { data: notifCount } = useQuery({
    queryKey: ["notifications", "count"],
    queryFn: () => api.get<{ count: number }>("/v1/notifications/count"),
  });

  const { data: modules = [] } = useEnabledModules();

  const recentSessions = sessions.slice(0, 5);
  const unread = notifCount?.count ?? 0;

  return (
    <div className="flex flex-1 flex-col p-6 gap-6 max-w-4xl mx-auto w-full">
      {/* Greeting */}
      <div>
        <h1 className="text-2xl font-semibold">
          Hola{user?.name ? `, ${user.name}` : ""}
        </h1>
        <p className="text-muted-foreground mt-1">Que necesitas hoy?</p>
      </div>

      {/* Quick stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Link
          href="/chat"
          className="flex flex-col gap-2 rounded-lg bg-card p-4 hover:bg-muted/50 transition-colors"
        >
          <MessageSquare className="size-5 text-muted-foreground" />
          <div>
            <p className="text-2xl font-semibold">{sessions.length}</p>
            <p className="text-xs text-muted-foreground">Conversaciones</p>
          </div>
        </Link>

        <Link
          href="/notifications"
          className="flex flex-col gap-2 rounded-lg bg-card p-4 hover:bg-muted/50 transition-colors"
        >
          <Bell className="size-5 text-muted-foreground" />
          <div>
            <p className="text-2xl font-semibold">{unread}</p>
            <p className="text-xs text-muted-foreground">Sin leer</p>
          </div>
        </Link>

        <div className="flex flex-col gap-2 rounded-lg bg-card p-4">
          <Database className="size-5 text-muted-foreground" />
          <div>
            <p className="text-2xl font-semibold">{modules.length}</p>
            <p className="text-xs text-muted-foreground">Modulos activos</p>
          </div>
        </div>

        <Link
          href="/chat"
          className="flex flex-col gap-2 rounded-lg bg-card p-4 hover:bg-muted/50 transition-colors items-center justify-center"
        >
          <Upload className="size-5 text-muted-foreground" />
          <p className="text-xs text-muted-foreground text-center">
            Nuevo chat
          </p>
        </Link>
      </div>

      {/* Recent sessions */}
      {recentSessions.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="font-semibold">Conversaciones recientes</h2>
            <Link
              href="/chat"
              className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"
            >
              Ver todas <ArrowRight className="size-3" />
            </Link>
          </div>
          <div className="flex flex-col gap-1">
            {recentSessions.map((session) => (
              <Link
                key={session.id}
                href="/chat"
                className="flex items-center justify-between px-4 py-3 rounded-lg hover:bg-muted/50 transition-colors"
              >
                <p className="text-sm truncate">{session.title}</p>
                <p className="text-xs text-muted-foreground shrink-0 ml-4">
                  {new Date(session.created_at).toLocaleDateString("es-AR", {
                    day: "numeric",
                    month: "short",
                  })}
                </p>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
