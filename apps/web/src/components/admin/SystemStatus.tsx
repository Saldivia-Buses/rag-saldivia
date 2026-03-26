"use client"

import { useRouter } from "next/navigation"
import { Users, Database, FolderOpen, AlertTriangle, RefreshCw, Loader2 } from "lucide-react"
import type { DbIngestionQueueItem } from "@rag-saldivia/db"
import { StatCard } from "@/components/ui/stat-card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table"

type Stats = {
  activeUsers: number
  areas: number
  collections: number
  activeJobs: number
  recentErrors: number
}

export function SystemStatus({
  stats,
  activeJobs,
}: {
  stats: Stats
  activeJobs: DbIngestionQueueItem[]
}) {
  const router = useRouter()

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Estado del sistema</h1>
          <p className="text-sm text-fg-muted mt-0.5">Métricas y jobs activos</p>
        </div>
        <Button size="sm" variant="outline" onClick={() => router.refresh()}>
          <RefreshCw className="h-3.5 w-3.5" />
          Actualizar
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Usuarios activos" value={stats.activeUsers} icon={Users} />
        <StatCard label="Áreas" value={stats.areas} icon={Database} />
        <StatCard label="Colecciones" value={stats.collections} icon={FolderOpen} />
        <StatCard
          label="Errores (24hs)"
          value={stats.recentErrors}
          icon={AlertTriangle}
          delta={stats.recentErrors > 0 ? -1 : 0}
        />
      </div>

      {/* Jobs activos */}
      <div>
        <h2 className="text-sm font-semibold text-fg mb-3">Jobs de ingesta activos</h2>
        {activeJobs.length === 0 ? (
          <p className="text-sm text-fg-muted">Sin jobs activos</p>
        ) : (
          <div className="rounded-xl border border-border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Colección</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead>Reintentos</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {activeJobs.map((job) => (
                  <TableRow key={job.id}>
                    <TableCell className="font-mono text-xs text-fg-muted">{job.id.slice(0, 8)}</TableCell>
                    <TableCell className="text-fg">{job.collection}</TableCell>
                    <TableCell>
                      <Badge variant={job.status === "locked" ? "default" : "secondary"}>
                        {job.status === "locked" && <Loader2 className="h-3 w-3 mr-1 animate-spin" />}
                        {job.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-fg-muted">{job.retryCount}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </div>
  )
}
