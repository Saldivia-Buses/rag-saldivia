"use client"

import { forwardRef, type HTMLAttributes } from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

// ---------------------------------------------------------------------------
// Color palette (from fluid) — used with `color` prop for semantic badges
// ---------------------------------------------------------------------------

const badgeColors = {
  gray: "#a3a3a3",
  red: "#ef4444",
  orange: "#f97316",
  amber: "#f59e0b",
  yellow: "#eab308",
  lime: "#84cc16",
  green: "#22c55e",
  emerald: "#10b981",
  teal: "#14b8a6",
  cyan: "#06b6d4",
  blue: "#3b82f6",
  indigo: "#6366f1",
  violet: "#8b5cf6",
  purple: "#a855f7",
  fuchsia: "#d946ef",
  pink: "#ec4899",
  rose: "#f43f5e",
} as const

type BadgeColor = keyof typeof badgeColors

// ---------------------------------------------------------------------------
// Variants (shadcnblocks base + fluid additions)
// ---------------------------------------------------------------------------

const badgeVariants = cva(
  "group/badge inline-flex w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-4xl border border-transparent px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-all focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 [&>svg]:pointer-events-none [&>svg]:size-3!",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground [a]:hover:bg-primary/80",
        secondary:
          "bg-secondary text-secondary-foreground [a]:hover:bg-secondary/80",
        destructive:
          "bg-destructive/10 text-destructive focus-visible:ring-destructive/20 dark:bg-destructive/20 dark:focus-visible:ring-destructive/40 [a]:hover:bg-destructive/20",
        outline:
          "border-border text-foreground [a]:hover:bg-muted [a]:hover:text-muted-foreground",
        ghost:
          "hover:bg-muted hover:text-muted-foreground dark:hover:bg-muted/50",
        link: "text-primary underline-offset-4 hover:underline",
        // Fluid-inspired: colored with dot indicator
        dot: "border border-border text-foreground",
      },
      size: {
        default: "h-5 px-2",
        sm: "h-4 px-1.5 text-[10px] gap-0.5",
        lg: "h-6 px-2.5 text-[13px] gap-1.5",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

interface BadgeProps
  extends Omit<HTMLAttributes<HTMLSpanElement>, "color">,
    VariantProps<typeof badgeVariants> {
  color?: BadgeColor
  prefixIcon?: React.ReactNode
  suffixIcon?: React.ReactNode
}

const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  (
    {
      className,
      variant = "default",
      size = "default",
      color,
      prefixIcon,
      suffixIcon,
      children,
      style,
      ...props
    },
    ref
  ) => {
    const isDot = variant === "dot"

    // Color-aware styles: when `color` is provided, override bg for solid variants
    const colorStyle =
      color && !isDot
        ? color === "gray"
          ? { backgroundColor: "var(--accent)", color: "var(--foreground)" }
          : {
              color: "var(--foreground)",
              backgroundColor: `color-mix(in srgb, ${badgeColors[color]} 15%, var(--background))`,
            }
        : {}

    const dotColor = color
      ? color === "gray"
        ? "var(--muted-foreground)"
        : badgeColors[color]
      : "var(--muted-foreground)"

    const dotSize = size === "sm" ? 5 : size === "lg" ? 7 : 6

    return (
      <span
        ref={ref}
        data-slot="badge"
        className={cn(badgeVariants({ variant, size }), className)}
        style={{ ...colorStyle, ...style }}
        {...props}
      >
        {isDot && (
          <span
            className="shrink-0 rounded-full"
            style={{
              width: dotSize,
              height: dotSize,
              backgroundColor: dotColor,
            }}
          />
        )}
        {prefixIcon}
        {children}
        {suffixIcon}
      </span>
    )
  }
)

Badge.displayName = "Badge"

export { Badge, badgeVariants, badgeColors }
export type { BadgeProps, BadgeColor }
