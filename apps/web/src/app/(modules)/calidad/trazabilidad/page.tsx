"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtNumber, fmtDate } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { SearchIcon } from "lucide-react";

interface StockMovement {
  id: string;
  article_code: string;
  article_name: string;
  movement_type: string;
  quantity: number;
  unit_cost: number;
  notes: string;
  created_at: string;
}

interface Article { id: string; code: string; name: string; }

const movementTypeBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  in: { label: "Ingreso", variant: "default" },
  out: { label: "Egreso", variant: "secondary" },
  transfer: { label: "Transferencia", variant: "outline" },
  adjustment: { label: "Ajuste", variant: "outline" },
};

export default function TrazabilidadPage() {
  const [search, setSearch] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");

  const { data: movements = [], isLoading: loadingMovements, error: movementsError } = useQuery({
    queryKey: [...erpKeys.all, "stock", "movements", { page_size: "100" }] as const,
    queryFn: () => api.get<{ movements: StockMovement[] }>("/v1/erp/stock/movements?page_size=100"),
    select: (d) => d.movements,
  });

  const { data: articles = [] } = useQuery({
    queryKey: erpKeys.stockArticles({ page_size: "200" }),
    queryFn: () => api.get<{ articles: Article[] }>("/v1/erp/stock/articles?page_size=200"),
    select: (d) => d.articles,
  });

  if (movementsError) return <ErrorState message="Error cargando movimientos de stock" onRetry={() => window.location.reload()} />;
  if (loadingMovements) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const searchLower = search.toLowerCase();

  const filtered = movements.filter((m) => {
    const matchesSearch =
      !search ||
      m.article_name.toLowerCase().includes(searchLower) ||
      m.article_code.toLowerCase().includes(searchLower) ||
      (m.notes ?? "").toLowerCase().includes(searchLower);

    const matchesFrom = !dateFrom || m.created_at >= dateFrom;
    const matchesTo = !dateTo || m.created_at <= dateTo + "T23:59:59";

    return matchesSearch && matchesFrom && matchesTo;
  });

  const articlesWithMovements = new Set(movements.map((m) => m.article_code)).size;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Trazabilidad</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Movimientos de stock — {filtered.length} de {movements.length} registros · {articlesWithMovements} artículos con actividad · {articles.length} artículos totales
          </p>
        </div>

        <div className="flex flex-wrap gap-4 mb-4">
          <div className="flex-1 min-w-[200px] relative">
            <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
            <Input
              className="pl-9"
              placeholder="Buscar por artículo o código..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <div className="flex items-center gap-2">
            <Label className="text-sm text-muted-foreground whitespace-nowrap">Desde</Label>
            <Input type="date" className="w-36" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
          </div>
          <div className="flex items-center gap-2">
            <Label className="text-sm text-muted-foreground whitespace-nowrap">Hasta</Label>
            <Input type="date" className="w-36" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
          </div>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-36">Fecha</TableHead>
                <TableHead className="w-24">Código</TableHead>
                <TableHead>Artículo</TableHead>
                <TableHead className="w-32">Tipo</TableHead>
                <TableHead className="text-right w-24">Cantidad</TableHead>
                <TableHead>Referencia</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((m) => {
                const badge = movementTypeBadge[m.movement_type] ?? { label: m.movement_type, variant: "outline" as const };
                return (
                  <TableRow key={m.id}>
                    <TableCell className="text-sm text-muted-foreground">{fmtDate(m.created_at)}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{m.article_code}</TableCell>
                    <TableCell className="text-sm">{m.article_name}</TableCell>
                    <TableCell><Badge variant={badge.variant}>{badge.label}</Badge></TableCell>
                    <TableCell className="text-right font-mono text-sm">{fmtNumber(m.quantity)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground truncate max-w-xs">{m.notes || "—"}</TableCell>
                  </TableRow>
                );
              })}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    {search || dateFrom || dateTo
                      ? "Sin resultados para los filtros aplicados."
                      : "Sin movimientos de stock registrados."}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
