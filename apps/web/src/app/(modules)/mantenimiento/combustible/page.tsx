"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtMoney, fmtNumber } from "@/lib/erp/format";
import type { FuelLog, MaintenanceAsset } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function CombustiblePage() {
  const [assetFilter, setAssetFilter] = useState<string>("");

  const { data: assets = [] } = useQuery({
    queryKey: erpKeys.maintenanceAssets(),
    queryFn: () => api.get<{ assets: MaintenanceAsset[] }>("/v1/erp/maintenance/assets"),
    select: (d) => d.assets,
  });

  const { data: logs = [], isLoading, error } = useQuery({
    queryKey: erpKeys.fuelLogs(assetFilter || undefined),
    queryFn: () => {
      const q = new URLSearchParams({ page_size: "100" });
      if (assetFilter) q.set("asset_id", assetFilter);
      return api.get<{ fuel_logs: FuelLog[] }>(`/v1/erp/maintenance/fuel-logs?${q}`);
    },
    select: (d) => d.fuel_logs,
  });

  if (error)
    return <ErrorState message="Error cargando combustible" onRetry={() => window.location.reload()} />;

  const totalLiters = logs.reduce((s, l) => s + (l.liters ?? 0), 0);
  const totalCost = logs.reduce((s, l) => s + (l.cost ?? 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Combustible</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Registro de cargas de combustible por equipo — litros, kilometraje y costo.
          </p>
        </div>

        <div className="mb-4 flex items-center gap-3">
          <Label className="shrink-0 text-sm">Equipo</Label>
          <Select
            value={assetFilter || "all"}
            onValueChange={(v) => setAssetFilter(!v || v === "all" ? "" : v)}
          >
            <SelectTrigger className="w-64 bg-card">
              <SelectValue placeholder="Todos los equipos" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Todos los equipos</SelectItem>
              {assets.map((a) => (
                <SelectItem key={a.id} value={a.id}>
                  {a.code} — {a.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <div className="ml-auto text-sm text-muted-foreground">
            {logs.length} cargas · {fmtNumber(totalLiters)} lts · {fmtMoney(totalCost)}
          </div>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Fecha</TableHead>
                <TableHead className="w-[120px]">Cód. equipo</TableHead>
                <TableHead>Equipo</TableHead>
                <TableHead className="w-[110px] text-right">Litros</TableHead>
                <TableHead className="w-[110px] text-right">Km</TableHead>
                <TableHead className="w-[130px] text-right">Costo</TableHead>
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
              {!isLoading && logs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-20 text-center text-sm text-muted-foreground">
                    Sin cargas registradas.
                  </TableCell>
                </TableRow>
              )}
              {logs.map((l) => (
                <TableRow key={l.id}>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(l.date)}</TableCell>
                  <TableCell className="font-mono text-xs">
                    {l.asset_id ? (
                      <Link
                        href={`/mantenimiento/equipos/${l.asset_id}`}
                        className="hover:underline"
                      >
                        {l.asset_code ?? "—"}
                      </Link>
                    ) : (
                      l.asset_code ?? "—"
                    )}
                  </TableCell>
                  <TableCell className="text-sm">{l.asset_name ?? "—"}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtNumber(l.liters)}</TableCell>
                  <TableCell className="text-right font-mono text-sm text-muted-foreground">
                    {fmtNumber(l.km_reading)}
                  </TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(l.cost)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
