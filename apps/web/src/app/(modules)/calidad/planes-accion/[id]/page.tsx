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

interface ActionPlanDetail {
  id: string;
  tenant_id: string;
  nonconformity_id: string | null;
  responsible_id: string | null;
  section_id: string | null;
  description: string;
  planned_start: string | null;
  target_date: string | null;
  closed_date: string | null;
  time_savings_hours: string | null;
  cost_savings: string | null;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  open: { label: "Abierto", variant: "outline" },
  in_progress: { label: "En curso", variant: "secondary" },
  closed: { label: "Cerrado", variant: "default" },
  cancelled: { label: "Cancelado", variant: "destructive" },
};

export default function ActionPlanDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: plan, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "action-plans", id] as const,
    queryFn: () => api.get<ActionPlanDetail>(`/v1/erp/quality/action-plans/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando plan de acción" onRetry={() => window.location.reload()} />;

  const status = plan ? (statusBadge[plan.status] ?? { label: plan.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/calidad/planes-accion"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a planes de acción
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {plan && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  Plan {plan.id.slice(0, 8)}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Creado {fmtDate(plan.created_at)} por {plan.created_by}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
              <div className="text-xs text-muted-foreground">Descripción</div>
              <p className="mt-1 text-sm whitespace-pre-wrap">{plan.description || "—"}</p>
            </div>

            <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Inicio planificado</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.planned_start ? fmtDate(plan.planned_start) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Fecha objetivo</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.target_date ? fmtDate(plan.target_date) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Cerrado</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.closed_date ? fmtDate(plan.closed_date) : "—"}
                </div>
              </div>
            </div>

            <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">No-conformidad</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.nonconformity_id ? (
                    <Link
                      href={`/calidad/no-conformidades/${plan.nonconformity_id}`}
                      className="hover:underline"
                    >
                      {plan.nonconformity_id.slice(0, 8)}
                    </Link>
                  ) : (
                    "—"
                  )}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Responsable</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.responsible_id ? plan.responsible_id.slice(0, 8) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Sección</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.section_id ? plan.section_id.slice(0, 8) : "—"}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Ahorro horas</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.time_savings_hours ? Number(plan.time_savings_hours).toFixed(2) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Ahorro costo</div>
                <div className="mt-1 font-mono text-sm">
                  {plan.cost_savings ? Number(plan.cost_savings).toFixed(2) : "—"}
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
