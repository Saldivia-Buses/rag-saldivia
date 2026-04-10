"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { AlertTriangleIcon, ClipboardCheckIcon, FileTextIcon } from "lucide-react";

interface NC { id: string; number: string; date: string; description: string; severity: string; status: string; assigned_name: string; }
interface QualityAudit { id: string; number: string; date: string; audit_type: string; scope: string; status: string; }
interface ControlledDoc { id: string; code: string; title: string; revision: number; status: string; }

const fmtDate = (s: string) => new Date(s).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });
const sevColor: Record<string, "default" | "secondary" | "destructive"> = { minor: "secondary", major: "default", critical: "destructive" };
const statusColor: Record<string, "default" | "secondary" | "outline"> = { open: "secondary", investigating: "outline", corrective_action: "outline", closed: "default", planned: "secondary", in_progress: "outline", completed: "default", draft: "secondary", approved: "default", obsolete: "secondary" };

export default function CalidadPage() {
  const [ncs, setNCs] = useState<NC[]>([]);
  const [audits, setAudits] = useState<QualityAudit[]>([]);
  const [docs, setDocs] = useState<ControlledDoc[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const [n, a, d] = await Promise.all([
        api.get<{ nonconformities: NC[] }>("/v1/erp/quality/nc?page_size=50"),
        api.get<{ audits: QualityAudit[] }>("/v1/erp/quality/audits?page_size=50"),
        api.get<{ documents: ControlledDoc[] }>("/v1/erp/quality/documents?page_size=50"),
      ]);
      setNCs(n.nonconformities); setAudits(a.audits); setDocs(d.documents);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);
  useEffect(() => { const unsub = wsManager.subscribe("erp_quality", fetch); return unsub; }, [fetch]);

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Calidad</h1>
          <p className="text-sm text-muted-foreground mt-0.5">No conformidades, auditorias y documentos controlados</p>
        </div>
        <Tabs defaultValue="nc">
          <TabsList className="mb-4">
            <TabsTrigger value="nc"><AlertTriangleIcon className="size-3.5 mr-1.5" />No Conformidades ({ncs.length})</TabsTrigger>
            <TabsTrigger value="audits"><ClipboardCheckIcon className="size-3.5 mr-1.5" />Auditorias</TabsTrigger>
            <TabsTrigger value="docs"><FileTextIcon className="size-3.5 mr-1.5" />Documentos</TabsTrigger>
          </TabsList>
          <TabsContent value="nc">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-20">NC</TableHead><TableHead className="w-28">Fecha</TableHead><TableHead>Descripcion</TableHead><TableHead className="w-20">Sev.</TableHead><TableHead className="w-28">Asignado</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{ncs.map((nc) => (<TableRow key={nc.id}><TableCell className="font-mono text-sm">{nc.number}</TableCell><TableCell className="text-sm text-muted-foreground">{fmtDate(nc.date)}</TableCell><TableCell className="text-sm truncate max-w-64">{nc.description}</TableCell><TableCell><Badge variant={sevColor[nc.severity] || "secondary"}>{nc.severity}</Badge></TableCell><TableCell className="text-sm">{nc.assigned_name || "—"}</TableCell><TableCell><Badge variant={statusColor[nc.status] || "secondary"}>{nc.status}</Badge></TableCell></TableRow>))}
                {ncs.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin no conformidades.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="audits">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-20">Nro</TableHead><TableHead className="w-28">Fecha</TableHead><TableHead className="w-24">Tipo</TableHead><TableHead>Alcance</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{audits.map((a) => (<TableRow key={a.id}><TableCell className="font-mono text-sm">{a.number}</TableCell><TableCell className="text-sm text-muted-foreground">{fmtDate(a.date)}</TableCell><TableCell><Badge variant="secondary">{a.audit_type}</Badge></TableCell><TableCell className="text-sm">{a.scope || "—"}</TableCell><TableCell><Badge variant={statusColor[a.status] || "secondary"}>{a.status}</Badge></TableCell></TableRow>))}
                {audits.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin auditorias.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
          <TabsContent value="docs">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table><TableHeader><TableRow><TableHead className="w-24">Codigo</TableHead><TableHead>Titulo</TableHead><TableHead className="w-16 text-center">Rev.</TableHead><TableHead className="w-28">Estado</TableHead></TableRow></TableHeader>
                <TableBody>{docs.map((d) => (<TableRow key={d.id}><TableCell className="font-mono text-sm">{d.code}</TableCell><TableCell className="text-sm">{d.title}</TableCell><TableCell className="text-center text-sm">{d.revision}</TableCell><TableCell><Badge variant={statusColor[d.status] || "secondary"}>{d.status}</Badge></TableCell></TableRow>))}
                {docs.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin documentos.</TableCell></TableRow>}</TableBody></Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
