"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function CalidadInspeccionesPage() {
  const { data: audits = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "audits"] as const,
    queryFn: () => api.get<{ audits: any[] }>("/v1/erp/quality/audits?page_size=50"),
    select: (d) => d.audits,
  });

  if (error) return <ErrorState message="Error cargando inspecciones" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Inspecciones de Calidad</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{audits.length} auditorías</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-20">Nro</TableHead><TableHead className="w-28">Fecha</TableHead>
              <TableHead className="w-24">Tipo</TableHead><TableHead>Alcance</TableHead>
              <TableHead className="w-28">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {audits.map((a: any) => (
                <TableRow key={a.id}>
                  <TableCell className="font-mono text-sm">{a.number}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(a.date)}</TableCell>
                  <TableCell><Badge variant="secondary">{a.audit_type}</Badge></TableCell>
                  <TableCell className="text-sm">{a.scope || "\u2014"}</TableCell>
                  <TableCell><Badge variant={a.status === "completed" ? "default" : "secondary"}>{a.status}</Badge></TableCell>
                </TableRow>
              ))}
              {audits.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin auditorías.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
