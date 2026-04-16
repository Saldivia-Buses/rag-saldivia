"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function LegajosPage() {
  const { data: employees = [], isLoading, error } = useQuery({
    queryKey: erpKeys.employees(),
    queryFn: () => api.get<{ employees: any[] }>("/v1/erp/hr/employees?page_size=100"),
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
            <TableHeader><TableRow>
              <TableHead>Legajo</TableHead><TableHead>Nombre</TableHead>
              <TableHead>Puesto</TableHead><TableHead>Horario</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {employees.map((e: any, i: number) => (
                <TableRow key={e.entity_code || i}>
                  <TableCell className="font-mono text-sm">{e.entity_code}</TableCell>
                  <TableCell className="text-sm font-medium">{e.entity_name}</TableCell>
                  <TableCell className="text-sm">{e.position || "\u2014"}</TableCell>
                  <TableCell><Badge variant="secondary">{e.schedule_type}</Badge></TableCell>
                </TableRow>
              ))}
              {employees.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin empleados.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
