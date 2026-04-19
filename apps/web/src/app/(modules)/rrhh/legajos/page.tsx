"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { EmployeeListRow } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function LegajosPage() {
  const { data: employees = [], isLoading, error } = useQuery({
    queryKey: erpKeys.employees(),
    queryFn: () => api.get<{ employees: EmployeeListRow[] }>("/v1/erp/hr/employees?page_size=100"),
    select: (d) => d.employees,
  });

  if (error) return <ErrorState message="Error cargando legajos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Legajos</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{employees.length} empleados</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[110px]">Legajo</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead>Puesto</TableHead>
                <TableHead className="w-[120px]">Ingreso</TableHead>
                <TableHead className="w-[120px]">Horario</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {employees.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="font-mono text-sm">
                    <Link href={`/rrhh/legajos/${e.entity_id}`} className="hover:underline">
                      {e.entity_code}
                    </Link>
                  </TableCell>
                  <TableCell className="text-sm font-medium">{e.entity_name}</TableCell>
                  <TableCell className="text-sm">{e.position || "—"}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {e.hire_date ? fmtDate(e.hire_date) : "—"}
                  </TableCell>
                  <TableCell><Badge variant="secondary">{e.schedule_type}</Badge></TableCell>
                </TableRow>
              ))}
              {employees.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin empleados.</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
