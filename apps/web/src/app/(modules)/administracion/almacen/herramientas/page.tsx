"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { Tool } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function HerramientasPage() {
  const [articleCode, setArticleCode] = useState("");

  const queryParams: Record<string, string> = {};
  if (articleCode) queryParams.article_code = articleCode;

  const { data: tools = [], isLoading, error } = useQuery({
    queryKey: erpKeys.tools(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ tools: Tool[] }>(`/v1/erp/admin/tools?${qs}`);
    },
    select: (d) => d.tools,
  });

  if (error)
    return <ErrorState message="Error cargando herramientas" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Herramientas</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Catálogo de herramientas / items serializados (erp_tools). Código, inventario, proveedor, fecha OC.
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="grid gap-1.5">
            <Label htmlFor="h-art" className="text-xs">Código de artículo</Label>
            <Input
              id="h-art"
              placeholder="Filtrar por código de artículo…"
              value={articleCode}
              onChange={(e) => setArticleCode(e.target.value)}
            />
          </div>
          <div className="flex items-end">
            <Button variant="outline" onClick={() => setArticleCode("")}>Limpiar</Button>
          </div>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Código</TableHead>
                <TableHead className="w-[120px]">Inventario</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-[140px]">Cód. artículo</TableHead>
                <TableHead className="w-[140px]">Cód. proveedor</TableHead>
                <TableHead className="w-[120px]">Fecha OC</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow><TableCell colSpan={6}><Skeleton className="h-32 w-full" /></TableCell></TableRow>
              )}
              {!isLoading && tools.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin herramientas en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {tools.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-mono text-sm">{t.code}</TableCell>
                  <TableCell className="font-mono text-xs">{t.inventory_code || "—"}</TableCell>
                  <TableCell className="max-w-[360px] truncate text-sm">{t.name}</TableCell>
                  <TableCell className="font-mono text-xs">{t.article_code || "—"}</TableCell>
                  <TableCell className="font-mono text-xs">{t.supplier_code || "—"}</TableCell>
                  <TableCell className="text-sm">
                    {t.purchase_order_date ? fmtDateShort(t.purchase_order_date) : "—"}
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
