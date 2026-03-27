# RAG Saldivia CLI

Interfaz de línea de comandos que habla con el servidor Next.js por **HTTP** (misma API que la UI). Requiere `SYSTEM_API_KEY` alineada con el servidor para endpoints admin.

## Instalación global

```bash
cd apps/cli
bun link
```

Verifica con `rag --help`.

## Variables de entorno

| Variable | Descripción |
|----------|-------------|
| `RAG_WEB_URL` | Base URL del servidor (default `http://localhost:3000`) |
| `SYSTEM_API_KEY` | Autenticación para rutas admin |

## Comandos

### `rag status`

Health check de servicios (Next.js, RAG, Milvus, etc.) con latencias.

### `rag users`

```bash
rag users list              # Tabla de usuarios
rag users create            # Asistente interactivo
rag users delete <id>       # Eliminar por ID
```

### `rag collections`

```bash
rag collections list
rag collections create [nombre]
rag collections delete [nombre]
```

### `rag ingest`

```bash
rag ingest start            # Wizard / flags -c / -p
rag ingest status
rag ingest cancel <jobId>
```

### `rag audit`

```bash
rag audit log                     # Últimos eventos
rag audit log -n 100 -l ERROR
rag audit replay <YYYY-MM-DD>
rag audit export
```

### `rag config`

```bash
rag config get [clave]
rag config set <clave> <valor>
rag config reset
```

### `rag db`

```bash
rag db migrate
rag db seed
rag db reset              # Destructivo — pide confirmación
```

### `rag setup`

Equivalente a `bun run setup` en la raíz del monorepo.

### Sin argumentos

`rag` abre un menú interactivo con comandos frecuentes.

## Referencia extendida

Ver [docs/cli.md](../../docs/cli.md) en el repositorio.
