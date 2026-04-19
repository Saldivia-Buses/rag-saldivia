"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { QualityAudit } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  planned: "outline",
  in_progress: "default",
  completed: "secondary",
  cancelled: "destructive",
};

const statusLabel: Record<string, string> = {
  planned: "Planificada",
  in_progress: "En curso",
  completed: "Completada",
  cancelled: "Cancelada",
};

export default function AuditoriasPage() {
  const { data: audits = [], isLoading, error } = useQuery({
    queryKey: erpKeys.qualityAudits(),
    queryFn: () => api.get<{ audits: QualityAudit[] }>("/v1/erp/quality/audits?page_size=100"),
    select: (d) => d.audits,
  });

  if (error)
    return <ErrorState message="Error cargando auditorías" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Auditorías de calidad</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Auditorías internas y externas. Registro de alcance, responsable, puntaje y estado.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Número</TableHead>
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[130px]">Tipo</TableHead>
                <TableHead>Alcance</TableHead>
                <TableHead className="w-[120px] text-right">Puntaje</TableHead>
                <TableHead className="w-[130px]">Estado</TableHead>
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
              {!isLoading && audits.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin auditorías registradas.
                  </TableCell>
                </TableRow>
              )}
              {audits.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="font-mono text-sm">{a.number || "—"}</TableCell>
                  <TableCell className="text-sm">{a.date ? fmtDateShort(a.date) : "—"}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{a.audit_type || "—"}</TableCell>
                  <TableCell className="max-w-[400px] truncate text-sm">{a.scope || "—"}</TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {a.score != null ? Number(a.score).toFixed(1) : "—"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[a.status] ?? "outline"}>
                      {statusLabel[a.status] ?? a.status}
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
