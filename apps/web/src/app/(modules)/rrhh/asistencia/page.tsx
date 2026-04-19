"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { Attendance } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function AsistenciaPage() {
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const queryParams: Record<string, string> = {};
  if (dateFrom) queryParams.date_from = dateFrom;
  if (dateTo) queryParams.date_to = dateTo;

  const { data: rows = [], isLoading, error } = useQuery({
    queryKey: erpKeys.attendance(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ attendance: Attendance[] }>(`/v1/erp/hr/attendance?${qs}`);
    },
    select: (d) => d.attendance,
  });

  if (error)
    return <ErrorState message="Error cargando asistencia" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Asistencia</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Registros de entrada/salida por empleado con horas trabajadas.
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div className="grid gap-1.5">
            <Label htmlFor="a-from" className="text-xs">Desde</Label>
            <Input id="a-from" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="a-to" className="text-xs">Hasta</Label>
            <Input id="a-to" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
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
                <TableHead className="w-[120px]">Fecha</TableHead>
                <TableHead className="w-[120px]">Entrada</TableHead>
                <TableHead className="w-[120px]">Salida</TableHead>
                <TableHead className="w-[110px] text-right">Horas</TableHead>
                <TableHead className="w-[130px]">Origen</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={5}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && rows.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin registros en este rango.
                  </TableCell>
                </TableRow>
              )}
              {rows.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="text-sm">
                    {a.date ? fmtDateShort(a.date) : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {a.clock_in ? new Date(a.clock_in).toLocaleTimeString("es-AR", { hour: "2-digit", minute: "2-digit" }) : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {a.clock_out ? new Date(a.clock_out).toLocaleTimeString("es-AR", { hour: "2-digit", minute: "2-digit" }) : "—"}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {a.hours != null ? Number(a.hours).toFixed(2) : "—"}
                  </TableCell>
                  <TableCell className="text-xs">
                    <Badge variant="outline">{a.source || "—"}</Badge>
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
