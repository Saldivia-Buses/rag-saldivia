const moneyFormatter = new Intl.NumberFormat("es-AR", {
  style: "currency",
  currency: "ARS",
  maximumFractionDigits: 0,
});

const numberFormatter = new Intl.NumberFormat("es-AR", {
  maximumFractionDigits: 2,
});

const dateFormatter = new Intl.DateTimeFormat("es-AR", {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
});

const shortDateFormatter = new Intl.DateTimeFormat("es-AR", {
  day: "2-digit",
  month: "short",
});

export function fmtMoney(n: number | null | undefined): string {
  if (n == null || n === 0) return "\u2014";
  return moneyFormatter.format(n);
}

export function fmtDate(d: string | null | undefined): string {
  if (!d) return "\u2014";
  return dateFormatter.format(new Date(d));
}

export function fmtDateShort(d: string | null | undefined): string {
  if (!d) return "\u2014";
  return shortDateFormatter.format(new Date(d));
}

export function fmtNumber(n: number | null | undefined): string {
  if (n == null) return "\u2014";
  return numberFormatter.format(n);
}

export function fmtPercent(n: number | null | undefined): string {
  if (n == null) return "\u2014";
  return `${n.toFixed(1)}%`;
}
