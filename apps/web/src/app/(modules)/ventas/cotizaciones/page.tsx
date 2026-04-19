"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import type { Quotation } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  sent: { label: "Enviada", variant: "default" },
  accepted: { label: "Aceptada", variant: "default" },
  rejected: { label: "Rechazada", variant: "outline" },
  expired: { label: "Vencida", variant: "secondary" },
};

export default function CotizacionesPage() {
  const { data: quotations = [], isLoading, error } = useQuery({
    queryKey: erpKeys.quotations(),
    queryFn: () => api.get<{ quotations: Quotation[] }>("/v1/erp/sales/quotations?page_size=50"),
    select: (d) => d.quotations,
  });

  if (error)
    return <ErrorState message="Error cargando cotizaciones" onRetry={() => window.location.reload()} />;
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
          <h1 className="text-xl font-semibold tracking-tight">Cotizaciones</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">{quotations.length} cotizaciones</p>
        </div>
        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-24">Número</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead>Cliente</TableHead>
                <TableHead className="w-28">Vence</TableHead>
                <TableHead className="w-28 text-right">Total</TableHead>
                <TableHead className="w-28">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {quotations.map((q) => {
                const s = statusBadge[q.status] ?? { label: q.status, variant: "outline" as const };
                return (
                  <TableRow key={q.id}>
                    <TableCell className="font-mono text-sm">
                      <Link href={`/ventas/cotizaciones/${q.id}`} className="hover:underline">
                        {q.number}
                      </Link>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(q.date)}</TableCell>
                    <TableCell className="text-sm">{q.customer_name ?? "—"}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(q.valid_until)}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtMoney(q.total)}</TableCell>
                    <TableCell>
                      <Badge variant={s.variant}>{s.label}</Badge>
                    </TableCell>
                  </TableRow>
                );
              })}
              {quotations.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    Sin cotizaciones.
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
