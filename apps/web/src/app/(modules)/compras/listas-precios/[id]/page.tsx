"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate, fmtMoney } from "@/lib/erp/format";
import type { PriceListDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function PriceListDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.priceList(id),
    queryFn: () => api.get<PriceListDetail>(`/v1/erp/sales/price-lists/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando lista de precios" onRetry={() => window.location.reload()} />;

  const pl = data?.price_list;
  const items = data?.items ?? [];

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/compras/listas-precios"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a listas de precios
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {pl && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{pl.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Vigencia {fmtDate(pl.valid_from)} — {fmtDate(pl.valid_until)}
                </p>
              </div>
              <Badge variant={pl.active ? "secondary" : "outline"}>
                {pl.active ? "Activa" : "Inactiva"}
              </Badge>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Items ({items.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px]">Cód. art.</TableHead>
                    <TableHead>Artículo</TableHead>
                    <TableHead>Descripción override</TableHead>
                    <TableHead className="w-[140px] text-right">Precio</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                        Sin items en la lista.
                      </TableCell>
                    </TableRow>
                  )}
                  {items.map((it) => (
                    <TableRow key={it.id}>
                      <TableCell className="font-mono text-xs">{it.article_code ?? "—"}</TableCell>
                      <TableCell className="text-sm">{it.article_name ?? "—"}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {it.description ?? "—"}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(it.price)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
