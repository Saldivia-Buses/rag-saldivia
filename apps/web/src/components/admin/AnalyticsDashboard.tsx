"use client"

import { useEffect, useState } from "react"
import {
  LineChart, Line, BarChart, Bar, PieChart, Pie, Cell,
  XAxis, YAxis, Tooltip, ResponsiveContainer,
} from "recharts"
import { StatCard } from "@/components/ui/stat-card"
import { SkeletonCard } from "@/components/ui/skeleton"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { BarChart2, MessageSquare, FolderOpen, ThumbsUp, Users } from "lucide-react"

type AnalyticsData = {
  queriesByDay: Array<{ day: string; queries: number }>
  topCollections: Array<{ name: string; queries: number }>
  feedbackDistribution: Array<{ name: string; value: number }>
  topUsers: Array<{ userId: number | null; queries: number }>
}

// Colores del design system — navy accent palette
const CHART_PRIMARY   = "#1a5276"
const CHART_SECONDARY = "#4a9fd4"
const CHART_COLORS    = ["#1a5276", "#4a9fd4", "#93c5e8", "#d4e8f7"]

export function AnalyticsDashboard() {
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch("/api/admin/analytics")
      .then((r) => r.json())
      .then((d: { ok: boolean } & AnalyticsData) => { if (d.ok) setData(d) })
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="p-6 space-y-6">
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {[0,1,2,3].map((i) => <SkeletonCard key={i} />)}
        </div>
        <SkeletonCard className="h-48" />
      </div>
    )
  }

  if (!data) {
    return (
      <div className="p-6">
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={BarChart2} />
          <EmptyPlaceholder.Title>Sin datos de analytics</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Los datos aparecerán cuando haya actividad en el sistema.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      </div>
    )
  }

  const totalQueries  = data.queriesByDay.reduce((s, d) => s + d.queries, 0)
  const totalFeedback = data.feedbackDistribution.reduce((s, d) => s + d.value, 0)
  const positiveFeedback = data.feedbackDistribution[0]?.value ?? 0
  const positiveRate  = totalFeedback > 0 ? Math.round((positiveFeedback / totalFeedback) * 100) : 0

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-lg font-semibold text-fg">Analytics</h1>
        <p className="text-sm text-fg-muted mt-0.5">Últimos 30 días</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <StatCard label="Queries (30d)" value={totalQueries.toLocaleString()} icon={MessageSquare} />
        <StatCard label="Colecciones activas" value={data.topCollections.length} icon={FolderOpen} />
        <StatCard label="Tasa de feedback positivo" value={`${positiveRate}%`} delta={positiveRate > 50 ? 1 : -1} icon={ThumbsUp} />
        <StatCard label="Usuarios activos" value={data.topUsers.length} icon={Users} />
      </div>

      {/* Queries por día */}
      <div className="rounded-xl border border-border bg-surface p-5">
        <h3 className="text-sm font-semibold text-fg mb-4">Queries por día</h3>
        {data.queriesByDay.length > 0 ? (
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={[...data.queriesByDay].reverse()}>
              <XAxis dataKey="day" tick={{ fontSize: 11, fill: "var(--fg-muted)" }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 11, fill: "var(--fg-muted)" }} axisLine={false} tickLine={false} />
              <Tooltip
                contentStyle={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: "8px", fontSize: 12 }}
                labelStyle={{ color: "var(--fg)" }}
              />
              <Line type="monotone" dataKey="queries" stroke={CHART_PRIMARY} strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <p className="text-sm text-center py-8 text-fg-muted">Sin datos aún</p>
        )}
      </div>

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
        {/* Top colecciones */}
        <div className="rounded-xl border border-border bg-surface p-5">
          <h3 className="text-sm font-semibold text-fg mb-4">Colecciones más consultadas</h3>
          {data.topCollections.length > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={data.topCollections} layout="vertical">
                <XAxis type="number" tick={{ fontSize: 11, fill: "var(--fg-muted)" }} axisLine={false} tickLine={false} />
                <YAxis type="category" dataKey="name" width={90} tick={{ fontSize: 11, fill: "var(--fg-muted)" }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: "8px", fontSize: 12 }}
                />
                <Bar dataKey="queries" fill={CHART_PRIMARY} radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-center py-8 text-fg-muted">Sin datos</p>
          )}
        </div>

        {/* Feedback */}
        <div className="rounded-xl border border-border bg-surface p-5">
          <h3 className="text-sm font-semibold text-fg mb-4">
            Distribución de feedback{totalFeedback > 0 && <span className="text-fg-muted font-normal ml-1">({totalFeedback} total)</span>}
          </h3>
          {totalFeedback > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <PieChart>
                <Pie
                  data={data.feedbackDistribution}
                  cx="50%" cy="50%"
                  innerRadius={50} outerRadius={80}
                  dataKey="value"
                  label={({ name, percent }: { name?: string; percent?: number }) =>
                    `${name ?? ""} ${((percent ?? 0) * 100).toFixed(0)}%`
                  }
                  labelLine={false}
                >
                  {data.feedbackDistribution.map((_, i) => (
                    <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ background: "var(--surface)", border: "1px solid var(--border)", borderRadius: "8px", fontSize: 12 }}
                />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-center py-8 text-fg-muted">Sin feedback aún</p>
          )}
        </div>
      </div>

      {/* Top usuarios */}
      {data.topUsers.length > 0 && (
        <div className="rounded-xl border border-border bg-surface p-5">
          <h3 className="text-sm font-semibold text-fg mb-3">Usuarios más activos</h3>
          <div className="divide-y divide-border">
            {data.topUsers.map((u, i) => (
              <div key={i} className="flex items-center justify-between py-2.5 text-sm">
                <span className="text-fg-muted">Usuario #{u.userId ?? "sistema"}</span>
                <span className="font-medium text-fg tabular-nums">{u.queries} queries</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
