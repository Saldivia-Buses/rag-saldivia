"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { ProductSection } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function ProductSectionsPage() {
  const { data: sections = [], isLoading, error } = useQuery({
    queryKey: erpKeys.productSections(),
    queryFn: () => api.get<{ sections: ProductSection[] }>("/v1/erp/admin/product-sections"),
    select: (d) => d.sections,
  });

  if (error)
    return <ErrorState message="Error cargando secciones de producto" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Secciones de producto</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Agrupaciones de atributos / características de producto usadas para la ficha de ingeniería.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px] text-right">Orden</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[110px]">Rubro</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={4}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && sections.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin secciones definidas.
                  </TableCell>
                </TableRow>
              )}
              {sections.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="text-right font-mono text-sm">{s.sort_order}</TableCell>
                  <TableCell className="text-sm font-medium">{s.name}</TableCell>
                  <TableCell className="font-mono text-xs">{s.rubro_id || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={s.active ? "secondary" : "outline"}>
                      {s.active ? "Activa" : "Inactiva"}
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
