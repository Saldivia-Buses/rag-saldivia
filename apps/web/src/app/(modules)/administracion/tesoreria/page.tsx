"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtMoney, fmtDateShort } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import type { TreasuryMovement, Check, BankBalance, Reconciliation, Receipt } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { BanknoteIcon, CreditCardIcon, LandmarkIcon, ScaleIcon, ReceiptIcon, PlusIcon } from "lucide-react";

const moveLabel: Record<string, string> = {
  cash_in: "Ingreso caja", cash_out: "Egreso caja", bank_deposit: "Depósito",
  bank_withdrawal: "Retiro", check_issued: "Cheque emitido",
  check_received: "Cheque recibido", transfer: "Transferencia",
};
const checkStatus: Record<string, string> = {
  in_portfolio: "En cartera", deposited: "Depositado", cashed: "Cobrado",
  rejected: "Rechazado", endorsed: "Endosado",
};

interface CreateMovementBody {
  number: string;
  date: string;
  movement_type: string;
  amount: string;
  entity_id?: string | null;
  description?: string;
}

interface CreateReceiptBody {
  number: string;
  date: string;
  receipt_type: string;
  entity_id: string;
  payments: Array<{ payment_method: string; amount: string; reference?: string }>;
  allocations: never[];
}

export default function TesoreriaPage() {
  const [movementOpen, setMovementOpen] = useState(false);
  const [receiptOpen, setReceiptOpen] = useState(false);
  const queryClient = useQueryClient();

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

  const createMovementMutation = useMutation({
    mutationFn: (data: CreateMovementBody) => api.post("/v1/erp/treasury/movements", data),
    onSuccess: () => {
      toast.success("Movimiento registrado");
      queryClient.invalidateQueries({ queryKey: erpKeys.treasuryMovements() });
      setMovementOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createReceiptMutation = useMutation({
    mutationFn: (data: CreateReceiptBody) => api.post("/v1/erp/treasury/receipts", data),
    onSuccess: () => {
      toast.success("Recibo creado");
      queryClient.invalidateQueries({ queryKey: erpKeys.receipts() });
      setReceiptOpen(false);
    },
    onError: permissionErrorToast,
  });

  if (error) return <ErrorState message="Error cargando tesorería" onRetry={() => window.location.reload()} />;
  if (isLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const totalBalance = balances.reduce((a, b) => a + (b.balance || 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Tesorería</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Movimientos, cheques y saldos bancarios</p>
          </div>
          <Button size="sm" onClick={() => setMovementOpen(true)}>
            <PlusIcon className="size-4 mr-1.5" />Nuevo movimiento
          </Button>
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
            <div className="flex justify-end mb-4">
              <Button size="sm" onClick={() => setReceiptOpen(true)}>
                <PlusIcon className="size-4 mr-1.5" />Nuevo recibo
              </Button>
            </div>
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
                      <TableCell className="font-mono text-sm">
                        <Link href={`/administracion/tesoreria/recibos/${r.id}`} className="hover:underline">
                          {r.number}
                        </Link>
                      </TableCell>
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

      <Dialog open={movementOpen} onOpenChange={(v) => !v && setMovementOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo movimiento</DialogTitle></DialogHeader>
          <CreateMovementForm
            onSubmit={(data) => createMovementMutation.mutate(data)}
            isPending={createMovementMutation.isPending}
            onClose={() => setMovementOpen(false)}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={receiptOpen} onOpenChange={(v) => !v && setReceiptOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo recibo</DialogTitle></DialogHeader>
          <CreateReceiptForm
            onSubmit={(data) => createReceiptMutation.mutate(data)}
            isPending={createReceiptMutation.isPending}
            onClose={() => setReceiptOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateMovementForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: CreateMovementBody) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [movementType, setMovementType] = useState("cash_in");
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [entityId, setEntityId] = useState("");

  const { data: entities = [] } = useQuery({
    queryKey: [...erpKeys.all, "entities", "all"],
    queryFn: () => api.get<{ entities: Array<{ id: string; name: string }> }>("/v1/erp/entities?page_size=200"),
    select: (d) => d.entities,
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!number || !date || !amount) return;
    onSubmit({
      number,
      date,
      movement_type: movementType,
      amount,
      entity_id: entityId || null,
      description: description || undefined,
    });
    setNumber("");
    setDate(today);
    setMovementType("cash_in");
    setAmount("");
    setDescription("");
    setEntityId("");
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="MOV-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Tipo de movimiento</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={movementType}
          onChange={(e) => setMovementType(e.target.value)}
        >
          <option value="cash_in">Ingreso caja</option>
          <option value="cash_out">Egreso caja</option>
          <option value="bank_deposit">Depósito bancario</option>
          <option value="bank_withdrawal">Retiro bancario</option>
          <option value="transfer">Transferencia</option>
        </select>
      </div>

      <div className="space-y-2">
        <Label>Monto</Label>
        <Input
          type="number"
          min="0"
          step="0.01"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0.00"
        />
      </div>

      <div className="space-y-2">
        <Label>Entidad (opcional)</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={entityId}
          onChange={(e) => setEntityId(e.target.value)}
        >
          <option value="">— Sin entidad —</option>
          {entities.map((en) => (
            <option key={en.id} value={en.id}>{en.name}</option>
          ))}
        </select>
      </div>

      <div className="space-y-2">
        <Label>Descripción (opcional)</Label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Descripción del movimiento..." />
      </div>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!number || !date || !amount || isPending}>
          {isPending ? "Registrando..." : "Registrar"}
        </Button>
      </div>
    </form>
  );
}

function CreateReceiptForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: CreateReceiptBody) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [number, setNumber] = useState("");
  const [date, setDate] = useState(today);
  const [receiptType, setReceiptType] = useState("collection");
  const [entityId, setEntityId] = useState("");
  const [paymentMethod, setPaymentMethod] = useState("cash");
  const [amount, setAmount] = useState("");
  const [reference, setReference] = useState("");

  const { data: entities = [] } = useQuery({
    queryKey: [...erpKeys.all, "entities", "all"],
    queryFn: () => api.get<{ entities: Array<{ id: string; name: string }> }>("/v1/erp/entities?page_size=200"),
    select: (d) => d.entities,
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!number || !date || !entityId || !amount) return;
    onSubmit({
      number,
      date,
      receipt_type: receiptType,
      entity_id: entityId,
      payments: [{ payment_method: paymentMethod, amount, reference: reference || undefined }],
      allocations: [],
    });
    setNumber("");
    setDate(today);
    setReceiptType("collection");
    setEntityId("");
    setPaymentMethod("cash");
    setAmount("");
    setReference("");
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Número</Label>
          <Input value={number} onChange={(e) => setNumber(e.target.value)} placeholder="REC-001" />
        </div>
        <div className="space-y-2">
          <Label>Fecha</Label>
          <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Tipo</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={receiptType}
          onChange={(e) => setReceiptType(e.target.value)}
        >
          <option value="collection">Cobro</option>
          <option value="payment">Pago</option>
        </select>
      </div>

      <div className="space-y-2">
        <Label>Entidad</Label>
        <select
          className="w-full rounded-md border px-3 py-2 text-sm bg-card"
          value={entityId}
          onChange={(e) => setEntityId(e.target.value)}
        >
          <option value="">Seleccionar entidad...</option>
          {entities.map((en) => (
            <option key={en.id} value={en.id}>{en.name}</option>
          ))}
        </select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Método de pago</Label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm bg-card"
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value)}
          >
            <option value="cash">Efectivo</option>
            <option value="transfer">Transferencia</option>
            <option value="check">Cheque</option>
          </select>
        </div>
        <div className="space-y-2">
          <Label>Monto</Label>
          <Input
            type="number"
            min="0"
            step="0.01"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="0.00"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Referencia (opcional)</Label>
        <Input value={reference} onChange={(e) => setReference(e.target.value)} placeholder="Nro. de operación, cheque, etc." />
      </div>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!number || !date || !entityId || !amount || isPending}>
          {isPending ? "Creando..." : "Crear recibo"}
        </Button>
      </div>
    </form>
  );
}
