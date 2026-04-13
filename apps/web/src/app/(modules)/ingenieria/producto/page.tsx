"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { EmptyState } from "@/components/erp/empty-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, SearchIcon, BoxIcon } from "lucide-react";

interface Article {
  id: string;
  code: string;
  name: string;
  article_type: string;
  min_stock: number;
  avg_cost: number;
  active: boolean;
  notes?: string;
}

export default function ProductoPage() {
  const [search, setSearch] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: articles = [], isLoading, error } = useQuery({
    queryKey: erpKeys.stockArticles({ article_type: "product" }),
    queryFn: () =>
      api.get<{ articles: Article[] }>(
        "/v1/erp/stock/articles?article_type=product&page_size=100"
      ),
    select: (d) => d.articles,
  });

  const createMutation = useMutation({
    mutationFn: (data: { code: string; name: string; article_type: string; min_stock: number; notes?: string }) =>
      api.post("/v1/erp/stock/articles", data),
    onSuccess: () => {
      toast.success("Producto creado exitosamente");
      queryClient.invalidateQueries({ queryKey: erpKeys.stockArticles() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const filtered = articles.filter(
    (a) =>
      a.name.toLowerCase().includes(search.toLowerCase()) ||
      a.code.toLowerCase().includes(search.toLowerCase())
  );

  if (error) return <ErrorState message="Error cargando productos" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Productos / Especificaciones de vehículos</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Modelos y especificaciones técnicas de vehículos — {articles.length} productos
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo producto
          </Button>
        </div>

        <div className="relative mb-4">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Buscar por código o nombre..."
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
                <TableHead>Nombre</TableHead>
                <TableHead className="w-36 text-right">Costo prom.</TableHead>
                <TableHead className="w-24 text-center">Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="font-mono text-sm">{a.code}</TableCell>
                  <TableCell className="text-sm">{a.name}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{fmtMoney(a.avg_cost)}</TableCell>
                  <TableCell className="text-center">
                    <Badge variant={a.active ? "default" : "secondary"}>
                      {a.active ? "Activo" : "Inactivo"}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4}>
                    <EmptyState
                      icon={BoxIcon}
                      title={search ? "Sin resultados" : "Sin productos"}
                      description={search ? "No se encontraron productos con ese criterio." : "Creá el primer producto para empezar."}
                      action={!search ? { label: "Nuevo producto", onClick: () => setCreateOpen(true) } : undefined}
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
          <DialogHeader><DialogTitle>Nuevo producto</DialogTitle></DialogHeader>
          <CreateProductForm
            onSubmit={(d) => createMutation.mutate({ ...d, article_type: "product", min_stock: 0 })}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateProductForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (d: { code: string; name: string; notes?: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [notes, setNotes] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (code.trim() && name.trim()) {
          const d: { code: string; name: string; notes?: string } = { code: code.trim(), name: name.trim() };
          if (notes.trim()) d.notes = notes.trim();
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
          placeholder="Ej: BUS-210, MICRO-450"
        />
      </div>
      <div className="space-y-2">
        <Label>Nombre</Label>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Ej: Colectivo urbano 12m — especificación 2025"
        />
      </div>
      <div className="space-y-2">
        <Label>Descripción <span className="text-muted-foreground">(opcional)</span></Label>
        <Input
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Notas adicionales..."
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!code.trim() || !name.trim() || isPending}>
          {isPending ? "Creando..." : "Crear"}
        </Button>
      </div>
    </form>
  );
}
