"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { StockArticle, StockMovement, Warehouse } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
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

const MOVEMENT_TYPES = ["in", "out", "transfer", "adjustment"] as const;
type MovementType = (typeof MOVEMENT_TYPES)[number];

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

        <NewMovementForm />

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

function NewMovementForm() {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [articleId, setArticleId] = useState("");
  const [warehouseId, setWarehouseId] = useState("");
  const [type, setType] = useState<MovementType>("in");
  const [quantity, setQuantity] = useState("");
  const [unitCost, setUnitCost] = useState("");
  const [notes, setNotes] = useState("");

  const { data: articles = [] } = useQuery({
    queryKey: erpKeys.stockArticles({ page_size: "200" }),
    queryFn: () =>
      api.get<{ articles: StockArticle[] }>("/v1/erp/stock/articles?page_size=200"),
    select: (d) => d.articles,
    enabled: open,
  });

  const { data: warehouses = [] } = useQuery({
    queryKey: erpKeys.warehouses(),
    queryFn: () => api.get<{ warehouses: Warehouse[] }>("/v1/erp/stock/warehouses"),
    select: (d) => d.warehouses,
    enabled: open,
  });

  const mutation = useMutation({
    mutationFn: (body: {
      article_id: string;
      warehouse_id: string;
      movement_type: string;
      quantity: string;
      unit_cost: string;
      notes: string;
    }) => api.post<StockMovement>("/v1/erp/stock/movements", body),
    onSuccess: () => {
      setArticleId("");
      setWarehouseId("");
      setQuantity("");
      setUnitCost("");
      setNotes("");
      setOpen(false);
      qc.invalidateQueries({ queryKey: [...erpKeys.all, "stock"] });
    },
  });

  if (!open) {
    return (
      <div className="mb-6">
        <Button type="button" onClick={() => setOpen(true)}>
          + Nuevo movimiento
        </Button>
      </div>
    );
  }

  return (
    <form
      className="mb-6 rounded-xl border border-border/40 bg-card p-4"
      onSubmit={(e) => {
        e.preventDefault();
        if (!articleId || !warehouseId || !quantity.trim()) return;
        mutation.mutate({
          article_id: articleId,
          warehouse_id: warehouseId,
          movement_type: type,
          quantity: quantity.trim(),
          unit_cost: unitCost.trim(),
          notes: notes.trim(),
        });
      }}
    >
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-sm font-medium">Nuevo movimiento</h3>
        <Button type="button" variant="ghost" size="sm" onClick={() => setOpen(false)}>
          Cancelar
        </Button>
      </div>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div className="grid gap-1.5">
          <Label className="text-xs">Artículo</Label>
          <select
            className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
            value={articleId}
            onChange={(e) => setArticleId(e.target.value)}
            required
            disabled={mutation.isPending}
          >
            <option value="">Seleccionar…</option>
            {articles.map((a) => (
              <option key={a.id} value={a.id}>
                {a.code} · {a.name}
              </option>
            ))}
          </select>
        </div>
        <div className="grid gap-1.5">
          <Label className="text-xs">Depósito</Label>
          <select
            className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
            value={warehouseId}
            onChange={(e) => setWarehouseId(e.target.value)}
            required
            disabled={mutation.isPending}
          >
            <option value="">Seleccionar…</option>
            {warehouses.map((w) => (
              <option key={w.id} value={w.id}>
                {w.code} · {w.name}
              </option>
            ))}
          </select>
        </div>
        <div className="grid gap-1.5">
          <Label className="text-xs">Tipo</Label>
          <select
            className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
            value={type}
            onChange={(e) => setType(e.target.value as MovementType)}
            disabled={mutation.isPending}
          >
            {MOVEMENT_TYPES.map((t) => (
              <option key={t} value={t}>
                {movementLabel[t]}
              </option>
            ))}
          </select>
        </div>
        <div className="grid gap-1.5">
          <Label className="text-xs">Cantidad</Label>
          <Input
            type="number"
            step="0.0001"
            placeholder="0.00"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            required
            disabled={mutation.isPending}
          />
        </div>
        <div className="grid gap-1.5">
          <Label className="text-xs">Costo unitario (opcional)</Label>
          <Input
            type="number"
            step="0.0001"
            placeholder="0.00"
            value={unitCost}
            onChange={(e) => setUnitCost(e.target.value)}
            disabled={mutation.isPending}
          />
        </div>
        <div className="grid gap-1.5">
          <Label className="text-xs">Notas (opcional)</Label>
          <Input
            placeholder="..."
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            disabled={mutation.isPending}
          />
        </div>
      </div>
      <div className="mt-3 flex items-center justify-between">
        {mutation.isError ? (
          <p className="text-xs text-destructive">Error al guardar movimiento.</p>
        ) : (
          <span />
        )}
        <Button
          type="submit"
          disabled={mutation.isPending || !articleId || !warehouseId || !quantity.trim()}
        >
          {mutation.isPending ? "Guardando…" : "Registrar movimiento"}
        </Button>
      </div>
    </form>
  );
}
