import type { Meta, StoryObj } from "@storybook/react"
import {
  Table, TableBody, TableCell, TableHead,
  TableHeader, TableRow, TableCaption,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"

const meta: Meta = {
  title: "Primitivos/Table",
  tags: ["autodocs"],
  parameters: { layout: "padded" },
}
export default meta

const users = [
  { email: "admin@saldivia.com",   name: "Admin",      role: "admin",  status: "Activo" },
  { email: "juan.p@saldivia.com",  name: "Juan Pérez", role: "user",   status: "Activo" },
  { email: "maria.g@saldivia.com", name: "María G.",   role: "editor", status: "Inactivo" },
  { email: "carlos.m@saldivia.com",name: "Carlos M.",  role: "user",   status: "Pendiente" },
  { email: "laura.v@saldivia.com", name: "Laura V.",   role: "admin",  status: "Activo" },
]

export const Default: StoryObj = {
  render: () => (
    <div className="rounded-lg border border-border overflow-hidden w-full max-w-2xl">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nombre</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Rol</TableHead>
            <TableHead>Estado</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {users.map((u) => (
            <TableRow key={u.email}>
              <TableCell className="font-medium">{u.name}</TableCell>
              <TableCell className="text-fg-muted">{u.email}</TableCell>
              <TableCell>
                <Badge variant={u.role === "admin" ? "default" : "secondary"}>
                  {u.role}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge variant={
                  u.status === "Activo" ? "success" :
                  u.status === "Pendiente" ? "warning" : "destructive"
                }>
                  {u.status}
                </Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  ),
}
