import type { Meta, StoryObj } from "@storybook/react"
import { StatCard } from "@/components/ui/stat-card"
import { MessageSquare, Users, FolderOpen, Zap } from "lucide-react"

const meta: Meta = {
  title: "Features/StatCard",
  tags: ["autodocs"],
  parameters: { layout: "padded" },
}
export default meta

export const Default: StoryObj = {
  render: () => (
    <div className="grid grid-cols-2 gap-4 max-w-2xl p-4">
      <StatCard label="Queries totales" value="12,847" delta={18} deltaLabel="vs mes anterior" icon={MessageSquare} />
      <StatCard label="Usuarios activos" value="94" delta={-3} deltaLabel="vs mes anterior" icon={Users} />
      <StatCard label="Colecciones" value="7" delta={0} deltaLabel="sin cambios" icon={FolderOpen} />
      <StatCard label="Tiempo respuesta" value="1.2s" delta={12} deltaLabel="más rápido" icon={Zap} />
    </div>
  ),
}
