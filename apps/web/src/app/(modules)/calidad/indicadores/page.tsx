"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { QualityIndicator } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

function thisYearPeriod(): { from: string; to: string } {
  const y = new Date().getFullYear();
  return { from: `${y}-01`, to: `${y}-12` };
}

export default function IndicadoresPage() {
  const initial = thisYearPeriod();
  const [periodFrom, setPeriodFrom] = useState(initial.from);
  const [periodTo, setPeriodTo] = useState(initial.to);

  const queryParams = { period_from: periodFrom, period_to: periodTo };

  const { data: indicators = [], isLoading, error } = useQuery({
    queryKey: erpKeys.qualityIndicators(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams(queryParams).toString();
      return api.get<{ indicators: QualityIndicator[] }>(`/v1/erp/quality/indicators?${qs}`);
    },
    select: (d) => d.indicators,
  });

  const grouped = useMemo(() => {
    const m = new Map<string, QualityIndicator[]>();
    for (const ind of indicators) {
      const arr = m.get(ind.period) ?? [];
      arr.push(ind);
      m.set(ind.period, arr);
    }
    return [...m.entries()].sort((a, b) => b[0].localeCompare(a[0]));
  }, [indicators]);

  if (error)
    return <ErrorState message="Error cargando indicadores" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Indicadores de calidad</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            KPIs por período (valor vs objetivo). Formato de período: AAAA-MM.
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="grid gap-1.5">
            <Label htmlFor="i-from" className="text-xs">Período desde</Label>
            <Input
              id="i-from"
              placeholder="2025-01"
              value={periodFrom}
              onChange={(e) => setPeriodFrom(e.target.value)}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="i-to" className="text-xs">Período hasta</Label>
            <Input
              id="i-to"
              placeholder="2025-12"
              value={periodTo}
              onChange={(e) => setPeriodTo(e.target.value)}
            />
          </div>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Período</TableHead>
                <TableHead>Indicador</TableHead>
                <TableHead className="w-[140px] text-right">Valor</TableHead>
                <TableHead className="w-[140px] text-right">Objetivo</TableHead>
                <TableHead className="w-[120px]">Cumplimiento</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={5}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && indicators.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin indicadores en este rango.
                  </TableCell>
                </TableRow>
              )}
              {grouped.flatMap(([period, rows]) =>
                rows.map((ind) => {
                  const hit =
                    ind.value != null && ind.target != null
                      ? Number(ind.value) >= Number(ind.target)
                      : null;
                  return (
                    <TableRow key={ind.id}>
                      <TableCell className="font-mono text-sm">{period}</TableCell>
                      <TableCell className="text-sm">{ind.indicator_type}</TableCell>
                      <TableCell className="text-right font-mono text-sm">
                        {ind.value != null ? Number(ind.value).toFixed(2) : "—"}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm text-muted-foreground">
                        {ind.target != null ? Number(ind.target).toFixed(2) : "—"}
                      </TableCell>
                      <TableCell>
                        {hit == null ? (
                          <span className="text-xs text-muted-foreground">—</span>
                        ) : hit ? (
                          <Badge variant="secondary">OK</Badge>
                        ) : (
                          <Badge variant="destructive">Bajo</Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                }),
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
