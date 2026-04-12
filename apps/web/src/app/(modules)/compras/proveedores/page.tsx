"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function ComprasProveedoresPage() {
  const { data: entities = [], isLoading, error } = useQuery({
    queryKey: erpKeys.entities("supplier"),
    queryFn: () => api.get<{ entities: any[]; total: number }>("/v1/erp/entities?type=supplier&page_size=100"),
    select: (d) => d.entities,
  });

  if (error) return <ErrorState message="Error cargando proveedores" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Proveedores</h1>
          <p className="text-sm text-muted-foreground mt-0.5">{entities.length} proveedores</p>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-28">Código</TableHead><TableHead>Nombre</TableHead>
              <TableHead>Email</TableHead><TableHead>Teléfono</TableHead>
              <TableHead className="w-24">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {entities.map((e: any) => (
                <TableRow key={e.id}>
                  <TableCell className="font-mono text-sm">{e.code}</TableCell>
                  <TableCell className="text-sm font-medium">{e.name}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{e.email || "\u2014"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{e.phone || "\u2014"}</TableCell>
                  <TableCell><Badge variant={e.active ? "default" : "secondary"}>{e.active ? "Activo" : "Inactivo"}</Badge></TableCell>
                </TableRow>
              ))}
              {entities.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin proveedores.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
