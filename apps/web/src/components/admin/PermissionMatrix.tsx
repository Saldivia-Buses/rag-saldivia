/**
 * Permission matrix — grid of checkboxes grouped by category.
 *
 * Each checkbox toggles a permission for the selected role.
 * Changes are applied immediately via actionSetRolePermissions.
 * Shows active/total count per category.
 *
 * Used by: AdminRoles (when a role is selected for editing)
 */

"use client"

import { useTransition } from "react"
import { Check } from "lucide-react"
import { actionSetRolePermissions } from "@/app/actions/roles"

type Permission = {
  id: number
  key: string
  label: string
  category: string
  description: string
}

export function PermissionMatrix({
  roleId,
  roleName,
  isSystem: _isSystem,
  locked = false,
  allPermissions,
  activeKeys,
  onUpdate,
}: {
  roleId: number
  roleName: string
  isSystem: boolean
  locked?: boolean
  allPermissions: Permission[]
  activeKeys: string[]
  onUpdate: (newKeys: string[]) => void
}) {
  const [isPending, startTransition] = useTransition()

  // Group by category
  const categories = new Map<string, Permission[]>()
  for (const p of allPermissions) {
    const list = categories.get(p.category) ?? []
    list.push(p)
    categories.set(p.category, list)
  }

  function handleToggle(key: string, checked: boolean) {
    const newKeys = checked
      ? [...activeKeys, key]
      : activeKeys.filter((k) => k !== key)

    onUpdate(newKeys)

    startTransition(async () => {
      try {
        await actionSetRolePermissions({ roleId, permissionKeys: newKeys })
      } catch {
        onUpdate(activeKeys)
      }
    })
  }

  function handleToggleCategory(category: string, perms: Permission[]) {
    const allActive = perms.every((p) => activeKeys.includes(p.key))
    const newKeys = allActive
      ? activeKeys.filter((k) => !perms.some((p) => p.key === k))
      : [...new Set([...activeKeys, ...perms.map((p) => p.key)])]

    onUpdate(newKeys)

    startTransition(async () => {
      try {
        await actionSetRolePermissions({ roleId, permissionKeys: newKeys })
      } catch {
        onUpdate(activeKeys)
      }
    })
  }

  const totalActive = activeKeys.length
  const totalPerms = allPermissions.length

  return (
    <div
      className="border border-border rounded-xl bg-surface overflow-hidden"
    >
      {/* Header */}
      <div
        className="flex items-center justify-between border-b border-border"
        style={{ padding: "16px 20px" }}
      >
        <div>
          <h3 className="text-sm font-semibold text-fg">
            Permisos de {roleName}
            {locked && <span className="text-xs font-normal text-fg-subtle" style={{ marginLeft: "8px" }}>(protegido)</span>}
          </h3>
          <p className="text-xs text-fg-subtle" style={{ marginTop: "2px" }}>
            {totalActive} de {totalPerms} permisos activos
            {locked && " — el rol Admin siempre tiene todos los permisos"}
            {isPending && <span style={{ marginLeft: "6px" }}>— guardando...</span>}
          </p>
        </div>
        <div
          className="text-xs font-medium rounded-full"
          style={{
            padding: "4px 12px",
            color: totalActive === totalPerms ? "var(--success)" : "var(--fg-muted)",
            background: totalActive === totalPerms
              ? "color-mix(in srgb, var(--success) 10%, transparent)"
              : "var(--surface-2)",
          }}
        >
          {totalActive}/{totalPerms}
        </div>
      </div>

      {/* Categories grid */}
      <div
        className="grid grid-cols-2 lg:grid-cols-3"
        style={{ gap: "0" }}
      >
        {Array.from(categories.entries()).map(([category, perms]) => {
          const activeInCat = perms.filter((p) => activeKeys.includes(p.key)).length
          const allActive = activeInCat === perms.length

          return (
            <div
              key={category}
              className="border-b border-r border-border last:border-r-0"
              style={{ padding: "16px 20px" }}
            >
              {/* Category header — click to toggle all */}
              <button
                type="button"
                onClick={() => !locked && handleToggleCategory(category, perms)}
                disabled={locked}
                className={`flex items-center justify-between w-full group ${locked ? "cursor-default" : ""}`}
                style={{ marginBottom: "10px" }}
              >
                <span className="text-xs font-semibold text-fg-muted uppercase tracking-wide group-hover:text-fg transition-colors">
                  {category}
                </span>
                <span
                  className="text-[10px] font-medium rounded-full"
                  style={{
                    padding: "1px 6px",
                    color: allActive ? "var(--success)" : "var(--fg-subtle)",
                    background: allActive
                      ? "color-mix(in srgb, var(--success) 10%, transparent)"
                      : "var(--surface-2)",
                  }}
                >
                  {activeInCat}/{perms.length}
                </span>
              </button>

              {/* Permission checkboxes */}
              <div className="flex flex-col" style={{ gap: "4px" }}>
                {perms.map((p) => {
                  const isActive = activeKeys.includes(p.key)
                  return (
                    <label
                      key={p.key}
                      className={`flex items-center text-sm rounded-md transition-colors group ${locked ? "cursor-default opacity-70" : "cursor-pointer hover:bg-surface-2"}`}
                      style={{ gap: "8px", padding: "5px 6px", margin: "0 -6px" }}
                      title={p.description}
                    >
                      {/* Custom checkbox */}
                      <div
                        className="flex items-center justify-center rounded shrink-0 transition-colors"
                        style={{
                          width: "16px",
                          height: "16px",
                          border: isActive ? "none" : "1.5px solid var(--border)",
                          background: isActive ? "var(--accent)" : "transparent",
                        }}
                      >
                        {isActive && <Check size={11} style={{ color: "white" }} strokeWidth={2.5} />}
                      </div>
                      <input
                        type="checkbox"
                        checked={isActive}
                        onChange={(e) => !locked && handleToggle(p.key, e.target.checked)}
                        disabled={locked}
                        className="sr-only"
                      />
                      <span className={isActive ? "text-fg" : "text-fg-muted"}>
                        {p.label}
                      </span>
                    </label>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
