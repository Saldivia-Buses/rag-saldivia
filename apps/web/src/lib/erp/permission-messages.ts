import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";

export const permissionMessages: Record<string, string> = {
  "erp.accounting.close": "No tenés permiso para cerrar el ejercicio",
  "erp.treasury.reconcile": "No tenés permiso para reconciliar la cuenta bancaria",
  "erp.invoicing.void": "No tenés permiso para anular el comprobante",
  "erp.purchasing.inspect": "No tenés permiso para inspeccionar la recepción",
  "erp.treasury.receipt": "No tenés permiso para registrar el recibo",
  "erp.stock.write": "No tenés permiso para modificar artículos",
  "erp.entities.write": "No tenés permiso para modificar entidades",
  "erp.catalogs.write": "No tenés permiso para modificar catálogos",
};

export function permissionErrorToast(err: unknown): void {
  if (err instanceof ApiError && err.status === 403) {
    const msg = "No tenés permiso para esta acción";
    toast.error(msg);
    return;
  }
  toast.error("Error inesperado", {
    description: err instanceof Error ? err.message : undefined,
  });
}
