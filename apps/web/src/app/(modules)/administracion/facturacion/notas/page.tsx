"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { InvoiceNote } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const systemLabel: Record<string, string> = {
  "01": "Ventas",
  "02": "Compras",
  "03": "Tesorería",
};

export default function NotasPage() {
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const queryParams: Record<string, string> = {};
  if (dateFrom) queryParams.date_from = dateFrom;
  if (dateTo) queryParams.date_to = dateTo;

  const { data: notes = [], isLoading, error } = useQuery({
    queryKey: erpKeys.invoiceNotes(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ notes: InvoiceNote[] }>(`/v1/erp/invoicing/invoice-notes?${qs}`);
    },
    select: (d) => d.notes,
  });

  if (error)
    return <ErrorState message="Error cargando notas de comprobantes" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Notas de comprobantes</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Observaciones de texto libre adjuntas a movimientos (REG_MOVIMIENTO_OBS). Una fila por observación — el comprobante original vive en Facturación / Tesorería según el subsistema.
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div className="grid gap-1.5">
            <Label htmlFor="n-from" className="text-xs">Desde</Label>
            <Input id="n-from" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="n-to" className="text-xs">Hasta</Label>
            <Input id="n-to" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
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
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[90px]">Hora</TableHead>
                <TableHead className="w-[100px]">Subsistema</TableHead>
                <TableHead className="w-[120px]">Comprobante</TableHead>
                <TableHead>Observación</TableHead>
                <TableHead className="w-[110px]">Usuario</TableHead>
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
              {!isLoading && notes.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin observaciones en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {notes.map((n) => (
                <TableRow key={n.id}>
                  <TableCell className="text-sm">
                    {n.observation_date ? fmtDateShort(n.observation_date) : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {n.observation_time ? n.observation_time.slice(0, 5) : "—"}
                  </TableCell>
                  <TableCell className="text-xs">
                    <Badge variant="outline">{systemLabel[n.system_code] ?? n.system_code ?? "—"}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {n.movement_no ? `${n.movement_voucher_class}-${n.movement_no}` : "—"}
                  </TableCell>
                  <TableCell className="max-w-[520px] whitespace-pre-wrap text-sm">
                    {n.observation || "—"}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">{n.login || "—"}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
