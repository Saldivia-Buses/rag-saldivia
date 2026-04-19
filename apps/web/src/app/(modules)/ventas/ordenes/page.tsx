"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtMoney } from "@/lib/erp/format";
import type { SalesOrder } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  confirmed: { label: "Confirmada", variant: "default" },
  in_production: { label: "En producción", variant: "outline" },
  delivered: { label: "Entregada", variant: "default" },
  cancelled: { label: "Cancelada", variant: "destructive" },
};

const typeLabel: Record<string, string> = {
  sale: "Venta",
  service: "Servicio",
  unit: "Unidad",
};

export default function SalesOrdersPage() {
  const { data: orders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.salesOrders(),
    queryFn: () => api.get<{ orders: SalesOrder[] }>("/v1/erp/sales/orders?page_size=100"),
    select: (d) => d.orders,
  });

  if (error)
    return <ErrorState message="Error cargando órdenes de venta" onRetry={() => window.location.reload()} />;
  if (isLoading)
    return (
      <div className="flex-1 p-8">
        <Skeleton className="h-[600px]" />
      </div>
    );

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Órdenes de venta</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">{orders.length} órdenes</p>
        </div>
        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-28">Número</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead className="w-28">Tipo</TableHead>
                <TableHead>Cliente</TableHead>
                <TableHead className="w-32 text-right">Total</TableHead>
                <TableHead className="w-32">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {orders.map((o) => {
                const s = statusBadge[o.status] ?? { label: o.status, variant: "outline" as const };
                return (
                  <TableRow key={o.id}>
                    <TableCell className="font-mono text-sm">
                      <Link href={`/ventas/ordenes/${o.id}`} className="hover:underline">
                        {o.number}
                      </Link>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(o.date)}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{typeLabel[o.order_type] ?? o.order_type}</Badge>
                    </TableCell>
                    <TableCell className="text-sm">{o.customer_name ?? "—"}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtMoney(o.total)}</TableCell>
                    <TableCell>
                      <Badge variant={s.variant}>{s.label}</Badge>
                    </TableCell>
                  </TableRow>
                );
              })}
              {orders.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    Sin órdenes de venta.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
