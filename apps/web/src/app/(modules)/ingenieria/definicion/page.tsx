"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { EmptyState } from "@/components/erp/empty-state";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PlusIcon, PackageIcon, LayersIcon } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface CarroceriaModel {
  id: string;
  code: string;
  model_code: string;
  description: string;
  abbreviation: string;
  double_deck: boolean;
  active: boolean;
}

interface BomItem {
  id: string;
  article_id: string;
  article_code: string;
  article_name: string;
  quantity: number;
  unit_of_use: string;
}

interface StockArticle {
  id: string;
  code: string;
  name: string;
}

// ---------------------------------------------------------------------------
// Query keys
// ---------------------------------------------------------------------------

const manufacturingKeys = {
  carroceriaModels: () => [...erpKeys.all, "manufacturing", "carroceria-models"] as const,
  bom: (modelId: string) => [...erpKeys.all, "manufacturing", "bom", modelId] as const,
};

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

export default function DefinicionPage() {
  const [selectedModelId, setSelectedModelId] = useState<string | null>(null);
  const [addItemOpen, setAddItemOpen] = useState(false);
  const queryClient = useQueryClient();

  // Left panel: carrocería models
  const {
    data: models = [],
    isLoading: modelsLoading,
    error: modelsError,
  } = useQuery({
    queryKey: manufacturingKeys.carroceriaModels(),
    queryFn: () =>
      api.get<{ carroceria_models: CarroceriaModel[] }>(
        "/v1/erp/manufacturing/carroceria-models"
      ),
    select: (d) => d.carroceria_models,
  });

  // Right panel: BOM for selected model
  const {
    data: bom = [],
    isLoading: bomLoading,
    error: bomError,
  } = useQuery({
    queryKey: manufacturingKeys.bom(selectedModelId ?? ""),
    queryFn: () =>
      api.get<{ bom: BomItem[] }>(
        `/v1/erp/manufacturing/carroceria-models/${selectedModelId}/bom`
      ),
    select: (d) => d.bom,
    enabled: !!selectedModelId,
  });

  // Articles for the add-item dialog
  const { data: articles = [] } = useQuery({
    queryKey: erpKeys.stockArticles({ page_size: "200" }),
    queryFn: () =>
      api.get<{ articles: StockArticle[] }>("/v1/erp/stock/articles?page_size=200"),
    select: (d) => d.articles,
    enabled: addItemOpen,
  });

  const addBomItemMutation = useMutation({
    mutationFn: (data: { article_id: string; quantity: number; unit_of_use?: string }) =>
      api.post(
        `/v1/erp/manufacturing/carroceria-models/${selectedModelId}/bom`,
        data
      ),
    onSuccess: () => {
      toast.success("Ítem agregado al BOM");
      queryClient.invalidateQueries({
        queryKey: manufacturingKeys.bom(selectedModelId ?? ""),
      });
      setAddItemOpen(false);
    },
    onError: permissionErrorToast,
  });

  const selectedModel = models.find((m) => m.id === selectedModelId);

  if (modelsError)
    return (
      <ErrorState
        message="Error cargando modelos de carrocería"
        onRetry={() => window.location.reload()}
      />
    );

  return (
    <div className="flex-1 overflow-hidden flex flex-col">
      <div className="mx-auto w-full max-w-7xl px-6 py-8 sm:px-8 flex-1 flex flex-col">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Definición técnica — BOM</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Seleccioná un modelo de carrocería para ver y editar su lista de materiales
          </p>
        </div>

        <div className="flex gap-6 flex-1 overflow-hidden min-h-0">
          {/* ── Left panel: model list ── */}
          <div className="w-72 shrink-0 flex flex-col gap-2">
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider px-1">
              Modelos
            </p>
            <div className="rounded-xl border border-border/40 bg-card overflow-y-auto flex-1">
              {modelsLoading ? (
                <div className="p-4 space-y-2">
                  {[...Array(5)].map((_, i) => (
                    <Skeleton key={i} className="h-10 rounded-lg" />
                  ))}
                </div>
              ) : models.length === 0 ? (
                <div className="p-4 text-sm text-muted-foreground text-center">
                  Sin modelos registrados
                </div>
              ) : (
                <ul className="py-2">
                  {models.map((m) => (
                    <li key={m.id}>
                      <button
                        type="button"
                        onClick={() => setSelectedModelId(m.id)}
                        className={[
                          "w-full text-left px-4 py-2.5 transition-colors text-sm",
                          "hover:bg-accent/60",
                          selectedModelId === m.id
                            ? "bg-accent text-accent-foreground font-medium"
                            : "text-foreground",
                        ].join(" ")}
                      >
                        <div className="font-mono text-xs text-muted-foreground">
                          {m.code}
                        </div>
                        <div className="mt-0.5 truncate">{m.description}</div>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>

          {/* ── Right panel: BOM ── */}
          <div className="flex-1 flex flex-col min-w-0">
            {!selectedModelId ? (
              <div className="flex-1 flex items-center justify-center rounded-xl border border-border/40 bg-card">
                <EmptyState
                  icon={LayersIcon}
                  title="Seleccioná un modelo"
                  description="Elegí un modelo de carrocería de la lista para ver su BOM."
                />
              </div>
            ) : (
              <>
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <p className="text-sm font-medium">{selectedModel?.description}</p>
                    <p className="text-xs text-muted-foreground">
                      {bom.length} {bom.length === 1 ? "ítem" : "ítems"} en el BOM
                    </p>
                  </div>
                  <Button size="sm" onClick={() => setAddItemOpen(true)}>
                    <PlusIcon className="size-4 mr-1.5" />
                    Agregar ítem
                  </Button>
                </div>

                {bomError ? (
                  <ErrorState
                    message="Error cargando BOM"
                    onRetry={() =>
                      queryClient.invalidateQueries({
                        queryKey: manufacturingKeys.bom(selectedModelId),
                      })
                    }
                  />
                ) : (
                  <div className="rounded-xl border border-border/40 bg-card overflow-hidden flex-1">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-32">Código</TableHead>
                          <TableHead>Artículo</TableHead>
                          <TableHead className="w-28 text-right">Cantidad</TableHead>
                          <TableHead className="w-28">Unidad</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {bomLoading
                          ? [...Array(4)].map((_, i) => (
                              <TableRow key={i}>
                                <TableCell colSpan={4}>
                                  <Skeleton className="h-5" />
                                </TableCell>
                              </TableRow>
                            ))
                          : bom.map((item) => (
                              <TableRow key={item.id}>
                                <TableCell className="font-mono text-sm">
                                  {item.article_code}
                                </TableCell>
                                <TableCell className="text-sm">{item.article_name}</TableCell>
                                <TableCell className="text-sm text-right tabular-nums">
                                  {item.quantity}
                                </TableCell>
                                <TableCell className="text-sm text-muted-foreground">
                                  {item.unit_of_use || "—"}
                                </TableCell>
                              </TableRow>
                            ))}
                        {!bomLoading && bom.length === 0 && (
                          <TableRow>
                            <TableCell colSpan={4}>
                              <EmptyState
                                icon={PackageIcon}
                                title="BOM vacío"
                                description="Este modelo no tiene ítems. Agregá el primero."
                                action={{
                                  label: "Agregar ítem",
                                  onClick: () => setAddItemOpen(true),
                                }}
                              />
                            </TableCell>
                          </TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>

      {/* ── Add BOM item dialog ── */}
      <Dialog open={addItemOpen} onOpenChange={(v) => !v && setAddItemOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Agregar ítem al BOM</DialogTitle>
          </DialogHeader>
          <AddBomItemForm
            articles={articles}
            onSubmit={(d) => addBomItemMutation.mutate(d)}
            isPending={addBomItemMutation.isPending}
            onClose={() => setAddItemOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Add BOM item form
// ---------------------------------------------------------------------------

function AddBomItemForm({
  articles,
  onSubmit,
  isPending,
  onClose,
}: {
  articles: StockArticle[];
  onSubmit: (d: { article_id: string; quantity: number; unit_of_use?: string }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [articleId, setArticleId] = useState("");
  const [quantity, setQuantity] = useState("");
  const [unitOfUse, setUnitOfUse] = useState("");

  const quantityNum = parseFloat(quantity);
  const valid = articleId && quantity && !isNaN(quantityNum) && quantityNum > 0;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!valid) return;
        const d: { article_id: string; quantity: number; unit_of_use?: string } = {
          article_id: articleId,
          quantity: quantityNum,
        };
        if (unitOfUse.trim()) d.unit_of_use = unitOfUse.trim();
        onSubmit(d);
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Artículo</Label>
        <select
          value={articleId}
          onChange={(e) => setArticleId(e.target.value)}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
        >
          <option value="">Seleccionar artículo...</option>
          {articles.map((a) => (
            <option key={a.id} value={a.id}>
              {a.code} — {a.name}
            </option>
          ))}
        </select>
      </div>
      <div className="space-y-2">
        <Label>Cantidad</Label>
        <Input
          type="number"
          min="0.001"
          step="any"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          placeholder="Ej: 4"
        />
      </div>
      <div className="space-y-2">
        <Label>
          Unidad de uso <span className="text-muted-foreground">(opcional)</span>
        </Label>
        <Input
          value={unitOfUse}
          onChange={(e) => setUnitOfUse(e.target.value)}
          placeholder="Ej: unidad, m², kg"
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>
          Cancelar
        </Button>
        <Button type="submit" disabled={!valid || isPending}>
          {isPending ? "Agregando..." : "Agregar"}
        </Button>
      </div>
    </form>
  );
}
