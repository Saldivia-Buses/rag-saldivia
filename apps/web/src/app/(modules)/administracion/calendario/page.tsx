"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { CalendarEvent } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function CalendarioPage() {
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const queryParams: Record<string, string> = {};
  if (dateFrom) queryParams.date_from = dateFrom;
  if (dateTo) queryParams.date_to = dateTo;

  const { data: events = [], isLoading, error } = useQuery({
    queryKey: erpKeys.calendar(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams(queryParams).toString();
      const url = qs ? `/v1/erp/admin/calendar?${qs}` : "/v1/erp/admin/calendar";
      return api.get<{ events: CalendarEvent[] }>(url);
    },
    select: (d) => d.events,
  });

  if (error)
    return <ErrorState message="Error cargando calendario" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Calendario</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Eventos y recordatorios internos — sucesor de HTXCALENDAR de Histrix.
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div className="grid gap-1.5">
            <Label htmlFor="c-from" className="text-xs">Desde</Label>
            <Input id="c-from" type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="c-to" className="text-xs">Hasta</Label>
            <Input id="c-to" type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
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
                <TableHead className="w-[140px]">Inicio</TableHead>
                <TableHead className="w-[140px]">Fin</TableHead>
                <TableHead>Título</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead className="w-[90px]">Todo-día</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={5}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && events.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin eventos en este rango.
                  </TableCell>
                </TableRow>
              )}
              {events.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="text-sm">{e.start_at ? fmtDateShort(e.start_at) : "—"}</TableCell>
                  <TableCell className="text-sm">{e.end_at ? fmtDateShort(e.end_at) : "—"}</TableCell>
                  <TableCell className="text-sm font-medium">{e.title}</TableCell>
                  <TableCell className="max-w-[420px] truncate text-xs text-muted-foreground">{e.description || "—"}</TableCell>
                  <TableCell>
                    {e.all_day ? <Badge variant="secondary">Sí</Badge> : <span className="text-xs text-muted-foreground">—</span>}
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
