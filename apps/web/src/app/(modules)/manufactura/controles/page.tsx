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
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { ListChecksIcon, AlertCircleIcon, PlayCircleIcon } from "lucide-react";

interface ManufacturingUnit {
  id: string;
  work_order_number: number;
  chassis_serial: string;
  carroceria_model_name: string;
  customer_name: string;
  status: string;
}

interface ProductionControl {
  id: string;
  station: string;
  station_seq: number;
  status: string;
  planned_start: string | null;
  planned_end: string | null;
  notes: string;
}

interface PendingControl {
  id: string;
  station: string;
  status: string;
}

type ExecType = "work" | "stoppage" | "rework" | "inspection";

const execTypeOptions: { value: ExecType; label: string }[] = [
  { value: "work", label: "Trabajo" },
  { value: "stoppage", label: "Detención" },
  { value: "rework", label: "Retrabajo" },
  { value: "inspection", label: "Inspección" },
];

const controlStatusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  pending: { label: "Pendiente", variant: "secondary" },
  in_progress: { label: "En curso", variant: "outline" },
  completed: { label: "Completado", variant: "default" },
  blocked: { label: "Bloqueado", variant: "secondary" },
};

const MFG_UNITS_KEY = [...erpKeys.all, "manufacturing", "units"] as const;

function unitControlsKey(unitId: string) {
  return [...erpKeys.all, "manufacturing", "units", unitId, "controls"] as const;
}

function unitPendingControlsKey(unitId: string) {
  return [...erpKeys.all, "manufacturing", "units", unitId, "pending-controls"] as const;
}

export default function ControlesPage() {
  const [selectedUnitId, setSelectedUnitId] = useState<string | null>(null);
  const [execDialog, setExecDialog] = useState<{ controlId: string } | null>(null);
  const queryClient = useQueryClient();

  const { data: units = [], isLoading: unitsLoading, error: unitsError } = useQuery({
    queryKey: [...MFG_UNITS_KEY, { status: "in_production" }] as const,
    queryFn: () => api.get<{ units: ManufacturingUnit[] }>("/v1/erp/manufacturing/units?status=in_production&page_size=50"),
    select: (d) => d.units,
  });

  const { data: controls = [], isLoading: controlsLoading } = useQuery({
    queryKey: unitControlsKey(selectedUnitId ?? ""),
    queryFn: () => api.get<{ controls: ProductionControl[] }>(`/v1/erp/manufacturing/units/${selectedUnitId}/controls`),
    enabled: !!selectedUnitId,
    select: (d) => d.controls,
  });

  const { data: pendingControls = [] } = useQuery({
    queryKey: unitPendingControlsKey(selectedUnitId ?? ""),
    queryFn: () => api.get<{ controls: PendingControl[] }>(`/v1/erp/manufacturing/units/${selectedUnitId}/pending-controls`),
    enabled: !!selectedUnitId,
    select: (d) => d.controls,
  });

  const executeMutation = useMutation({
    mutationFn: ({ controlId, exec_type, notes }: { controlId: string; exec_type: ExecType; notes: string }) =>
      api.post(`/v1/erp/manufacturing/units/${selectedUnitId}/controls/${controlId}/execute`, { exec_type, notes }),
    onSuccess: () => {
      toast.success("Control ejecutado");
      queryClient.invalidateQueries({ queryKey: unitControlsKey(selectedUnitId ?? "") });
      queryClient.invalidateQueries({ queryKey: unitPendingControlsKey(selectedUnitId ?? "") });
      setExecDialog(null);
    },
    onError: permissionErrorToast,
  });

  if (unitsError) return <ErrorState message="Error cargando unidades" onRetry={() => window.location.reload()} />;
  if (unitsLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const selectedUnit = units.find((u) => u.id === selectedUnitId);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Controles</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Controles de producción por unidad — {units.length} unidades en producción
          </p>
        </div>

        <div className="flex gap-6">
          {/* Unit list */}
          <div className="w-72 shrink-0 rounded-xl border border-border/40 bg-card overflow-hidden self-start">
            <div className="px-4 py-3 border-b border-border/40">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">En producción</p>
            </div>
            {units.length === 0 ? (
              <p className="px-4 py-6 text-sm text-center text-muted-foreground">Sin unidades en producción.</p>
            ) : (
              <ul>
                {units.map((u) => (
                  <li key={u.id}>
                    <button
                      type="button"
                      className={`w-full text-left px-4 py-3 text-sm border-b border-border/20 last:border-0 transition-colors hover:bg-accent/40 ${selectedUnitId === u.id ? "bg-accent/60" : ""}`}
                      onClick={() => setSelectedUnitId(u.id)}
                    >
                      <p className="font-mono font-semibold">OT {u.work_order_number}</p>
                      <p className="text-xs text-muted-foreground mt-0.5 truncate">
                        {u.carroceria_model_name || u.chassis_serial || "\u2014"}
                      </p>
                      {u.customer_name && (
                        <p className="text-xs text-muted-foreground truncate">{u.customer_name}</p>
                      )}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>

          {/* Detail panel */}
          <div className="flex-1 min-w-0">
            {!selectedUnit ? (
              <div className="flex items-center justify-center h-64 rounded-xl border border-dashed border-border/40 text-muted-foreground text-sm">
                Seleccioná una unidad para ver sus controles.
              </div>
            ) : (
              <>
                <div className="mb-4">
                  <p className="font-semibold">OT {selectedUnit.work_order_number}</p>
                  <p className="text-sm text-muted-foreground">
                    {selectedUnit.carroceria_model_name || selectedUnit.chassis_serial || "\u2014"}
                    {selectedUnit.customer_name ? ` — ${selectedUnit.customer_name}` : ""}
                  </p>
                </div>

                <Tabs defaultValue="controls">
                  <TabsList className="mb-4">
                    <TabsTrigger value="controls">
                      <ListChecksIcon className="size-3.5 mr-1.5" />Controles
                    </TabsTrigger>
                    <TabsTrigger value="pending">
                      <AlertCircleIcon className="size-3.5 mr-1.5" />
                      Pendientes
                      {pendingControls.length > 0 && (
                        <span className="ml-1.5 inline-flex items-center justify-center size-4 rounded-full bg-destructive/20 text-destructive text-[10px] font-semibold">
                          {pendingControls.length}
                        </span>
                      )}
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value="controls">
                    {controlsLoading ? (
                      <Skeleton className="h-48" />
                    ) : (
                      <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead className="w-12 text-center">#</TableHead>
                              <TableHead>Estación</TableHead>
                              <TableHead className="w-32">Estado</TableHead>
                              <TableHead className="w-24">Inicio plan.</TableHead>
                              <TableHead className="w-24">Fin plan.</TableHead>
                              <TableHead>Notas</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {controls.map((c) => {
                              const s = controlStatusBadge[c.status] ?? controlStatusBadge.pending;
                              return (
                                <TableRow key={c.id}>
                                  <TableCell className="text-center font-mono text-sm">{c.station_seq}</TableCell>
                                  <TableCell className="text-sm font-medium">{c.station}</TableCell>
                                  <TableCell>
                                    <Badge variant={s.variant} className="text-xs">{s.label}</Badge>
                                  </TableCell>
                                  <TableCell className="text-sm text-muted-foreground">
                                    {c.planned_start ? fmtDateShort(c.planned_start) : "\u2014"}
                                  </TableCell>
                                  <TableCell className="text-sm text-muted-foreground">
                                    {c.planned_end ? fmtDateShort(c.planned_end) : "\u2014"}
                                  </TableCell>
                                  <TableCell className="text-sm text-muted-foreground truncate max-w-[200px]">
                                    {c.notes || "\u2014"}
                                  </TableCell>
                                </TableRow>
                              );
                            })}
                            {controls.length === 0 && (
                              <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                                  Sin controles registrados.
                                </TableCell>
                              </TableRow>
                            )}
                          </TableBody>
                        </Table>
                      </div>
                    )}
                  </TabsContent>

                  <TabsContent value="pending">
                    <div className="space-y-2">
                      {pendingControls.length === 0 ? (
                        <div className="flex items-center justify-center h-24 rounded-xl border border-dashed border-border/40 text-muted-foreground text-sm">
                          Sin controles pendientes.
                        </div>
                      ) : (
                        pendingControls.map((c) => (
                          <div key={c.id} className="flex items-center justify-between rounded-lg border border-border/40 bg-card px-4 py-3">
                            <div>
                              <p className="text-sm font-medium">{c.station}</p>
                              <p className="text-xs text-muted-foreground capitalize">{c.status}</p>
                            </div>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => setExecDialog({ controlId: c.id })}
                            >
                              <PlayCircleIcon className="size-3.5 mr-1.5" />Ejecutar
                            </Button>
                          </div>
                        ))
                      )}
                    </div>
                  </TabsContent>
                </Tabs>
              </>
            )}
          </div>
        </div>
      </div>

      {execDialog && (
        <Dialog open onOpenChange={(v) => !v && setExecDialog(null)}>
          <DialogContent className="max-w-sm">
            <DialogHeader><DialogTitle>Ejecutar control</DialogTitle></DialogHeader>
            <ExecuteControlForm
              onSubmit={(exec_type, notes) =>
                executeMutation.mutate({ controlId: execDialog.controlId, exec_type, notes })
              }
              isPending={executeMutation.isPending}
              onClose={() => setExecDialog(null)}
            />
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}

function ExecuteControlForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (exec_type: ExecType, notes: string) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [execType, setExecType] = useState<ExecType>("work");
  const [notes, setNotes] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit(execType, notes);
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Tipo de ejecución</Label>
        <Select value={execType} onValueChange={(v) => setExecType(v as ExecType)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            {execTypeOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Notas</Label>
        <Textarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Observaciones del control..."
          rows={3}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={isPending}>{isPending ? "Ejecutando..." : "Ejecutar"}</Button>
      </div>
    </form>
  );
}
