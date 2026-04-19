"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { ProductAttribute } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";

type Tab = "active" | "all";

export default function ProductAttributesPage() {
  const [tab, setTab] = useState<Tab>("active");
  const activeOnly = tab === "active";

  const { data: attrs = [], isLoading, error } = useQuery({
    queryKey: erpKeys.productAttributes(activeOnly),
    queryFn: () =>
      api.get<{ attributes: ProductAttribute[] }>(
        `/v1/erp/admin/product-attributes?active=${activeOnly ? "true" : "false"}`,
      ),
    select: (d) => d.attributes,
  });

  if (error)
    return <ErrorState message="Error cargando atributos de producto" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Atributos de producto</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Diccionario de atributos aplicables a productos (tipo, sección, código de artículo).
          </p>
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as Tab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="active">Activos</TabsTrigger>
            <TabsTrigger value="all">Todos</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[90px] text-right">Orden</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[130px]">Tipo</TableHead>
                <TableHead className="w-[140px]">Cód. artículo</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={5}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && attrs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                    Sin atributos en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {attrs.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="text-right font-mono text-sm">{a.sort_order}</TableCell>
                  <TableCell className="text-sm font-medium">{a.name}</TableCell>
                  <TableCell className="text-xs text-muted-foreground capitalize">{a.attribute_type || "—"}</TableCell>
                  <TableCell className="font-mono text-xs">{a.article_code || "—"}</TableCell>
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
