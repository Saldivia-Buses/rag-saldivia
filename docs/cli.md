# RAG Saldivia CLI — Referencia de comandos

## Instalación

```bash
# Desde el repo
bun run setup

# Instalar como comando global
cd apps/cli && bun link
# Ahora 'rag' está disponible en todo el sistema
```

## Configuración

La CLI lee las siguientes variables de entorno:

```bash
RAG_WEB_URL=http://localhost:3000    # URL del servidor Next.js
SYSTEM_API_KEY=...                   # Clave de autenticación service-to-service
RAG_SERVER_URL=http://localhost:8081 # URL del RAG Server (para rag status)
```

## Comandos

### `rag status`

Health check de todos los servicios con latencias.

```
✓ Next.js server     http://localhost:3000  (23ms)
✓ RAG Server         http://localhost:8081  (145ms)
✓ Milvus             localhost:19530        (8ms)
⚠ Mode Manager       http://localhost:8082  timeout — opcional
✗ OpenRouter Proxy   http://localhost:8083  connection refused

2/3 servicios críticos operativos
```

### `rag users`

```bash
rag users list                    # Tabla con todos los usuarios, roles, áreas
rag users create                  # Wizard interactivo
rag users delete <id>             # Con confirmación
```

### `rag collections`

```bash
rag collections list              # Lista con stats
rag collections create [nombre]   # Crear colección
rag collections delete [nombre]   # Con confirmación si tiene docs
```

### `rag ingest`

```bash
rag ingest start                  # Iniciar ingesta (interactivo)
rag ingest start -c manual -p /docs/   # Con flags
rag ingest status                 # Tabla de jobs con barra de progreso
rag ingest cancel <jobId>         # Con confirmación
```

### `rag config`

```bash
rag config get                    # Mostrar parámetros RAG actuales
rag config set temperature 0.3    # Cambiar un parámetro
rag config reset                  # Resetear a valores por defecto
```

### `rag audit`

```bash
rag audit log                     # Últimos 50 eventos
rag audit log -n 100 -l ERROR     # Filtrar por nivel
rag audit log -t rag.query        # Filtrar por tipo
rag audit replay 2026-03-24       # Black box replay desde fecha
rag audit export > events.json    # Exportar todos los eventos
```

### `rag db`

```bash
rag db migrate                    # Correr migraciones pendientes
rag db seed                       # Seed de datos de desarrollo
rag db reset                      # BORRAR DB y rehacer (con confirmación)
```

### `rag setup`

Equivalente a `bun run setup` — ejecuta el onboarding completo.

## Modo interactivo

Sin argumentos, `rag` abre un menú interactivo:

```bash
rag
```

```
  RAG Saldivia
  Modo interactivo

  ¿Qué querés hacer?
  ● rag status — Estado del sistema
  ○ rag users list — Lista usuarios
  ○ rag collections list — Colecciones
  ...
```
