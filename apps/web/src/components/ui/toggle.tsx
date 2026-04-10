"use client"

import { Toggle as TogglePrimitive } from "@base-ui/react/toggle"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const toggleVariants = cva(
  "group/toggle inline-flex shrink-0 items-center justify-center gap-1.5 rounded-lg border border-transparent text-sm font-medium whitespace-nowrap transition-all outline-none select-none cursor-pointer focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        default:
          "bg-transparent hover:bg-muted hover:text-foreground data-pressed:bg-muted data-pressed:text-foreground dark:hover:bg-muted/50 dark:data-pressed:bg-muted/50",
        outline:
          "border-border bg-transparent hover:bg-muted hover:text-foreground data-pressed:bg-accent data-pressed:text-accent-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
      },
      size: {
        default: "h-8 px-2.5",
        sm: "h-7 px-2 text-[0.8rem]",
        lg: "h-9 px-3",
        icon: "size-8",
        "icon-sm": "size-7",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

interface ToggleProps
  extends TogglePrimitive.Props,
    VariantProps<typeof toggleVariants> {
  label?: string
}

function Toggle({
  className,
  variant,
  size,
  label,
  children,
  ...props
}: ToggleProps) {
  return (
    <TogglePrimitive
      data-slot="toggle"
      className={cn(toggleVariants({ variant, size }), className)}
      {...props}
    >
      {children}
      {label && (
        <span className="text-sm">{label}</span>
      )}
    </TogglePrimitive>
  )
}

export { Toggle, toggleVariants }
export type { ToggleProps }
