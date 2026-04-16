"use client"

import { ToggleGroup as ToggleGroupPrimitive } from "@base-ui/react/toggle-group"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const toggleGroupVariants = cva(
  "flex w-fit items-stretch gap-0.5 *:focus-visible:relative *:focus-visible:z-10",
  {
    variants: {
      orientation: {
        horizontal: "flex-row",
        vertical: "flex-col",
      },
    },
    defaultVariants: {
      orientation: "horizontal",
    },
  }
)

interface ToggleGroupProps
  extends Omit<ToggleGroupPrimitive.Props, "orientation">,
    VariantProps<typeof toggleGroupVariants> {}

function ToggleGroup({
  className,
  orientation = "horizontal",
  ...props
}: ToggleGroupProps) {
  return (
    <ToggleGroupPrimitive
      data-slot="toggle-group"
      data-orientation={orientation}
      orientation={orientation ?? "horizontal"}
      className={cn(toggleGroupVariants({ orientation }), className)}
      {...props}
    />
  )
}

export { ToggleGroup, toggleGroupVariants }
export type { ToggleGroupProps }
