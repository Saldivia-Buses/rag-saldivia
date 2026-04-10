"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { BanknoteIcon, CreditCardIcon, LandmarkIcon, CalculatorIcon } from "lucide-react";

interface Movement { id: string; date: string; number: string; movement_type: string; amount: number; entity_name: string | null; notes: string; status: string; }
interface Check { id: string; direction: string; number: string; bank_name: string; amount: number; issue_date: string; due_date: string; status: string; }
interface BankBalance { bank_name: string; account_number: string; total_in: number; total_out: number; balance: number; }

const fmtMoney = (n: number) => n === 0 ? "—" : new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
const fmtDate = (s: string) => new Date(s).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });
const moveLabel: Record<string, string> = { cash_in: "Ingreso caja", cash_out: "Egreso caja", bank_deposit: "Deposito", bank_withdrawal: "Retiro", check_issued: "Cheque emitido", check_received: "Cheque recibido", transfer: "Transferencia" };
const checkStatus: Record<string, string> = { in_portfolio: "En cartera", deposited: "Depositado", cashed: "Cobrado", rejected: "Rechazado", endorsed: "Endosado" };

export default function TesoreriaPage() {
  const [movements, setMovements] = useState<Movement[]>([]);
  const [checks, setChecks] = useState<Check[]>([]);
  const [balances, setBalances] = useState<BankBalance[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const [m, c, b] = await Promise.all([
        api.get<{ movements: Movement[] }>("/v1/erp/treasury/movements?page_size=50"),
        api.get<{ checks: Check[] }>("/v1/erp/treasury/checks"),
        api.get<{ balances: BankBalance[] }>("/v1/erp/treasury/balance"),
      ]);
      setMovements(m.movements); setChecks(c.checks); setBalances(b.balances);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);
  useEffect(() => { const unsub = wsManager.subscribe("erp_treasury", fetch); return unsub; }, [fetch]);

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const totalBalance = balances.reduce((a, b) => a + (b.balance || 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Tesoreria</h1>
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
                      <TableCell className="text-sm text-muted-foreground">{fmtDate(m.date)}</TableCell>
                      <TableCell className="font-mono text-sm">{m.number}</TableCell>
                      <TableCell className="text-sm">{moveLabel[m.movement_type] || m.movement_type}</TableCell>
                      <TableCell className="text-sm">{m.entity_name || "—"}</TableCell>
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
                      <TableCell className="text-sm text-muted-foreground">{fmtDate(c.due_date)}</TableCell>
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
        </Tabs>
      </div>
    </div>
  );
}
