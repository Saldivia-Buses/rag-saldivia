"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { BankAccount } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function CuentasBancariasPage() {
  const { data: accounts = [], isLoading, error } = useQuery({
    queryKey: erpKeys.bankAccountsCatalog(),
    queryFn: () => api.get<{ bank_accounts: BankAccount[] }>("/v1/erp/treasury/bank-accounts"),
    select: (d) => d.bank_accounts,
  });

  if (error)
    return <ErrorState message="Error cargando cuentas bancarias" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Cuentas bancarias</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de cuentas bancarias del tenant — banco, sucursal, número, CBU, alias.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Banco</TableHead>
                <TableHead className="w-[140px]">Sucursal</TableHead>
                <TableHead className="w-[180px]">Cuenta</TableHead>
                <TableHead className="w-[220px]">CBU</TableHead>
                <TableHead className="w-[160px]">Alias</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
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
              {!isLoading && accounts.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin cuentas bancarias.
                  </TableCell>
                </TableRow>
              )}
              {accounts.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="text-sm font-medium">{a.bank_name}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{a.branch || "—"}</TableCell>
                  <TableCell className="font-mono text-sm">{a.account_number}</TableCell>
                  <TableCell className="font-mono text-xs">{a.cbu || "—"}</TableCell>
                  <TableCell className="text-xs">{a.alias || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={a.active ? "secondary" : "outline"}>
                      {a.active ? "Activa" : "Baja"}
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
