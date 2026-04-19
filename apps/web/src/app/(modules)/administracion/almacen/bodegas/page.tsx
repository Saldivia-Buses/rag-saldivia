"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { Warehouse } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function BodegasPage() {
  const { data: warehouses = [], isLoading, error } = useQuery({
    queryKey: erpKeys.warehouses2(),
    queryFn: () => api.get<{ warehouses: Warehouse[] }>("/v1/erp/stock/warehouses?active=true"),
    select: (d) => d.warehouses,
  });

  if (error)
    return <ErrorState message="Error cargando bodegas" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Bodegas / almacenes</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de bodegas activas. Cada artículo de stock tiene saldo por bodega.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Código</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead>Ubicación</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={4}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && warehouses.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin bodegas activas.
                  </TableCell>
                </TableRow>
              )}
              {warehouses.map((w) => (
                <TableRow key={w.id}>
                  <TableCell className="font-mono text-sm">{w.code}</TableCell>
                  <TableCell className="text-sm font-medium">{w.name}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{w.location || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={w.active ? "secondary" : "outline"}>
                      {w.active ? "Activa" : "Inactiva"}
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
