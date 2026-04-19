"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtNumber } from "@/lib/erp/format";
import type { WorkOrderDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  open: { label: "Abierta", variant: "destructive" },
  in_progress: { label: "En curso", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

const typeLabel: Record<string, string> = {
  corrective: "Correctivo",
  preventive: "Preventivo",
  predictive: "Predictivo",
};

const priorityLabel: Record<string, string> = {
  low: "Baja",
  normal: "Normal",
  high: "Alta",
  urgent: "Urgente",
};

export default function WorkOrderDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.workOrder(id),
    queryFn: () => api.get<WorkOrderDetail>(`/v1/erp/maintenance/work-orders/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando orden de trabajo" onRetry={() => window.location.reload()} />;

  const wo = data?.order;
  const parts = data?.parts ?? [];
  const status = wo ? (statusBadge[wo.status] ?? { label: wo.status, variant: "outline" as const }) : null;

  const backHref = wo?.work_type === "preventive" ? "/mantenimiento/preventivo" : "/mantenimiento/correctivo";

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href={backHref}
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a mantenimiento
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {wo && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">OT {wo.number}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {typeLabel[wo.work_type] ?? wo.work_type} · {fmtDate(wo.date)}
                  {wo.asset_id && (
                    <>
                      {" "}· equipo{" "}
                      <Link href={`/mantenimiento/equipos/${wo.asset_id}`} className="hover:underline">
                        {wo.asset_code ?? wo.asset_id.slice(0, 8)}
                      </Link>
                    </>
                  )}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Prioridad" value={priorityLabel[wo.priority] ?? wo.priority ?? "—"} />
              <Metric label="Completada" value={fmtDate(wo.completed_at)} />
              <Metric label="Repuestos" value={String(parts.length)} />
              <Metric label="Creada" value={fmtDate(wo.created_at)} />
            </div>

            {wo.description && (
              <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Descripción</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{wo.description}</p>
              </div>
            )}

            {wo.notes && (
              <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Notas</div>
                <p className="mt-1 text-sm whitespace-pre-wrap text-muted-foreground">{wo.notes}</p>
              </div>
            )}

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Repuestos usados ({parts.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px]">Cód. art.</TableHead>
                    <TableHead>Artículo</TableHead>
                    <TableHead className="w-[120px] text-right">Cantidad</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {parts.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                        Sin repuestos registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {parts.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell className="font-mono text-xs">
                        <Link
                          href={`/administracion/almacen/articulos/${p.article_id}`}
                          className="hover:underline"
                        >
                          {p.article_code}
                        </Link>
                      </TableCell>
                      <TableCell className="text-sm">{p.article_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtNumber(p.quantity)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
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
