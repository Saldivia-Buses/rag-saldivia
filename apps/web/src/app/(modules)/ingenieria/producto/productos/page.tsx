"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { Product } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function ProductsCatalogPage() {
  const { data: products = [], isLoading, error } = useQuery({
    queryKey: erpKeys.products(),
    queryFn: () => api.get<{ products: Product[] }>("/v1/erp/admin/products?page_size=100"),
    select: (d) => d.products,
  });

  if (error)
    return <ErrorState message="Error cargando catálogo de productos" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Catálogo de productos</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Productos de ingeniería con código de proveedor — distinto de `erp_articles` (stock SKUs).
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Legacy ID</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead className="w-[140px]">Cód. proveedor</TableHead>
                <TableHead className="w-[130px]">Creado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={4}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && products.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin productos registrados.
                  </TableCell>
                </TableRow>
              )}
              {products.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-mono text-xs">{p.legacy_id}</TableCell>
                  <TableCell className="text-sm">{p.description}</TableCell>
                  <TableCell className="font-mono text-xs">{p.supplier_code || "—"}</TableCell>
                  <TableCell className="text-sm">
                    {p.created_at ? fmtDateShort(p.created_at) : "—"}
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
