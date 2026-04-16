"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api/client";
import { erpKeys } from "@/lib/erp/queries";
import { fmtDateShort } from "@/lib/erp/format";
import { permissionErrorToast } from "@/lib/erp/permission-messages";
import { ErrorState } from "@/components/erp/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { PlusIcon, FileCheckIcon, AwardIcon, CheckCircleIcon } from "lucide-react";

interface ManufacturingUnit {
  id: string;
  work_order_number: number;
  chassis_serial: string;
  carroceria_model_name: string;
  customer_name: string;
  status: string;
}

interface CnrtWorkOrder {
  id: string;
  inspection_type: string;
  inspector_name: string;
  inspection_date: string;
  approved: boolean;
  status: string;
  observations: string;
}

interface Certificate {
  id: string;
  certificate_number: string;
  cert_type: string;
  issued_at: string | null;
  valid_from: string;
  valid_until: string | null;
  status: string;
  authority: string;
}

type CnrtInspectionType = "initial" | "periodic" | "extraordinary" | "modification";

const cnrtInspectionTypeOptions: { value: CnrtInspectionType; label: string }[] = [
  { value: "initial", label: "Inicial" },
  { value: "periodic", label: "Periódica" },
  { value: "extraordinary", label: "Extraordinaria" },
  { value: "modification", label: "Modificación" },
];

const cnrtStatusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  pending: { label: "Pendiente", variant: "secondary" },
  approved: { label: "Aprobada", variant: "default" },
  rejected: { label: "Rechazada", variant: "secondary" },
  in_progress: { label: "En curso", variant: "outline" },
};

const certStatusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  draft: { label: "Borrador", variant: "secondary" },
  issued: { label: "Emitido", variant: "default" },
  expired: { label: "Vencido", variant: "secondary" },
  revoked: { label: "Revocado", variant: "secondary" },
};

const unitStatusBadge: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
  pending: { label: "Pendiente", variant: "secondary" },
  in_production: { label: "En producción", variant: "outline" },
  completed: { label: "Terminada", variant: "default" },
  delivered: { label: "Entregada", variant: "default" },
  returned: { label: "Devuelta", variant: "secondary" },
};

const MFG_UNITS_KEY = [...erpKeys.all, "manufacturing", "units"] as const;

function cnrtKey(unitId: string) {
  return [...erpKeys.all, "manufacturing", "units", unitId, "cnrt"] as const;
}

function certKey(unitId: string) {
  return [...erpKeys.all, "manufacturing", "units", unitId, "certificate"] as const;
}

export default function CertificacionesPage() {
  const [selectedUnitId, setSelectedUnitId] = useState<string | null>(null);
  const [cnrtDialogOpen, setCnrtDialogOpen] = useState(false);
  const [certDialogOpen, setCertDialogOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: units = [], isLoading: unitsLoading, error: unitsError } = useQuery({
    queryKey: MFG_UNITS_KEY,
    queryFn: () => api.get<{ units: ManufacturingUnit[] }>("/v1/erp/manufacturing/units?page_size=100"),
    select: (d) => d.units,
  });

  const { data: cnrtOrders = [], isLoading: cnrtLoading } = useQuery({
    queryKey: cnrtKey(selectedUnitId ?? ""),
    queryFn: () => api.get<{ cnrt_work_orders: CnrtWorkOrder[] }>(`/v1/erp/manufacturing/units/${selectedUnitId}/cnrt`),
    enabled: !!selectedUnitId,
    select: (d) => d.cnrt_work_orders,
  });

  const { data: certificate, isLoading: certLoading } = useQuery({
    queryKey: certKey(selectedUnitId ?? ""),
    queryFn: () => api.get<{ certificate: Certificate | null }>(`/v1/erp/manufacturing/units/${selectedUnitId}/certificate`),
    enabled: !!selectedUnitId,
    select: (d) => d.certificate,
  });

  const createCnrtMutation = useMutation({
    mutationFn: (data: {
      inspection_type: CnrtInspectionType;
      inspector_name: string;
      inspection_date: string;
      observations: string;
    }) => api.post(`/v1/erp/manufacturing/units/${selectedUnitId}/cnrt`, data),
    onSuccess: () => {
      toast.success("Inspección CNRT registrada");
      queryClient.invalidateQueries({ queryKey: cnrtKey(selectedUnitId ?? "") });
      setCnrtDialogOpen(false);
    },
    onError: permissionErrorToast,
  });

  const createCertMutation = useMutation({
    mutationFn: (data: {
      certificate_number: string;
      cert_type: string;
      authority: string;
      valid_from: string;
      valid_until: string;
      observations: string;
    }) => api.post(`/v1/erp/manufacturing/units/${selectedUnitId}/certificate`, data),
    onSuccess: () => {
      toast.success("Certificado creado");
      queryClient.invalidateQueries({ queryKey: certKey(selectedUnitId ?? "") });
      setCertDialogOpen(false);
    },
    onError: permissionErrorToast,
  });

  const issueCertMutation = useMutation({
    mutationFn: (certId: string) =>
      api.post(`/v1/erp/manufacturing/units/${selectedUnitId}/certificate/${certId}/issue`, {}),
    onSuccess: () => {
      toast.success("Certificado emitido");
      queryClient.invalidateQueries({ queryKey: certKey(selectedUnitId ?? "") });
    },
    onError: permissionErrorToast,
  });

  if (unitsError) return <ErrorState message="Error cargando unidades" onRetry={() => window.location.reload()} />;
  if (unitsLoading) return <div className="flex-1 p-8"><Skeleton className="h-[600px]" /></div>;

  const selectedUnit = units.find((u) => u.id === selectedUnitId);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="mx-auto w-full max-w-6xl px-6 py-8 sm:px-8">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Certificaciones</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Inspecciones CNRT y certificados de fabricación — {units.length} unidades
          </p>
        </div>

        <div className="flex gap-6">
          {/* Unit list */}
          <div className="w-72 shrink-0 rounded-xl border border-border/40 bg-card overflow-hidden self-start">
            <div className="px-4 py-3 border-b border-border/40">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Unidades</p>
            </div>
            {units.length === 0 ? (
              <p className="px-4 py-6 text-sm text-center text-muted-foreground">Sin unidades registradas.</p>
            ) : (
              <ul>
                {units.map((u) => {
                  const s = unitStatusBadge[u.status] ?? unitStatusBadge.pending;
                  return (
                    <li key={u.id}>
                      <button
                        type="button"
                        className={`w-full text-left px-4 py-3 text-sm border-b border-border/20 last:border-0 transition-colors hover:bg-accent/40 ${selectedUnitId === u.id ? "bg-accent/60" : ""}`}
                        onClick={() => setSelectedUnitId(u.id)}
                      >
                        <div className="flex items-center justify-between gap-2">
                          <p className="font-mono font-semibold">OT {u.work_order_number}</p>
                          <Badge variant={s.variant} className="text-[10px]">{s.label}</Badge>
                        </div>
                        <p className="text-xs text-muted-foreground mt-0.5 truncate">
                          {u.carroceria_model_name || u.chassis_serial || "\u2014"}
                        </p>
                        {u.customer_name && (
                          <p className="text-xs text-muted-foreground truncate">{u.customer_name}</p>
                        )}
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>

          {/* Detail panel */}
          <div className="flex-1 min-w-0">
            {!selectedUnit ? (
              <div className="flex items-center justify-center h-64 rounded-xl border border-dashed border-border/40 text-muted-foreground text-sm">
                Seleccioná una unidad para ver sus certificaciones.
              </div>
            ) : (
              <>
                <div className="mb-4">
                  <p className="font-semibold">OT {selectedUnit.work_order_number}</p>
                  <p className="text-sm text-muted-foreground">
                    {selectedUnit.carroceria_model_name || selectedUnit.chassis_serial || "\u2014"}
                    {selectedUnit.customer_name ? ` — ${selectedUnit.customer_name}` : ""}
                  </p>
                </div>

                <Tabs defaultValue="cnrt">
                  <TabsList className="mb-4">
                    <TabsTrigger value="cnrt">
                      <FileCheckIcon className="size-3.5 mr-1.5" />CNRT
                    </TabsTrigger>
                    <TabsTrigger value="certificate">
                      <AwardIcon className="size-3.5 mr-1.5" />Certificado
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value="cnrt">
                    <div className="flex items-center justify-between mb-3">
                      <p className="text-sm text-muted-foreground">{cnrtOrders.length} inspecciones</p>
                      <Button size="sm" variant="outline" onClick={() => setCnrtDialogOpen(true)}>
                        <PlusIcon className="size-3.5 mr-1.5" />Nueva inspección
                      </Button>
                    </div>
                    {cnrtLoading ? (
                      <Skeleton className="h-48" />
                    ) : (
                      <div className="rounded-xl border border-border/40 bg-card overflow-hidden">
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>Tipo</TableHead>
                              <TableHead>Inspector</TableHead>
                              <TableHead className="w-24">Fecha</TableHead>
                              <TableHead className="w-24 text-center">Aprobada</TableHead>
                              <TableHead className="w-28">Estado</TableHead>
                              <TableHead>Observaciones</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {cnrtOrders.map((c) => {
                              const s = cnrtStatusBadge[c.status] ?? cnrtStatusBadge.pending;
                              const typeLabel = cnrtInspectionTypeOptions.find((o) => o.value === c.inspection_type)?.label ?? c.inspection_type;
                              return (
                                <TableRow key={c.id}>
                                  <TableCell className="text-sm font-medium">{typeLabel}</TableCell>
                                  <TableCell className="text-sm">{c.inspector_name}</TableCell>
                                  <TableCell className="text-sm text-muted-foreground">{fmtDateShort(c.inspection_date)}</TableCell>
                                  <TableCell className="text-center">
                                    {c.approved ? (
                                      <CheckCircleIcon className="size-4 text-green-600 mx-auto" />
                                    ) : (
                                      <span className="text-muted-foreground text-xs">\u2014</span>
                                    )}
                                  </TableCell>
                                  <TableCell>
                                    <Badge variant={s.variant} className="text-xs">{s.label}</Badge>
                                  </TableCell>
                                  <TableCell className="text-sm text-muted-foreground truncate max-w-[160px]">
                                    {c.observations || "\u2014"}
                                  </TableCell>
                                </TableRow>
                              );
                            })}
                            {cnrtOrders.length === 0 && (
                              <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                                  Sin inspecciones CNRT.
                                </TableCell>
                              </TableRow>
                            )}
                          </TableBody>
                        </Table>
                      </div>
                    )}
                  </TabsContent>

                  <TabsContent value="certificate">
                    {certLoading ? (
                      <Skeleton className="h-32" />
                    ) : certificate ? (
                      <div className="rounded-xl border border-border/40 bg-card p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div>
                            <p className="font-semibold font-mono">{certificate.certificate_number}</p>
                            <p className="text-sm text-muted-foreground mt-0.5">{certificate.cert_type} — {certificate.authority}</p>
                          </div>
                          <div className="flex items-center gap-2">
                            <Badge variant={certStatusBadge[certificate.status]?.variant ?? "secondary"}>
                              {certStatusBadge[certificate.status]?.label ?? certificate.status}
                            </Badge>
                            {certificate.status === "draft" && (
                              <Button
                                size="sm"
                                onClick={() => issueCertMutation.mutate(certificate.id)}
                                disabled={issueCertMutation.isPending}
                              >
                                <CheckCircleIcon className="size-3.5 mr-1.5" />
                                {issueCertMutation.isPending ? "Emitiendo..." : "Emitir"}
                              </Button>
                            )}
                          </div>
                        </div>
                        <div className="grid grid-cols-2 gap-4 text-sm">
                          <div>
                            <p className="text-xs text-muted-foreground">Válido desde</p>
                            <p className="font-medium">{fmtDateShort(certificate.valid_from)}</p>
                          </div>
                          <div>
                            <p className="text-xs text-muted-foreground">Válido hasta</p>
                            <p className="font-medium">
                              {certificate.valid_until ? fmtDateShort(certificate.valid_until) : "Sin vencimiento"}
                            </p>
                          </div>
                          {certificate.issued_at && (
                            <div>
                              <p className="text-xs text-muted-foreground">Emitido</p>
                              <p className="font-medium">{fmtDateShort(certificate.issued_at)}</p>
                            </div>
                          )}
                        </div>
                      </div>
                    ) : (
                      <div className="flex flex-col items-center justify-center h-40 rounded-xl border border-dashed border-border/40 gap-3">
                        <p className="text-sm text-muted-foreground">Sin certificado emitido.</p>
                        <Button size="sm" onClick={() => setCertDialogOpen(true)}>
                          <PlusIcon className="size-3.5 mr-1.5" />Emitir certificado
                        </Button>
                      </div>
                    )}
                  </TabsContent>
                </Tabs>
              </>
            )}
          </div>
        </div>
      </div>

      {/* CNRT dialog */}
      <Dialog open={cnrtDialogOpen} onOpenChange={(v) => !v && setCnrtDialogOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nueva inspección CNRT</DialogTitle></DialogHeader>
          <CnrtInspectionForm
            onSubmit={(data) => createCnrtMutation.mutate(data)}
            isPending={createCnrtMutation.isPending}
            onClose={() => setCnrtDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Certificate dialog */}
      <Dialog open={certDialogOpen} onOpenChange={(v) => !v && setCertDialogOpen(false)}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Nuevo certificado</DialogTitle></DialogHeader>
          <CertificateForm
            onSubmit={(data) => createCertMutation.mutate(data)}
            isPending={createCertMutation.isPending}
            onClose={() => setCertDialogOpen(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CnrtInspectionForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: {
    inspection_type: CnrtInspectionType;
    inspector_name: string;
    inspection_date: string;
    observations: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [inspectionType, setInspectionType] = useState<CnrtInspectionType>("initial");
  const [inspectorName, setInspectorName] = useState("");
  const [inspectionDate, setInspectionDate] = useState(today);
  const [observations, setObservations] = useState("");

  const canSubmit = !!inspectorName && !!inspectionDate;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!canSubmit) return;
        onSubmit({ inspection_type: inspectionType, inspector_name: inspectorName, inspection_date: inspectionDate, observations });
      }}
      className="space-y-4"
    >
      <div className="space-y-2">
        <Label>Tipo de inspección</Label>
        <Select value={inspectionType} onValueChange={(v) => setInspectionType(v as CnrtInspectionType)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            {cnrtInspectionTypeOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Inspector <span className="text-destructive">*</span></Label>
        <Input value={inspectorName} onChange={(e) => setInspectorName(e.target.value)} placeholder="Nombre del inspector" />
      </div>
      <div className="space-y-2">
        <Label>Fecha <span className="text-destructive">*</span></Label>
        <Input type="date" value={inspectionDate} onChange={(e) => setInspectionDate(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label>Observaciones</Label>
        <Textarea value={observations} onChange={(e) => setObservations(e.target.value)} rows={2} placeholder="Observaciones..." />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Registrando..." : "Registrar"}</Button>
      </div>
    </form>
  );
}

function CertificateForm({
  onSubmit,
  isPending,
  onClose,
}: {
  onSubmit: (data: {
    certificate_number: string;
    cert_type: string;
    authority: string;
    valid_from: string;
    valid_until: string;
    observations: string;
  }) => void;
  isPending: boolean;
  onClose: () => void;
}) {
  const today = new Date().toISOString().split("T")[0];
  const [certificateNumber, setCertificateNumber] = useState("");
  const [certType, setCertType] = useState("");
  const [authority, setAuthority] = useState("");
  const [validFrom, setValidFrom] = useState(today);
  const [validUntil, setValidUntil] = useState("");
  const [observations, setObservations] = useState("");

  const canSubmit = !!certificateNumber && !!certType && !!authority && !!validFrom;

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!canSubmit) return;
        onSubmit({ certificate_number: certificateNumber, cert_type: certType, authority, valid_from: validFrom, valid_until: validUntil, observations });
      }}
      className="space-y-4"
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Nro certificado <span className="text-destructive">*</span></Label>
          <Input value={certificateNumber} onChange={(e) => setCertificateNumber(e.target.value)} placeholder="CERT-001" />
        </div>
        <div className="space-y-2">
          <Label>Tipo <span className="text-destructive">*</span></Label>
          <Input value={certType} onChange={(e) => setCertType(e.target.value)} placeholder="Fabricación" />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Autoridad emisora <span className="text-destructive">*</span></Label>
        <Input value={authority} onChange={(e) => setAuthority(e.target.value)} placeholder="CNRT / Ministerio..." />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>Válido desde <span className="text-destructive">*</span></Label>
          <Input type="date" value={validFrom} onChange={(e) => setValidFrom(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Válido hasta <span className="text-muted-foreground">(opcional)</span></Label>
          <Input type="date" value={validUntil} onChange={(e) => setValidUntil(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Observaciones</Label>
        <Textarea value={observations} onChange={(e) => setObservations(e.target.value)} rows={2} placeholder="Observaciones..." />
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>Cancelar</Button>
        <Button type="submit" disabled={!canSubmit || isPending}>{isPending ? "Creando..." : "Crear"}</Button>
      </div>
    </form>
  );
}
