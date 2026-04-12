"use client";

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, SearchIcon, FolderIcon, TagIcon } from "lucide-react";

interface Catalog {
  id: string; type: string; code: string; name: string; parent_id: string | null;
  sort_order: number; active: boolean; metadata: Record<string, unknown>;
  created_at: string; updated_at: string;
}

export default function CatalogosPage() {
  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: types = [], isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "catalogs", "types"] as const,
    queryFn: () => api.get<{ types: string[] }>("/v1/erp/catalogs/types"),
    select: (d) => d.types,
  });

  // Auto-select first type
  useEffect(() => {
    if (types.length > 0 && !selectedType) setSelectedType(types[0]);
  }, [types, selectedType]);

  const { data: catalogs = [] } = useQuery({
    queryKey: erpKeys.catalogs(selectedType ?? undefined),
    queryFn: () => api.get<{ catalogs: Catalog[] }>(`/v1/erp/catalogs?type=${encodeURIComponent(selectedType!)}&active=false`),
    enabled: !!selectedType,
    select: (d) => d.catalogs,
  });

  const createMutation = useMutation({
    mutationFn: (data: { type: string; code: string; name: string }) => api.post("/v1/erp/catalogs", data),
    onSuccess: () => {
      toast.success("Entrada creada exitosamente");
      queryClient.invalidateQueries({ queryKey: erpKeys.catalogs() });
      queryClient.invalidateQueries({ queryKey: [...erpKeys.all, "catalogs", "types"] });
      setCreateOpen(false);
    },
    onError: (err) => toast.error("Error al crear entrada", { description: err instanceof Error ? err.message : undefined }),
  });

  const toggleMutation = useMutation({
    mutationFn: (catalog: Catalog) => api.put(`/v1/erp/catalogs/${catalog.id}`, { code: catalog.code, name: catalog.name, sort_order: catalog.sort_order, active: !catalog.active }),
    onSuccess: (_data, catalog) => {
      toast.success(catalog.active ? "Entrada desactivada" : "Entrada activada");
      queryClient.invalidateQueries({ queryKey: erpKeys.catalogs(selectedType ?? undefined) });
    },
    onError: (err) => toast.error("Error al actualizar", { description: err instanceof Error ? err.message : undefined }),
  });

  const filtered = catalogs.filter((c) => c.name.toLowerCase().includes(search.toLowerCase()) || c.code.toLowerCase().includes(search.toLowerCase()));
  const typeLabel = (t: string) => t.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());

  if (error) return <ErrorState message="Error cargando catálogos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 overflow-y-auto"><div className="mx-auto w-full max-w-6xl px-6 py-8"><Skeleton className="h-8 w-48 mb-6" /><div className="flex gap-6"><Skeleton className="h-[500px] w-56" /><Skeleton className="h-[500px] flex-1" /></div></div></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Catálogos</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Tablas de referencia del sistema — {types.length} tipos, {catalogs.length} entradas</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva entrada</Button>
        </div>

        <div className="flex gap-6">
          <div className="w-56 shrink-0">
            <ScrollArea className="h-[calc(100vh-12rem)]">
              <div className="space-y-1">
                {types.map((t) => (
                  <button key={t} onClick={() => setSelectedType(t)} className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors flex items-center gap-2 ${selectedType === t ? "bg-primary/10 text-primary font-medium" : "text-muted-foreground hover:bg-muted/50"}`}>
                    <FolderIcon className="size-3.5 shrink-0" /><span className="truncate">{typeLabel(t)}</span>
                  </button>
                ))}
                {types.length === 0 && <p className="text-sm text-muted-foreground px-3 py-4">Sin catálogos todavía. Creá el primero.</p>}
              </div>
            </ScrollArea>
          </div>

          <div className="flex-1 min-w-0">
            {selectedType ? (
              <>
                <div className="flex items-center gap-3 mb-4">
                  <div className="relative flex-1">
                    <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                    <Input placeholder={`Buscar en ${typeLabel(selectedType)}...`} className="pl-9 bg-card" value={search} onChange={(e) => setSearch(e.target.value)} />
                  </div>
                  <Badge variant="secondary">{filtered.length} entradas</Badge>
                </div>
                <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
                  <Table>
                    <TableHeader><TableRow>
                      <TableHead className="w-28">Código</TableHead><TableHead>Nombre</TableHead>
                      <TableHead className="w-20 text-center">Orden</TableHead><TableHead className="w-24 text-center">Estado</TableHead>
                      <TableHead className="w-24 text-right">Acción</TableHead>
                    </TableRow></TableHeader>
                    <TableBody>
                      {filtered.length > 0 ? filtered.map((c) => (
                        <TableRow key={c.id} className={!c.active ? "opacity-50" : ""}>
                          <TableCell><div className="flex items-center gap-1.5"><TagIcon className="size-3.5 text-muted-foreground" /><span className="font-mono text-sm">{c.code}</span></div></TableCell>
                          <TableCell className="text-sm">{c.name}</TableCell>
                          <TableCell className="text-center text-sm text-muted-foreground">{c.sort_order}</TableCell>
                          <TableCell className="text-center"><Badge variant={c.active ? "default" : "secondary"}>{c.active ? "Activo" : "Inactivo"}</Badge></TableCell>
                          <TableCell className="text-right"><Button variant="ghost" size="sm" onClick={() => toggleMutation.mutate(c)}>{c.active ? "Desactivar" : "Activar"}</Button></TableCell>
                        </TableRow>
                      )) : (
                        <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">{search ? "No se encontraron entradas." : "Este catálogo está vacío."}</TableCell></TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">Seleccioná un tipo de catálogo de la lista.</div>
            )}
          </div>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva entrada de catálogo</DialogTitle></DialogHeader>
          <CreateCatalogForm types={types} defaultType={selectedType} onSubmit={(d) => createMutation.mutate(d)} isPending={createMutation.isPending} onClose={() => setCreateOpen(false)} />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateCatalogForm({ types, defaultType, onSubmit, isPending, onClose }: {
  types: string[]; defaultType: string | null;
  onSubmit: (d: { type: string; code: string; name: string }) => void;
  isPending: boolean; onClose: () => void;
}) {
  const [type, setType] = useState(defaultType || "");
  const [code, setCode] = useState(""); const [name, setName] = useState("");

  useEffect(() => { setType(defaultType || ""); setCode(""); setName(""); }, [defaultType]);

  return (
    <form onSubmit={(e) => { e.preventDefault(); if (type.trim() && code.trim() && name.trim()) onSubmit({ type: type.trim(), code: code.trim(), name: name.trim() }); }} className="space-y-4">
      <div className="space-y-2"><Label>Tipo</Label><Input value={type} onChange={(e) => setType(e.target.value)} placeholder="Ej: province, currency" list="catalog-types" /><datalist id="catalog-types">{types.map((t) => <option key={t} value={t} />)}</datalist></div>
      <div className="space-y-2"><Label>Código</Label><Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Ej: BA, ARS" /></div>
      <div className="space-y-2"><Label>Nombre</Label><Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Ej: Buenos Aires" /></div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!type.trim() || !code.trim() || !name.trim() || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
