"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FileTextIcon, WalletIcon, AlertTriangleIcon } from "lucide-react";

interface Balance { entity_name: string; entity_type: string; direction: string; open_balance: number; }
interface Overdue { entity_name: string; invoice_number: string; due_date: string; amount: number; balance: number; }

const fmtMoney = (n: number) => n === 0 ? "—" : new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
const fmtDate = (s: string) => new Date(s).toLocaleDateString("es-AR", { day: "2-digit", month: "short" });

export default function CuentasPage() {
  const [balances, setBalances] = useState<Balance[]>([]);
  const [overdue, setOverdue] = useState<Overdue[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const [b, o] = await Promise.all([
        api.get<{ balances: Balance[] }>("/v1/erp/accounts/balances"),
        api.get<{ overdue: Overdue[] }>("/v1/erp/accounts/overdue"),
      ]);
      setBalances(b.balances); setOverdue(o.overdue);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, []);

  useEffect(() => { fetch(); }, [fetch]);
  useEffect(() => { const unsub = wsManager.subscribe("erp_accounts", fetch); return unsub; }, [fetch]);

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const receivable = balances.filter(b => b.direction === "receivable");
  const payable = balances.filter(b => b.direction === "payable");

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Cuentas Corrientes</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Saldos, vencidos y estado de cuenta</p>
        </div>

        <Tabs defaultValue="balances">
          <TabsList className="mb-4">
            <TabsTrigger value="balances"><WalletIcon className="size-3.5 mr-1.5" />Saldos</TabsTrigger>
            <TabsTrigger value="overdue"><AlertTriangleIcon className="size-3.5 mr-1.5" />Vencidos ({overdue.length})</TabsTrigger>
          </TabsList>

          <TabsContent value="balances">
            <div className="grid grid-cols-2 gap-6">
              <div>
                <h3 className="text-sm font-medium mb-3">A cobrar (clientes nos deben)</h3>
                <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
                  <Table>
                    <TableHeader><TableRow><TableHead>Entidad</TableHead><TableHead className="text-right">Saldo</TableHead></TableRow></TableHeader>
                    <TableBody>
                      {receivable.map((b, i) => (
                        <TableRow key={i}>
                          <TableCell className="text-sm">{b.entity_name}</TableCell>
                          <TableCell className="text-right font-mono text-sm text-green-600">{fmtMoney(b.open_balance)}</TableCell>
                        </TableRow>
                      ))}
                      {receivable.length === 0 && <TableRow><TableCell colSpan={2} className="h-16 text-center text-muted-foreground">Sin saldos a cobrar.</TableCell></TableRow>}
                    </TableBody>
                  </Table>
                </div>
              </div>
              <div>
                <h3 className="text-sm font-medium mb-3">A pagar (debemos a proveedores)</h3>
                <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
                  <Table>
                    <TableHeader><TableRow><TableHead>Entidad</TableHead><TableHead className="text-right">Saldo</TableHead></TableRow></TableHeader>
                    <TableBody>
                      {payable.map((b, i) => (
                        <TableRow key={i}>
                          <TableCell className="text-sm">{b.entity_name}</TableCell>
                          <TableCell className="text-right font-mono text-sm text-red-500">{fmtMoney(b.open_balance)}</TableCell>
                        </TableRow>
                      ))}
                      {payable.length === 0 && <TableRow><TableCell colSpan={2} className="h-16 text-center text-muted-foreground">Sin saldos a pagar.</TableCell></TableRow>}
                    </TableBody>
                  </Table>
                </div>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="overdue">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead>Entidad</TableHead>
                  <TableHead className="w-32">Factura</TableHead>
                  <TableHead className="w-28">Vencimiento</TableHead>
                  <TableHead className="text-right w-28">Monto</TableHead>
                  <TableHead className="text-right w-28">Saldo</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {overdue.map((o, i) => (
                    <TableRow key={i}>
                      <TableCell className="text-sm">{o.entity_name}</TableCell>
                      <TableCell className="font-mono text-sm">{o.invoice_number}</TableCell>
                      <TableCell className="text-sm text-red-500 font-medium">{fmtDate(o.due_date)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(o.amount)}</TableCell>
                      <TableCell className="text-right font-mono text-sm font-medium text-red-500">{fmtMoney(o.balance)}</TableCell>
                    </TableRow>
                  ))}
                  {overdue.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin facturas vencidas.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
