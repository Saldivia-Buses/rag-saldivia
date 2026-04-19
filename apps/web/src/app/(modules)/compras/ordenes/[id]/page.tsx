"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney, fmtNumber } from "@/lib/erp/format";
import type { PurchaseOrderDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  approved: { label: "Aprobada", variant: "default" },
  partial: { label: "Parcial", variant: "outline" },
  received: { label: "Recibida", variant: "default" },
  cancelled: { label: "Cancelada", variant: "secondary" },
};

export default function PurchaseOrderDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.purchaseOrder(id),
    queryFn: () => api.get<PurchaseOrderDetail>(`/v1/erp/purchasing/orders/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando orden de compra" onRetry={() => window.location.reload()} />;

  const order = data?.order;
  const lines = data?.lines ?? [];
  const status = order ? (statusBadge[order.status] ?? { label: order.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/compras/ordenes"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a órdenes de compra
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {order && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">OC {order.number}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Fecha {fmtDate(order.date)}
                  {order.notes ? ` · ${order.notes}` : ""}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="Total" value={fmtMoney(order.total)} />
              <Metric label="Líneas" value={String(lines.length)} />
              <Metric label="Creada" value={fmtDate(order.created_at)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Líneas</h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[120px]">Cód. art.</TableHead>
                    <TableHead>Artículo</TableHead>
                    <TableHead className="w-[100px] text-right">Cantidad</TableHead>
                    <TableHead className="w-[100px] text-right">Recibida</TableHead>
                    <TableHead className="w-[120px] text-right">Precio unit.</TableHead>
                    <TableHead className="w-[120px] text-right">Subtotal</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {lines.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                        Sin líneas.
                      </TableCell>
                    </TableRow>
                  )}
                  {lines.map((l) => {
                    const subtotal =
                      l.quantity != null && l.unit_price != null ? l.quantity * l.unit_price : null;
                    return (
                      <TableRow key={l.id}>
                        <TableCell className="font-mono text-xs">{l.article_code}</TableCell>
                        <TableCell className="text-sm">{l.article_name}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtNumber(l.quantity)}</TableCell>
                        <TableCell className="text-right font-mono text-sm text-muted-foreground">
                          {fmtNumber(l.received_qty)}
                        </TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtMoney(l.unit_price)}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtMoney(subtotal)}</TableCell>
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
