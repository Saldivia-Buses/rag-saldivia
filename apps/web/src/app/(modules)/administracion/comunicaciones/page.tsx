"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { Communication } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const priorityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  low: "outline",
  normal: "secondary",
  high: "default",
  urgent: "destructive",
};

const priorityLabel: Record<string, string> = {
  low: "Baja",
  normal: "Normal",
  high: "Alta",
  urgent: "Urgente",
};

export default function ComunicacionesPage() {
  const { data: comms = [], isLoading, error } = useQuery({
    queryKey: erpKeys.communications(),
    queryFn: () => api.get<{ communications: Communication[] }>("/v1/erp/admin/communications?page_size=100"),
    select: (d) => d.communications,
  });

  if (error)
    return <ErrorState message="Error cargando comunicaciones" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Comunicaciones internas</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Mensajes / anuncios publicados por la empresa — reemplaza HTXMAIL + HTX_MEDIA de Histrix (ADR 027 Phase 3 ingest agents van a leer esto).
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Fecha</TableHead>
                <TableHead>Asunto</TableHead>
                <TableHead className="w-[110px]">Prioridad</TableHead>
                <TableHead className="w-[140px]">Remitente</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={4}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && comms.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin comunicaciones.
                  </TableCell>
                </TableRow>
              )}
              {comms.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="text-sm">
                    {c.created_at ? fmtDateShort(c.created_at) : "—"}
                  </TableCell>
                  <TableCell className="max-w-[520px] truncate text-sm font-medium">{c.subject}</TableCell>
                  <TableCell>
                    <Badge variant={priorityVariant[c.priority] ?? "outline"}>
                      {priorityLabel[c.priority] ?? c.priority}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">{c.sender_id || "—"}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
