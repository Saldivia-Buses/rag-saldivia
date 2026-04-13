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
import { AlertTriangleIcon, ClipboardCheckIcon, FileTextIcon, PlusIcon } from "lucide-react";

interface NC { id: string; number: string; date: string; description: string; severity: string; status: string; assigned_name: string; }
interface QualityAudit { id: string; number: string; date: string; audit_type: string; scope: string; status: string; }
interface ControlledDoc { id: string; code: string; title: string; revision: number; status: string; }

const sevColor: Record<string, "default" | "secondary" | "destructive"> = { minor: "secondary", major: "default", critical: "destructive" };
const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", investigating: "outline", corrective_action: "outline", closed: "default", planned: "secondary", in_progress: "outline", completed: "default", draft: "secondary", approved: "default", obsolete: "secondary" };

export default function CalidadPage() {
  const queryClient = useQueryClient();
  const [ncOpen, setNcOpen] = useState(false);
  const [auditOpen, setAuditOpen] = useState(false);

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
