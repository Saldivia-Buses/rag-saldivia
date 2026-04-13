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
import { Textarea } from "@/components/ui/textarea";
import { AlertTriangleIcon, PlusIcon } from "lucide-react";

interface Incident {
  id: string;
  number: string;
  date: string;
  description: string;
  severity: string;
  status: string;
}

const sevLabel: Record<string, string> = { minor: "Menor", major: "Mayor", critical: "Crítico" };
const sevColor: Record<string, "default" | "secondary" | "destructive"> = { minor: "secondary", major: "default", critical: "destructive" };
const statusLabel: Record<string, string> = { open: "Abierto", investigating: "Investigando", closed: "Cerrado" };
const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", investigating: "outline", closed: "default" };

export default function IncidentesPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const { data: incidents = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "nc", "seguridad"] as const,
    queryFn: () => api.get<{ nonconformities: Incident[] }>("/v1/erp/quality/nc?source=seguridad&page_size=50"),
    select: (d) => d.nonconformities,
  });

  const createMutation = useMutation({
    mutationFn: (data: { Number: string; Date: string; Description: string; Severity: string; Source: string }) =>
      api.post("/v1/erp/quality/nc", data),
    onSuccess: () => {
      toast.success("Incidente registrado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "nc", "seguridad"] });
      setOpen(false);
    },
    onError: permissionErrorToast,
  });

  const closeMutation = useMutation({
    mutationFn: (id: string) => api.patch(`/v1/erp/quality/nc/${id}/status`, { Status: "closed" }),
    onSuccess: () => {
      toast.success("Incidente cerrado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "nc", "seguridad"] });
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando incidentes" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight flex items-center gap-2">
              <AlertTriangleIcon className="size-5 text-destructive" />
              Incidentes
            </h1>
            <p className="text-sm text-muted-foreground mt-0.5">Registro de accidentes e incidentes de seguridad vial</p>
          </div>
          <Button size="sm" onClick={() => setOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Registrar incidente
          </Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-24">Número</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead className="w-28">Gravedad</TableHead>
                <TableHead className="w-28">Estado</TableHead>
                <TableHead className="w-24">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {incidents.map((inc) => (
                <TableRow key={inc.id}>
                  <TableCell className="font-mono text-sm">{inc.number}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(inc.date)}</TableCell>
                  <TableCell className="text-sm truncate max-w-72">{inc.description}</TableCell>
                  <TableCell>
                    <Badge variant={sevColor[inc.severity] || "secondary"}>
                      {sevLabel[inc.severity] || inc.severity}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusColor[inc.status] || "secondary"}>
                      {statusLabel[inc.status] || inc.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {inc.status !== "closed" && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => closeMutation.mutate(inc.id)}
                        disabled={closeMutation.isPending}
                      >
                        Cerrar
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {incidents.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    Sin incidentes registrados.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={open} onOpenChange={(v) => !v && setOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Registrar Incidente</DialogTitle></DialogHeader>
          <CreateIncidentForm
            onSubmit={(data) => createMutation.mutate({ ...data, Source: "seguridad" })}
            isPending={createMutation.isPending}
            onClose={() => setOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateIncidentForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { Number: string; Date: string; Description: string; Severity: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [severity, setSeverity] = useState("minor");
  const [description, setDescription] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (number && date && description) {
          onSubmit({ Number: number, Date: date, Description: description, Severity: severity });
        }
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="INC-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Gravedad</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={severity}
          onChange={(e) => setSeverity(e.target.value)}
        >
          <option value="minor">Menor</option>
          <option value="major">Mayor</option>
          <option value="critical">Crítico</option>
        </select>
      </div>
      <div className="space-y-2">
        <Label>Descripción</Label>
        <Textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describa el incidente..."
          rows={3}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!number || !date || !description || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}
