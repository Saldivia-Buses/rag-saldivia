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

interface Accident {
  id: string;
  entity_name: string;
  accident_type_name: string;
  body_part_description: string;
  incident_date: string;
  recovery_date: string;
  lost_days: number;
  status: string;
  observations: string;
  reported_by: string;
}

interface AccidentType {
  id: string;
  name: string;
  severity_idx: number;
}

interface BodyPart {
  id: string;
  description: string;
}

interface Entity {
  id: string;
  name: string;
}

const statusLabel: Record<string, string> = {
  open: "Abierto",
  investigating: "Investigando",
  closed: "Cerrado",
};
const statusColor: Record<string, "default" | "secondary" | "outline"> = {
  open: "secondary",
  investigating: "outline",
  closed: "default",
};

export default function IncidentesPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const { data: accidents = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "safety", "accidents"] as const,
    queryFn: () =>
      api.get<{ accidents: Accident[] }>("/v1/erp/safety/accidents?page_size=50"),
    select: (d) => d.accidents,
  });

  const { data: accidentTypes = [] } = useQuery({
    queryKey: [...erpKeys.all, "safety", "accident-types"] as const,
    queryFn: () =>
      api.get<{ accident_types: AccidentType[] }>("/v1/erp/safety/accident-types"),
    select: (d) => d.accident_types,
  });

  const { data: bodyParts = [] } = useQuery({
    queryKey: [...erpKeys.all, "safety", "body-parts"] as const,
    queryFn: () =>
      api.get<{ body_parts: BodyPart[] }>("/v1/erp/safety/body-parts"),
    select: (d) => d.body_parts,
  });

  const { data: entities = [] } = useQuery({
    queryKey: erpKeys.entities("employee"),
    queryFn: () =>
      api.get<{ entities: Entity[] }>("/v1/erp/entities?type=employee&page_size=200"),
    select: (d) => d.entities,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      entity_id?: string;
      accident_type_id?: string;
      body_part_id?: string;
      incident_date: string;
      lost_days?: number;
      observations?: string;
      reported_by?: string;
    }) => api.post("/v1/erp/safety/accidents", data),
    onSuccess: () => {
      toast.success("Accidente registrado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "safety", "accidents"] });
      setOpen(false);
    },
    onError: permissionErrorToast,
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      api.patch(`/v1/erp/safety/accidents/${id}/status`, { status }),
    onSuccess: () => {
      toast.success("Estado actualizado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "safety", "accidents"] });
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
            <p className="text-sm text-muted-foreground mt-0.5">
              Registro de accidentes e incidentes de seguridad laboral
            </p>
          </div>
          <Button size="sm" onClick={() => setOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Registrar accidente
          </Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Empleado</TableHead>
                <TableHead className="w-40">Tipo de accidente</TableHead>
                <TableHead className="w-36">Parte afectada</TableHead>
                <TableHead className="w-28">Fecha</TableHead>
                <TableHead className="w-24">Días perdidos</TableHead>
                <TableHead className="w-32">Estado</TableHead>
                <TableHead>Reportado por</TableHead>
                <TableHead className="w-28">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {accidents.map((acc) => (
                <TableRow key={acc.id}>
                  <TableCell className="text-sm font-medium">{acc.entity_name || "—"}</TableCell>
                  <TableCell className="text-sm">{acc.accident_type_name || "—"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{acc.body_part_description || "—"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(acc.incident_date)}</TableCell>
                  <TableCell className="text-sm text-center">{acc.lost_days ?? "—"}</TableCell>
                  <TableCell>
                    <Badge variant={statusColor[acc.status] ?? "secondary"}>
                      {statusLabel[acc.status] ?? acc.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">{acc.reported_by || "—"}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      {acc.status === "open" && (
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-7 text-xs px-2"
                          onClick={() => statusMutation.mutate({ id: acc.id, status: "investigating" })}
                          disabled={statusMutation.isPending}
                        >
                          Investigar
                        </Button>
                      )}
                      {acc.status === "investigating" && (
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-7 text-xs px-2"
                          onClick={() => statusMutation.mutate({ id: acc.id, status: "closed" })}
                          disabled={statusMutation.isPending}
                        >
                          Cerrar
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {accidents.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8} className="h-24 text-center text-muted-foreground">
                    Sin accidentes registrados.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={open} onOpenChange={(v) => !v && setOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Registrar Accidente</DialogTitle></DialogHeader>
          <CreateAccidentForm
            entities={entities}
            accidentTypes={accidentTypes}
            bodyParts={bodyParts}
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateAccidentForm({
  entities,
  accidentTypes,
  bodyParts,
  onSubmit,
  isPending,
  onClose,
}: {
  entities: Entity[];
  accidentTypes: AccidentType[];
  bodyParts: BodyPart[];
  onSubmit: (data: {
    entity_id?: string;
    accident_type_id?: string;
    body_part_id?: string;
    incident_date: string;
    lost_days?: number;
    observations?: string;
    reported_by?: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [entityId, setEntityId] = useState("");
  const [accidentTypeId, setAccidentTypeId] = useState("");
  const [bodyPartId, setBodyPartId] = useState("");
  const [incidentDate, setIncidentDate] = useState(today);
  const [lostDays, setLostDays] = useState("");
  const [observations, setObservations] = useState("");
  const [reportedBy, setReportedBy] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!incidentDate) return;
        onSubmit({
          entity_id: entityId || undefined,
          accident_type_id: accidentTypeId || undefined,
          body_part_id: bodyPartId || undefined,
          incident_date: incidentDate,
          lost_days: lostDays ? parseInt(lostDays, 10) : undefined,
          observations: observations || undefined,
          reported_by: reportedBy || undefined,
        });
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
          <option value="">Sin especificar</option>
          {entities.map((ent) => (
            <option key={ent.id} value={ent.id}>{ent.name}</option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Tipo de accidente</Label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm bg-card"
            value={accidentTypeId}
            onChange={(e) => setAccidentTypeId(e.target.value)}
          >
            <option value="">Sin especificar</option>
            {accidentTypes.map((at) => (
              <option key={at.id} value={at.id}>{at.name}</option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label>Parte corporal</Label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm bg-card"
            value={bodyPartId}
            onChange={(e) => setBodyPartId(e.target.value)}
          >
            <option value="">Sin especificar</option>
            {bodyParts.map((bp) => (
              <option key={bp.id} value={bp.id}>{bp.description}</option>
            ))}
          </select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Fecha del accidente</Label>
          <Input type="date" value={incidentDate} onChange={(e) => setIncidentDate(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <Label>Días perdidos</Label>
          <Input
            type="number"
            min="0"
            value={lostDays}
            onChange={(e) => setLostDays(e.target.value)}
            placeholder="0"
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Reportado por</Label>
        <Input
          value={reportedBy}
          onChange={(e) => setReportedBy(e.target.value)}
          placeholder="Nombre del responsable"
        />
      </div>
      <div className="space-y-2">
        <Label>Notas</Label>
        <Textarea
          value={observations}
          onChange={(e) => setObservations(e.target.value)}
          placeholder="Descripción del accidente..."
          rows={3}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!incidentDate || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}
