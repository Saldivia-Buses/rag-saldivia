"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { useAuthStore } from "@/lib/auth/store";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { ScrollArea } from "@/components/ui/scroll-area";
import CardStandard4 from "@/components/card-standard-4";
import { cn } from "@/lib/utils";

interface TenantUser {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: string;
}

function getInitials(name: string): string {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join("") || "U";
}

function UsersSidebar({ users, currentUserId }: { users: TenantUser[]; currentUserId?: string }) {
  return (
    <div className="w-72 shrink-0 flex flex-col min-h-0">
      <div className="px-4 py-4">
        <p className="text-xs text-muted-foreground uppercase tracking-wide">Equipo</p>
      </div>
      <ScrollArea className="flex-1">
        <div className="space-y-0.5 px-2">
          {users.map((user) => {
            const isYou = user.id === currentUserId;
            return (
              <div
                key={user.id}
                className="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-accent/50 transition-colors"
              >
                <span
                  className="size-2 shrink-0 rounded-full bg-muted-foreground/40"
                  aria-hidden="true"
                />
                <Avatar className="size-7">
                  <AvatarFallback className="bg-primary text-xs font-medium text-primary-foreground">
                    {getInitials(user.name)}
                  </AvatarFallback>
                </Avatar>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="truncate text-sm font-medium">{user.name}</span>
                    {isYou && (
                      <span className="text-[10px] text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                        Vos
                      </span>
                    )}
                  </div>
                  <p className="truncate text-xs text-muted-foreground capitalize">{user.role}</p>
                </div>
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user);

  const { data: users = [] } = useQuery({
    queryKey: ["auth", "users"],
    queryFn: () => api.get<TenantUser[]>("/v1/auth/users"),
  });

  return (
    <div className="flex flex-1 min-h-0">
      {/* Main content */}
      <div className="flex-1 overflow-y-auto p-8">
        <div className="flex flex-wrap gap-4">
          <CardStandard4 title="Producción" description="Órdenes y seguimiento de unidades" href="/produccion" />
          <CardStandard4 title="Calidad" description="Inspecciones y trazabilidad" href="/calidad" />
          <CardStandard4 title="Compras" description="Órdenes de compra y proveedores" href="/compras" />
          <CardStandard4 title="Documentos" description="Base de conocimiento empresarial" href="/documents" />
        </div>
      </div>

      {/* Users sidebar */}
      <UsersSidebar users={users} currentUserId={user?.id} />
    </div>
  );
}
