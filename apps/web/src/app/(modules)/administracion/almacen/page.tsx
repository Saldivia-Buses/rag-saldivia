"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PlusIcon, SearchIcon, PackageIcon, WarehouseIcon, ArrowRightLeftIcon } from "lucide-react";

interface Article {
  id: string; code: string; name: string; article_type: string;
  min_stock: number; avg_cost: number; active: boolean;
}
interface StockLevel {
  article_code: string; article_name: string; warehouse_code: string;
  warehouse_name: string; quantity: number; reserved: number;
}
interface StockMovement {
  id: string; article_code: string; article_name: string; movement_type: string;
  quantity: number; unit_cost: number; user_id: string; notes: string; created_at: string;
}

const fmt = (n: number) => new Intl.NumberFormat("es-AR", { maximumFractionDigits: 2 }).format(n);
const fmtMoney = (n: number) => n === 0 ? "—" : new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
const typeBadge: Record<string, string> = { in: "Ingreso", out: "Egreso", transfer: "Transferencia", adjustment: "Ajuste" };

export default function AlmacenPage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [levels, setLevels] = useState<StockLevel[]>([]);
  const [movements, setMovements] = useState<StockMovement[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  const fetchArticles = useCallback(async () => {
    try {
      const q = new URLSearchParams({ page_size: "100" });
      if (search) q.set("search", search);
      const data = await api.get<{ articles: Article[] }>(`/v1/erp/stock/articles?${q}`);
      setArticles(data.articles);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, [search]);

  const fetchLevels = useCallback(async () => {
    try {
      const data = await api.get<{ levels: StockLevel[] }>("/v1/erp/stock/levels");
      setLevels(data.levels);
    } catch (err) { console.error(err); }
  }, []);

  const fetchMovements = useCallback(async () => {
    try {
      const data = await api.get<{ movements: StockMovement[] }>("/v1/erp/stock/movements?page_size=50");
      setMovements(data.movements);
    } catch (err) { console.error(err); }
  }, []);

  useEffect(() => { fetchArticles(); fetchLevels(); fetchMovements(); }, [fetchArticles, fetchLevels, fetchMovements]);

  useEffect(() => {
    const handler = () => { fetchArticles(); fetchLevels(); fetchMovements(); };
    const unsub = wsManager.subscribe("erp_stock", handler);
    return unsub;
  }, [fetchArticles, fetchLevels, fetchMovements]);

  const handleCreate = async (code: string, name: string, type: string) => {
    await api.post("/v1/erp/stock/articles", { code, name, article_type: type || "material" });
    setCreateOpen(false);
    fetchArticles();
  };

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Almacen</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              Articulos, stock y movimientos — {articles.length} articulos
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo articulo
          </Button>
        </div>

        <Tabs defaultValue="articles">
          <TabsList className="mb-4">
            <TabsTrigger value="articles"><PackageIcon className="size-3.5 mr-1.5" />Articulos</TabsTrigger>
            <TabsTrigger value="stock"><WarehouseIcon className="size-3.5 mr-1.5" />Stock</TabsTrigger>
            <TabsTrigger value="movements"><ArrowRightLeftIcon className="size-3.5 mr-1.5" />Movimientos</TabsTrigger>
          </TabsList>

          <TabsContent value="articles">
            <div className="relative mb-4">
              <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
              <Input placeholder="Buscar por codigo o nombre..." className="pl-9 bg-card" value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-28">Codigo</TableHead>
                  <TableHead>Nombre</TableHead>
                  <TableHead className="w-28">Tipo</TableHead>
                  <TableHead className="w-28 text-right">Costo prom.</TableHead>
                  <TableHead className="w-20 text-center">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {articles.map((a) => (
                    <TableRow key={a.id}>
                      <TableCell className="font-mono text-sm">{a.code}</TableCell>
                      <TableCell className="text-sm">{a.name}</TableCell>
                      <TableCell><Badge variant="secondary">{a.article_type}</Badge></TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(a.avg_cost)}</TableCell>
                      <TableCell className="text-center"><Badge variant={a.active ? "default" : "secondary"}>{a.active ? "Activo" : "Inactivo"}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {articles.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin articulos.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="stock">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead>Articulo</TableHead>
                  <TableHead>Deposito</TableHead>
                  <TableHead className="text-right">Cantidad</TableHead>
                  <TableHead className="text-right">Reservado</TableHead>
                  <TableHead className="text-right">Disponible</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {levels.map((l, i) => (
                    <TableRow key={i}>
                      <TableCell><span className="font-mono text-sm">{l.article_code}</span> {l.article_name}</TableCell>
                      <TableCell className="text-sm">{l.warehouse_name}</TableCell>
                      <TableCell className="text-right font-mono">{fmt(l.quantity)}</TableCell>
                      <TableCell className="text-right font-mono text-muted-foreground">{fmt(l.reserved)}</TableCell>
                      <TableCell className="text-right font-mono font-medium">{fmt(l.quantity - l.reserved)}</TableCell>
                    </TableRow>
                  ))}
                  {levels.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin stock registrado.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="movements">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-36">Fecha</TableHead>
                  <TableHead>Articulo</TableHead>
                  <TableHead className="w-28">Tipo</TableHead>
                  <TableHead className="text-right w-24">Cantidad</TableHead>
                  <TableHead>Notas</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {movements.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell className="text-sm text-muted-foreground">{new Date(m.created_at).toLocaleDateString("es-AR", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" })}</TableCell>
                      <TableCell><span className="font-mono text-sm">{m.article_code}</span> {m.article_name}</TableCell>
                      <TableCell><Badge variant={m.movement_type === "in" ? "default" : "secondary"}>{typeBadge[m.movement_type] || m.movement_type}</Badge></TableCell>
                      <TableCell className={`text-right font-mono ${m.movement_type === "in" ? "text-green-600" : "text-red-500"}`}>{m.movement_type === "in" ? "+" : "-"}{fmt(m.quantity)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground truncate max-w-48">{m.notes || "—"}</TableCell>
                    </TableRow>
                  ))}
                  {movements.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin movimientos.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo articulo</DialogTitle></DialogHeader>
          <CreateArticleForm onCreate={handleCreate} onClose={() => setCreateOpen(false)} />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateArticleForm({ onCreate, onClose }: { onCreate: (code: string, name: string, type: string) => void; onClose: () => void }) {
  const [code, setCode] = useState(""); const [name, setName] = useState(""); const [type, setType] = useState("material");
  return (
    <form onSubmit={(e) => { e.preventDefault(); if (code && name) onCreate(code, name, type); }} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2"><Label>Codigo</Label><Input value={code} onChange={(e) => setCode(e.target.value)} /></div>
        <div className="space-y-2"><Label>Tipo</Label>
          <select className="w-full rounded-md border px-3 py-2 text-sm bg-card" value={type} onChange={(e) => setType(e.target.value)}>
            <option value="material">Material</option><option value="product">Producto</option>
            <option value="tool">Herramienta</option><option value="spare">Repuesto</option>
            <option value="consumable">Consumible</option>
          </select>
        </div>
      </div>
      <div className="space-y-2"><Label>Nombre</Label><Input value={name} onChange={(e) => setName(e.target.value)} /></div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!code || !name}>Crear</Button>
      </div>
    </form>
  );
}
