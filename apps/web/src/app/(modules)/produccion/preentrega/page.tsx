"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  in_production: { label: "En Producción", variant: "outline" },
  ready: { label: "Lista", variant: "default" },
  delivered: { label: "Entregada", variant: "default" },
};

export default function PreentregaPage() {
  const { data: units = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "production", "units"] as const,
    queryFn: () => api.get<{ units: any[] }>("/v1/erp/production/units?page_size=50"),
    select: (d) => d.units,
  });

  if (error) return <ErrorState message="Error cargando unidades" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Pre-entrega</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Unidades listas para entrega — {units.length} unidades</p>
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
              {units.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin unidades.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
