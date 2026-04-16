"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  in_production: { label: "En Producción", variant: "outline" },
  ready: { label: "Lista para entrega", variant: "default" },
  delivered: { label: "Entregada", variant: "secondary" },
};

export default function SeguimientoPage() {
  const { data: units = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "production", "units"] as const,
    queryFn: () => api.get<{ units: any[] }>("/v1/erp/production/units?page_size=100"),
    select: (d) => d.units.filter((u: any) => u.status !== "delivered"),
  });

  if (error) return <ErrorState message="Error cargando seguimiento" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const inProd = units.filter((u: any) => u.status === "in_production").length;
  const ready = units.filter((u: any) => u.status === "ready").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Seguimiento de Producción</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {inProd} en producción · {ready} listas para entrega
          </p>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-6 sm:grid-cols-3">
          <div className="rounded-xl border border-border/40 bg-card px-5 py-4">
            <p className="text-xs text-muted-foreground mb-1">En producción</p>
            <p className="text-2xl font-semibold tabular-nums">{inProd}</p>
          </div>
          <div className="rounded-xl border border-border/40 bg-card px-5 py-4">
            <p className="text-xs text-muted-foreground mb-1">Listas para entrega</p>
            <p className="text-2xl font-semibold tabular-nums">{ready}</p>
          </div>
          <div className="rounded-xl border border-border/40 bg-card px-5 py-4">
            <p className="text-xs text-muted-foreground mb-1">Total activas</p>
            <p className="text-2xl font-semibold tabular-nums">{units.length}</p>
          </div>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-36">Chasis</TableHead>
              <TableHead className="w-20">Interno</TableHead>
              <TableHead>Modelo</TableHead>
              <TableHead className="w-28">Patente</TableHead>
              <TableHead className="w-32">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {units.map((u: any) => {
                const s = statusBadge[u.status] || statusBadge.in_production;
                return (
                  <TableRow key={u.id}>
                    <TableCell className="font-mono text-sm">{u.chassis_number}</TableCell>
                    <TableCell className="text-sm">{u.internal_number || "\u2014"}</TableCell>
                    <TableCell className="text-sm">{u.model || "\u2014"}</TableCell>
                    <TableCell className="font-mono text-sm">{u.patent || "\u2014"}</TableCell>
                    <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                  </TableRow>
                );
              })}
              {units.length === 0 && (
                <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                  No hay unidades activas en producción.
                </TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
