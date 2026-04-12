"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import type { Invoice, Withholding } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FileTextIcon, ShieldIcon } from "lucide-react";

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

export default function FacturacionPage() {
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

  if (error) return <ErrorState message="Error cargando facturación" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Facturación</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Comprobantes, libro IVA y retenciones — {invoices.length} comprobantes</p>
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
                </TableRow></TableHeader>
                <TableBody>
                  {invoices.map((inv) => {
                    const s = statusBadge[inv.status] || statusBadge.draft;
                    return (
                      <TableRow key={inv.id}>
                        <TableCell className="font-mono text-sm">{inv.number}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{fmtDateShort(inv.date)}</TableCell>
                        <TableCell><Badge variant="secondary">{typeLabel[inv.invoice_type] || inv.invoice_type}</Badge></TableCell>
                        <TableCell className="text-sm">{inv.entity_name}</TableCell>
                        <TableCell className="text-right font-mono text-sm">{fmtMoney(inv.total)}</TableCell>
                        <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                      </TableRow>
                    );
                  })}
                  {invoices.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin comprobantes.</TableCell></TableRow>}
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
                  {withholdings.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin retenciones.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
