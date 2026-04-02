import type { Meta, StoryObj } from "@storybook/react"
import { Input } from "@/components/ui/input"

const meta: Meta<typeof Input> = {
  title: "Primitivos/Input",
  component: Input,
  tags: ["autodocs"],
  parameters: { layout: "centered" },
}
export default meta
type Story = StoryObj<typeof Input>

export const Default: Story = {
  args: { placeholder: "usuario@empresa.com", type: "email" },
}

export const AllStates: Story = {
  render: () => (
    <div className="flex flex-col gap-3 w-72 p-4">
      <Input placeholder="Estado normal" />
      <Input defaultValue="Con valor" />
      <Input placeholder="Deshabilitado" disabled />
      <Input type="password" defaultValue="password123" />
    </div>
  ),
}
