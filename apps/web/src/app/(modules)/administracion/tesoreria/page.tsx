"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import type { TreasuryMovement, Check, BankBalance, Reconciliation, Receipt } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { BanknoteIcon, CreditCardIcon, LandmarkIcon, ScaleIcon, ReceiptIcon } from "lucide-react";

const moveLabel: Record<string, string> = {
  cash_in: "Ingreso caja", cash_out: "Egreso caja", bank_deposit: "Depósito",
  bank_withdrawal: "Retiro", check_issued: "Cheque emitido",
  check_received: "Cheque recibido", transfer: "Transferencia",
};
const checkStatus: Record<string, string> = {
  in_portfolio: "En cartera", deposited: "Depositado", cashed: "Cobrado",
  rejected: "Rechazado", endorsed: "Endosado",
};

export default function TesoreriaPage() {
  const { data: movements = [], isLoading, error } = useQuery({
    queryKey: erpKeys.treasuryMovements(),
    queryFn: () => api.get<{ movements: TreasuryMovement[] }>("/v1/erp/treasury/movements?page_size=50"),
    select: (d) => d.movements,
  });

  const { data: checks = [] } = useQuery({
    queryKey: erpKeys.checks(),
    queryFn: () => api.get<{ checks: Check[] }>("/v1/erp/treasury/checks"),
    select: (d) => d.checks,
  });

  const { data: balances = [] } = useQuery({
    queryKey: erpKeys.treasuryBalance(),
    queryFn: () => api.get<{ balances: BankBalance[] }>("/v1/erp/treasury/balance"),
    select: (d) => d.balances,
  });

  const { data: reconciliations = [] } = useQuery({
    queryKey: [...erpKeys.all, "treasury", "reconciliations"] as const,
    queryFn: () => api.get<{ reconciliations: Reconciliation[] }>("/v1/erp/treasury/reconciliations"),
    select: (d) => d.reconciliations,
  });

  const { data: receipts = [] } = useQuery({
    queryKey: erpKeys.receipts(),
    queryFn: () => api.get<{ receipts: Receipt[] }>("/v1/erp/treasury/receipts?page_size=50"),
    select: (d) => d.receipts,
  });

  if (error) return <ErrorState message="Error cargando tesorería" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const totalBalance = balances.reduce((a, b) => a + (b.balance || 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Tesorería</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Movimientos, cheques y saldos bancarios</p>
        </div>

        <div className="grid grid-cols-3 gap-3 mb-6">
          {balances.slice(0, 3).map((b) => (
            <div key={b.account_number} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground mb-1">{b.bank_name}</p>
              <p className={`text-xl font-semibold ${b.balance >= 0 ? "text-green-500" : "text-red-500"}`}>{fmtMoney(b.balance)}</p>
            </div>
          ))}
          {balances.length === 0 && (
            <div className="col-span-3 rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground">Total disponible</p>
              <p className="text-xl font-semibold">{fmtMoney(totalBalance)}</p>
            </div>
          )}
        </div>

        <Tabs defaultValue="movements">
          <TabsList className="mb-4">
            <TabsTrigger value="movements"><BanknoteIcon className="size-3.5 mr-1.5" />Movimientos</TabsTrigger>
            <TabsTrigger value="checks"><CreditCardIcon className="size-3.5 mr-1.5" />Cheques</TabsTrigger>
            <TabsTrigger value="banks"><LandmarkIcon className="size-3.5 mr-1.5" />Bancos</TabsTrigger>
            <TabsTrigger value="reconciliation"><ScaleIcon className="size-3.5 mr-1.5" />Reconciliación</TabsTrigger>
            <TabsTrigger value="receipts"><ReceiptIcon className="size-3.5 mr-1.5" />Recibos</TabsTrigger>
          </TabsList>

          <TabsContent value="movements">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-28">Fecha</TableHead>
                  <TableHead className="w-20">Nro</TableHead>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Entidad</TableHead>
                  <TableHead className="text-right">Monto</TableHead>
                  <TableHead className="w-24">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {movements.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(m.date)}</TableCell>
                      <TableCell className="font-mono text-sm">{m.number}</TableCell>
                      <TableCell className="text-sm">{moveLabel[m.movement_type] || m.movement_type}</TableCell>
                      <TableCell className="text-sm">{m.entity_name || "\u2014"}</TableCell>
                      <TableCell className={`text-right font-mono text-sm ${m.movement_type.includes("in") || m.movement_type.includes("deposit") || m.movement_type.includes("received") ? "text-green-600" : "text-red-500"}`}>{fmtMoney(m.amount)}</TableCell>
                      <TableCell><Badge variant={m.status === "confirmed" ? "default" : "secondary"}>{m.status}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {movements.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin movimientos.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="checks">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-24">Nro</TableHead>
                  <TableHead>Banco</TableHead>
                  <TableHead className="w-24">Tipo</TableHead>
                  <TableHead className="text-right">Monto</TableHead>
                  <TableHead className="w-28">Vencimiento</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {checks.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="font-mono text-sm">{c.number}</TableCell>
                      <TableCell className="text-sm">{c.bank_name}</TableCell>
                      <TableCell><Badge variant="secondary">{c.direction === "received" ? "Recibido" : "Emitido"}</Badge></TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(c.amount)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(c.due_date)}</TableCell>
                      <TableCell><Badge variant="outline">{checkStatus[c.status] || c.status}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {checks.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin cheques.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="banks">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead>Banco</TableHead>
                  <TableHead>Cuenta</TableHead>
                  <TableHead className="text-right">Ingresos</TableHead>
                  <TableHead className="text-right">Egresos</TableHead>
                  <TableHead className="text-right">Saldo</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {balances.map((b, i) => (
                    <TableRow key={i}>
                      <TableCell className="text-sm font-medium">{b.bank_name}</TableCell>
                      <TableCell className="font-mono text-sm">{b.account_number}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-green-600">{fmtMoney(b.total_in)}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-red-500">{fmtMoney(b.total_out)}</TableCell>
                      <TableCell className={`text-right font-mono text-sm font-medium ${b.balance >= 0 ? "" : "text-red-500"}`}>{fmtMoney(b.balance)}</TableCell>
                    </TableRow>
                  ))}
                  {balances.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin cuentas bancarias.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="reconciliation">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead>Banco</TableHead><TableHead className="w-28">Período</TableHead>
                  <TableHead className="text-right">Saldo extracto</TableHead><TableHead className="text-right">Saldo libros</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {reconciliations.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="text-sm">{r.bank_name}</TableCell>
                      <TableCell className="font-mono text-sm">{r.period}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(r.statement_balance)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(r.book_balance)}</TableCell>
                      <TableCell><Badge variant={r.status === "confirmed" ? "default" : "secondary"}>{r.status === "confirmed" ? "Confirmada" : "Borrador"}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {reconciliations.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin reconciliaciones.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="receipts">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-28">Número</TableHead><TableHead className="w-28">Fecha</TableHead>
                  <TableHead className="w-24">Tipo</TableHead><TableHead>Entidad</TableHead>
                  <TableHead className="text-right w-28">Total</TableHead><TableHead className="w-28">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {receipts.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-sm">{r.number}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fmtDateShort(r.date)}</TableCell>
                      <TableCell><Badge variant="secondary">{r.receipt_type === "collection" ? "Cobro" : "Pago"}</Badge></TableCell>
                      <TableCell className="text-sm">{r.entity_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(r.total)}</TableCell>
                      <TableCell><Badge variant={r.status === "confirmed" ? "default" : "secondary"}>{r.status === "confirmed" ? "Confirmado" : r.status}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {receipts.length === 0 && <TableRow><TableCell colSpan={6} className="h-24 text-center text-muted-foreground">Sin recibos.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
