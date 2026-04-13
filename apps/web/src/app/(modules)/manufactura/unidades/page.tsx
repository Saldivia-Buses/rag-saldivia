"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
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
import { PlusIcon } from "lucide-react";

interface ManufacturingUnit {
  id: string;
  work_order_number: number;
  chassis_serial: string;
  engine_number: string;
  chassis_brand_name: string;
  chassis_model_name: string;
  carroceria_model_name: string;
  customer_name: string;
  entry_date: string;
  expected_completion: string | null;
  status: string;
  observations: string;
}

interface CarroceriaModel {
  id: string;
  code: string;
  model_code: string;
  description: string;
}

interface ChassisBrand {
  id: string;
  code: string;
  name: string;
}

interface ChassisModel {
  id: string;
  brand_id: string;
  code: string;
  name: string;
}

interface Entity {
  id: string;
  name: string;
}

type UnitStatus = "pending" | "in_production" | "completed" | "delivered" | "returned";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  pending: { label: "Pendiente", variant: "secondary" },
  in_production: { label: "En producción", variant: "outline" },
  completed: { label: "Terminada", variant: "default" },
  delivered: { label: "Entregada", variant: "default" },
  returned: { label: "Devuelta", variant: "secondary" },
};

const statusOptions: { value: UnitStatus; label: string }[] = [
  { value: "pending", label: "Pendiente" },
  { value: "in_production", label: "En producción" },
  { value: "completed", label: "Terminada" },
  { value: "delivered", label: "Entregada" },
  { value: "returned", label: "Devuelta" },
];

const MFG_UNITS_KEY = [...erpKeys.all, "manufacturing", "units"] as const;

export default function UnidadesPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: units = [], isLoading, error } = useQuery({
    queryKey: MFG_UNITS_KEY,
    queryFn: () => api.get<{ units: ManufacturingUnit[] }>("/v1/erp/manufacturing/units?page=1&page_size=50"),
    select: (d) => d.units,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      work_order_number: number;
      chassis_serial: string;
      engine_number: string;
      chassis_brand_id?: string;
      chassis_model_id?: string;
      carroceria_model_id?: string;
      customer_id?: string;
      entry_date: string;
      expected_completion?: string;
    }) => api.post("/v1/erp/manufacturing/units", data),
    onSuccess: () => {
      toast.success("Unidad registrada");
      queryClient.invalidateQueries({ queryKey: MFG_UNITS_KEY });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: UnitStatus }) =>
      api.patch(`/v1/erp/manufacturing/units/${id}/status`, { status }),
    onSuccess: () => {
      toast.success("Estado actualizado");
      queryClient.invalidateQueries({ queryKey: MFG_UNITS_KEY });
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando unidades" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Unidades</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Buses en producción — {units.length} unidades
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nueva unidad
          </Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">OT</TableHead>
                <TableHead className="w-36">Chasis</TableHead>
                <TableHead className="w-36">Motor</TableHead>
                <TableHead>Carrocería</TableHead>
                <TableHead>Cliente</TableHead>
                <TableHead className="w-24">Ingreso</TableHead>
                <TableHead className="w-24">Vencimiento</TableHead>
                <TableHead className="w-44">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {units.map((u) => {
                const s = statusBadge[u.status] ?? statusBadge.pending;
                return (
                  <TableRow key={u.id}>
                    <TableCell className="font-mono text-sm">{u.work_order_number}</TableCell>
                    <TableCell className="font-mono text-sm">{u.chassis_serial || "\u2014"}</TableCell>
                    <TableCell className="font-mono text-sm">{u.engine_number || "\u2014"}</TableCell>
                    <TableCell className="text-sm">{u.carroceria_model_name || "\u2014"}</TableCell>
                    <TableCell className="text-sm">{u.customer_name || "\u2014"}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDateShort(u.entry_date)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {u.expected_completion ? fmtDateShort(u.expected_completion) : "\u2014"}
                    </TableCell>
                    <TableCell>
                      <Select
                        value={u.status}
                        onValueChange={(v) => updateStatusMutation.mutate({ id: u.id, status: v as UnitStatus })}
                      >
                        <SelectTrigger className="h-7 text-xs w-36">
                          <Badge variant={s.variant} className="text-xs">{s.label}</Badge>
                        </SelectTrigger>
                        <SelectContent>
                          {statusOptions.map((opt) => (
                            <SelectItem key={opt.value} value={opt.value} className="text-xs">
                              {opt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                  </TableRow>
                );
              })}
              {units.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8} className="h-24 text-center text-muted-foreground">
                    Sin unidades registradas.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva unidad</DialogTitle></DialogHeader>
          <CreateUnitForm
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateUnitForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: {
    work_order_number: number;
    chassis_serial: string;
    engine_number: string;
    chassis_brand_id?: string;
    chassis_model_id?: string;
    carroceria_model_id?: string;
    customer_id?: string;
    entry_date: string;
    expected_completion?: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [workOrderNumber, setWorkOrderNumber] = useState("");
  const [chassisSerial, setChassisSerial] = useState("");
  const [engineNumber, setEngineNumber] = useState("");
  const [carroceriaModelId, setCarroceriaModelId] = useState("");
  const [customerId, setCustomerId] = useState("");
  const [entryDate, setEntryDate] = useState(today);
  const [expectedCompletion, setExpectedCompletion] = useState("");

  const { data: carroceriaModels = [] } = useQuery({
    queryKey: [...erpKeys.all, "manufacturing", "carroceria-models"] as const,
    queryFn: () => api.get<{ models: CarroceriaModel[] }>("/v1/erp/manufacturing/carroceria-models"),
    select: (d) => d.models,
  });

  const { data: customers = [] } = useQuery({
    queryKey: erpKeys.entities(undefined, undefined),
    queryFn: () => api.get<{ entities: Entity[] }>("/v1/erp/entities?page_size=200"),
    select: (d) => d.entities,
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const ot = parseInt(workOrderNumber);
    if (!ot) return;
    const payload: Parameters<typeof onSubmit>[0] = {
      work_order_number: ot,
      chassis_serial: chassisSerial,
      engine_number: engineNumber,
      entry_date: entryDate,
    };
    if (carroceriaModelId) payload.carroceria_model_id = carroceriaModelId;
    if (customerId) payload.customer_id = customerId;
    if (expectedCompletion) payload.expected_completion = expectedCompletion;
    onSubmit(payload);
  }

  const canSubmit = !!workOrderNumber && !!entryDate;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>Nro OT <span className="text-destructive">*</span></Label>
        <Input
          type="number"
          min="1"
          value={workOrderNumber}
          onChange={(e) => setWorkOrderNumber(e.target.value)}
          placeholder="1001"
        />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Nro chasis</Label>
          <Input value={chassisSerial} onChange={(e) => setChassisSerial(e.target.value)} placeholder="ABC123456" />
        </div>
        <div className="space-y-2">
          <Label>Nro motor</Label>
          <Input value={engineNumber} onChange={(e) => setEngineNumber(e.target.value)} placeholder="MOT-789" />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Carrocería</Label>
        <Select value={carroceriaModelId} onValueChange={(v) => v && setCarroceriaModelId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar modelo..." /></SelectTrigger>
          <SelectContent>
            {(carroceriaModels as CarroceriaModel[]).map((m) => (
              <SelectItem key={m.id} value={m.id}>
                {m.model_code} — {m.description}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Cliente</Label>
        <Select value={customerId} onValueChange={(v) => v && setCustomerId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar cliente..." /></SelectTrigger>
          <SelectContent>
            {(customers as Entity[]).map((c) => (
              <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Fecha ingreso <span className="text-destructive">*</span></Label>
          <Input type="date" value={entryDate} onChange={(e) => setEntryDate(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Fecha est. entrega <span className="text-muted-foreground">(opcional)</span></Label>
          <Input type="date" value={expectedCompletion} onChange={(e) => setExpectedCompletion(e.target.value)} />
        </div>
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
