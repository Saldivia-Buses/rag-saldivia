/**
 * RBAC — Role Based Access Control
 * Reemplaza la lógica de permisos dispersa en gateway.py
 */

import type { JwtClaims, Role } from "@rag-saldivia/shared"

// Jerarquía de roles: admin > area_manager > user
const ROLE_LEVELS: Record<Role, number> = {
  admin: 3,
  area_manager: 2,
  user: 1,
}

export function hasRole(claims: JwtClaims, minRole: Role): boolean {
  return ROLE_LEVELS[claims.role] >= ROLE_LEVELS[minRole]
}

export function isAdmin(claims: JwtClaims): boolean {
  return claims.role === "admin"
}

export function isAreaManager(claims: JwtClaims): boolean {
  return claims.role === "admin" || claims.role === "area_manager"
}

// Rutas protegidas por rol mínimo requerido
const PROTECTED_ROUTES: Array<{ pattern: RegExp; minRole: Role }> = [
  { pattern: /^\/admin(\/|$)/, minRole: "admin" },
  { pattern: /^\/api\/admin(\/|$)/, minRole: "admin" },
  { pattern: /^\/audit(\/|$)/, minRole: "area_manager" },
  { pattern: /^\/api\/audit(\/|$)/, minRole: "area_manager" },
]

export function getRequiredRole(pathname: string): Role | null {
  for (const route of PROTECTED_ROUTES) {
    if (route.pattern.test(pathname)) {
      return route.minRole
    }
  }
  return null // público o solo requiere autenticación
}

export function canAccessRoute(claims: JwtClaims, pathname: string): boolean {
  const required = getRequiredRole(pathname)
  if (!required) return true
  return hasRole(claims, required)
}
