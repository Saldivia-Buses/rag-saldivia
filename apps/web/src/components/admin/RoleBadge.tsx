/**
 * Role badge with dynamic color from the role's color field.
 *
 * Renders a small colored badge with the role name.
 * Used in: AdminUsers (multi-role display), AdminRoles (role cards)
 */

export function RoleBadge({
  name,
  color,
  size = "sm",
}: {
  name: string
  color: string
  size?: "sm" | "xs"
}) {
  return (
    <span
      className="inline-flex items-center rounded-md font-medium"
      style={{
        padding: size === "xs" ? "1px 6px" : "2px 8px",
        fontSize: size === "xs" ? "10px" : "12px",
        color,
        backgroundColor: `color-mix(in srgb, ${color} 12%, transparent)`,
        border: `1px solid color-mix(in srgb, ${color} 25%, transparent)`,
      }}
    >
      {name}
    </span>
  )
}
