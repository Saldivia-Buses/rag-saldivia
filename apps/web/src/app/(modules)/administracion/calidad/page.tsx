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
import { AlertTriangleIcon, CheckIcon, ClipboardCheckIcon, FileTextIcon, ListChecksIcon, PlusIcon, XIcon } from "lucide-react";

interface NC { id: string; number: string; date: string; description: string; severity: string; status: string; assigned_name: string; }
interface QualityAudit { id: string; number: string; date: string; audit_type: string; scope: string; status: string; }
interface ControlledDoc { id: string; code: string; title: string; revision: number; status: string; }
interface ActionPlan { id: string; nonconformity_id: string | null; nc_number: string | null; responsible_name: string | null; description: string; target_date: string | null; cost_savings: number; status: string; time_savings_hours: number; }
interface ActionTask { id: string; description: string; leader_name: string | null; target_date: string | null; completed: boolean; }
interface NCOption { id: string; number: string; }
interface Employee { id: string; name: string; }

const sevColor: Record<string, "default" | "secondary" | "destructive"> = { minor: "secondary", major: "default", critical: "destructive" };
const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", investigating: "outline", corrective_action: "outline", closed: "default", planned: "secondary", in_progress: "outline", completed: "default", draft: "secondary", approved: "default", obsolete: "secondary" };

const planStatusLabel: Record<string, string> = { draft: "Borrador", active: "Activo", closed: "Cerrado", cancelled: "Cancelado" };
const planStatusVariant: Record<string, "default" | "secondary" | "outline"> = { draft: "secondary", active: "outline", closed: "default", cancelled: "secondary" };

function fmtCurrency(n: number) {
  if (!n) return "—";
  return new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
}

export default function CalidadPage() {
  const queryClient = useQueryClient();
  const [ncOpen, setNcOpen] = useState(false);
  const [auditOpen, setAuditOpen] = useState(false);
  const [planOpen, setPlanOpen] = useState(false);
  const [taskOpen, setTaskOpen] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null);

  const { data: ncs = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "nc"] as const,
    queryFn: () => api.get<{ nonconformities: NC[] }>("/v1/erp/quality/nc?page_size=50"),
    select: (d) => d.nonconformities,
  });

  const { data: audits = [] } = useQuery({
    queryKey: [...erpKeys.all, "quality", "audits"] as const,
    queryFn: () => api.get<{ audits: QualityAudit[] }>("/v1/erp/quality/audits?page_size=50"),
    select: (d) => d.audits,
  });

  const { data: docs = [] } = useQuery({
    queryKey: [...erpKeys.all, "quality", "documents"] as const,
    queryFn: () => api.get<{ documents: ControlledDoc[] }>("/v1/erp/quality/documents?page_size=50"),
    select: (d) => d.documents,
  });

  const { data: actionPlans = [] } = useQuery({
    queryKey: [...erpKeys.all, "quality", "action-plans"] as const,
    queryFn: () => api.get<{ action_plans: ActionPlan[] }>("/v1/erp/quality/action-plans"),
    select: (d) => d.action_plans,
  });

  const { data: planTasks = [] } = useQuery({
    queryKey: [...erpKeys.all, "quality", "action-plans", selectedPlanId, "tasks"] as const,
    queryFn: () => api.get<{ tasks: ActionTask[] }>(`/v1/erp/quality/action-plans/${selectedPlanId}/tasks`),
    select: (d) => d.tasks,
    enabled: !!selectedPlanId,
  });

  const { data: ncOptions = [] } = useQuery({
    queryKey: [...erpKeys.all, "quality", "nc-options"] as const,
    queryFn: () => api.get<{ nonconformities: NC[] }>("/v1/erp/quality/nc?page_size=100"),
    select: (d): NCOption[] => d.nonconformities.map((n) => ({ id: n.id, number: n.number })),
    enabled: planOpen,
  });

  const { data: employees = [] } = useQuery({
    queryKey: [...erpKeys.all, "entities", "employees"] as const,
    queryFn: () => api.get<{ entities: Employee[] }>("/v1/erp/entities?type=employee&page_size=200"),
    select: (d) => d.entities,
    enabled: planOpen || taskOpen,
  });

  const createNCMutation = useMutation({
    mutationFn: (data: { Number: string; Date: string; Description: string; Severity: string }) =>
      api.post("/v1/erp/quality/nc", data),
    onSuccess: () => {
      toast.success("No conformidad registrada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "nc"] });
      setNcOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createAuditMutation = useMutation({
    mutationFn: (data: { Number: string; Date: string; AuditType: string; Scope: string }) =>
      api.post("/v1/erp/quality/audits", data),
    onSuccess: () => {
      toast.success("Auditoría creada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "audits"] });
      setAuditOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createPlanMutation = useMutation({
    mutationFn: (data: { nonconformity_id?: string; responsible_id?: string; description: string; target_date?: string; cost_savings?: number; time_savings_hours?: number }) =>
      api.post("/v1/erp/quality/action-plans", data),
    onSuccess: () => {
      toast.success("Plan de acción creado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "action-plans"] });
      setPlanOpen(false);
    },
    onError: permissionErrorToast,
  });

  const updatePlanStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      api.patch(`/v1/erp/quality/action-plans/${id}/status`, { status }),
    onSuccess: () => {
      toast.success("Estado actualizado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "action-plans"] });
    },
    onError: permissionErrorToast,
  });

  const createTaskMutation = useMutation({
    mutationFn: (data: { description: string; leader_id?: string; target_date?: string }) =>
      api.post(`/v1/erp/quality/action-plans/${selectedPlanId}/tasks`, data),
    onSuccess: () => {
      toast.success("Tarea agregada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "action-plans", selectedPlanId, "tasks"] });
      setTaskOpen(false);
    },
    onError: permissionErrorToast,
  });

  const completeTaskMutation = useMutation({
    mutationFn: ({ planId, taskId }: { planId: string; taskId: string }) =>
      api.patch(`/v1/erp/quality/action-plans/${planId}/tasks/${taskId}/complete`, {}),
    onSuccess: () => {
      toast.success("Tarea completada");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "action-plans", selectedPlanId, "tasks"] });
    },
    onError: permissionErrorToast,
  });

  const selectedPlan = selectedPlanId ? actionPlans.find((p) => p.id === selectedPlanId) : null;

  if (error) return <ErrorState message="Error cargando calidad" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Calidad</h1>
            <p className="text-sm text-muted-foreground mt-0.5">No conformidades, auditorías y documentos controlados</p>
          </div>
          <Button size="sm" onClick={() => setNcOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva NC</Button>
        </div>
        <Tabs defaultValue="nc">
          <TabsList className="mb-4">
            <TabsTrigger value="nc"><AlertTriangleIcon className="size-3.5 mr-1.5" />No Conformidades ({ncs.length})</TabsTrigger>
            <TabsTrigger value="audits"><ClipboardCheckIcon className="size-3.5 mr-1.5" />Auditorías</TabsTrigger>
            <TabsTrigger value="docs"><FileTextIcon className="size-3.5 mr-1.5" />Documentos</TabsTrigger>
            <TabsTrigger value="action-plans"><ListChecksIcon className="size-3.5 mr-1.5" />Planes de acción</TabsTrigger>
          </TabsList>
          <TabsContent value="nc">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-20">NC</TableHead><TableHead className="w-28">Fecha</TableHead><TableHead>Descripción</TableHead><TableHead className="w-20">Sev.</TableHead><TableHead className="w-28">Asignado</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{ncs.map((nc) => (<TableRow key={nc.id}><TableCell className="font-mono text-sm">{nc.number}</TableCell><TableCell className="text-sm text-muted-foreground">{fmtDateShort(nc.date)}</TableCell><TableCell className="text-sm truncate max-w-64">{nc.description}</TableCell><TableCell><Badge variant={sevColor[nc.severity] || "secondary"}>{nc.severity}</Badge></TableCell><TableCell className="text-sm">{nc.assigned_name || "\u2014"}</TableCell><TableCell><Badge variant={statusColor[nc.status] || "secondary"}>{nc.status}</Badge></TableCell></TableRow>))}
                {ncs.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin no conformidades.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="audits">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setAuditOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva auditoría</Button>
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-20">Nro</TableHead><TableHead className="w-28">Fecha</TableHead><TableHead className="w-24">Tipo</TableHead><TableHead>Alcance</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{audits.map((a) => (<TableRow key={a.id}><TableCell className="font-mono text-sm">{a.number}</TableCell><TableCell className="text-sm text-muted-foreground">{fmtDateShort(a.date)}</TableCell><TableCell><Badge variant="secondary">{a.audit_type}</Badge></TableCell><TableCell className="text-sm">{a.scope || "\u2014"}</TableCell><TableCell><Badge variant={statusColor[a.status] || "secondary"}>{a.status}</Badge></TableCell></TableRow>))}
                {audits.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin auditorías.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="docs">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-24">Código</TableHead><TableHead>Título</TableHead><TableHead className="w-16 text-center">Rev.</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{docs.map((d) => (<TableRow key={d.id}><TableCell className="font-mono text-sm">{d.code}</TableCell><TableCell className="text-sm">{d.title}</TableCell><TableCell className="text-center text-sm">{d.revision}</TableCell><TableCell><Badge variant={statusColor[d.status] || "secondary"}>{d.status}</Badge></TableCell></TableRow>))}
                {docs.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin documentos.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>

          <TabsContent value="action-plans">
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setPlanOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nuevo plan</Button>
            </div>
            <div className="flex gap-4">
              {/* Plans list */}
              <div className="flex-1 rounded-xl border border-border/40 bg-card overflow-hidden">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-20">NC</TableHead>
                      <TableHead>Descripción</TableHead>
                      <TableHead className="w-32">Responsable</TableHead>
                      <TableHead className="w-24">Objetivo</TableHead>
                      <TableHead className="w-28">Ahorro $</TableHead>
                      <TableHead className="w-28">Estado</TableHead>
                      <TableHead className="w-20"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {actionPlans.map((plan) => (
                      <TableRow
                        key={plan.id}
                        className={`cursor-pointer ${selectedPlanId === plan.id ? "bg-accent/50" : ""}`}
                        onClick={() => setSelectedPlanId(selectedPlanId === plan.id ? null : plan.id)}
                      >
                        <TableCell className="font-mono text-sm">{plan.nc_number || "—"}</TableCell>
                        <TableCell className="text-sm truncate max-w-48">{plan.description}</TableCell>
                        <TableCell className="text-sm">{plan.responsible_name || "—"}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{plan.target_date ? fmtDateShort(plan.target_date) : "—"}</TableCell>
                        <TableCell className="text-sm">{fmtCurrency(plan.cost_savings)}</TableCell>
                        <TableCell>
                          <Badge variant={planStatusVariant[plan.status] || "secondary"}>
                            {planStatusLabel[plan.status] || plan.status}
                          </Badge>
                        </TableCell>
                        <TableCell onClick={(e) => e.stopPropagation()}>
                          {plan.status === "draft" && (
                            <Button
                              size="sm"
                              variant="outline"
                              className="h-7 text-xs"
                              onClick={() => updatePlanStatusMutation.mutate({ id: plan.id, status: "active" })}
                              disabled={updatePlanStatusMutation.isPending}
                            >
                              Activar
                            </Button>
                          )}
                          {plan.status === "active" && (
                            <Button
                              size="sm"
                              variant="outline"
                              className="h-7 text-xs"
                              onClick={() => updatePlanStatusMutation.mutate({ id: plan.id, status: "closed" })}
                              disabled={updatePlanStatusMutation.isPending}
                            >
                              Cerrar
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                    {actionPlans.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={7} className="h-24 text-center text-muted-foreground">Sin planes de acción.</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>

              {/* Task panel */}
              {selectedPlan && (
                <div className="w-80 shrink-0 rounded-xl border border-border/40 bg-card flex flex-col">
                  <div className="flex items-center justify-between px-4 py-3 border-b border-border/40">
                    <div>
                      <p className="text-sm font-medium">Tareas</p>
                      <p className="text-xs text-muted-foreground truncate max-w-52">{selectedPlan.description}</p>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={() => setTaskOpen(true)} title="Nueva tarea">
                        <PlusIcon className="size-3.5" />
                      </Button>
                      <Button size="sm" variant="ghost" className="h-7 w-7 p-0" onClick={() => setSelectedPlanId(null)} title="Cerrar panel">
                        <XIcon className="size-3.5" />
                      </Button>
                    </div>
                  </div>
                  <div className="flex-1 overflow-y-auto divide-y divide-border/30">
                    {planTasks.map((task) => (
                      <div key={task.id} className="px-4 py-3 flex items-start gap-3">
                        <div className="flex-1 min-w-0">
                          <p className={`text-sm ${task.completed ? "line-through text-muted-foreground" : ""}`}>{task.description}</p>
                          <div className="flex items-center gap-2 mt-1">
                            {task.leader_name && <span className="text-xs text-muted-foreground">{task.leader_name}</span>}
                            {task.target_date && <span className="text-xs text-muted-foreground">{fmtDateShort(task.target_date)}</span>}
                            {task.completed && <Badge variant="default" className="text-xs h-4 px-1">Completada</Badge>}
                          </div>
                        </div>
                        {!task.completed && (
                          <Button
                            size="sm"
                            variant="ghost"
                            className="h-6 w-6 p-0 shrink-0 text-muted-foreground hover:text-foreground"
                            onClick={() => completeTaskMutation.mutate({ planId: selectedPlan.id, taskId: task.id })}
                            disabled={completeTaskMutation.isPending}
                            title="Completar"
                          >
                            <CheckIcon className="size-3.5" />
                          </Button>
                        )}
                      </div>
                    ))}
                    {planTasks.length === 0 && (
                      <div className="h-24 flex items-center justify-center text-sm text-muted-foreground">Sin tareas.</div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={ncOpen} onOpenChange={(v) => !v && setNcOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva No Conformidad</DialogTitle></DialogHeader>
          <CreateNCForm
            onSubmit={(data) => createNCMutation.mutate(data)}
            isPending={createNCMutation.isPending}
            onClose={() => setNcOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={auditOpen} onOpenChange={(v) => !v && setAuditOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Auditoría</DialogTitle></DialogHeader>
          <CreateAuditForm
            onSubmit={(data) => createAuditMutation.mutate(data)}
            isPending={createAuditMutation.isPending}
            onClose={() => setAuditOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={planOpen} onOpenChange={(v) => !v && setPlanOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo Plan de Acción</DialogTitle></DialogHeader>
          <CreatePlanForm
            ncOptions={ncOptions}
            employees={employees}
            onSubmit={(data) => createPlanMutation.mutate(data)}
            isPending={createPlanMutation.isPending}
            onClose={() => setPlanOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={taskOpen} onOpenChange={(v) => !v && setTaskOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva Tarea</DialogTitle></DialogHeader>
          <CreateTaskForm
            employees={employees}
            onSubmit={(data) => createTaskMutation.mutate(data)}
            isPending={createTaskMutation.isPending}
            onClose={() => setTaskOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateNCForm({
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
        if (number && date && description) onSubmit({ Number: number, Date: date, Description: description, Severity: severity });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="NC-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Severidad</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={severity}
          onChange={(e) => setSeverity(e.target.value)}
        >
          <option value="minor">Menor</option>
          <option value="major">Mayor</option>
          <option value="critical">Crítica</option>
        </select>
      </div>
      <div className="space-y-2">
        <Label>Descripción</Label>
        <Textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describa la no conformidad..."
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

function CreateAuditForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { Number: string; Date: string; AuditType: string; Scope: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().slice(0, 10);
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [auditType, setAuditType] = useState("internal");
  const [scope, setScope] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (number && date) onSubmit({ Number: number, Date: date, AuditType: auditType, Scope: scope });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="AUD-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Tipo</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={auditType}
          onChange={(e) => setAuditType(e.target.value)}
        >
          <option value="internal">Interna</option>
          <option value="external">Externa</option>
          <option value="supplier">Proveedor</option>
        </select>
      </div>
      <div className="space-y-2">
        <Label>Alcance</Label>
        <Input value={scope} onChange={(e) => setScope(e.target.value)} placeholder="Área o proceso auditado" />
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

function CreatePlanForm({
  ncOptions,
  employees,
  onSubmit,
  isPending,
  onClose,
}: {
  ncOptions: NCOption[];
  employees: Employee[];
  onSubmit: (data: { nonconformity_id?: string; responsible_id?: string; description: string; target_date?: string; cost_savings?: number; time_savings_hours?: number }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [ncId, setNcId] = useState("");
  const [responsibleId, setResponsibleId] = useState("");
  const [description, setDescription] = useState("");
  const [targetDate, setTargetDate] = useState("");
  const [costSavings, setCostSavings] = useState("");
  const [timeSavings, setTimeSavings] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!description) return;
        onSubmit({
          ...(ncId ? { nonconformity_id: ncId } : {}),
          ...(responsibleId ? { responsible_id: responsibleId } : {}),
          description,
          ...(targetDate ? { target_date: targetDate } : {}),
          ...(costSavings ? { cost_savings: parseFloat(costSavings) } : {}),
          ...(timeSavings ? { time_savings_hours: parseFloat(timeSavings) } : {}),
        });
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>No Conformidad (opcional)</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={ncId}
          onChange={(e) => setNcId(e.target.value)}
        >
          <option value="">— Sin NC asociada —</option>
          {ncOptions.map((nc) => (
            <option key={nc.id} value={nc.id}>{nc.number}</option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label>Descripción <span className="text-destructive">*</span></Label>
        <Textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describa el plan de acción..."
          rows={3}
          required
        />
      </div>
      <div className="space-y-2">
        <Label>Responsable (opcional)</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={responsibleId}
          onChange={(e) => setResponsibleId(e.target.value)}
        >
          <option value="">— Sin asignar —</option>
          {employees.map((emp) => (
            <option key={emp.id} value={emp.id}>{emp.name}</option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Fecha objetivo</Label>
          <Input type="date" value={targetDate} onChange={(e) => setTargetDate(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Ahorro estimado $</Label>
          <Input type="number" min="0" step="0.01" value={costSavings} onChange={(e) => setCostSavings(e.target.value)} placeholder="0" />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Ahorro de tiempo (hs)</Label>
        <Input type="number" min="0" step="0.5" value={timeSavings} onChange={(e) => setTimeSavings(e.target.value)} placeholder="0" />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!description || isPending}>
          {isPending ? "Creando..." : "Crear plan"}
        </Button>
      </div>
    </form>
  );
}

function CreateTaskForm({
  employees,
  onSubmit,
  isPending,
  onClose,
}: {
  employees: Employee[];
  onSubmit: (data: { description: string; leader_id?: string; target_date?: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [description, setDescription] = useState("");
  const [leaderId, setLeaderId] = useState("");
  const [targetDate, setTargetDate] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!description) return;
        onSubmit({
          description,
          ...(leaderId ? { leader_id: leaderId } : {}),
          ...(targetDate ? { target_date: targetDate } : {}),
        });
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Descripción <span className="text-destructive">*</span></Label>
        <Textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Describa la tarea..."
          rows={3}
          required
        />
      </div>
      <div className="space-y-2">
        <Label>Responsable (opcional)</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={leaderId}
          onChange={(e) => setLeaderId(e.target.value)}
        >
          <option value="">— Sin asignar —</option>
          {employees.map((emp) => (
            <option key={emp.id} value={emp.id}>{emp.name}</option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label>Fecha objetivo</Label>
        <Input type="date" value={targetDate} onChange={(e) => setTargetDate(e.target.value)} />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!description || isPending}>
          {isPending ? "Agregando..." : "Agregar tarea"}
        </Button>
      </div>
    </form>
  );
}
