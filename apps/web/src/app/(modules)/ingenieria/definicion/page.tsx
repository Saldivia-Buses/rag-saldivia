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
import { PlusIcon, TagIcon } from "lucide-react";

interface Catalog {
  id: string; type: string; code: string; name: string;
  sort_order: number; active: boolean;
}

export default function DefinicionPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: parts = [], isLoading, error } = useQuery({
    queryKey: erpKeys.catalogs("parts"),
    queryFn: () => api.get<{ catalogs: Catalog[] }>("/v1/erp/catalogs?type=parts&active=false&page_size=100"),
    select: (d) => d.catalogs,
  });

  const createMutation = useMutation({
    mutationFn: (data: { type: string; code: string; name: string }) => api.post("/v1/erp/catalogs", data),
    onSuccess: () => {
      toast.success("Definición creada");
      queryClient.invalidateQueries({ queryKey: erpKeys.catalogs("parts") });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando definiciones" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Definición técnica</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Catálogo de partes y componentes — {parts.length} definiciones</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva definición</Button>
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-32">Código</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead className="w-28 text-center">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {parts.map((p) => (
                <TableRow key={p.id} className={!p.active ? "opacity-50" : ""}>
                  <TableCell>
                    <div className="flex items-center gap-1.5">
                      <TagIcon className="size-3.5 text-muted-foreground" />
                      <span className="font-mono text-sm">{p.code}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm">{p.name}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={p.active ? "default" : "secondary"}>
                      {p.active ? "Activo" : "Inactivo"}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
              {parts.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="h-24 text-center text-muted-foreground">
                    Sin definiciones. Creá la primera.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva definición técnica</DialogTitle></DialogHeader>
          <CreateDefinitionForm
            onSubmit={(code, name) => createMutation.mutate({ type: "parts", code, name })}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateDefinitionForm({ onSubmit, isPending, onClose }: {
  onSubmit: (code: string, name: string) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");

  return (
    <form
      onSubmit={(e) => { e.preventDefault(); if (code.trim() && name.trim()) onSubmit(code.trim(), name.trim()); }}
      className="space-y-4"
    >
      <div className="space-y-2"><Label>Código</Label><Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Ej: PART-001" /></div>
      <div className="space-y-2"><Label>Nombre</Label><Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Ej: Eje delantero" /></div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!code.trim() || !name.trim() || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
