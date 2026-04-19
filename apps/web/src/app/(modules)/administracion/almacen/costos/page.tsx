"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtMoney } from "@/lib/erp/format";
import type { ArticleCost } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function CostosArticulosPage() {
  const [supplierCode, setSupplierCode] = useState("");
  const [articleId, setArticleId] = useState("");

  const queryParams: Record<string, string> = {};
  if (supplierCode) queryParams.supplier_code = supplierCode;
  if (articleId) queryParams.article_id = articleId;

  const { data: costs = [], isLoading, error } = useQuery({
    queryKey: erpKeys.articleCosts(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ costs: ArticleCost[] }>(`/v1/erp/stock/article-costs?${qs}`);
    },
    select: (d) => d.costs,
  });

  if (error)
    return <ErrorState message="Error cargando costos de artículos" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Costos de artículos</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Ledger de costos por (artículo, proveedor) con última fecha de actualización (STKINSPR → erp_article_costs, 189,863 filas live).
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div className="grid gap-1.5">
            <Label htmlFor="c-sup" className="text-xs">Código de proveedor</Label>
            <Input
              id="c-sup"
              placeholder="Filtrar por cod. proveedor…"
              value={supplierCode}
              onChange={(e) => setSupplierCode(e.target.value)}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="c-art" className="text-xs">Artículo (UUID)</Label>
            <Input
              id="c-art"
              placeholder="Filtrar por artículo UUID…"
              value={articleId}
              onChange={(e) => setArticleId(e.target.value)}
            />
          </div>
          <div className="flex items-end">
            <Button variant="outline" onClick={() => { setSupplierCode(""); setArticleId(""); }}>
              Limpiar
            </Button>
          </div>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Cód. art.</TableHead>
                <TableHead>Artículo</TableHead>
                <TableHead className="w-[120px]">Cód. prov.</TableHead>
                <TableHead>Proveedor</TableHead>
                <TableHead className="w-[130px] text-right">Costo</TableHead>
                <TableHead className="w-[110px]">Últ. act.</TableHead>
                <TableHead className="w-[110px]">Factura</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={7}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && costs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="h-20 text-center text-sm text-muted-foreground">
                    Sin costos en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {costs.map((c) => (
                <TableRow key={c.id}>
                  <TableCell className="font-mono text-xs">{c.article_code || "—"}</TableCell>
                  <TableCell className="max-w-[260px] truncate text-sm">
                    {c.article_name ?? <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{c.supplier_code || "—"}</TableCell>
                  <TableCell className="max-w-[240px] truncate text-sm">
                    {c.supplier_name ?? <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(c.cost)}</TableCell>
                  <TableCell className="text-sm">
                    {c.last_update_date ? fmtDateShort(c.last_update_date) : "—"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {c.invoice_date ? fmtDateShort(c.invoice_date) : "—"}
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
