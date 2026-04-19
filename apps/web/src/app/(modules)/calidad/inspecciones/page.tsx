"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtNumber } from "@/lib/erp/format";
import type { QCInspection } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  pending: { label: "Pendiente", variant: "outline" },
  passed: { label: "Aprobada", variant: "default" },
  failed: { label: "Rechazada", variant: "destructive" },
  partial: { label: "Parcial", variant: "secondary" },
};

export default function CalidadInspeccionesPage() {
  const { data: inspections = [], isLoading, error } = useQuery({
    queryKey: erpKeys.qcInspections(),
    queryFn: () =>
      api.get<{ inspections: QCInspection[] }>("/v1/erp/purchasing/inspections?page_size=100"),
    select: (d) => d.inspections,
  });

  if (error)
    return <ErrorState message="Error cargando inspecciones" onRetry={() => window.location.reload()} />;
  if (isLoading)
    return (
      <div className="flex-1 p-8">
        <Skeleton className="h-[600px]" />
      </div>
    );

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Inspecciones de Calidad</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            {inspections.length} inspecciones de recepción — control de materiales entrantes.
          </p>
        </div>
        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead className="w-32">Recepción</TableHead>
                <TableHead className="w-28">Cód. art.</TableHead>
                <TableHead>Artículo</TableHead>
                <TableHead className="w-24 text-right">Cantidad</TableHead>
                <TableHead className="w-24 text-right">Aceptada</TableHead>
                <TableHead className="w-24 text-right">Rechazada</TableHead>
                <TableHead className="w-28">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {inspections.map((i) => {
                const s = statusBadge[i.status] ?? { label: i.status, variant: "outline" as const };
                return (
                  <TableRow key={i.id}>
                    <TableCell className="text-sm text-muted-foreground">
                      <Link href={`/calidad/inspecciones/${i.id}`} className="hover:underline">
                        {fmtDateShort(i.created_at)}
                      </Link>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{i.receipt_number ?? "—"}</TableCell>
                    <TableCell className="font-mono text-xs">{i.article_code ?? "—"}</TableCell>
                    <TableCell className="text-sm">{i.article_name ?? "—"}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtNumber(i.quantity)}</TableCell>
                    <TableCell className="text-right font-mono text-sm text-emerald-600">
                      {fmtNumber(i.accepted_qty)}
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm text-destructive">
                      {fmtNumber(i.rejected_qty)}
                    </TableCell>
                    <TableCell>
                      <Badge variant={s.variant}>{s.label}</Badge>
                    </TableCell>
                  </TableRow>
                );
              })}
              {inspections.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8} className="h-24 text-center text-muted-foreground">
                    Sin inspecciones registradas.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
