"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDate } from "@/lib/erp/format";
import { ErrorState } from "@/components/erp/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface ToolDetail {
  id: string;
  tenant_id: string;
  legacy_id: number;
  code: string;
  article_code: string;
  article_id: string | null;
  inventory_code: string;
  name: string;
  characteristic: string;
  group_code: number;
  tool_type: number;
  status_code: number;
  purchase_order_no: number;
  purchase_order_date: string | null;
  delivery_note_date: string | null;
  delivery_note_post: number;
  delivery_note_no: number;
  supplier_code: number;
  pending_oc: string | null;
  observation: string;
  manufacture_no: number;
  generated_at: string | null;
  created_at: string;
}

interface ToolMovement {
  id: string;
  tool_id: string | null;
  tool_code: string;
  user_code: string;
  quantity: string;
  movement_date: string | null;
  concept_code: string;
  created_at: string;
}

const statusBadge: Record<number, { label: string; variant: "default" | "secondary" | "outline" | "destructive" }> = {
  0: { label: "Activa", variant: "default" },
  1: { label: "En uso", variant: "secondary" },
  2: { label: "Mantenimiento", variant: "outline" },
  3: { label: "Baja", variant: "destructive" },
};

export default function ToolDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: tool, isLoading, error } = useQuery({
    queryKey: [...erpKeys.all, "admin", "tools", id] as const,
    queryFn: () => api.get<ToolDetail>(`/v1/erp/admin/tools/${id}`),
    enabled: !!id,
  });

  const { data: movements = [] } = useQuery({
    queryKey: [...erpKeys.all, "admin", "tools", id, "movements"] as const,
    queryFn: () =>
      api.get<{ movements: ToolMovement[] }>(
        `/v1/erp/admin/tool-movements?tool_id=${id}&page_size=50`,
      ),
    select: (d) => d.movements ?? [],
    enabled: !!id,
  });

  if (error)
    return <ErrorState message="Error cargando herramienta" onRetry={() => window.location.reload()} />;

  const status = tool ? (statusBadge[tool.status_code] ?? { label: `#${tool.status_code}`, variant: "outline" as const }) : null;

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-4xl px-6 py-8 sm:px-8">
        <Link
          href="/administracion/almacen/herramientas"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Volver a herramientas
        </Link>

        {isLoading && <Skeleton className="h-48 w-full" />}

        {tool && (
          <>
            <div className="mb-6 flex items-baseline justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight">{tool.name}</h1>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  Código <span className="font-mono">{tool.code}</span>
                  {tool.inventory_code ? ` · Inventario ${tool.inventory_code}` : ""}
                </p>
              </div>
              {status && <Badge variant={status.variant}>{status.label}</Badge>}
            </div>

            <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Artículo</div>
                <div className="mt-1 font-mono text-sm">
                  {tool.article_id ? (
                    <Link href={`/administracion/almacen/articulos/${tool.article_id}`} className="hover:underline">
                      {tool.article_code || tool.article_id.slice(0, 8)}
                    </Link>
                  ) : (
                    tool.article_code || "—"
                  )}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Grupo / Tipo</div>
                <div className="mt-1 font-mono text-sm">
                  {tool.group_code} / {tool.tool_type}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Proveedor (legacy)</div>
                <div className="mt-1 font-mono text-sm">{tool.supplier_code || "—"}</div>
              </div>
            </div>

            {tool.characteristic && (
              <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Característica</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{tool.characteristic}</p>
              </div>
            )}

            <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">OC Nº</div>
                <div className="mt-1 font-mono text-sm">
                  {tool.purchase_order_no || "—"}
                  {tool.purchase_order_date ? ` · ${fmtDate(tool.purchase_order_date)}` : ""}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Remito</div>
                <div className="mt-1 font-mono text-sm">
                  {tool.delivery_note_no || "—"}
                  {tool.delivery_note_date ? ` · ${fmtDate(tool.delivery_note_date)}` : ""}
                </div>
              </div>
              <div className="rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Pendiente OC</div>
                <div className="mt-1 font-mono text-sm">
                  {tool.pending_oc ? Number(tool.pending_oc).toFixed(2) : "—"}
                </div>
              </div>
            </div>

            {tool.observation && (
              <div className="mb-6 rounded-lg border border-border/40 bg-card px-4 py-3">
                <div className="text-xs text-muted-foreground">Observación</div>
                <p className="mt-1 text-sm whitespace-pre-wrap">{tool.observation}</p>
              </div>
            )}

            <h2 className="mb-3 text-sm font-medium text-muted-foreground">
              Historial de movimientos ({movements.length})
            </h2>
            <div className="overflow-hidden rounded-xl border border-border/40 bg-card">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[120px]">Fecha</TableHead>
                    <TableHead className="w-[120px]">Operario</TableHead>
                    <TableHead className="w-[110px] text-right">Cantidad</TableHead>
                    <TableHead>Concepto</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {movements.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="h-16 text-center text-sm text-muted-foreground">
                        Sin movimientos registrados.
                      </TableCell>
                    </TableRow>
                  )}
                  {movements.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell className="font-mono text-xs">
                        {m.movement_date ? fmtDate(m.movement_date) : "—"}
                      </TableCell>
                      <TableCell className="font-mono text-xs">{m.user_code}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{Number(m.quantity).toFixed(2)}</TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">{m.concept_code || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
