"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusColor: Record<string, "default" | "secondary" | "outline"> = { draft: "secondary", approved: "default", obsolete: "secondary" };

export default function CalidadSGCPage() {
  const { data: docs = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "documents"] as const,
    queryFn: () => api.get<{ documents: any[] }>("/v1/erp/quality/documents?page_size=50"),
    select: (d) => d.documents,
  });

  if (error) return <ErrorState message="Error cargando documentos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Sistema de Gestión de Calidad</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Documentos controlados — {docs.length} documentos</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-24">Código</TableHead><TableHead>Título</TableHead>
              <TableHead className="w-16 text-center">Rev.</TableHead><TableHead className="w-28">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {docs.map((d: any) => (
                <TableRow key={d.id}>
                  <TableCell className="font-mono text-sm">{d.code}</TableCell>
                  <TableCell className="text-sm">{d.title}</TableCell>
                  <TableCell className="text-center text-sm">{d.revision}</TableCell>
                  <TableCell><Badge variant={statusColor[d.status] || "secondary"}>{d.status}</Badge></TableCell>
                </TableRow>
              ))}
              {docs.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin documentos.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
