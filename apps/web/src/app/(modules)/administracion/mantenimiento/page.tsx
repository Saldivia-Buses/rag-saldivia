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
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ClipboardListIcon, FuelIcon, PlusIcon } from "lucide-react";

interface WorkOrder { id: string; number: string; date: string; asset_code: string; asset_name: string; work_type: string; status: string; priority: string; }
interface FuelLog { id: string; asset_code: string; asset_name: string; date: string; liters: number; km_reading: number; cost: number; }
interface Asset { id: string; code: string; name: string; }

const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", in_progress: "outline", completed: "default", cancelled: "secondary" };
const prioColor: Record<string, string> = { low: "text-muted-foreground", normal: "", high: "text-amber-500", urgent: "text-red-500 font-medium" };
const typeLabel: Record<string, string> = { preventive: "Preventivo", corrective: "Correctivo", inspection: "Inspección" };

export default function MantenimientoPage() {
  const queryClient = useQueryClient();
  const [woOpen, setWoOpen] = useState(false);
  const [fuelOpen, setFuelOpen] = useState(false);

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

  if (error) return <ErrorState message="Error cargando mantenimiento" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Mantenimiento</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Órdenes de trabajo y combustible</p>
          </div>
          <Button size="sm" onClick={() => setWoOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva OT</Button>
        </div>
        <Tabs defaultValue="work-orders">
          <TabsList className="mb-4">
            <TabsTrigger value="work-orders"><ClipboardListIcon className="size-3.5 mr-1.5" />Órdenes de Trabajo</TabsTrigger>
            <TabsTrigger value="fuel"><FuelIcon className="size-3.5 mr-1.5" />Combustible</TabsTrigger>
          </TabsList>
          <TabsContent value="work-orders">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-20">OT</TableHead><TableHead className="w-28">Fecha</TableHead><TableHead>Equipo</TableHead><TableHead className="w-28">Tipo</TableHead><TableHead className="w-20">Prior.</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{workOrders.map((wo) => (<TableRow key={wo.id}><TableCell className="font-mono text-sm">{wo.number}</TableCell><TableCell className="text-sm text-muted-foreground">{fmtDateShort(wo.date)}</TableCell><TableCell><span className="font-mono text-xs text-muted-foreground">{wo.asset_code}</span> {wo.asset_name}</TableCell><TableCell><Badge variant="secondary">{typeLabel[wo.work_type] || wo.work_type}</Badge></TableCell><TableCell className={`text-sm ${prioColor[wo.priority] || ""}`}>{wo.priority}</TableCell><TableCell><Badge variant={statusColor[wo.status] || "secondary"}>{wo.status}</Badge></TableCell></TableRow>))}
                {workOrders.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin órdenes de trabajo.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="fuel">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setFuelOpen(true)}><PlusIcon className="size-4 mr-1.5" />Registrar carga</Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-28">Fecha</TableHead><TableHead>Equipo</TableHead><TableHead className="text-right w-20">Litros</TableHead><TableHead className="text-right w-24">Km</TableHead><TableHead className="text-right w-28">Costo</TableHead></TableRow></TableHeader>
                <TableBody>{fuelLogs.map((fl) => (<TableRow key={fl.id}><TableCell className="text-sm text-muted-foreground">{fmtDateShort(fl.date)}</TableCell><TableCell><span className="font-mono text-xs text-muted-foreground">{fl.asset_code}</span> {fl.asset_name}</TableCell><TableCell className="text-right font-mono text-sm">{fmtNumber(fl.liters)}</TableCell><TableCell className="text-right font-mono text-sm">{fl.km_reading || "\u2014"}</TableCell><TableCell className="text-right font-mono text-sm">{fl.cost ? `$${fmtNumber(fl.cost)}` : "\u2014"}</TableCell></TableRow>))}
                {fuelLogs.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin registros de combustible.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

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
    </div>
  );
}

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
