"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { EmptyState } from "@/components/erp/empty-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, SearchIcon, BoxIcon } from "lucide-react";

interface CarroceriaModel {
  id: string;
  code: string;
  model_code: string;
  description: string;
  abbreviation: string;
  double_deck: boolean;
  axle_weight_pct: number;
  active: boolean;
}

const manufacturingKeys = {
  carroceriaModels: () => [...erpKeys.all, "manufacturing", "carroceria-models"] as const,
};

export default function ProductoPage() {
  const [search, setSearch] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: models = [], isLoading, error } = useQuery({
    queryKey: manufacturingKeys.carroceriaModels(),
    queryFn: () =>
      api.get<{ carroceria_models: CarroceriaModel[] }>(
        "/v1/erp/manufacturing/carroceria-models"
      ),
    select: (d) => d.carroceria_models,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      code: string;
      model_code: string;
      description: string;
      abbreviation?: string;
      double_deck?: boolean;
    }) => api.post("/v1/erp/manufacturing/carroceria-models", data),
    onSuccess: () => {
      toast.success("Modelo de carrocería creado exitosamente");
      queryClient.invalidateQueries({ queryKey: manufacturingKeys.carroceriaModels() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const filtered = models.filter(
    (m) =>
      m.description.toLowerCase().includes(search.toLowerCase()) ||
      m.code.toLowerCase().includes(search.toLowerCase()) ||
      m.model_code.toLowerCase().includes(search.toLowerCase())
  );

  if (error) return <ErrorState message="Error cargando modelos de carrocería" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Modelos de carrocería</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Especificaciones de modelos de carrocería — {models.length} modelos
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo modelo
          </Button>
        </div>

        <div className="relative mb-4">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Buscar por código o descripción..."
            className="pl-9 bg-card"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-32">Código</TableHead>
                <TableHead className="w-36">Modelo</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead className="w-24">Abrev.</TableHead>
                <TableHead className="w-28 text-center">Doble piso</TableHead>
                <TableHead className="w-24 text-center">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="font-mono text-sm">{m.code}</TableCell>
                  <TableCell className="font-mono text-sm">{m.model_code}</TableCell>
                  <TableCell className="text-sm">{m.description}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{m.abbreviation || "—"}</TableCell>
                  <TableCell className="text-center">
                    {m.double_deck ? (
                      <Badge variant="secondary">Doble piso</Badge>
                    ) : (
                      <span className="text-xs text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell className="text-center">
                    <Badge variant={m.active ? "default" : "secondary"}>
                      {m.active ? "Activo" : "Inactivo"}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6}>
                    <EmptyState
                      icon={BoxIcon}
                      title={search ? "Sin resultados" : "Sin modelos"}
                      description={
                        search
                          ? "No se encontraron modelos con ese criterio."
                          : "Creá el primer modelo de carrocería para empezar."
                      }
                      action={!search ? { label: "Nuevo modelo", onClick: () => setCreateOpen(true) } : undefined}
                    />
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo modelo de carrocería</DialogTitle></DialogHeader>
          <CreateCarroceriaModelForm
            onSubmit={(d) => createMutation.mutate(d)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateCarroceriaModelForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (d: {
    code: string;
    model_code: string;
    description: string;
    abbreviation?: string;
    double_deck?: boolean;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [code, setCode] = useState("");
  const [modelCode, setModelCode] = useState("");
  const [description, setDescription] = useState("");
  const [abbreviation, setAbbreviation] = useState("");
  const [doubleDeck, setDoubleDeck] = useState(false);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (code.trim() && modelCode.trim() && description.trim()) {
          const d: {
            code: string;
            model_code: string;
            description: string;
            abbreviation?: string;
            double_deck?: boolean;
          } = {
            code: code.trim(),
            model_code: modelCode.trim(),
            description: description.trim(),
          };
          if (abbreviation.trim()) d.abbreviation = abbreviation.trim();
          if (doubleDeck) d.double_deck = true;
          onSubmit(d);
        }
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Código</Label>
        <Input
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="Ej: CAR-210"
        />
      </div>
      <div className="space-y-2">
        <Label>Modelo</Label>
        <Input
          value={modelCode}
          onChange={(e) => setModelCode(e.target.value)}
          placeholder="Ej: O500-U"
        />
      </div>
      <div className="space-y-2">
        <Label>Descripción</Label>
        <Input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Ej: Colectivo urbano 12m"
        />
      </div>
      <div className="space-y-2">
        <Label>Abreviación <span className="text-muted-foreground">(opcional)</span></Label>
        <Input
          value={abbreviation}
          onChange={(e) => setAbbreviation(e.target.value)}
          placeholder="Ej: URB-12"
        />
      </div>
      <div className="flex items-center gap-2">
        <Checkbox
          id="double-deck"
          checked={doubleDeck}
          onCheckedChange={(v) => setDoubleDeck(v === true)}
        />
        <Label htmlFor="double-deck" className="cursor-pointer">Doble piso</Label>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button
          type="submit"
          disabled={!code.trim() || !modelCode.trim() || !description.trim() || isPending}
        >
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}
