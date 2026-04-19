"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtNumber } from "@/lib/erp/format";
import type { BOMEntry, StockArticle } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function ArticleDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: articles = [] } = useQuery({
    queryKey: erpKeys.stockArticles(),
    queryFn: () => api.get<{ articles: StockArticle[] }>("/v1/erp/stock/articles?page_size=500"),
    select: (d) => d.articles,
  });

  const { data: bom = [], isLoading, error } = useQuery({
    queryKey: erpKeys.articleBOM(id),
    queryFn: () => api.get<{ bom: BOMEntry[] }>(`/v1/erp/stock/articles/${id}/bom`),
    select: (d) => d.bom,
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando detalle del artículo" onRetry={() => window.location.reload()} />;

  const article = articles.find((a) => a.id === id);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/almacen"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a almacén
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        <div className="mb-6 flex items-baseline justify-between gap-4">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">
              {article?.name ?? "Artículo"}{" "}
              {article ? (
                <span className="font-mono text-sm text-muted-foreground">{article.code}</span>
              ) : null}
            </h1>
            {article && (
              <p className="mt-0.5 text-sm text-muted-foreground">
                Tipo {article.article_type} · costo prom. {fmtMoney(article.avg_cost)}
              </p>
            )}
          </div>
          {article && (
            <Badge variant={article.active ? "default" : "secondary"}>
              {article.active ? "Activo" : "Inactivo"}
            </Badge>
          )}
        </div>

        <h2 className="mb-3 text-sm font-medium text-muted-foreground">
          Lista de materiales ({bom.length})
        </h2>
        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[140px]">Cód. componente</TableHead>
                <TableHead>Componente</TableHead>
                <TableHead className="w-[120px] text-right">Cantidad</TableHead>
                <TableHead>Notas</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {bom.length === 0 && !isLoading && (
                <TableRow>
                  <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                    Sin componentes definidos.
                  </TableCell>
                </TableRow>
              )}
              {bom.map((b) => (
                <TableRow key={b.id}>
                  <TableCell className="font-mono text-xs">
                    <Link href={`/administracion/almacen/articulos/${b.child_id}`} className="hover:underline">
                      {b.child_code}
                    </Link>
                  </TableCell>
                  <TableCell className="text-sm">{b.child_name}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtNumber(b.quantity)}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{b.notes || "—"}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
