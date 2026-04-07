"use client";

import { useState, useEffect } from "react";
import { Plus } from "lucide-react";
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
    <div className="flex-1 p-8">
      <div className="flex items-center mb-6">
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="h-4 w-4" />
              Nueva colección
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Crear colección</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 pt-2">
              <div className="space-y-2">
                <Label htmlFor="name">Nombre</Label>
                <Input
                  id="name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Ej: Manuales técnicos"
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="desc">Descripción (opcional)</Label>
                <Input
                  id="desc"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Breve descripción de la colección"
                />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <Button onClick={handleCreate} disabled={creating || !name.trim()} className="w-full">
                {creating ? "Creando..." : "Crear"}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {loading ? (
        <p className="text-muted-foreground text-sm">Cargando...</p>
      ) : collections.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <p className="text-muted-foreground mb-4">No hay colecciones todavía</p>
          <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Crear la primera
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {collections.map((col) => (
            <div
              key={col.id}
              className="rounded-lg border p-4 hover:border-foreground/20 transition-colors"
            >
              <h3 className="font-medium">{col.name}</h3>
              {col.description && (
                <p className="text-sm text-muted-foreground mt-1">{col.description}</p>
              )}
              <p className="text-xs text-muted-foreground/60 mt-3">
                {new Date(col.created_at).toLocaleDateString("es-AR")}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
