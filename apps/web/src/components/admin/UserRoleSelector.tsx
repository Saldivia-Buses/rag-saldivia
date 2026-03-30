/**
 * Multi-select dropdown for assigning roles to a user.
 *
 * Shows all available roles as checkboxes. Changes are applied
 * immediately via actionSetUserRoles.
 *
 * Used by: AdminUsers (role column)
 */

"use client"

import { useState, useRef, useEffect, useTransition } from "react"
import { ChevronDown } from "lucide-react"
import { RoleBadge } from "./RoleBadge"
import { actionSetUserRoles } from "@/app/actions/roles"

type Role = {
  id: number
  name: string
  color: string
  level: number
}

export function UserRoleSelector({
  userId,
  allRoles,
  assignedRoleIds,
  onUpdate,
  disabled,
}: {
  userId: number
  allRoles: Role[]
  assignedRoleIds: number[]
  onUpdate: (newRoleIds: number[]) => void
  disabled?: boolean
}) {
  const [open, setOpen] = useState(false)
  const [isPending, startTransition] = useTransition()
  const ref = useRef<HTMLDivElement>(null)

  // Close on outside click
  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClick)
    return () => document.removeEventListener("mousedown", handleClick)
  }, [open])

  const assignedRoles = allRoles
    .filter((r) => assignedRoleIds.includes(r.id))
    .sort((a, b) => b.level - a.level)

  function handleToggle(roleId: number, checked: boolean) {
    const newIds = checked
      ? [...assignedRoleIds, roleId]
      : assignedRoleIds.filter((id) => id !== roleId)

    // Must have at least one role
    if (newIds.length === 0) return

    onUpdate(newIds)

    startTransition(async () => {
      try {
        await actionSetUserRoles({ userId, roleIds: newIds })
      } catch {
        onUpdate(assignedRoleIds)
      }
    })
  }

  return (
    <div ref={ref} className="relative">
      {/* Badges + trigger */}
      <button
        type="button"
        onClick={() => !disabled && setOpen(!open)}
        disabled={disabled}
        className="flex items-center flex-wrap cursor-pointer disabled:cursor-not-allowed disabled:opacity-50"
        style={{ gap: "4px" }}
      >
        {assignedRoles.map((r) => (
          <RoleBadge key={r.id} name={r.name} color={r.color} size="xs" />
        ))}
        {!disabled && (
          <ChevronDown
            size={12}
            className="text-fg-subtle"
            style={{ marginLeft: "2px" }}
          />
        )}
      </button>

      {/* Dropdown */}
      {open && (
        <div
          className="absolute left-0 z-50 border border-border rounded-lg bg-surface shadow-lg"
          style={{ top: "calc(100% + 4px)", minWidth: "180px", padding: "6px" }}
        >
          {allRoles
            .sort((a, b) => b.level - a.level)
            .map((role) => (
              <label
                key={role.id}
                className="flex items-center text-sm text-fg cursor-pointer hover:bg-surface-2 rounded-md transition-colors"
                style={{ gap: "8px", padding: "6px 8px" }}
              >
                <input
                  type="checkbox"
                  checked={assignedRoleIds.includes(role.id)}
                  onChange={(e) => handleToggle(role.id, e.target.checked)}
                  className="accent-[var(--accent)]"
                  style={{ width: "14px", height: "14px" }}
                />
                <RoleBadge name={role.name} color={role.color} size="xs" />
              </label>
            ))}
          {isPending && (
            <div className="text-xs text-fg-subtle text-center" style={{ padding: "4px" }}>
              Guardando...
            </div>
          )}
        </div>
      )}
    </div>
  )
}
