"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, ShieldCheckIcon } from "lucide-react";

interface ControlledDoc {
  id: string; code: string; title: string; revision: number; status: string;
}

const statusVariant: Record<string, "default" | "secondary" | "outline"> = {
  draft: "secondary",
  approved: "default",
  obsolete: "outline",
};

const statusLabel: Record<string, string> = {
  draft: "Borrador",
  approved: "Aprobado",
  obsolete: "Obsoleto",
};

export default function LegalPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: docs = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "quality", "documents"] as const,
    queryFn: () => api.get<{ documents: ControlledDoc[] }>("/v1/erp/quality/documents?page_size=50"),
    select: (d) => d.documents,
  });

  const createMutation = useMutation({
    mutationFn: (data: { Code: string; Title: string; Revision: number }) =>
      api.post("/v1/erp/quality/documents", data),
    onSuccess: () => {
      toast.success("Documento creado");
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "quality", "documents"] });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando documentos legales" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Homologación y legal</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Certificaciones, homologaciones y documentación legal — {docs.length} documentos</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nuevo documento</Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-32">Código</TableHead>
                <TableHead>Título</TableHead>
                <TableHead className="w-24 text-center">Revisión</TableHead>
                <TableHead className="w-28 text-center">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {docs.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-mono text-sm">{d.code}</TableCell>
                  <TableCell className="text-sm">{d.title}</TableCell>
                  <TableCell className="text-center text-sm font-mono">{d.revision}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={statusVariant[d.status] ?? "secondary"}>
                      {statusLabel[d.status] ?? d.status}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
              {docs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                    <div className="flex flex-col items-center gap-2">
                      <ShieldCheckIcon className="size-8 text-muted-foreground/40" />
                      <span>Sin documentos. Creá el primero.</span>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo documento legal</DialogTitle></DialogHeader>
          <CreateDocumentForm
            onSubmit={(code, title, revision) => createMutation.mutate({ Code: code, Title: title, Revision: revision })}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateDocumentForm({ onSubmit, isPending, onClose }: {
  onSubmit: (code: string, title: string, revision: number) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [code, setCode] = useState("");
  const [title, setTitle] = useState("");
  const [revision, setRevision] = useState("1");

  const canSubmit = code.trim() && title.trim();

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (canSubmit) onSubmit(code.trim(), title.trim(), parseInt(revision) || 1);
      }}
      className="space-y-4"
    >
      <div className="space-y-2"><Label>Código</Label><Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Ej: CERT-001" /></div>
      <div className="space-y-2"><Label>Título</Label><Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Ej: Homologación CNRT Resolución 85" /></div>
      <div className="space-y-2">
        <Label>Revisión</Label>
        <Input type="number" min="1" value={revision} onChange={(e) => setRevision(e.target.value)} />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
