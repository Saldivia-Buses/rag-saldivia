"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { PriceList } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function ListasPreciosPage() {
  const { data: lists = [], isLoading, error } = useQuery({
    queryKey: erpKeys.priceLists(),
    queryFn: () => api.get<{ price_lists: PriceList[] }>("/v1/erp/sales/price-lists"),
    select: (d) => d.price_lists,
  });

  if (error)
    return <ErrorState message="Error cargando listas de precios" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Listas de precios</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de listas de precios vigentes y sus períodos de validez.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[130px]">Desde</TableHead>
                <TableHead className="w-[130px]">Hasta</TableHead>
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
              {!isLoading && lists.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin listas de precios.
                  </TableCell>
                </TableRow>
              )}
              {lists.map((l) => (
                <TableRow key={l.id}>
                  <TableCell className="text-sm font-medium">{l.name}</TableCell>
                  <TableCell className="text-sm">
                    {l.valid_from ? fmtDateShort(l.valid_from) : "—"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {l.valid_until ? fmtDateShort(l.valid_until) : "—"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={l.active ? "secondary" : "outline"}>
                      {l.active ? "Activa" : "Inactiva"}
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
