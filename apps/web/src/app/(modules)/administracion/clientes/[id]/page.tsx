"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import type { EntityDetail } from "@/lib/erp/types";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export default function CustomerDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, error } = useQuery({
    queryKey: erpKeys.entity(id),
    queryFn: () => api.get<EntityDetail>(`/v1/erp/entities/${id}`),
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando cliente" onRetry={() => window.location.reload()} />;

  const entity = data?.entity;
  const contacts = data?.contacts ?? [];
  const notes = data?.notes ?? [];
  const documents = data?.documents ?? [];

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/clientes"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a clientes
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

            <div className="mb-6 grid grid-cols-3 gap-4">
              <Metric label="Contactos" value={String(contacts.length)} />
              <Metric label="Notas" value={String(notes.length)} />
              <Metric label="Documentos" value={String(documents.length)} />
            </div>

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Contactos ({contacts.length})
            </h2>
            <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
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
                      <TableCell colSpan={3} className="h-20 text-center text-sm text-muted-foreground">
                        Sin contactos registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {contacts.map((c) => (
                    <TableRow key={c.id}>
                      <TableCell className="text-sm">{c.type}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{c.label || "—"}</TableCell>
                      <TableCell className="text-sm">{c.value}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            {documents.length > 0 && (
              <>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground">
                  Documentos ({documents.length})
                </h2>
                <div className="mb-6 overflow-hidden rounded-xl border border-border/40 bg-card">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[160px]">Tipo</TableHead>
                        <TableHead>Archivo</TableHead>
                        <TableHead className="w-[150px]">Fecha</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {documents.map((d) => (
                        <TableRow key={d.id}>
                          <TableCell className="text-sm">{d.doc_type}</TableCell>
                          <TableCell className="font-mono text-xs text-muted-foreground">{d.filename}</TableCell>
                          <TableCell className="font-mono text-xs text-muted-foreground">{fmtDate(d.created_at)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </>
            )}

            {notes.length > 0 && (
              <>
                <h2 className="mb-3 text-sm font-medium text-muted-foreground">Notas ({notes.length})</h2>
                <div className="space-y-2">
                  {notes.map((n) => (
                    <div key={n.id} className="rounded-lg border border-border/40 bg-card px-4 py-3">
                      <div className="text-xs text-muted-foreground">
                        {fmtDate(n.created_at)} · {n.user_id}
                      </div>
                      <p className="mt-1 text-sm whitespace-pre-wrap">{n.body}</p>
                    </div>
                  ))}
                </div>
              </>
            )}
          </>
        )}
      </div>
    </div>
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
