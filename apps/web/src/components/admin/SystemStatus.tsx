"use client"

import { useRouter } from "next/navigation"
import { Users, Database, FolderOpen, Loader2, AlertTriangle, RefreshCw } from "lucide-react"
import type { DbIngestionQueueItem } from "@rag-saldivia/db"

type Stats = {
  activeUsers: number
  areas: number
  collections: number
  activeJobs: number
  recentErrors: number
}

function StatCard({ icon, label, value, color }: {
  icon: React.ReactNode
  label: string
  value: number
  color?: string
}) {
  return (
    <div className="p-5 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
      <div className="flex items-center justify-between">
        <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>{label}</span>
        <span style={{ color: color ?? "var(--muted-foreground)" }}>{icon}</span>
      </div>
      <p className="text-3xl font-bold">{value}</p>
    </div>
  )
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
    <div className="space-y-6">
      {/* Stats cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard icon={<Users size={18} />} label="Usuarios activos" value={stats.activeUsers} />
        <StatCard icon={<Database size={18} />} label="Áreas" value={stats.areas} />
        <StatCard icon={<FolderOpen size={18} />} label="Colecciones" value={stats.collections} />
        <StatCard
          icon={<AlertTriangle size={18} />}
          label="Errores (24hs)"
          value={stats.recentErrors}
          color={stats.recentErrors > 0 ? "var(--destructive)" : undefined}
        />
      </div>

      {/* Refresh */}
      <div className="flex justify-end">
        <button
          onClick={() => router.refresh()}
          className="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm border"
          style={{ borderColor: "var(--border)" }}
        >
          <RefreshCw size={14} />
          Actualizar
        </button>
      </div>

      {/* Jobs activos */}
      <div>
        <h3 className="font-medium text-sm mb-3">Jobs de ingesta activos</h3>
        {activeJobs.length === 0 ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Sin jobs activos</p>
        ) : (
          <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
            <table className="w-full text-sm">
              <thead>
                <tr style={{ background: "var(--muted)" }}>
                  <th className="text-left px-4 py-3 font-medium">ID</th>
                  <th className="text-left px-4 py-3 font-medium">Colección</th>
                  <th className="text-left px-4 py-3 font-medium">Estado</th>
                  <th className="text-left px-4 py-3 font-medium">Reintentos</th>
                </tr>
              </thead>
              <tbody>
                {activeJobs.map((job, i) => (
                  <tr key={job.id} style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}>
                    <td className="px-4 py-3 font-mono text-xs">{job.id.slice(0, 8)}</td>
                    <td className="px-4 py-3">{job.collection}</td>
                    <td className="px-4 py-3">
                      <span className="flex items-center gap-1.5">
                        {job.status === "locked" && <Loader2 size={12} className="animate-spin" />}
                        <span className={job.status === "locked" ? "text-blue-600" : ""}>
                          {job.status}
                        </span>
                      </span>
                    </td>
                    <td className="px-4 py-3" style={{ color: "var(--muted-foreground)" }}>
                      {job.retryCount}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
