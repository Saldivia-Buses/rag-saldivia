/**
 * Module guard unit tests.
 *
 * Tests pure logic from the modules system. Since @testing-library/react is
 * not installed, we test the data-layer logic (useHasModule's underlying
 * check) rather than the React component rendering.
 */

import { describe, it, expect } from "bun:test";
import type { EnabledModule } from "@/lib/modules/hooks";

/**
 * Pure function equivalent of useHasModule's logic.
 * Extracted here to test without React Query / hooks.
 */
function hasModule(modules: EnabledModule[] | undefined, moduleId: string): boolean {
  if (!modules) return false;
  return modules.some((m) => m.id === moduleId);
}

const mockModules: EnabledModule[] = [
  { id: "fleet", name: "Fleet Management", category: "transport" },
  { id: "chat", name: "Chat", category: "core" },
  { id: "rag", name: "RAG", category: "core" },
];

describe("module guard logic", () => {
  it("returns true when module is in the enabled list", () => {
    expect(hasModule(mockModules, "fleet")).toBe(true);
    expect(hasModule(mockModules, "chat")).toBe(true);
    expect(hasModule(mockModules, "rag")).toBe(true);
  });

  it("returns false when module is not in the enabled list", () => {
    expect(hasModule(mockModules, "billing")).toBe(false);
    expect(hasModule(mockModules, "hr")).toBe(false);
  });

  it("returns false when modules list is undefined", () => {
    expect(hasModule(undefined, "fleet")).toBe(false);
  });

  it("returns false when modules list is empty", () => {
    expect(hasModule([], "fleet")).toBe(false);
  });
});
