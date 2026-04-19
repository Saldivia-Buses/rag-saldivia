"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { CustomerVehicle, VehicleIncident } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function VehicleDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const vehicleQ = useQuery({
    queryKey: erpKeys.customerVehicle(id),
    queryFn: () => api.get<CustomerVehicle>(`/v1/erp/workshop/vehicles/${id}`),
    enabled: !!id,
  });

  const incidentsQ = useQuery({
    queryKey: erpKeys.vehicleIncidents({ vehicle_id: id }),
    queryFn: () =>
      api.get<{ incidents: VehicleIncident[] }>(
        `/v1/erp/workshop/incidents?vehicle_id=${id}&page_size=100`
      ),
    select: (d) => d.incidents,
    enabled: !!id,
  });

  if (vehicleQ.error)
    return <ErrorState message="Error cargando vehículo" onRetry={() => window.location.reload()} />;

  const vehicle = vehicleQ.data;
  const incidents = incidentsQ.data ?? [];
  const openIncidents = incidents.filter((i) => i.status !== "resolved").length;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/mantenimiento/taller/vehiculos"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a vehículos
        </Link>

        {vehicleQ.isLoading && <Skeleton className="h-48 w-full" />}

        {vehicle && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">
                  {vehicle.plate || "—"}
                  {vehicle.internal_number != null ? (
                    <span className="ml-2 text-sm font-normal text-muted-foreground">
                      Interno {vehicle.internal_number}
                    </span>
                  ) : null}
                </h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {vehicle.brand}
                  {vehicle.chassis_serial ? ` · Chasis ${vehicle.chassis_serial}` : ""}
                  {vehicle.body_serial ? ` · Carrocería ${vehicle.body_serial}` : ""}
                </p>
              </div>
              <Badge variant={vehicle.active ? "default" : "secondary"}>
                {vehicle.active ? "Activo" : "Baja"}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Año" value={vehicle.model_year != null ? String(vehicle.model_year) : "—"} />
              <Metric label="Plazas" value={vehicle.seating_capacity ? String(vehicle.seating_capacity) : "—"} />
              <Metric label="Combustible" value={vehicle.fuel_type || "—"} />
              <Metric label="Incidentes abiertos" value={String(openIncidents)} />
            </div>

            {(vehicle.destination || vehicle.color || vehicle.observations) && (
              <div className="mb-6 rounded-xl border border-border/40 bg-card px-4 py-3 text-sm">
                {vehicle.destination && (
                  <div>
                    <span className="text-muted-foreground">Destino: </span>
                    {vehicle.destination}
                  </div>
                )}
                {vehicle.color && (
                  <div>
                    <span className="text-muted-foreground">Color: </span>
                    {vehicle.color}
                  </div>
                )}
                {vehicle.observations && (
                  <div className="mt-1 whitespace-pre-wrap">
                    <span className="text-muted-foreground">Observaciones: </span>
                    {vehicle.observations}
                  </div>
                )}
              </div>
            )}

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Incidentes ({incidents.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[130px]">Fecha</TableHead>
                    <TableHead className="w-[160px]">Tipo</TableHead>
                    <TableHead>Detalle</TableHead>
                    <TableHead className="w-[140px]">Responsable</TableHead>
                    <TableHead className="w-[110px]">Estado</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {incidents.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                        Sin incidentes registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {incidents.map((inc) => (
                    <TableRow key={inc.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {inc.incident_date ? fmtDate(inc.incident_date) : "—"}
                      </TableCell>
                      <TableCell className="text-sm">{inc.incident_type_name ?? "—"}</TableCell>
                      <TableCell className="text-sm">
                        {inc.location ? <span className="text-muted-foreground">{inc.location}</span> : null}
                        {inc.location && inc.notes ? " — " : null}
                        {inc.notes || (inc.location ? "" : "—")}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">{inc.responsible || "—"}</TableCell>
                      <TableCell>
                        <Badge variant={inc.status === "resolved" ? "secondary" : "default"}>
                          {inc.status}
                        </Badge>
                      </TableCell>
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

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 font-mono text-sm">{value}</div>
    </div>
  );
}
