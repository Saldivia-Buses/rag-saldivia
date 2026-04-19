"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtNumber } from "@/lib/erp/format";
import type { ProductionOrderDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  planned: { label: "Planificada", variant: "secondary" },
  in_progress: { label: "En producción", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "destructive" },
};

const stepStatusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  pending: { label: "Pendiente", variant: "secondary" },
  in_progress: { label: "En curso", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
};

const inspectionResultBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  passed: { label: "Aprobada", variant: "default" },
  failed: { label: "Rechazada", variant: "destructive" },
  pending: { label: "Pendiente", variant: "outline" },
};

export default function ProductionOrderDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.productionOrder(id),
    queryFn: () => api.get<ProductionOrderDetail>(`/v1/erp/production/orders/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando orden de producción" onRetry={() => window.location.reload()} />;

  const o = data?.order;
  const materials = data?.materials ?? [];
  const steps = data?.steps ?? [];
  const inspections = data?.inspections ?? [];
  const status = o ? (statusBadge[o.status] ?? { label: o.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/produccion/ordenes"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a órdenes de producción
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {o && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">OP {o.number}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {o.product_name ?? o.product_code ?? "producto —"} · {fmtDate(o.date)}
                  {o.order_id && (
                    <>
                      {" "}· venta origen{" "}
                      <Link href={`/ventas/ordenes/${o.order_id}`} className="hover:underline">
                        {o.order_id.slice(0, 8)}
                      </Link>
                    </>
                  )}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Cantidad" value={fmtNumber(o.quantity)} />
              <Metric label="Prioridad" value={String(o.priority)} />
              <Metric label="Inicio" value={fmtDate(o.start_date)} />
              <Metric label="Fin planificado" value={fmtDate(o.end_date)} />
            </div>

            {o.notes && (
              <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Notas</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{o.notes}</p>
              </div>
            )}

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Materiales ({materials.length})</h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px]">Cód. art.</TableHead>
                    <TableHead>Artículo</TableHead>
                    <TableHead className="w-[110px] text-right">Requerida</TableHead>
                    <TableHead className="w-[110px] text-right">Consumida</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {materials.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                        Sin materiales asignados.
                      </TableCell>
                    </TableRow>
                  )}
                  {materials.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell className="font-mono text-xs">
                        <Link
                          href={`/administracion/almacen/articulos/${m.article_id}`}
                          className="hover:underline"
                        >
                          {m.article_code}
                        </Link>
                      </TableCell>
                      <TableCell className="text-sm">{m.article_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtNumber(m.required_qty)}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-muted-foreground">
                        {fmtNumber(m.consumed_qty)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Pasos ({steps.length})</h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[60px]">#</TableHead>
                    <TableHead>Paso</TableHead>
                    <TableHead className="w-[130px]">Inicio</TableHead>
                    <TableHead className="w-[130px]">Fin</TableHead>
                    <TableHead className="w-[130px]">Estado</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {steps.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                        Sin pasos definidos.
                      </TableCell>
                    </TableRow>
                  )}
                  {steps.map((st) => {
                    const sb = stepStatusBadge[st.status] ?? { label: st.status, variant: "outline" as const };
                    return (
                      <TableRow key={st.id}>
                        <TableCell className="font-mono text-xs">{st.sort_order}</TableCell>
                        <TableCell className="text-sm">{st.step_name}</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">{fmtDate(st.started_at)}</TableCell>
                        <TableCell className="font-mono text-xs">{fmtDate(st.completed_at)}</TableCell>
                        <TableCell>
                          <Badge variant={sb.variant}>{sb.label}</Badge>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Inspecciones ({inspections.length})</h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[130px]">Fecha</TableHead>
                    <TableHead className="w-[120px]">Inspector</TableHead>
                    <TableHead className="w-[130px]">Resultado</TableHead>
                    <TableHead>Observaciones</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {inspections.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                        Sin inspecciones.
                      </TableCell>
                    </TableRow>
                  )}
                  {inspections.map((ins) => {
                    const ib = inspectionResultBadge[ins.result] ?? { label: ins.result, variant: "outline" as const };
                    return (
                      <TableRow key={ins.id}>
                        <TableCell className="font-mono text-xs text-muted-foreground">{fmtDate(ins.created_at)}</TableCell>
                        <TableCell className="font-mono text-xs">{ins.inspector_id}</TableCell>
                        <TableCell>
                          <Badge variant={ib.variant}>{ib.label}</Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{ins.observations || "—"}</TableCell>
                      </TableRow>
                    );
                  })}
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
