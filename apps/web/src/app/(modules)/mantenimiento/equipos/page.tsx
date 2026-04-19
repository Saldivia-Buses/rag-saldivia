"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import type { MaintenanceAsset } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon } from "lucide-react";

const typeLabel: Record<string, string> = {
  vehicle: "Vehículo",
  machine: "Máquina",
  tool: "Herramienta",
  facility: "Instalación",
};

export default function EquiposPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: assets = [], isLoading, error } = useQuery({
    queryKey: erpKeys.maintenanceAssets(),
    queryFn: () => api.get<{ assets: MaintenanceAsset[] }>("/v1/erp/maintenance/assets"),
    select: (d) => d.assets,
  });

  const createMutation = useMutation({
    mutationFn: (data: { code: string; name: string; asset_type: string; location?: string }) =>
      api.post("/v1/erp/maintenance/assets", data),
    onSuccess: () => {
      toast.success("Equipo creado exitosamente");
      queryClient.invalidateQueries({ queryKey: erpKeys.maintenanceAssets() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando equipos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Equipos</h1>
            <p className="text-sm text-muted-foreground mt-0.5">{assets.length} equipos registrados</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo equipo
          </Button>
        </div>
        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader><TableRow>
              <TableHead className="w-28">Código</TableHead>
              <TableHead>Nombre</TableHead>
              <TableHead className="w-28">Tipo</TableHead>
              <TableHead>Ubicación</TableHead>
              <TableHead className="w-24 text-center">Estado</TableHead>
            </TableRow></TableHeader>
            <TableBody>
              {assets.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="font-mono text-sm">
                    <Link href={`/mantenimiento/equipos/${a.id}`} className="hover:underline">
                      {a.code}
                    </Link>
                  </TableCell>
                  <TableCell className="text-sm font-medium">{a.name}</TableCell>
                  <TableCell><Badge variant="secondary">{typeLabel[a.asset_type] || a.asset_type}</Badge></TableCell>
                  <TableCell className="text-sm text-muted-foreground">{a.location || "\u2014"}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={a.active ? "default" : "secondary"}>{a.active ? "Activo" : "Inactivo"}</Badge>
                  </TableCell>
                </TableRow>
              ))}
              {assets.length === 0 && (
                <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin equipos registrados.</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo equipo</DialogTitle></DialogHeader>
          <CreateAssetForm
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateAssetForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: { code: string; name: string; asset_type: string; location?: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [assetType, setAssetType] = useState("vehicle");
  const [location, setLocation] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (code && name) onSubmit({ code, name, asset_type: assetType, location: location || undefined });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Código</Label>
          <Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="EQ-001" />
        </div>
        <div className="space-y-2">
          <Label>Tipo</Label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm bg-card"
            value={assetType}
            onChange={(e) => setAssetType(e.target.value)}
          >
            <option value="vehicle">Vehículo</option>
            <option value="machine">Máquina</option>
            <option value="tool">Herramienta</option>
            <option value="facility">Instalación</option>
          </select>
        </div>
      </div>
      <div className="space-y-2">
        <Label>Nombre</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Nombre del equipo" />
      </div>
      <div className="space-y-2">
        <Label>Ubicación <span className="text-muted-foreground">(opcional)</span></Label>
        <Input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="Planta / Sector" />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!code || !name || isPending}>
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}
