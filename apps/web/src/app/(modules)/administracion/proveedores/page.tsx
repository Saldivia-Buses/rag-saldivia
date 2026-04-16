"use client";

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { useERPSearch } from "@/lib/erp/use-erp-search";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { EmptyState } from "@/components/erp/empty-state";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, SearchIcon, BuildingIcon, PencilIcon, TrashIcon } from "lucide-react";

interface Entity {
  id: string;
  code: string;
  name: string;
  entity_type: "customer" | "supplier" | "employee" | "carrier" | "other";
  tax_id: string | null;
  email: string | null;
  phone: string | null;
  address: string | null;
  city: string | null;
  active: boolean;
  created_at: string;
}

interface EntityBody {
  code: string;
  name: string;
  tax_id?: string;
  email?: string;
  phone?: string;
  address?: string;
  city?: string;
}

interface CreateEntityInput extends EntityBody {
  entity_type: string;
}

export default function ProveedoresPage() {
  const { search, setSearch, deferredSearch } = useERPSearch(0);
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Entity | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Entity | null>(null);
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.entities("supplier", deferredSearch || undefined),
    queryFn: () => {
      const q = new URLSearchParams({ type: "supplier", page: "1", page_size: "50" });
      if (deferredSearch) q.set("search", deferredSearch);
      return api.get<{ entities: Entity[]; total: number }>(`/v1/erp/entities?${q}`);
    },
  });

  const entities = data?.entities ?? [];
  const total = data?.total ?? 0;

  const createMutation = useMutation({
    mutationFn: (d: CreateEntityInput) => api.post("/v1/erp/entities", d),
    onSuccess: () => {
      toast.success("Proveedor creado exitosamente");
      queryClient.invalidateQueries({ queryKey: erpKeys.entities("supplier") });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, ...data }: { id: string } & EntityBody) =>
      api.put(`/v1/erp/entities/${id}`, data),
    onSuccess: () => {
      toast.success("Proveedor actualizado");
      queryClient.invalidateQueries({ queryKey: erpKeys.entities("supplier") });
      setEditTarget(null);
    },
    onError: permissionErrorToast,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/v1/erp/entities/${id}`),
    onSuccess: () => {
      toast.success("Proveedor eliminado");
      queryClient.invalidateQueries({ queryKey: erpKeys.entities("supplier") });
      setDeleteTarget(null);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando proveedores" onRetry={() => window.location.reload()} />;

  if (isLoading) {
    return (
      <div className="flex-1 p-8">
        <Skeleton className="h-8 w-48 mb-6" />
        <Skeleton className="h-[600px]" />
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Proveedores</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Registro de proveedores — {total} registros</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo proveedor
          </Button>
        </div>

        <div className="relative mb-4">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Buscar por nombre o código..."
            className="pl-9 bg-card"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-28">Código</TableHead>
                <TableHead>Nombre / Razón social</TableHead>
                <TableHead className="w-36">CUIT</TableHead>
                <TableHead className="w-48">Email</TableHead>
                <TableHead className="w-36">Teléfono</TableHead>
                <TableHead className="w-36">Ciudad</TableHead>
                <TableHead className="w-20 text-center">Estado</TableHead>
                <TableHead className="w-20 text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {entities.map((e) => (
                <TableRow key={e.id}>
                  <TableCell className="font-mono text-sm">{e.code}</TableCell>
                  <TableCell className="text-sm font-medium">{e.name}</TableCell>
                  <TableCell className="font-mono text-sm text-muted-foreground">{e.tax_id ?? "—"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground truncate max-w-48">{e.email ?? "—"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{e.phone ?? "—"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">{e.city ?? "—"}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={e.active ? "default" : "secondary"}>
                      {e.active ? "Activo" : "Inactivo"}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button size="sm" variant="ghost" onClick={() => setEditTarget(e)}>
                        <PencilIcon className="size-4" />
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => setDeleteTarget(e)}>
                        <TrashIcon className="size-4 text-destructive" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {entities.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8}>
                    <EmptyState
                      icon={BuildingIcon}
                      title={search ? "Sin resultados" : "Sin proveedores"}
                      description={search ? `No se encontraron proveedores para "${search}".` : "Creá el primer proveedor para empezar."}
                      action={!search ? { label: "Nuevo proveedor", onClick: () => setCreateOpen(true) } : undefined}
                    />
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <ProveedorDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSubmit={(d) => createMutation.mutate({ entity_type: "supplier", ...d })}
        isPending={createMutation.isPending}
      />

      <ProveedorDialog
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        initialValues={editTarget ?? undefined}
        onSubmit={(d) => {
          if (editTarget) updateMutation.mutate({ id: editTarget.id, ...d });
        }}
        isPending={updateMutation.isPending}
      />

      <AlertDialog open={!!deleteTarget} onOpenChange={(v) => !v && setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>¿Eliminar proveedor {deleteTarget?.name}?</AlertDialogTitle>
            <AlertDialogDescription>
              Esta acción no se puede deshacer. El proveedor será eliminado permanentemente.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancelar</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? "Eliminando..." : "Eliminar"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function ProveedorDialog({
  open,
  onClose,
  onSubmit,
  isPending,
  initialValues,
}: {
  open: boolean;
  onClose: () => void;
  onSubmit: (d: EntityBody) => void;
  isPending: boolean;
  initialValues?: Pick<Entity, "code" | "name" | "tax_id" | "email" | "phone" | "address" | "city">;
}) {
  const isEdit = !!initialValues;
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [taxId, setTaxId] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [address, setAddress] = useState("");
  const [city, setCity] = useState("");

  useEffect(() => {
    if (open) {
      setCode(initialValues?.code ?? "");
      setName(initialValues?.name ?? "");
      setTaxId(initialValues?.tax_id ?? "");
      setEmail(initialValues?.email ?? "");
      setPhone(initialValues?.phone ?? "");
      setAddress(initialValues?.address ?? "");
      setCity(initialValues?.city ?? "");
    }
  }, [open, initialValues]);

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Editar proveedor" : "Nuevo proveedor"}</DialogTitle>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (code.trim() && name.trim()) {
              onSubmit({
                code: code.trim(),
                name: name.trim(),
                tax_id: taxId || undefined,
                email: email || undefined,
                phone: phone || undefined,
                address: address || undefined,
                city: city || undefined,
              });
            }
          }}
          className="space-y-4"
        >
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2"><Label>Código</Label><Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Ej: PROV-001" required /></div>
            <div className="space-y-2"><Label>CUIT</Label><Input value={taxId} onChange={(e) => setTaxId(e.target.value)} placeholder="30-12345678-9" /></div>
          </div>
          <div className="space-y-2"><Label>Nombre / Razón social</Label><Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Razón social o nombre comercial" required /></div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2"><Label>Email</Label><Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} /></div>
            <div className="space-y-2"><Label>Teléfono</Label><Input value={phone} onChange={(e) => setPhone(e.target.value)} /></div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2"><Label>Dirección</Label><Input value={address} onChange={(e) => setAddress(e.target.value)} /></div>
            <div className="space-y-2"><Label>Ciudad</Label><Input value={city} onChange={(e) => setCity(e.target.value)} /></div>
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
            <Button type="submit" disabled={!code.trim() || !name.trim() || isPending}>
              {isPending ? (isEdit ? "Guardando..." : "Creando...") : (isEdit ? "Guardar cambios" : "Crear proveedor")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
