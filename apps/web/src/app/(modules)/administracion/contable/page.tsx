"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "@/lib/api/client";
import { wsManager } from "@/lib/ws/manager";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { BookOpenIcon, ListTreeIcon, BarChart3Icon } from "lucide-react";

interface Account {
  id: string; code: string; name: string; account_type: string;
  is_detail: boolean; active: boolean; parent_id: string | null;
}
interface JournalEntry {
  id: string; number: string; date: string; concept: string;
  entry_type: string; status: string; user_id: string; created_at: string;
}
interface JournalLine {
  account_code: string; account_name: string; debit: number;
  credit: number; description: string;
}
interface Balance {
  account_code: string; account_name: string;
  total_debit: number; total_credit: number; balance: number;
}

const fmtMoney = (n: number) => n === 0 ? "—" : new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(n);
const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  posted: { label: "Contabilizado", variant: "default" },
  reversed: { label: "Reversado", variant: "outline" },
};

export default function ContablePage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [entries, setEntries] = useState<JournalEntry[]>([]);
  const [selectedEntry, setSelectedEntry] = useState<{ entry: JournalEntry; lines: JournalLine[] } | null>(null);
  const [balances, setBalances] = useState<Balance[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchAccounts = useCallback(async () => {
    try {
      const data = await api.get<{ accounts: Account[] }>("/v1/erp/accounting/accounts");
      setAccounts(data.accounts);
    } catch (err) { console.error(err); } finally { setLoading(false); }
  }, []);

  const fetchEntries = useCallback(async () => {
    try {
      const data = await api.get<{ entries: JournalEntry[] }>("/v1/erp/accounting/entries?page_size=50");
      setEntries(data.entries);
    } catch (err) { console.error(err); }
  }, []);

  const fetchBalance = useCallback(async () => {
    try {
      const data = await api.get<{ balances: Balance[] }>("/v1/erp/accounting/balance");
      setBalances(data.balances);
    } catch (err) { console.error(err); }
  }, []);

  const fetchEntryDetail = useCallback(async (id: string) => {
    try {
      const data = await api.get<{ entry: JournalEntry; lines: JournalLine[] }>(`/v1/erp/accounting/entries/${id}`);
      setSelectedEntry(data);
    } catch (err) { console.error(err); }
  }, []);

  useEffect(() => { fetchAccounts(); fetchEntries(); fetchBalance(); }, [fetchAccounts, fetchEntries, fetchBalance]);

  useEffect(() => {
    const handler = () => { fetchEntries(); fetchBalance(); };
    const unsub = wsManager.subscribe("erp_accounting", handler);
    return unsub;
  }, [fetchEntries, fetchBalance]);

  if (loading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const totalDebit = balances.reduce((a, b) => a + (b.total_debit || 0), 0);
  const totalCredit = balances.reduce((a, b) => a + (b.total_credit || 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Contabilidad</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Plan de cuentas, libro diario y balance — {accounts.length} cuentas, {entries.length} asientos
          </p>
        </div>

        <div className="grid grid-cols-3 gap-3 mb-6">
          {[
            { label: "Total Debe", value: fmtMoney(totalDebit), color: "text-blue-500" },
            { label: "Total Haber", value: fmtMoney(totalCredit), color: "text-blue-500" },
            { label: "Balance", value: fmtMoney(totalDebit - totalCredit), color: totalDebit === totalCredit ? "text-green-500" : "text-red-500" },
          ].map((s) => (
            <div key={s.label} className="rounded-xl border border-border/40 bg-card p-4">
              <p className="text-xs text-muted-foreground mb-1">{s.label}</p>
              <p className={`text-xl font-semibold ${s.color}`}>{s.value}</p>
            </div>
          ))}
        </div>

        <Tabs defaultValue="diary">
          <TabsList className="mb-4">
            <TabsTrigger value="diary"><BookOpenIcon className="size-3.5 mr-1.5" />Libro Diario</TabsTrigger>
            <TabsTrigger value="accounts"><ListTreeIcon className="size-3.5 mr-1.5" />Plan de Cuentas</TabsTrigger>
            <TabsTrigger value="balance"><BarChart3Icon className="size-3.5 mr-1.5" />Balance</TabsTrigger>
          </TabsList>

          <TabsContent value="diary">
            <div className="flex gap-6">
              <div className="flex-1 min-w-0 rounded-xl border border-border/40 bg-card overflow-hidden">
                <Table>
                  <TableHeader><TableRow>
                    <TableHead className="w-24">Asiento</TableHead>
                    <TableHead className="w-28">Fecha</TableHead>
                    <TableHead>Concepto</TableHead>
                    <TableHead className="w-28">Estado</TableHead>
                  </TableRow></TableHeader>
                  <TableBody>
                    {entries.map((e) => {
                      const s = statusBadge[e.status] || statusBadge.draft;
                      return (
                        <TableRow key={e.id} className="cursor-pointer" onClick={() => fetchEntryDetail(e.id)}>
                          <TableCell className="font-mono text-sm">{e.number}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">{new Date(e.date).toLocaleDateString("es-AR", { day: "2-digit", month: "short" })}</TableCell>
                          <TableCell className="text-sm">{e.concept}</TableCell>
                          <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                        </TableRow>
                      );
                    })}
                    {entries.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin asientos.</TableCell></TableRow>}
                  </TableBody>
                </Table>
              </div>

              {selectedEntry && (
                <div className="w-80 shrink-0 rounded-xl border border-border/40 bg-card p-5">
                  <h3 className="font-semibold mb-1">{selectedEntry.entry.number}</h3>
                  <p className="text-sm text-muted-foreground mb-4">{selectedEntry.entry.concept}</p>
                  <div className="space-y-2">
                    {selectedEntry.lines.map((l, i) => (
                      <div key={i} className="flex justify-between text-sm">
                        <div>
                          <p className="font-mono text-xs text-muted-foreground">{l.account_code}</p>
                          <p>{l.account_name}</p>
                        </div>
                        <div className="text-right font-mono">
                          {l.debit > 0 && <p className="text-blue-600">{fmtMoney(l.debit)}</p>}
                          {l.credit > 0 && <p className="text-muted-foreground">{fmtMoney(l.credit)}</p>}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </TabsContent>

          <TabsContent value="accounts">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-32">Codigo</TableHead>
                  <TableHead>Nombre</TableHead>
                  <TableHead className="w-28">Tipo</TableHead>
                  <TableHead className="w-28 text-center">Imputable</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {accounts.map((a) => (
                    <TableRow key={a.id}>
                      <TableCell className="font-mono text-sm">{a.code}</TableCell>
                      <TableCell className={`text-sm ${!a.is_detail ? "font-semibold" : ""}`}>{a.name}</TableCell>
                      <TableCell><Badge variant="secondary">{a.account_type}</Badge></TableCell>
                      <TableCell className="text-center text-sm">{a.is_detail ? "Si" : "No"}</TableCell>
                    </TableRow>
                  ))}
                  {accounts.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin cuentas.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="balance">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-28">Cuenta</TableHead>
                  <TableHead>Nombre</TableHead>
                  <TableHead className="text-right">Debe</TableHead>
                  <TableHead className="text-right">Haber</TableHead>
                  <TableHead className="text-right">Saldo</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {balances.map((b, i) => (
                    <TableRow key={i}>
                      <TableCell className="font-mono text-sm">{b.account_code}</TableCell>
                      <TableCell className="text-sm">{b.account_name}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(b.total_debit)}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{fmtMoney(b.total_credit)}</TableCell>
                      <TableCell className={`text-right font-mono text-sm font-medium ${b.balance >= 0 ? "" : "text-red-500"}`}>{fmtMoney(b.balance)}</TableCell>
                    </TableRow>
                  ))}
                  {balances.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin datos de balance.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
