/**
 * Module guard unit tests.
 *
 * Tests pure logic from the modules system. Since @testing-library/react is
 * not installed, we test the data-layer logic (useHasModule's underlying
 * check) rather than the React component rendering.
 *
 * ModuleGuard component is not tested here — it requires a React render
 * environment with QueryClientProvider. Those would be Playwright e2e tests.
 */

import { describe, it, expect } from "bun:test";
import type { EnabledModule } from "@/lib/modules/hooks";
import { MODULE_REGISTRY } from "@/lib/modules/registry";

/**
 * Pure function equivalent of useHasModule's logic.
 * Extracted here to test without React Query / hooks.
 */
function hasModule(modules: EnabledModule[] | undefined, moduleId: string): boolean {
  if (!modules) return false;
  return modules.some((m) => m.id === moduleId);
}

/**
 * Pure function equivalent of ModuleGuard's allow logic.
 * Only enforces guard when isSuccess=true AND modules is non-empty.
 */
function guardAllows(
  modules: EnabledModule[] | undefined,
  isSuccess: boolean,
  moduleId: string,
): boolean {
  if (isSuccess && modules && modules.length > 0) {
    return modules.some((m) => m.id === moduleId);
  }
  // Fail-open: loading, API offline, or no modules configured
  return true;
}

const mockModules: EnabledModule[] = [
  { id: "fleet", name: "Fleet Management", category: "transport" },
  { id: "chat", name: "Chat", category: "core" },
  { id: "rag", name: "RAG", category: "core" },
];

// ---------------------------------------------------------------------------
// hasModule — pure logic
// ---------------------------------------------------------------------------

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

  it("returns false for an empty string module ID", () => {
    expect(hasModule(mockModules, "")).toBe(false);
  });

  it("returns false for unknown module not in registry or list", () => {
    expect(hasModule(mockModules, "does-not-exist-xyz")).toBe(false);
  });

  it("returns false for module the user doesn't have even if it exists in registry", () => {
    // rrhh is in MODULE_REGISTRY but not in mockModules
    expect(hasModule(mockModules, "rrhh")).toBe(false);
  });

  it("is case-sensitive — 'Fleet' does not match 'fleet'", () => {
    // IDs are lowercase in the system; uppercase lookups must fail
    expect(hasModule(mockModules, "Fleet")).toBe(false);
    expect(hasModule(mockModules, "FLEET")).toBe(false);
    expect(hasModule(mockModules, "fleet")).toBe(true);
  });

  it("is case-sensitive — 'Chat' does not match 'chat'", () => {
    expect(hasModule(mockModules, "Chat")).toBe(false);
    expect(hasModule(mockModules, "CHAT")).toBe(false);
  });

  it("handles a single-module list correctly", () => {
    const single: EnabledModule[] = [{ id: "solo", name: "Solo", category: "intelligence" }];
    expect(hasModule(single, "solo")).toBe(true);
    expect(hasModule(single, "fleet")).toBe(false);
  });

  it("handles a large module list without throwing", () => {
    const large: EnabledModule[] = Array.from({ length: 100 }, (_, i) => ({
      id: `module-${i}`,
      name: `Module ${i}`,
      category: "test",
    }));
    expect(hasModule(large, "module-50")).toBe(true);
    expect(hasModule(large, "module-99")).toBe(true);
    expect(hasModule(large, "module-100")).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// ModuleGuard fail-open behavior
// ---------------------------------------------------------------------------

describe("ModuleGuard fail-open logic", () => {
  it("allows access when isSuccess=false (still loading)", () => {
    // API hasn't returned yet — fail-open so the page renders
    expect(guardAllows(undefined, false, "fleet")).toBe(true);
  });

  it("allows access when modules list is undefined and isSuccess=true", () => {
    // isSuccess=true but modules is undefined (edge case in query state)
    expect(guardAllows(undefined, true, "fleet")).toBe(true);
  });

  it("allows access when modules list is empty and isSuccess=true", () => {
    // Empty list → fail-open (no modules configured yet, don't lock users out)
    expect(guardAllows([], true, "fleet")).toBe(true);
  });

  it("blocks access to unlisted module when isSuccess=true and list is non-empty", () => {
    expect(guardAllows(mockModules, true, "billing")).toBe(false);
  });

  it("allows access to listed module when isSuccess=true", () => {
    expect(guardAllows(mockModules, true, "fleet")).toBe(true);
  });

  it("blocks unknown module name without throwing", () => {
    expect(guardAllows(mockModules, true, "totally-unknown-module")).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// MODULE_REGISTRY — structural integrity
// ---------------------------------------------------------------------------

describe("MODULE_REGISTRY", () => {
  it("every entry has a valid id that matches its key", () => {
    for (const [key, manifest] of Object.entries(MODULE_REGISTRY)) {
      expect(manifest.id).toBe(key);
    }
  });

  it("every entry has a nav with label, path, and position", () => {
    for (const [key, manifest] of Object.entries(MODULE_REGISTRY)) {
      expect(typeof manifest.nav.label).toBe("string");
      expect(manifest.nav.label.length).toBeGreaterThan(0);
      expect(typeof manifest.nav.path).toBe("string");
      expect(manifest.nav.path.startsWith("/")).toBe(true);
      expect(typeof manifest.nav.position).toBe("number");
    }
  });

  it("every entry has at least one route", () => {
    for (const [key, manifest] of Object.entries(MODULE_REGISTRY)) {
      expect(manifest.routes.length).toBeGreaterThan(0);
      // First route should match the nav path
      expect(manifest.routes[0]).toBe(manifest.nav.path);
    }
  });

  it("no two entries share the same nav position", () => {
    const positions = Object.values(MODULE_REGISTRY).map((m) => m.nav.position);
    const unique = new Set(positions);
    expect(unique.size).toBe(positions.length);
  });

  it("known modules are present", () => {
    expect("fleet" in MODULE_REGISTRY || "produccion" in MODULE_REGISTRY).toBe(true);
    expect("feedback" in MODULE_REGISTRY).toBe(true);
  });
});
