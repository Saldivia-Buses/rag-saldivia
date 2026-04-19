"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { ProductionCenter } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function CentrosProduccionPage() {
  const { data: centers = [], isLoading, error } = useQuery({
    queryKey: erpKeys.productionCenters(),
    queryFn: () => api.get<{ centers: ProductionCenter[] }>("/v1/erp/production/centers"),
    select: (d) => d.centers,
  });

  if (error)
    return <ErrorState message="Error cargando centros de producción" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Centros de producción</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de talleres / centros donde se ejecutan las órdenes de producción.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Código</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={3}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && centers.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                    Sin centros de producción.
                  </TableCell>
                </TableRow>
              )}
              {centers.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="font-mono text-sm">{c.code}</TableCell>
                  <TableCell className="text-sm font-medium">{c.name}</TableCell>
                  <TableCell>
                    <Badge variant={c.active ? "secondary" : "outline"}>
                      {c.active ? "Activo" : "Inactivo"}
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
