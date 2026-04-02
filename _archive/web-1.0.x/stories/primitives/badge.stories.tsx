import type { Meta, StoryObj } from "@storybook/react"
import { Badge } from "@/components/ui/badge"

const meta: Meta<typeof Badge> = {
  title: "Primitivos/Badge",
  component: Badge,
  tags: ["autodocs"],
  parameters: { layout: "centered" },
}
export default meta
type Story = StoryObj<typeof Badge>

export const Default: Story = { args: { children: "Admin", variant: "default" } }

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2 p-4">
      <Badge variant="default">Default</Badge>
      <Badge variant="secondary">Secondary</Badge>
      <Badge variant="outline">Outline</Badge>
      <Badge variant="destructive">Error</Badge>
      <Badge variant="success">Completado</Badge>
      <Badge variant="warning">Pendiente</Badge>
    </div>
  ),
}

export const UseCases: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2 p-4">
      <Badge variant="default">admin</Badge>
      <Badge variant="secondary">user</Badge>
      <Badge variant="success">Activo</Badge>
      <Badge variant="destructive">Inactivo</Badge>
      <Badge variant="warning">Pendiente</Badge>
      <Badge variant="outline">área_legal</Badge>
    </div>
  ),
}
