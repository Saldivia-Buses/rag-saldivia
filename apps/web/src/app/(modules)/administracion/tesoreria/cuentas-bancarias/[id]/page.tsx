"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { BankAccount, BankReconciliation } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function BankAccountDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const accountQ = useQuery({
    queryKey: erpKeys.bankAccount(id),
    queryFn: () => api.get<BankAccount>(`/v1/erp/treasury/bank-accounts/${id}`),
    enabled: !!id,
  });

  const reconciliationsQ = useQuery({
    queryKey: [...erpKeys.all, "treasury", "reconciliations", { bank_account_id: id }] as const,
    queryFn: () =>
      api.get<{ reconciliations: BankReconciliation[] }>(
        `/v1/erp/treasury/reconciliations?bank_account_id=${id}`,
      ),
    select: (d) => d.reconciliations,
    enabled: !!id,
  });

  if (accountQ.error)
    return <ErrorState message="Error cargando cuenta bancaria" onRetry={() => window.location.reload()} />;

  const account = accountQ.data;
  const reconciliations = reconciliationsQ.data ?? [];

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/tesoreria/cuentas-bancarias"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a cuentas bancarias
        </Link>

        {accountQ.isLoading && <Skeleton className="h-48 w-full" />}

        {account && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{account.bank_name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {account.branch ? `${account.branch} · ` : ""}
                  <span className="font-mono">{account.account_number}</span>
                </p>
              </div>
              <Badge variant={account.active ? "default" : "secondary"}>
                {account.active ? "Activa" : "Baja"}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-3">
              <Metric label="CBU" value={account.cbu || "—"} />
              <Metric label="Alias" value={account.alias || "—"} />
              <Metric label="Reconciliaciones" value={String(reconciliations.length)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Reconciliaciones ({reconciliations.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[110px]">Período</TableHead>
                    <TableHead className="w-[150px] text-right">Saldo extracto</TableHead>
                    <TableHead className="w-[150px] text-right">Saldo libros</TableHead>
                    <TableHead className="w-[110px]">Estado</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {reconciliations.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                        Sin reconciliaciones registradas.
                      </TableCell>
                    </TableRow>
                  )}
                  {reconciliations.map((rec) => (
                    <TableRow key={rec.id}>
                      <TableCell className="font-mono text-xs">{rec.period}</TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {rec.statement_balance != null ? rec.statement_balance.toFixed(2) : "—"}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {rec.book_balance != null ? rec.book_balance.toFixed(2) : "—"}
                      </TableCell>
                      <TableCell>
                        <Badge variant={rec.status === "confirmed" ? "secondary" : "default"}>
                          {rec.status}
                        </Badge>
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
