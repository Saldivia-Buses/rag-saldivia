/**
 * Cliente Redis singleton para @rag-saldivia/db.
 *
 * Redis es una dependencia requerida del sistema — sin fallback.
 * Si REDIS_URL no está configurado, lanza un error claro con instrucciones.
 *
 * ADR-010: Redis como dependencia requerida del sistema.
 *
 * IMPORTANTE: NO importar @rag-saldivia/logger aquí.
 * Logger importa de @rag-saldivia/db → importar logger aquí crearía
 * dependencia circular: db → logger → db (violación de ADR-005).
 * Usar console.error para errores de conexión.
 */

import Redis from "ioredis"

let _client: Redis | null = null

export function getRedisClient(): Redis {
  if (!_client) {
    const url = process.env["REDIS_URL"]
    if (!url) {
      throw new Error(
        "REDIS_URL no configurado.\n" +
        "Redis es requerido (ADR-010). Para dev local:\n" +
        "  docker run -d -p 6379:6379 redis:alpine\n" +
        "  Agregar REDIS_URL=redis://localhost:6379 en .env.local"
      )
    }
    _client = new Redis(url, { maxRetriesPerRequest: 3 })
    _client.on("error", (err) => {
      console.error("[Redis] connection error:", String(err))
    })
  }
  return _client
}

/** Solo para tests — permite reiniciar el singleton entre tests. */
export function _resetRedisForTesting(): void {
  if (_client) {
    _client.disconnect()
    _client = null
  }
}
