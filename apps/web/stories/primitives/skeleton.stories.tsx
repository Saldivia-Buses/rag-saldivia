import type { Meta, StoryObj } from "@storybook/react"
import { Skeleton, SkeletonText, SkeletonAvatar, SkeletonCard, SkeletonTable } from "@/components/ui/skeleton"

const meta: Meta = {
  title: "Primitivos/Skeleton",
  tags: ["autodocs"],
  parameters: { layout: "padded" },
}
export default meta

export const AllVariants: StoryObj = {
  render: () => (
    <div className="space-y-8 max-w-lg p-4">
      <div>
        <p className="text-xs text-fg-muted mb-2 uppercase tracking-wide">Text</p>
        <SkeletonText lines={3} />
      </div>
      <div>
        <p className="text-xs text-fg-muted mb-2 uppercase tracking-wide">Avatar</p>
        <div className="flex gap-3">
          <SkeletonAvatar size="sm" />
          <SkeletonAvatar size="md" />
          <SkeletonAvatar size="lg" />
        </div>
      </div>
      <div>
        <p className="text-xs text-fg-muted mb-2 uppercase tracking-wide">Card</p>
        <SkeletonCard />
      </div>
      <div>
        <p className="text-xs text-fg-muted mb-2 uppercase tracking-wide">Table</p>
        <div className="rounded-lg border border-border overflow-hidden">
          <SkeletonTable rows={4} cols={3} />
        </div>
      </div>
    </div>
  ),
}
