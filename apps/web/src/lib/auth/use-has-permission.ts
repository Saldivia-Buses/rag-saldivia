import { useAuthStore } from "./store";

export function useHasPermission(perm: string): boolean {
  const perms = useAuthStore((s) => s.user?.perms);
  if (!perms) return false;
  return perms.includes(perm);
}

export function useHasAnyPermission(...perms: string[]): boolean {
  const userPerms = useAuthStore((s) => s.user?.perms);
  if (!userPerms) return false;
  return perms.some((p) => userPerms.includes(p));
}

export function useHasAllPermissions(...perms: string[]): boolean {
  const userPerms = useAuthStore((s) => s.user?.perms);
  if (!userPerms) return false;
  return perms.every((p) => userPerms.includes(p));
}
