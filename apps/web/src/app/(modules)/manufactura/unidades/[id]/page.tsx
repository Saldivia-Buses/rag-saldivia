"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface ManufacturingUnitDetail {
  id: string;
  work_order_number: number;
  chassis_serial: string;
  engine_number: string;
  chassis_brand_name: string | null;
  chassis_model_name: string | null;
  carroceria_model_name: string | null;
  customer_id: string | null;
  customer_name: string | null;
  entry_date: string;
  expected_completion: string | null;
  delivery_date: string | null;
  status: string;
  observations: string | null;
}

interface PendingControl {
  id: string;
  control_id: string;
  name: string;
  sequence: number;
  required: boolean;
}

interface ControlExecution {
  id: string;
  control_id: string;
  control_name: string;
  executed_at: string;
  executed_by: string;
  result: string;
  observations: string | null;
}

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  pending: { label: "Pendiente", variant: "secondary" },
  in_production: { label: "En producción", variant: "outline" },
  completed: { label: "Terminada", variant: "default" },
  delivered: { label: "Entregada", variant: "default" },
  returned: { label: "Devuelta", variant: "destructive" },
};

export default function UnitDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: unit, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "units", id] as const,
    queryFn: () => api.get<ManufacturingUnitDetail>(`/v1/erp/manufacturing/units/${id}`),
    enabled: !!id,
  });

  const { data: pending = [] } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "units", id, "pending-controls"] as const,
    queryFn: () =>
      api.get<{ controls: PendingControl[] }>(`/v1/erp/manufacturing/units/${id}/pending-controls`),
    select: (d) => d.controls,
    enabled: !!id,
  });

  const { data: executions = [] } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "units", id, "executions"] as const,
    queryFn: () =>
      api.get<{ executions: ControlExecution[] }>(`/v1/erp/manufacturing/units/${id}/executions`),
    select: (d) => d.executions,
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando unidad" onRetry={() => window.location.reload()} />;

  const status = unit ? (statusBadge[unit.status] ?? { label: unit.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <Link
          href="/manufactura/unidades"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a unidades
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {unit && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  OT {unit.work_order_number}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Chasis <span className="font-mono">{unit.chassis_serial || "—"}</span>
                  {" · "}Motor <span className="font-mono">{unit.engine_number || "—"}</span>
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Cliente</div>
                <div className="mt-1 text-sm">{unit.customer_name || "—"}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Carrocería</div>
                <div className="mt-1 text-sm">{unit.carroceria_model_name || "—"}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Chasis modelo</div>
                <div className="mt-1 text-sm">
                  {unit.chassis_brand_name ?? ""} {unit.chassis_model_name ?? ""}
                  {!unit.chassis_brand_name && !unit.chassis_model_name ? "—" : ""}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Ingreso</div>
                <div className="mt-1 font-mono text-sm">{fmtDateShort(unit.entry_date)}</div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Entrega esperada</div>
                <div className="mt-1 font-mono text-sm">
                  {unit.expected_completion ? fmtDateShort(unit.expected_completion) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Entregada</div>
                <div className="mt-1 font-mono text-sm">
                  {unit.delivery_date ? fmtDateShort(unit.delivery_date) : "—"}
                </div>
              </div>
            </div>

            {unit.observations && (
              <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Observaciones</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{unit.observations}</p>
              </div>
            )}

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Controles pendientes ({pending.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[70px]">Seq</TableHead>
                    <TableHead>Control</TableHead>
                    <TableHead className="w-[120px]">Obligatorio</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pending.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className="h-16 text-center text-sm text-muted-foreground">
                        Sin controles pendientes.
                      </TableCell>
                    </TableRow>
                  )}
                  {pending.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell className="font-mono text-xs">{p.sequence}</TableCell>
                      <TableCell className="text-sm">{p.name}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {p.required ? "Sí" : "No"}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Historial de controles ({executions.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[120px]">Fecha</TableHead>
                    <TableHead>Control</TableHead>
                    <TableHead className="w-[120px]">Resultado</TableHead>
                    <TableHead className="w-[140px]">Operario</TableHead>
                    <TableHead>Observaciones</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {executions.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-16 text-center text-sm text-muted-foreground">
                        Sin ejecuciones registradas.
                      </TableCell>
                    </TableRow>
                  )}
                  {executions.map((ex) => (
                    <TableRow key={ex.id}>
                      <TableCell className="font-mono text-xs">{fmtDateShort(ex.executed_at)}</TableCell>
                      <TableCell className="text-sm">{ex.control_name}</TableCell>
                      <TableCell className="text-sm">
                        <Badge variant={ex.result === "passed" ? "default" : "destructive"}>
                          {ex.result}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono text-xs">{ex.executed_by}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {ex.observations || "—"}
                      </TableCell>
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
