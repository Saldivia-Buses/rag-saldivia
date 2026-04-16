"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  planned: { label: "Planificada", variant: "secondary" },
  in_progress: { label: "En Producción", variant: "outline" },
  completed: { label: "Completada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

export default function PCPPage() {
  const { data: orders = [], isLoading: loadingOrders, error } = useQuery({
    queryKey: erpKeys.productionOrders(),
    queryFn: () => api.get<{ orders: any[] }>("/v1/erp/production/orders?page_size=100"),
    select: (d) => d.orders,
  });

  const { data: centers = [], isLoading: loadingCenters } = useQuery({
    queryKey: [...erpKeys.all, "production", "centers"] as const,
    queryFn: () => api.get<{ centers: any[] }>("/v1/erp/production/centers"),
    select: (d) => d.centers,
  });

  if (error) return <ErrorState message="Error cargando PCP" onRetry={() => window.location.reload()} />;
  if (loadingOrders || loadingCenters) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const planned = orders.filter((o: any) => o.status === "planned").length;
  const inProgress = orders.filter((o: any) => o.status === "in_progress").length;
  const completed = orders.filter((o: any) => o.status === "completed").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Planificación y Control de Producción</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {orders.length} órdenes · {centers.length} centros de trabajo
          </p>
        </div>

        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="rounded-xl border border-border/40 bg-card px-5 py-4">
            <p className="text-xs text-muted-foreground mb-1">Planificadas</p>
            <p className="text-2xl font-semibold tabular-nums">{planned}</p>
          </div>
          <div className="rounded-xl border border-border/40 bg-card px-5 py-4">
            <p className="text-xs text-muted-foreground mb-1">En producción</p>
            <p className="text-2xl font-semibold tabular-nums">{inProgress}</p>
          </div>
          <div className="rounded-xl border border-border/40 bg-card px-5 py-4">
            <p className="text-xs text-muted-foreground mb-1">Completadas</p>
            <p className="text-2xl font-semibold tabular-nums">{completed}</p>
          </div>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-24">OP</TableHead>
              <TableHead className="w-28">Fecha</TableHead>
              <TableHead>Producto</TableHead>
              <TableHead className="text-right w-20">Cant.</TableHead>
              <TableHead className="w-32">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {orders.map((o: any) => {
                const s = statusBadge[o.status] || statusBadge.planned;
                return (
                  <TableRow key={o.id}>
                    <TableCell className="font-mono text-sm">{o.number}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                    <TableCell className="text-sm">{o.product_name || o.product_code}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{o.quantity}</TableCell>
                    <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                  </TableRow>
                );
              })}
              {orders.length === 0 && (
                <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                  Sin órdenes de producción.
                </TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
