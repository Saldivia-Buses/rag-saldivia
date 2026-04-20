"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney } from "@/lib/erp/format";
import type { BankReconciliation } from "@/lib/erp/types";
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

export default function ReconciliacionesPage() {
  const { data: recons = [], isLoading, error } = useQuery({
    queryKey: erpKeys.bankReconciliations(),
    queryFn: () => api.get<{ reconciliations: BankReconciliation[] }>("/v1/erp/treasury/reconciliations"),
    select: (d) => d.reconciliations,
  });

  if (error)
    return <ErrorState message="Error cargando reconciliaciones" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Reconciliaciones bancarias</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Cierres mensuales por cuenta — saldo del extracto vs saldo libro.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Período</TableHead>
                <TableHead>Banco</TableHead>
                <TableHead className="w-[140px]">Cuenta</TableHead>
                <TableHead className="w-[140px] text-right">Saldo extracto</TableHead>
                <TableHead className="w-[140px] text-right">Saldo libro</TableHead>
                <TableHead className="w-[130px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={6}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && recons.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin reconciliaciones registradas.
                  </TableCell>
                </TableRow>
              )}
              {recons.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="font-mono text-sm">
                    <Link
                      href={`/tesoreria/reconciliaciones/${r.id}`}
                      className="hover:underline"
                    >
                      {r.period}
                    </Link>
                  </TableCell>
                  <TableCell className="text-sm">{r.bank_name}</TableCell>
                  <TableCell className="font-mono text-xs">{r.account_number}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(r.statement_balance)}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(r.book_balance)}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[r.status] ?? "outline"}>
                      {statusLabel[r.status] ?? r.status}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
