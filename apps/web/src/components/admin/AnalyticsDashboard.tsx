"use client"

import { useEffect, useState } from "react"
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts"

type AnalyticsData = {
  queriesByDay: Array<{ day: string; queries: number }>
  topCollections: Array<{ name: string; queries: number }>
  feedbackDistribution: Array<{ name: string; value: number }>
  topUsers: Array<{ userId: number | null; queries: number }>
}

const ACCENT = "#7C6AF5"
const COLORS = ["#7C6AF5", "#9D8FF8", "#BDB4FC", "#D4CFFD"]

export function AnalyticsDashboard() {
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch("/api/admin/analytics")
      .then((r) => r.json())
      .then((d: { ok: boolean } & AnalyticsData) => {
        if (d.ok) setData(d)
      })
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Cargando analytics...</p>
  }

  if (!data) {
    return <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>No hay datos disponibles.</p>
  }

  const totalQueries = data.queriesByDay.reduce((s, d) => s + d.queries, 0)
  const totalFeedback = data.feedbackDistribution.reduce((s, d) => s + d.value, 0)

  return (
    <div className="space-y-8">
      {/* Stats cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {[
          { label: "Queries (30d)", value: totalQueries },
          { label: "Colecciones activas", value: data.topCollections.length },
          { label: "Feedback positivo", value: data.feedbackDistribution[0]?.value ?? 0 },
          { label: "Usuarios activos", value: data.topUsers.length },
        ].map((stat) => (
          <div
            key={stat.label}
            className="p-4 rounded-xl border space-y-1"
            style={{ borderColor: "var(--border)" }}
          >
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{stat.label}</p>
            <p className="text-2xl font-semibold">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Queries por día */}
      <div className="rounded-xl border p-4" style={{ borderColor: "var(--border)" }}>
        <h3 className="text-sm font-medium mb-4">Queries por día (últimos 30 días)</h3>
        {data.queriesByDay.length > 0 ? (
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={[...data.queriesByDay].reverse()}>
              <XAxis dataKey="day" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} />
              <Tooltip />
              <Line type="monotone" dataKey="queries" stroke={ACCENT} strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <p className="text-sm text-center py-8" style={{ color: "var(--muted-foreground)" }}>Sin datos aún</p>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
        {/* Top colecciones */}
        <div className="rounded-xl border p-4" style={{ borderColor: "var(--border)" }}>
          <h3 className="text-sm font-medium mb-4">Colecciones más consultadas</h3>
          {data.topCollections.length > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={data.topCollections} layout="vertical">
                <XAxis type="number" tick={{ fontSize: 11 }} />
                <YAxis type="category" dataKey="name" width={90} tick={{ fontSize: 11 }} />
                <Tooltip />
                <Bar dataKey="queries" fill={ACCENT} radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-center py-8" style={{ color: "var(--muted-foreground)" }}>Sin datos</p>
          )}
        </div>

        {/* Feedback */}
        <div className="rounded-xl border p-4" style={{ borderColor: "var(--border)" }}>
          <h3 className="text-sm font-medium mb-4">
            Distribución de feedback {totalFeedback > 0 && `(${totalFeedback} total)`}
          </h3>
          {totalFeedback > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <PieChart>
                <Pie
                  data={data.feedbackDistribution}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={80}
                  dataKey="value"
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  labelLine={false}
                >
                  {data.feedbackDistribution.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-center py-8" style={{ color: "var(--muted-foreground)" }}>Sin feedback aún</p>
          )}
        </div>
      </div>

      {/* Top usuarios */}
      {data.topUsers.length > 0 && (
        <div className="rounded-xl border p-4" style={{ borderColor: "var(--border)" }}>
          <h3 className="text-sm font-medium mb-3">Usuarios más activos</h3>
          <div className="space-y-1">
            {data.topUsers.map((u, i) => (
              <div key={i} className="flex items-center justify-between text-sm py-1">
                <span style={{ color: "var(--muted-foreground)" }}>Usuario #{u.userId ?? "sistema"}</span>
                <span className="font-medium">{u.queries} queries</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
