"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const typeLabel: Record<string, string> = { preventive: "Preventivo", corrective: "Correctivo", inspection: "Inspección" };

export default function EquiposPage() {
  const { data: workOrders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.workOrders(),
    queryFn: () => api.get<{ work_orders: any[] }>("/v1/erp/maintenance/work-orders?page_size=50"),
    select: (d) => d.work_orders,
  });

  if (error) return <ErrorState message="Error cargando equipos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Equipos y Órdenes de Trabajo</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{workOrders.length} órdenes</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-20">OT</TableHead><TableHead className="w-28">Fecha</TableHead>
              <TableHead>Equipo</TableHead><TableHead className="w-28">Tipo</TableHead>
              <TableHead className="w-28">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {workOrders.map((wo: any) => (
                <TableRow key={wo.id}>
                  <TableCell className="font-mono text-sm">{wo.number}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(wo.date)}</TableCell>
                  <TableCell className="text-sm">{wo.asset_name}</TableCell>
                  <TableCell><Badge variant="secondary">{typeLabel[wo.work_type] || wo.work_type}</Badge></TableCell>
                  <TableCell><Badge variant={wo.status === "completed" ? "default" : "secondary"}>{wo.status}</Badge></TableCell>
                </TableRow>
              ))}
              {workOrders.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin órdenes de trabajo.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
