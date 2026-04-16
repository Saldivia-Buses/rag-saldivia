/**
 * ERP format utility unit tests.
 *
 * Tests pure formatter functions from lib/erp/format.ts:
 * - fmtMoney: ARS currency formatting, null/zero handling
 * - fmtNumber: numeric formatting with separators
 * - fmtDate: full date (dd/mm/yyyy)
 * - fmtDateShort: short date (dd mmm)
 * - fmtPercent: percentage with one decimal
 *
 * All formatters use es-AR locale. Tests run in bun (no DOM needed).
 */

import { describe, it, expect } from "bun:test";
import {
  fmtMoney,
  fmtNumber,
  fmtDate,
  fmtDateShort,
  fmtPercent,
} from "@/lib/erp/format";

const EM_DASH = "\u2014";

describe("fmtMoney", () => {
  it("returns em-dash for null", () => {
    expect(fmtMoney(null)).toBe(EM_DASH);
  });

  it("returns em-dash for undefined", () => {
    expect(fmtMoney(undefined)).toBe(EM_DASH);
  });

  it("returns em-dash for zero", () => {
    // Zero is treated as empty/no value in the ERP context
    expect(fmtMoney(0)).toBe(EM_DASH);
  });

  it("formats a positive integer", () => {
    const result = fmtMoney(1000);
    // Should contain the numeric value — locale formats vary across runtimes,
    // so check structure rather than exact string
    expect(result).toContain("1");
    expect(result).not.toBe(EM_DASH);
  });

  it("formats a large number with separators", () => {
    const result = fmtMoney(1234567);
    expect(result).toContain("1");
    expect(result).toContain("234");
    expect(result).not.toBe(EM_DASH);
  });

  it("formats a negative number", () => {
    const result = fmtMoney(-500);
    expect(result).toContain("500");
    expect(result).not.toBe(EM_DASH);
  });

  it("truncates fractional digits (maximumFractionDigits: 0)", () => {
    const result = fmtMoney(1234.89);
    // Should NOT contain ".89" — currency formatter rounds to 0 decimals
    expect(result).not.toContain(".89");
    expect(result).not.toContain(",89");
  });

  it("does not throw for very large numbers", () => {
    expect(() => fmtMoney(999_999_999)).not.toThrow();
  });
});

describe("fmtNumber", () => {
  it("returns em-dash for null", () => {
    expect(fmtNumber(null)).toBe(EM_DASH);
  });

  it("returns em-dash for undefined", () => {
    expect(fmtNumber(undefined)).toBe(EM_DASH);
  });

  it("formats zero as '0'", () => {
    expect(fmtNumber(0)).toBe("0");
  });

  it("formats a whole number", () => {
    const result = fmtNumber(1000);
    expect(result).toContain("1");
    expect(result).not.toBe(EM_DASH);
  });

  it("formats a decimal number (up to 2 decimals)", () => {
    const result = fmtNumber(1234.5);
    expect(result).toContain("1");
    expect(result).not.toBe(EM_DASH);
  });

  it("does not show more than 2 decimal digits", () => {
    const result = fmtNumber(1.999);
    // maximumFractionDigits: 2 — rounds to at most 2 places
    // "2" in es-AR is just "2"
    expect(result).toBe("2");
  });

  it("formats a negative number", () => {
    const result = fmtNumber(-42);
    expect(result).toContain("42");
  });
});

describe("fmtDate", () => {
  it("returns em-dash for null", () => {
    expect(fmtDate(null)).toBe(EM_DASH);
  });

  it("returns em-dash for undefined", () => {
    expect(fmtDate(undefined)).toBe(EM_DASH);
  });

  it("returns em-dash for empty string", () => {
    expect(fmtDate("")).toBe(EM_DASH);
  });

  it("formats a valid ISO date string", () => {
    // 2024-01-15 should produce something with 15, 01, 2024 in es-AR dd/mm/yyyy
    const result = fmtDate("2024-01-15");
    expect(result).toContain("2024");
    expect(result).toContain("15");
    expect(result).not.toBe(EM_DASH);
  });

  it("formats a date-time string (uses date part)", () => {
    const result = fmtDate("2024-06-30T12:00:00Z");
    expect(result).toContain("2024");
    expect(result).not.toBe(EM_DASH);
  });

  it("throws RangeError for an invalid date string", () => {
    // new Date("not-a-date") → Invalid Date — Intl.DateTimeFormat.format() throws
    // RangeError for Invalid Date. Callers must pass valid ISO strings.
    expect(() => fmtDate("not-a-date")).toThrow(RangeError);
  });
});

describe("fmtDateShort", () => {
  it("returns em-dash for null", () => {
    expect(fmtDateShort(null)).toBe(EM_DASH);
  });

  it("returns em-dash for undefined", () => {
    expect(fmtDateShort(undefined)).toBe(EM_DASH);
  });

  it("returns em-dash for empty string", () => {
    expect(fmtDateShort("")).toBe(EM_DASH);
  });

  it("formats a valid date to short form (dd mmm)", () => {
    const result = fmtDateShort("2024-03-21");
    // Should contain "21" and some abbreviated month
    expect(result).toContain("21");
    expect(result).not.toBe(EM_DASH);
    // Should NOT contain the full year
    expect(result).not.toContain("2024");
  });

  it("throws RangeError for an invalid date string", () => {
    expect(() => fmtDateShort("not-a-date")).toThrow(RangeError);
  });
});

describe("fmtPercent", () => {
  it("returns em-dash for null", () => {
    expect(fmtPercent(null)).toBe(EM_DASH);
  });

  it("returns em-dash for undefined", () => {
    expect(fmtPercent(undefined)).toBe(EM_DASH);
  });

  it("formats zero as '0.0%'", () => {
    expect(fmtPercent(0)).toBe("0.0%");
  });

  it("formats a whole number with one decimal", () => {
    expect(fmtPercent(50)).toBe("50.0%");
  });

  it("formats a decimal to one place", () => {
    expect(fmtPercent(12.34)).toBe("12.3%");
  });

  it("formats a negative percentage", () => {
    expect(fmtPercent(-5.5)).toBe("-5.5%");
  });

  it("includes the percent sign", () => {
    const result = fmtPercent(75);
    expect(result.endsWith("%")).toBe(true);
  });
});
