"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { EmployeeDetail, EntityDetail, Attendance } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function EmployeeDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const employeeQ = useQuery({
    queryKey: erpKeys.employee(id),
    queryFn: () => api.get<EmployeeDetail>(`/v1/erp/hr/employees/${id}`),
    enabled: !!id,
  });

  const entityQ = useQuery({
    queryKey: erpKeys.entity(id),
    queryFn: () => api.get<EntityDetail>(`/v1/erp/entities/${id}`),
    enabled: !!id,
  });

  const attendanceQ = useQuery({
    queryKey: erpKeys.attendance({ entity_id: id, page_size: "50" }),
    queryFn: () =>
      api.get<{ attendance: Attendance[] }>(
        `/v1/erp/hr/attendance?entity_id=${id}&page_size=50`
      ),
    select: (d) => d.attendance,
    enabled: !!id,
  });

  if (employeeQ.error || entityQ.error)
    return <ErrorState message="Error cargando legajo" onRetry={() => window.location.reload()} />;

  const employee = employeeQ.data;
  const entity = entityQ.data?.entity;
  const attendance = attendanceQ.data ?? [];
  const contacts = entityQ.data?.contacts ?? [];
  const notes = entityQ.data?.notes ?? [];
  const isLoading = employeeQ.isLoading || entityQ.isLoading;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/rrhh/legajos"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a legajos
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {entity && employee && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{entity.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Legajo <span className="font-mono">{entity.code}</span>
                  {entity.email ? ` · ${entity.email}` : ""}
                  {entity.phone ? ` · ${entity.phone}` : ""}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {employee.termination_date && (
                  <Badge variant="outline">Baja {fmtDate(employee.termination_date)}</Badge>
                )}
                <Badge variant={entity.active ? "default" : "secondary"}>
                  {entity.active ? "Activo" : "Inactivo"}
                </Badge>
              </div>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Puesto" value={employee.position || "—"} />
              <Metric label="Horario" value={employee.schedule_type || "—"} />
              <Metric
                label="Ingreso"
                value={employee.hire_date ? fmtDate(employee.hire_date) : "—"}
              />
              <Metric label="Registros asistencia" value={String(attendance.length)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Asistencia reciente ({attendance.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[130px]">Fecha</TableHead>
                    <TableHead className="w-[100px]">Entrada</TableHead>
                    <TableHead className="w-[100px]">Salida</TableHead>
                    <TableHead className="w-[90px] text-right">Horas</TableHead>
                    <TableHead>Fuente</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {attendance.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-20 text-center text-sm text-muted-foreground">
                        Sin registros de asistencia.
                      </TableCell>
                    </TableRow>
                  )}
                  {attendance.map((a) => (
                    <TableRow key={a.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {a.date ? fmtDate(a.date) : "—"}
                      </TableCell>
                      <TableCell className="font-mono text-xs">{a.clock_in ?? "—"}</TableCell>
                      <TableCell className="font-mono text-xs">{a.clock_out ?? "—"}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{a.hours ?? "—"}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{a.source || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            {contacts.length > 0 && (
              <>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground">
                  Contactos ({contacts.length})
                </h2>
                <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[140px]">Tipo</TableHead>
                        <TableHead className="w-[180px]">Etiqueta</TableHead>
                        <TableHead>Valor</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {contacts.map((c) => (
                        <TableRow key={c.id}>
                          <TableCell className="text-sm">{c.type}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">{c.label || "—"}</TableCell>
                          <TableCell className="text-sm">{c.value}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </>
            )}

            {notes.length > 0 && (
              <>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground">Notas ({notes.length})</h2>
                <div className="space-y-2">
                  {notes.map((n) => (
                    <div key={n.id} className="rounded-lg border border-border/40 bg-card px-4 py-3">
                      <div className="text-xs text-muted-foreground">
                        {fmtDate(n.created_at)} · {n.user_id}
                      </div>
                      <p className="mt-1 text-sm whitespace-pre-wrap">{n.body}</p>
                    </div>
                  ))}
                </div>
              </>
            )}
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
