import type { Meta, StoryObj } from "@storybook/react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

const meta: Meta = {
  title: "Primitivos/Avatar",
  tags: ["autodocs"],
  parameters: { layout: "centered" },
}
export default meta

export const WithFallback: StoryObj = {
  render: () => (
    <div className="flex gap-3 items-center p-4">
      <Avatar className="h-8 w-8"><AvatarFallback>EA</AvatarFallback></Avatar>
      <Avatar><AvatarFallback>RS</AvatarFallback></Avatar>
      <Avatar className="h-12 w-12"><AvatarFallback>JD</AvatarFallback></Avatar>
    </div>
  ),
}

export const WithImage: StoryObj = {
  render: () => (
    <Avatar>
      <AvatarImage src="https://github.com/shadcn.png" alt="shadcn" />
      <AvatarFallback>SC</AvatarFallback>
    </Avatar>
  ),
}
