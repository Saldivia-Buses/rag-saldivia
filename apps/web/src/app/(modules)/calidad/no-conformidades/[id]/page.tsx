"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

interface NCDetail {
  id: string;
  tenant_id: string;
  number: string;
  date: string | null;
  type_id: string | null;
  origin_id: string | null;
  description: string;
  severity: string;
  status: string;
  assigned_to: string | null;
  closed_at: string | null;
  user_id: string;
  created_at: string;
}

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  open: { label: "Abierta", variant: "destructive" },
  in_progress: { label: "En curso", variant: "secondary" },
  closed: { label: "Cerrada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "outline" },
};

const severityBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  low: { label: "Baja", variant: "outline" },
  medium: { label: "Media", variant: "secondary" },
  high: { label: "Alta", variant: "destructive" },
  critical: { label: "Crítica", variant: "destructive" },
};

export default function NCDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: nc, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "nc", id] as const,
    queryFn: () => api.get<NCDetail>(`/v1/erp/quality/nc/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando no-conformidad" onRetry={() => window.location.reload()} />;

  const status = nc ? (statusBadge[nc.status] ?? { label: nc.status, variant: "outline" as const }) : null;
  const severity = nc ? (severityBadge[nc.severity] ?? { label: nc.severity, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/calidad/no-conformidades"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a no-conformidades
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {nc && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  NC {nc.number}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {nc.date ? fmtDate(nc.date) : "—"} · reportada por {nc.user_id}
                </p>
              </div>
              <div className="flex gap-2">
                {severity && <Badge variant={severity.variant}>{severity.label}</Badge>}
                {status && <Badge variant={status.variant}>{status.label}</Badge>}
              </div>
            </div>

            <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
              <div className="text-xs text-muted-foreground">Descripción</div>
              <p className="mt-1 text-sm whitespace-pre-wrap">{nc.description || "—"}</p>
            </div>

            <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Tipo</div>
                <div className="mt-1 font-mono text-sm">
                  {nc.type_id ? nc.type_id.slice(0, 8) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Origen</div>
                <div className="mt-1 font-mono text-sm">
                  {nc.origin_id ? nc.origin_id.slice(0, 8) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Asignada a</div>
                <div className="mt-1 font-mono text-sm">
                  {nc.assigned_to ? nc.assigned_to.slice(0, 8) : "—"}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Creada</div>
                <div className="mt-1 font-mono text-sm">{fmtDate(nc.created_at)}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Cerrada</div>
                <div className="mt-1 font-mono text-sm">
                  {nc.closed_at ? fmtDate(nc.closed_at) : "—"}
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
