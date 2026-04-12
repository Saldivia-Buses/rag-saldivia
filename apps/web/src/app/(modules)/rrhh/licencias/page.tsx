"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const eventLabel: Record<string, string> = { absence: "Falta", leave: "Licencia", accident: "Accidente", vacation: "Vacaciones", overtime: "Hora extra" };

export default function LicenciasPage() {
  const { data: events = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "hr", "events"] as const,
    queryFn: () => api.get<{ events: any[] }>("/v1/erp/hr/events?page_size=50"),
    select: (d) => d.events,
  });

  if (error) return <ErrorState message="Error cargando licencias" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Licencias y Novedades</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{events.length} novedades</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead>Tipo</TableHead><TableHead>Desde</TableHead>
              <TableHead>Hasta</TableHead><TableHead>Horas</TableHead><TableHead>Notas</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {events.map((ev: any) => (
                <TableRow key={ev.id}>
                  <TableCell><Badge variant="outline">{eventLabel[ev.event_type] || ev.event_type}</Badge></TableCell>
                  <TableCell className="text-sm">{fmtDateShort(ev.date_from)}</TableCell>
                  <TableCell className="text-sm">{fmtDateShort(ev.date_to)}</TableCell>
                  <TableCell className="text-sm font-mono">{ev.hours || "\u2014"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground truncate max-w-48">{ev.notes || "\u2014"}</TableCell>
                </TableRow>
              ))}
              {events.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin novedades.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
