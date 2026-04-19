"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { SupplierScorecard } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

function scoreVariant(score: number | null): "default" | "secondary" | "destructive" | "outline" {
  if (score == null) return "outline";
  if (score >= 80) return "secondary";
  if (score >= 60) return "default";
  return "destructive";
}

function scoreLabel(score: number | null): string {
  if (score == null) return "—";
  return Number(score).toFixed(1);
}

export default function SupplierScorecardsPage() {
  const { data: cards = [], isLoading, error } = useQuery({
    queryKey: erpKeys.supplierScorecards(),
    queryFn: () => api.get<{ scorecards: SupplierScorecard[] }>("/v1/erp/quality/supplier-scorecards?page_size=100"),
    select: (d) => d.scorecards,
  });

  if (error)
    return <ErrorState message="Error cargando scorecards de proveedores" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Scorecards de proveedores</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Calificación de calidad agregada por (proveedor, período) — total recibos, aceptados, rechazados, demeritos, quality score.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Período</TableHead>
                <TableHead>Proveedor</TableHead>
                <TableHead className="w-[110px] text-right">Recibos</TableHead>
                <TableHead className="w-[110px] text-right">Aceptado</TableHead>
                <TableHead className="w-[110px] text-right">Rechazado</TableHead>
                <TableHead className="w-[110px] text-right">Demeritos</TableHead>
                <TableHead className="w-[130px] text-right">Quality score</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={7}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && cards.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="h-20 text-center text-sm text-muted-foreground">
                    Sin scorecards registrados.
                  </TableCell>
                </TableRow>
              )}
              {cards.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="font-mono text-sm">{c.period}</TableCell>
                  <TableCell className="max-w-[320px] truncate text-sm">{c.supplier_name}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{c.total_receipts}</TableCell>
                  <TableCell className="text-right font-mono text-sm text-muted-foreground">
                    {c.accepted_qty != null ? Number(c.accepted_qty).toFixed(0) : "—"}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {c.rejected_qty != null ? Number(c.rejected_qty).toFixed(0) : "—"}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">{c.total_demerits}</TableCell>
                  <TableCell className="text-right">
                    <Badge variant={scoreVariant(c.quality_score != null ? Number(c.quality_score) : null)}>
                      {scoreLabel(c.quality_score != null ? Number(c.quality_score) : null)}
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
