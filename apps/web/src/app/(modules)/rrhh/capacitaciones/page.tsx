"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function CapacitacionesPage() {
  const { data: training = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "hr", "training"] as const,
    queryFn: () => api.get<{ training: any[] }>("/v1/erp/hr/training?page_size=50"),
    select: (d) => d.training,
  });

  if (error) return <ErrorState message="Error cargando capacitaciones" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Capacitaciones</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{training.length} cursos</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead>Curso</TableHead><TableHead>Instructor</TableHead>
              <TableHead>Fecha</TableHead><TableHead>Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {training.map((t: any) => (
                <TableRow key={t.id}>
                  <TableCell className="text-sm">{t.name}</TableCell>
                  <TableCell className="text-sm">{t.instructor || "\u2014"}</TableCell>
                  <TableCell className="text-sm">{fmtDateShort(t.date_from)}</TableCell>
                  <TableCell><Badge variant="secondary">{t.status}</Badge></TableCell>
                </TableRow>
              ))}
              {training.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin cursos.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
