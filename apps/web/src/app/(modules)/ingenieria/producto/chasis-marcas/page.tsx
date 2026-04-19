"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { ChassisBrand } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function ChassisBrandsPage() {
  const { data: brands = [], isLoading, error } = useQuery({
    queryKey: erpKeys.chassisBrands(),
    queryFn: () => api.get<{ chassis_brands: ChassisBrand[] }>("/v1/erp/manufacturing/chassis-brands"),
    select: (d) => d.chassis_brands,
  });

  if (error)
    return <ErrorState message="Error cargando marcas de chasis" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Marcas de chasis</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">Catálogo de marcas de chasis utilizadas en manufactura.</p>
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
                <TableRow><TableCell colSpan={3}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && brands.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                    Sin marcas registradas.
                  </TableCell>
                </TableRow>
              )}
              {brands.map((b) => (
                <TableRow key={b.id}>
                  <TableCell className="font-mono text-sm">{b.code}</TableCell>
                  <TableCell className="text-sm font-medium">{b.name}</TableCell>
                  <TableCell>
                    <Badge variant={b.active ? "secondary" : "outline"}>
                      {b.active ? "Activa" : "Inactiva"}
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
