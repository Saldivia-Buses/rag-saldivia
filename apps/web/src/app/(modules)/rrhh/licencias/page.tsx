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
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { PlusIcon } from "lucide-react";

const eventLabel: Record<string, string> = {
  absence: "Falta",
  leave: "Licencia",
  accident: "Accidente",
  vacation: "Vacaciones",
  overtime: "Hora extra",
  transfer: "Transferencia",
  promotion: "Promoción",
  sanction: "Sanción",
};

export default function LicenciasPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: events = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "hr", "events"] as const,
    queryFn: () => api.get<{ events: any[] }>("/v1/erp/hr/events?page_size=50"),
    select: (d) => d.events,
  });

  const { data: employees = [] } = useQuery({
    queryKey: erpKeys.employees(),
    queryFn: () => api.get<{ employees: any[] }>("/v1/erp/hr/employees?page_size=200"),
    select: (d) => d.employees,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      entity_id: string;
      event_type: string;
      date_from: string;
      date_to?: string;
      hours?: string;
      notes?: string;
    }) => api.post("/v1/erp/hr/events", data),
    onSuccess: () => {
      toast.success("Novedad registrada exitosamente");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "hr", "events"] });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando licencias" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Licencias y Novedades</h1>
            <p className="text-sm text-muted-foreground mt-0.5">{events.length} novedades</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nueva novedad
          </Button>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead>Tipo</TableHead>
              <TableHead className="w-28">Desde</TableHead>
              <TableHead className="w-28">Hasta</TableHead>
              <TableHead className="w-20">Horas</TableHead>
              <TableHead>Notas</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {events.map((ev: any) => (
                <TableRow key={ev.id}>
                  <TableCell>
                    <Badge variant="outline">{eventLabel[ev.event_type] || ev.event_type}</Badge>
                  </TableCell>
                  <TableCell className="text-sm">{fmtDateShort(ev.date_from)}</TableCell>
                  <TableCell className="text-sm">{fmtDateShort(ev.date_to)}</TableCell>
                  <TableCell className="text-sm font-mono">{ev.hours || "\u2014"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground truncate max-w-48">
                    {ev.notes || "\u2014"}
                  </TableCell>
                </TableRow>
              ))}
              {events.length === 0 && (
                <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                  Sin novedades.
                </TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva novedad</DialogTitle></DialogHeader>
          <CreateEventForm
            employees={employees}
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateEventForm({
  employees,
  onSubmit,
  isPending,
  onClose,
}: {
  employees: any[];
  onSubmit: (data: {
    entity_id: string;
    event_type: string;
    date_from: string;
    date_to?: string;
    hours?: string;
    notes?: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [entityId, setEntityId] = useState("");
  const [eventType, setEventType] = useState("leave");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [hours, setHours] = useState("");
  const [notes, setNotes] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!entityId || !eventType || !dateFrom) return;
        onSubmit({
          entity_id: entityId,
          event_type: eventType,
          date_from: dateFrom,
          date_to: dateTo || undefined,
          hours: hours || undefined,
          notes: notes || undefined,
        });
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Empleado</Label>
        <Select value={entityId} onValueChange={(v) => setEntityId(v ?? "")}>
          <SelectTrigger className="bg-card">
            <SelectValue placeholder="Seleccionar empleado..." />
          </SelectTrigger>
          <SelectContent>
            {employees.map((e: any) => (
              <SelectItem key={e.id} value={e.id}>{e.entity_code} — {e.entity_name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Tipo de novedad</Label>
        <Select value={eventType} onValueChange={(v) => setEventType(v ?? "leave")}>
          <SelectTrigger className="bg-card">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {Object.entries(eventLabel).map(([val, label]) => (
              <SelectItem key={val} value={val}>{label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Desde</Label>
          <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Hasta <span className="text-muted-foreground">(opcional)</span></Label>
          <Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Horas <span className="text-muted-foreground">(opcional)</span></Label>
        <Input type="number" min="0" step="0.5" value={hours} onChange={(e) => setHours(e.target.value)} placeholder="8" />
      </div>
      <div className="space-y-2">
        <Label>Notas <span className="text-muted-foreground">(opcional)</span></Label>
        <Textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2} placeholder="Observaciones..." />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!entityId || !eventType || !dateFrom || isPending}>
          {isPending ? "Guardando..." : "Guardar"}
        </Button>
      </div>
    </form>
  );
}
