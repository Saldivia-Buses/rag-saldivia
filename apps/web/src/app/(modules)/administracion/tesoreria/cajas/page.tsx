"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { CashRegister } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function CajasPage() {
  const { data: registers = [], isLoading, error } = useQuery({
    queryKey: erpKeys.cashRegisters(),
    queryFn: () => api.get<{ cash_registers: CashRegister[] }>("/v1/erp/treasury/cash-registers"),
    select: (d) => d.cash_registers,
  });

  if (error)
    return <ErrorState message="Error cargando cajas" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Cajas</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de cajas de efectivo — cada una con su cuenta contable asociada.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={2}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && registers.length === 0 && (
                <TableRow>
                  <TableCell colSpan={2} className="h-20 text-center text-sm text-muted-foreground">
                    Sin cajas registradas.
                  </TableCell>
                </TableRow>
              )}
              {registers.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="text-sm font-medium">
                    <Link href={`/administracion/tesoreria/cajas/${r.id}`} className="hover:underline">
                      {r.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant={r.active ? "secondary" : "outline"}>
                      {r.active ? "Activa" : "Inactiva"}
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
