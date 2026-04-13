"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtNumber, fmtDateShort } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { AlertTriangleIcon, PlusIcon, ShoppingCartIcon } from "lucide-react";

interface Article { id: string; code: string; name: string; min_stock: number; active: boolean; }
interface StockLevel { article_code: string; article_name: string; warehouse_code: string; warehouse_name: string; quantity: number; }
interface Supplier { id: string; name: string; }
interface POLine { description: string; quantity: string; unit_price: string; }

interface LowStockRow {
  articleId: string;
  code: string;
  name: string;
  minStock: number;
  currentStock: number;
  deficit: number;
  warehouseName: string;
}

export default function AbastecimientoPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [prefill, setPrefill] = useState<{ name: string; deficit: number } | null>(null);
  const queryClient = useQueryClient();

  const { data: articles = [], isLoading: loadingArticles, error: articlesError } = useQuery({
    queryKey: erpKeys.stockArticles({ active: "true", page_size: "200" }),
    queryFn: () => api.get<{ articles: Article[] }>("/v1/erp/stock/articles?active=true&page_size=200"),
    select: (d) => d.articles,
  });

  const { data: levels = [], isLoading: loadingLevels, error: levelsError } = useQuery({
    queryKey: erpKeys.stockLevels(),
    queryFn: () => api.get<{ levels: StockLevel[] }>("/v1/erp/stock/levels"),
    select: (d) => d.levels,
  });

  const createMutation = useMutation({
    mutationFn: (data: { number: string; supplier_id: string; date: string; lines: POLine[] }) =>
      api.post("/v1/erp/purchasing/orders", data),
    onSuccess: () => {
      toast.success("Orden de compra creada");
      queryClient.invalidateQueries({ queryKey: erpKeys.purchaseOrders() });
      setCreateOpen(false);
      setPrefill(null);
    },
    onError: permissionErrorToast,
  });

  const isLoading = loadingArticles || loadingLevels;
  const error = articlesError || levelsError;

  if (error) return <ErrorState message="Error cargando datos de stock" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  // Build a map of article_code → total stock across all warehouses
  const stockByCode = new Map<string, { quantity: number; warehouseName: string }>();
  for (const lvl of levels) {
    const existing = stockByCode.get(lvl.article_code);
    if (!existing) {
      stockByCode.set(lvl.article_code, { quantity: lvl.quantity, warehouseName: lvl.warehouse_name });
    } else {
      stockByCode.set(lvl.article_code, { quantity: existing.quantity + lvl.quantity, warehouseName: existing.warehouseName });
    }
  }

  const lowStockRows: LowStockRow[] = articles
    .filter((a) => a.min_stock > 0)
    .map((a) => {
      const level = stockByCode.get(a.code);
      const currentStock = level?.quantity ?? 0;
      return {
        articleId: a.id,
        code: a.code,
        name: a.name,
        minStock: a.min_stock,
        currentStock,
        deficit: a.min_stock - currentStock,
        warehouseName: level?.warehouseName ?? "—",
      };
    })
    .filter((r) => r.deficit > 0)
    .sort((a, b) => b.deficit - a.deficit);

  function openCreate(row: LowStockRow) {
    setPrefill({ name: row.name, deficit: row.deficit });
    setCreateOpen(true);
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Abastecimiento</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Artículos por debajo del stock mínimo — {lowStockRows.length} con déficit
            </p>
          </div>
          <Button size="sm" onClick={() => { setPrefill(null); setCreateOpen(true); }}>
            <PlusIcon className="size-4 mr-1.5" />Nueva OC
          </Button>
        </div>

        {lowStockRows.length === 0 ? (
          <div className="rounded-xl border border-border/40 bg-card flex flex-col items-center justify-center py-16 text-center">
            <ShoppingCartIcon className="size-10 text-muted-foreground/40 mb-3" />
            <p className="text-sm font-medium">Todo el stock está en nivel óptimo</p>
            <p className="text-xs text-muted-foreground mt-1">No hay artículos por debajo del mínimo configurado.</p>
          </div>
        ) : (
          <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-24">Código</TableHead>
                  <TableHead>Artículo</TableHead>
                  <TableHead className="text-right w-28">Stock mínimo</TableHead>
                  <TableHead className="text-right w-28">Stock actual</TableHead>
                  <TableHead className="text-right w-28">Diferencia</TableHead>
                  <TableHead className="w-32">Depósito</TableHead>
                  <TableHead className="w-28" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {lowStockRows.map((row) => (
                  <TableRow key={row.articleId}>
                    <TableCell className="font-mono text-xs text-muted-foreground">{row.code}</TableCell>
                    <TableCell className="text-sm font-medium">{row.name}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtNumber(row.minStock)}</TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtNumber(row.currentStock)}</TableCell>
                    <TableCell className="text-right">
                      <Badge variant="destructive" className="font-mono">
                        <AlertTriangleIcon className="size-3 mr-1" />
                        -{fmtNumber(row.deficit)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{row.warehouseName}</TableCell>
                    <TableCell>
                      <Button size="sm" variant="outline" onClick={() => openCreate(row)}>
                        <ShoppingCartIcon className="size-3.5 mr-1" />Generar OC
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => { if (!v) { setCreateOpen(false); setPrefill(null); } }}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader><DialogTitle>Nueva orden de compra</DialogTitle></DialogHeader>
          <CreateOrderForm
            prefill={prefill}
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => { setCreateOpen(false); setPrefill(null); }}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateOrderForm({
  prefill,
  onSubmit,
  isPending,
  onClose,
}: {
  prefill: { name: string; deficit: number } | null;
  onSubmit: (data: { number: string; supplier_id: string; date: string; lines: POLine[] }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(new Date().toISOString().split("T")[0]);
  const [supplierId, setSupplierId] = useState("");
  const [lines, setLines] = useState<POLine[]>(() => [
    prefill
      ? { description: prefill.name, quantity: String(Math.ceil(prefill.deficit)), unit_price: "0" }
      : { description: "", quantity: "1", unit_price: "0" },
  ]);

  const { data: suppliers = [] } = useQuery({
    queryKey: erpKeys.entities("supplier"),
    queryFn: () => api.get<{ entities: Supplier[] }>("/v1/erp/entities?type=supplier&page_size=200"),
    select: (d) => d.entities,
  });

  function addLine() {
    setLines((prev) => [...prev, { description: "", quantity: "1", unit_price: "0" }]);
  }

  function updateLine(i: number, field: keyof POLine, value: string) {
    setLines((prev) => prev.map((l, idx) => idx === i ? { ...l, [field]: value } : l));
  }

  function removeLine(i: number) {
    setLines((prev) => prev.filter((_, idx) => idx !== i));
  }

  const canSubmit = !!supplierId && lines.length > 0 && lines.every((l) => l.description);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!canSubmit) return;
        onSubmit({ number, supplier_id: supplierId, date, lines });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="OC-0001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Proveedor</Label>
        <Select value={supplierId} onValueChange={(v) => v && setSupplierId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar proveedor..." /></SelectTrigger>
          <SelectContent>
            {(suppliers as Supplier[]).map((s) => (
              <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Líneas</Label>
          <Button type="button" size="sm" variant="outline" onClick={addLine}>
            <PlusIcon className="size-3.5 mr-1" />Agregar línea
          </Button>
        </div>
        <div className="space-y-2">
          {lines.map((line, i) => (
            <div key={i} className="grid grid-cols-12 gap-2 items-start">
              <div className="col-span-7">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Descripción</p>}
                <Input value={line.description} onChange={(e) => updateLine(i, "description", e.target.value)} placeholder="Descripción del ítem" />
              </div>
              <div className="col-span-2">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Cantidad</p>}
                <Input type="number" value={line.quantity} onChange={(e) => updateLine(i, "quantity", e.target.value)} placeholder="1" />
              </div>
              <div className="col-span-2">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Precio unit.</p>}
                <Input type="number" value={line.unit_price} onChange={(e) => updateLine(i, "unit_price", e.target.value)} placeholder="0" />
              </div>
              <div className="col-span-1 flex items-end">
                {i === 0 && <div className="h-[18px]" />}
                <Button type="button" size="sm" variant="ghost" disabled={lines.length === 1}
                  onClick={() => removeLine(i)} className="px-2 text-muted-foreground hover:text-destructive">×</Button>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear OC"}</Button>
      </div>
    </form>
  );
}
