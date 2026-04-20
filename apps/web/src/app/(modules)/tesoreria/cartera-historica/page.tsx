"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtMoney } from "@/lib/erp/format";
import type { CheckHistoryEntry } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const checkTypeLabel: Record<number, string> = {
  1: "Portador",
  2: "A la orden",
};

export default function CarteraHistoricaPage() {
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const queryParams: Record<string, string> = {};
  if (dateFrom) queryParams.date_from = dateFrom;
  if (dateTo) queryParams.date_to = dateTo;

  const { data: history = [], isLoading, error } = useQuery({
    queryKey: erpKeys.checkHistory(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ history: CheckHistoryEntry[] }>(`/v1/erp/treasury/check-history?${qs}`);
    },
    select: (d) => d.history,
  });

  if (error)
    return <ErrorState message="Error cargando cartera histórica" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Cartera histórica de cheques</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Cheques procesados o archivados (CARCHEHI). Vista de consulta — los cheques vigentes están en
            <span className="font-medium"> Cartera activa</span> (tesorería principal).
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div className="grid gap-1.5">
            <Label htmlFor="ch-from" className="text-xs">Desde</Label>
            <Input id="ch-from" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="ch-to" className="text-xs">Hasta</Label>
            <Input id="ch-to" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
          </div>
          <div className="flex items-end">
            <Button variant="outline" onClick={() => { setDateFrom(""); setDateTo(""); }}>
              Limpiar
            </Button>
          </div>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[90px]">Operación</TableHead>
                <TableHead className="w-[100px]">Nº interno</TableHead>
                <TableHead className="w-[120px]">Nº cheque</TableHead>
                <TableHead>Banco</TableHead>
                <TableHead className="w-[130px] text-right">Importe</TableHead>
                <TableHead className="w-[90px]">Tipo</TableHead>
                <TableHead className="w-[100px]">Emitido</TableHead>
                <TableHead className="w-[110px]">Acreditado</TableHead>
                <TableHead className="w-[90px]">Prov.</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={9}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && history.length === 0 && (
                <TableRow>
                  <TableCell colSpan={9} className="h-20 text-center text-sm text-muted-foreground">
                    Sin cheques en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {history.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="text-sm">
                    {c.operation_date ? fmtDateShort(c.operation_date) : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-sm">{c.legacy_carint || "—"}</TableCell>
                  <TableCell className="font-mono text-sm">{c.number || "—"}</TableCell>
                  <TableCell className="max-w-[260px] truncate text-sm">{c.bank_name || "—"}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(c.amount)}</TableCell>
                  <TableCell className="text-xs">
                    {checkTypeLabel[c.check_type] ?? `—`}
                  </TableCell>
                  <TableCell className="text-sm">
                    {c.issue_date ? fmtDateShort(c.issue_date) : "—"}
                  </TableCell>
                  <TableCell>
                    {c.credited_at ? (
                      <Badge variant="secondary">{fmtDateShort(c.credited_at)}</Badge>
                    ) : c.returned_at ? (
                      <Badge variant="destructive">Rechazado</Badge>
                    ) : (
                      <span className="text-xs text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{c.entity_legacy_code || "—"}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
