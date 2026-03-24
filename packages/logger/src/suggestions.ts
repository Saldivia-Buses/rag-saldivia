/**
 * Mapeo de errores conocidos → mensajes accionables.
 * Cuando el logger captura un error, busca aquí si hay una sugerencia mejor
 * que el stack trace genérico.
 */

type ErrorSuggestion = {
  pattern: RegExp
  message: string
  suggestion: string
}

export const ERROR_SUGGESTIONS: ErrorSuggestion[] = [
  {
    pattern: /ECONNREFUSED.*8081/,
    message: "No se puede conectar al RAG Server",
    suggestion: "El RAG Server no está corriendo en el puerto 8081.\n→ Verificá con: make status\n→ Para levantar: make deploy PROFILE=workstation-1gpu",
  },
  {
    pattern: /ECONNREFUSED.*19530/,
    message: "No se puede conectar a Milvus",
    suggestion: "Milvus no está corriendo en el puerto 19530.\n→ Verificá los contenedores Docker: docker ps | grep milvus",
  },
  {
    pattern: /jwt expired/i,
    message: "Token JWT expirado",
    suggestion: "El token de sesión expiró.\n→ El cliente debe hacer refresh del token o re-autenticarse.",
  },
  {
    pattern: /invalid signature/i,
    message: "Firma JWT inválida",
    suggestion: "JWT_SECRET puede haber cambiado o el token es de otra instancia.\n→ Verificá que JWT_SECRET en .env.local sea consistente.",
  },
  {
    pattern: /SQLITE_BUSY/,
    message: "SQLite bloqueado",
    suggestion: "Hay otra transacción bloqueando la DB.\n→ Verificá que no haya múltiples workers de ingesta corriendo.\n→ Si persiste: bun run setup:reset",
  },
  {
    pattern: /SQLITE_CORRUPT/,
    message: "Base de datos SQLite corrupta",
    suggestion: "El archivo de DB puede estar corrupto.\n→ Restaurá desde backup en data/app.db.bak\n→ O reseteá la DB: bun run setup:reset (BORRA todos los datos)",
  },
  {
    pattern: /collection.*not found/i,
    message: "Colección no encontrada en Milvus",
    suggestion: "La colección no existe en Milvus.\n→ Verificá con: rag collections list\n→ Creá la colección: rag collections create <nombre>",
  },
  {
    pattern: /EADDRINUSE.*3000/,
    message: "Puerto 3000 en uso",
    suggestion: "Otro proceso está usando el puerto 3000.\n→ Verificá: lsof -i :3000 (Linux/Mac) o netstat -ano | findstr :3000 (Windows)\n→ Cambiá el puerto: PORT=3001 bun run dev",
  },
  {
    pattern: /Cannot find module/,
    message: "Módulo no encontrado",
    suggestion: "Falta una dependencia.\n→ Corré: bun install\n→ Si persiste, verificá que el paquete esté en package.json",
  },
]

export function getSuggestion(errorMessage: string): string | null {
  const suggestion = ERROR_SUGGESTIONS.find((s) => s.pattern.test(errorMessage))
  return suggestion ? suggestion.suggestion : null
}
