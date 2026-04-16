"use client";

import { ModuleGuard } from "@/lib/modules/guard";

export default function ManufacturaLayout({ children }: { children: React.ReactNode }) {
  return <ModuleGuard moduleId="manufactura">{children}</ModuleGuard>;
}
