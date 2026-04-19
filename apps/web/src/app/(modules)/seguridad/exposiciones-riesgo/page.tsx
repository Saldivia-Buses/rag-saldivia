"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { RiskExposure } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const levelVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  low: "secondary",
  medium: "default",
  high: "destructive",
};

const levelLabel: Record<string, string> = {
  low: "Bajo",
  medium: "Medio",
  high: "Alto",
};

export default function ExposicionesRiesgoPage() {
  const { data: exposures = [], isLoading, error } = useQuery({
    queryKey: erpKeys.riskExposures(),
    queryFn: () => api.get<{ risk_exposures: RiskExposure[] }>("/v1/erp/safety/risk-exposures"),
    select: (d) => d.risk_exposures,
  });

  if (error)
    return <ErrorState message="Error cargando exposiciones" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Exposiciones de riesgo</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Registro de personas expuestas a agentes de riesgo laboral — nivel, horas-día, mitigación.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Nivel</TableHead>
                <TableHead className="w-[110px] text-right">h/día</TableHead>
                <TableHead>Mitigación</TableHead>
                <TableHead className="w-[120px]">Registro</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={4}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && exposures.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin exposiciones registradas.
                  </TableCell>
                </TableRow>
              )}
              {exposures.map((e) => (
                <TableRow key={e.id}>
                  <TableCell>
                    <Badge variant={levelVariant[e.level] ?? "outline"}>
                      {levelLabel[e.level] ?? e.level}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {e.hours_per_day != null ? Number(e.hours_per_day).toFixed(1) : "—"}
                  </TableCell>
                  <TableCell className="max-w-[420px] truncate text-sm text-muted-foreground">
                    {e.mitigation || "—"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {e.created_at ? fmtDateShort(e.created_at) : "—"}
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
