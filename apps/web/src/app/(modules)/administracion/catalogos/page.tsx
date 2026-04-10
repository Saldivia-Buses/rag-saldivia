"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, SearchIcon, FolderIcon, TagIcon } from "lucide-react";

// ─── Types ──────────────────────────────────────────────────────────────────

interface Catalog {
  id: string;
  type: string;
  code: string;
  name: string;
  parent_id: string | null;
  sort_order: number;
  active: boolean;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// ─── Page ───────────────────────────────────────────────────────────────────

export default function CatalogosPage() {
  const [types, setTypes] = useState<string[]>([]);
  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [catalogs, setCatalogs] = useState<Catalog[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  // Fetch available catalog types
  const fetchTypes = useCallback(async () => {
    try {
      const data = await api.get<{ types: string[] }>("/v1/erp/catalogs/types");
      setTypes(data.types);
      if (data.types.length > 0 && !selectedType) {
        setSelectedType(data.types[0]);
      }
    } catch (err) {
      console.error("Failed to fetch catalog types:", err);
    } finally {
      setLoading(false);
    }
  }, [selectedType]);

  // Fetch catalogs for selected type
  const fetchCatalogs = useCallback(async () => {
    if (!selectedType) return;
    try {
      const data = await api.get<{ catalogs: Catalog[] }>(
        `/v1/erp/catalogs?type=${encodeURIComponent(selectedType)}&active=false`
      );
      setCatalogs(data.catalogs);
    } catch (err) {
      console.error("Failed to fetch catalogs:", err);
    }
  }, [selectedType]);

  useEffect(() => { fetchTypes(); }, [fetchTypes]);
  useEffect(() => { fetchCatalogs(); }, [fetchCatalogs]);

  // Real-time updates
  useEffect(() => {
    const handler = () => { fetchTypes(); fetchCatalogs(); };
    const unsubscribe = wsManager.subscribe("erp_catalogs", handler);
    return unsubscribe;
  }, [fetchTypes, fetchCatalogs]);

  // Create catalog entry
  const handleCreate = async (code: string, name: string, type: string) => {
    await api.post("/v1/erp/catalogs", { type: type || selectedType, code, name });
    setCreateOpen(false);
    fetchCatalogs();
    fetchTypes();
  };

  // Toggle active status
  const handleToggleActive = async (catalog: Catalog) => {
    await api.put(`/v1/erp/catalogs/${catalog.id}`, {
      code: catalog.code,
      name: catalog.name,
      sort_order: catalog.sort_order,
      active: !catalog.active,
    });
    fetchCatalogs();
  };

  // Filtered catalogs
  const filtered = catalogs.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.code.toLowerCase().includes(search.toLowerCase())
  );

  const typeLabel = (t: string) => t.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());

  if (loading) {
    return (
      <div className="flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-6xl px-6 py-8">
          <Skeleton className="h-8 w-48 mb-6" />
          <div className="flex gap-6">
            <Skeleton className="h-[500px] w-56" />
            <Skeleton className="h-[500px] flex-1" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Catálogos</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Tablas de referencia del sistema — {types.length} tipos, {catalogs.length} entradas
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />
            Nueva entrada
          </Button>
        </div>

        <div className="flex gap-6">
          {/* Sidebar: type list */}
          <div className="w-56 shrink-0">
            <ScrollArea className="h-[calc(100vh-12rem)]">
              <div className="space-y-1">
                {types.map((t) => (
                  <button
                    key={t}
                    onClick={() => setSelectedType(t)}
                    className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors flex items-center gap-2 ${
                      selectedType === t
                        ? "bg-primary/10 text-primary font-medium"
                        : "text-muted-foreground hover:bg-muted/50"
                    }`}
                  >
                    <FolderIcon className="size-3.5 shrink-0" />
                    <span className="truncate">{typeLabel(t)}</span>
                  </button>
                ))}
                {types.length === 0 && (
                  <p className="text-sm text-muted-foreground px-3 py-4">
                    Sin catálogos todavía. Creá el primero.
                  </p>
                )}
              </div>
            </ScrollArea>
          </div>

          {/* Main: catalog entries table */}
          <div className="flex-1 min-w-0">
            {selectedType ? (
              <>
                <div className="flex items-center gap-3 mb-4">
                  <div className="relative flex-1">
                    <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                    <Input
                      placeholder={`Buscar en ${typeLabel(selectedType)}...`}
                      className="pl-9 bg-card"
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                    />
                  </div>
                  <Badge variant="secondary">{filtered.length} entradas</Badge>
                </div>

                <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-28">Código</TableHead>
                        <TableHead>Nombre</TableHead>
                        <TableHead className="w-20 text-center">Orden</TableHead>
                        <TableHead className="w-24 text-center">Estado</TableHead>
                        <TableHead className="w-24 text-right">Acción</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filtered.length > 0 ? (
                        filtered.map((c) => (
                          <TableRow key={c.id} className={!c.active ? "opacity-50" : ""}>
                            <TableCell>
                              <div className="flex items-center gap-1.5">
                                <TagIcon className="size-3.5 text-muted-foreground" />
                                <span className="font-mono text-sm">{c.code}</span>
                              </div>
                            </TableCell>
                            <TableCell className="text-sm">{c.name}</TableCell>
                            <TableCell className="text-center text-sm text-muted-foreground">
                              {c.sort_order}
                            </TableCell>
                            <TableCell className="text-center">
                              <Badge variant={c.active ? "default" : "secondary"}>
                                {c.active ? "Activo" : "Inactivo"}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-right">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleToggleActive(c)}
                              >
                                {c.active ? "Desactivar" : "Activar"}
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                            {search
                              ? "No se encontraron entradas."
                              : "Este catálogo está vacío."}
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
                Seleccioná un tipo de catálogo de la lista.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create dialog */}
      <CreateDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreate={handleCreate}
        defaultType={selectedType}
        types={types}
      />
    </div>
  );
}

// ─── Create Dialog ──────────────────────────────────────────────────────────

function CreateDialog({
  open,
  onClose,
  onCreate,
  defaultType,
  types,
}: {
  open: boolean;
  onClose: () => void;
  onCreate: (code: string, name: string, type: string) => void;
  defaultType: string | null;
  types: string[];
}) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [type, setType] = useState(defaultType || "");

  useEffect(() => {
    if (open) {
      setCode("");
      setName("");
      setType(defaultType || "");
    }
  }, [open, defaultType]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim() || !name.trim() || !type.trim()) return;
    onCreate(code.trim(), name.trim(), type.trim());
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Nueva entrada de catálogo</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Tipo</Label>
            <Input
              value={type}
              onChange={(e) => setType(e.target.value)}
              placeholder="Ej: province, currency, payment_method"
              list="catalog-types"
            />
            <datalist id="catalog-types">
              {types.map((t) => (
                <option key={t} value={t} />
              ))}
            </datalist>
          </div>
          <div className="space-y-2">
            <Label>Código</Label>
            <Input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="Ej: BA, ARS, cash"
            />
          </div>
          <div className="space-y-2">
            <Label>Nombre</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ej: Buenos Aires, Peso Argentino"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancelar
            </Button>
            <Button type="submit" disabled={!code.trim() || !name.trim() || !type.trim()}>
              Crear
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
