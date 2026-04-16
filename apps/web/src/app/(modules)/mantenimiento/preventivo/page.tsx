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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PlusIcon } from "lucide-react";

export default function PreventivoPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [selectedAsset, setSelectedAsset] = useState<string>("");
  const queryClient = useQueryClient();

  const { data: assets = [], isLoading: loadingAssets } = useQuery({
    queryKey: [...erpKeys.all, "maintenance", "assets"] as const,
    queryFn: () => api.get<{ assets: any[] }>("/v1/erp/maintenance/assets"),
    select: (d) => d.assets,
  });

  const { data: plans = [], isLoading: loadingPlans, error } = useQuery({
    queryKey: [...erpKeys.all, "maintenance", "plans", selectedAsset] as const,
    queryFn: () =>
      api.get<{ plans: any[] }>(`/v1/erp/maintenance/assets/${selectedAsset}/plans`),
    select: (d) => d.plans,
    enabled: !!selectedAsset,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      asset_id: string;
      name: string;
      frequency_days?: string;
      frequency_km?: string;
      frequency_hours?: string;
    }) => api.post(`/v1/erp/maintenance/assets/${data.asset_id}/plans`, data),
    onSuccess: () => {
      toast.success("Plan de mantenimiento creado");
      queryClient.invalidateQueries({
        queryKey: [...erpKeys.all, "maintenance", "plans", selectedAsset],
      });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const isLoading = loadingAssets || (!!selectedAsset && loadingPlans);
  if (error) return <ErrorState message="Error cargando planes" onRetry={() => window.location.reload()} />;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Mantenimiento Preventivo</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Planes de mantenimiento por frecuencia, kilometraje u horas de uso
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)} disabled={assets.length === 0}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo plan
          </Button>
        </div>

        <div className="mb-4 flex items-center gap-3">
          <Label className="text-sm shrink-0">Equipo</Label>
          <Select value={selectedAsset} onValueChange={(v) => setSelectedAsset(v ?? "")}>
            <SelectTrigger className="w-64 bg-card">
              <SelectValue placeholder="Seleccionar equipo..." />
            </SelectTrigger>
            <SelectContent>
              {assets.map((a: any) => (
                <SelectItem key={a.id} value={a.id}>{a.code} — {a.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {isLoading ? (
          <Skeleton className="h-[400px]" />
        ) : !selectedAsset ? (
          <div className="rounded-xl border border-border/40 bg-card h-40 flex items-center justify-center">
            <p className="text-sm text-muted-foreground">Seleccioná un equipo para ver sus planes de mantenimiento.</p>
          </div>
        ) : (
          <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
            <Table>
              <TableHeader><TableRow>
                <TableHead>Plan</TableHead>
                <TableHead className="w-28 text-right">Frec. días</TableHead>
                <TableHead className="w-28 text-right">Frec. km</TableHead>
                <TableHead className="w-28 text-right">Frec. hs</TableHead>
                <TableHead className="w-28">Próxima vez</TableHead>
                <TableHead className="w-24 text-center">Estado</TableHead>
              </TableRow></TableHeader>
              <TableBody>
                {plans.map((p: any) => (
                  <TableRow key={p.id}>
                    <TableCell className="text-sm font-medium">{p.name}</TableCell>
                    <TableCell className="text-right font-mono text-sm">
                      {p.frequency_days?.Int ?? "\u2014"}
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm">
                      {p.frequency_km?.Int ?? "\u2014"}
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm">
                      {p.frequency_hours?.Int ?? "\u2014"}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {p.next_due?.Time ? fmtDateShort(p.next_due.Time) : "\u2014"}
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge variant={p.active ? "default" : "secondary"}>
                        {p.active ? "Activo" : "Inactivo"}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
                {plans.length === 0 && (
                  <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                    Sin planes para este equipo.
                  </TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo plan de mantenimiento</DialogTitle></DialogHeader>
          <CreatePlanForm
            assets={assets}
            defaultAssetId={selectedAsset}
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreatePlanForm({
  assets,
  defaultAssetId,
  onSubmit,
  isPending,
  onClose,
}: {
  assets: any[];
  defaultAssetId: string;
  onSubmit: (data: { asset_id: string; name: string; frequency_days?: string; frequency_km?: string; frequency_hours?: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [assetId, setAssetId] = useState(defaultAssetId);
  const [name, setName] = useState("");
  const [freqDays, setFreqDays] = useState("");
  const [freqKm, setFreqKm] = useState("");
  const [freqHours, setFreqHours] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!assetId || !name) return;
        onSubmit({
          asset_id: assetId,
          name,
          frequency_days: freqDays || undefined,
          frequency_km: freqKm || undefined,
          frequency_hours: freqHours || undefined,
        });
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Equipo</Label>
        <Select value={assetId} onValueChange={(v) => setAssetId(v ?? "")}>
          <SelectTrigger className="bg-card">
            <SelectValue placeholder="Seleccionar..." />
          </SelectTrigger>
          <SelectContent>
            {assets.map((a: any) => (
              <SelectItem key={a.id} value={a.id}>{a.code} — {a.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Nombre del plan</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Ej: Cambio de aceite" />
      </div>
      <div className="grid grid-cols-3 gap-3">
        <div className="space-y-2">
          <Label className="text-xs">Cada N días</Label>
          <Input type="number" min="1" value={freqDays} onChange={(e) => setFreqDays(e.target.value)} placeholder="90" />
        </div>
        <div className="space-y-2">
          <Label className="text-xs">Cada N km</Label>
          <Input type="number" min="1" value={freqKm} onChange={(e) => setFreqKm(e.target.value)} placeholder="5000" />
        </div>
        <div className="space-y-2">
          <Label className="text-xs">Cada N hs</Label>
          <Input type="number" min="1" value={freqHours} onChange={(e) => setFreqHours(e.target.value)} placeholder="200" />
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!assetId || !name || isPending}>
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}
