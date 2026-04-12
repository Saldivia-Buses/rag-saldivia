"use client";

import { useState } from "react";
import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDate, fmtDateShort } from "@/lib/erp/format";
import type { Account, JournalEntry, JournalLine, AccountBalance, FiscalYear, CostCenter } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { ERPPagination } from "@/components/erp/pagination";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { BookOpenIcon, ListTreeIcon, BarChart3Icon, CalendarIcon, TargetIcon } from "lucide-react";

const statusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  posted: { label: "Contabilizado", variant: "default" },
  reversed: { label: "Reversado", variant: "outline" },
};

export default function ContablePage() {
  const [selectedEntryId, setSelectedEntryId] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const pageSize = 25;

  const { data: accounts = [], isLoading: loadingAccounts, error: errorAccounts } = useQuery({
    queryKey: erpKeys.accounts(),
    queryFn: () => api.get<{ accounts: Account[] }>("/v1/erp/accounting/accounts"),
    select: (d) => d.accounts,
  });

  const { data: entriesData } = useQuery({
    queryKey: erpKeys.entries({ page: String(page), page_size: String(pageSize) }),
    queryFn: () => api.get<{ entries: JournalEntry[]; total: number }>(`/v1/erp/accounting/entries?page=${page}&page_size=${pageSize}`),
    placeholderData: keepPreviousData,
  });
  const entries = entriesData?.entries ?? [];
  const entriesTotal = entriesData?.total ?? 0;

  const { data: balances = [] } = useQuery({
    queryKey: erpKeys.balance(),
    queryFn: () => api.get<{ balances: AccountBalance[] }>("/v1/erp/accounting/balance"),
    select: (d) => d.balances,
  });

  const { data: fiscalYears = [] } = useQuery({
    queryKey: erpKeys.fiscalYears(),
    queryFn: () => api.get<{ fiscal_years: FiscalYear[] }>("/v1/erp/accounting/fiscal-years"),
    select: (d) => d.fiscal_years,
  });

  const { data: costCenters = [] } = useQuery({
    queryKey: [...erpKeys.all, "cost-centers"] as const,
    queryFn: () => api.get<{ cost_centers: CostCenter[] }>("/v1/erp/accounting/cost-centers"),
    select: (d) => d.cost_centers,
  });

  const { data: selectedEntry } = useQuery({
    queryKey: erpKeys.entry(selectedEntryId!),
    queryFn: () => api.get<{ entry: JournalEntry; lines: JournalLine[] }>(`/v1/erp/accounting/entries/${selectedEntryId}`),
    enabled: !!selectedEntryId,
  });

  if (errorAccounts) return <ErrorState message="Error cargando contabilidad" onRetry={() => window.location.reload()} />;
  if (loadingAccounts) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

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
            <TabsTrigger value="fiscal-years"><CalendarIcon className="size-3.5 mr-1.5" />Ejercicios</TabsTrigger>
            <TabsTrigger value="cost-centers"><TargetIcon className="size-3.5 mr-1.5" />Centros de Costo</TabsTrigger>
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
                        <TableRow key={e.id} className="cursor-pointer" onClick={() => setSelectedEntryId(e.id)}>
                          <TableCell className="font-mono text-sm">{e.number}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">{fmtDateShort(e.date)}</TableCell>
                          <TableCell className="text-sm">{e.concept}</TableCell>
                          <TableCell><Badge variant={s.variant}>{s.label}</Badge></TableCell>
                        </TableRow>
                      );
                    })}
                    {entries.length === 0 && <TableRow><TableCell colSpan={4} className="h-24 text-center text-muted-foreground">Sin asientos.</TableCell></TableRow>}
                  </TableBody>
                </Table>
                <ERPPagination page={page} pageSize={pageSize} total={entriesTotal} onPageChange={setPage} />
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
                  <TableHead className="w-32">Código</TableHead>
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
                      <TableCell className="text-center text-sm">{a.is_detail ? "Sí" : "No"}</TableCell>
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

          <TabsContent value="fiscal-years">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-20">Año</TableHead>
                  <TableHead className="w-32">Inicio</TableHead>
                  <TableHead className="w-32">Fin</TableHead>
                  <TableHead className="w-28">Estado</TableHead>
                  <TableHead className="w-36">Cerrado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {fiscalYears.map((fy) => (
                    <TableRow key={fy.id}>
                      <TableCell className="font-mono text-sm font-medium">{fy.year}</TableCell>
                      <TableCell className="text-sm">{fmtDate(fy.start_date)}</TableCell>
                      <TableCell className="text-sm">{fmtDate(fy.end_date)}</TableCell>
                      <TableCell><Badge variant={fy.status === "open" ? "default" : "secondary"}>{fy.status === "open" ? "Abierto" : "Cerrado"}</Badge></TableCell>
                      <TableCell className="text-sm text-muted-foreground">{fy.closed_at ? fmtDate(fy.closed_at) : "\u2014"}</TableCell>
                    </TableRow>
                  ))}
                  {fiscalYears.length === 0 && <TableRow><TableCell colSpan={5} className="h-24 text-center text-muted-foreground">Sin ejercicios fiscales.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="cost-centers">
            <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
              <Table>
                <TableHeader><TableRow>
                  <TableHead className="w-28">Código</TableHead>
                  <TableHead>Nombre</TableHead>
                  <TableHead className="w-24 text-center">Estado</TableHead>
                </TableRow></TableHeader>
                <TableBody>
                  {costCenters.map((cc) => (
                    <TableRow key={cc.id}>
                      <TableCell className="font-mono text-sm">{cc.code}</TableCell>
                      <TableCell className="text-sm">{cc.name}</TableCell>
                      <TableCell className="text-center"><Badge variant={cc.active ? "default" : "secondary"}>{cc.active ? "Activo" : "Inactivo"}</Badge></TableCell>
                    </TableRow>
                  ))}
                  {costCenters.length === 0 && <TableRow><TableCell colSpan={3} className="h-24 text-center text-muted-foreground">Sin centros de costo.</TableCell></TableRow>}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
