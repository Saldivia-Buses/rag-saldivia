"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { StockMovement } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";

type TypeTab = "all" | "in" | "out" | "transfer";

const typeToFilter: Record<TypeTab, string> = {
  all: "",
  in: "in",
  out: "out",
  transfer: "transfer",
};

const movementLabel: Record<string, string> = {
  in: "Ingreso",
  out: "Egreso",
  transfer: "Transf.",
  adjustment: "Ajuste",
  inventory: "Inventario",
};

const movementVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  in: "secondary",
  out: "destructive",
  transfer: "outline",
  adjustment: "default",
  inventory: "outline",
};

export default function MovimientosStockPage() {
  const [tab, setTab] = useState<TypeTab>("all");
  const [articleCode, setArticleCode] = useState("");

  const queryParams: Record<string, string> = {};
  if (articleCode) queryParams.article_code = articleCode;

  const { data: movements = [], isLoading, error } = useQuery({
    queryKey: erpKeys.stockMovements({ ...queryParams, tab }),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ movements: StockMovement[] }>(`/v1/erp/stock/movements?${qs}`);
    },
    select: (d) => {
      if (tab === "all") return d.movements;
      return d.movements.filter((m) => m.movement_type === typeToFilter[tab]);
    },
  });

  if (error)
    return <ErrorState message="Error cargando movimientos de stock" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Movimientos de stock</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Historial de ingresos, egresos, transferencias y ajustes por artículo (erp_stock_movements).
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="grid gap-1.5">
            <Label htmlFor="m-art" className="text-xs">Código de artículo (opcional)</Label>
            <Input
              id="m-art"
              placeholder="Filtrar por código de artículo…"
              value={articleCode}
              onChange={(e) => setArticleCode(e.target.value)}
            />
          </div>
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as TypeTab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="all">Todos</TabsTrigger>
            <TabsTrigger value="in">Ingresos</TabsTrigger>
            <TabsTrigger value="out">Egresos</TabsTrigger>
            <TabsTrigger value="transfer">Transferencias</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[120px]">Cód. art.</TableHead>
                <TableHead>Artículo</TableHead>
                <TableHead className="w-[110px]">Tipo</TableHead>
                <TableHead className="w-[110px] text-right">Cantidad</TableHead>
                <TableHead className="w-[130px] text-right">Costo unit.</TableHead>
                <TableHead>Notas</TableHead>
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
              {!isLoading && movements.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="h-20 text-center text-sm text-muted-foreground">
                    Sin movimientos en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {movements.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="text-sm">
                    {m.created_at ? fmtDateShort(m.created_at) : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{m.article_code || "—"}</TableCell>
                  <TableCell className="max-w-[300px] truncate text-sm">{m.article_name || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={movementVariant[m.movement_type] ?? "outline"}>
                      {movementLabel[m.movement_type] ?? m.movement_type}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    {Number(m.quantity).toFixed(2)}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm text-muted-foreground">
                    {m.unit_cost != null ? Number(m.unit_cost).toFixed(2) : "—"}
                  </TableCell>
                  <TableCell className="max-w-[260px] truncate text-xs text-muted-foreground">
                    {m.notes || "—"}
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
