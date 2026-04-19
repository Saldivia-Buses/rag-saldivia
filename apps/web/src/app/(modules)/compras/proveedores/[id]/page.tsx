"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { EntityContact, EntityDetail, EntityNote, SupplierDemerit } from "@/lib/erp/types";
import { Textarea } from "@/components/ui/textarea";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const CONTACT_TYPES = ["phone", "email", "address", "bank_account"] as const;
type ContactType = (typeof CONTACT_TYPES)[number];

export default function SupplierDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const qc = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.entity(id),
    queryFn: () => api.get<EntityDetail>(`/v1/erp/entities/${id}`),
    enabled: !!id,
  });

  const { data: demerits = [] } = useQuery({
    queryKey: erpKeys.supplierDemerits(id),
    queryFn: () =>
      api.get<{ demerits: SupplierDemerit[] }>(`/v1/erp/purchasing/suppliers/${id}/demerits`),
    select: (d) => d.demerits,
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando proveedor" onRetry={() => window.location.reload()} />;

  const entity = data?.entity;
  const contacts = data?.contacts ?? [];
  const notes = data?.notes ?? [];
  const totalPoints = demerits.reduce((s, d) => s + (d.points ?? 0), 0);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/compras/proveedores"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a proveedores
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {entity && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{entity.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Código <span className="font-mono">{entity.code}</span>
                  {entity.email ? ` · ${entity.email}` : ""}
                  {entity.phone ? ` · ${entity.phone}` : ""}
                </p>
              </div>
              <Badge variant={entity.active ? "default" : "secondary"}>
                {entity.active ? "Activo" : "Inactivo"}
              </Badge>
            </div>

            <div className="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Metric label="Contactos" value={String(contacts.length)} />
              <Metric label="Notas" value={String(notes.length)} />
              <Metric label="Demeritos" value={String(demerits.length)} />
              <Metric label="Puntos" value={String(totalPoints)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Contactos ({contacts.length})</h2>
            <div className="mb-4 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px]">Tipo</TableHead>
                    <TableHead className="w-[180px]">Etiqueta</TableHead>
                    <TableHead>Valor</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {contacts.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className="h-16 text-center text-sm text-muted-foreground">
                        Sin contactos registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {contacts.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="text-sm">{labelForType(c.type)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{c.label || "—"}</TableCell>
                      <TableCell className="text-sm">{c.value}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <AddContactForm
              entityId={id}
              onSuccess={() => qc.invalidateQueries({ queryKey: erpKeys.entity(id) })}
            />

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Demeritos ({demerits.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[130px]">Fecha</TableHead>
                    <TableHead className="w-[90px] text-right">Puntos</TableHead>
                    <TableHead>Motivo</TableHead>
                    <TableHead className="w-[180px]">Inspección</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {demerits.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-20 text-center text-sm text-muted-foreground">
                        Sin demeritos registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {demerits.map((d) => (
                    <TableRow key={d.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">{fmtDate(d.created_at)}</TableCell>
                      <TableCell className="text-right font-mono text-sm text-destructive">{d.points}</TableCell>
                      <TableCell className="text-sm">{d.reason}</TableCell>
                      <TableCell className="font-mono text-xs">
                        {d.inspection_id ? (
                          <Link href={`/calidad/inspecciones/${d.inspection_id}`} className="hover:underline">
                            {d.inspection_id.slice(0, 8)}
                          </Link>
                        ) : (
                          "—"
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">Notas ({notes.length})</h2>
            {notes.length > 0 ? (
              <div className="mb-4 space-y-2">
                {notes.map((n) => (
                  <div key={n.id} className="rounded-lg border border-border/40 bg-card px-4 py-3">
                    <div className="text-xs text-muted-foreground">
                      {fmtDate(n.created_at)} · {n.user_id}
                      {n.type && n.type !== "note" ? ` · ${n.type}` : ""}
                    </div>
                    <p className="mt-1 text-sm whitespace-pre-wrap">{n.body}</p>
                  </div>
                ))}
              </div>
            ) : (
              <p className="mb-4 text-sm text-muted-foreground">Sin notas registradas.</p>
            )}

            <AddNoteForm
              entityId={id}
              onSuccess={() => qc.invalidateQueries({ queryKey: erpKeys.entity(id) })}
            />
          </>
        )}
      </div>
    </div>
  );
}

function labelForType(type: string) {
  switch (type) {
    case "phone":
      return "Teléfono";
    case "email":
      return "Email";
    case "address":
      return "Dirección";
    case "bank_account":
      return "Cuenta bancaria";
    default:
      return type;
  }
}

function AddContactForm({ entityId, onSuccess }: { entityId: string; onSuccess: () => void }) {
  const [type, setType] = useState<ContactType>("phone");
  const [label, setLabel] = useState("");
  const [value, setValue] = useState("");

  const mutation = useMutation({
    mutationFn: (body: { type: string; label: string; value: string }) =>
      api.post<EntityContact>(`/v1/erp/entities/${entityId}/contacts`, body),
    onSuccess: () => {
      setLabel("");
      setValue("");
      onSuccess();
    },
  });

  return (
    <form
      className="mb-6 rounded-xl border border-border/40 bg-card p-4"
      onSubmit={(e) => {
        e.preventDefault();
        if (!value.trim()) return;
        mutation.mutate({ type, label: label.trim(), value: value.trim() });
      }}
    >
      <h3 className="mb-3 text-sm font-medium">Agregar contacto</h3>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-[140px_180px_1fr_auto]">
        <select
          className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
          value={type}
          onChange={(e) => setType(e.target.value as ContactType)}
          disabled={mutation.isPending}
        >
          {CONTACT_TYPES.map((t) => (
            <option key={t} value={t}>
              {labelForType(t)}
            </option>
          ))}
        </select>
        <Input
          placeholder="Etiqueta (opcional)"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          disabled={mutation.isPending}
        />
        <Input
          placeholder="Valor"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          disabled={mutation.isPending}
          required
        />
        <Button type="submit" disabled={mutation.isPending || !value.trim()}>
          {mutation.isPending ? "Guardando…" : "Agregar"}
        </Button>
      </div>
      {mutation.isError && (
        <p className="mt-2 text-xs text-destructive">Error al guardar contacto.</p>
      )}
    </form>
  );
}

function AddNoteForm({ entityId, onSuccess }: { entityId: string; onSuccess: () => void }) {
  const [body, setBody] = useState("");

  const mutation = useMutation({
    mutationFn: (payload: { type: string; body: string }) =>
      api.post<EntityNote>(`/v1/erp/entities/${entityId}/notes`, payload),
    onSuccess: () => {
      setBody("");
      onSuccess();
    },
  });

  return (
    <form
      className="mb-6 rounded-xl border border-border/40 bg-card p-4"
      onSubmit={(e) => {
        e.preventDefault();
        if (!body.trim()) return;
        mutation.mutate({ type: "note", body: body.trim() });
      }}
    >
      <h3 className="mb-3 text-sm font-medium">Agregar nota</h3>
      <Textarea
        placeholder="Escribí la nota…"
        value={body}
        onChange={(e) => setBody(e.target.value)}
        disabled={mutation.isPending}
        rows={3}
        required
      />
      <div className="mt-3 flex items-center justify-between">
        {mutation.isError ? (
          <p className="text-xs text-destructive">Error al guardar nota.</p>
        ) : (
          <span />
        )}
        <Button type="submit" disabled={mutation.isPending || !body.trim()}>
          {mutation.isPending ? "Guardando…" : "Agregar nota"}
        </Button>
      </div>
    </form>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 font-mono text-sm">{value}</div>
    </div>
  );
}
