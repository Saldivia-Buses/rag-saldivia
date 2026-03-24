# Onboarding — 5 minutos para arrancar

## Prerrequisitos

- [Bun](https://bun.sh) >= 1.0 (`curl -fsSL https://bun.sh/install | bash`)
- [Docker](https://docs.docker.com/get-docker/) con Docker Compose v2
- Git

## Setup — Ubuntu nativo (producción / workstation)

El caso más simple. Todo funciona out-of-the-box:

```bash
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia
git checkout experimental/ultra-optimize
bun run setup
bun run dev
```

## Setup — WSL2 (desarrollo en Windows)

WSL2 tiene una particularidad: `bun install` en el filesystem Windows (`/mnt/c/`) no crea symlinks correctamente. La solución es clonar el repo en el filesystem **nativo de Linux**:

```bash
# En la terminal WSL2 — clonar en HOME de Linux, NO en /mnt/c/
cd ~
git clone https://github.com/Camionerou/rag-saldivia
cd rag-saldivia
git checkout experimental/ultra-optimize
bun install

# Crear symlinks de @libsql (solo necesario en WSL2)
bash scripts/link-libsql.sh

# Setup del .env y la DB
cp .env.example apps/web/.env.local
# Editar apps/web/.env.local: cambiar JWT_SECRET, SYSTEM_API_KEY, DATABASE_PATH a ruta absoluta

bun packages/db/src/migrate.ts
bun packages/db/src/seed.ts

# Levantar el servidor
node_modules/.bin/next dev /path/to/rag-saldivia/apps/web --port 3000
```

El script de setup hace todo:

1. Verifica que Bun, Docker y los puertos estén disponibles
2. Crea `.env.local` desde `.env.example`
3. Instala todas las dependencias con Bun workspaces
4. Crea la base de datos SQLite y aplica migraciones
5. Crea datos de desarrollo (usuario `admin@localhost`)
6. Muestra un resumen del estado

## Arrancar el servidor

```bash
bun run dev
```

Abrí http://localhost:3000

Credenciales de desarrollo:
- Email: `admin@localhost`
- Contraseña: `changeme`

## Sin Docker (modo mock)

Si no tenés Docker, podés desarrollar la UI sin el RAG real:

```bash
echo "MOCK_RAG=true" >> .env.local
bun run dev
```

El servidor simula respuestas del RAG. Útil para trabajo en el frontend.

## Variables de entorno importantes

Ver `.env.example` para la lista completa documentada.

Las más importantes:

```env
JWT_SECRET=...          # Generar: openssl rand -base64 32
SYSTEM_API_KEY=...      # Generar: openssl rand -hex 32
RAG_SERVER_URL=...      # Default: http://localhost:8081
MOCK_RAG=false          # true para desarrollo sin Docker
```

## Comandos útiles

```bash
make setup              # Onboarding completo
make dev                # Servidor de desarrollo
make status             # Estado de los servicios
make test               # Todos los tests
make reset              # Limpiar DB y rehacer seed

# CLI (instalar globalmente: cd apps/cli && bun link)
rag status              # Estado del sistema
rag users list          # Lista usuarios
rag audit log           # Últimos eventos del sistema
```

## Estructura del proyecto

Ver [docs/architecture.md](./architecture.md) para el diagrama completo.

## Deployar a producción

```bash
git checkout main        # La branch main tiene el stack estable (Python + SvelteKit)
make deploy PROFILE=workstation-1gpu
```

La branch `experimental/ultra-optimize` es el nuevo stack que está en desarrollo.
Cuando esté listo, se mergeará a main.
