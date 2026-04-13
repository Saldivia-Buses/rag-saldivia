"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { HeartPulseIcon, PlusIcon } from "lucide-react";

interface Employee {
  id: string;
  first_name: string;
  last_name: string;
}

interface MedicalEvent {
  id: string;
  entity_id: string;
  entity_name?: string;
  event_type: string;
  date_from: string;
  date_to: string;
  notes: string;
}

function certStatus(dateTo: string): "Vigente" | "Vencido" {
  if (!dateTo) return "Vencido";
  return new Date(dateTo) >= new Date() ? "Vigente" : "Vencido";
}

export default function MedicinaPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const { data: events = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "hr", "events", "medical"] as const,
    queryFn: () => api.get<{ events: MedicalEvent[] }>("/v1/erp/hr/events?event_type=medical&page_size=50"),
    select: (d) => d.events,
  });

  const { data: employees = [] } = useQuery({
    queryKey: erpKeys.employees(),
    queryFn: () => api.get<{ employees: Employee[] }>("/v1/erp/hr/employees?page_size=200"),
    select: (d) => d.employees,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      entity_id: string;
      event_type: string;
      date_from: string;
      date_to: string;
      hours: null;
      notes: string;
    }) => api.post("/v1/erp/hr/events", data),
    onSuccess: () => {
      toast.success("Certificado registrado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "hr", "events", "medical"] });
      setOpen(false);
    },
    onError: permissionErrorToast,
  });

  const employeeMap = new Map(employees.map((e) => [e.id, `${e.first_name} ${e.last_name}`]));

  if (error) return <ErrorState message="Error cargando certificados médicos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight flex items-center gap-2">
              <HeartPulseIcon className="size-5 text-rose-500" />
              Medicina Laboral
            </h1>
            <p className="text-sm text-muted-foreground mt-0.5">Certificados psicofísicos y aptitud médica de conductores</p>
          </div>
          <Button size="sm" onClick={() => setOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo certificado
          </Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Empleado</TableHead>
                <TableHead className="w-28">Tipo</TableHead>
                <TableHead className="w-28">Válido desde</TableHead>
                <TableHead className="w-28">Válido hasta</TableHead>
                <TableHead>Notas</TableHead>
                <TableHead className="w-24">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {events.map((ev) => {
                const status = certStatus(ev.date_to);
                const empName = ev.entity_name || employeeMap.get(ev.entity_id) || ev.entity_id;
                return (
                  <TableRow key={ev.id}>
                    <TableCell className="text-sm font-medium">{empName}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">Psicofísico</Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDate(ev.date_from)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{fmtDate(ev.date_to)}</TableCell>
                    <TableCell className="text-sm truncate max-w-56">{ev.notes || "—"}</TableCell>
                    <TableCell>
                      <Badge variant={status === "Vigente" ? "default" : "destructive"}>
                        {status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                );
              })}
              {events.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    Sin certificados registrados.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={open} onOpenChange={(v) => !v && setOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo Certificado Médico</DialogTitle></DialogHeader>
          <CreateCertForm
            employees={employees}
            onSubmit={(data) =>
              createMutation.mutate({
                entity_id: data.entity_id,
                event_type: "leave",
                date_from: data.date_from,
                date_to: data.date_to,
                hours: null,
                notes: data.notes,
              })
            }
            isPending={createMutation.isPending}
            onClose={() => setOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateCertForm({
  employees,
  onSubmit,
  isPending,
  onClose,
}: {
  employees: Employee[];
  onSubmit: (data: { entity_id: string; date_from: string; date_to: string; notes: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [entityId, setEntityId] = useState("");
  const [dateFrom, setDateFrom] = useState(today);
  const [dateTo, setDateTo] = useState("");
  const [notes, setNotes] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (entityId && dateFrom && dateTo) {
          onSubmit({ entity_id: entityId, date_from: dateFrom, date_to: dateTo, notes });
        }
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Empleado</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={entityId}
          onChange={(e) => setEntityId(e.target.value)}
        >
          <option value="">Seleccionar empleado...</option>
          {employees.map((emp) => (
            <option key={emp.id} value={emp.id}>
              {emp.first_name} {emp.last_name}
            </option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Válido desde</Label>
          <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Válido hasta</Label>
          <Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Notas</Label>
        <Textarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Ej: Apto psicofísico certificado Nro 1234"
          rows={2}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!entityId || !dateFrom || !dateTo || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}
