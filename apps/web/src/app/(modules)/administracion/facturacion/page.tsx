"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import type { Invoice, Withholding } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FileTextIcon, ShieldIcon, PlusIcon } from "lucide-react";

const typeLabel: Record<string, string> = {
  invoice_a: "Factura A", invoice_b: "Factura B", invoice_c: "Factura C",
  invoice_e: "Factura E", credit_note: "Nota Crédito", debit_note: "Nota Débito",
  delivery_note: "Remito",
};
const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  posted: { label: "Contabilizada", variant: "default" },
  paid: { label: "Cobrada", variant: "default" },
  cancelled: { label: "Anulada", variant: "secondary" },
};

interface InvoiceLine { description: string; quantity: string; unit_price: string; tax_rate: string; }
interface Entity { id: string; name: string; }

export default function FacturacionPage() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [voidOpen, setVoidOpen] = useState<string | null>(null);
  const [voidReason, setVoidReason] = useState("");

  const { data: invoices = [], isLoading, error } = useQuery({
    queryKey: erpKeys.invoices({ page_size: "50" }),
    queryFn: () => api.get<{ invoices: Invoice[] }>("/v1/erp/invoicing/invoices?page_size=50"),
    select: (d) => d.invoices,
  });

  const { data: withholdings = [] } = useQuery({
    queryKey: erpKeys.withholdings(),
    queryFn: () => api.get<{ withholdings: Withholding[] }>("/v1/erp/invoicing/withholdings?page_size=50"),
    select: (d) => d.withholdings,
  });

  const createMutation = useMutation({
    mutationFn: (data: {
      number: string; date: string; invoice_type: string; direction: string;
      entity_id: string; lines: InvoiceLine[];
    }) => api.post("/v1/erp/invoicing/invoices", data),
    onSuccess: () => {
      toast.success("Comprobante creado");
      queryClient.invalidateQueries({ queryKey: erpKeys.invoices() });
      setCreateOpen(false);
    },
    onError: permissionErrorToast,
  });

  const postMutation = useMutation({
    mutationFn: (id: string) => api.post(`/v1/erp/invoicing/invoices/${id}/post`),
    onSuccess: () => {
      toast.success("Comprobante contabilizado");
      queryClient.invalidateQueries({ queryKey: erpKeys.invoices() });
    },
    onError: permissionErrorToast,
  });

  const voidMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      api.post(`/v1/erp/invoicing/invoices/${id}/void`, { reason }),
    onSuccess: () => {
      toast.success("Comprobante anulado");
      queryClient.invalidateQueries({ queryKey: erpKeys.invoices() });
      setVoidOpen(null);
      setVoidReason("");
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando facturación" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Facturación</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Comprobantes, libro IVA y retenciones — {invoices.length} comprobantes</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}><PlusIcon className="size-4 mr-1.5" />Nueva factura</Button>
        </div>

        <Tabs defaultValue="invoices">
          <TabsList className="mb-4">
            <TabsTrigger value="invoices"><FileTextIcon className="size-3.5 mr-1.5" />Comprobantes</TabsTrigger>
            <TabsTrigger value="withholdings"><ShieldIcon className="size-3.5 mr-1.5" />Retenciones</TabsTrigger>
          </TabsList>

          <TabsContent value="invoices">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-36">Número</TableHead>
                  <TableHead className="w-28">Fecha</TableHead>
                  <TableHead className="w-28">Tipo</TableHead>
                  <TableHead>Entidad</TableHead>
                  <TableHead className="text-right w-28">Total</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                  <TableHead className="w-28" />
                </TableRow></TableHeader>
                <TableBody>
                  {invoices.map((inv) => {
                    const s = statusBadge[inv.status] || statusBadge.draft;
                    return (
                      <TableRow key={inv.id}>
                        <TableCell className="font-mono text-sm">
                          <Link href={`/administracion/facturacion/${inv.id}`} className="hover:underline">
                            {inv.number}
                          </Link>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{fmtDateShort(inv.date)}</TableCell>
                        <TableCell><Badge variant="secondary">{typeLabel[inv.invoice_type] || inv.invoice_type}</Badge></TableCell>
                        <TableCell className="text-sm">{inv.entity_name}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtMoney(inv.total)}</TableCell>
                        <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                        <TableCell>
                          {inv.status === "draft" && (
                            <Button size="sm" variant="outline" disabled={postMutation.isPending}
                              onClick={() => postMutation.mutate(inv.id)}>
                              Contabilizar
                            </Button>
                          )}
                          {inv.status === "posted" && (
                            <Button size="sm" variant="outline" onClick={() => setVoidOpen(inv.id)}>
                              Anular
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    );
                  })}
                  {invoices.length === 0 && (
                    <TableRow><TableCell colSpan={7} className="h-24 text-center text-muted-foreground">Sin comprobantes.</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="withholdings">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-28">Fecha</TableHead>
                  <TableHead>Entidad</TableHead>
                  <TableHead className="w-20">Tipo</TableHead>
                  <TableHead className="text-right w-20">Tasa</TableHead>
                  <TableHead className="text-right w-28">Base</TableHead>
                  <TableHead className="text-right w-28">Monto</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {withholdings.map((w) => (
                    <TableRow key={w.id}>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(w.date)}</TableCell>
                      <TableCell className="text-sm">{w.entity_name}</TableCell>
                      <TableCell><Badge variant="outline">{w.type.toUpperCase()}</Badge></TableCell>
                      <TableCell className="text-right font-mono text-sm">{w.rate}%</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(w.base_amount)}</TableCell>
                      <TableCell className="text-right font-mono text-sm font-medium">{fmtMoney(w.amount)}</TableCell>
                    </TableRow>
                  ))}
                  {withholdings.length === 0 && (
                    <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin retenciones.</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={(v) => !v && setCreateOpen(false)}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader><DialogTitle>Nueva factura</DialogTitle></DialogHeader>
          <CreateInvoiceForm
            onSubmit={(data) => createMutation.mutate(data)}
            isPending={createMutation.isPending}
            onClose={() => setCreateOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Void dialog */}
      <Dialog open={!!voidOpen} onOpenChange={(v) => !v && (setVoidOpen(null), setVoidReason(""))}>
        <DialogContent className="max-w-sm">
          <DialogHeader><DialogTitle>Anular comprobante</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Motivo de anulación</Label>
              <textarea
                className="w-full rounded-md border px-3 py-2 text-sm bg-card min-h-[80px] resize-none"
                value={voidReason}
                onChange={(e) => setVoidReason(e.target.value)}
                placeholder="Ingresá el motivo..."
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => { setVoidOpen(null); setVoidReason(""); }}>Cancelar</Button>
              <Button
                variant="destructive"
                disabled={!voidReason.trim() || voidMutation.isPending}
                onClick={() => voidOpen && voidMutation.mutate({ id: voidOpen, reason: voidReason })}
              >
                {voidMutation.isPending ? "Anulando..." : "Anular"}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateInvoiceForm({
  onSubmit, isPending, onClose,
}: {
  onSubmit: (data: { number: string; date: string; invoice_type: string; direction: string; entity_id: string; lines: InvoiceLine[] }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(new Date().toISOString().split("T")[0]);
  const [invoiceType, setInvoiceType] = useState("invoice_a");
  const [direction, setDirection] = useState("outgoing");
  const [entityId, setEntityId] = useState("");
  const [lines, setLines] = useState<InvoiceLine[]>([{ description: "", quantity: "1", unit_price: "", tax_rate: "21" }]);

  const { data: entities = [] } = useQuery({
    queryKey: erpKeys.entities(),
    queryFn: () => api.get<{ entities: Entity[] }>("/v1/erp/entities?page_size=200"),
    select: (d) => d.entities,
  });

  function addLine() {
    setLines((prev) => [...prev, { description: "", quantity: "1", unit_price: "", tax_rate: "21" }]);
  }

  function updateLine(i: number, field: keyof InvoiceLine, value: string) {
    setLines((prev) => prev.map((l, idx) => idx === i ? { ...l, [field]: value } : l));
  }

  function removeLine(i: number) {
    setLines((prev) => prev.filter((_, idx) => idx !== i));
  }

  const canSubmit = !!entityId && lines.length > 0 && lines.every((l) => l.description && l.unit_price);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (canSubmit) onSubmit({ number, date, invoice_type: invoiceType, direction, entity_id: entityId, lines });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="0001-00000001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Tipo de comprobante</Label>
          <Select value={invoiceType} onValueChange={(v) => v && setInvoiceType(v)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="invoice_a">Factura A</SelectItem>
              <SelectItem value="invoice_b">Factura B</SelectItem>
              <SelectItem value="invoice_c">Factura C</SelectItem>
              <SelectItem value="invoice_e">Factura E</SelectItem>
              <SelectItem value="credit_note">Nota Crédito</SelectItem>
              <SelectItem value="debit_note">Nota Débito</SelectItem>
              <SelectItem value="delivery_note">Remito</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Dirección</Label>
          <Select value={direction} onValueChange={(v) => v && setDirection(v)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="outgoing">Emisión</SelectItem>
              <SelectItem value="incoming">Recepción</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <Label>Entidad</Label>
        <Select value={entityId} onValueChange={(v) => v && setEntityId(v)}>
          <SelectTrigger><SelectValue placeholder="Seleccionar entidad..." /></SelectTrigger>
          <SelectContent>
            {(entities as Entity[]).map((e) => (
              <SelectItem key={e.id} value={e.id}>{e.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label>Líneas</Label>
          <Button type="button" size="sm" variant="outline" onClick={addLine}><PlusIcon className="size-3.5 mr-1" />Agregar línea</Button>
        </div>
        <div className="space-y-2">
          {lines.map((line, i) => (
            <div key={i} className="grid grid-cols-12 gap-2 items-start">
              <div className="col-span-5">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Descripción</p>}
                <Input value={line.description} onChange={(e) => updateLine(i, "description", e.target.value)} placeholder="Descripción" />
              </div>
              <div className="col-span-2">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Cantidad</p>}
                <Input type="number" value={line.quantity} onChange={(e) => updateLine(i, "quantity", e.target.value)} placeholder="1" />
              </div>
              <div className="col-span-2">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">Precio unit.</p>}
                <Input type="number" value={line.unit_price} onChange={(e) => updateLine(i, "unit_price", e.target.value)} placeholder="0.00" />
              </div>
              <div className="col-span-2">
                {i === 0 && <p className="text-xs text-muted-foreground mb-1">IVA %</p>}
                <Input type="number" value={line.tax_rate} onChange={(e) => updateLine(i, "tax_rate", e.target.value)} placeholder="21" />
              </div>
              <div className="col-span-1 flex items-end">
                {i === 0 && <div className="h-[18px]" />}
                <Button type="button" size="sm" variant="ghost" disabled={lines.length === 1}
                  onClick={() => removeLine(i)} className="px-2 text-muted-foreground hover:text-destructive">×</Button>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
