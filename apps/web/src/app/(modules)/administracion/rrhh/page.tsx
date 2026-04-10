"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UsersIcon, CalendarIcon, BookOpenIcon, ClockIcon } from "lucide-react";

interface Employee { entity_name: string; entity_code: string; position: string; department_id: string; schedule_type: string; }
interface HREvent { id: string; entity_id: string; event_type: string; date_from: string; date_to: string; hours: number; notes: string; }
interface Training { id: string; name: string; instructor: string; date_from: string; status: string; }

const fmtDate = (s: string) => s ? new Date(s).toLocaleDateString("es-AR", { day: "2-digit", month: "short" }) : "—";
const eventLabel: Record<string, string> = { absence: "Falta", leave: "Licencia", accident: "Accidente", transfer: "Traslado", promotion: "Ascenso", sanction: "Sancion", overtime: "Hora extra", vacation: "Vacaciones" };

export default function RRHHPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [events, setEvents] = useState<HREvent[]>([]);
  const [training, setTraining] = useState<Training[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const [e, ev, t] = await Promise.all([
        api.get<{ employees: Employee[] }>("/v1/erp/hr/employees?page_size=100"),
        api.get<{ events: HREvent[] }>("/v1/erp/hr/events?page_size=50"),
        api.get<{ training: Training[] }>("/v1/erp/hr/training?page_size=50"),
      ]);
      setEmployees(e.employees); setEvents(ev.events); setTraining(t.training);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);
  useEffect(() => { const unsub = wsManager.subscribe("erp_hr", fetch); return unsub; }, [fetch]);

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Recursos Humanos</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Empleados, novedades, asistencia y capacitacion</p>
        </div>
        <Tabs defaultValue="employees">
          <TabsList className="mb-4">
            <TabsTrigger value="employees"><UsersIcon className="size-3.5 mr-1.5" />Empleados</TabsTrigger>
            <TabsTrigger value="events"><CalendarIcon className="size-3.5 mr-1.5" />Novedades</TabsTrigger>
            <TabsTrigger value="training"><BookOpenIcon className="size-3.5 mr-1.5" />Capacitacion</TabsTrigger>
          </TabsList>
          <TabsContent value="employees">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Legajo</TableHead><TableHead>Nombre</TableHead><TableHead>Puesto</TableHead><TableHead>Horario</TableHead></TableRow></TableHeader>
                <TableBody>{employees.map((e, i) => (<TableRow key={i}><TableCell className="font-mono text-sm">{e.entity_code}</TableCell><TableCell className="text-sm">{e.entity_name}</TableCell><TableCell className="text-sm">{e.position || "—"}</TableCell><TableCell><Badge variant="secondary">{e.schedule_type}</Badge></TableCell></TableRow>))}
                {employees.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin empleados registrados.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="events">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Tipo</TableHead><TableHead>Desde</TableHead><TableHead>Hasta</TableHead><TableHead>Hs</TableHead><TableHead>Notas</TableHead></TableRow></TableHeader>
                <TableBody>{events.map((ev) => (<TableRow key={ev.id}><TableCell><Badge variant="outline">{eventLabel[ev.event_type] || ev.event_type}</Badge></TableCell><TableCell className="text-sm">{fmtDate(ev.date_from)}</TableCell><TableCell className="text-sm">{fmtDate(ev.date_to)}</TableCell><TableCell className="text-sm font-mono">{ev.hours || "—"}</TableCell><TableCell className="text-sm text-muted-foreground truncate max-w-48">{ev.notes || "—"}</TableCell></TableRow>))}
                {events.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin novedades.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="training">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead>Curso</TableHead><TableHead>Instructor</TableHead><TableHead>Fecha</TableHead><TableHead>Estado</TableHead></TableRow></TableHeader>
                <TableBody>{training.map((t) => (<TableRow key={t.id}><TableCell className="text-sm">{t.name}</TableCell><TableCell className="text-sm">{t.instructor || "—"}</TableCell><TableCell className="text-sm">{fmtDate(t.date_from)}</TableCell><TableCell><Badge variant="secondary">{t.status}</Badge></TableCell></TableRow>))}
                {training.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin cursos.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
