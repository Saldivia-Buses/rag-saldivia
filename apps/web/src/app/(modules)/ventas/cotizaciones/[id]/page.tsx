"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney, fmtNumber } from "@/lib/erp/format";
import type { QuotationDetail, QuotationOption } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  sent: { label: "Enviada", variant: "default" },
  accepted: { label: "Aceptada", variant: "default" },
  rejected: { label: "Rechazada", variant: "outline" },
  expired: { label: "Vencida", variant: "secondary" },
};

export default function QuotationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.quotation(id),
    queryFn: () => api.get<QuotationDetail>(`/v1/erp/sales/quotations/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando cotización" onRetry={() => window.location.reload()} />;

  const q = data?.quotation;
  const lines = data?.lines ?? [];
  const options = data?.options ?? [];
  const status = q ? (statusBadge[q.status] ?? { label: q.status, variant: "outline" as const }) : null;

  const optionsBySection = options.reduce<Record<number, QuotationOption[]>>((acc, opt) => {
    (acc[opt.section_legacy_id] ??= []).push(opt);
    return acc;
  }, {});
  const sectionIds = Object.keys(optionsBySection).map(Number).sort((a, b) => a - b);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/ventas/cotizaciones"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a cotizaciones
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {q && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">Cotización {q.number}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Fecha {fmtDate(q.date)}
                  {q.valid_until ? ` · válida hasta ${fmtDate(q.valid_until)}` : ""}
                  {q.notes ? ` · ${q.notes}` : ""}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="Total" value={fmtMoney(q.total)} />
              <Metric label="Líneas" value={String(lines.length)} />
              <Metric label="Opciones" value={String(options.length)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Líneas</h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Descripción</TableHead>
                    <TableHead className="w-[100px] text-right">Cantidad</TableHead>
                    <TableHead className="w-[130px] text-right">Precio unit.</TableHead>
                    <TableHead className="w-[130px] text-right">Subtotal</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {lines.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                        Sin líneas.
                      </TableCell>
                    </TableRow>
                  )}
                  {lines.map((l) => {
                    const subtotal =
                      l.quantity != null && l.unit_price != null ? l.quantity * l.unit_price : null;
                    return (
                      <TableRow key={l.id}>
                        <TableCell className="text-sm">{l.description}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtNumber(l.quantity)}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtMoney(l.unit_price)}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtMoney(subtotal)}</TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>

            {options.length > 0 && (
              <>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground">Opciones por sección</h2>
                <div className="space-y-4">
                  {sectionIds.map((sid) => (
                    <div
                      key={sid}
                      className="overflow-hidden rounded-xl border border-border/40 bg-card"
                    >
                      <div className="border-b border-border/40 bg-muted/30 px-4 py-2 text-xs font-medium text-muted-foreground">
                        Sección {sid}
                      </div>
                      <ul className="divide-y divide-border/40">
                        {optionsBySection[sid].map((opt) => (
                          <li key={opt.id} className="px-4 py-2 text-sm">
                            {opt.description}
                          </li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>
              </>
            )}
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
