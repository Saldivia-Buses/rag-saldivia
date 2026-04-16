"use client";

import { useState, useEffect } from "react";
import { Database, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api/client";

type Collection = {
  id: string;
  name: string;
  description: string | null;
  created_at: string;
};

export default function CollectionsPage() {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");

  const fetchCollections = async () => {
    try {
      const data = await api.get<Collection[]>("/v1/ingest/collections");
      setCollections(data);
    } catch {
      // empty on error
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCollections();
  }, []);

  const handleCreate = async () => {
    if (!name.trim()) return;
    setCreating(true);
    setError("");
    try {
      await api.post<Collection>("/v1/ingest/collections", {
        name: name.trim(),
        description: description.trim(),
      });
      setName("");
      setDescription("");
      setOpen(false);
      fetchCollections();
    } catch (e: any) {
      setError(e?.message || "Error al crear colección");
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-8">
      <div className="mx-auto w-full max-w-4xl">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-lg font-semibold">Colecciones</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Base de conocimiento empresarial
            </p>
          </div>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger
              render={
                <Button size="sm" leadingIcon={Plus}>
                  Nueva colección
                </Button>
              }
            />
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Crear colección</DialogTitle>
              </DialogHeader>
              <div className="space-y-4 pt-2">
                <div className="space-y-2">
                  <Label htmlFor="col-name" className="text-sm">Nombre</Label>
                  <Input
                    id="col-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Ej: Manuales técnicos"
                    className="h-10"
                    autoFocus
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="col-desc" className="text-sm">Descripción (opcional)</Label>
                  <Input
                    id="col-desc"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Breve descripción de la colección"
                    className="h-10"
                  />
                </div>
                {error && (
                  <p className="text-sm text-destructive bg-destructive/10 px-3 py-2 rounded-lg">
                    {error}
                  </p>
                )}
                <Button
                  onClick={handleCreate}
                  disabled={creating || !name.trim()}
                  loading={creating}
                  className="w-full"
                >
                  Crear colección
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>

        {/* Content */}
        {loading ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="rounded-xl border border-border/40 bg-card p-5 animate-pulse">
                <div className="h-4 bg-muted rounded w-2/3 mb-3" />
                <div className="h-3 bg-muted rounded w-full" />
              </div>
            ))}
          </div>
        ) : collections.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <div className="flex size-14 items-center justify-center rounded-2xl bg-card border border-border/40 mb-4">
              <Database className="size-7 text-muted-foreground" />
            </div>
            <h2 className="text-base font-medium">No hay colecciones todavía</h2>
            <p className="text-sm text-muted-foreground mt-1 max-w-sm">
              Las colecciones organizan tus documentos para que la IA pueda consultarlos.
            </p>
            <Button
              variant="outline"
              size="sm"
              className="mt-4"
              onClick={() => setOpen(true)}
              leadingIcon={Plus}
            >
              Crear la primera
            </Button>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {collections.map((col) => (
              <div
                key={col.id}
                className="group rounded-xl border border-border/40 bg-card p-5 transition-all hover:border-border hover:shadow-sm cursor-pointer"
              >
                <div className="flex items-start gap-3">
                  <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 mt-0.5">
                    <Database className="size-4 text-primary" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <h3 className="font-medium text-sm group-hover:text-foreground transition-colors">
                      {col.name}
                    </h3>
                    {col.description && (
                      <p className="text-sm text-muted-foreground mt-0.5 line-clamp-2">
                        {col.description}
                      </p>
                    )}
                    <p className="text-[11px] text-muted-foreground/60 mt-2">
                      {new Date(col.created_at).toLocaleDateString("es-AR", {
                        day: "numeric",
                        month: "short",
                        year: "numeric",
                      })}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
