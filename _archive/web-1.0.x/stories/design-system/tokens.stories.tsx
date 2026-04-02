import type { Meta, StoryObj } from "@storybook/react"

const meta: Meta = {
  title: "Design System/Tokens",
  parameters: { layout: "fullscreen" },
}
export default meta

function Swatch({ name, value, label }: { name: string; value: string; label?: string }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div
        className="h-12 w-full rounded-lg border border-black/5 shadow-sm"
        style={{ background: value }}
      />
      <div>
        <p className="text-xs font-mono font-medium text-fg">{name}</p>
        {label && <p className="text-xs text-fg-muted">{label}</p>}
      </div>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-10">
      <h2 className="text-sm font-semibold text-fg-subtle uppercase tracking-wide mb-4">{title}</h2>
      <div className="grid grid-cols-4 gap-4 sm:grid-cols-6 lg:grid-cols-8">{children}</div>
    </div>
  )
}

export const Palette: StoryObj = {
  render: () => (
    <div className="p-8 bg-bg min-h-screen">
      <h1 className="text-2xl font-semibold text-fg mb-2">Design Tokens</h1>
      <p className="text-sm text-fg-muted mb-8">Paleta completa de RAG Saldivia — Warm Intelligence</p>

      <Section title="Fondos">
        <Swatch name="--bg" value="var(--bg)" label="#faf8f4" />
        <Swatch name="--surface" value="var(--surface)" label="#f0ebe0" />
        <Swatch name="--surface-2" value="var(--surface-2)" label="#e8e1d4" />
      </Section>

      <Section title="Bordes">
        <Swatch name="--border" value="var(--border)" label="#ede9e0" />
        <Swatch name="--border-strong" value="var(--border-strong)" label="#d5cfc4" />
      </Section>

      <Section title="Texto">
        <Swatch name="--fg" value="var(--fg)" label="#1a1a1a" />
        <Swatch name="--fg-muted" value="var(--fg-muted)" label="#5a5048" />
        <Swatch name="--fg-subtle" value="var(--fg-subtle)" label="#9a9088" />
      </Section>

      <Section title="Acento Navy">
        <Swatch name="--accent" value="var(--accent)" label="#1a5276" />
        <Swatch name="--accent-hover" value="var(--accent-hover)" label="#154360" />
        <Swatch name="--accent-dark" value="var(--accent-dark)" label="#0d3349" />
        <Swatch name="--accent-subtle" value="var(--accent-subtle)" label="#d4e8f7" />
        <Swatch name="--accent-fg" value="var(--accent-fg)" label="#ffffff" />
      </Section>

      <Section title="Estados">
        <Swatch name="--destructive" value="var(--destructive)" label="Error" />
        <Swatch name="--destructive-subtle" value="var(--destructive-subtle)" label="Error bg" />
        <Swatch name="--success" value="var(--success)" label="Éxito" />
        <Swatch name="--success-subtle" value="var(--success-subtle)" label="Éxito bg" />
        <Swatch name="--warning" value="var(--warning)" label="Advertencia" />
        <Swatch name="--warning-subtle" value="var(--warning-subtle)" label="Warning bg" />
      </Section>

      <Section title="Tipografía">
        {(["xs", "sm", "base", "lg", "xl", "2xl", "3xl", "4xl"] as const).map((size) => (
          <div key={size} className="col-span-8 flex items-baseline gap-4 border-b border-border pb-3">
            <span className="text-xs font-mono text-fg-muted w-16">text-{size}</span>
            <span style={{ fontSize: `var(--text-${size})` }} className="text-fg">
              RAG Saldivia — Warm Intelligence
            </span>
          </div>
        ))}
      </Section>
    </div>
  ),
}
