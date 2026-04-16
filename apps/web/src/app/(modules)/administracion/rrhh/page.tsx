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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { PlusIcon, UsersIcon, CalendarIcon, BookOpenIcon } from "lucide-react";

interface Employee { entity_name: string; entity_code: string; position: string; department_id: string; schedule_type: string; }
interface EmployeeWithId { id: string; entity_name: string; entity_code: string; }
interface HREvent { id: string; entity_id: string; event_type: string; date_from: string; date_to: string; hours: number; notes: string; }
interface Training { id: string; name: string; instructor: string; date_from: string; status: string; }

type EventType = "absence" | "leave" | "overtime" | "vacation" | "sanction";

const eventLabel: Record<string, string> = { absence: "Falta", leave: "Licencia", accident: "Accidente", transfer: "Traslado", promotion: "Ascenso", sanction: "Sanción", overtime: "Hora extra", vacation: "Vacaciones" };

const eventTypeOptions: { value: EventType; label: string }[] = [
  { value: "absence", label: "Falta" },
  { value: "leave", label: "Licencia" },
  { value: "overtime", label: "Hora extra" },
  { value: "vacation", label: "Vacaciones" },
  { value: "sanction", label: "Sanción" },
];

export default function RRHHPage() {
  const [eventOpen, setEventOpen] = useState(false);
  const [trainingOpen, setTrainingOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: employees = [], isLoading, error } = useQuery({
    queryKey: erpKeys.employees(),
    queryFn: () => api.get<{ employees: Employee[] }>("/v1/erp/hr/employees?page_size=100"),
    select: (d) => d.employees,
  });

  const { data: employeesWithId = [] } = useQuery({
    queryKey: [...erpKeys.all, "hr", "employees-full"] as const,
    queryFn: () => api.get<{ employees: EmployeeWithId[] }>("/v1/erp/hr/employees?page_size=200"),
    select: (d) => d.employees,
  });

  const { data: events = [] } = useQuery({
    queryKey: [...erpKeys.all, "hr", "events"] as const,
    queryFn: () => api.get<{ events: HREvent[] }>("/v1/erp/hr/events?page_size=50"),
    select: (d) => d.events,
  });

  const { data: training = [] } = useQuery({
    queryKey: [...erpKeys.all, "hr", "training"] as const,
    queryFn: () => api.get<{ training: Training[] }>("/v1/erp/hr/training?page_size=50"),
    select: (d) => d.training,
  });

  const createEventMutation = useMutation({
    mutationFn: (data: { entity_id: string; event_type: EventType; date_from: string; date_to: string; hours: number | null; notes: string }) =>
      api.post("/v1/erp/hr/events", data),
    onSuccess: () => {
      toast.success("Novedad registrada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "hr", "events"] });
      setEventOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createTrainingMutation = useMutation({
    mutationFn: (data: { name: string; instructor: string; date_from: string; date_to: string }) =>
      api.post("/v1/erp/hr/training", data),
    onSuccess: () => {
      toast.success("Capacitación creada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "hr", "training"] });
      setTrainingOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando recursos humanos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Recursos Humanos</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Empleados, novedades, asistencia y capacitación</p>
        </div>
        <Tabs defaultValue="employees">
          <TabsList className="mb-4">
            <TabsTrigger value="employees"><UsersIcon className="size-3.5 mr-1.5" />Empleados</TabsTrigger>
            <TabsTrigger value="events"><CalendarIcon className="size-3.5 mr-1.5" />Novedades</TabsTrigger>
            <TabsTrigger value="training"><BookOpenIcon className="size-3.5 mr-1.5" />Capacitación</TabsTrigger>
          </TabsList>

          <TabsContent value="employees">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Legajo</TableHead><TableHead>Nombre</TableHead><TableHead>Puesto</TableHead><TableHead>Horario</TableHead></TableRow></TableHeader>
                <TableBody>{employees.map((e, i) => (<TableRow key={i}><TableCell className="font-mono text-sm">{e.entity_code}</TableCell><TableCell className="text-sm">{e.entity_name}</TableCell><TableCell className="text-sm">{e.position || "\u2014"}</TableCell><TableCell><Badge variant="secondary">{e.schedule_type}</Badge></TableCell></TableRow>))}
                {employees.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin empleados registrados.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>

          <TabsContent value="events">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setEventOpen(true)}><PlusIcon className="size-4 mr-1.5" />Registrar novedad</Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Tipo</TableHead><TableHead>Desde</TableHead><TableHead>Hasta</TableHead><TableHead>Hs</TableHead><TableHead>Notas</TableHead></TableRow></TableHeader>
                <TableBody>{events.map((ev) => (<TableRow key={ev.id}><TableCell><Badge variant="outline">{eventLabel[ev.event_type] || ev.event_type}</Badge></TableCell><TableCell className="text-sm">{fmtDateShort(ev.date_from)}</TableCell><TableCell className="text-sm">{fmtDateShort(ev.date_to)}</TableCell><TableCell className="text-sm font-mono">{ev.hours || "\u2014"}</TableCell><TableCell className="text-sm text-muted-foreground truncate max-w-48">{ev.notes || "\u2014"}</TableCell></TableRow>))}
                {events.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin novedades.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>

          <TabsContent value="training">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setTrainingOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva capacitación</Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Curso</TableHead><TableHead>Instructor</TableHead><TableHead>Fecha</TableHead><TableHead>Estado</TableHead></TableRow></TableHeader>
                <TableBody>{training.map((t) => (<TableRow key={t.id}><TableCell className="text-sm">{t.name}</TableCell><TableCell className="text-sm">{t.instructor || "\u2014"}</TableCell><TableCell className="text-sm">{fmtDateShort(t.date_from)}</TableCell><TableCell><Badge variant="secondary">{t.status}</Badge></TableCell></TableRow>))}
                {training.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin cursos.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={eventOpen} onOpenChange={(v) => !v && setEventOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Registrar novedad</DialogTitle></DialogHeader>
          <CreateHREventForm
            employees={employeesWithId}
            onSubmit={(data) => createEventMutation.mutate(data)}
            isPending={createEventMutation.isPending}
            onClose={() => setEventOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={trainingOpen} onOpenChange={(v) => !v && setTrainingOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva capacitación</DialogTitle></DialogHeader>
          <CreateTrainingForm
            onSubmit={(data) => createTrainingMutation.mutate(data)}
            isPending={createTrainingMutation.isPending}
            onClose={() => setTrainingOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateHREventForm({
  employees,
  onSubmit,
  isPending,
  onClose,
}: {
  employees: EmployeeWithId[];
  onSubmit: (data: { entity_id: string; event_type: EventType; date_from: string; date_to: string; hours: number | null; notes: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [entityId, setEntityId] = useState("");
  const [eventType, setEventType] = useState<EventType>("absence");
  const [dateFrom, setDateFrom] = useState(today);
  const [dateTo, setDateTo] = useState(today);
  const [hours, setHours] = useState("");
  const [notes, setNotes] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!entityId || !eventType || !dateFrom || !dateTo) return;
    onSubmit({
      entity_id: entityId,
      event_type: eventType,
      date_from: dateFrom,
      date_to: dateTo,
      hours: hours ? parseFloat(hours) : null,
      notes,
    });
  }

  const canSubmit = entityId && eventType && dateFrom && dateTo;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>Empleado</Label>
        <Select value={entityId} onValueChange={(v) => setEntityId(v ?? "")}>
          <SelectTrigger><SelectValue placeholder="Seleccionar empleado..." /></SelectTrigger>
          <SelectContent>
            {employees.map((emp) => (
              <SelectItem key={emp.id} value={emp.id}>{emp.entity_name} — {emp.entity_code}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Tipo de novedad</Label>
        <Select value={eventType} onValueChange={(v) => setEventType(v as EventType)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            {eventTypeOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Desde</Label><Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} /></div>
        <div className="space-y-2"><Label>Hasta</Label><Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} /></div>
      </div>
      <div className="space-y-2"><Label>Horas (opcional)</Label><Input type="number" min="0" step="0.5" placeholder="—" value={hours} onChange={(e) => setHours(e.target.value)} /></div>
      <div className="space-y-2"><Label>Notas</Label><Textarea rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} /></div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Guardando..." : "Registrar"}</Button>
      </div>
    </form>
  );
}

function CreateTrainingForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { name: string; instructor: string; date_from: string; date_to: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [name, setName] = useState("");
  const [instructor, setInstructor] = useState("");
  const [dateFrom, setDateFrom] = useState(today);
  const [dateTo, setDateTo] = useState(today);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name || !dateFrom || !dateTo) return;
    onSubmit({ name, instructor, date_from: dateFrom, date_to: dateTo });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2"><Label>Nombre del curso</Label><Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Seguridad industrial" /></div>
      <div className="space-y-2"><Label>Instructor</Label><Input value={instructor} onChange={(e) => setInstructor(e.target.value)} placeholder="Nombre del instructor" /></div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Desde</Label><Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} /></div>
        <div className="space-y-2"><Label>Hasta</Label><Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} /></div>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!name || !dateFrom || !dateTo || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
