"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney } from "@/lib/erp/format";
import type { ReceiptDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const methodLabel: Record<string, string> = {
  cash: "Efectivo",
  check: "Cheque",
  transfer: "Transferencia",
  card: "Tarjeta",
};

export default function ReceiptDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.receipt(id),
    queryFn: () => api.get<ReceiptDetail>(`/v1/erp/treasury/receipts/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando recibo" onRetry={() => window.location.reload()} />;

  const r = data?.receipt;
  const payments = data?.payments ?? [];
  const allocations = data?.allocations ?? [];

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/tesoreria"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a tesorería
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {r && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  {r.receipt_type === "collection" ? "Cobro" : "Pago"} {r.number}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Fecha {fmtDate(r.date)}
                  {r.entity_name ? ` · ${r.entity_name}` : ""}
                  {r.notes ? ` · ${r.notes}` : ""}
                </p>
              </div>
              <Badge variant={r.status === "confirmed" ? "default" : "secondary"}>
                {r.status === "confirmed" ? "Confirmado" : r.status}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="Total" value={fmtMoney(r.total)} />
              <Metric label="Pagos" value={String(payments.length)} />
              <Metric label="Imputaciones" value={String(allocations.length)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Pagos ({payments.length})</h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px]">Método</TableHead>
                    <TableHead className="w-[140px] text-right">Monto</TableHead>
                    <TableHead>Notas</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {payments.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                        Sin pagos.
                      </TableCell>
                    </TableRow>
                  )}
                  {payments.map((p) => (
                    <TableRow key={p.id}>
                      <TableCell>
                        <Badge variant="outline">{methodLabel[p.payment_method] ?? p.payment_method}</Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(p.amount)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{p.notes || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Imputaciones ({allocations.length})</h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[160px]">Comprobante</TableHead>
                    <TableHead className="w-[160px] text-right">Total comp.</TableHead>
                    <TableHead className="w-[160px] text-right">Imputado</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {allocations.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                        Sin imputaciones.
                      </TableCell>
                    </TableRow>
                  )}
                  {allocations.map((a) => (
                    <TableRow key={a.id}>
                      <TableCell className="font-mono text-sm">
                        <Link
                          href={`/administracion/facturacion/${a.invoice_id}`}
                          className="hover:underline"
                        >
                          {a.invoice_number}
                        </Link>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm text-muted-foreground">
                        {fmtMoney(a.invoice_total)}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(a.amount)}</TableCell>
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
