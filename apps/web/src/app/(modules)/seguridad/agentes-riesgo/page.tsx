"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { RiskAgent } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function AgentesRiesgoPage() {
  const { data: agents = [], isLoading, error } = useQuery({
    queryKey: erpKeys.riskAgents(),
    queryFn: () => api.get<{ risk_agents: RiskAgent[] }>("/v1/erp/safety/risk-agents"),
    select: (d) => d.risk_agents,
  });

  if (error)
    return <ErrorState message="Error cargando agentes de riesgo" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Agentes de riesgo</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de agentes laborales (químicos, físicos, biológicos, ergonómicos, psicosociales).
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[160px]">Tipo</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={3}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && agents.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                    Sin agentes registrados.
                  </TableCell>
                </TableRow>
              )}
              {agents.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="text-sm font-medium">{a.name}</TableCell>
                  <TableCell className="text-xs text-muted-foreground capitalize">{a.risk_type || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={a.active ? "secondary" : "outline"}>
                      {a.active ? "Activo" : "Inactivo"}
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
