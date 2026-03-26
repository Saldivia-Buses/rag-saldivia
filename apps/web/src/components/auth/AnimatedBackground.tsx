"use client"

/**
 * Fondo animado para la página de login.
 * Gradiente mesh CSS animado — sin WebGL, funciona en todos los browsers.
 * En dark mode usa tonos warm-dark; en light mode usa tonos crema-navy.
 */
export function AnimatedBackground() {
  return (
    <div className="fixed inset-0 -z-10 overflow-hidden">
      {/* Capa base */}
      <div className="absolute inset-0 bg-bg" />

      {/* Orbe 1 — accent-subtle, top-left */}
      <div
        className="absolute -top-32 -left-32 w-[500px] h-[500px] rounded-full opacity-40 blur-3xl animate-pulse"
        style={{
          background: "radial-gradient(circle, var(--accent-subtle) 0%, transparent 70%)",
          animationDuration: "4s",
        }}
      />

      {/* Orbe 2 — accent, bottom-right */}
      <div
        className="absolute -bottom-48 -right-24 w-[600px] h-[600px] rounded-full opacity-20 blur-3xl"
        style={{
          background: "radial-gradient(circle, var(--accent) 0%, transparent 70%)",
          animation: "float 6s ease-in-out infinite",
        }}
      />

      {/* Orbe 3 — surface-2, center */}
      <div
        className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] h-[400px] rounded-full opacity-30 blur-3xl"
        style={{
          background: "radial-gradient(circle, var(--surface-2) 0%, transparent 70%)",
          animation: "float 8s ease-in-out infinite reverse",
        }}
      />

      <style>{`
        @keyframes float {
          0%, 100% { transform: translate(-50%, -50%) scale(1); }
          50%       { transform: translate(-50%, -52%) scale(1.05); }
        }
      `}</style>
    </div>
  )
}
