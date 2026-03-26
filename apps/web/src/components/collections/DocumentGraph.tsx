"use client"

import { useEffect, useRef, useState } from "react"
import { ZoomIn, ZoomOut, RotateCcw } from "lucide-react"
import { Button } from "@/components/ui/button"

type Node = { id: string; name: string; group?: number; x?: number; y?: number; vx?: number; vy?: number; fx?: number | null; fy?: number | null }
type Edge = { source: string; target: string; weight: number }

type Props = {
  collection: string
  onNodeClick?: (docName: string) => void
}

const COLORS = ["#7C6AF5", "#9D8FF8", "#BDB4FC", "#22c55e", "#f59e0b", "#ef4444"]

export function DocumentGraph({ collection, onNodeClick }: Props) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [transform, setTransform] = useState({ x: 0, y: 0, k: 1 })

  useEffect(() => {
    fetch(`/api/collections/${encodeURIComponent(collection)}/embeddings`)
      .then((r) => r.json())
      .then((d: { ok: boolean; nodes?: Node[]; edges?: Edge[]; error?: string }) => {
        if (!d.ok) { setError(d.error ?? "Error"); return }
        // Posiciones iniciales aleatorias
        const initialNodes = (d.nodes ?? []).map((n) => ({
          ...n,
          x: 200 + Math.random() * 400,
          y: 150 + Math.random() * 300,
          vx: 0, vy: 0, fx: null, fy: null,
        }))
        setNodes(initialNodes)
        setEdges(d.edges ?? [])
      })
      .catch(() => setError("Error de red"))
      .finally(() => setLoading(false))
  }, [collection])

  // Simulación force-directed simple (sin d3-force para reducir bundle)
  useEffect(() => {
    if (nodes.length === 0) return

    let frame: number
    let current = [...nodes]

    function tick() {
      const W = svgRef.current?.clientWidth ?? 800
      const H = svgRef.current?.clientHeight ?? 500

      current = current.map((n) => {
        let fx = 0; let fy = 0

        // Repulsión entre nodos
        for (const other of current) {
          if (other.id === n.id) continue
          const dx = (n.x ?? 0) - (other.x ?? 0)
          const dy = (n.y ?? 0) - (other.y ?? 0)
          const dist = Math.sqrt(dx * dx + dy * dy) || 1
          const force = 2000 / (dist * dist)
          fx += (dx / dist) * force
          fy += (dy / dist) * force
        }

        // Atracción hacia el centro
        fx += (W / 2 - (n.x ?? 0)) * 0.01
        fy += (H / 2 - (n.y ?? 0)) * 0.01

        // Atracción por aristas
        for (const e of edges) {
          const otherId = e.source === n.id ? e.target : e.source === n.id ? e.target : null
          const other = otherId ? current.find((c) => c.id === otherId) : null
          if (!other) continue
          const dx = (other.x ?? 0) - (n.x ?? 0)
          const dy = (other.y ?? 0) - (n.y ?? 0)
          fx += dx * e.weight * 0.05
          fy += dy * e.weight * 0.05
        }

        const damping = 0.85
        const vx = ((n.vx ?? 0) + fx * 0.1) * damping
        const vy = ((n.vy ?? 0) + fy * 0.1) * damping
        return {
          ...n,
          vx, vy,
          x: Math.max(20, Math.min(W - 20, (n.x ?? 0) + vx)),
          y: Math.max(20, Math.min(H - 20, (n.y ?? 0) + vy)),
        }
      })

      setNodes([...current])
      frame = requestAnimationFrame(tick)
    }

    // Correr la simulación por 60 frames y parar
    let count = 0
    function run() {
      tick()
      count++
      if (count < 80) frame = requestAnimationFrame(run)
    }
    frame = requestAnimationFrame(run)
    return () => cancelAnimationFrame(frame)
  }, [edges, nodes.length])

  if (loading) return <div className="flex items-center justify-center h-64 text-sm" style={{ color: "var(--muted-foreground)" }}>Cargando grafo...</div>
  if (error) return <div className="flex items-center justify-center h-64 text-sm" style={{ color: "var(--destructive)" }}>{error}</div>

  const nodeMap = new Map(nodes.map((n) => [n.id, n]))

  return (
    <div className="relative rounded-xl border overflow-hidden" style={{ borderColor: "var(--border)", height: 500 }}>
      <div className="absolute top-3 right-3 flex gap-1 z-10">
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setTransform((t) => ({ ...t, k: Math.min(t.k * 1.2, 3) }))}>
          <ZoomIn size={13} />
        </Button>
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setTransform((t) => ({ ...t, k: Math.max(t.k / 1.2, 0.3) }))}>
          <ZoomOut size={13} />
        </Button>
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setTransform({ x: 0, y: 0, k: 1 })}>
          <RotateCcw size={13} />
        </Button>
      </div>

      <svg
        ref={svgRef}
        className="w-full h-full"
        style={{ background: "var(--muted)" }}
      >
        <g transform={`translate(${transform.x},${transform.y}) scale(${transform.k})`}>
          {/* Aristas */}
          {edges.map((e, i) => {
            const s = nodeMap.get(e.source)
            const t = nodeMap.get(e.target)
            if (!s || !t) return null
            return (
              <line
                key={i}
                x1={s.x} y1={s.y} x2={t.x} y2={t.y}
                stroke="var(--border)"
                strokeWidth={e.weight * 2}
                strokeOpacity={0.6}
              />
            )
          })}

          {/* Nodos */}
          {nodes.map((n) => {
            const color = COLORS[(n.group ?? 0) % COLORS.length]
            const label = n.name.replace(/\.[^.]+$/, "").slice(0, 15)
            return (
              <g
                key={n.id}
                transform={`translate(${n.x},${n.y})`}
                className="cursor-pointer"
                onClick={() => onNodeClick?.(n.name)}
              >
                <circle r={18} fill={color} fillOpacity={0.85} stroke="white" strokeWidth={1.5} />
                <text
                  textAnchor="middle"
                  y={30}
                  fontSize={10}
                  fill="var(--foreground)"
                  className="select-none pointer-events-none"
                >
                  {label}
                </text>
              </g>
            )
          })}
        </g>
      </svg>
    </div>
  )
}
