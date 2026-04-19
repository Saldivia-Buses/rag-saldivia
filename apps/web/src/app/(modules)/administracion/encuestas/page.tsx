"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { Survey } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  draft: "outline",
  open: "default",
  closed: "secondary",
};

const statusLabel: Record<string, string> = {
  draft: "Borrador",
  open: "Abierta",
  closed: "Cerrada",
};

export default function EncuestasPage() {
  const { data: surveys = [], isLoading, error } = useQuery({
    queryKey: erpKeys.surveys(),
    queryFn: () => api.get<{ surveys: Survey[] }>("/v1/erp/admin/surveys?page_size=100"),
    select: (d) => d.surveys,
  });

  if (error)
    return <ErrorState message="Error cargando encuestas" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Encuestas</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Encuestas internas lanzadas al personal — sucesor del motor de cuestionarios de Histrix.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Fecha</TableHead>
                <TableHead>Título</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={4}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && surveys.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin encuestas registradas.
                  </TableCell>
                </TableRow>
              )}
              {surveys.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="text-sm">
                    {s.created_at ? fmtDateShort(s.created_at) : "—"}
                  </TableCell>
                  <TableCell className="text-sm font-medium">{s.title}</TableCell>
                  <TableCell className="max-w-[420px] truncate text-xs text-muted-foreground">{s.description || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[s.status] ?? "outline"}>
                      {statusLabel[s.status] ?? s.status}
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
