"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtNumber } from "@/lib/erp/format";
import type { QCInspection } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  pending: { label: "Pendiente", variant: "outline" },
  passed: { label: "Aprobada", variant: "default" },
  failed: { label: "Rechazada", variant: "destructive" },
  partial: { label: "Parcial", variant: "secondary" },
};

export default function InspectionDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.qcInspection(id),
    queryFn: () => api.get<QCInspection>(`/v1/erp/purchasing/inspections/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando inspección" onRetry={() => window.location.reload()} />;

  const i = data;
  const status = i ? (statusBadge[i.status] ?? { label: i.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/calidad/inspecciones"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a inspecciones
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {i && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  Inspección {i.id.slice(0, 8)}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Fecha {fmtDate(i.created_at)}
                  {i.completed_at ? ` · completada ${fmtDate(i.completed_at)}` : ""}
                  {i.inspector_id ? ` · inspector ${i.inspector_id}` : ""}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="Cantidad" value={fmtNumber(i.quantity)} />
              <Metric label="Aceptada" value={fmtNumber(i.accepted_qty)} />
              <Metric label="Rechazada" value={fmtNumber(i.rejected_qty)} />
            </div>

            <div className="mb-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Artículo</div>
                {i.article_id ? (
                  <Link
                    href={`/administracion/almacen/articulos/${i.article_id}`}
                    className="mt-1 block font-mono text-sm hover:underline"
                  >
                    {i.article_code ?? i.article_id.slice(0, 8)}
                    {i.article_name ? ` — ${i.article_name}` : ""}
                  </Link>
                ) : (
                  <div className="mt-1 font-mono text-sm">—</div>
                )}
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Recepción</div>
                <div className="mt-1 font-mono text-sm">
                  {i.receipt_number ?? (i.receipt_id ? i.receipt_id.slice(0, 8) : "—")}
                </div>
              </div>
            </div>

            {i.notes && (
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Notas</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{i.notes}</p>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 font-mono text-sm">{value}</div>
    </div>
  );
}
