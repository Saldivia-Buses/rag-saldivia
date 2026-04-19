"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import type { EntityCreditRating, EntitySearchResult } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { EntityPicker } from "@/components/erp/entity-picker";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { SearchIcon, XIcon } from "lucide-react";

type RatingTab = "all" | "A" | "B" | "C" | "X";

const ratingLabel: Record<string, string> = {
  A: "A — Aprobado",
  B: "B — Atención",
  C: "C — Riesgo",
  X: "X — Bloqueado",
  "-": "Sin calificación",
};

const ratingBadgeVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  A: "secondary",
  B: "outline",
  C: "destructive",
  X: "destructive",
};

export default function CalificacionesPage() {
  const [tab, setTab] = useState<RatingTab>("all");
  const [pickerOpen, setPickerOpen] = useState(false);
  const [entity, setEntity] = useState<EntitySearchResult | null>(null);

  const queryParams: Record<string, string> = {};
  if (tab !== "all") queryParams.rating = tab;
  if (entity) queryParams.entity_id = entity.id;

  const { data: ratings = [], isLoading, error } = useQuery({
    queryKey: erpKeys.creditRatings(queryParams),
    queryFn: () => {
      const qs = new URLSearchParams({ ...queryParams, page_size: "100" }).toString();
      return api.get<{ ratings: EntityCreditRating[] }>(`/v1/erp/entities/credit-ratings?${qs}`);
    },
    select: (d) => d.ratings,
  });

  if (error)
    return <ErrorState message="Error cargando calificaciones" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Calificaciones de cuentas</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Historial de calificaciones crediticias de proveedores y clientes — una fila por evento de calificación (REG_CUENTA_CALIFICACION).
          </p>
        </div>

        <div className="mb-4 flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2">
            {entity ? (
              <div className="flex items-center gap-2 rounded-md border border-border/60 bg-muted/30 px-3 py-1.5">
                <div>
                  <div className="text-sm font-medium">{entity.name}</div>
                  <div className="font-mono text-xs text-muted-foreground">cod. {entity.code}</div>
                </div>
                <Button size="sm" variant="ghost" onClick={() => setEntity(null)}>
                  <XIcon className="size-3.5" />
                </Button>
              </div>
            ) : (
              <Button variant="outline" size="sm" onClick={() => setPickerOpen(true)}>
                <SearchIcon className="mr-1.5 size-3.5" />
                Filtrar por cuenta
              </Button>
            )}
          </div>
          <EntityPicker
            type="supplier"
            open={pickerOpen}
            onOpenChange={setPickerOpen}
            onSelect={setEntity}
          />
        </div>

        <Tabs value={tab} onValueChange={(v) => setTab(v as RatingTab)} className="mb-4">
          <TabsList>
            <TabsTrigger value="all">Todas</TabsTrigger>
            <TabsTrigger value="A">A</TabsTrigger>
            <TabsTrigger value="B">B</TabsTrigger>
            <TabsTrigger value="C">C</TabsTrigger>
            <TabsTrigger value="X">X</TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Fecha</TableHead>
                <TableHead className="w-[90px]">Cód.</TableHead>
                <TableHead>Cuenta</TableHead>
                <TableHead className="w-[90px]">Tipo</TableHead>
                <TableHead className="w-[150px]">Calificación</TableHead>
                <TableHead>Referencia</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={6}>
                    <Skeleton className="h-32 w-full" />
                  </TableCell>
                </TableRow>
              )}
              {!isLoading && ratings.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin calificaciones en esta vista.
                  </TableCell>
                </TableRow>
              )}
              {ratings.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="text-sm">
                    {r.rated_at ? fmtDateShort(r.rated_at) : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-sm">{r.entity_legacy_id || "—"}</TableCell>
                  <TableCell className="max-w-[300px] truncate text-sm">
                    {r.entity_name ?? <span className="text-muted-foreground">—</span>}
                  </TableCell>
                  <TableCell className="text-xs capitalize text-muted-foreground">
                    {r.entity_type ?? "—"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={ratingBadgeVariant[r.rating] ?? "outline"}>
                      {ratingLabel[r.rating] ?? r.rating}
                    </Badge>
                  </TableCell>
                  <TableCell className="max-w-[380px] truncate text-sm">{r.reference || "—"}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
