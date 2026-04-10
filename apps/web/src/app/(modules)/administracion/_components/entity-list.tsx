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
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { PlusIcon, SearchIcon, UserIcon, BuildingIcon, MailIcon, PhoneIcon } from "lucide-react";

// ─── Types ──────────────────────────────────────────────────────────────────

interface Entity {
  id: string;
  type: string;
  code: string;
  name: string;
  tax_id_hash: string | null;
  email: string | null;
  phone: string | null;
  address: Record<string, unknown>;
  metadata: Record<string, unknown>;
  active: boolean;
  created_at: string;
}

interface EntityContact {
  id: string;
  type: string;
  label: string;
  value: string;
}

interface EntityNote {
  id: string;
  user_id: string;
  type: string;
  body: string;
  created_at: string;
}

interface EntityDetail {
  entity: Entity;
  contacts: EntityContact[];
  documents: unknown[];
  notes: EntityNote[];
  relations: unknown[];
}

// ─── Props ──────────────────────────────────────────────────────────────────

interface EntityListProps {
  entityType: "employee" | "customer" | "supplier";
  title: string;
  subtitle: string;
  codeLabel: string; // "Legajo", "CUIT", "Código"
}

// ─── Component ──────────────────────────────────────────────────────────────

export function EntityList({ entityType, title, subtitle, codeLabel }: EntityListProps) {
  const [entities, setEntities] = useState<Entity[]>([]);
  const [total, setTotal] = useState(0);
  const [selected, setSelected] = useState<EntityDetail | null>(null);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  const fetchEntities = useCallback(async () => {
    try {
      const q = new URLSearchParams({ type: entityType, page_size: "100" });
      if (search) q.set("search", search);
      const data = await api.get<{ entities: Entity[]; total: number }>(`/v1/erp/entities?${q}`);
      setEntities(data.entities);
      setTotal(data.total);
    } catch (err) {
      console.error("Failed to fetch entities:", err);
    } finally {
      setLoading(false);
    }
  }, [entityType, search]);

  const fetchDetail = useCallback(async (id: string) => {
    try {
      const data = await api.get<EntityDetail>(`/v1/erp/entities/${id}`);
      setSelected(data);
    } catch (err) {
      console.error("Failed to fetch entity detail:", err);
    }
  }, []);

  useEffect(() => { fetchEntities(); }, [fetchEntities]);

  useEffect(() => {
    const handler = () => {
      fetchEntities();
      if (selected) fetchDetail(selected.entity.id);
    };
    const unsubscribe = wsManager.subscribe("erp_entities", handler);
    return unsubscribe;
  }, [selected, fetchEntities, fetchDetail]);

  const handleCreate = async (code: string, name: string, email?: string, phone?: string, taxId?: string) => {
    await api.post("/v1/erp/entities", {
      type: entityType, code, name,
      ...(email ? { email } : {}),
      ...(phone ? { phone } : {}),
      ...(taxId ? { tax_id: taxId } : {}),
    });
    setCreateOpen(false);
    fetchEntities();
  };

  const fmtDate = (s: string) => new Date(s).toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric" });

  if (loading) {
    return (
      <div className="flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-6xl px-6 py-8">
          <Skeleton className="h-8 w-48 mb-6" />
          <div className="flex gap-6">
            <Skeleton className="h-[600px] flex-1" />
            <Skeleton className="h-[600px] w-80" />
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
            <h1 className="text-xl font-semibold tracking-tight">{title}</h1>
            <p className="text-sm text-muted-foreground mt-0.5">{subtitle} — {total} registros</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />
            Nuevo
          </Button>
        </div>

        <div className="flex gap-6">
          {/* List */}
          <div className="flex-1 min-w-0">
            <div className="relative mb-4">
              <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
              <Input
                placeholder={`Buscar por nombre o ${codeLabel.toLowerCase()}...`}
                className="pl-9 bg-card"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>

            <ScrollArea className="h-[calc(100vh-14rem)]">
              <div className="space-y-1">
                {entities.map((e) => (
                  <button
                    key={e.id}
                    onClick={() => fetchDetail(e.id)}
                    className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
                      selected?.entity.id === e.id
                        ? "bg-primary/10 border border-primary/20"
                        : "hover:bg-muted/50 border border-transparent"
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      <div className="size-9 rounded-full bg-muted flex items-center justify-center shrink-0">
                        {entityType === "employee" ? (
                          <UserIcon className="size-4 text-muted-foreground" />
                        ) : (
                          <BuildingIcon className="size-4 text-muted-foreground" />
                        )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium truncate">{e.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {codeLabel}: {e.code}
                          {e.email && ` · ${e.email}`}
                        </p>
                      </div>
                      {!e.active && <Badge variant="secondary">Inactivo</Badge>}
                    </div>
                  </button>
                ))}
                {entities.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    {search ? "Sin resultados." : "Sin registros todavía."}
                  </p>
                )}
              </div>
            </ScrollArea>
          </div>

          {/* Detail panel */}
          <div className="w-80 shrink-0">
            {selected ? (
              <div className="rounded-xl border border-border/40 bg-card p-5">
                <div className="flex items-center gap-3 mb-4">
                  <div className="size-11 rounded-full bg-primary/10 flex items-center justify-center">
                    {entityType === "employee" ? (
                      <UserIcon className="size-5 text-primary" />
                    ) : (
                      <BuildingIcon className="size-5 text-primary" />
                    )}
                  </div>
                  <div>
                    <h2 className="font-semibold">{selected.entity.name}</h2>
                    <p className="text-xs text-muted-foreground">{codeLabel}: {selected.entity.code}</p>
                  </div>
                </div>

                <Separator className="mb-4" />

                {/* Contact info */}
                <div className="space-y-2 mb-4">
                  {selected.entity.email && (
                    <div className="flex items-center gap-2 text-sm">
                      <MailIcon className="size-3.5 text-muted-foreground" />
                      {selected.entity.email}
                    </div>
                  )}
                  {selected.entity.phone && (
                    <div className="flex items-center gap-2 text-sm">
                      <PhoneIcon className="size-3.5 text-muted-foreground" />
                      {selected.entity.phone}
                    </div>
                  )}
                </div>

                {/* Extra contacts */}
                {selected.contacts.length > 0 && (
                  <>
                    <p className="text-xs font-medium text-muted-foreground mb-2">Contactos</p>
                    <div className="space-y-1 mb-4">
                      {selected.contacts.map((c) => (
                        <div key={c.id} className="text-sm">
                          <span className="text-muted-foreground">{c.label || c.type}:</span> {c.value}
                        </div>
                      ))}
                    </div>
                  </>
                )}

                {/* Notes */}
                {selected.notes.length > 0 && (
                  <>
                    <p className="text-xs font-medium text-muted-foreground mb-2">Notas recientes</p>
                    <div className="space-y-2">
                      {selected.notes.slice(0, 5).map((n) => (
                        <div key={n.id} className="text-sm bg-muted/30 rounded-lg p-2.5">
                          <p className="text-xs text-muted-foreground mb-1">{fmtDate(n.created_at)}</p>
                          <p>{n.body}</p>
                        </div>
                      ))}
                    </div>
                  </>
                )}

                <Separator className="my-4" />
                <p className="text-xs text-muted-foreground">
                  Creado {fmtDate(selected.entity.created_at)}
                </p>
              </div>
            ) : (
              <div className="rounded-xl border border-border/40 bg-card p-5 flex items-center justify-center h-64 text-sm text-muted-foreground">
                Seleccioná un registro para ver el detalle.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create dialog */}
      <CreateEntityDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreate={handleCreate}
        codeLabel={codeLabel}
        entityType={entityType}
      />
    </div>
  );
}

// ─── Create Dialog ──────────────────────────────────────────────────────────

function CreateEntityDialog({
  open,
  onClose,
  onCreate,
  codeLabel,
  entityType,
}: {
  open: boolean;
  onClose: () => void;
  onCreate: (code: string, name: string, email?: string, phone?: string, taxId?: string) => void;
  codeLabel: string;
  entityType: string;
}) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [taxId, setTaxId] = useState("");

  useEffect(() => {
    if (open) { setCode(""); setName(""); setEmail(""); setPhone(""); setTaxId(""); }
  }, [open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim() || !name.trim()) return;
    onCreate(code.trim(), name.trim(), email || undefined, phone || undefined, taxId || undefined);
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {entityType === "employee" ? "Nuevo empleado" : entityType === "customer" ? "Nuevo cliente" : "Nuevo proveedor"}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{codeLabel}</Label>
              <Input value={code} onChange={(e) => setCode(e.target.value)} placeholder="Ej: 1234" />
            </div>
            <div className="space-y-2">
              <Label>CUIT/CUIL</Label>
              <Input value={taxId} onChange={(e) => setTaxId(e.target.value)} placeholder="20-12345678-9" />
            </div>
          </div>
          <div className="space-y-2">
            <Label>Nombre</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Nombre completo o razón social" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Email</Label>
              <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>Teléfono</Label>
              <Input value={phone} onChange={(e) => setPhone(e.target.value)} />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
            <Button type="submit" disabled={!code.trim() || !name.trim()}>Crear</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
