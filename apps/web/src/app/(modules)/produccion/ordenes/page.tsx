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

export default function ProduccionOrdenesPage() {
  const { data: orders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.productionOrders(),
    queryFn: () => api.get<{ orders: any[] }>("/v1/erp/production/orders?page_size=50"),
    select: (d) => d.orders,
  });

  if (error) return <ErrorState message="Error cargando órdenes de producción" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Órdenes de Producción</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{orders.length} órdenes</p>
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
              {orders.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin órdenes de producción.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
