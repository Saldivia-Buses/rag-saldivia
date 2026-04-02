import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Formatea un timestamp como dd/mm/yyyy en zona horaria local argentina. */
export function formatDate(ts: number | string | Date): string {
  return new Date(ts).toLocaleDateString("es-AR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  })
}

/** Formatea un timestamp como dd/mm/yyyy hh:mm en zona horaria local argentina. */
export function formatDateTime(ts: number | string | Date): string {
  return new Date(ts).toLocaleString("es-AR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}
