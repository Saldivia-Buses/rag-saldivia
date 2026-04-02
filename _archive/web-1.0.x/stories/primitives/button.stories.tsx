import type { Meta, StoryObj } from "@storybook/react"
import { Button } from "@/components/ui/button"
import { Mail, Loader2, Trash2 } from "lucide-react"

const meta: Meta<typeof Button> = {
  title: "Primitivos/Button",
  component: Button,
  tags: ["autodocs"],
  parameters: { layout: "centered" },
  argTypes: {
    variant: {
      control: "select",
      options: ["default", "destructive", "outline", "secondary", "ghost", "link"],
    },
    size: {
      control: "select",
      options: ["default", "sm", "lg", "icon"],
    },
  },
}
export default meta
type Story = StoryObj<typeof Button>

export const Default: Story = {
  args: { children: "Iniciar sesión", variant: "default" },
}

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-3 items-center p-4">
      <Button variant="default">Default</Button>
      <Button variant="destructive">Destructive</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="link">Link</Button>
    </div>
  ),
}

export const AllSizes: Story = {
  render: () => (
    <div className="flex gap-3 items-center p-4">
      <Button size="lg">Large</Button>
      <Button size="default">Default</Button>
      <Button size="sm">Small</Button>
      <Button size="icon"><Mail /></Button>
    </div>
  ),
}

export const WithIcon: Story = {
  render: () => (
    <div className="flex gap-3 p-4">
      <Button><Mail className="h-4 w-4" /> Enviar email</Button>
      <Button variant="destructive"><Trash2 className="h-4 w-4" /> Eliminar</Button>
    </div>
  ),
}

export const Loading: Story = {
  render: () => (
    <Button disabled>
      <Loader2 className="h-4 w-4 animate-spin" /> Cargando...
    </Button>
  ),
}

export const Disabled: Story = {
  args: { children: "Deshabilitado", disabled: true },
}
