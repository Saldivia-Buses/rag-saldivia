"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Check, X } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney } from "@/lib/erp/format";
import type { ReconciliationDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  draft: "outline",
  confirmed: "secondary",
};

const statusLabel: Record<string, string> = {
  draft: "Borrador",
  confirmed: "Confirmada",
};

export default function ReconciliacionDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.bankReconciliation(id),
    queryFn: () => api.get<ReconciliationDetail>(`/v1/erp/treasury/reconciliations/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando reconciliación" onRetry={() => window.location.reload()} />;

  const recon = data?.reconciliation;
  const lines = data?.lines ?? [];

  const diff =
    recon?.statement_balance != null && recon?.book_balance != null
      ? recon.statement_balance - recon.book_balance
      : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/tesoreria/reconciliaciones"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a reconciliaciones
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {recon && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  Reconciliación {recon.period}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {recon.bank_name} · cuenta {recon.account_number}
                </p>
              </div>
              <Badge variant={statusVariant[recon.status] ?? "outline"}>
                {statusLabel[recon.status] ?? recon.status}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Saldo extracto" value={fmtMoney(recon.statement_balance)} />
              <Metric label="Saldo libro" value={fmtMoney(recon.book_balance)} />
              <Metric label="Diferencia" value={fmtMoney(diff)} />
              <Metric label="Confirmada" value={fmtDate(recon.confirmed_at)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Líneas del extracto ({lines.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[110px]">Fecha</TableHead>
                    <TableHead>Descripción</TableHead>
                    <TableHead className="w-[140px]">Referencia</TableHead>
                    <TableHead className="w-[140px] text-right">Importe</TableHead>
                    <TableHead className="w-[100px] text-center">Conciliada</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {lines.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                        Sin líneas importadas.
                      </TableCell>
                    </TableRow>
                  )}
                  {lines.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell className="font-mono text-xs">{fmtDate(l.date)}</TableCell>
                      <TableCell className="text-sm">{l.description}</TableCell>
                      <TableCell className="font-mono text-xs">{l.reference || "—"}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(l.amount)}</TableCell>
                      <TableCell className="text-center">
                        {l.matched ? (
                          <Check className="mx-auto h-4 w-4 text-emerald-500" />
                        ) : (
                          <X className="mx-auto h-4 w-4 text-muted-foreground" />
                        )}
                      </TableCell>
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
