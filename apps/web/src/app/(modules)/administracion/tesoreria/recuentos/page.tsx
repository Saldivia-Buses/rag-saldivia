"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtMoney } from "@/lib/erp/format";
import type { CashCount } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function RecuentosCajaPage() {
  const { data: counts = [], isLoading, error } = useQuery({
    queryKey: erpKeys.cashCounts(),
    queryFn: () => api.get<{ cash_counts: CashCount[] }>("/v1/erp/treasury/cash-counts?page_size=100"),
    select: (d) => d.cash_counts,
  });

  if (error)
    return <ErrorState message="Error cargando recuentos de caja" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Recuentos de caja</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Cierres de caja — saldo esperado vs contado, con diferencia por arqueo.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[140px] text-right">Esperado</TableHead>
                <TableHead className="w-[140px] text-right">Contado</TableHead>
                <TableHead className="w-[140px] text-right">Diferencia</TableHead>
                <TableHead>Notas</TableHead>
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
              {!isLoading && counts.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin recuentos de caja registrados.
                  </TableCell>
                </TableRow>
              )}
              {counts.map((c) => {
                const diff = c.difference != null ? Number(c.difference) : null;
                const balanced = diff != null && Math.abs(diff) < 0.01;
                return (
                  <TableRow key={c.id}>
                    <TableCell className="text-sm">
                      {c.date ? fmtDateShort(c.date) : "—"}
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtMoney(c.expected)}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtMoney(c.counted)}</TableCell>
                    <TableCell className="text-right font-mono text-sm">
                      {fmtMoney(c.difference)}
                    </TableCell>
                    <TableCell className="max-w-[340px] truncate text-xs text-muted-foreground">
                      {c.notes || "—"}
                    </TableCell>
                    <TableCell>
                      {balanced ? (
                        <Badge variant="secondary">Cuadra</Badge>
                      ) : diff != null ? (
                        <Badge variant="destructive">Descuadre</Badge>
                      ) : (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
