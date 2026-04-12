"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UsersIcon, CalendarIcon, BookOpenIcon } from "lucide-react";

interface Employee { entity_name: string; entity_code: string; position: string; department_id: string; schedule_type: string; }
interface HREvent { id: string; entity_id: string; event_type: string; date_from: string; date_to: string; hours: number; notes: string; }
interface Training { id: string; name: string; instructor: string; date_from: string; status: string; }

const eventLabel: Record<string, string> = { absence: "Falta", leave: "Licencia", accident: "Accidente", transfer: "Traslado", promotion: "Ascenso", sanction: "Sanción", overtime: "Hora extra", vacation: "Vacaciones" };

export default function RRHHPage() {
  const { data: employees = [], isLoading, error } = useQuery({
    queryKey: erpKeys.employees(),
    queryFn: () => api.get<{ employees: Employee[] }>("/v1/erp/hr/employees?page_size=100"),
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
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Tipo</TableHead><TableHead>Desde</TableHead><TableHead>Hasta</TableHead><TableHead>Hs</TableHead><TableHead>Notas</TableHead></TableRow></TableHeader>
                <TableBody>{events.map((ev) => (<TableRow key={ev.id}><TableCell><Badge variant="outline">{eventLabel[ev.event_type] || ev.event_type}</Badge></TableCell><TableCell className="text-sm">{fmtDateShort(ev.date_from)}</TableCell><TableCell className="text-sm">{fmtDateShort(ev.date_to)}</TableCell><TableCell className="text-sm font-mono">{ev.hours || "\u2014"}</TableCell><TableCell className="text-sm text-muted-foreground truncate max-w-48">{ev.notes || "\u2014"}</TableCell></TableRow>))}
                {events.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin novedades.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="training">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Curso</TableHead><TableHead>Instructor</TableHead><TableHead>Fecha</TableHead><TableHead>Estado</TableHead></TableRow></TableHeader>
                <TableBody>{training.map((t) => (<TableRow key={t.id}><TableCell className="text-sm">{t.name}</TableCell><TableCell className="text-sm">{t.instructor || "\u2014"}</TableCell><TableCell className="text-sm">{fmtDateShort(t.date_from)}</TableCell><TableCell><Badge variant="secondary">{t.status}</Badge></TableCell></TableRow>))}
                {training.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin cursos.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
