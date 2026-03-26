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

      {/* Orbe 1 — azul claro, top-left */}
      <div
        className="absolute -top-24 -left-24 w-[520px] h-[520px] rounded-full blur-3xl"
        style={{
          background: "radial-gradient(circle at 40% 40%, #bfdbf7 0%, #d4e8f7 30%, transparent 70%)",
          opacity: 0.75,
          animation: "pulse-slow 5s ease-in-out infinite",
        }}
      />

      {/* Orbe 2 — navy suave, bottom-right */}
      <div
        className="absolute -bottom-32 -right-16 w-[580px] h-[580px] rounded-full blur-3xl"
        style={{
          background: "radial-gradient(circle at 60% 60%, #93c5e8 0%, #1a527622 60%, transparent 80%)",
          opacity: 0.6,
          animation: "float 7s ease-in-out infinite",
        }}
      />

      {/* Orbe 3 — crema cálido, off-center */}
      <div
        className="absolute top-[55%] right-[25%] w-[350px] h-[350px] rounded-full blur-2xl"
        style={{
          background: "radial-gradient(circle, #e8d5c0 0%, transparent 65%)",
          opacity: 0.5,
          animation: "float 9s ease-in-out infinite reverse",
        }}
      />

      <style>{`
        @keyframes float {
          0%, 100% { transform: scale(1) translateY(0); }
          50%       { transform: scale(1.06) translateY(-16px); }
        }
        @keyframes pulse-slow {
          0%, 100% { opacity: 0.75; transform: scale(1); }
          50%       { opacity: 0.55; transform: scale(1.04); }
        }
      `}</style>
    </div>
  )
}
