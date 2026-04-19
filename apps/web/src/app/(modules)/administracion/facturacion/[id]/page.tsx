"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney, fmtNumber, fmtPercent } from "@/lib/erp/format";
import type { InvoiceDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const typeLabel: Record<string, string> = {
  invoice_a: "Factura A",
  invoice_b: "Factura B",
  invoice_c: "Factura C",
  invoice_e: "Factura E",
  credit_note: "Nota Crédito",
  debit_note: "Nota Débito",
  delivery_note: "Remito",
};

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  posted: { label: "Contabilizada", variant: "default" },
  paid: { label: "Cobrada", variant: "default" },
  cancelled: { label: "Anulada", variant: "secondary" },
};

export default function InvoiceDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.invoice(id),
    queryFn: () => api.get<InvoiceDetail>(`/v1/erp/invoicing/invoices/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando comprobante" onRetry={() => window.location.reload()} />;

  const inv = data?.invoice;
  const lines = data?.lines ?? [];
  const status = inv ? (statusBadge[inv.status] ?? { label: inv.status, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/facturacion"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a facturación
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {inv && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  {typeLabel[inv.invoice_type] ?? inv.invoice_type} {inv.number}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Fecha {fmtDate(inv.date)}
                  {inv.due_date ? ` · vence ${fmtDate(inv.due_date)}` : ""}
                  {inv.direction === "incoming" ? " · recepción" : " · emisión"}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Subtotal" value={fmtMoney(inv.subtotal ?? null)} />
              <Metric label="IVA" value={fmtMoney(inv.tax_amount ?? null)} />
              <Metric label="Total" value={fmtMoney(inv.total)} />
              <Metric label="CAE" value={inv.afip_cae ?? "\u2014"} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Líneas ({lines.length})</h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Descripción</TableHead>
                    <TableHead className="w-[90px] text-right">Cant.</TableHead>
                    <TableHead className="w-[120px] text-right">Precio unit.</TableHead>
                    <TableHead className="w-[80px] text-right">IVA %</TableHead>
                    <TableHead className="w-[120px] text-right">IVA</TableHead>
                    <TableHead className="w-[130px] text-right">Subtotal</TableHead>
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
                  {lines.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell className="text-sm">{l.description}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtNumber(l.quantity)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(l.unit_price)}</TableCell>
                      <TableCell className="text-right font-mono text-xs text-muted-foreground">
                        {fmtPercent(l.tax_rate)}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(l.tax_amount)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(l.line_total)}</TableCell>
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

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 font-mono text-sm">{value}</div>
    </div>
  );
}
