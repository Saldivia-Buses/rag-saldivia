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
import { ClipboardCheckIcon, PlusIcon } from "lucide-react";

interface VehicleInspection {
  id: string;
  number: string;
  date: string;
  audit_type: string;
  scope: string;
  status: string;
}

const statusLabel: Record<string, string> = {
  planned: "Planificada",
  in_progress: "En curso",
  completed: "Completada",
  cancelled: "Cancelada",
};
const statusColor: Record<string, "default" | "secondary" | "outline" | "destructive"> = {
  planned: "secondary",
  in_progress: "outline",
  completed: "default",
  cancelled: "destructive",
};

export default function InspeccionesPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const { data: inspections = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "audits", "vehicular"] as const,
    queryFn: () => api.get<{ audits: VehicleInspection[] }>("/v1/erp/quality/audits?audit_type=vehicular&page_size=50"),
    select: (d) => d.audits,
  });

  const createMutation = useMutation({
    mutationFn: (data: { Number: string; Date: string; AuditType: string; Scope: string }) =>
      api.post("/v1/erp/quality/audits", data),
    onSuccess: () => {
      toast.success("Inspección creada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "audits", "vehicular"] });
      setOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando inspecciones" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight flex items-center gap-2">
              <ClipboardCheckIcon className="size-5 text-blue-500" />
              Inspecciones Vehiculares
            </h1>
            <p className="text-sm text-muted-foreground mt-0.5">Inspecciones de seguridad de unidades y chasis</p>
          </div>
          <Button size="sm" onClick={() => setOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nueva inspección
          </Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-24">Número</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead>Unidad / Alcance</TableHead>
                <TableHead className="w-32">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {inspections.map((ins) => (
                <TableRow key={ins.id}>
                  <TableCell className="font-mono text-sm">{ins.number}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(ins.date)}</TableCell>
                  <TableCell className="text-sm">{ins.scope || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={statusColor[ins.status] || "secondary"}>
                      {statusLabel[ins.status] || ins.status}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
              {inspections.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                    Sin inspecciones registradas.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={open} onOpenChange={(v) => !v && setOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Inspección Vehicular</DialogTitle></DialogHeader>
          <CreateInspectionForm
            onSubmit={(data) => createMutation.mutate({ ...data, AuditType: "vehicular" })}
            isPending={createMutation.isPending}
            onClose={() => setOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateInspectionForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { Number: string; Date: string; Scope: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [scope, setScope] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (number && date) {
          onSubmit({ Number: number, Date: date, Scope: scope });
        }
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="INS-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Alcance (unidad / vehículo)</Label>
        <Input
          value={scope}
          onChange={(e) => setScope(e.target.value)}
          placeholder="Ej: Unidad 12 — interno/chasis"
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!number || !date || isPending}>
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}
