"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import type { CustomerVehicle } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function VehiculosClientesPage() {
  const { data: vehicles = [], isLoading, error } = useQuery({
    queryKey: erpKeys.customerVehicles(),
    queryFn: () => api.get<{ vehicles: CustomerVehicle[] }>("/v1/erp/workshop/vehicles?page_size=100"),
    select: (d) => d.vehicles,
  });

  if (error)
    return <ErrorState message="Error cargando vehículos" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Vehículos de clientes</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Parque vehicular registrado por cliente — dominio, chasis, marca, modelo, destino.
          </p>
        </div>

        <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Dominio</TableHead>
                <TableHead className="w-[90px]">Interno</TableHead>
                <TableHead>Marca / Chasis</TableHead>
                <TableHead className="w-[80px]">Año</TableHead>
                <TableHead className="w-[80px]">Plazas</TableHead>
                <TableHead>Destino</TableHead>
                <TableHead className="w-[110px]">Estado</TableHead>
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
              {!isLoading && vehicles.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="h-20 text-center text-sm text-muted-foreground">
                    Sin vehículos registrados.
                  </TableCell>
                </TableRow>
              )}
              {vehicles.map((v) => (
                <TableRow key={v.id}>
                  <TableCell className="font-mono text-sm">
                    <Link href={`/mantenimiento/taller/vehiculos/${v.id}`} className="hover:underline">
                      {v.plate || v.id.slice(0, 8)}
                    </Link>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {v.internal_number != null ? v.internal_number : "—"}
                  </TableCell>
                  <TableCell className="max-w-[280px] truncate text-sm">
                    {v.brand}
                    {v.chassis_serial ? <span className="ml-1 text-xs text-muted-foreground">· {v.chassis_serial}</span> : null}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{v.model_year ?? "—"}</TableCell>
                  <TableCell className="text-right font-mono text-xs">{v.seating_capacity || "—"}</TableCell>
                  <TableCell className="max-w-[200px] truncate text-xs text-muted-foreground">
                    {v.destination || "—"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={v.active ? "secondary" : "outline"}>
                      {v.active ? "Activo" : "Baja"}
                    </Badge>
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
