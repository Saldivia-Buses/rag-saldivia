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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { HeartPulseIcon, PlusIcon } from "lucide-react";

interface MedicalLeave {
  id: string;
  entity_name: string;
  leave_type: string;
  date_from: string;
  date_to: string;
  working_days: number;
  body_part_description: string;
  status: string;
  approved_by: string;
  observations: string;
}

interface MedicalConsultation {
  id: string;
  patient_name: string;
  consult_date: string;
  consult_time: string;
  symptoms: string;
  prescription: string;
  medic_user: string;
}

interface Entity {
  id: string;
  name: string;
}

interface BodyPart {
  id: string;
  description: string;
}

const leaveTypeLabel: Record<string, string> = {
  illness: "Enfermedad",
  accident: "Accidente laboral",
  vacation: "Vacaciones",
  leave: "Licencia",
  other: "Otro",
};

const leaveTypeVariant: Record<string, "default" | "secondary" | "outline" | "destructive"> = {
  illness: "destructive",
  accident: "destructive",
  vacation: "default",
  leave: "secondary",
  other: "outline",
};

const leaveStatusLabel: Record<string, string> = {
  pending: "Pendiente",
  approved: "Aprobada",
  rejected: "Rechazada",
};

const leaveStatusVariant: Record<string, "default" | "secondary" | "outline" | "destructive"> = {
  pending: "outline",
  approved: "default",
  rejected: "destructive",
};

export default function MedicinaPage() {
  const queryClient = useQueryClient();
  const [leaveDialogOpen, setLeaveDialogOpen] = useState(false);
  const [logDialogOpen, setLogDialogOpen] = useState(false);

  // Medical Leaves
  const { data: leaves = [], isLoading: leavesLoading, error: leavesError } = useQuery({
    queryKey: [...erpKeys.all, "safety", "medical-leaves"] as const,
    queryFn: () =>
      api.get<{ medical_leaves: MedicalLeave[] }>("/v1/erp/safety/medical-leaves?page_size=50"),
    select: (d) => d.medical_leaves,
  });

  // Medical Log
  const { data: consultations = [], isLoading: logLoading, error: logError } = useQuery({
    queryKey: [...erpKeys.all, "safety", "medical-log"] as const,
    queryFn: () =>
      api.get<{ consultations: MedicalConsultation[] }>("/v1/erp/safety/medical-log?page_size=50"),
    select: (d) => d.consultations,
  });

  // Catalogs
  const { data: entities = [] } = useQuery({
    queryKey: erpKeys.entities("employee"),
    queryFn: () =>
      api.get<{ entities: Entity[] }>("/v1/erp/entities?type=employee&page_size=200"),
    select: (d) => d.entities,
  });

  const { data: bodyParts = [] } = useQuery({
    queryKey: [...erpKeys.all, "safety", "body-parts"] as const,
    queryFn: () =>
      api.get<{ body_parts: BodyPart[] }>("/v1/erp/safety/body-parts"),
    select: (d) => d.body_parts,
  });

  // Mutations
  const createLeaveMutation = useMutation({
    mutationFn: (data: {
      entity_id: string;
      leave_type: string;
      date_from: string;
      date_to: string;
      working_days?: number;
      observations?: string;
      body_part_id?: string;
    }) => api.post("/v1/erp/safety/medical-leaves", data),
    onSuccess: () => {
      toast.success("Licencia registrada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "safety", "medical-leaves"] });
      setLeaveDialogOpen(false);
    },
    onError: permissionErrorToast,
  });

  const approveLeaveMutation = useMutation({
    mutationFn: ({ id, approved_by }: { id: string; approved_by: string }) =>
      api.patch(`/v1/erp/safety/medical-leaves/${id}/approve`, { approved_by }),
    onSuccess: () => {
      toast.success("Licencia aprobada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "safety", "medical-leaves"] });
    },
    onError: permissionErrorToast,
  });

  const createLogMutation = useMutation({
    mutationFn: (data: {
      consult_date: string;
      patient_name?: string;
      symptoms: string;
      prescription: string;
      medic_user?: string;
    }) => api.post("/v1/erp/safety/medical-log", data),
    onSuccess: () => {
      toast.success("Consulta registrada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "safety", "medical-log"] });
      setLogDialogOpen(false);
    },
    onError: permissionErrorToast,
  });

  const anyError = leavesError || logError;
  const anyLoading = leavesLoading || logLoading;

  if (anyError) return <ErrorState message="Error cargando medicina laboral" onRetry={() => window.location.reload()} />;
  if (anyLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight flex items-center gap-2">
              <HeartPulseIcon className="size-5 text-rose-500" />
              Medicina Laboral
            </h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Licencias médicas y libro de consultas diarias
            </p>
          </div>
        </div>

        <Tabs defaultValue="licencias">
          <TabsList className="mb-4">
            <TabsTrigger value="licencias">Licencias</TabsTrigger>
            <TabsTrigger value="libro">Libro médico</TabsTrigger>
          </TabsList>

          {/* TAB 1: Licencias */}
          <TabsContent value="licencias">
            <div className="flex justify-end mb-3">
              <Button size="sm" onClick={() => setLeaveDialogOpen(true)}>
                <PlusIcon className="size-4 mr-1.5" />Nueva licencia
              </Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Empleado</TableHead>
                    <TableHead className="w-36">Tipo</TableHead>
                    <TableHead className="w-28">Desde</TableHead>
                    <TableHead className="w-28">Hasta</TableHead>
                    <TableHead className="w-20">Días</TableHead>
                    <TableHead>Parte corporal</TableHead>
                    <TableHead className="w-28">Estado</TableHead>
                    <TableHead className="w-28">Acciones</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {leaves.map((lv) => (
                    <TableRow key={lv.id}>
                      <TableCell className="text-sm font-medium">{lv.entity_name || "—"}</TableCell>
                      <TableCell>
                        <Badge variant={leaveTypeVariant[lv.leave_type] ?? "secondary"}>
                          {leaveTypeLabel[lv.leave_type] ?? lv.leave_type}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(lv.date_from)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(lv.date_to)}</TableCell>
                      <TableCell className="text-sm text-center">{lv.working_days ?? "—"}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{lv.body_part_description || "—"}</TableCell>
                      <TableCell>
                        <Badge variant={leaveStatusVariant[lv.status] ?? "outline"}>
                          {leaveStatusLabel[lv.status] ?? lv.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {lv.status === "pending" && (
                          <Button
                            size="sm"
                            variant="outline"
                            className="h-7 text-xs px-2"
                            onClick={() =>
                              approveLeaveMutation.mutate({ id: lv.id, approved_by: "admin" })
                            }
                            disabled={approveLeaveMutation.isPending}
                          >
                            Aprobar
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                  {leaves.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={8} className="h-24 text-center text-muted-foreground">
                        Sin licencias registradas.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          {/* TAB 2: Libro médico */}
          <TabsContent value="libro">
            <div className="flex justify-end mb-3">
              <Button size="sm" onClick={() => setLogDialogOpen(true)}>
                <PlusIcon className="size-4 mr-1.5" />Nueva consulta
              </Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-28">Fecha</TableHead>
                    <TableHead>Paciente</TableHead>
                    <TableHead>Síntomas</TableHead>
                    <TableHead>Prescripción</TableHead>
                    <TableHead className="w-36">Médico/Enfermero</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {consultations.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="text-sm text-muted-foreground">
                        {fmtDateShort(c.consult_date)}
                        {c.consult_time ? ` ${c.consult_time.slice(0, 5)}` : ""}
                      </TableCell>
                      <TableCell className="text-sm font-medium">{c.patient_name || "—"}</TableCell>
                      <TableCell className="text-sm text-muted-foreground truncate max-w-48">{c.symptoms || "—"}</TableCell>
                      <TableCell className="text-sm text-muted-foreground truncate max-w-48">{c.prescription || "—"}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{c.medic_user || "—"}</TableCell>
                    </TableRow>
                  ))}
                  {consultations.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                        Sin consultas registradas.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* Dialog: Nueva licencia */}
      <Dialog open={leaveDialogOpen} onOpenChange={(v) => !v && setLeaveDialogOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Licencia Médica</DialogTitle></DialogHeader>
          <CreateLeaveForm
            entities={entities}
            bodyParts={bodyParts}
            onSubmit={(data) => createLeaveMutation.mutate(data)}
            isPending={createLeaveMutation.isPending}
            onClose={() => setLeaveDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Dialog: Nueva consulta */}
      <Dialog open={logDialogOpen} onOpenChange={(v) => !v && setLogDialogOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Consulta Médica</DialogTitle></DialogHeader>
          <CreateConsultationForm
            onSubmit={(data) => createLogMutation.mutate(data)}
            isPending={createLogMutation.isPending}
            onClose={() => setLogDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateLeaveForm({
  entities,
  bodyParts,
  onSubmit,
  isPending,
  onClose,
}: {
  entities: Entity[];
  bodyParts: BodyPart[];
  onSubmit: (data: {
    entity_id: string;
    leave_type: string;
    date_from: string;
    date_to: string;
    working_days?: number;
    observations?: string;
    body_part_id?: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [entityId, setEntityId] = useState("");
  const [leaveType, setLeaveType] = useState("illness");
  const [dateFrom, setDateFrom] = useState(today);
  const [dateTo, setDateTo] = useState("");
  const [workingDays, setWorkingDays] = useState("");
  const [observations, setObservations] = useState("");
  const [bodyPartId, setBodyPartId] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!entityId || !dateFrom || !dateTo) return;
        onSubmit({
          entity_id: entityId,
          leave_type: leaveType,
          date_from: dateFrom,
          date_to: dateTo,
          working_days: workingDays ? parseInt(workingDays, 10) : undefined,
          observations: observations || undefined,
          body_part_id: bodyPartId || undefined,
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
          required
        >
          <option value="">Seleccionar empleado...</option>
          {entities.map((ent) => (
            <option key={ent.id} value={ent.id}>{ent.name}</option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label>Tipo de licencia</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={leaveType}
          onChange={(e) => setLeaveType(e.target.value)}
        >
          <option value="illness">Enfermedad</option>
          <option value="accident">Accidente laboral</option>
          <option value="vacation">Vacaciones</option>
          <option value="leave">Licencia</option>
          <option value="other">Otro</option>
        </select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Desde</Label>
          <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <Label>Hasta</Label>
          <Input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)} required />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Días hábiles</Label>
          <Input
            type="number"
            min="0"
            value={workingDays}
            onChange={(e) => setWorkingDays(e.target.value)}
            placeholder="0"
          />
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
      <div className="space-y-2">
        <Label>Notas</Label>
        <Textarea
          value={observations}
          onChange={(e) => setObservations(e.target.value)}
          placeholder="Observaciones adicionales..."
          rows={2}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!entityId || !dateFrom || !dateTo || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}

function CreateConsultationForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: {
    consult_date: string;
    patient_name?: string;
    symptoms: string;
    prescription: string;
    medic_user?: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [consultDate, setConsultDate] = useState(today);
  const [patientName, setPatientName] = useState("");
  const [symptoms, setSymptoms] = useState("");
  const [prescription, setPrescription] = useState("");
  const [medicUser, setMedicUser] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!consultDate || !symptoms) return;
        onSubmit({
          consult_date: consultDate,
          patient_name: patientName || undefined,
          symptoms,
          prescription,
          medic_user: medicUser || undefined,
        });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={consultDate} onChange={(e) => setConsultDate(e.target.value)} required />
        </div>
        <div className="space-y-2">
          <Label>Médico/Enfermero</Label>
          <Input
            value={medicUser}
            onChange={(e) => setMedicUser(e.target.value)}
            placeholder="Nombre"
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Nombre del paciente</Label>
        <Input
          value={patientName}
          onChange={(e) => setPatientName(e.target.value)}
          placeholder="Nombre completo"
        />
      </div>
      <div className="space-y-2">
        <Label>Síntomas</Label>
        <Textarea
          value={symptoms}
          onChange={(e) => setSymptoms(e.target.value)}
          placeholder="Descripción de síntomas..."
          rows={2}
          required
        />
      </div>
      <div className="space-y-2">
        <Label>Prescripción</Label>
        <Textarea
          value={prescription}
          onChange={(e) => setPrescription(e.target.value)}
          placeholder="Tratamiento indicado..."
          rows={2}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!consultDate || !symptoms || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}
