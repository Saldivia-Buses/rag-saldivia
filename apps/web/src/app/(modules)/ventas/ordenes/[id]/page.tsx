"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney } from "@/lib/erp/format";
import type { SalesOrder } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";

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

export default function SalesOrderDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: order, isLoading, error } = useQuery({
    queryKey: erpKeys.salesOrder(id),
    queryFn: () => api.get<SalesOrder>(`/v1/erp/sales/orders/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando orden de venta" onRetry={() => window.location.reload()} />;

  const status = order ? (statusBadge[order.status] ?? { label: order.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/ventas/ordenes"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a órdenes de venta
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {order && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">Orden {order.number}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {typeLabel[order.order_type] ?? order.order_type} · {fmtDate(order.date)}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="Total" value={fmtMoney(order.total)} />
              <Metric label="Creada" value={fmtDate(order.created_at)} />
              <Metric label="Usuario" value={order.user_id || "—"} />
            </div>

            <div className="mb-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Cliente</div>
                <div className="mt-1 font-mono text-sm">
                  {order.customer_id ? order.customer_id.slice(0, 8) : "—"}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Cotización origen</div>
                {order.quotation_id ? (
                  <Link
                    href={`/ventas/cotizaciones/${order.quotation_id}`}
                    className="mt-1 block font-mono text-sm hover:underline"
                  >
                    {order.quotation_id.slice(0, 8)} →
                  </Link>
                ) : (
                  <div className="mt-1 font-mono text-sm">—</div>
                )}
              </div>
            </div>

            {order.notes && (
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Notas</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{order.notes}</p>
              </div>
            )}

            <p className="mt-6 text-xs text-muted-foreground">
              Las órdenes de venta son header-only — no hay tabla de líneas en el esquema actual.
              Los items se describen en la cotización origen.
            </p>
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
