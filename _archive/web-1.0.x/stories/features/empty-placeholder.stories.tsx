import type { Meta, StoryObj } from "@storybook/react"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { Button } from "@/components/ui/button"
import { MessageSquare, FolderOpen, Upload, Users } from "lucide-react"

const meta: Meta = {
  title: "Features/EmptyPlaceholder",
  tags: ["autodocs"],
  parameters: { layout: "padded" },
}
export default meta

export const Chat: StoryObj = {
  render: () => (
    <EmptyPlaceholder className="max-w-sm">
      <EmptyPlaceholder.Icon icon={MessageSquare} />
      <EmptyPlaceholder.Title>Sin conversaciones</EmptyPlaceholder.Title>
      <EmptyPlaceholder.Description>
        Hacé una pregunta sobre tus documentos y obtené respuestas fundamentadas.
      </EmptyPlaceholder.Description>
      <Button size="sm">Nueva conversación</Button>
    </EmptyPlaceholder>
  ),
}

export const Collections: StoryObj = {
  render: () => (
    <EmptyPlaceholder className="max-w-sm">
      <EmptyPlaceholder.Icon icon={FolderOpen} />
      <EmptyPlaceholder.Title>Sin colecciones</EmptyPlaceholder.Title>
      <EmptyPlaceholder.Description>
        Creá una colección para empezar a ingestar documentos.
      </EmptyPlaceholder.Description>
    </EmptyPlaceholder>
  ),
}

export const AllVariants: StoryObj = {
  render: () => (
    <div className="grid grid-cols-2 gap-4 p-4">
      <EmptyPlaceholder>
        <EmptyPlaceholder.Icon icon={MessageSquare} />
        <EmptyPlaceholder.Title>Sin mensajes</EmptyPlaceholder.Title>
        <EmptyPlaceholder.Description>Empezá una conversación.</EmptyPlaceholder.Description>
      </EmptyPlaceholder>
      <EmptyPlaceholder>
        <EmptyPlaceholder.Icon icon={Upload} />
        <EmptyPlaceholder.Title>Sin documentos</EmptyPlaceholder.Title>
        <EmptyPlaceholder.Description>Subí documentos a esta colección.</EmptyPlaceholder.Description>
        <Button size="sm" variant="outline">Subir documentos</Button>
      </EmptyPlaceholder>
    </div>
  ),
}
