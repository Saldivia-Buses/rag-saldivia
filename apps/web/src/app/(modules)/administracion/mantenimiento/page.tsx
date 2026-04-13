"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort, fmtNumber } from "@/lib/erp/format";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ClipboardListIcon, FuelIcon, BusIcon, PlusIcon, CheckCircleIcon, XIcon } from "lucide-react";

// ── Existing types ─────────────────────────────────────────────────────────────
interface WorkOrder { id: string; number: string; date: string; asset_code: string; asset_name: string; work_type: string; status: string; priority: string; }
interface FuelLog { id: string; asset_code: string; asset_name: string; date: string; liters: number; km_reading: number; cost: number; }
interface Asset { id: string; code: string; name: string; }

// ── Workshop: customer vehicles & incidents ────────────────────────────────────
interface CustomerVehicle {
  id: string; plate: string; chassis_serial: string; body_serial: string;
  brand: string; model_year: number | null; seating_capacity: number;
  fuel_type: string; owner_name: string | null; driver_name: string | null;
  destination: string; active: boolean;
}
interface VehicleIncident {
  id: string; incident_type_name: string | null; incident_date: string;
  location: string; responsible: string; notes: string; status: string;
}
interface IncidentType { id: string; name: string; }
interface Entity { id: string; name: string; }

// ── Lookup maps ────────────────────────────────────────────────────────────────
const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", in_progress: "outline", completed: "default", cancelled: "secondary" };
const prioColor: Record<string, string> = { low: "text-muted-foreground", normal: "", high: "text-amber-500", urgent: "text-red-500 font-medium" };
const typeLabel: Record<string, string> = { preventive: "Preventivo", corrective: "Correctivo", inspection: "Inspección" };

const fuelLabel: Record<string, string> = {
  diesel: "Diesel", gasolina: "Gasolina", gnc: "GNC", electric: "Eléctrico", hybrid: "Híbrido",
};

const incidentStatusVariant: Record<string, "default" | "secondary" | "outline"> = {
  pending: "secondary", resolved: "default",
};
const incidentStatusLabel: Record<string, string> = { pending: "Pendiente", resolved: "Resuelto" };

// ── Page ───────────────────────────────────────────────────────────────────────
export default function MantenimientoPage() {
  const queryClient = useQueryClient();
  const [woOpen, setWoOpen] = useState(false);
  const [fuelOpen, setFuelOpen] = useState(false);
  const [vehicleOpen, setVehicleOpen] = useState(false);
  const [incidentOpen, setIncidentOpen] = useState(false);
  const [selectedVehicleId, setSelectedVehicleId] = useState<string | null>(null);

  // ── Work orders ──────────────────────────────────────────────────────────────
  const { data: workOrders = [], isLoading, error } = useQuery({
    queryKey: erpKeys.workOrders(),
    queryFn: () => api.get<{ work_orders: WorkOrder[] }>("/v1/erp/maintenance/work-orders?page_size=50"),
    select: (d) => d.work_orders,
  });

  const { data: fuelLogs = [] } = useQuery({
    queryKey: [...erpKeys.all, "maintenance", "fuel-logs"] as const,
    queryFn: () => api.get<{ fuel_logs: FuelLog[] }>("/v1/erp/maintenance/fuel-logs?page_size=50"),
    select: (d) => d.fuel_logs,
  });

  const { data: assets = [] } = useQuery({
    queryKey: [...erpKeys.all, "maintenance", "assets"] as const,
    queryFn: () => api.get<{ assets: Asset[] }>("/v1/erp/maintenance/assets?page_size=200"),
    select: (d) => d.assets,
  });

  // ── Customer vehicles ────────────────────────────────────────────────────────
  const { data: vehicles = [] } = useQuery({
    queryKey: [...erpKeys.all, "workshop", "vehicles"] as const,
    queryFn: () => api.get<{ vehicles: CustomerVehicle[] }>("/v1/erp/workshop/vehicles?page_size=50"),
    select: (d) => d.vehicles,
  });

  const { data: entities = [] } = useQuery({
    queryKey: erpKeys.entities(),
    queryFn: () => api.get<{ entities: Entity[] }>("/v1/erp/entities?page_size=200"),
    select: (d) => d.entities,
  });

  // ── Incidents for selected vehicle ──────────────────────────────────────────
  const { data: incidents = [] } = useQuery({
    queryKey: [...erpKeys.all, "workshop", "incidents", selectedVehicleId] as const,
    queryFn: () => api.get<{ incidents: VehicleIncident[] }>(`/v1/erp/workshop/incidents?vehicle_id=${selectedVehicleId}&page_size=50`),
    enabled: !!selectedVehicleId,
    select: (d) => d.incidents,
  });

  const { data: incidentTypes = [] } = useQuery({
    queryKey: [...erpKeys.all, "workshop", "incident-types"] as const,
    queryFn: () => api.get<{ incident_types: IncidentType[] }>("/v1/erp/workshop/incident-types"),
    select: (d) => d.incident_types,
  });

  // ── Mutations ────────────────────────────────────────────────────────────────
  const createWorkOrderMutation = useMutation({
    mutationFn: (data: { Number: string; AssetID: string; WorkType: string; Description: string; Date: string; Priority: string }) =>
      api.post("/v1/erp/maintenance/work-orders", data),
    onSuccess: () => {
      toast.success("Orden de trabajo creada");
      queryClient.invalidateQueries({ queryKey: erpKeys.workOrders() });
      setWoOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createFuelLogMutation = useMutation({
    mutationFn: (data: { AssetID: string; Date: string; Liters: string; Cost: string; KmReading: number | null }) =>
      api.post("/v1/erp/maintenance/fuel-logs", data),
    onSuccess: () => {
      toast.success("Carga registrada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "maintenance", "fuel-logs"] });
      setFuelOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createVehicleMutation = useMutation({
    mutationFn: (data: {
      plate: string; brand: string; model_year: number | null; seating_capacity: number | null;
      owner_id: string; chassis_serial: string; body_serial: string;
      fuel_type: string; destination: string; observations: string;
    }) => api.post("/v1/erp/workshop/vehicles", data),
    onSuccess: () => {
      toast.success("Vehículo registrado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "workshop", "vehicles"] });
      setVehicleOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createIncidentMutation = useMutation({
    mutationFn: (data: { vehicle_id: string; incident_type_id: string; incident_date: string; location: string; responsible: string; notes: string }) =>
      api.post("/v1/erp/workshop/incidents", data),
    onSuccess: () => {
      toast.success("Novedad registrada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "workshop", "incidents", selectedVehicleId] });
      setIncidentOpen(false);
    },
    onError: permissionErrorToast,
  });

  const resolveIncidentMutation = useMutation({
    mutationFn: (incidentId: string) => api.patch(`/v1/erp/workshop/incidents/${incidentId}/resolve`),
    onSuccess: () => {
      toast.success("Novedad resuelta");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "workshop", "incidents", selectedVehicleId] });
    },
    onError: permissionErrorToast,
  });

  const selectedVehicle = vehicles.find((v) => v.id === selectedVehicleId) ?? null;

  if (error) return <ErrorState message="Error cargando mantenimiento" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Mantenimiento</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Órdenes de trabajo, combustible y vehículos de clientes</p>
          </div>
          <Button size="sm" onClick={() => setWoOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva OT</Button>
        </div>

        <Tabs defaultValue="work-orders">
          <TabsList className="mb-4">
            <TabsTrigger value="work-orders"><ClipboardListIcon className="size-3.5 mr-1.5" />Órdenes de Trabajo</TabsTrigger>
            <TabsTrigger value="fuel"><FuelIcon className="size-3.5 mr-1.5" />Combustible</TabsTrigger>
            <TabsTrigger value="vehicles"><BusIcon className="size-3.5 mr-1.5" />Vehículos de clientes</TabsTrigger>
          </TabsList>

          {/* ── Work Orders tab ──────────────────────────────────────────────── */}
          <TabsContent value="work-orders">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow>
                <TableHead className="w-20">OT</TableHead><TableHead className="w-28">Fecha</TableHead>
                <TableHead>Equipo</TableHead><TableHead className="w-28">Tipo</TableHead>
                <TableHead className="w-20">Prior.</TableHead><TableHead className="w-28">Estado</TableHead>
              </TableRow></TableHeader>
                <TableBody>
                  {workOrders.map((wo) => (
                    <TableRow key={wo.id}>
                      <TableCell className="font-mono text-sm">{wo.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(wo.date)}</TableCell>
                      <TableCell><span className="font-mono text-xs text-muted-foreground">{wo.asset_code}</span> {wo.asset_name}</TableCell>
                      <TableCell><Badge variant="secondary">{typeLabel[wo.work_type] || wo.work_type}</Badge></TableCell>
                      <TableCell className={`text-sm ${prioColor[wo.priority] || ""}`}>{wo.priority}</TableCell>
                      <TableCell><Badge variant={statusColor[wo.status] || "secondary"}>{wo.status}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {workOrders.length === 0 && (
                    <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin órdenes de trabajo.</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          {/* ── Fuel tab ─────────────────────────────────────────────────────── */}
          <TabsContent value="fuel">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setFuelOpen(true)}><PlusIcon className="size-4 mr-1.5" />Registrar carga</Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow>
                <TableHead className="w-28">Fecha</TableHead><TableHead>Equipo</TableHead>
                <TableHead className="text-right w-20">Litros</TableHead>
                <TableHead className="text-right w-24">Km</TableHead>
                <TableHead className="text-right w-28">Costo</TableHead>
              </TableRow></TableHeader>
                <TableBody>
                  {fuelLogs.map((fl) => (
                    <TableRow key={fl.id}>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(fl.date)}</TableCell>
                      <TableCell><span className="font-mono text-xs text-muted-foreground">{fl.asset_code}</span> {fl.asset_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtNumber(fl.liters)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fl.km_reading || "\u2014"}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fl.cost ? `$${fmtNumber(fl.cost)}` : "\u2014"}</TableCell>
                    </TableRow>
                  ))}
                  {fuelLogs.length === 0 && (
                    <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin registros de combustible.</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          {/* ── Customer Vehicles tab ────────────────────────────────────────── */}
          <TabsContent value="vehicles">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setVehicleOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nuevo vehículo</Button>
            </div>
            <div className="flex gap-6">
              {/* Vehicle list */}
              <div className="flex-1 min-w-0 rounded-xl border border-border/40 bg-card overflow-hidden">
                <Table>
                  <TableHeader><TableRow>
                    <TableHead className="w-28">Patente</TableHead>
                    <TableHead>Marca</TableHead>
                    <TableHead className="w-16 text-right">Año</TableHead>
                    <TableHead className="w-16 text-right">Asientos</TableHead>
                    <TableHead>Cliente</TableHead>
                    <TableHead>Destino</TableHead>
                    <TableHead className="w-24">Combustible</TableHead>
                    <TableHead className="w-20">Estado</TableHead>
                  </TableRow></TableHeader>
                  <TableBody>
                    {vehicles.map((v) => (
                      <TableRow
                        key={v.id}
                        className={`cursor-pointer ${selectedVehicleId === v.id ? "bg-muted/50" : ""}`}
                        onClick={() => setSelectedVehicleId(selectedVehicleId === v.id ? null : v.id)}
                      >
                        <TableCell className="font-mono text-sm font-medium">{v.plate}</TableCell>
                        <TableCell className="text-sm">{v.brand || "\u2014"}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{v.model_year ?? "\u2014"}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{v.seating_capacity || "\u2014"}</TableCell>
                        <TableCell className="text-sm">{v.owner_name || "\u2014"}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{v.destination || "\u2014"}</TableCell>
                        <TableCell>
                          {v.fuel_type ? (
                            <Badge variant="outline" className="text-xs">{fuelLabel[v.fuel_type] || v.fuel_type}</Badge>
                          ) : "\u2014"}
                        </TableCell>
                        <TableCell>
                          <Badge variant={v.active ? "default" : "secondary"}>{v.active ? "Activo" : "Inactivo"}</Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                    {vehicles.length === 0 && (
                      <TableRow><TableCell colSpan={8} className="h-24 text-center text-muted-foreground">Sin vehículos registrados.</TableCell></TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>

              {/* Incident side panel */}
              {selectedVehicle && (
                <div className="w-96 shrink-0 rounded-xl border border-border/40 bg-card flex flex-col">
                  <div className="p-5 border-b border-border/40">
                    <div className="flex items-start justify-between">
                      <div>
                        <h3 className="font-semibold font-mono">{selectedVehicle.plate}</h3>
                        <p className="text-sm text-muted-foreground mt-0.5">
                          {selectedVehicle.brand}{selectedVehicle.model_year ? ` · ${selectedVehicle.model_year}` : ""}{selectedVehicle.owner_name ? ` · ${selectedVehicle.owner_name}` : ""}
                        </p>
                      </div>
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setIncidentOpen(true)}
                        >
                          <PlusIcon className="size-3.5 mr-1" />Nueva novedad
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="size-7 text-muted-foreground"
                          onClick={() => setSelectedVehicleId(null)}
                        >
                          <XIcon className="size-4" />
                        </Button>
                      </div>
                    </div>
                  </div>

                  <div className="flex-1 overflow-y-auto p-5">
                    <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">Novedades</p>
                    {incidents.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-8">Sin novedades registradas.</p>
                    ) : (
                      <div className="space-y-3">
                        {incidents.map((inc) => (
                          <div key={inc.id} className="rounded-lg border border-border/40 bg-background p-3">
                            <div className="flex items-start justify-between gap-2 mb-1">
                              <div className="flex items-center gap-2 flex-wrap">
                                <Badge variant={incidentStatusVariant[inc.status] || "secondary"} className="text-xs">
                                  {incidentStatusLabel[inc.status] || inc.status}
                                </Badge>
                                {inc.incident_type_name && (
                                  <span className="text-xs text-muted-foreground">{inc.incident_type_name}</span>
                                )}
                              </div>
                              {inc.status === "pending" && (
                                <Button
                                  size="sm"
                                  variant="outline"
                                  className="h-6 text-xs px-2 shrink-0"
                                  disabled={resolveIncidentMutation.isPending}
                                  onClick={() => resolveIncidentMutation.mutate(inc.id)}
                                >
                                  <CheckCircleIcon className="size-3 mr-1" />Resolver
                                </Button>
                              )}
                            </div>
                            <p className="text-xs text-muted-foreground">{fmtDateShort(inc.incident_date)}{inc.location ? ` · ${inc.location}` : ""}{inc.responsible ? ` · ${inc.responsible}` : ""}</p>
                            {inc.notes && <p className="text-sm mt-1">{inc.notes}</p>}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* ── Work Order dialog ────────────────────────────────────────────────── */}
      <Dialog open={woOpen} onOpenChange={(v) => !v && setWoOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Orden de Trabajo</DialogTitle></DialogHeader>
          <CreateWorkOrderForm
            assets={assets}
            onSubmit={(data) => createWorkOrderMutation.mutate(data)}
            isPending={createWorkOrderMutation.isPending}
            onClose={() => setWoOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* ── Fuel log dialog ──────────────────────────────────────────────────── */}
      <Dialog open={fuelOpen} onOpenChange={(v) => !v && setFuelOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Registrar Carga de Combustible</DialogTitle></DialogHeader>
          <CreateFuelLogForm
            assets={assets}
            onSubmit={(data) => createFuelLogMutation.mutate(data)}
            isPending={createFuelLogMutation.isPending}
            onClose={() => setFuelOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* ── New vehicle dialog ───────────────────────────────────────────────── */}
      <Dialog open={vehicleOpen} onOpenChange={(v) => !v && setVehicleOpen(false)}>
        <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
          <DialogHeader><DialogTitle>Nuevo Vehículo de Cliente</DialogTitle></DialogHeader>
          <CreateVehicleForm
            entities={entities}
            onSubmit={(data) => createVehicleMutation.mutate(data)}
            isPending={createVehicleMutation.isPending}
            onClose={() => setVehicleOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* ── New incident dialog ──────────────────────────────────────────────── */}
      <Dialog open={incidentOpen} onOpenChange={(v) => !v && setIncidentOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Novedad</DialogTitle></DialogHeader>
          {selectedVehicleId && (
            <CreateIncidentForm
              vehicleId={selectedVehicleId}
              incidentTypes={incidentTypes}
              onSubmit={(data) => createIncidentMutation.mutate(data)}
              isPending={createIncidentMutation.isPending}
              onClose={() => setIncidentOpen(false)}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ── Create Work Order form ─────────────────────────────────────────────────────
function CreateWorkOrderForm({
  assets,
  onSubmit,
  isPending,
  onClose,
}: {
  assets: Asset[];
  onSubmit: (data: { Number: string; AssetID: string; WorkType: string; Description: string; Date: string; Priority: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [number, setNumber] = useState("");
  const [assetID, setAssetID] = useState("");
  const [workType, setWorkType] = useState("preventive");
  const [description, setDescription] = useState("");
  const [date, setDate] = useState(today);
  const [priority, setPriority] = useState("normal");

  const canSubmit = number && assetID && description && date;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (canSubmit) onSubmit({ Number: number, AssetID: assetID, WorkType: workType, Description: description, Date: date, Priority: priority });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número OT</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="OT-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Equipo</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={assetID}
          onChange={(e) => setAssetID(e.target.value)}
        >
          <option value="">Seleccionar equipo...</option>
          {assets.map((a) => (
            <option key={a.id} value={a.id}>{a.code} — {a.name}</option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Tipo</Label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm bg-card"
            value={workType}
            onChange={(e) => setWorkType(e.target.value)}
          >
            <option value="preventive">Preventivo</option>
            <option value="corrective">Correctivo</option>
            <option value="inspection">Inspección</option>
          </select>
        </div>
        <div className="space-y-2">
          <Label>Prioridad</Label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm bg-card"
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
          >
            <option value="low">Baja</option>
            <option value="normal">Normal</option>
            <option value="high">Alta</option>
            <option value="urgent">Urgente</option>
          </select>
        </div>
      </div>
      <div className="space-y-2">
        <Label>Descripción</Label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Descripción del trabajo" />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}

// ── Create Fuel Log form ───────────────────────────────────────────────────────
function CreateFuelLogForm({
  assets,
  onSubmit,
  isPending,
  onClose,
}: {
  assets: Asset[];
  onSubmit: (data: { AssetID: string; Date: string; Liters: string; Cost: string; KmReading: number | null }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [assetID, setAssetID] = useState("");
  const [date, setDate] = useState(today);
  const [liters, setLiters] = useState("");
  const [cost, setCost] = useState("");
  const [kmReading, setKmReading] = useState("");

  const canSubmit = assetID && date && liters && cost;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (canSubmit) {
          const km = kmReading ? parseInt(kmReading, 10) : null;
          onSubmit({ AssetID: assetID, Date: date, Liters: liters, Cost: cost, KmReading: km });
        }
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Equipo</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={assetID}
          onChange={(e) => setAssetID(e.target.value)}
        >
          <option value="">Seleccionar equipo...</option>
          {assets.map((a) => (
            <option key={a.id} value={a.id}>{a.code} — {a.name}</option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label>Fecha</Label>
        <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Litros</Label>
          <Input type="number" step="0.01" min="0" value={liters} onChange={(e) => setLiters(e.target.value)} placeholder="0.00" />
        </div>
        <div className="space-y-2">
          <Label>Costo</Label>
          <Input type="number" step="0.01" min="0" value={cost} onChange={(e) => setCost(e.target.value)} placeholder="0.00" />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Km lectura <span className="text-muted-foreground">(opcional)</span></Label>
        <Input type="number" min="0" value={kmReading} onChange={(e) => setKmReading(e.target.value)} placeholder="150000" />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}

// ── Create Vehicle form ────────────────────────────────────────────────────────
function CreateVehicleForm({
  entities,
  onSubmit,
  isPending,
  onClose,
}: {
  entities: Entity[];
  onSubmit: (data: {
    plate: string; brand: string; model_year: number | null; seating_capacity: number | null;
    owner_id: string; chassis_serial: string; body_serial: string;
    fuel_type: string; destination: string; observations: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [plate, setPlate] = useState("");
  const [brand, setBrand] = useState("");
  const [modelYear, setModelYear] = useState("");
  const [seatingCapacity, setSeatingCapacity] = useState("");
  const [ownerId, setOwnerId] = useState("");
  const [chassisSerial, setChassisSerial] = useState("");
  const [bodySerial, setBodySerial] = useState("");
  const [fuelType, setFuelType] = useState("");
  const [destination, setDestination] = useState("");
  const [observations, setObservations] = useState("");

  const canSubmit = !!plate.trim();

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!canSubmit) return;
        onSubmit({
          plate: plate.trim().toUpperCase(),
          brand: brand.trim(),
          model_year: modelYear ? parseInt(modelYear, 10) : null,
          seating_capacity: seatingCapacity ? parseInt(seatingCapacity, 10) : null,
          owner_id: ownerId,
          chassis_serial: chassisSerial.trim(),
          body_serial: bodySerial.trim(),
          fuel_type: fuelType,
          destination: destination.trim(),
          observations: observations.trim(),
        });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Patente <span className="text-destructive">*</span></Label>
          <Input
            value={plate}
            onChange={(e) => setPlate(e.target.value)}
            placeholder="ABC123"
            className="uppercase"
          />
        </div>
        <div className="space-y-2">
          <Label>Marca</Label>
          <Input value={brand} onChange={(e) => setBrand(e.target.value)} placeholder="Mercedes-Benz" />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Año modelo</Label>
          <Input type="number" min="1900" max="2099" value={modelYear} onChange={(e) => setModelYear(e.target.value)} placeholder="2020" />
        </div>
        <div className="space-y-2">
          <Label>Asientos</Label>
          <Input type="number" min="1" value={seatingCapacity} onChange={(e) => setSeatingCapacity(e.target.value)} placeholder="40" />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Cliente / Dueño</Label>
        <Select value={ownerId} onValueChange={(v) => v && setOwnerId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar entidad..." /></SelectTrigger>
          <SelectContent>
            {entities.map((e) => (
              <SelectItem key={e.id} value={e.id}>{e.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Nro Chasis</Label>
          <Input value={chassisSerial} onChange={(e) => setChassisSerial(e.target.value)} placeholder="9BM123..." />
        </div>
        <div className="space-y-2">
          <Label>Nro Carrocería</Label>
          <Input value={bodySerial} onChange={(e) => setBodySerial(e.target.value)} placeholder="CAR-001" />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Combustible</Label>
          <Select value={fuelType} onValueChange={(v) => v && setFuelType(v)}>
            <SelectTrigger><SelectValue placeholder="Seleccionar..." /></SelectTrigger>
            <SelectContent>
              <SelectItem value="diesel">Diesel</SelectItem>
              <SelectItem value="gasolina">Gasolina</SelectItem>
              <SelectItem value="gnc">GNC</SelectItem>
              <SelectItem value="electric">Eléctrico</SelectItem>
              <SelectItem value="hybrid">Híbrido</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Destino</Label>
          <Input value={destination} onChange={(e) => setDestination(e.target.value)} placeholder="Urbano, Interurbano..." />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Notas <span className="text-muted-foreground">(opcional)</span></Label>
        <textarea
          className="w-full rounded-md border px-3 py-2 text-sm bg-card resize-none min-h-[80px]"
          value={observations}
          onChange={(e) => setObservations(e.target.value)}
          placeholder="Observaciones adicionales..."
        />
      </div>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>
          {isPending ? "Guardando..." : "Guardar"}
        </Button>
      </div>
    </form>
  );
}

// ── Create Incident form ───────────────────────────────────────────────────────
function CreateIncidentForm({
  vehicleId,
  incidentTypes,
  onSubmit,
  isPending,
  onClose,
}: {
  vehicleId: string;
  incidentTypes: IncidentType[];
  onSubmit: (data: { vehicle_id: string; incident_type_id: string; incident_date: string; location: string; responsible: string; notes: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [incidentTypeId, setIncidentTypeId] = useState("");
  const [incidentDate, setIncidentDate] = useState(today);
  const [location, setLocation] = useState("");
  const [responsible, setResponsible] = useState("");
  const [notes, setNotes] = useState("");

  const canSubmit = !!incidentDate;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!canSubmit) return;
        onSubmit({
          vehicle_id: vehicleId,
          incident_type_id: incidentTypeId,
          incident_date: incidentDate,
          location: location.trim(),
          responsible: responsible.trim(),
          notes: notes.trim(),
        });
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Tipo de novedad</Label>
        <Select value={incidentTypeId} onValueChange={(v) => v && setIncidentTypeId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar tipo..." /></SelectTrigger>
          <SelectContent>
            {incidentTypes.map((t) => (
              <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>Fecha <span className="text-destructive">*</span></Label>
        <Input type="date" value={incidentDate} onChange={(e) => setIncidentDate(e.target.value)} />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Lugar</Label>
          <Input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="Terminal Norte..." />
        </div>
        <div className="space-y-2">
          <Label>Responsable</Label>
          <Input value={responsible} onChange={(e) => setResponsible(e.target.value)} placeholder="Nombre..." />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Notas</Label>
        <textarea
          className="w-full rounded-md border px-3 py-2 text-sm bg-card resize-none min-h-[80px]"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Descripción de la novedad..."
        />
      </div>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}
