"use client";

import { ModuleGuard } from "@/lib/modules/guard";

export default function Layout({ children }: { children: React.ReactNode }) {
  return <ModuleGuard moduleId="rrhh">{children}</ModuleGuard>;
}
